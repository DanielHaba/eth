package eth

import (
	"context"
	"errors"
	"sync"
)

type Bridge struct {
	ctx      context.Context
	log      *Logger
	fdb      Fdb
	ch       chan *Frame
	links    map[int]Interface
	linkLock sync.RWMutex
	stop     func()
	ifstop   map[int]func()
}

func NewBridge(ctx context.Context) *Bridge {
	ctx, stop := context.WithCancel(ctx)
	br := &Bridge{
		log:  NewLogger(ctx),
		ch:   make(chan *Frame, 4096),
		ctx:  ctx,
		stop: stop,
	}
	go br.send(context.Background())

	return br
}

func (br *Bridge) Link(ifce Interface) error {
	br.linkLock.Lock()
	defer br.linkLock.Unlock()

	if err := br.link(ifce); err != nil {
		return err
	}
	if br.ifstop == nil {
		br.ifstop = map[int]func(){}
	}
	ctx, stop := context.WithCancel(context.Background())
	br.ifstop[ifce.Index()] = stop

	go func() {
		defer br.Unlink(ifce)
		br.recv(ctx, ifce)
	}()

	return nil
}

func (br *Bridge) Unlink(ifce Interface) error {
	br.linkLock.Lock()
	defer br.linkLock.Unlock()

	return br.unlink(ifce)
}

func (br *Bridge) Close() error {
	br.stop()
	return nil
}

func (br *Bridge) link(ifce Interface) error {
	if br.links == nil {
		br.links = map[int]Interface{}
	}
	if _, ok := br.links[ifce.Index()]; ok {
		return errors.New("already linked")
	}
	br.links[ifce.Index()] = ifce

	return nil
}

func (br *Bridge) unlink(ifce Interface) error {
	if br.ifstop != nil {
		if stop, ok := br.ifstop[ifce.Index()]; ok {
			stop()
			delete(br.ifstop, ifce.Index())
		}
	}
	if br.links != nil {
		if _, ok := br.links[ifce.Index()]; ok {
			delete(br.links, ifce.Index())
			br.fdb.Clear(ifce)
			return nil
		}
	}
	return errors.New("not linked")
}

func (br *Bridge) send(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-br.ctx.Done():
			return

		case frame := <-br.ch:
			br.broadcast(frame)
		}
	}
}

func (br *Bridge) recv(ctx context.Context, ifce Interface) {
	var buffer [64 * 1024]byte
	for {
		select {
		case <-ctx.Done():
			return
		case <-br.ctx.Done():
			return

		default:
			if frame, err := br.recvfrom(ifce, buffer[:]); err == nil && frame != nil {
				br.ch <- frame
			}
		}
	}

}

func (br *Bridge) broadcast(frame *Frame) error {
	br.linkLock.RLock()
	defer br.linkLock.RUnlock()
	for _, link := range br.links {
		if link != frame.Origin {
			go br.sendto(link, frame.Copy())
		}
	}
	return nil
}

func (br *Bridge) recvfrom(src Interface, buffer []byte) (frame *Frame, err error) {
	n, _, err := src.Recv(buffer)
	if err != nil {
		return
	}

	data := make([]byte, n)
	copy(data, buffer)
	if err != nil {
		return
	}
	frame = NewFrame(src, data)
	br.log.Recive(src, frame)

	return
}

func (br *Bridge) sendto(dst Interface, frame *Frame) error {
	frame.Source = frame.Origin.MAC()
	br.log.Send(dst, frame)
	return dst.Send(frame.Destination, frame.Data)
}

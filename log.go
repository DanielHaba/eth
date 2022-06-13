package eth

import (
	"context"
	"log"
)

type Direction uint8

const (
	Incoming Direction = 0
	Outgoing Direction = 1
)

type DirectFrame struct {
	*Frame
	Origin Interface
	Dir    Direction
}

type Logger struct {
	ch chan *DirectFrame
}

func NewLogger(ctx context.Context) *Logger {
	logger := &Logger{ch: make(chan *DirectFrame, 2048)}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return

			case frame := <-logger.ch:
				logger.logFrame(frame)
			}
		}
	}()

	return logger
}

func (logger *Logger) Recive(org Interface, frame *Frame) {
	if !frame.Source.IsZero() {
		logger.ch <- &DirectFrame{Frame: frame, Origin: org, Dir: Outgoing}
	}
}

func (logger *Logger) Send(org Interface, frame *Frame) {
	if !frame.Destination.IsZero() {
		logger.ch <- &DirectFrame{Frame: frame, Origin: org, Dir: Incoming}
	}
}

func (*Logger) logFrame(frame *DirectFrame) {
	dir := "<"
	payload := ""
	typ := frame.Type
	if frame.Dir == Outgoing {
		dir = ">"
	}
	if typ == VLAN {
		typ = frame.VType
	}
	switch typ {
	case ARP:
		payload = (*ARPPayload)(frame.Payload).String()
	}
	log.Println(frame.Origin.Name(), dir, frame.Header.String(), payload)
}

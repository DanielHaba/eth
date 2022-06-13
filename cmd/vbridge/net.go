package main

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/DanielHaba/eth"
)

var bridges = map[string]*eth.Bridge{}
var bridgeLock sync.Mutex
var ifces = map[string]*sync.Pool{}
var ifceLock sync.Mutex

func Up(ctx context.Context, name string) error {
	bridgeLock.Lock()
	defer bridgeLock.Unlock()

	if _, ok := bridges[name]; ok {
		return errors.New("bridge already exists")
	}
	bridges[name] = eth.NewBridge(ctx)

	return nil
}

func Down(name string) error {
	bridgeLock.Lock()
	defer bridgeLock.Unlock()

	if bridge, ok := bridges[name]; ok {
		bridge.Close()
		delete(bridges, name)
	}
	return errors.New("bridge not exists")
}

func Link(br string, ifce string) error {
	bridge, err := Bridge(br)
	if err != nil {
		return err
	}
	sock, err := Ifce(ifce)
	if err != nil {
		return err
	}
	return bridge.Link(sock)
}

func Unlink(br string, ifce string) error {
	bridge, err := Bridge(br)
	if err != nil {
		return err
	}
	sock, err := Ifce(ifce)
	if err != nil {
		return err
	}
	return bridge.Unlink(sock)
}

func Bridge(name string) (*eth.Bridge, error) {
	bridgeLock.Lock()
	defer bridgeLock.Unlock()

	if bridge, ok := bridges[name]; ok {
		return bridge, nil
	}
	return nil, errors.New("bridge not exists")
}

func Ifce(name string) (ifce eth.Interface, err error) {
	ifceLock.Lock()
	defer ifceLock.Unlock()

	if ifces == nil {
		ifces = map[string]*sync.Pool{}
	}
	pool, ok := ifces[name]
	if !ok {
		pool = &sync.Pool{
			New: func() any { 
				ifce, err := eth.Open(name)
				if err != nil {
					panic(err)
				}
				return ifce
			},
		}
		ifces[name] = pool
	}

	defer func() {
		if rec := recover(); rec != nil {
			err = fmt.Errorf("cannot open interface %s: %s", name, rec)
		}
	}()

	ifce = pool.Get().(eth.Interface)
	pool.Put(ifce)

	return
}

func Cleanup() {
	bridgeLock.Lock()
	defer bridgeLock.Unlock()
	ifceLock.Lock()
	defer ifceLock.Unlock()

	for _, br := range bridges {
		br.Close()
	}
	for _, pool := range ifces {
		ifce := pool.Get().(eth.Interface)
		ifce.Close()
	}
}
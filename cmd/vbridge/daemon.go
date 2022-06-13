package main

import (
	"context"
	"errors"
	"net"
	"os"
)

const daemonSock = "/var/run/vbridge.sock"

func Start(ctx context.Context) error {

	if err := os.RemoveAll(daemonSock); err != nil {
		return err
	}

	sock, err := net.Listen("unix", daemonSock)
	if err != nil {
		return err
	}
	defer os.RemoveAll(daemonSock)
	defer sock.Close()

	ctx, stop := context.WithCancel(ctx)
	defer stop()

	go func() {
		for {
			conn, err := sock.Accept()
			if errors.Is(err, net.ErrClosed) {
				return
			}
			if err == nil {
				be := &Daemon{
					Local:  Local{ReadWriteCloser: conn},
					Closer: Stopper(stop),
				}
				Exec(ctx, be)
			}
			if conn != nil {
				conn.Close()
			}
		}
	}()

	<-ctx.Done()
	return nil
}

func Connect(ctx context.Context) (net.Conn, error) {
	if !HasDaemon() {
		return nil, errors.New("daemon is not running")
	}
	return net.Dial("unix", daemonSock)
}

func HasDaemon() bool {
	_, err := os.Stat(daemonSock)
	return !os.IsNotExist(err)
}

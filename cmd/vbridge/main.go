package main

import (
	"context"
	"io"
	"os"
	"os/signal"
	"strings"
)

func main() {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		os.Kill,
	)
	defer cancel()

	ctx, stop := context.WithCancel(ctx)
	defer stop()

	io := &IO{
		Reader: strings.NewReader(strings.Join(os.Args[1:], " ")),
		Writer: os.Stdout,
		Closer: Stopper(stop),
	}

	var be Backend
	if conn, err := Connect(ctx); err == nil {
		defer conn.Close()
		be = &Remote{Conn: conn, ReadWriteCloser: io}
	} else {
		be = &Local{ReadWriteCloser: io}
	}

	Exec(ctx, be)
	<-ctx.Done()
	Cleanup()
}

type IO struct {
	io.Reader
	io.Writer
	io.Closer
}

type Stopper func()

func (fn Stopper) Close() error {
	fn()
	return nil
}

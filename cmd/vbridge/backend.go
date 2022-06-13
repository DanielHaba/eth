package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
)

type Backend interface {
	io.ReadWriteCloser
	Up(ctx context.Context, br string) error
	Down(br string) error
	Link(br string, ifce string) error
	Unlink(br string, ifce string) error
	Start(ctx context.Context) error
	Stop() error
}

func Exec(ctx context.Context, be Backend) {
	var in io.Reader = be
	var out io.Writer = be
	s := bufio.NewScanner(in)
	s.Split(bufio.ScanWords)

	for s.Scan() {
		cmd := s.Text()
		switch cmd {
		case "close":
			return

		case "start":
			if err := be.Start(ctx); err != nil {
				fmt.Fprintf(out, "cannot start daemon: %s", err)
			}
			fmt.Fprintf(out, "success\n")
			be.Close()
			return

		case "stop":
			be.Stop()
			fmt.Fprintf(out, "success\n")
			return

		case "up":
			args, err := ScanN(s, 1)
			if err != nil {
				fmt.Fprintf(out, "invalid arguments: %s\n", err)
				return
			}
			if err = be.Up(ctx, args[0]); err != nil {
				fmt.Fprintf(out, "cannot setup bridge %s: %s\n", args[0], err)
				return
			}
			fmt.Fprintf(out, "success: %s set up\n", args[0])

		case "down":
			args, err := ScanN(s, 1)
			if err != nil {
				fmt.Fprintf(out, "invalid arguments: %s\n", err)
				return
			}
			if err = be.Down(args[0]); err != nil {
				fmt.Fprintf(out, "cannot destroy bridge %s: %s\n", args[0], err)
				return
			}
			fmt.Fprintf(out, "success: %s put down\n", args[0])

		case "link":
			args, err := ScanN(s, 2)
			if err != nil {
				fmt.Fprintf(out, "invalid arguments: %s\n", err)
				return
			}
			if err = be.Link(args[0], args[1]); err != nil {
				fmt.Fprintf(out, "cannot link %s to %s: %s\n", args[1], args[0], err)
				return
			}
			fmt.Fprintf(out, "success: %s linked to %s\n", args[1], args[0])

		case "unlink":
			args, err := ScanN(s, 2)
			if err != nil {
				fmt.Fprintf(out, "invalid arguments: %s\n", err)
				return
			}
			if err = be.Unlink(args[0], args[1]); err != nil {
				fmt.Fprintf(out, "cannot unlink %s from %s: %s\n", args[0], args[1], err)
				return
			}
			fmt.Fprintf(out, "success: %s unlinked from %s\n", args[0], args[1])

		default:
			fmt.Fprintf(out, "invalid command: %s\n", cmd)
			return
		}
	}
}

func ScanN(s *bufio.Scanner, n int) ([]string, error) {
	r := make([]string, 0, n)
	for n > 0 {
		if !s.Scan() {
			break
		}
		n--
		r = append(r, s.Text())
	}
	return r, s.Err()
}

type Local struct {
	io.ReadWriteCloser
}

func (*Local) Up(ctx context.Context, br string) error {
	return Up(ctx, br)
}

func (*Local) Down(br string) error {
	return Down(br)
}

func (*Local) Link(br string, ifce string) error {
	return Link(br, ifce)
}

func (*Local) Unlink(br string, ifce string) error {
	return Unlink(br, ifce)
}

func (*Local) Start(ctx context.Context) error {
	return Start(ctx)
}

func (*Local) Stop() error {
	return errors.New("not supported on local backend")
}

type Remote struct {
	io.ReadWriteCloser
	Conn io.ReadWriteCloser
}

func (r *Remote) Up(ctx context.Context, br string) error {
	fmt.Fprintf(r.Conn, "up %s\n", br)
	return r.Err()
}

func (r *Remote) Down(br string) error {
	fmt.Fprintf(r.Conn, "down %s\n", br)
	return r.Err()
}

func (r *Remote) Link(br string, ifce string) error {
	fmt.Fprintf(r.Conn, "link %s %s\n", br, ifce)
	return r.Err()
}

func (r *Remote) Unlink(br string, ifce string) error {
	fmt.Fprintf(r.Conn, "unlink %s %s\n", br, ifce)
	return r.Err()
}

func (r *Remote) Start(ctx context.Context) error {
	return errors.New("not supported on remote backend")
}

func (r *Remote) Stop() error {
	fmt.Fprintf(r.Conn, "stop\n")
	return r.Err()
}

func (r *Remote) Err() error {
	var data [1024]byte
	var res string

	for {
		n, err := r.Conn.Read(data[:])
		if err != nil {
			return err
		}
		res += string(data[0:n])

		if n < 1024 {
			break
		}
	}
	res = strings.TrimRight(res, "\n")
	if !strings.HasPrefix(res, "success") {
		v := strings.SplitN(res, ":", 2)
		res = strings.TrimSpace(v[len(v) - 1])
		return errors.New(res)
	}
	return nil
}

type Daemon struct {
	Local
	io.Closer
}

func (d *Daemon) Stop() error {
	return d.Closer.Close()
}

func (d *Daemon) Close() error {
	return d.ReadWriteCloser.Close()
}

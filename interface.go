package eth


type Interface interface {
	Name() string
	Index() int
	MAC() MAC
	Send(addr MAC, p []byte) (err error)
	Recv(p []byte) (n int, addr MAC, err error)
	Close() error
}

func Close(v interface{ Close() error }) {
	v.Close()
}

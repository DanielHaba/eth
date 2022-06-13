package eth

import (
	"net"
	"runtime"

	"golang.org/x/sys/unix"
)


func Open(name string) (ifce Interface, err error) {
	i, err := net.InterfaceByName(name)
	if err != nil {
		return
	}

	in, err := open(i)
	if err != nil {
		return
	}
	ifce = in

	// return

	out, err := open(i)
	if err != nil {
		return
	}

	if err = in.SetPromisc(true); err != nil {
		return
	}

	sock := &BiSocket{Socket: in, Output: out}
	runtime.SetFinalizer(sock, Close)
	ifce = sock

	return
}

type BiSocket struct {
	*Socket
	Output *Socket
}

func (sock *BiSocket) Send(addr MAC, p []byte) (err error) {
	return sock.Output.Send(addr, p)
}

func (sock *BiSocket) Close() error {
	err1 := sock.Socket.Close()
	err2 := sock.Output.Close()
	if err1 != nil {
		return err1
	}
	return err2
}


type Socket struct {
	ifce *net.Interface
	fd   int
}

func open(ifce *net.Interface) (sock *Socket, err error) {
	fd, err := unix.Socket(
		unix.AF_PACKET,
		unix.SOCK_RAW,
		0x0300,
	)
	if err != nil {
		return
	}
	if err = unix.BindToDevice(fd, ifce.Name); err != nil {
		unix.Close(fd)
		return
	}
	sock = &Socket{fd: fd, ifce: ifce}
	runtime.SetFinalizer(sock, Close)

	return
}

func (sock *Socket) MAC() MAC {
	addr := sock.ifce.HardwareAddr
	return MAC{addr[0], addr[1], addr[2], addr[3], addr[4], addr[5]}
}

func (sock *Socket) Recv(p []byte) (n int, from MAC, err error) {
	var addr unix.Sockaddr
	n, _, _, addr, err = unix.Recvmsg(sock.fd, p, []byte{}, 0)
	if err != nil {
		return
	}
	laddr := addr.(*unix.SockaddrLinklayer)
	if laddr.Pkttype&unix.PACKET_OUTGOING != 0 {
		err = unix.ENOMSG
		return
	}
	copy(from[:], laddr.Addr[0:6])

	return
}

func (sock *Socket) Send(addr MAC, p []byte) (err error) {
	err = unix.Sendto(sock.fd, p, 0, &unix.SockaddrLinklayer{
		Ifindex: sock.ifce.Index,
		Halen:   6,
		Addr:    [8]byte{addr[0], addr[1], addr[2], addr[3], addr[4], addr[5]},
	})
	return
}

func (sock *Socket) Close() error {
	return unix.Close(sock.fd)
}

func (sock *Socket) SetPromisc(f bool) (err error) {
	ifreq, err := unix.NewIfreq(sock.ifce.Name)
	if err = unix.IoctlIfreq(sock.fd, unix.SIOCGIFFLAGS, ifreq); err != nil {
		return
	}
	if f {
		ifreq.SetUint32(ifreq.Uint32() | unix.IFF_PROMISC)
	} else {
		ifreq.SetUint32(ifreq.Uint32() & ^uint32(unix.IFF_PROMISC))
	}
	if err = unix.IoctlIfreq(sock.fd, unix.SIOCGIFFLAGS, ifreq); err != nil {
		return
	}
	return
}

func (sock *Socket) Name() string {
	return sock.ifce.Name
}

func (sock *Socket) Index() int {
	return sock.ifce.Index
}
package eth

import (
	"fmt"
	"unsafe"
)

const (
	SizeHeader     = 14
	SizeVLANHeader = 18
)

type Header struct {
	Destination MAC
	Source      MAC
	Type        EtherType
	VID         uint16
	VType       EtherType
}

func (header *Header) String() string {
	typ := header.Type
	vlan := ""
	if typ == VLAN {
		typ = header.VType
		vlan = fmt.Sprintf("VLAN(%d) ", header.VID)
	}

	return fmt.Sprintf("{%s%s Destination: %s Source: %s}", vlan, typ, header.Destination, header.Source)
}

type Frame struct {
	Origin Interface
	*Header
	Data    []byte
	Payload unsafe.Pointer
}

func NewFrame(org Interface, data []byte) *Frame {
	header := (*Header)(unsafe.Pointer(&data[0]))
	offset := SizeHeader
	if header.Type == VLAN {
		offset = SizeVLANHeader
	}
	payload := unsafe.Pointer(&data[offset])

	return &Frame{
		Origin:  org,
		Data:    data,
		Header:  header,
		Payload: payload,
	}
}

func (frame *Frame) Copy() *Frame {
	data := make([]byte, len(frame.Data))
	copy(data, frame.Data)
	return NewFrame(frame.Origin, data)
}

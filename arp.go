package eth

import (
	"fmt"
	"unsafe"
)

type ARPOp uint16

const (
	ARPRequest ARPOp = 0x0100
	ARPReply   ARPOp = 0x0200
)

type ARPPayload struct {
	HType uint16
	PType EtherType
	HLen  uint8
	PLen  uint8
	Op    ARPOp
	SHA   MAC
	SPA   IP4
	THA   MAC
	TPA   IP4
}

func (arp *ARPPayload) String() string {
	if arp.Op == ARPRequest {
		return fmt.Sprintf("request who has %s, tell %s", arp.TPA, arp.SPA)
	}
	return fmt.Sprintf("reply %s is at %s", arp.TPA, arp.THA)
}

func (arp *ARPPayload) raw() []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(arp)), szArpPayload)
}

var arpOpName = map[ARPOp]string {
	ARPReply: "reply",
	ARPRequest: "request",
}

const szArpPayload = 28

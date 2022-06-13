package eth

import "fmt"

type EtherType uint16

const (
	IPv4 EtherType = 0x0008
	IPv6 EtherType = 0xDD86
	VLAN EtherType = 0x0018
	ARP  EtherType = 0x0608
)

func (t EtherType) String() string {
	if n, ok := ethNames[t]; ok {
		return n
	}
	return fmt.Sprintf("0x%04x", uint16(t))
}

var szEth = map[EtherType]int{
	ARP: 26,
}

var ethNames = map[EtherType]string{
	VLAN: "VLAN",
	IPv4: "IPv4",
	IPv6: "IPv6",
	ARP:  "ARP",
}

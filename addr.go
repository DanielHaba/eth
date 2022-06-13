package eth

import "fmt"

const (
	SizeIP  = 4
	SIZEMAC = 6
)

type IP4 [4]byte

func (ip IP4) String() string {
	return fmt.Sprintf("%d.%d.%d.%d", ip[0], ip[1], ip[2], ip[3])
}

type MAC [6]byte

func (mac MAC) Equals(v MAC) bool {
	return mac[0] == v[0] &&
		mac[1] == v[1] &&
		mac[2] == v[2] &&
		mac[3] == v[3] &&
		mac[4] == v[4] &&
		mac[5] == v[5]
}

func (mac MAC) IsZero() bool {
	return mac[0] == 0 && mac[1] == 0 && mac[2] == 0 && mac[3] == 0 && mac[4] == 0 && mac[5] == 0
}

func (mac MAC) IsBroadcast() bool {
	return mac[0] == 0xFF && mac[1] == 0xFF && mac[2] == 0xFF && mac[3] == 0xFF && mac[4] == 0xFF && mac[5] == 0xFF
}

func (mac MAC) String() string {
	return string([]byte{
		hexes[mac[0]>>4], hexes[mac[0]&0xF], ':',
		hexes[mac[1]>>4], hexes[mac[1]&0xF], ':',
		hexes[mac[2]>>4], hexes[mac[2]&0xF], ':',
		hexes[mac[3]>>4], hexes[mac[3]&0xF], ':',
		hexes[mac[4]>>4], hexes[mac[4]&0xF], ':',
		hexes[mac[5]>>4], hexes[mac[5]&0xF],
	})
}

func (mac MAC) Uint64() uint64 {
	return uint64(mac[5])<<40 |
		uint64(mac[4])<<32 |
		uint64(mac[3])<<24 |
		uint64(mac[2])<<16 |
		uint64(mac[1])<<8 |
		uint64(mac[0])
}

const hexes = "0123456789ABCDEF"

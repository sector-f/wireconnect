package server

import (
	"math/bits"
	"net"
)

func cidr(mask net.IPMask) uint {
	var cidrmask uint

	for _, b := range mask {
		cidrmask += uint(bits.OnesCount(uint(b)))
	}

	return cidrmask
}

package server

import (
	"net"
	"strings"

	"github.com/sector-f/wireconnect"
)

func CidrList(s string) ([]wireconnect.Address, error) {
	addresses := []wireconnect.Address{}

	for _, addr := range strings.Split(s, ",") {
		ip, net, err := net.ParseCIDR(addr)
		if err != nil {
			return nil, err
		}
		addresses = append(addresses, wireconnect.Address{Address: ip, Mask: net.Mask})
	}

	return addresses, nil
}

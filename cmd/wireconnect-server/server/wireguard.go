package server

import (
	"net"
	// "github.com/WireGuard/wgctrl-go"
)

func (s *Server) makeIface(addrString string) error {
	_, _, err := net.ParseCIDR(addrString)
	if err != nil {
		return err
	}

	return nil
}

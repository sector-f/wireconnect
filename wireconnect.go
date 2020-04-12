package wireconnect

import (
	"fmt"
	"math/bits"
	"net"
)

type ConnectionRequest struct {
	PeerName  string `json:"peer_name"`
	PublicKey string `json:"public_key"`
}

type ConnectionReply struct {
	PublicKey     string `json:"public_key"`
	ClientAddress string `json:"client_address"`
}

type CreatePeerRequest struct {
	Name            string `json:"name"`
	Address         string `json:"address"`
	ServerInterface string `json:"server_interface"`
}

type DisconnectionRequest struct {
	Peers []string `json:"peers"`
}

type BanList struct {
	Addresses []string
}

type Address struct {
	Address net.IP
	Mask    net.IPMask
}

func (a Address) String() string {
	var cidrmask uint

	for _, b := range a.Mask {
		cidrmask += uint(bits.OnesCount(uint(b)))
	}

	return fmt.Sprintf("%v/%v", a.Address, cidrmask)
}

func ParseAddress(s string) (Address, error) {
	ip, net, err := net.ParseCIDR(s)
	if err != nil {
		return Address{}, err
	}

	return Address{ip, net.Mask}, nil
}

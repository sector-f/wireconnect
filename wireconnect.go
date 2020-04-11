package wireconnect

import (
	"net"
)

type Address struct {
	Address net.IP
	Mask    net.IPMask
}

type Request struct {
	PublicKey string `json:"public_key"`
}

type Reply struct {
	PublicKey     string `json:"public_key"`
	ClientAddress string `json:"client_address"`
	ServerAddress string `json:"server_address"`
}

type BanList struct {
	Addresses []string
}

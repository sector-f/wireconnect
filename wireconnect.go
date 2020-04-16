package wireconnect

import (
	"encoding/json"
	"fmt"
	"math/bits"
	"net"
	"net/http"
)

var (
	DatabaseError      = ErrorResponse{http.StatusInternalServerError, "Database error"}
	ParseJsonError     = ErrorResponse{http.StatusBadRequest, "Improperly-formed request body"}
	IncompleteReqError = ErrorResponse{http.StatusBadRequest, "Incomplete request"}
)

type SuccessResponse struct {
	Status  int
	Payload interface{}
}

type ErrorResponse struct {
	Status  int
	Message string
}

func (e ErrorResponse) Error() string {
	return e.Message
}

type ServerInterface struct {
	Name      string    `json:"name"`
	Addresses []Address `json:"addresses"` // TODO: Maybe change this to []string?
}

func (s ServerInterface) MarshalJSON() ([]byte, error) {
	retAddr := []string{}
	for _, addr := range s.Addresses {
		retAddr = append(retAddr, addr.String())
	}

	retVal := struct {
		Name      string   `json:"name"`
		Addresses []string `json:"addresses"`
	}{s.Name, retAddr}

	return json.Marshal(&retVal)
}

type ConnectionRequest struct {
	PeerName  string `json:"peer_name"`
	PublicKey string `json:"public_key"`
}

type ConnectionReply struct {
	PublicKey     string `json:"public_key"`
	ClientAddress string `json:"client_address"`
}

type CreatePeerRequest struct {
	UserName        string `json:"user_name"`
	PeerName        string `json:"peer_name"`
	Address         string `json:"address"`
	ServerInterface string `json:"server_interface"`
}

type CreateUserRequest struct {
	UserName string `json:"user_name"`
	Password string `json:"password"`
	IsAdmin  bool   `json:"is_admin"`
}

type DisconnectionRequest struct {
	PeerName string `json:"peer_name"`
}

type BanList struct {
	Addresses []string
}

type Address struct {
	Address net.IP     `json:"address"`
	Mask    net.IPMask `json:"mask"`
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

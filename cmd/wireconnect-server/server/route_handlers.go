package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sector-f/wireconnect"
	"github.com/sector-f/wireconnect/cmd/wireconnect-server/database"
)

var limiter = NewLimiter()

func (s *Server) createPeerHandler(r *http.Request) (*wireconnect.SuccessResponse, error) {
	jsonDecoder := json.NewDecoder(r.Body)

	request := wireconnect.CreatePeerRequest{}
	err := jsonDecoder.Decode(&request)
	if err != nil {
		return nil, wireconnect.ParseJsonError
	}

	if request.UserName == "" || request.PeerName == "" || request.Address == "" || request.ServerInterface == "" {
		return nil, wireconnect.IncompleteReqError
	}

	err = s.db.CreatePeer(request)
	if err != nil {
		return nil, wireconnect.DatabaseError
	}

	return &wireconnect.SuccessResponse{http.StatusCreated, fmt.Sprintf("Created peer: %s\n", request.PeerName)}, nil
}

func (s *Server) getInterfacesHandler(r *http.Request) (*wireconnect.SuccessResponse, error) {
	interfaces, err := s.db.Ifaces()
	if err != nil {
		return nil, wireconnect.DatabaseError
	}

	wireIfaces := []wireconnect.ServerInterface{}
	for _, iface := range interfaces {
		wireIfaces = append(
			wireIfaces,
			wireconnect.ServerInterface{
				Name:      iface.Name,
				Addresses: iface.Addresses,
			},
		)
	}

	return &wireconnect.SuccessResponse{http.StatusOK, wireIfaces}, nil
}

func (s *Server) connectHandler(r *http.Request) (*wireconnect.SuccessResponse, error) {
	jsonDecoder := json.NewDecoder(r.Body)

	username, _, _ := r.BasicAuth()

	request := wireconnect.ConnectionRequest{}
	err := jsonDecoder.Decode(&request)
	if err != nil {
		return nil, wireconnect.ParseJsonError
	}

	if request.PeerName == "" || request.PublicKey == "" {
		return nil, wireconnect.IncompleteReqError
	}

	peer := s.db.GetPeer(username, request.PeerName)
	if peer == nil {
		return nil, wireconnect.ErrorResponse{http.StatusNotFound, "No peer with that name exists"}
	}

	err = s.makeIface(peer.DBIface)
	if err != nil {
		return nil, wireconnect.DatabaseError
	}

	err = s.addPeer(username, request)
	if err != nil {
		return nil, wireconnect.DatabaseError
	}

	wgDev, _ := s.wgClient.Device(peer.DBIface.Name)
	resp := wireconnect.ConnectionReply{
		PublicKey:     wgDev.PublicKey.String(),
		ClientAddress: peer.Address.String(),
	}

	return &wireconnect.SuccessResponse{http.StatusOK, resp}, nil
}

func (s *Server) getBansHandler(r *http.Request) (*wireconnect.SuccessResponse, error) {
	bans := wireconnect.BanList{limiter.getBans()}
	return &wireconnect.SuccessResponse{http.StatusOK, bans}, nil
}

func (s *Server) disconnectHandler(r *http.Request) (*wireconnect.SuccessResponse, error) {
	jsonDecoder := json.NewDecoder(r.Body)

	username, _, _ := r.BasicAuth()

	request := wireconnect.DisconnectionRequest{}
	err := jsonDecoder.Decode(&request)
	if err != nil {
		return nil, wireconnect.ParseJsonError
	}

	if request.PeerName == "" {
		return nil, wireconnect.IncompleteReqError
	}

	err = s.removePeer(username, request.PeerName)
	if err != nil {
		return nil, wireconnect.IncompleteReqError
	}

	return &wireconnect.SuccessResponse{http.StatusOK, fmt.Sprintf("Disconnected peer: %s\n", request.PeerName)}, nil
}

func (s *Server) addUserHandler(r *http.Request) (*wireconnect.SuccessResponse, error) {
	jsonDecoder := json.NewDecoder(r.Body)

	request := wireconnect.CreateUserRequest{}
	err := jsonDecoder.Decode(&request)
	if err != nil {
		return nil, wireconnect.ParseJsonError
	}

	if request.UserName == "" || request.Password == "" {
		return nil, wireconnect.IncompleteReqError
	}

	err = s.db.AddUser(database.User{
		Username: request.UserName,
		Password: []byte(request.Password),
		IsAdmin:  request.IsAdmin,
	})
	if err != nil {
		return nil, wireconnect.DatabaseError
	}

	return &wireconnect.SuccessResponse{http.StatusCreated, "User created"}, nil
}

package main

import (
	// "encoding/json"
	"log"
	"net/http"

	// "crypto/tls"

	// "github.com/gorilla/mux"
	"github.com/sector-f/wireconnect"
	"github.com/sector-f/wireconnect/cmd/wireconnect-server/server"
	// "github.com/WireGuard/wgctrl-go"
	// "github.com/juju/ratelimit"
)

var auth wireconnect.AuthProvider = mockAuth{}

type Route struct {
	Pattern    string
	Method     string
	Handler    http.HandlerFunc
	Permission string
}

func main() {
	config := server.NewConfig()
	server := server.NewServer(config)
	log.Fatal(server.ListenAndServe())
}

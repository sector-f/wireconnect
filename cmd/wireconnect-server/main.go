package main

import (
	"log"
	"net/http"

	"github.com/sector-f/wireconnect"
	"github.com/sector-f/wireconnect/cmd/wireconnect-server/server"
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
	config.DSN = "file:./wireconnect.sqlite"

	server, err := server.NewServer(config)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Listening on %s\n", config.Address)
	log.Fatal(server.ListenAndServe())
}

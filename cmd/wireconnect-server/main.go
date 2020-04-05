package main

import (
	"log"

	"github.com/sector-f/wireconnect/cmd/wireconnect-server/server"
)

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

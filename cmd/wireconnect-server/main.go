package main

import (
	"log"

	"github.com/sector-f/wireconnect/cmd/wireconnect-server/server"
	flag "github.com/spf13/pflag"
)

func main() {
	keyfile := flag.StringP("key", "k", "", "Path to keyfile")
	certfile := flag.StringP("cert", "c", "", "Path to certfile")
	flag.Parse()

	if *keyfile == "" || *certfile == "" {
		log.Fatalln("Key and cert must be specified")
	}

	config := server.NewConfig()
	config.DSN = "file:./wireconnect.sqlite"

	server, err := server.NewServer(config)
	if err != nil {
		server.Shutdown()
		log.Fatal(err)
	}

	log.Printf("Listening on %s\n", config.Address)
	err = server.ListenAndServeTLS(*certfile, *keyfile)
	if err != nil {
		server.Shutdown()
		log.Fatal(err)
	}
}

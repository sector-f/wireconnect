package main

import (
	"errors"
	"log"

	"github.com/sector-f/wireconnect/cmd/wireconnect-server/server"
	flag "github.com/spf13/pflag"
)

func main() {
	flag.ErrHelp = errors.New("Help requested")

	keyfile := flag.StringP("key", "k", "", "Path to keyfile")
	certfile := flag.StringP("cert", "c", "", "Path to certfile")
	dbfile := flag.StringP("database", "d", "file:/var/local/wireconnect.sqlite", "SQLite DSN for wireconnect database")
	flag.Parse()

	if *keyfile == "" || *certfile == "" {
		log.Fatalln("Key and cert must be specified")
	}

	config := server.NewConfig()
	config.DSN = *dbfile
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

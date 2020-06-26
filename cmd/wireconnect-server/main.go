package main

import (
	"crypto/tls"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/sector-f/wireconnect/cmd/wireconnect-server/reloadablecert"
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
	wcServer, err := server.NewServer(config)
	if err != nil {
		wcServer.Shutdown()
		log.Fatal(err)
	}

	cert, err := reloadablecert.New(*certfile, *keyfile)
	if err != nil {
		log.Fatal(err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGUSR1)

	go func() {
		for _ = range sigChan {
			err := cert.Reload()
			if err != nil {
				log.Printf("Failed to reload TLS key/certificate: %v\n", err)
			} else {
				log.Println("Reloaded TLS key/certificate")
			}
		}
	}()

	tlsConfig := &tls.Config{
		GetCertificate:           cert.GetCertificate,
		PreferServerCipherSuites: true,
		MinVersion:               tls.VersionTLS12,
	}

	listener, err := tls.Listen("tcp", config.Address, tlsConfig)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Listening on %s\n", config.Address)
	err = wcServer.Serve(listener)
	if err != nil {
		wcServer.Shutdown()
		log.Fatal(err)
	}
}

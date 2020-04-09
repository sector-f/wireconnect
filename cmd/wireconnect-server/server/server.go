package server

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	// "github.com/WireGuard/wgctrl-go"
)

type route struct {
	pattern  string
	handlers []handler
}

type handler struct {
	method      string
	handlerFunc http.HandlerFunc
	needsAdmin  bool
}

type Config struct {
	Address      string
	DSN          string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func NewConfig() Config {
	return Config{
		Address:      "0.0.0.0:8080",
		DSN:          "file:/var/local/wireconnect.sqlite",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
}

type Server struct {
	db *sql.DB
	*http.Server
}

func NewServer(conf Config) (*Server, error) {
	db, err := sql.Open("sqlite3", conf.DSN)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	httpServer := &http.Server{
		Addr:         conf.Address,
		ReadTimeout:  conf.ReadTimeout,
		WriteTimeout: conf.WriteTimeout,
	}
	server := Server{db, httpServer}
	err = server.initDB()
	if err != nil {
		return nil, err
	}

	userCount, err := server.userCount()
	if err != nil {
		return nil, err
	}
	if userCount == 0 {
		server.makeFirstUser()
	}

	ifaceCount, err := server.ifaceCount()
	if err != nil {
		return nil, err
	}
	if ifaceCount == 0 {
		server.makeFirstIface()
	}

	ifaces, err := server.ifaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		log.Printf("Creating interface %v\n", iface.Name)
		for _, addr := range iface.Addresses {
			log.Printf("\t%v/%v\n", addr.Address, cidr(addr.Mask))
		}
	}

	router := mux.NewRouter()
	routes := []route{
		route{
			pattern: "/bans",
			handlers: []handler{
				handler{
					method:      "GET",
					handlerFunc: server.getBansHandler,
					needsAdmin:  true,
				},
			},
		},
		route{
			pattern: "/connect",
			handlers: []handler{
				handler{
					method:      "POST",
					handlerFunc: server.connectHandler,
					needsAdmin:  false,
				},
			},
		},
		route{
			pattern: "/disconnect",
			handlers: []handler{
				handler{
					method:      "POST",
					handlerFunc: server.disconnectHandler,
					needsAdmin:  false,
				},
			},
		},
	}

	for _, route := range routes {
		methodHandler := make(handlers.MethodHandler)
		for _, handler := range route.handlers {
			var h http.Handler = handler.handlerFunc

			if handler.needsAdmin {
				h = server.adminHandler(h)
			}

			h = server.authLimit(h)

			methodHandler[handler.method] = h

			if handler.method == "GET" {
				methodHandler["HEAD"] = h
			}
		}

		router.Path(route.pattern).Handler(methodHandler)
	}

	httpServer.Handler = handlers.LoggingHandler(os.Stdout, router)

	return &server, nil
}

package server

import (
	"database/sql"
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
	server.initDB()

	userCount, err := server.userCount()
	if err != nil {
		return nil, err
	}

	if userCount == 0 {
		server.makeUserInteractive()
	}

	router := mux.NewRouter()
	routes := []route{
		route{
			pattern: "/connect",
			handlers: []handler{
				handler{
					method:      "POST",
					handlerFunc: server.connectHandler,
				},
			},
		},
		route{
			pattern: "/disconnect",
			handlers: []handler{
				handler{
					method:      "POST",
					handlerFunc: server.disconnectHandler,
				},
			},
		},
	}

	for _, route := range routes {
		methodHandler := make(handlers.MethodHandler)
		for _, handler := range route.handlers {
			h := server.authLimit(handler.handlerFunc)
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

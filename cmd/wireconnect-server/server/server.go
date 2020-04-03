package server

import (
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type route struct {
	pattern  string
	handlers []handler
}

type handler struct {
	method      string
	handlerFunc http.HandlerFunc

	permission string
}

type Config struct {
	Address      string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func NewConfig() Config {
	return Config{
		Address:      "0.0.0.0:8080",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
}

type Server struct {
	*http.Server
}

func NewServer(conf Config) Server {
	routes := []route{
		route{
			pattern: "/connect",
			handlers: []handler{
				handler{
					method:      "POST",
					handlerFunc: connectHandler,
				},
			},
		},
	}

	router := mux.NewRouter()
	for _, route := range routes {
		methodHandler := make(handlers.MethodHandler)
		for _, handler := range route.handlers {
			methodHandler[handler.method] = handler.handlerFunc

			if handler.method == "GET" {
				methodHandler["HEAD"] = handler.handlerFunc
			}
		}

		router.Path(route.pattern).Handler(methodHandler)
	}

	server := &http.Server{
		Handler:      handlers.LoggingHandler(os.Stdout, router),
		Addr:         conf.Address,
		ReadTimeout:  conf.ReadTimeout,
		WriteTimeout: conf.WriteTimeout,
	}

	return Server{server}
}

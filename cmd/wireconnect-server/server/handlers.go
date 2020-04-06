package server

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
)

var limiter = NewLimiter()

func (s *Server) connectHandler(w http.ResponseWriter, r *http.Request) {
	username, _, _ := r.BasicAuth()
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, fmt.Sprintf("Successfully authenticated as %s\n", username))
}

func (s *Server) disconnectHandler(w http.ResponseWriter, r *http.Request) {
	username, _, _ := r.BasicAuth()
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, fmt.Sprintf("Successfully authenticated as %s\n", username))
}

func (s *Server) getBansHandler(w http.ResponseWriter, r *http.Request) {
	bans := limiter.getBans()
	for _, ban := range bans {
		io.WriteString(w, ban+"\n")
	}
}

func (s *Server) adminHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, _, _ := r.BasicAuth()

		isAdmin, err := s.isAdmin(username)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "Internal server error\n")
			return
		}

		if !isAdmin {
			w.WriteHeader(http.StatusUnauthorized)
			io.WriteString(w, "Only administrators can access this resource\n")
			return
		}

		h.ServeHTTP(w, r)
	})
}

func (s *Server) authLimit(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var sourceAddr string
		if strings.Contains(r.RemoteAddr, ":") {
			sourceAddr, _, _ = net.SplitHostPort(r.RemoteAddr)
		} else {
			sourceAddr = r.RemoteAddr
		}

		bucket := limiter.getIP(sourceAddr)
		if bucket.TakeAvailable(1) == 0 {
			w.WriteHeader(http.StatusTooManyRequests)
			io.WriteString(w, "Rate limit has been reached\n")
			return
		}

		username, password, ok := r.BasicAuth()
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			io.WriteString(w, "Authentication required\n")
			return
		}

		err := s.authenticate(username, password)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			io.WriteString(w, "Bad username or password\n")
			return
		}

		limiter.delIP(sourceAddr)

		h.ServeHTTP(w, r)
	})
}

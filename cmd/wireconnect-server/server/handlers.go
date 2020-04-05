package server

import (
	"fmt"
	"io"
	"net/http"
)

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

func (s *Server) authLimit(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		h.ServeHTTP(w, r)
	})
}

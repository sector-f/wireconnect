package server

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"

	"github.com/sector-f/wireconnect"
)

type apiFunc = func(*http.Request) (*wireconnect.SuccessResponse, error)

func jsonHandler(internal apiFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		returnVal, err := internal(r)
		if err != nil {
			switch val := err.(type) {
			case wireconnect.ErrorResponse:
				w.WriteHeader(val.Status)

				json, err := json.MarshalIndent(val, "", "  ")
				if err != nil {
					w.Write([]byte(val.Message))
					return
				}

				w.Write(append(json, byte('\n')))
				return
			default:
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		json, err := json.MarshalIndent(returnVal.Payload, "", "  ")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(returnVal.Status)
		w.Write(append(json, byte('\n')))
	})
}

func (s *Server) adminHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, _, _ := r.BasicAuth()

		isAdmin, err := s.db.IsAdmin(username)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			// TODO: replace with json
			// io.WriteString(w, "Internal server error\n")
			return
		}

		if !isAdmin {
			w.WriteHeader(http.StatusUnauthorized)
			// TODO: replace with json
			// io.WriteString(w, "Only administrators can access this resource\n")
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
			// io.WriteString(w, "Rate limit has been reached\n")
			return
		}

		username, password, ok := r.BasicAuth()
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			// io.WriteString(w, "Authentication required\n")
			return
		}

		err := s.db.Authenticate(username, password)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			// io.WriteString(w, "Bad username or password\n")
			return
		}

		limiter.delIP(sourceAddr)

		h.ServeHTTP(w, r)
	})
}

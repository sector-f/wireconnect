package server

import (
	"fmt"
	"io"
	"net/http"
)

func connectHandler(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, "Authentication required\n")
		return
	}

	w.WriteHeader(http.StatusOK)
	io.WriteString(w, fmt.Sprintf("Username: %s\n", username))
	io.WriteString(w, fmt.Sprintf("Password: %s\n", password))
}

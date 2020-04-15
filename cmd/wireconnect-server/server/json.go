package server

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Status  int
	Message string
}

func (e ErrorResponse) Error() string {
	return e.Message
}

func jsonHandler(internal func(*http.Request) (interface{}, error)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		returnVal, err := internal(r)
		if err != nil {
			switch val := err.(type) {
			case ErrorResponse:
				w.WriteHeader(val.Status)

				json, err := json.Marshal(val)
				if err != nil {
					w.Write([]byte(val.Message))
					return
				}

				w.Write(json)
				return
			default:
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		json, err := json.Marshal(returnVal)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write(json)
	})
}

func jsonUsernameTest(r *http.Request) (interface{}, error) {
	username, _, ok := r.BasicAuth()
	if !ok {
		return nil, ErrorResponse{http.StatusTeapot, "Basic Auth not used"}
	}

	return username, nil
}

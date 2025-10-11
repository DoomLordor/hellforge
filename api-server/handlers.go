package apiserver

import (
	"io"
	"net/http"
)

func livenessCheckHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, `{"live": true}`)
}

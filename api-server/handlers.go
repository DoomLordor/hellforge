package apiserver

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/DoomLordor/hellforge/logger"
)

type logLevel struct {
	ModuleName string `json:"module_name"`
	LogLevel   string `json:"log_level"`
}

func setLogLevel(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	ll := &logLevel{}
	err := json.NewDecoder(r.Body).Decode(ll)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, `{"status": "ok"}`)
		return
	}

	if ll.ModuleName == "" {
		ll.ModuleName = logger.BaseLoggerName
	}

	if ll.LogLevel == "" {
		ll.LogLevel = "info"
	}

	logger.SetLevel(ll.ModuleName, ll.LogLevel)

	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, `{"status": "ok"}`)
}

func livenessCheckHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, `{"live": true}`)
}

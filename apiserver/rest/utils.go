package rest

import (
	"context"
	"io"
	"net/http"
)

const (
	UserKey = "user"
	bearer  = "Bearer "
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type WsResponse struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

type AuthFunc func(ctx context.Context, token string) (any, error)

func notFound(w http.ResponseWriter, _ *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	_, _ = io.WriteString(w, `{"error": "url not found"}`)
}

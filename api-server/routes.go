package apiserver

import (
	"net/http"
)

type HandlerApiFunc func(*http.Request) (any, int, error)

type RouteApi struct {
	Methods     []string
	Pattern     string
	Metrics     bool
	Version     uint8
	ContentType string
	DataType    DataType
	Middlewares func(http.Handler) http.Handler
	HandlerFunc HandlerApiFunc
}

type RoutesApi []*RouteApi
type RouteApiMap map[string]RoutesApi

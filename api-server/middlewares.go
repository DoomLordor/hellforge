package apiserver

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/DoomLordor/hellforge/logger"
)

type middlewares struct {
	logger  zerolog.Logger
	tracer  trace.Tracer
	metrics *metrics
	erc     ErrorResponseConstructor
}

type metrics struct {
	requestCount  *prometheus.CounterVec
	responseCount *prometheus.CounterVec
	latency       *prometheus.HistogramVec
}

func newMiddlewares(tracer trace.Tracer, erc ErrorResponseConstructor) *middlewares {
	if erc == nil {
		erc = newErrorResponse
	}

	return &middlewares{
		logger: logger.NewLogger("middlewares"),
		tracer: tracer,
		erc:    erc,
		metrics: &metrics{
			requestCount: prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Name: "total_request",
					Help: "Total number of HTTP requests",
				},
				[]string{"path"},
			),
			responseCount: prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Name: "total_response",
					Help: "Total number of error HTTP requests",
				},
				[]string{"path", "code"},
			),
			latency: prometheus.NewHistogramVec(
				prometheus.HistogramOpts{
					Name:    "request_latency",
					Help:    "Response latency in seconds",
					Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 2},
				},
				[]string{"path"},
			),
		},
	}
}

func (m *middlewares) headersMiddleware(next http.Handler) http.Handler {
	return cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowedHeaders: []string{
			"Accept",
			"Content-Type",
			"Content-Length",
			"Accept-Encoding",
			"X-CSRF-Token",
			"Authorization",
			"Access-Control-Request-Headers",
			"Access-Control-Request-Method",
			"Connection",
			"Host",
			"Origin",
			"User-Agent",
			"Referer",
			"Cache-Control",
			"X-header",
		},
		AllowCredentials: false,
	}).Handler(next)
}

func (m *middlewares) recoveryMiddleware(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			rc := recover()
			if rc != nil {
				m.logger.Err(errors.New("panic")).
					Str("panic", fmt.Sprintf("%v", rc)).
					Msg(string(debug.Stack()))
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(f)
}

func (m *middlewares) tracingMiddleware(hf HandlerApiFunc) HandlerApiFunc {
	if m.tracer == nil {
		return hf
	}

	return func(r *http.Request) (any, int, error) {
		ctx, span := m.tracer.Start(r.Context(), r.URL.Path)
		defer span.End()

		res, code, err := hf(r.WithContext(ctx))
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "succeeded")
		}

		return res, code, err
	}
}

func (m *middlewares) loggingMiddleware(hf HandlerApiFunc) HandlerApiFunc {
	f := func(r *http.Request) (res any, code int, err error) {

		start := time.Now().UnixMilli()

		m.logger.Info().
			Str("method", r.Method).
			Str("url", r.RequestURI).
			Msg("Start")

		defer func() {
			end := time.Now().UnixMilli() - start
			event := m.logger.Info().
				Ctx(r.Context()).
				Str("method", r.Method).
				Str("url", r.RequestURI).
				Int("code", code).
				Int64("response_time", end)

			if err != nil {
				event = event.Str("error", err.Error())
			}

			event.Msg("End")
		}()

		return hf(r)
	}

	return f
}

func (m *middlewares) metricsMiddleware(path string, hf HandlerApiFunc) HandlerApiFunc {
	return func(r *http.Request) (res any, code int, err error) {
		m.metrics.requestCount.WithLabelValues(path).Inc()

		start := time.Now()

		defer func() {
			delta := time.Since(start).Seconds()
			m.metrics.latency.WithLabelValues(path).Observe(delta)

			if code >= http.StatusMultipleChoices {
				m.metrics.responseCount.WithLabelValues(path, strconv.Itoa(code)).Inc()
			}
		}()

		return hf(r)
	}
}

func (m *middlewares) handleApiWrap(hf HandlerApiFunc, contentType string, dataType DataType) http.Handler {
	if contentType == "" {
		contentType = ContentTypeJSON
	}

	f := func(w http.ResponseWriter, r *http.Request) {
		res, code, err := hf(r)
		if err != nil {
			res = m.erc(err)
			if code >= http.StatusInternalServerError {
				m.logger.Err(err).Str("method", r.Method).Str("url", r.RequestURI).Send()
			} else {
				m.logger.Warn().
					Str("method", r.Method).
					Str("url", r.RequestURI).
					Str("warning", err.Error()).
					Send()
			}
		}

		if res != nil {
			w.Header().Add("Content-Type", contentType)

			switch dataType {
			case DataTypeBytes:
				bytes, ok := res.([]byte)
				if !ok {
					_, _ = io.WriteString(w, `{"error": "error cast response to bytes"}`)
					return
				}
				_, _ = w.Write(bytes)
			case DataTypeString:
				str, ok := res.(string)
				if !ok {
					_, _ = io.WriteString(w, `{"error": "error cast response to string"}`)
					return
				}
				_, _ = io.WriteString(w, str)
			case DataTypeStructJSON:
				err = json.NewEncoder(w).Encode(res)
				if err != nil {
					m.logger.
						Err(err).
						Str("method", r.Method).
						Str("url", r.RequestURI).
						Msg("Error encoding json response")
					_, _ = io.WriteString(w, `{"error": "error encoding response"}`)
				}
			case DataTypeStructXML:
				err = xml.NewEncoder(w).Encode(res)
				if err != nil {
					m.logger.
						Err(err).
						Str("method", r.Method).
						Str("url", r.RequestURI).
						Msg("Error encoding xml response")
					_, _ = io.WriteString(w, `<error>error encoding response</error>`)
				}
			}
		}

		w.WriteHeader(code)
	}

	return http.HandlerFunc(f)
}

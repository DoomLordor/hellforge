package logger

import (
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

type tracingHook struct{}

func (h *tracingHook) Run(e *zerolog.Event, _ zerolog.Level, _ string) {
	span := trace.SpanFromContext(e.GetCtx())
	if span.SpanContext().IsValid() {
		traceID := span.SpanContext().TraceID().String()
		e.Str("trace-id", traceID)
	}
}

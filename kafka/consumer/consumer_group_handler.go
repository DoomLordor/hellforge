package consumer

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"

	"github.com/IBM/sarama"
	"github.com/rs/zerolog"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type consumerGroupHandler struct {
	logger  zerolog.Logger
	tracer  trace.Tracer
	handler Handler
}

func newHandler(logger zerolog.Logger, topics []string, handler Handler, tracer trace.Tracer) *consumerGroupHandler {
	h := &consumerGroupHandler{
		logger: logger.With().Strs("topics", topics).Logger(),
		tracer: tracer,
	}

	handler = h.withRecover(handler)
	handler = h.withTracing(handler)
	h.handler = handler
	return h
}

func (h *consumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) ConsumeClaim(
	session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim,
) error {
	done := session.Context().Done()
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				h.logger.Debug().Ctx(session.Context()).Msg("message channel was closed for topic")
				return nil
			}

			err := h.handler(session.Context(), message)
			if err != nil {
				h.logger.Err(err).Ctx(session.Context()).Msg("handler error")
			}

			session.MarkMessage(message, "")
			session.Commit()
		case <-done:
			return nil
		}
	}
}

func (h *consumerGroupHandler) withRecover(handler Handler) Handler {
	return func(ctx context.Context, msg *sarama.ConsumerMessage) error {
		defer func() {
			r := recover()
			if r != nil {
				h.logger.Err(errors.New("panic")).
					Ctx(ctx).
					Str("panic", fmt.Sprintf("%v", r)).
					Msgf("recovered stack: %s", string(debug.Stack()))
			}
		}()

		return handler(ctx, msg)
	}
}

func (h *consumerGroupHandler) withTracing(handler Handler) Handler {
	if h.tracer == nil {
		return handler
	}

	return func(ctx context.Context, msg *sarama.ConsumerMessage) error {
		ctx, span := h.tracer.Start(ctx, msg.Topic)
		defer span.End()

		err := handler(ctx, msg)
		if err != nil {
			span.SetStatus(otelcodes.Error, err.Error())
		} else {
			span.SetStatus(otelcodes.Ok, "succeeded")
		}

		return err
	}
}

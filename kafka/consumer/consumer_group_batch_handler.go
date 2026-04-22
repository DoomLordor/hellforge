package consumer

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type consumerGroupBatchHandler struct {
	logger       zerolog.Logger
	tracer       trace.Tracer
	handler      BatchHandler
	batchSize    int
	batchTimeout time.Duration
}

func newBatchHandler(logger zerolog.Logger, topics []string, handler BatchHandler, batchSize int, batchTimeout time.Duration, tracer trace.Tracer) *consumerGroupBatchHandler {
	h := &consumerGroupBatchHandler{
		logger:       logger.With().Strs("topics", topics).Logger(),
		tracer:       tracer,
		batchSize:    batchSize,
		batchTimeout: batchTimeout,
	}

	handler = h.withRecover(handler)
	handler = h.withTracing(handler, topics)
	h.handler = handler
	return h
}

func (h *consumerGroupBatchHandler) Setup(_ sarama.ConsumerGroupSession) error { return nil }

func (h *consumerGroupBatchHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *consumerGroupBatchHandler) ConsumeClaim(
	session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim,
) error {
	ctx := session.Context()
	done := ctx.Done()
	batch := make([]*sarama.ConsumerMessage, 0, h.batchSize)

	timer := time.NewTimer(h.batchTimeout)
	defer timer.Stop()

	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				h.logger.Debug().Ctx(ctx).Msg("message channel was closed for topic")
				return nil
			}

			batch = append(batch, message)

			if len(batch) >= h.batchSize {
				err := h.handler(ctx, batch)
				if err != nil {
					h.logger.Err(err).Ctx(ctx).Msg("handler error")
				}

				batch = h.commitBatch(session, batch, timer)
			}

		case <-timer.C:
			if len(batch) > 0 {
				err := h.handler(ctx, batch)
				if err != nil {
					h.logger.Err(err).Ctx(ctx).Msg("handler error")
				}

				batch = h.commitBatch(session, batch, timer)
			} else {
				timer.Reset(h.batchTimeout)
			}

		case <-done:
			return nil
		}
	}
}

func (h *consumerGroupBatchHandler) commitBatch(session sarama.ConsumerGroupSession, batch []*sarama.ConsumerMessage, timer *time.Timer) []*sarama.ConsumerMessage {
	for _, msg := range batch {
		session.MarkMessage(msg, "")
	}

	session.Commit()
	timer.Reset(h.batchTimeout)

	return batch[:0]
}

func (h *consumerGroupBatchHandler) withRecover(handler BatchHandler) BatchHandler {
	return func(ctx context.Context, msgs []*sarama.ConsumerMessage) error {
		if len(msgs) == 0 {
			return nil
		}

		defer func() {
			r := recover()
			if r != nil {
				h.logger.Err(errors.New("panic")).
					Ctx(ctx).
					Str("panic", fmt.Sprintf("%v", r)).
					Msgf("recovered stack: %s", string(debug.Stack()))
			}
		}()

		return handler(ctx, msgs)
	}
}

func (h *consumerGroupBatchHandler) withTracing(handler BatchHandler, topics []string) BatchHandler {
	if h.tracer == nil {
		return handler
	}

	spanName := strings.Join(topics, "; ")

	return func(ctx context.Context, msgs []*sarama.ConsumerMessage) error {
		if len(msgs) == 0 {
			return nil
		}

		ctx, span := h.tracer.Start(ctx, spanName) //TODO: think about span name
		defer span.End()

		span.SetAttributes(attribute.Int("count", len(msgs)))

		err := handler(ctx, msgs)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelcodes.Error, err.Error())
		} else {
			span.SetStatus(otelcodes.Ok, "success")
		}

		return err
	}
}

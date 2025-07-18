package consumer

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/IBM/sarama"

	"github.com/DoomLordor/hellforge/logger"
)

type consumerGroupBatchHandler struct {
	logger       *logger.Logger
	bh           BatchHandler
	batchSize    int
	batchTimeout time.Duration
}

func newBatchHandler(log *logger.Logger, bh BatchHandler, batchSize int, batchTimeout time.Duration) *consumerGroupBatchHandler {
	return &consumerGroupBatchHandler{
		logger:       log,
		bh:           batchHandlerRecover(bh),
		batchSize:    batchSize,
		batchTimeout: batchTimeout,
	}
}

func (h *consumerGroupBatchHandler) Setup(_ sarama.ConsumerGroupSession) error { return nil }

func (h *consumerGroupBatchHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *consumerGroupBatchHandler) ConsumeClaim(
	session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim,
) error {
	done := session.Context().Done()
	batch := make([]*sarama.ConsumerMessage, 0, h.batchSize)

	timer := time.NewTimer(h.batchTimeout)
	defer timer.Stop()

	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				h.logger.Debug().
					Ctx(session.Context()).
					Str("topic", claim.Topic()).
					Msg("message channel was closed for topic")
				return nil
			}

			batch = append(batch, message)

			if len(batch) >= h.batchSize {
				err := h.bh(session.Context(), batch)
				if err != nil {
					h.logger.Err(err).
						Ctx(session.Context()).
						Str("topic", message.Topic).
						Send()
				}

				batch = h.commitBatch(session, batch, timer)
			}

		case <-timer.C:
			if len(batch) > 0 {
				err := h.bh(session.Context(), batch)
				if err != nil {
					h.logger.Err(err).
						Ctx(session.Context()).
						Str("topic", batch[0].Topic).
						Send()
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

func batchHandlerRecover(bh BatchHandler) BatchHandler {
	return func(ctx context.Context, msgs []*sarama.ConsumerMessage) (err error) {
		if len(msgs) == 0 {
			return nil
		}

		defer func() {
			r := recover()
			if r != nil {
				err = fmt.Errorf("kafka batch handler recovered from panic: %s\n stack: %s", r, debug.Stack())
			}
		}()

		return bh(ctx, msgs)
	}
}

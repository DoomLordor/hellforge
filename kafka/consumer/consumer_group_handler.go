package consumer

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/IBM/sarama"

	"github.com/DoomLordor/hellforge/logger"
)

type consumerGroupHandler struct {
	logger *logger.Logger
	h      Handler
}

func newHandler(log *logger.Logger, h Handler) *consumerGroupHandler {
	return &consumerGroupHandler{
		logger: log,
		h:      handlerRecover(h),
	}
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
				h.logger.Debug().Ctx(session.Context()).Str("topic", claim.Topic()).Msg("message channel was closed for topic")
				return nil
			}

			err := h.h(session.Context(), message)
			if err != nil {
				h.logger.Err(err).
					Ctx(session.Context()).
					Str("topic", message.Topic).
					Send()
			}

			session.MarkMessage(message, "")
			session.Commit()
		case <-done:
			return nil
		}
	}
}

func handlerRecover(h Handler) Handler {
	return func(ctx context.Context, msg *sarama.ConsumerMessage) (err error) {
		defer func() {
			r := recover()
			if r != nil {
				err = fmt.Errorf("kafka handler recovered from panic: %s\n stack: %s", r, debug.Stack())
			}
		}()

		return h(ctx, msg)
	}
}

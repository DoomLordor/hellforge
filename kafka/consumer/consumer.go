package consumer

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/IBM/sarama"

	"github.com/DoomLordor/hellforge/kafka"
	"github.com/DoomLordor/hellforge/logger"
)

// Handler for kafka consumer
type Handler func(ctx context.Context, msg *sarama.ConsumerMessage) error

// BatchHandler for batch message processing
type BatchHandler func(ctx context.Context, msgs []*sarama.ConsumerMessage) error

// Consumer is an interface for work with kafka consumer
type Consumer interface {
	Run(ctx context.Context, enabled bool, topics []string, h Handler) error
	RunBatch(ctx context.Context, enabled bool, topics []string, bh BatchHandler, batchSize int, batchTimeout time.Duration) error
}

type consumer struct {
	kafkaBrokers []string
	groupID      string
	config       *sarama.Config
	logger       *logger.Logger
}

// NewConsumer constructor for new consumer
func NewConsumer(name string, kafkaBrokers []string, groupID string, opts ...kafka.Option) Consumer {
	kafkaCfg := sarama.NewConfig()
	kafkaCfg.Consumer.Offsets.AutoCommit.Enable = false
	kafkaCfg.Consumer.Return.Errors = true

	for _, opt := range opts {
		opt(kafkaCfg)
	}

	return &consumer{
		logger:       logger.NewLogger(fmt.Sprintf("kafka-consumer-%s", name)),
		kafkaBrokers: kafkaBrokers,
		config:       kafkaCfg,
		groupID:      groupID,
	}
}

func (c *consumer) Run(ctx context.Context, enabled bool, topics []string, h Handler) error {
	return c.run(ctx, enabled, topics, newHandler(c.logger, h))
}

func (c *consumer) RunBatch(ctx context.Context, enabled bool, topics []string, bh BatchHandler, batchSize int, batchTimeout time.Duration) error {
	return c.run(ctx, enabled, topics, newBatchHandler(c.logger, bh, batchSize, batchTimeout))
}

func (c *consumer) run(ctx context.Context, enabled bool, topics []string, h sarama.ConsumerGroupHandler) error {
	if !enabled {
		c.logger.Debug().Ctx(ctx).Msg("kafka is disabled from config")
		return nil
	}

	consumerGroup, err := sarama.NewConsumerGroup(c.kafkaBrokers, c.groupID, c.config)
	if err != nil {
		return fmt.Errorf("failed to create new consumer group: %w", err)
	}

	c.logger.Debug().Ctx(ctx).Msg("Consumer started...")
	go func() {
		done := ctx.Done()
		for {
			select {
			case <-done:
				err = consumerGroup.Close()
				if err != nil {
					c.logger.Err(err).Msg("kafka consumer was not closed correctly")
				}

				return
			default:
				err = consumerGroup.Consume(ctx, topics, h)
				if err != nil {
					if errors.Is(err, sarama.ErrClosedConsumerGroup) {
						return
					}

					c.logger.Err(err).Strs("topics", topics).Send()
				}
			}
		}
	}()

	return nil
}

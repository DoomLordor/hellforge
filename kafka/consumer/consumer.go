package consumer

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog"

	"github.com/DoomLordor/hellforge/logger"
)

// Handler for kafka consumer
type Handler func(ctx context.Context, msg *sarama.ConsumerMessage) error

// BatchHandler for batch message processing
type BatchHandler func(ctx context.Context, msgs []*sarama.ConsumerMessage) error

// Consumer is an interface for work with kafka consumer
type Consumer interface {
	Run(ctx context.Context, enabled bool, groupID string, topics []string, h Handler) error
	RunBatch(ctx context.Context, enabled bool, groupID string, topics []string, bh BatchHandler, batchSize int, batchTimeout time.Duration) error
	Close() error
}

type consumer struct {
	logger       zerolog.Logger
	config       *config
	kafkaBrokers []string
	kafkaConfig  *sarama.Config
	closers      []func() error
	quit         chan struct{}
}

// NewConsumer constructor for new consumer
func NewConsumer(name string, kafkaBrokers []string, options ...Option) Consumer {
	cfg := newConfig()
	for _, option := range options {
		option(cfg)
	}

	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Consumer.Offsets.AutoCommit.Enable = false
	kafkaConfig.Consumer.Return.Errors = true

	for _, option := range cfg.kafkaOptions {
		option(kafkaConfig)
	}

	return &consumer{
		logger:       logger.NewLogger("kafka-consumer").With().Str("name", name).Logger(),
		config:       cfg,
		kafkaBrokers: kafkaBrokers,
		kafkaConfig:  kafkaConfig,
		closers:      make([]func() error, 0, 5),
		quit:         make(chan struct{}),
	}
}

func (c *consumer) Run(ctx context.Context, enabled bool, groupID string, topics []string, h Handler) error {
	handler := newHandler(c.logger, topics, h, c.config.tracer)
	return c.run(ctx, enabled, groupID, topics, handler)
}

func (c *consumer) RunBatch(ctx context.Context, enabled bool, groupID string, topics []string, bh BatchHandler, batchSize int, batchTimeout time.Duration) error {
	handler := newBatchHandler(c.logger, topics, bh, batchSize, batchTimeout, c.config.tracer)
	return c.run(ctx, enabled, groupID, topics, handler)
}

func (c *consumer) run(ctx context.Context, enabled bool, groupID string, topics []string, h sarama.ConsumerGroupHandler) error {
	if !enabled {
		c.logger.Debug().Ctx(ctx).Strs("topics", topics).Msg("kafka is disabled from config")
		return nil
	}

	consumerGroup, err := sarama.NewConsumerGroup(c.kafkaBrokers, groupID, c.kafkaConfig)
	if err != nil {
		return fmt.Errorf("failed to create new consumer group: %w", err)
	}

	c.closers = append(c.closers, consumerGroup.Close)

	c.logger.Debug().Ctx(ctx).Strs("topics", topics).Msg("Consumer started...")
	go func() {
		done := ctx.Done()
		for {
			select {
			case <-c.quit:
				return
			case <-done:
				err = consumerGroup.Close()
				if err != nil {
					c.logger.Err(err).Ctx(ctx).Strs("topics", topics).Msg("kafka consumer was not closed correctly")
				}

				return
			default:
				err = consumerGroup.Consume(ctx, topics, h)
				if err != nil {
					if errors.Is(err, sarama.ErrClosedConsumerGroup) {
						return
					}

					c.logger.Err(err).Ctx(ctx).Strs("topics", topics).Send()
				}
			}
		}
	}()

	return nil
}

func (c *consumer) Close() error {
	close(c.quit)

	var multiErr *multierror.Error
	for _, closer := range c.closers {
		multiErr = multierror.Append(multiErr, closer())
	}

	return multiErr.ErrorOrNil()
}

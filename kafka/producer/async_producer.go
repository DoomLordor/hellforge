package producer

import (
	"time"

	"github.com/IBM/sarama"
	"github.com/rs/zerolog"

	"github.com/DoomLordor/hellforge/kafka"
	"github.com/DoomLordor/hellforge/logger"
)

type AsyncProducer interface {
	SendMessages(msgs ...*sarama.ProducerMessage)
	Close() error
}

// AsyncProducer definition
type asyncProducer struct {
	producer sarama.AsyncProducer
	done     chan struct{}
	logger   zerolog.Logger
}

// NewAsyncProducer async producer constructor
func NewAsyncProducer(name string, brokers []string, enabled bool, opts ...kafka.Option) (AsyncProducer, error) {
	if !enabled {
		return &asyncProducer{}, nil
	}

	kafkaCfg := sarama.NewConfig()
	kafkaCfg.Producer.Return.Successes = false
	kafkaCfg.Producer.Return.Errors = true
	kafkaCfg.Producer.Compression = sarama.CompressionLZ4
	kafkaCfg.Producer.CompressionLevel = 12
	kafkaCfg.Producer.Flush.Messages = 1000
	kafkaCfg.Producer.Flush.Frequency = 500 * time.Millisecond
	kafkaCfg.Producer.Flush.MaxMessages = 10000

	kafkaCfg.Net.DialTimeout = 5 * time.Second
	kafkaCfg.Metadata.Retry.Max = 1
	kafkaCfg.Metadata.Retry.Backoff = 1 * time.Second
	kafkaCfg.Metadata.Timeout = 5 * time.Second

	for _, optFn := range opts {
		optFn(kafkaCfg)
	}

	kafkaProducer, err := sarama.NewAsyncProducer(brokers, kafkaCfg)
	if err != nil {
		return nil, err
	}

	producer := &asyncProducer{
		producer: kafkaProducer,
		done:     make(chan struct{}),
		logger:   logger.NewLogger("kafka-async-producer").With().Str("name", name).Logger(),
	}

	go producer.logErrors()

	return producer, nil
}

// SendMessages produces a given messages asynchronously
func (p *asyncProducer) SendMessages(msgs ...*sarama.ProducerMessage) {
	if p.producer == nil {
		return
	}

	for _, msg := range msgs {
		p.producer.Input() <- msg
	}
}

// Close shuts down the producer
func (p *asyncProducer) Close() error {
	if p.producer == nil {
		return nil
	}

	close(p.done)
	return p.producer.Close()
}

func (p *asyncProducer) logErrors() {
	for {
		select {
		case err := <-p.producer.Errors():
			p.logger.Err(err).Str("topic", err.Msg.Topic).Send()
		case <-p.done:
			return
		}
	}
}

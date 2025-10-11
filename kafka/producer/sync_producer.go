package producer

import (
	"time"

	"github.com/IBM/sarama"

	"github.com/DoomLordor/hellforge/kafka"
)

type SyncProducer interface {
	SendMessage(msg *sarama.ProducerMessage) (partition int32, offset int64, err error)
	Close() error
}

// SyncProducer definition
type syncProducer struct {
	producer sarama.SyncProducer
}

// NewSyncProducer sync producer constructor
func NewSyncProducer(brokers []string, enabled bool, opts ...kafka.Option) (SyncProducer, error) {
	if !enabled {
		return &syncProducer{}, nil
	}

	kafkaCfg := sarama.NewConfig()
	kafkaCfg.Producer.Return.Successes = true
	kafkaCfg.Producer.Compression = sarama.CompressionLZ4
	kafkaCfg.Producer.CompressionLevel = 12

	kafkaCfg.Net.DialTimeout = 5 * time.Second
	kafkaCfg.Metadata.Retry.Max = 1
	kafkaCfg.Metadata.Retry.Backoff = 1 * time.Second
	kafkaCfg.Metadata.Timeout = 5 * time.Second

	for _, optFn := range opts {
		optFn(kafkaCfg)
	}

	kafkaProducer, err := sarama.NewSyncProducer(brokers, kafkaCfg)
	if err != nil {
		return nil, err
	}

	return &syncProducer{kafkaProducer}, nil
}

// SendMessage produces a given message
func (p *syncProducer) SendMessage(msg *sarama.ProducerMessage) (partition int32, offset int64, err error) {
	if p.producer == nil {
		return 0, 0, nil
	}

	return p.producer.SendMessage(msg)
}

// Close shuts down the producer
func (p *syncProducer) Close() error {
	if p.producer == nil {
		return nil
	}

	return p.producer.Close()
}

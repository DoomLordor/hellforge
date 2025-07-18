package kafka

import (
	"time"

	"github.com/IBM/sarama"
	"github.com/xdg-go/scram"
)

// Option func signature for adding additional options to sarama config
type Option func(*sarama.Config)

// WithSASL option for sasl authorization (consumer/producer)
func WithSASL(username, password string, saslMechanism sarama.SASLMechanism) Option {
	return func(kafkaCfg *sarama.Config) {
		if len(username) > 0 && len(password) > 0 {
			kafkaCfg.Net.SASL.Enable = true
			kafkaCfg.Net.SASL.Handshake = true
			kafkaCfg.Net.SASL.Mechanism = saslMechanism
			kafkaCfg.Net.SASL.User = username
			kafkaCfg.Net.SASL.Password = password

			switch saslMechanism {
			case sarama.SASLTypeSCRAMSHA256:
				kafkaCfg.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
					return &XDGSCRAMClient{HashGeneratorFcn: scram.SHA256}
				}
			case sarama.SASLTypeSCRAMSHA512:
				kafkaCfg.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
					return &XDGSCRAMClient{HashGeneratorFcn: scram.SHA512}
				}
			default:
			}
		}
	}
}

// WithOffset option for offset setting (consumer)
func WithOffset(offsetType int64) Option {
	return func(kafkaCfg *sarama.Config) {
		kafkaCfg.Consumer.Offsets.Initial = offsetType
	}
}

// WithDialTimeout option for setting tcp dial timeout (producer/consumer)
func WithDialTimeout(timeout time.Duration) Option {
	return func(kafkaCfg *sarama.Config) {
		kafkaCfg.Net.DialTimeout = timeout
	}
}

// WithMetadataTimeout option for setting metadata obtaining timeout (producer/consumer)
func WithMetadataTimeout(timeout time.Duration) Option {
	return func(kafkaCfg *sarama.Config) {
		kafkaCfg.Metadata.Timeout = timeout
	}
}

// WithProducerBufferSettings option for setting batching params for (async producer)
func WithProducerBufferSettings(batchSize, bufferSize int, flushFrequency time.Duration) Option {
	return func(kafkaCfg *sarama.Config) {
		kafkaCfg.Producer.Flush.Messages = batchSize
		kafkaCfg.Producer.Flush.Frequency = flushFrequency
		kafkaCfg.Producer.Flush.MaxMessages = bufferSize
	}
}

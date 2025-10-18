package defaults

import (
	"github.com/IBM/sarama"
	"go.opentelemetry.io/otel/trace"

	"github.com/DoomLordor/hellforge/kafka"
	"github.com/DoomLordor/hellforge/kafka/consumer"
	"github.com/DoomLordor/hellforge/kafka/producer"
)

type Consumer struct {
	Username string   `env:"USERNAME" envDefault:""`
	Password string   `env:"PASSWORD" envDefault:""`
	SASLType string   `env:"SASL_TYPE" envDefault:"PLAIN"`
	Brokers  []string `env:"BROKERS" envDefault:"localhost:9092"`
	Group    string   `env:"GROUP" envDefault:"cg_local"`
	Enabled  bool     `env:"ENABLED" envDefault:"true"`
}

func (c *Consumer) NewKafkaConsumer(name string, tracer trace.Tracer) consumer.Consumer {
	return consumer.NewConsumer(
		name,
		c.Brokers,
		consumer.WithTracer(tracer),
		consumer.WithKafkaOptions(
			kafka.WithOffset(sarama.OffsetOldest),
			kafka.WithSASL(c.Username, c.Password, sarama.SASLMechanism(c.SASLType)),
		),
	)
}

type Producer struct {
	Username string   `env:"USERNAME" envDefault:""`
	Password string   `env:"PASSWORD" envDefault:""`
	SASLType string   `env:"SASL_TYPE" envDefault:"PLAIN"`
	Brokers  []string `env:"BROKERS" envDefault:"localhost:9092"`
	Enabled  bool     `env:"ENABLED" envDefault:"true"`
}

func (p *Producer) NewSyncProducer() (producer.SyncProducer, error) {
	return producer.NewSyncProducer(
		p.Brokers,
		p.Enabled,
		kafka.WithSASL(p.Username, p.Password, sarama.SASLMechanism(p.SASLType)),
	)
}

func (p *Producer) NewAsyncProducer(name string) (producer.AsyncProducer, error) {
	return producer.NewAsyncProducer(
		name,
		p.Brokers,
		p.Enabled,
		kafka.WithSASL(p.Username, p.Password, sarama.SASLMechanism(p.SASLType)),
	)
}

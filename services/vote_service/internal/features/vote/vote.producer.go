package vote

import (
	"context"
	"log/slog"

	"github.com/IBM/sarama"
)

type Producer struct {
	producer sarama.SyncProducer
	logger *slog.Logger
}

type ProducerConfig struct {
	Producer sarama.SyncProducer
	Logger *slog.Logger
}

func NewProducer(producerConfig ProducerConfig) *Producer {

	return &Producer{
		producer: producerConfig.Producer,
		logger: producerConfig.Logger,
	}
}

func (p *Producer) Publish(
    ctx context.Context,
    topic string,
    key string,
    value []byte,
) error {

    msg := &sarama.ProducerMessage{
        Topic: topic,
        Key:   sarama.StringEncoder(key),
        Value: sarama.ByteEncoder(value),
    }

    _, _, err := p.producer.SendMessage(msg)
    return err
}
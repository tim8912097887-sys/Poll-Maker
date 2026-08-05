package event

import (
	"context"
	"log/slog"

	"github.com/IBM/sarama"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared/shutdown"
)

type EventBroker struct {
	producer sarama.SyncProducer
	brokers []string
	shutdownManager *shutdown.Manager
	logger *slog.Logger
}

type EventBrokerConfig struct {
	Brokers []string
	ShutdownManager *shutdown.Manager
	Logger *slog.Logger
}

func NewEventBroker(eventBrokerConfig EventBrokerConfig) *EventBroker {

	return &EventBroker{
		brokers: eventBrokerConfig.Brokers,
		shutdownManager: eventBrokerConfig.ShutdownManager,
		logger: eventBrokerConfig.Logger,
	}
}

func (e *EventBroker) Init() error {
	config := sarama.NewConfig()

    config.Version = sarama.V3_8_0_0
    config.Producer.Return.Successes = true
    config.Producer.RequiredAcks = sarama.WaitForAll
    config.Producer.Retry.Max = 3

    producer, err := sarama.NewSyncProducer(e.brokers, config)
    if err != nil {
        return err
    }

	e.producer = producer

	e.logger.Info("Registered shutdown handler for event broker")
	e.shutdownManager.Register(e.Close)
	e.logger.Info("Event broker initialized")
	return nil
}

func (e *EventBroker) Close(ctx context.Context) error {
    return e.producer.Close()
}

func (e *EventBroker) GetBroker() sarama.SyncProducer {
	return e.producer
}
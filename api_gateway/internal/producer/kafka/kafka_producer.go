package kafka

import (
	"api_gateway/internal/config"
	e "api_gateway/internal/domain/errors"
	"api_gateway/internal/domain/models"
	kafka2 "api_gateway/internal/kafka"
	"api_gateway/logger"
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"log/slog"
	"time"
)

type Producer struct {
	toImage *kafka.Conn
	toText  *kafka.Conn
	log     *slog.Logger
}

func NewProducer(ctx context.Context, cfg *config.KafkaConfig, log *slog.Logger) (*Producer, error) {
	conn1, err := kafka2.ConnectKafka(ctx, "kafka:9092", "content.text", 0)
	if err != nil {
		log.Error("Kafka connect error", err.Error())
		return nil, err
	}
	conn2, err := kafka2.ConnectKafka(ctx, "kafka:9092", "content.image", 0)
	if err != nil {
		log.Error("Kafka connect error", err.Error())
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	for _, broker := range cfg.Brokers {
		conn, err := kafka.DialContext(ctx, "tcp", broker)
		if err != nil {
			log.Error("failed to connect to kafka broker",
				logger.Err(err),
				slog.String("broker", broker))
			continue
		}
		defer conn.Close()

		brokers, err := conn.Brokers()
		if err != nil {
			log.Error("failed to get brokers list", logger.Err(err))
			continue
		}

		log.Info("kafka connection established",
			slog.String("broker", broker),
			slog.Any("available_brokers", brokers))
		break
	}

	return &Producer{
		toText:  conn1,
		toImage: conn2,
		log:     log,
	}, nil
}

func (p *Producer) ProduceContent(content *models.Content) error {
	const op = "kafka.Producer.ProduceContent"
	log := p.log.With(
		slog.String("op", op),
		slog.String("content_id", content.ID),
	)

	var conn *kafka.Conn
	switch content.Type {
	case models.ContentTypeText:
		conn = p.toText
	case models.ContentTypeImage:
		conn = p.toImage
	default:
		log.Error("invalid content type", slog.String("type", string(content.Type)))
		return e.ErrInvalidContentType
	}

	msg, err := json.Marshal(content)
	if err != nil {
		log.Error("failed to marshal content", logger.Err(err))
		return err
	}

	err = kafka2.SendToTopic(conn, msg)
	if err != nil {
		log.Error("failed to produce message",
			logger.Err(err))
		return err
	}

	log.Info("content successfully produced",
		slog.String("topic", string(content.Type)),
		slog.String("content_id", content.ID))
	return nil
}

func (p *Producer) Close() error {
	var err error

	if closeErr := p.toImage.Close(); closeErr != nil {
		p.log.Error("failed to close text writer", logger.Err(closeErr))
		err = closeErr
	}

	if closeErr := p.toText.Close(); closeErr != nil {
		p.log.Error("failed to close image writer", logger.Err(closeErr))
		err = closeErr
	}

	return err
}

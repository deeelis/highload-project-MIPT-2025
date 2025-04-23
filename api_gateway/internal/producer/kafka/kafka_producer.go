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
	const op = "kafka.NewProducer"
	log = log.With(slog.String("op", op))

	log.Info("connecting to Kafka...",
		slog.Any("brokers", cfg.Brokers),
		slog.String("text_topic", cfg.TextTopic),
		slog.String("image_topic", cfg.ImageTopic))

	conn1, err := kafka2.ConnectKafka(ctx, cfg.Brokers[0], cfg.TextTopic, 0)
	if err != nil {
		log.Error("failed to connect to text topic",
			logger.Err(err),
			slog.String("topic", cfg.TextTopic))
		return nil, err
	}
	log.Info("successfully connected to text topic",
		slog.String("topic", cfg.TextTopic))

	conn2, err := kafka2.ConnectKafka(ctx, cfg.Brokers[0], cfg.ImageTopic, 0)
	if err != nil {
		log.Error("failed to connect to image topic",
			logger.Err(err),
			slog.String("topic", cfg.ImageTopic))
		_ = conn1.Close()
		return nil, err
	}
	log.Info("successfully connected to image topic",
		slog.String("topic", cfg.ImageTopic))

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var connectedBroker string
	for _, broker := range cfg.Brokers {
		log.Debug("trying to connect to broker", slog.String("broker", broker))
		conn, err := kafka.DialContext(ctx, "tcp", broker)
		if err != nil {
			log.Warn("failed to connect to broker",
				logger.Err(err),
				slog.String("broker", broker))
			continue
		}
		defer conn.Close()

		brokers, err := conn.Brokers()
		if err != nil {
			log.Warn("failed to get brokers list",
				logger.Err(err),
				slog.String("broker", broker))
			continue
		}

		connectedBroker = broker
		log.Info("successfully connected to broker",
			slog.String("broker", broker),
			slog.Any("available_brokers", brokers))
		break
	}

	if connectedBroker == "" {
		log.Error("no available brokers")
		return nil, e.ErrKafkaUnavailable
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
		slog.String("content_type", string(content.Type)),
	)

	log.Debug("producing content...")

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
	log.Debug("content marshaled successfully")

	startTime := time.Now()
	err = kafka2.SendToTopic(conn, msg)
	if err != nil {
		log.Error("failed to produce message",
			logger.Err(err),
			slog.Duration("duration", time.Since(startTime)))
		return err
	}

	log.Info("content successfully produced",
		slog.String("topic", string(content.Type)),
		slog.Duration("duration", time.Since(startTime)))
	return nil
}

func (p *Producer) Close() error {
	const op = "kafka.Producer.Close"
	log := p.log.With(slog.String("op", op))

	log.Info("closing Kafka connections...")
	var err error

	if closeErr := p.toImage.Close(); closeErr != nil {
		log.Error("failed to close image writer", logger.Err(closeErr))
		err = closeErr
	} else {
		log.Debug("image writer closed successfully")
	}

	if closeErr := p.toText.Close(); closeErr != nil {
		log.Error("failed to close text writer", logger.Err(closeErr))
		err = closeErr
	} else {
		log.Debug("text writer closed successfully")
	}

	if err == nil {
		log.Info("all connections closed successfully")
	} else {
		log.Warn("closed with errors")
	}

	return err
}

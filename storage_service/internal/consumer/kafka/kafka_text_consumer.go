package kafka

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"log/slog"
	"storage_service/internal/config"
	"storage_service/internal/domain/models"
	"storage_service/internal/domain/repositories"
	kafka2 "storage_service/internal/kafka"
	services "storage_service/internal/usecases"
	"strconv"
	"time"
)

type TextConsumer struct {
	fromText *kafka.Conn
	log      *slog.Logger
	cfg      *config.Config
	usecase  repositories.StorageUsecase
}

func NewTextConsumer(ctx context.Context, cfg *config.Config, log *slog.Logger) (*TextConsumer, error) {
	conn1, err := kafka2.ConnectKafka(ctx, cfg.Kafka.Brokers[0], cfg.Kafka.TextTopic, 0)
	if err != nil {
		log.Error("Kafka connect error", err.Error())
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	for _, broker := range cfg.Kafka.Brokers {
		conn, err := kafka.DialContext(ctx, "tcp", broker)
		if err != nil {
			log.Error("failed to connect to kafka broker", err.Error(),
				slog.String("broker", broker))
			continue
		}
		defer conn.Close()

		brokers, err := conn.Brokers()
		if err != nil {
			log.Error("failed to get brokers list", err.Error())
			continue
		}

		log.Info("kafka connection established",
			slog.String("broker", broker),
			slog.Any("available_brokers", brokers))
		break
	}

	usecase, err := services.NewStorageUsecase(ctx, cfg, log)
	if err != nil {
		return nil, err
	}

	return &TextConsumer{
		fromText: conn1,
		log:      log,
		usecase:  usecase,
		cfg:      cfg,
	}, nil
}

func (c *TextConsumer) ConsumeText(ctx context.Context) error {
	const op = "kafka.Consumer.ConsumeContent"
	log := c.log.With(
		slog.String("op", op),
	)

	c.log.Info("starting kafka consumer")

	for {
		msg, err := kafka2.ReadFromTopic(c.fromText)
		if err != nil {
			c.log.Error("failed to fetch message", err.Error())
			return err
		}

		var content models.TextMessage
		var mes models.TextKafkaMessage
		if err := json.Unmarshal(msg, &mes); err != nil {
			c.log.Error("failed to unmarshal message",
				err.Error(),
			)
			continue
		}

		content.ID = mes.ID
		content.Content = mes.Data
		content.UserID = mes.UserID
		metadata := make(map[string]string)
		metadata["is_approved"] = strconv.FormatBool(mes.Analysis.IsApproved)
		metadata["is_spam"] = strconv.FormatBool(mes.Analysis.IsSpam)
		metadata["has_sensitive"] = strconv.FormatBool(mes.Analysis.HasSensitive)
		metadata["language"] = mes.Analysis.Language

		content.Analysis = metadata
		c.log.Debug("received message",
			"text_id", content.ID,
		)

		err = c.usecase.ProcessTextMessage(ctx, &content)
		if err != nil {
			c.log.Error("failed to process text",
				"text_id", content.ID,
				err.Error(),
			)
			continue
		}

		log.Info("content successfully saved",
			slog.String("topic", content.Content),
			slog.String("content_id", content.ID))
	}
}

func (c *TextConsumer) Close() error {
	var err error

	if closeErr := c.fromText.Close(); closeErr != nil {
		c.log.Error("failed to close text writer", err.Error())
		err = closeErr
	}

	return err
}

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
	"time"
)

type ImageConsumer struct {
	fromImage *kafka.Conn
	log       *slog.Logger
	cfg       *config.Config
	usecase   repositories.StorageUsecase
}

func NewImageConsumer(ctx context.Context, cfg *config.Config, log *slog.Logger) (*ImageConsumer, error) {
	conn1, err := kafka2.ConnectKafka(ctx, cfg.Kafka.Brokers[0], cfg.Kafka.ImageTopic, 0)
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

	return &ImageConsumer{
		fromImage: conn1,
		log:       log,
		cfg:       cfg,
		usecase:   usecase,
	}, nil
}

func (c *ImageConsumer) ConsumeImages(ctx context.Context) error {
	const op = "kafka.Consumer.ConsumeContent"
	log := c.log.With(
		slog.String("op", op),
	)

	c.log.Info("starting kafka consumer")

	for {
		msg, err := kafka2.ReadFromTopic(c.fromImage)
		if err != nil {
			c.log.Error("failed to fetch message", err.Error())
			return err
		}

		var content models.ImageKafkaMessage
		if err := json.Unmarshal(msg, &content); err != nil {
			c.log.Error("failed to unmarshal message",
				err.Error(),
			)
			continue
		}

		c.log.Debug("received message",
			"image_id", content.ID,
		)
        imageMessage := &models.ImageMessage{
          ID:         content.ID,
          UserID:     content.UserID,
          Data:       content.Data,
          NsfwScores: models.NsfwScoresResult{
            Drawings: content.NsfwScores.Drawings,
            Hentai:   content.NsfwScores.Hentai,
            Neutral:  content.NsfwScores.Neutral,
            Porn:     content.NsfwScores.Porn,
            Sexy:     content.NsfwScores.Sexy,
          },
        }
        
		err = c.usecase.ProcessImageMessage(ctx, imageMessage)
		if err != nil {
			c.log.Error("failed to process image",
				"image_id", content.ID,
				err.Error(),
			)
			continue
		}

		log.Info("content successfully saved",
			slog.String("topic", content.Data),
			slog.String("content_id", content.ID))
	}
}

func (c *ImageConsumer) Close() error {
	var err error

	if closeErr := c.fromImage.Close(); closeErr != nil {
		c.log.Error("failed to close text writer", err.Error())
		err = closeErr
	}

	return err
}

package kafka

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"log/slog"
	"text_analyzer_service/internal/config"
	"text_analyzer_service/internal/domain/models"
	kafka2 "text_analyzer_service/internal/kafka"
	"text_analyzer_service/internal/usecases"
	"text_analyzer_service/internal/usecases/text_analyzer_usecase"
	"text_analyzer_service/logger"
	"time"
)

type Consumer struct {
	fromGateway *kafka.Conn
	toStorage   *kafka.Conn
	usecase     usecases.TextAnalyzerUsecase
	log         *slog.Logger
}

func NewConsumer(ctx context.Context, cfg *config.KafkaConfig, log *slog.Logger) (*Consumer, error) {
	conn1, err := kafka2.ConnectKafka(ctx, cfg.Brokers[0], cfg.InputTopic, 0)
	if err != nil {
		log.Error("Kafka connect error", err.Error())
		return nil, err
	}
	conn2, err := kafka2.ConnectKafka(ctx, cfg.Brokers[0], cfg.ResultTopic, 0)
	if err != nil {
		log.Error("Kafka connect error", err.Error())
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	for _, broker := range cfg.Brokers {
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

	usecase := text_analyzer_usecase.NewTextAnalyzerUseCase(log)
	return &Consumer{
		fromGateway: conn1,
		toStorage:   conn2,
		log:         log,
		usecase:     usecase,
	}, nil
}

func (c *Consumer) ConsumeContent(ctx context.Context) error {
	const op = "kafka.Consumer.ConsumeContent"
	log := c.log.With(
		slog.String("op", op),
	)

	c.log.Info("starting kafka consumer")

	for {
		msg, err := kafka2.ReadFromTopic(c.fromGateway)
		if err != nil {
			c.log.Error("failed to fetch message", logger.Err(err))
			return err
		}

		var text models.TextContent
		if err := json.Unmarshal(msg, &text); err != nil {
			c.log.Error("failed to unmarshal message",
				logger.Err(err),
			)
			continue
		}

		c.log.Debug("received message",
			"text_id", text.ID,
		)

		result, err := c.usecase.ProcessText(ctx, &text)
		if err != nil {
			c.log.Error("failed to process text",
				"text_id", text.ID,
				logger.Err(err),
			)
			continue
		}

		message := models.TextAnalysisContent{
			ID:     text.ID,
			Data:   text.Data,
			UserID: text.UserID,
			Result: *result,
		}

		log.Info(message.Data, text.Data, message.UserID, text.UserID)
		msg, err = json.Marshal(message)
		if err != nil {
			log.Error("failed to marshal content", logger.Err(err))
			return err
		}

		err = kafka2.SendToTopic(c.toStorage, msg)
		if err != nil {
			log.Error("failed to produce message",
				logger.Err(err))
			return err
		}

		log.Info("content successfully produced",
			slog.String("topic", text.Data),
			slog.String("content_id", text.ID))
	}
}

func (c *Consumer) Close() error {
	var err error

	if closeErr := c.fromGateway.Close(); closeErr != nil {
		c.log.Error("failed to close text writer", logger.Err(closeErr))
		err = closeErr
	}

	if closeErr := c.toStorage.Close(); closeErr != nil {
		c.log.Error("failed to close image writer", logger.Err(closeErr))
		err = closeErr
	}

	return err
}

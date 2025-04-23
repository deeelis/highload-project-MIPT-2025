package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"text_analyzer_service/internal/config"
	"text_analyzer_service/internal/consumer/kafka"
	"text_analyzer_service/logger"
)

func main() {
	cfg, err := config.MustLoad()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	log := logger.SetUpLogger(cfg.Env)
	log.Info("starting service",
		"env", cfg.Env,
		"kafka_brokers", cfg.Kafka.Brokers,
		"input_topic", cfg.Kafka.InputTopic,
	)

	log.Info("Text analyzer service started", "env", cfg.Env)

	consumer, err := kafka.NewConsumer(context.Background(), cfg.Kafka, log)
	if err != nil {

	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		if err := consumer.ConsumeContent(ctx); err != nil {
			log.Error("consumer error", logger.Err(err))
			cancel()
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Info("Service is running", "pid", os.Getpid())
	<-sigChan

	log.Info("Shutting down...")
	cancel()
	log.Info("Service stopped")
}

package kafka

import (
	"context"
	"github.com/segmentio/kafka-go"
)

func ConnectKafka(ctx context.Context, address string, topic string, partition int) (*kafka.Conn, error) {
	return kafka.DialLeader(ctx, "tcp", "kafka:9092", topic, partition)
}

func SendToTopic(conn *kafka.Conn, message []byte) error {
	_, err := conn.WriteMessages(
		kafka.Message{Value: message},
	)
	return err
}

func ReadFromTopic(conn *kafka.Conn) ([]byte, error) {
	b := make([]byte, 10e3)
	n, err := conn.Read(b)
	if err != nil {
		return nil, err
	}
	return b[:n], nil
}

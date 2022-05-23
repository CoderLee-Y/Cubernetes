package kafka

import (
	"Cubernetes/pkg/utils/kafka/options"
	"context"
	"log"
	"net"
)
import "github.com/segmentio/kafka-go"

func NewKafkaClientByTopic(topic string, partition int, host string) *kafka.Conn {
	address := net.JoinHostPort(host, options.KafkaPort)
	conn, err := kafka.DialLeader(context.Background(), "tcp", address, topic, partition)

	if err != nil {
		log.Println("[Error]: Get Consumer client failed when subscribe topic", topic)
		return nil
	}

	return conn
}

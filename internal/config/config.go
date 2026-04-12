package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all runtime configuration.
type Config struct {
	Brokers           []string
	Topic             string
	GroupID           string
	NumOfPartitions   int
	ReplicationFactor int
	HTTPPort          string
}

// LoadConfig loads env vars from .env if present and returns Config.
func LoadConfig(forConsumer bool) Config {
	_ = godotenv.Load()

	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		log.Fatal("KAFKA_BROKERS not set")
	}
	brokersList := strings.Split(brokers, ",")

	topic := os.Getenv("KAFKA_TOPIC")
	if topic == "" {
		topic = "orders"
	}

	groupID := ""
	if forConsumer {
		groupID = os.Getenv("KAFKA_GROUP_ID")
		if groupID == "" {
			groupID = "order-consumer-group"
		}
	}

	numOfPartitions := 3 // default
	if os.Getenv("KAFKA_NUM_PARTITIONS") != "" {
		fmt.Sscanf(os.Getenv("KAFKA_NUM_PARTITIONS"), "%d", &numOfPartitions)
	}

	replicationFactor := 3
	if os.Getenv("KAFKA_REPLICATION_FACTOR") != "" {
		fmt.Sscanf(os.Getenv("KAFKA_REPLICATION_FACTOR"), "%d", &replicationFactor)
	}

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "4000"
	}

	return Config{
		Brokers:           brokersList,
		Topic:             topic,
		GroupID:           groupID,
		NumOfPartitions:   numOfPartitions,
		ReplicationFactor: replicationFactor,
		HTTPPort:          httpPort,
	}
}

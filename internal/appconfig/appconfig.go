package appconfig

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	Brokers           []string
	Topic             string
	GroupID           string
	NumOfPartitions   int
	ReplicationFactor int
	HTTPPort          string
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	ShutdownTimeout   time.Duration
}

func LoadProducerConfig() AppConfig {
	return loadConfig(false)
}

func LoadConsumerConfig() AppConfig {
	return loadConfig(true)
}

func loadConfig(forConsumer bool) AppConfig {
	_ = godotenv.Load()

	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		slog.Error("KAFKA_BROKERS environment variable is required but not set")
		os.Exit(1)
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

	numOfPartitions := 3
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

	return AppConfig{
		Brokers:           brokersList,
		Topic:             topic,
		GroupID:           groupID,
		NumOfPartitions:   numOfPartitions,
		ReplicationFactor: replicationFactor,
		HTTPPort:          httpPort,
		ReadTimeout:       parseDuration("HTTP_READ_TIMEOUT", 15*time.Second),
		WriteTimeout:      parseDuration("HTTP_WRITE_TIMEOUT", 15*time.Second),
		IdleTimeout:       parseDuration("HTTP_IDLE_TIMEOUT", 60*time.Second),
		ReadHeaderTimeout: parseDuration("HTTP_READ_HEADER_TIMEOUT", 5*time.Second),
		ShutdownTimeout:   parseDuration("SHUTDOWN_TIMEOUT", 30*time.Second),
	}
}

func parseDuration(envKey string, fallback time.Duration) time.Duration {
	val := os.Getenv(envKey)
	if val == "" {
		return fallback
	}
	d, err := time.ParseDuration(val)
	if err != nil {
		slog.Warn("Invalid duration for env var, using default",
			"key", envKey, "value", val, "default", fallback)
		return fallback
	}
	return d
}

package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type Config struct {
	Brokers           []string
	Topic             string
	GroupID           string
	NumOfPartitions   int
	ReplicationFactor int
}

func (c Config) GetBrokers() string {
	var brokers = c.Brokers[0]
	if len(c.Brokers) > 1 {
		for _, broker := range c.Brokers[1:] {
			brokers = brokers + "," + broker
		}
	}
	return brokers
}

type OrderEvent struct {
	OrderID    string    `json:"order_id"`
	CustomerID string    `json:"customer_id"`
	Amount     float64   `json:"amount"`
	CreatedAt  time.Time `json:"created_at"`
}

func NewProducer(cfg Config) (*kafka.Producer, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": cfg.GetBrokers(),
		"acks":              "all",
		"retries":           5,
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to create producer: %v", err)
	}
	return p, nil
}

func ProduceOrder(ctx context.Context, producer *kafka.Producer, order OrderEvent) error {
	message, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal order: %v", err)
	}

	topic := ctx.Value("topic").(string)
	err = producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          message,
	}, nil)
	if err != nil {
		return fmt.Errorf("failed to produce message: %v", err)
	}

	select {
	case ev := <-producer.Events():
		switch e := ev.(type) {
		case *kafka.Message:
			if e.TopicPartition.Error != nil {
				return fmt.Errorf("delivery failed: %v", e.TopicPartition.Error)
			}
			fmt.Printf("Produce order: %s\n", string(message))

		case kafka.Error:
			return fmt.Errorf("producer error: %v", e)
		}
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

func EnsureTopic(ctx context.Context, cfg Config) error {
	admin, err := kafka.NewAdminClient(&kafka.ConfigMap{
		"bootstrap.servers": cfg.GetBrokers(),
	})
	if err != nil {
		return fmt.Errorf("failed to create admin client: %v", err)
	}
	defer admin.Close()

	metadata, err := admin.GetMetadata(&cfg.Topic, false, 5000)
	if err == nil && len(metadata.Topics) > 0 && len(metadata.Topics[cfg.Topic].Partitions) > 0 {
		slog.Log(ctx, slog.LevelDebug, "topic %s already exists", cfg.Topic, nil)
		return nil
	}

	topicSpec := []kafka.TopicSpecification{
		{
			Topic:             cfg.Topic,
			NumPartitions:     cfg.NumOfPartitions,
			ReplicationFactor: cfg.ReplicationFactor,
		},
	}
	maxDur, err := time.ParseDuration("60s")
	if err != nil {
		panic("ParseDuration(60s)")
	}
	results, err := admin.CreateTopics(ctx, topicSpec, kafka.SetAdminOperationTimeout(maxDur))
	if err != nil {
		slog.Log(ctx, slog.LevelError, "failed to create topic %s: %v", cfg.Topic, err)
		return err
	}

	for i, result := range results {
		if result.Error.Code() != kafka.ErrNoError && result.Error.Code() != kafka.ErrTopicAlreadyExists {
			slog.WarnContext(ctx, "%s: %s %v", strconv.Itoa(i), result.Topic, result.Error)
		}
	}

	slog.Debug("")
	return nil
}

package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type Config struct {
	Brokers []string
	Topic   string
	GroupID string
}

func (c Config) GetBrokers() string {
	var brokers = ""
	for _, broker := range c.Brokers {
		brokers = brokers + broker
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
		"boostrap.servers": cfg.GetBrokers(),
		"acks":             "all",
		"retries":          5,
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

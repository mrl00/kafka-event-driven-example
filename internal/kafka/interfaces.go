package kafka

import (
	"context"
	"time"

	ckafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type Producer interface {
	Produce(msg *ckafka.Message, deliveryChan chan ckafka.Event) error
	Events() chan ckafka.Event
	Flush(timeoutMs int) int
	Close()
}

type Consumer interface {
	SubscribeTopics(topics []string, rebalanceCb ckafka.RebalanceCb) error
	ReadMessage(timeout time.Duration) (*ckafka.Message, error)
	Close() error
}

type AdminClient interface {
	CreateTopics(ctx context.Context, topics []ckafka.TopicSpecification, options ...ckafka.CreateTopicsAdminOption) ([]ckafka.TopicResult, error)
	Close()
}

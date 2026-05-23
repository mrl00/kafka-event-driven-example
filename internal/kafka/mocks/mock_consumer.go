package mocks

import (
	"time"

	ckafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type MockConsumer struct {
	SubscribeTopicsFunc func(topics []string, rebalanceCb ckafka.RebalanceCb) error
	ReadMessageFunc     func(timeout time.Duration) (*ckafka.Message, error)
	CloseFunc           func() error
}

func (m *MockConsumer) SubscribeTopics(topics []string, rebalanceCb ckafka.RebalanceCb) error {
	if m.SubscribeTopicsFunc == nil {
		return nil
	}
	return m.SubscribeTopicsFunc(topics, rebalanceCb)
}

func (m *MockConsumer) ReadMessage(timeout time.Duration) (*ckafka.Message, error) {
	if m.ReadMessageFunc == nil {
		return nil, nil
	}
	return m.ReadMessageFunc(timeout)
}

func (m *MockConsumer) Close() error {
	if m.CloseFunc == nil {
		return nil
	}
	return m.CloseFunc()
}

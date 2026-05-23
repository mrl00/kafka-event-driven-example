package mocks

import (
	ckafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type MockProducer struct {
	ProduceFunc func(msg *ckafka.Message, deliveryChan chan ckafka.Event) error
	EventsFunc  func() chan ckafka.Event
	FlushFunc   func(timeoutMs int) int
	CloseFunc   func()
}

func (m *MockProducer) Produce(msg *ckafka.Message, deliveryChan chan ckafka.Event) error {
	if m.ProduceFunc == nil {
		return nil
	}
	return m.ProduceFunc(msg, deliveryChan)
}

func (m *MockProducer) Events() chan ckafka.Event {
	if m.EventsFunc == nil {
		return make(chan ckafka.Event)
	}
	return m.EventsFunc()
}

func (m *MockProducer) Flush(timeoutMs int) int {
	if m.FlushFunc == nil {
		return 0
	}
	return m.FlushFunc(timeoutMs)
}

func (m *MockProducer) Close() {
	if m.CloseFunc != nil {
		m.CloseFunc()
	}
}

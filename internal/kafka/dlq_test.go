package kafka

import (
	"testing"

	ckafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

func TestDLQTopicNaming(t *testing.T) {
	cfg := Config{Topic: "orders"}
	dlq, _ := NewDLQProducer(cfg)
	if dlq == nil {
		t.Fatal("esperava DLQProducer válido")
	}
	if dlq.topic != "orders.dlq" {
		t.Errorf("esperava orders.dlq, recebeu %s", dlq.topic)
	}
}

func TestDLQSendHeaders(t *testing.T) {
	topic := "test-topic"
	msg := &ckafka.Message{
		TopicPartition: ckafka.TopicPartition{
			Topic:     &topic,
			Partition: 3,
			Offset:    42,
		},
		Value: []byte(`{"order_id":"ORD001"}`),
	}

	dlq := &DLQProducer{topic: "test-topic.dlq"}
	headers := dlq.buildHeaders(msg, "validation error", "validation", 2)

	headerMap := make(map[string]string)
	for _, h := range headers {
		headerMap[h.Key] = string(h.Value)
	}

	tests := []struct {
		key      string
		expected string
	}{
		{"dlq.error", "validation error"},
		{"dlq.error_type", "validation"},
		{"dlq.original_topic", "test-topic"},
		{"dlq.original_partition", "3"},
		{"dlq.original_offset", "42"},
		{"dlq.retry_count", "2"},
	}

	for _, tt := range tests {
		got, ok := headerMap[tt.key]
		if !ok {
			t.Errorf("header %s nao encontrado", tt.key)
			continue
		}
		if got != tt.expected {
			t.Errorf("header %s: esperava %s, recebeu %s", tt.key, tt.expected, got)
		}
	}

	if _, ok := headerMap["dlq.timestamp"]; !ok {
		t.Error("header dlq.timestamp nao encontrado")
	}
}

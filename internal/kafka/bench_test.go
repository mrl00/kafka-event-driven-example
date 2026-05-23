package kafka

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	ckafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/mrl00/kafka-event-driven-example/internal/kafka/mocks"
)

func BenchmarkGetBrokers(b *testing.B) {
	cfg := Config{Brokers: []string{"kafka1:9092", "kafka2:9092", "kafka3:9092"}}
	b.ResetTimer()
	for b.Loop() {
		cfg.GetBrokers()
	}
}

func BenchmarkProduceOrder(b *testing.B) {
	mockProducer := &mocks.MockProducer{
		ProduceFunc: func(msg *ckafka.Message, deliveryChan chan ckafka.Event) error {
			go func() {
				deliveryChan <- &ckafka.Message{
					TopicPartition: ckafka.TopicPartition{Partition: 1, Offset: ckafka.Offset(100)},
				}
			}()
			return nil
		},
	}
	event := OrderEvent{
		OrderID:    "ORD001",
		CustomerID: testCustomerID,
		Amount:     99.99,
		CreatedAt:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	ctx := context.Background()
	b.ResetTimer()
	for b.Loop() {
		_ = ProduceOrder(ctx, mockProducer, "orders", event)
	}
}

func BenchmarkProduceOrder_Parallel(b *testing.B) {
	mockProducer := &mocks.MockProducer{
		ProduceFunc: func(msg *ckafka.Message, deliveryChan chan ckafka.Event) error {
			go func() {
				deliveryChan <- &ckafka.Message{
					TopicPartition: ckafka.TopicPartition{Partition: 1, Offset: ckafka.Offset(100)},
				}
			}()
			return nil
		},
	}
	event := OrderEvent{
		OrderID:    "ORD001",
		CustomerID: testCustomerID,
		Amount:     99.99,
		CreatedAt:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	ctx := context.Background()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = ProduceOrder(ctx, mockProducer, "orders", event)
		}
	})
}

func BenchmarkConsumeOrders(b *testing.B) {
	target := int64(b.N)
	var processed atomic.Int64
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumer := &mocks.MockConsumer{
		ReadMessageFunc: func(timeout time.Duration) (*ckafka.Message, error) {
			if processed.Add(1) > target {
				cancel()
				return nil, ckafka.NewError(ckafka.ErrTimedOut, "timeout", false)
			}
			return &ckafka.Message{
				Value: []byte(`{"order_id":"ORD001","customer_id":"CUST001","amount":100.50,"created_at":"2024-01-01T00:00:00Z"}`),
			}, nil
		},
	}

	b.ResetTimer()
	_ = ConsumeOrders(ctx, consumer, nil)
}

func BenchmarkConsumeOrders_WithDLQ(b *testing.B) {
	target := int64(b.N)
	var processed atomic.Int64
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dlq := NewDLQProducerWithProducer(
		&mocks.MockProducer{
			ProduceFunc: func(msg *ckafka.Message, deliveryChan chan ckafka.Event) error {
				go func() {
					deliveryChan <- &ckafka.Message{
						TopicPartition: ckafka.TopicPartition{Partition: 0, Offset: ckafka.Offset(1)},
					}
				}()
				return nil
			},
		},
		"test",
	)
	consumer := &mocks.MockConsumer{
		ReadMessageFunc: func(timeout time.Duration) (*ckafka.Message, error) {
			if processed.Add(1) > target {
				cancel()
				return nil, ckafka.NewError(ckafka.ErrTimedOut, "timeout", false)
			}
			return &ckafka.Message{
				Value: []byte(`invalid json`),
			}, nil
		},
	}

	b.ResetTimer()
	_ = ConsumeOrders(ctx, consumer, dlq)
}

func BenchmarkWrapRetryable(b *testing.B) {
	err := ckafka.NewError(ckafka.ErrAllBrokersDown, "all brokers down", false)
	b.ResetTimer()
	for b.Loop() {
		_ = wrapRetryable(err)
	}
}

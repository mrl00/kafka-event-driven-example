package kafka

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	ckafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/mrl00/kafka-event-driven-example/internal/kafka/mocks"
)

const testOrderID = "ORD001"

func TestWrapRetryable(t *testing.T) {
	t.Run("deve retornar nil se o erro for nil", func(t *testing.T) {
		err := wrapRetryable(nil)
		if err != nil {
			t.Errorf("esperava nil, recebeu %v", err)
		}
	})

	t.Run("deve marcar erro como retryable", func(t *testing.T) {
		originalErr := errors.New("kafka connection failed")
		err := wrapRetryable(originalErr)

		rErr, ok := err.(kafkaError)
		if !ok {
			t.Fatal("erro deveria ser do tipo kafkaError")
		}

		if !rErr.Retryable() {
			t.Error("erro deveria ser retryable")
		}

		if rErr.Error() != originalErr.Error() {
			t.Errorf("mensagem de erro esperada %s, recebeu %s", originalErr.Error(), rErr.Error())
		}
	})
}

func TestKafkaConfig_GetBrokers(t *testing.T) {
	tests := []struct {
		name     string
		brokers  []string
		expected string
	}{
		{"lista vazia", []string{}, ""},
		{"um broker", []string{"localhost:9092"}, "localhost:9092"},
		{"múltiplos brokers", []string{"k1:9092", "k2:9092"}, "k1:9092,k2:9092"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{Brokers: tt.brokers}
			if got := cfg.GetBrokers(); got != tt.expected {
				t.Errorf("GetBrokers() = %v, esperava %v", got, tt.expected)
			}
		})
	}
}

func TestProduceOrder_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := ProduceOrder(ctx, nil, "test-topic", OrderEvent{OrderID: "123"})
	if err == nil {
		t.Error("esperava erro por contexto cancelado, recebeu nil")
	}
}

func TestProduceOrder_Success(t *testing.T) {
	var producedMsg *ckafka.Message
	mockProducer := &mocks.MockProducer{
		ProduceFunc: func(msg *ckafka.Message, deliveryChan chan ckafka.Event) error {
			producedMsg = msg
			go func() {
				deliveryChan <- &ckafka.Message{
					TopicPartition: ckafka.TopicPartition{
						Partition: 1,
						Offset:    ckafka.Offset(100),
					},
				}
			}()
			return nil
		},
	}

	event := OrderEvent{
		OrderID:    testOrderID,
		CustomerID: "CUST001",
		Amount:     99.99,
		CreatedAt:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	err := ProduceOrder(context.Background(), mockProducer, "orders", event)
	if err != nil {
		t.Fatalf("ProduceOrder retornou erro inesperado: %v", err)
	}

	if producedMsg == nil {
		t.Fatal("esperava mensagem produzida")
	}

	topic := *producedMsg.TopicPartition.Topic
	if topic != "orders" {
		t.Errorf("topic esperado 'orders', recebeu %s", topic)
	}

	expected := `{"order_id":"ORD001","customer_id":"CUST001","amount":99.99,"created_at":"2024-01-01T00:00:00Z"}`
	if string(producedMsg.Value) != expected {
		t.Errorf("valor esperado %s, recebeu %s", expected, string(producedMsg.Value))
	}
}

func TestProduceOrder_ProduceError(t *testing.T) {
	mockProducer := &mocks.MockProducer{
		ProduceFunc: func(msg *ckafka.Message, deliveryChan chan ckafka.Event) error {
			return errors.New("kafka unavailable")
		},
	}

	err := ProduceOrder(context.Background(), mockProducer, "orders", OrderEvent{OrderID: testOrderID})
	if err == nil {
		t.Fatal("esperava erro, recebeu nil")
	}
}

func TestProduceOrder_DeliveryError(t *testing.T) {
	mockProducer := &mocks.MockProducer{
		ProduceFunc: func(msg *ckafka.Message, deliveryChan chan ckafka.Event) error {
			go func() {
				deliveryChan <- &ckafka.Message{
					TopicPartition: ckafka.TopicPartition{
						Error: errors.New("leader not available"),
					},
				}
			}()
			return nil
		},
	}

	err := ProduceOrder(context.Background(), mockProducer, "orders", OrderEvent{OrderID: testOrderID})
	if err == nil {
		t.Fatal("esperava erro de delivery, recebeu nil")
	}
}

func TestProduceOrder_ContextCancelledDuringDelivery(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	started := make(chan struct{})
	var once sync.Once

	mockProducer := &mocks.MockProducer{
		ProduceFunc: func(msg *ckafka.Message, deliveryChan chan ckafka.Event) error {
			once.Do(func() { close(started) })
			return nil
		},
	}

	go func() {
		<-started
		cancel()
	}()

	err := ProduceOrder(ctx, mockProducer, "orders", OrderEvent{OrderID: testOrderID})
	if err == nil {
		t.Fatal("esperava erro de contexto cancelado, recebeu nil")
	}
}

func TestConsumeOrders_Success(t *testing.T) {
	var readCount atomic.Int32
	consumer := &mocks.MockConsumer{
		ReadMessageFunc: func(timeout time.Duration) (*ckafka.Message, error) {
			if readCount.Add(1) > 2 {
				return nil, ckafka.NewError(ckafka.ErrTimedOut, "timeout", false)
			}
			return &ckafka.Message{
				Value: []byte(`{"order_id":"ORD001","customer_id":"CUST001","amount":100.50,"created_at":"2024-01-01T00:00:00Z"}`),
			}, nil
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := ConsumeOrders(ctx, consumer, nil)
	if err != nil && err != context.DeadlineExceeded {
		t.Fatalf("ConsumeOrders retornou erro inesperado: %v", err)
	}
}

func TestConsumeOrders_InvalidJSON(t *testing.T) {
	var readCount atomic.Int32
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
			if readCount.Add(1) > 2 {
				return nil, ckafka.NewError(ckafka.ErrTimedOut, "timeout", false)
			}
			return &ckafka.Message{
				Value: []byte(`not json`),
			}, nil
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := ConsumeOrders(ctx, consumer, dlq)
	if err != nil && err != context.DeadlineExceeded {
		t.Fatalf("ConsumeOrders nao deveria falhar com JSON invalido: %v", err)
	}
}

func TestConsumeOrders_Timeout(t *testing.T) {
	var callCount atomic.Int32
	consumer := &mocks.MockConsumer{
		ReadMessageFunc: func(timeout time.Duration) (*ckafka.Message, error) {
			callCount.Add(1)
			return nil, ckafka.NewError(ckafka.ErrTimedOut, "timeout", false)
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := ConsumeOrders(ctx, consumer, nil)
	if err != nil && err != context.DeadlineExceeded {
		t.Fatalf("timeouts repetidos nao deveriam causar erro: %v", err)
	}
	if calls := callCount.Load(); calls < 2 {
		t.Errorf("esperava multiplas chamadas a ReadMessage (loop continuou), recebeu %d", calls)
	}
}

func TestConsumeOrders_ContextCancelled(t *testing.T) {
	consumer := &mocks.MockConsumer{
		ReadMessageFunc: func(timeout time.Duration) (*ckafka.Message, error) {
			return nil, ckafka.NewError(ckafka.ErrTimedOut, "timeout", false)
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := ConsumeOrders(ctx, consumer, nil)
	if err == nil {
		t.Fatal("esperava erro de contexto cancelado, recebeu nil")
	}
}

func TestConsumeOrders_KafkaError(t *testing.T) {
	var callCount atomic.Int32
	consumer := &mocks.MockConsumer{
		ReadMessageFunc: func(timeout time.Duration) (*ckafka.Message, error) {
			callCount.Add(1)
			return nil, ckafka.NewError(ckafka.ErrAllBrokersDown, "all brokers down", false)
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := ConsumeOrders(ctx, consumer, nil)
	if err != nil && err != context.DeadlineExceeded {
		t.Fatalf("erro nao-fatal do kafka nao deveria parar o consumo: %v", err)
	}
	if calls := callCount.Load(); calls < 2 {
		t.Errorf("esperava multiplas chamadas a ReadMessage (loop continuou), recebeu %d", calls)
	}
}

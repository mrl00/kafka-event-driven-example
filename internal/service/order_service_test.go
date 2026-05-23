package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	ckafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/mrl00/kafka-event-driven-example/internal/kafka"
	"github.com/mrl00/kafka-event-driven-example/internal/kafka/mocks"
)

const (
	testORD1  = "ORD001"
	testORD2  = "ORD002"
	testORD3  = "ORD003"
	testCUST1 = "CUST001"
	testCUST2 = "CUST002"
	testCUST3 = "CUST003"
)

func TestProcessOrder_Success(t *testing.T) {
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

	svc := NewOrderService(mockProducer, "orders")
	orders := []kafka.OrderEvent{
		{OrderID: testORD1, CustomerID: testCUST1, Amount: 100.50, CreatedAt: time.Now()},
		{OrderID: testORD2, CustomerID: testCUST2, Amount: 200.75, CreatedAt: time.Now()},
	}

	failedIDs := svc.ProcessOrder(context.Background(), orders)

	if len(failedIDs) != 0 {
		t.Errorf("esperava 0 falhas, recebeu %d: %v", len(failedIDs), failedIDs)
	}
}

func TestProcessOrder_PartialFailure(t *testing.T) {
	mockProducer := &mocks.MockProducer{
		ProduceFunc: func(msg *ckafka.Message, deliveryChan chan ckafka.Event) error {
			var event struct {
				OrderID string `json:"order_id"`
			}
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				return err
			}
			if event.OrderID == testORD2 {
				return errors.New("kafka unavailable")
			}
			go func() {
				deliveryChan <- &ckafka.Message{
					TopicPartition: ckafka.TopicPartition{Partition: 1, Offset: ckafka.Offset(100)},
				}
			}()
			return nil
		},
	}

	svc := NewOrderService(mockProducer, "orders")
	orders := []kafka.OrderEvent{
		{OrderID: testORD1, CustomerID: testCUST1, Amount: 100.50, CreatedAt: time.Now()},
		{OrderID: testORD2, CustomerID: testCUST2, Amount: 200.75, CreatedAt: time.Now()},
		{OrderID: testORD3, CustomerID: testCUST3, Amount: 50.25, CreatedAt: time.Now()},
	}

	failedIDs := svc.ProcessOrder(context.Background(), orders)

	if len(failedIDs) != 1 {
		t.Fatalf("esperava 1 falha, recebeu %d: %v", len(failedIDs), failedIDs)
	}
	if failedIDs[0] != testORD2 {
		t.Errorf("esperava ORD002 como falha, recebeu %s", failedIDs[0])
	}
}

func TestProcessOrder_AllFailed(t *testing.T) {
	mockProducer := &mocks.MockProducer{
		ProduceFunc: func(msg *ckafka.Message, deliveryChan chan ckafka.Event) error {
			return errors.New("kafka unavailable")
		},
	}

	svc := NewOrderService(mockProducer, "orders")
	orders := []kafka.OrderEvent{
		{OrderID: testORD1, CustomerID: testCUST1, Amount: 100.50, CreatedAt: time.Now()},
		{OrderID: testORD2, CustomerID: testCUST2, Amount: 200.75, CreatedAt: time.Now()},
	}

	failedIDs := svc.ProcessOrder(context.Background(), orders)

	if len(failedIDs) != 2 {
		t.Errorf("esperava 2 falhas, recebeu %d: %v", len(failedIDs), failedIDs)
	}
}

func TestProcessOrder_EmptyOrders(t *testing.T) {
	mockProducer := &mocks.MockProducer{
		ProduceFunc: func(msg *ckafka.Message, deliveryChan chan ckafka.Event) error {
			t.Error("Produce nao deveria ser chamado para slice vazio")
			return nil
		},
	}

	svc := NewOrderService(mockProducer, "orders")
	failedIDs := svc.ProcessOrder(context.Background(), nil)

	if len(failedIDs) != 0 {
		t.Errorf("esperava 0 falhas para slice vazio, recebeu %d", len(failedIDs))
	}
}

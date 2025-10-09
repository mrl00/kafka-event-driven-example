package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type OrderEvent struct {
	OrderID    string    `json:"order_id"`
	CustomerID string    `json:"customer_id"`
	Amount     float64   `json:"amount"`
	CreatedAt  time.Time `json:"created_at"`
}

const (
	brokerAddress = "localhost:29092"
	topic         = "orders"
	consumerGroup = "order-processing-group"
)

func createProducer() *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP(brokerAddress),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
}

func produceOrderEvent(ctx context.Context, order OrderEvent) error {
	w := createProducer()
	defer w.Close()

	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal order: %v", err)
	}

	err = w.WriteMessages(ctx,
		kafka.Message{
			Key:   []byte(order.OrderID),
			Value: data,
		},
	)

	if err != nil {
		return fmt.Errorf("failed to write message: %v", err)
	}

	log.Printf("Produce order event: %s", order.OrderID)
	return nil
}

func createConsumer() *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{brokerAddress},
		Topic:    topic,
		GroupID:  consumerGroup,
		MinBytes: 10e3, //10KB
		MaxBytes: 10e6, //10MB
		MaxWait:  1 * time.Second,
	})
}

func consumerOrderEvents(ctx context.Context) {
	r := createConsumer()
	defer r.Close()

	for {
		msg, err := r.ReadMessage(ctx)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			continue
		}

		var order OrderEvent
		if err := json.Unmarshal(msg.Value, &order); err != nil {
			log.Printf("Error unmarshaling order: %v", err)
			continue
		}

		log.Printf("Processed order: ID=%s, Customer=%s, Amount=%.2f, CreatedAt=%v",
			order.OrderID, order.CustomerID, order.Amount, order.CreatedAt)

		if err := r.CommitMessages(ctx, msg); err != nil {
			log.Printf("Error comitting message: %v", err)
		}
	}
}

func main() {
	ctx := context.Background()

	go func() {
		log.Println("Starting consumer...")
		consumerOrderEvents(ctx)
	}()

	orders := []OrderEvent{
		{
			OrderID:    "ORD001",
			CustomerID: "CUST001",
			Amount:     199.99,
			CreatedAt:  time.Now(),
		},
		{
			OrderID:    "ORD002",
			CustomerID: "CUST002",
			Amount:     299.50,
			CreatedAt:  time.Now(),
		},
	}

	for _, order := range orders {
		if err := produceOrderEvent(ctx, order); err != nil {
			log.Printf("Error producing order event: %v", err)
		}

		time.Sleep(1 * time.Second)
	}

	select {}
}

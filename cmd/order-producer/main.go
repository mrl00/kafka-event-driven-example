package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/mrl00/kafka-event-driven-example/internal/kafka"
	"github.com/mrl00/kafka-event-driven-example/internal/router"
)

func server() {
	r := router.New()
	if err := http.ListenAndServe(":4000", r); err != nil {
		log.Fatal("failed to start server: ", err)
	}
}

func main() {

	go server()

	var ctx = context.Background()

	cfg := kafka.Config{
		Brokers:           []string{"kafka1:19092", "kafka2:19092", "kafka3:19092"},
		Topic:             "orders",
		NumOfPartitions:   3,
		ReplicationFactor: 3,
	}

	ctx = context.WithValue(ctx, "topic", cfg.Topic)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if err := kafka.EnsureTopic(ctx, cfg); err != nil {
		log.Fatalf("ensure topic error: %v", err)
	}

	producer, err := kafka.NewProducer(cfg)
	if err != nil {
		log.Fatalf("failed to create producer: %v", err)
	}
	defer producer.Close()

	orders := []kafka.OrderEvent{
		{OrderID: "ORD001", CustomerID: "CUST001", Amount: 199.99},
		{OrderID: "ORD002", CustomerID: "CUST002", Amount: 299.99},
		{OrderID: "ORD003", CustomerID: "CUST003", Amount: 149.50},
		{OrderID: "ORD004", CustomerID: "CUST004", Amount: 499.99},
		{OrderID: "ORD005", CustomerID: "CUST005", Amount: 79.90},
		{OrderID: "ORD006", CustomerID: "CUST006", Amount: 249.75},
		{OrderID: "ORD007", CustomerID: "CUST007", Amount: 399.00},
		{OrderID: "ORD008", CustomerID: "CUST008", Amount: 99.99},
		{OrderID: "ORD009", CustomerID: "CUST009", Amount: 199.00},
		{OrderID: "ORD010", CustomerID: "CUST010", Amount: 599.95},
		{OrderID: "ORD011", CustomerID: "CUST011", Amount: 129.49},
		{OrderID: "ORD012", CustomerID: "CUST012", Amount: 349.99},
		{OrderID: "ORD013", CustomerID: "CUST013", Amount: 89.90},
		{OrderID: "ORD014", CustomerID: "CUST014", Amount: 279.99},
		{OrderID: "ORD015", CustomerID: "CUST015", Amount: 450.00},
		{OrderID: "ORD016", CustomerID: "CUST016", Amount: 69.99},
		{OrderID: "ORD017", CustomerID: "CUST017", Amount: 189.50},
		{OrderID: "ORD018", CustomerID: "CUST018", Amount: 529.99},
		{OrderID: "ORD019", CustomerID: "CUST019", Amount: 109.75},
		{OrderID: "ORD020", CustomerID: "CUST020", Amount: 399.49},
	}

	for _, order := range orders {
		if err := kafka.ProduceOrder(ctx, producer, order); err != nil {
			log.Printf("Failed to produce order: %v", err)
		}
		time.Sleep(1 * time.Second)
	}

	select {}
}

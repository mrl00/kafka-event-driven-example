package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/mrl00/kafka-event-driven-example/internal/config"
	"github.com/mrl00/kafka-event-driven-example/internal/kafka"
	"github.com/mrl00/kafka-event-driven-example/internal/router"
)

func server(port string) {
	r := router.New()
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal("failed to start server: ", err)
	}
}

func main() {

	cfg := config.LoadConfig(false)
	go server(cfg.HTTPPort)

	var ctx = context.Background()

	kcfg := kafka.Config{
		Brokers:           cfg.Brokers,
		Topic:             cfg.Topic,
		NumOfPartitions:   cfg.NumOfPartitions,
		ReplicationFactor: cfg.ReplicationFactor,
	}

	ctx = context.WithValue(ctx, "topic", kcfg.Topic)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if err := kafka.EnsureTopic(ctx, kcfg); err != nil {
		log.Fatalf("ensure topic error: %v", err)
	}

	producer, err := kafka.NewProducer(kcfg)
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

package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"

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

	ctx := context.Background()
	cfg := kafka.Config{
		Brokers: []string{"kafka1:19092", "kafka2:19092", "kafka3:19092"},
		Topic:   "orders",
		GroupID: "order-consumer-group",
	}
	if err := kafka.EnsureTopic(ctx, cfg); err != nil {
		log.Fatalf("ensure topic error: %v", err)
	}

	consumer, err := kafka.NewConsumer(ctx, cfg)
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer consumer.Close()
	slog.InfoContext(ctx, "consumer creaated")

	if err := kafka.ConsumeOrders(ctx, consumer); err != nil {
		log.Fatalf("Failed to consume orders: %v", err)
	}
}

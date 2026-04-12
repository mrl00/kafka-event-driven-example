package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"

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
	cfg := config.LoadConfig(true)
	go server(cfg.HTTPPort)

	ctx := context.Background()
	kcfg := kafka.Config{
		Brokers: cfg.Brokers,
		Topic:   cfg.Topic,
		GroupID: cfg.GroupID,
	}
	if err := kafka.EnsureTopic(ctx, kcfg); err != nil {
		log.Fatalf("ensure topic error: %v", err)
	}

	consumer, err := kafka.NewConsumer(ctx, kcfg)
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer consumer.Close()
	slog.InfoContext(ctx, "consumer created")

	if err := kafka.ConsumeOrders(ctx, consumer); err != nil {
		log.Fatalf("Failed to consume orders: %v", err)
	}
}

package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/mrl00/kafka-event-driven-example/internal/appconfig"
	"github.com/mrl00/kafka-event-driven-example/internal/kafka"
	"github.com/mrl00/kafka-event-driven-example/internal/lifecycle"
	"github.com/mrl00/kafka-event-driven-example/internal/router"
	"github.com/mrl00/kafka-event-driven-example/internal/server"
)

func main() {
	cfg := appconfig.LoadConsumerConfig()

	ctx, cancel := context.WithCancel(context.Background())

	kcfg := kafka.Config{
		Brokers: cfg.Brokers,
		Topic:   cfg.Topic,
		GroupID: cfg.GroupID,
	}

	if err := kafka.EnsureTopic(ctx, kcfg); err != nil {
		slog.Error("erro ao assegurar tópico", "error", err)
		return
	}

	consumer, err := kafka.NewConsumer(ctx, kcfg)
	if err != nil {
		slog.Error("falha ao criar consumer", "error", err)
		return
	}
	slog.Info("Consumer Kafka criado com sucesso")

	r := router.ConsumerRouter()

	srv := server.StartServer(server.Config{
		Name:              "consumer",
		Port:              cfg.HTTPPort,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
	}, r)

	dlq, err := kafka.NewDLQProducer(kcfg)
	if err != nil {
		slog.Error("falha ao criar DLQ producer", "error", err)
		return
	}

	go func() {
		slog.Info("Iniciando consumo de ordens...")
		if err := kafka.ConsumeOrders(ctx, consumer, dlq); err != nil {
			slog.Error("erro durante o consumo de ordens", "error", err)
		}
	}()

	cleanups := []func(){
		func() {
			slog.Info("Encerrando servidor HTTP...")
			sdCtx, sdCancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer sdCancel()
			if err := srv.Shutdown(sdCtx); err != nil {
				slog.Error("erro ao desligar servidor", "error", err)
			}
		},
		func() {
			slog.Info("Fechando conexão com Kafka Consumer...")
			if err := consumer.Close(); err != nil {
				slog.Error("erro ao fechar consumer", "error", err)
			}
		},
		func() {
			slog.Info("Fechando DLQ Producer...")
			dlq.Close()
		},
	}

	lifecycle.WaitForShutdownSignal(ctx, cancel, cfg.ShutdownTimeout, cleanups...)
}

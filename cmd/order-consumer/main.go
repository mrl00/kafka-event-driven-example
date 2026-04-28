package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/mrl00/kafka-event-driven-example/internal/config"
	"github.com/mrl00/kafka-event-driven-example/internal/lifecycle"
	"github.com/mrl00/kafka-event-driven-example/internal/mykafka"
	"github.com/mrl00/kafka-event-driven-example/internal/server"
)

func main() {
	cfg := config.LoadConfig(true)

	ctx, cancel := context.WithCancel(context.Background())

	kcfg := mykafka.KafkaConfig{
		Brokers: cfg.Brokers,
		Topic:   cfg.Topic,
		GroupID: cfg.GroupID,
	}

	if err := mykafka.EnsureTopic(ctx, kcfg); err != nil {
		slog.Error("erro ao assegurar tópico", "error", err)
		return
	}

	consumer, err := mykafka.NewConsumer(ctx, kcfg)
	if err != nil {
		slog.Error("falha ao criar consumer", "error", err)
		return
	}
	slog.Info("Consumer Kafka criado com sucesso")

	srv := server.StartServer("consumer", cfg.HTTPPort)

	dlq, err := mykafka.NewDLQProducer(kcfg)
	if err != nil {
		slog.Error("falha ao criar DLQ producer", "error", err)
		return
	}

	go func() {
		slog.Info("Iniciando consumo de ordens...")
		if err := mykafka.ConsumeOrders(ctx, consumer, dlq); err != nil {
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

	lifecycle.WaitForShutdownSignal(ctx, cancel, cleanups...)
}

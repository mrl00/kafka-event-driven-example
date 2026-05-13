package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/mrl00/kafka-event-driven-example/internal/appconfig"
	"github.com/mrl00/kafka-event-driven-example/internal/handler"
	"github.com/mrl00/kafka-event-driven-example/internal/kafka"
	"github.com/mrl00/kafka-event-driven-example/internal/lifecycle"
	"github.com/mrl00/kafka-event-driven-example/internal/router"
	"github.com/mrl00/kafka-event-driven-example/internal/server"
	"github.com/mrl00/kafka-event-driven-example/internal/service"
)

func main() {
	cfg := appconfig.LoadProducerConfig()

	ctx, cancel := context.WithCancel(context.Background())

	kcfg := kafka.Config{
		Brokers:           cfg.Brokers,
		Topic:             cfg.Topic,
		NumOfPartitions:   cfg.NumOfPartitions,
		ReplicationFactor: cfg.ReplicationFactor,
	}

	if err := kafka.EnsureTopic(ctx, kcfg); err != nil {
		slog.Error("erro ao assegurar tópico", "error", err)
		return
	}

	producer, err := kafka.NewProducer(kcfg)
	if err != nil {
		slog.Error("falha ao criar producer", "error", err)
		return
	}

	orderSvc := service.NewOrderService(producer, cfg.Topic)

	orderHandler := handler.NewOrderHandler(orderSvc)

	r := router.ProducerRouter(orderHandler)

	srv := server.StartServer(server.Config{
		Name:              "producer",
		Port:              cfg.HTTPPort,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
	}, r)

	cleanups := []func(){
		func() {
			slog.Info("Encerrando servidor HTTP...")
			sdCtx, sdCancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer sdCancel()
			if err := srv.Shutdown(sdCtx); err != nil {
				slog.Error("erro no shutdown do servidor", "error", err)
			}
		},
		func() {
			slog.Info("Limpando buffers e fechando producer Kafka...")
			producer.Flush(15 * 1000)
			producer.Close()
		},
	}

	lifecycle.WaitForShutdownSignal(ctx, cancel, cfg.ShutdownTimeout, cleanups...)
}


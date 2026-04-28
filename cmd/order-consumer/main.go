package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/mrl00/kafka-event-driven-example/internal/config"
	"github.com/mrl00/kafka-event-driven-example/internal/kafka"
	"github.com/mrl00/kafka-event-driven-example/internal/lifecycle"
	"github.com/mrl00/kafka-event-driven-example/internal/server"
)

func main() {
	// Carrega configurações
	cfg := config.LoadConfig(true)

	// Contexto raiz que será propagado para o Kafka e Servidor
	ctx, cancel := context.WithCancel(context.Background())

	// Inicialização do Kafka
	kcfg := kafka.KafkaConfig{
		Brokers: cfg.Brokers,
		Topic:   cfg.Topic,
		GroupID: cfg.GroupID,
	}

	// Verifica se o tópico existe antes de começar
	if err := kafka.EnsureTopic(ctx, kcfg); err != nil {
		slog.Error("erro ao assegurar tópico", "error", err)
		return
	}

	// Cria o consumer
	consumer, err := kafka.NewConsumer(ctx, kcfg)
	if err != nil {
		slog.Error("falha ao criar consumer", "error", err)
		return
	}
	slog.Info("Consumer Kafka criado com sucesso")

	// Inicia o servidor (Health checks/Metrics)
	srv := server.StartServer("consumer", cfg.HTTPPort)

	// Goroutine para processamento de mensagens
	go func() {
		slog.Info("Iniciando consumo de ordens...")
		if err := kafka.ConsumeOrders(ctx, consumer); err != nil {
			slog.Error("erro durante o consumo de ordens", "error", err)
		}
	}()

	// Lista de tarefas de limpeza (LIFO)
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
			// O Close() aqui garante que o grupo de consumo seja notificado
			// e o rebalanceamento ocorra mais rápido para outros consumers
			if err := consumer.Close(); err != nil {
				slog.Error("erro ao fechar consumer", "error", err)
			}
		},
	}

	// Aguarda sinais SIGINT/SIGTERM e gerencia o encerramento
	lifecycle.WaitForShutdownSignal(ctx, cancel, cleanups...)
}

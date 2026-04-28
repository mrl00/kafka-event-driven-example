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

	cfg := config.LoadConfig(false)

	ctx, cancel := context.WithCancel(context.Background())

	kcfg := kafka.KafkaConfig{
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

	srv := server.StartServer("producer", cfg.HTTPPort)

	go func() {
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

		for _, orderEvent := range orders {
			select {
			case <-ctx.Done(): // Para o loop se o shutdown começar
				return
			default:
				if err := kafka.ProduceOrder(ctx, producer, cfg.Topic, orderEvent); err != nil {
					slog.Error("falha ao produzir ordem", "order_id", orderEvent.OrderID, "error", err)
				}
				time.Sleep(1 * time.Second)
			}
		}
	}()

	// Lista de funções de limpeza (Cleanups)
	cleanups := []func(){
		func() {
			slog.Info("Encerrando servidor HTTP...")
			// Contexto de timeout específico para o shutdown do server
			sdCtx, sdCancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer sdCancel()
			if err := srv.Shutdown(sdCtx); err != nil {
				slog.Error("erro no shutdown do servidor", "error", err)
			}
		},
		func() {
			slog.Info("Limpando buffers e fechando producer Kafka...")
			// Flush garante que mensagens no buffer local sejam enviadas
			producer.Flush(15 * 1000) // 15 segundos em ms
			producer.Close()
		},
	}

	// Aguarda o sinal de encerramento
	lifecycle.WaitForShutdownSignal(ctx, cancel, cleanups...)
}

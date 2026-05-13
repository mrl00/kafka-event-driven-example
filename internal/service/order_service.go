package service

import (
	"context"
	"log/slog"

	"github.com/mrl00/kafka-event-driven-example/internal/kafka"
)

type OrderService struct {
	producer *kafka.Producer
	topic    string
}

func NewOrderService(producer *kafka.Producer, topic string) *OrderService {
	return &OrderService{producer: producer, topic: topic}
}

func (s *OrderService) ProcessOrder(ctx context.Context, orders []kafka.OrderEvent) []string {
	var failedIDs []string

	for _, order := range orders {
		if err := ctx.Err(); err != nil {
			slog.Error("Contexto cancelado durante processamento de lote")
			failedIDs = append(failedIDs, order.OrderID)
			continue
		}

		err := kafka.ProduceOrder(ctx, s.producer, s.topic, order)
		if err != nil {
			slog.Error("falha ao produzir ordem", "order_id", order.OrderID, "error", err)
			failedIDs = append(failedIDs, order.OrderID)
		}
	}

	return failedIDs
}

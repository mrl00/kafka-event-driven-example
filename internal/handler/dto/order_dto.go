package dto

import (
	"time"

	"github.com/mrl00/kafka-event-driven-example/internal/kafka"
)

type CreateOrderDTO struct {
	OrderID    string  `json:"order_id"`
	CustomerID string  `json:"customer_id"`
	Amount     float64 `json:"amount"`
}

func NewOrderDTO(orderID string, customerID string, amount float64) CreateOrderDTO {
	return CreateOrderDTO{
		OrderID:    orderID,
		CustomerID: customerID,
		Amount:     amount,
	}
}

func (dto CreateOrderDTO) ToEvent() kafka.OrderEvent {
	return kafka.OrderEvent{
		OrderID:    dto.OrderID,
		CustomerID: dto.CustomerID,
		Amount:     dto.Amount,
		CreatedAt:  time.Now(),
	}
}

package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/IBM/fp-go/v2/array"
	"github.com/mrl00/kafka-event-driven-example/internal/handler/dto"
	"github.com/mrl00/kafka-event-driven-example/internal/kafka"
)

type OrderProcessor interface {
	ProcessOrder(ctx context.Context, orders []kafka.OrderEvent) []string
}

type OrderHandler struct {
	service OrderProcessor
}

func NewOrderHandler(service OrderProcessor) *OrderHandler {
	return &OrderHandler{service: service}
}

func (h *OrderHandler) HandleCreateOrder() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input []dto.CreateOrderDTO
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		events := array.Map(func(i dto.CreateOrderDTO) kafka.OrderEvent {
			return i.ToEvent()
		})(input)

		failedIDs := h.service.ProcessOrder(r.Context(), events)

		if len(failedIDs) > 0 {
			if len(failedIDs) == len(events) {
				w.WriteHeader(http.StatusServiceUnavailable)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"error":  "kafka unavailable",
					"failed": failedIDs,
				})
				return
			}
			w.WriteHeader(http.StatusMultiStatus)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"failed": failedIDs})
			return
		}
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("Orders sent to processing"))
	}
}

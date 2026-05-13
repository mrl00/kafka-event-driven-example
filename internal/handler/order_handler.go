package handler

import (
	"encoding/json"
	"net/http"

	"github.com/IBM/fp-go/v2/array"
	"github.com/mrl00/kafka-event-driven-example/internal/handler/dto"
	"github.com/mrl00/kafka-event-driven-example/internal/kafka"
	"github.com/mrl00/kafka-event-driven-example/internal/service"
)

type OrderHandler struct {
	service *service.OrderService
}

func NewOrderHandler(service *service.OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}

func (h *OrderHandler) HandleCreateOrder() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input []dto.CreateOrderDTO
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		failedIDs := h.service.ProcessOrder(r.Context(), array.Map(func(i dto.CreateOrderDTO) kafka.OrderEvent {
			return i.ToEvent()
		})(input))

		if len(failedIDs) > 0 {
			w.WriteHeader(http.StatusMultiStatus)
			json.NewEncoder(w).Encode(map[string]interface{}{"failed": failedIDs})
			return
		}
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("Orders sent to processing"))
	}
}

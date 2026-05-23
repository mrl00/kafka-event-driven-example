package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mrl00/kafka-event-driven-example/internal/handler"
	"github.com/mrl00/kafka-event-driven-example/internal/kafka"
)

const testOrderID = "ORD001"

type mockOrderProcessor struct {
	processFunc func(ctx context.Context, orders []kafka.OrderEvent) []string
}

func (m *mockOrderProcessor) ProcessOrder(ctx context.Context, orders []kafka.OrderEvent) []string {
	return m.processFunc(ctx, orders)
}

func TestHandleCreateOrder(t *testing.T) {
	t.Run("deve retornar 202 quando todas as orders sao publicadas", func(t *testing.T) {
		mock := &mockOrderProcessor{
			processFunc: func(ctx context.Context, orders []kafka.OrderEvent) []string {
				return nil
			},
		}
		h := handler.NewOrderHandler(mock)

		body := []byte(`[{"order_id":"ORD001","customer_id":"CUST001","amount":100.50}]`)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")

		h.HandleCreateOrder()(w, r)

		if w.Code != http.StatusAccepted {
			t.Errorf("esperava %d, recebeu %d", http.StatusAccepted, w.Code)
		}
	})

	t.Run("deve retornar 400 para JSON invalido", func(t *testing.T) {
		mock := &mockOrderProcessor{
			processFunc: func(ctx context.Context, orders []kafka.OrderEvent) []string {
				return nil
			},
		}
		h := handler.NewOrderHandler(mock)

		body := []byte(`not json`)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")

		h.HandleCreateOrder()(w, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("esperava %d, recebeu %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("deve retornar 207 quando algumas orders falham", func(t *testing.T) {
		mock := &mockOrderProcessor{
			processFunc: func(ctx context.Context, orders []kafka.OrderEvent) []string {
				return []string{testOrderID}
			},
		}
		h := handler.NewOrderHandler(mock)

		body := []byte(`[
			{"order_id":"ORD001","customer_id":"CUST001","amount":100.50},
			{"order_id":"ORD002","customer_id":"CUST002","amount":200.75}
		]`)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")

		h.HandleCreateOrder()(w, r)

		if w.Code != http.StatusMultiStatus {
			t.Errorf("esperava %d, recebeu %d", http.StatusMultiStatus, w.Code)
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatal("falha ao decodificar resposta:", err)
		}
		failed, ok := resp["failed"].([]interface{})
		if !ok || len(failed) != 1 || failed[0] != "ORD001" {
			t.Errorf("esperava failed=[ORD001], recebeu %v", resp["failed"])
		}
	})

	t.Run("deve retornar 503 quando todas as orders falham", func(t *testing.T) {
		mock := &mockOrderProcessor{
			processFunc: func(ctx context.Context, orders []kafka.OrderEvent) []string {
				return []string{testOrderID, "ORD002"}
			},
		}
		h := handler.NewOrderHandler(mock)

		body := []byte(`[
			{"order_id":"ORD001","customer_id":"CUST001","amount":100.50},
			{"order_id":"ORD002","customer_id":"CUST002","amount":200.75}
		]`)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")

		h.HandleCreateOrder()(w, r)

		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("esperava %d, recebeu %d", http.StatusServiceUnavailable, w.Code)
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatal("falha ao decodificar resposta:", err)
		}
		if resp["error"] != "kafka unavailable" {
			t.Errorf("esperava error='kafka unavailable', recebeu %v", resp["error"])
		}
	})

	t.Run("deve retornar 400 para corpo vazio", func(t *testing.T) {
		mock := &mockOrderProcessor{
			processFunc: func(ctx context.Context, orders []kafka.OrderEvent) []string {
				return nil
			},
		}
		h := handler.NewOrderHandler(mock)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/orders", http.NoBody)
		r.Header.Set("Content-Type", "application/json")

		h.HandleCreateOrder()(w, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("esperava %d, recebeu %d", http.StatusBadRequest, w.Code)
		}
	})
}

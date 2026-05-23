package router_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mrl00/kafka-event-driven-example/internal/handler"
	"github.com/mrl00/kafka-event-driven-example/internal/kafka"
	"github.com/mrl00/kafka-event-driven-example/internal/router"
)

type noopProcessor struct{}

func (n *noopProcessor) ProcessOrder(ctx context.Context, orders []kafka.OrderEvent) []string {
	return nil
}

func TestRouter_HealthRoute(t *testing.T) {
	orderHandler := handler.NewOrderHandler(&noopProcessor{})
	r := router.ProducerRouter(orderHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GET /health esperava %d, recebeu %d", http.StatusOK, w.Code)
	}

	body := strings.TrimSpace(w.Body.String())
	if body != "It's Working" {
		t.Errorf("body esperado 'It\\'s Working', recebeu '%s'", body)
	}
}

func TestRouter_NotFound(t *testing.T) {
	orderHandler := handler.NewOrderHandler(&noopProcessor{})
	r := router.ProducerRouter(orderHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/nonexistent", http.NoBody)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("GET /nonexistent esperava %d, recebeu %d", http.StatusNotFound, w.Code)
	}
}

func TestRouter_MethodNotAllowed(t *testing.T) {
	orderHandler := handler.NewOrderHandler(&noopProcessor{})
	r := router.ProducerRouter(orderHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/health", http.NoBody)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST /health esperava %d, recebeu %d", http.StatusMethodNotAllowed, w.Code)
	}
}

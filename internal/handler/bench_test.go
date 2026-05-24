package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mrl00/kafka-event-driven-example/internal/kafka"
)

const (
	testSingleOrderPayload = `[{"order_id":"ORD001","customer_id":"CUST001","amount":100.50}]`
	testMultiOrderPayload  = `[{"order_id":"ORD001","customer_id":"CUST001","amount":100.50},
		{"order_id":"ORD002","customer_id":"CUST002","amount":200.75},
		{"order_id":"ORD003","customer_id":"CUST003","amount":50.25}]`
)

type mockProcessor struct {
	failedIDs []string
}

func (m *mockProcessor) ProcessOrder(ctx context.Context, orders []kafka.OrderEvent) []string {
	return m.failedIDs
}

func BenchmarkHandleCreateOrder_Success(b *testing.B) {
	handler := NewOrderHandler(&mockProcessor{})

	b.ResetTimer()
	for b.Loop() {
		req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(testSingleOrderPayload))
		w := httptest.NewRecorder()
		handler.HandleCreateOrder()(w, req)
	}
}

func BenchmarkHandleCreateOrder_PartialFailure(b *testing.B) {
	handler := NewOrderHandler(&mockProcessor{failedIDs: []string{"ORD001"}})

	b.ResetTimer()
	for b.Loop() {
		req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(testMultiOrderPayload))
		w := httptest.NewRecorder()
		handler.HandleCreateOrder()(w, req)
	}
}

func BenchmarkHandleCreateOrder_AllFailed(b *testing.B) {
	handler := NewOrderHandler(&mockProcessor{failedIDs: []string{"ORD001", "ORD002", "ORD003"}})

	b.ResetTimer()
	for b.Loop() {
		req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(testMultiOrderPayload))
		w := httptest.NewRecorder()
		handler.HandleCreateOrder()(w, req)
	}
}

func BenchmarkHandleCreateOrder_InvalidJSON(b *testing.B) {
	handler := NewOrderHandler(&mockProcessor{})

	b.ResetTimer()
	for b.Loop() {
		req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader("not json"))
		w := httptest.NewRecorder()
		handler.HandleCreateOrder()(w, req)
	}
}

func BenchmarkHandleCreateOrder_MultipleOrders(b *testing.B) {
	handler := NewOrderHandler(&mockProcessor{})

	b.ResetTimer()
	for b.Loop() {
		req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(testMultiOrderPayload))
		w := httptest.NewRecorder()
		handler.HandleCreateOrder()(w, req)
	}
}

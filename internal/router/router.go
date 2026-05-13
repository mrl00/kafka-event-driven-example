package router

import (
	"net/http"

	"github.com/mrl00/kafka-event-driven-example/internal/handler"
)

func ProducerRouter(orderHandler *handler.OrderHandler) http.Handler {
	r := http.NewServeMux()

	r.HandleFunc("GET /health", handler.HealthCheck())

	r.HandleFunc("POST /orders", orderHandler.HandleCreateOrder())

	return r
}

func ConsumerRouter() http.Handler {
	r := http.NewServeMux()

	r.HandleFunc("GET /health", handler.HealthCheck())

	return r
}

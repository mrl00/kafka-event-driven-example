package handler_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mrl00/kafka-event-driven-example/internal/handler"
)

func Test_HealthCheck(t *testing.T) {
	t.Run("deve retornar 200 OK", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		handler.HealthCheck()(w, r)

		if w.Result().StatusCode != http.StatusOK {
			t.Errorf("esperava %d, recebeu %d", http.StatusOK, w.Result().StatusCode)
		}
	})

	t.Run("deve retornar 'It\\'s Working' no body", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		handler.HealthCheck()(w, r)

		body := strings.TrimSpace(w.Body.String())
		if body != "It's Working" {
			t.Errorf("esperava 'It\\'s Working', recebeu '%s'", body)
		}
	})

	t.Run("deve retornar Content-Type text/plain", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		handler.HealthCheck()(w, r)

		ct := w.Result().Header.Get("Content-Type")
		if ct == "" {
			t.Error("Content-Type header nao deveria estar vazio")
		}
		if !strings.HasPrefix(ct, "text/plain") {
			t.Errorf("Content-Type esperado text/plain, recebeu %s", ct)
		}
	})
}

package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mrl00/kafka-event-driven-example/internal/handler"
)

func Test_HealthCheck(t *testing.T) {
	testCases := []struct {
		name           string
		expectedStatus int
	}{
		{
			name:           "Ok",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)

			handler.HealthCheck()(w, r)

			if w.Result().StatusCode != tc.expectedStatus {
				t.Errorf("expected: %d, got: %d\n", tc.expectedStatus, w.Result().StatusCode)
			}
		})

	}

}

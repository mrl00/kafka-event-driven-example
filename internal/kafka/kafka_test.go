package kafka

import (
	"context"
	"errors"
	"testing"
)

func TestWrapRetryable(t *testing.T) {
	t.Run("deve retornar nil se o erro for nil", func(t *testing.T) {
		err := wrapRetryable(nil)
		if err != nil {
			t.Errorf("esperava nil, recebeu %v", err)
		}
	})

	t.Run("deve marcar erro como retryable", func(t *testing.T) {
		originalErr := errors.New("kafka connection failed")
		err := wrapRetryable(originalErr)

		rErr, ok := err.(kafkaError)
		if !ok {
			t.Fatal("erro deveria ser do tipo kafkaError")
		}

		if !rErr.Retryable() {
			t.Error("erro deveria ser retryable")
		}

		if rErr.Error() != originalErr.Error() {
			t.Errorf("mensagem de erro esperada %s, recebeu %s", originalErr.Error(), rErr.Error())
		}
	})
}

func TestKafkaConfig_GetBrokers(t *testing.T) {
	tests := []struct {
		name     string
		brokers  []string
		expected string
	}{
		{"lista vazia", []string{}, ""},
		{"um broker", []string{"localhost:9092"}, "localhost:9092"},
		{"múltiplos brokers", []string{"k1:9092", "k2:9092"}, "k1:9092,k2:9092"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := KafkaConfig{Brokers: tt.brokers}
			if got := cfg.GetBrokers(); got != tt.expected {
				t.Errorf("GetBrokers() = %v, esperava %v", got, tt.expected)
			}
		})
	}
}

// Nota: Testar NewProducer ou ProduceOrder de forma unitária completa
// exigiria um mock da interface do Producer do confluent-kafka-go.
// Este teste foca na lógica de Marshalling e tratamento de contexto.
func TestProduceOrder_ContextCancellation(t *testing.T) {
	// Criamos um contexto já cancelado
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// O ProduceOrder deve falhar imediatamente devido ao contexto
	err := ProduceOrder(ctx, nil, "test-topic", OrderEvent{OrderID: "123"})

	if err == nil {
		t.Error("esperava erro por contexto cancelado, recebeu nil")
	}
}


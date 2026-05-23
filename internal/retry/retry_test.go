package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fatalError struct{ error }

func (e fatalError) Retryable() bool { return false }

func TestDo(t *testing.T) {
	t.Run("deve retornar sucesso se a operação eventualmente funcionar", func(t *testing.T) {
		attempts := 0
		cfg := RetryConfig{
			MaxRetries:     3,
			InitialBackoff: 1 * time.Millisecond,
			Multiplier:     2.0,
		}

		err := Do(context.Background(), cfg, func(ctx context.Context) error {
			attempts++
			if attempts < 2 {
				return errors.New("erro temporário")
			}
			return nil
		})
		if err != nil {
			t.Errorf("esperava sucesso, recebeu erro: %v", err)
		}
		if attempts != 2 {
			t.Errorf("esperava 2 tentativas, mas foram %d", attempts)
		}
	})

	t.Run("deve parar imediatamente se receber um erro não-retryable", func(t *testing.T) {
		attempts := 0
		cfg := DefaultConfig()
		cfg.InitialBackoff = 1 * time.Millisecond

		err := Do(context.Background(), cfg, func(ctx context.Context) error {
			attempts++
			return fatalError{errors.New("erro fatal")}
		})

		if err == nil {
			t.Error("esperava um erro, mas recebeu nil")
		}
		if attempts != 1 {
			t.Errorf("esperava apenas 1 tentativa, mas foram %d", attempts)
		}
	})

	t.Run("deve respeitar o cancelamento do contexto durante o backoff", func(t *testing.T) {
		cfg := RetryConfig{
			MaxRetries:     5,
			InitialBackoff: 1 * time.Second, // Longo para forçar a espera
			Multiplier:     2.0,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		start := time.Now()
		err := Do(ctx, cfg, func(ctx context.Context) error {
			//nolint:misspell // Portuguese word, intentional
			return errors.New("falha persistente")
		})

		duration := time.Since(start)

		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("esperava erro de timeout, recebeu: %v", err)
		}

		if duration > 500*time.Millisecond {
			t.Errorf("o retry demorou demais para cancelar: %v", duration)
		}
	})
}

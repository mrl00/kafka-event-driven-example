package retry

import (
	"context"
	"log/slog"
	"math/rand"
	"time"
)

type RetryConfig struct {
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	Multiplier     float64
	Jitter         bool
}

func DefaultConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:     5,
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     10 * time.Second,
		Multiplier:     2.0,
		Jitter:         true,
	}
}

type RetryableError interface {
	error
	Retryable() bool
}

func Do(ctx context.Context, cfg RetryConfig, operation func(ctx context.Context) error) error {
	var lastErr error
	backoff := cfg.InitialBackoff

	for attempt := 1; attempt <= cfg.MaxRetries; attempt++ {
		lastErr = operation(ctx)
		if lastErr == nil {
			return nil
		}

		if rErr, ok := lastErr.(RetryableError); ok && !rErr.Retryable() {
			slog.DebugContext(ctx, "erro fatal detectado, interrompendo retry")
			return lastErr
		}

		if attempt == cfg.MaxRetries {
			break
		}

		sleepTime := backoff
		if cfg.Jitter {
			ratio := 0.8 + (rand.Float64() * 0.4)
			sleepTime = time.Duration(float64(backoff) * ratio)
		}

		slog.WarnContext(ctx, "failed operation, trying again", "attempt", attempt, "next_retry_in", sleepTime, "error", lastErr)

		select {
		case <-time.After(sleepTime):
			backoff = min(time.Duration(float64(backoff)*cfg.Multiplier), cfg.MaxBackoff)
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return lastErr
}

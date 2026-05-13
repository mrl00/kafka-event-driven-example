package lifecycle

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func WaitForShutdownSignal(ctx context.Context, cancel context.CancelFunc, shutdownTimeout time.Duration, cleanups ...func()) {
	sigCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-sigCtx.Done()
	slog.Info("Shutdown signal received, starting graceful shutdown...", "timeout", shutdownTimeout)

	cancel()

	done := make(chan struct{})
	go func() {
		for i := len(cleanups) - 1; i >= 0; i-- {
			cleanups[i]()
		}
		close(done)
	}()

	select {
	case <-done:
		slog.Info("Shutdown complete")
	case <-time.After(shutdownTimeout):
		slog.Warn("Shutdown timed out, forcing exit", "timeout", shutdownTimeout)
	}
}

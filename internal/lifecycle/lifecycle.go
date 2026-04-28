package lifecycle

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func WaitForShutdownSignal(ctx context.Context, cancel context.CancelFunc, cleanups ...func()) {
	sigCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-sigCtx.Done()
	slog.Info("Shutting down...")

	cancel()

	timeoutStr := os.Getenv("SHUTDOWN_TIMEOUT")
	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		timeout = 5 * time.Second
	}

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
	case <-time.After(timeout):
		slog.Info("Shutdown timed out")
	}
}

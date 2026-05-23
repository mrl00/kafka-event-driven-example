package lifecycle_test

import (
	"context"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/mrl00/kafka-event-driven-example/internal/lifecycle"
)

func TestWaitForShutdownSignal(t *testing.T) {
	t.Run("deve executar cleanups em ordem reversa ao receber SIGTERM", func(t *testing.T) {
		var order []int64
		var orderMu int64

		cleanup1 := func() {
			atomic.AddInt64(&orderMu, 1)
			order = append(order, 1)
		}
		cleanup2 := func() {
			atomic.AddInt64(&orderMu, 1)
			order = append(order, 2)
		}
		cleanup3 := func() {
			atomic.AddInt64(&orderMu, 1)
			order = append(order, 3)
		}

		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			time.Sleep(50 * time.Millisecond)
			_ = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		}()

		done := make(chan struct{})
		go func() {
			lifecycle.WaitForShutdownSignal(ctx, cancel, 5*time.Second, cleanup1, cleanup2, cleanup3)
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatal("shutdown timeout")
		}

		if len(order) != 3 {
			t.Fatalf("esperava 3 cleanups, executou %d", len(order))
		}
		if order[0] != 3 || order[1] != 2 || order[2] != 1 {
			t.Errorf("cleanups nao executados em ordem reversa: %v", order)
		}
	})

	t.Run("deve respeitar o timeout de shutdown", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		blockingCleanup := func() {
			time.Sleep(1 * time.Second)
		}

		go func() {
			time.Sleep(50 * time.Millisecond)
			_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		}()

		done := make(chan struct{})
		go func() {
			lifecycle.WaitForShutdownSignal(ctx, cancel, 100*time.Millisecond, blockingCleanup)
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatal("shutdown nao retornou apos timeout")
		}
	})

	t.Run("deve cancelar o contexto ao receber sinal", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			time.Sleep(50 * time.Millisecond)
			_ = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		}()

		done := make(chan struct{})
		go func() {
			lifecycle.WaitForShutdownSignal(ctx, cancel, 5*time.Second)
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatal("shutdown timeout")
		}

		if ctx.Err() == nil {
			t.Error("contexto deveria estar cancelado")
		}
	})
}

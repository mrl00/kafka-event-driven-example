package server

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/mrl00/kafka-event-driven-example/internal/router"
)

type Config struct {
	Name              string
	Port              string
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ReadHeaderTimeout time.Duration
}

func StartServer(cfg Config) *http.Server {
	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		MaxHeaderBytes:    1 << 20, // 1MB
		Handler:           router.New(),
	}

	go func() {
		slog.Info("Started "+cfg.Name+" server", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Failed to start "+cfg.Name+" server", "err", err)
		}
	}()

	return srv
}

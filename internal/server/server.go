package server

import (
	"log/slog"
	"net/http"
	"time"
)

type Config struct {
	Name              string
	Port              string
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ReadHeaderTimeout time.Duration
}

func StartServer(cfg Config, r http.Handler) *http.Server {
	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		MaxHeaderBytes:    1 << 20, // 1MB
		Handler:           r,
	}

	go func() {
		slog.Info("Started "+cfg.Name+" server", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Failed to start "+cfg.Name+" server", "err", err)
		}
	}()

	return srv
}

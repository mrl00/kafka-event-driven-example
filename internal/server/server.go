package server

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/mrl00/kafka-event-driven-example/internal/router"
)

func StartServer(name string, port string) *http.Server {
	srv := &http.Server{
		Addr:              ":" + port,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       15 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		Handler:           router.New(),
	}

	go func() {
		slog.Info("Started "+name+" server", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Failed to start "+name+" server", "err", err)
		}
	}()

	return srv
}

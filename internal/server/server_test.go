package server

import (
	"net/http"
	"testing"
	"time"
)

const testServerName = "test"

func TestStartServer_ConfigApplied(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	srv := StartServer(Config{
		Name:              testServerName,
		Port:              "0",
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}, handler)

	defer func() {
		_ = srv.Close()
	}()

	if srv.Addr != ":0" {
		t.Errorf("Addr esperado ':0', recebeu %s", srv.Addr)
	}
	if srv.ReadTimeout != 5*time.Second {
		t.Errorf("ReadTimeout esperado 5s, recebeu %v", srv.ReadTimeout)
	}
	if srv.WriteTimeout != 10*time.Second {
		t.Errorf("WriteTimeout esperado 10s, recebeu %v", srv.WriteTimeout)
	}
	if srv.IdleTimeout != 120*time.Second {
		t.Errorf("IdleTimeout esperado 120s, recebeu %v", srv.IdleTimeout)
	}
	if srv.ReadHeaderTimeout != 2*time.Second {
		t.Errorf("ReadHeaderTimeout esperado 2s, recebeu %v", srv.ReadHeaderTimeout)
	}
}

func TestStartServer_StartsAndCloses(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	srv := StartServer(Config{
		Name: "test",
		Port: "0",
	}, handler)

	if err := srv.Close(); err != nil {
		t.Fatalf("Close falhou: %v", err)
	}
}

func TestStartServer_DefaultMaxHeaderBytes(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	srv := StartServer(Config{
		Name: "test",
		Port: "0",
	}, handler)
	defer func() {
		_ = srv.Close()
	}()

	if srv.MaxHeaderBytes != 1<<20 {
		t.Errorf("MaxHeaderBytes esperado %d, recebeu %d", 1<<20, srv.MaxHeaderBytes)
	}
}

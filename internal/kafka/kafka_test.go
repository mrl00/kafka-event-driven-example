package kafka_test

import (
	"testing"

	"github.com/mrl00/kafka-event-driven-example/internal/kafka"
)

func TestGetBrokers(t *testing.T) {
	cfg := kafka.Config{
		Brokers: []string{"1", "2", "3"},
		Topic:   "asd",
		GroupID: "asd",
	}

	t.Run("testing concat", func(t *testing.T) {
		got := cfg.GetBrokers()
		if got != "123" {
			t.Error("getBrokers must return 123")
		}
	})
}

package appconfig

import (
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestLoadConfig_AllDefaults(t *testing.T) {
	t.Setenv("KAFKA_BROKERS", "localhost:9092")
	t.Setenv("KAFKA_TOPIC", "")
	t.Setenv("KAFKA_GROUP_ID", "")
	t.Setenv("KAFKA_NUM_PARTITIONS", "")
	t.Setenv("KAFKA_REPLICATION_FACTOR", "")
	t.Setenv("HTTP_PORT", "")
	t.Setenv("DEMO_MODE", "")

	cfg := LoadProducerConfig()

	if len(cfg.Brokers) != 1 || cfg.Brokers[0] != "localhost:9092" {
		t.Errorf("Brokers esperado [localhost:9092], recebeu %v", cfg.Brokers)
	}
	if cfg.Topic != "orders" {
		t.Errorf("Topic esperado 'orders', recebeu %s", cfg.Topic)
	}
	if cfg.GroupID != "" {
		t.Errorf("GroupID esperado '', recebeu %s", cfg.GroupID)
	}
	if cfg.NumOfPartitions != 3 {
		t.Errorf("NumOfPartitions esperado 3, recebeu %d", cfg.NumOfPartitions)
	}
	if cfg.ReplicationFactor != 3 {
		t.Errorf("ReplicationFactor esperado 3, recebeu %d", cfg.ReplicationFactor)
	}
	if cfg.HTTPPort != "4000" {
		t.Errorf("HTTPPort esperado '4000', recebeu %s", cfg.HTTPPort)
	}
	if cfg.DemoMode {
		t.Error("DemoMode esperado false")
	}
	if cfg.ReadTimeout != 15*time.Second {
		t.Errorf("ReadTimeout esperado 15s, recebeu %v", cfg.ReadTimeout)
	}
}

func TestLoadConfig_CustomValues(t *testing.T) {
	t.Setenv("KAFKA_BROKERS", "broker1:9092,broker2:9092")
	t.Setenv("KAFKA_TOPIC", "my-topic")
	t.Setenv("KAFKA_GROUP_ID", "my-group")
	t.Setenv("KAFKA_NUM_PARTITIONS", "5")
	t.Setenv("KAFKA_REPLICATION_FACTOR", "2")
	t.Setenv("HTTP_PORT", "8080")
	t.Setenv("HTTP_READ_TIMEOUT", "10s")
	t.Setenv("HTTP_WRITE_TIMEOUT", "20s")
	t.Setenv("HTTP_IDLE_TIMEOUT", "30s")
	t.Setenv("HTTP_READ_HEADER_TIMEOUT", "3s")
	t.Setenv("SHUTDOWN_TIMEOUT", "45s")
	t.Setenv("DEMO_MODE", "true")

	cfg := LoadProducerConfig()

	if len(cfg.Brokers) != 2 {
		t.Errorf("Brokers esperado 2, recebeu %d", len(cfg.Brokers))
	}
	if cfg.Topic != "my-topic" {
		t.Errorf("Topic esperado 'my-topic', recebeu %s", cfg.Topic)
	}
	if cfg.GroupID != "" {
		t.Errorf("Producer GroupID esperado '', recebeu %s", cfg.GroupID)
	}
	if cfg.NumOfPartitions != 5 {
		t.Errorf("NumOfPartitions esperado 5, recebeu %d", cfg.NumOfPartitions)
	}
	if cfg.ReplicationFactor != 2 {
		t.Errorf("ReplicationFactor esperado 2, recebeu %d", cfg.ReplicationFactor)
	}
	if cfg.HTTPPort != "8080" {
		t.Errorf("HTTPPort esperado '8080', recebeu %s", cfg.HTTPPort)
	}
	if !cfg.DemoMode {
		t.Error("DemoMode esperado true")
	}
	if cfg.ReadTimeout != 10*time.Second {
		t.Errorf("ReadTimeout esperado 10s, recebeu %v", cfg.ReadTimeout)
	}
	if cfg.WriteTimeout != 20*time.Second {
		t.Errorf("WriteTimeout esperado 20s, recebeu %v", cfg.WriteTimeout)
	}
	if cfg.IdleTimeout != 30*time.Second {
		t.Errorf("IdleTimeout esperado 30s, recebeu %v", cfg.IdleTimeout)
	}
	if cfg.ReadHeaderTimeout != 3*time.Second {
		t.Errorf("ReadHeaderTimeout esperado 3s, recebeu %v", cfg.ReadHeaderTimeout)
	}
	if cfg.ShutdownTimeout != 45*time.Second {
		t.Errorf("ShutdownTimeout esperado 45s, recebeu %v", cfg.ShutdownTimeout)
	}
}

func TestLoadConfig_MissingBrokers(t *testing.T) {
	if os.Getenv("TEST_MISSING_BROKERS") == "1" {
		os.Unsetenv("KAFKA_BROKERS")
		LoadProducerConfig()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestLoadConfig_MissingBrokers")
	var filteredEnv []string
	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, "KAFKA_BROKERS=") {
			filteredEnv = append(filteredEnv, e)
		}
	}
	cmd.Env = append(filteredEnv, "TEST_MISSING_BROKERS=1")

	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatal("process should have exited with non-zero status")
}

func TestLoadConfig_ConsumerGroupID(t *testing.T) {
	t.Setenv("KAFKA_BROKERS", "localhost:9092")
	t.Setenv("KAFKA_GROUP_ID", "")

	producerCfg := LoadProducerConfig()
	if producerCfg.GroupID != "" {
		t.Errorf("Producer GroupID esperado '', recebeu %s", producerCfg.GroupID)
	}

	consumerCfg := LoadConsumerConfig()
	if consumerCfg.GroupID != "order-consumer-group" {
		t.Errorf("Consumer GroupID esperado 'order-consumer-group', recebeu %s", consumerCfg.GroupID)
	}

	t.Setenv("KAFKA_GROUP_ID", "custom-group")
	consumerCfg2 := LoadConsumerConfig()
	if consumerCfg2.GroupID != "custom-group" {
		t.Errorf("Consumer GroupID custom esperado 'custom-group', recebeu %s", consumerCfg2.GroupID)
	}
}

func TestLoadConfig_InvalidNumPartitions(t *testing.T) {
	t.Setenv("KAFKA_BROKERS", "localhost:9092")
	t.Setenv("KAFKA_NUM_PARTITIONS", "not-a-number")

	cfg := LoadProducerConfig()

	if cfg.NumOfPartitions != 3 {
		t.Errorf("NumOfPartitions esperado 3 (default para valor invalido), recebeu %d", cfg.NumOfPartitions)
	}
}

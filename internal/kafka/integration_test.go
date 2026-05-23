//go:build integration

package kafka

// testOrderEvent uses different fields (Item, Quantity, Price) from the
// production OrderEvent (CustomerID, Amount) because this test sends raw
// JSON through Kafka — it validates the wire format, not the domain type.
import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	ckafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/testcontainers/testcontainers-go"
	tckafka "github.com/testcontainers/testcontainers-go/modules/kafka"
)

var testBroker string

func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	container, err := tckafka.Run(ctx, "confluentinc/confluent-local:7.9.0",
		tckafka.WithClusterID("test-cluster"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start kafka container: %v\n", err)
		os.Exit(1)
	}

	brokers, err := container.Brokers(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get brokers: %v\n", err)
		os.Exit(1)
	}
	testBroker = brokers[0]

	code := m.Run()

	if err := container.Terminate(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "failed to terminate container: %v\n", err)
	}
	os.Exit(code)
}

type testOrderEvent struct {
	OrderID  string  `json:"order_id"`
	Item     string  `json:"item"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

func createTopic(t *testing.T, topic string) {
	t.Helper()
	admin, err := ckafka.NewAdminClient(&ckafka.ConfigMap{
		"bootstrap.servers": testBroker,
	})
	if err != nil {
		t.Fatalf("failed to create admin client: %v", err)
	}
	defer admin.Close()

	ctx := context.Background()
	_, err = admin.CreateTopics(ctx,
		[]ckafka.TopicSpecification{{
			Topic:         topic,
			NumPartitions: 1,
		}},
	)
	if err != nil {
		t.Fatalf("failed to create topic: %v", err)
	}
}

func TestIntegration_ProduceAndConsume(t *testing.T) {
	topic := "test-integration-topic"
	createTopic(t, topic)

	producer, err := ckafka.NewProducer(&ckafka.ConfigMap{
		"bootstrap.servers": testBroker,
	})
	if err != nil {
		t.Fatalf("failed to create producer: %v", err)
	}
	defer producer.Close()

	order := testOrderEvent{
		OrderID:  "int-001",
		Item:     "integration-test-item",
		Quantity: 3,
		Price:    29.99,
	}

	value, err := json.Marshal(order)
	if err != nil {
		t.Fatalf("failed to marshal order: %v", err)
	}

	deliveryChan := make(chan ckafka.Event, 1)
	err = producer.Produce(&ckafka.Message{
		TopicPartition: ckafka.TopicPartition{Topic: &topic, Partition: ckafka.PartitionAny},
		Value:          value,
	}, deliveryChan)
	if err != nil {
		t.Fatalf("failed to produce message: %v", err)
	}

	ev := <-deliveryChan
	msg, ok := ev.(*ckafka.Message)
	if !ok {
		t.Fatalf("unexpected delivery event type: %T", ev)
	}
	if msg.TopicPartition.Error != nil {
		t.Fatalf("delivery error: %v", msg.TopicPartition.Error)
	}
	t.Logf("Produced to partition %d at offset %d", msg.TopicPartition.Partition, msg.TopicPartition.Offset)

	consumer, err := ckafka.NewConsumer(&ckafka.ConfigMap{
		"bootstrap.servers":  testBroker,
		"group.id":           "test-integration-group",
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": false,
	})
	if err != nil {
		t.Fatalf("failed to create consumer: %v", err)
	}
	defer consumer.Close()

	if err := consumer.Subscribe(topic, nil); err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}

	msg, err = consumer.ReadMessage(30 * time.Second)
	if err != nil {
		t.Fatalf("failed to read message: %v", err)
	}

	var consumed testOrderEvent
	if err := json.Unmarshal(msg.Value, &consumed); err != nil {
		t.Fatalf("failed to unmarshal message: %v", err)
	}

	if consumed.OrderID != order.OrderID {
		t.Errorf("expected order_id %q, got %q", order.OrderID, consumed.OrderID)
	}
	if consumed.Item != order.Item {
		t.Errorf("expected item %q, got %q", order.Item, consumed.Item)
	}
	if consumed.Quantity != order.Quantity {
		t.Errorf("expected quantity %d, got %d", order.Quantity, consumed.Quantity)
	}
	if consumed.Price != order.Price {
		t.Errorf("expected price %.2f, got %.2f", order.Price, consumed.Price)
	}
}

func TestIntegration_TopicCreationIdempotent(t *testing.T) {
	admin, err := ckafka.NewAdminClient(&ckafka.ConfigMap{
		"bootstrap.servers": testBroker,
	})
	if err != nil {
		t.Fatalf("failed to create admin client: %v", err)
	}
	defer admin.Close()

	ctx := context.Background()
	results, err := admin.CreateTopics(ctx,
		[]ckafka.TopicSpecification{{
			Topic:             "test-idempotent-topic",
			NumPartitions:     1,
			ReplicationFactor: 1,
		}},
	)
	if err != nil {
		t.Fatalf("first creation failed: %v", err)
	}
	t.Logf("First creation result: %s (code=%d)", results[0].Topic, results[0].Error.Code())

	results, err = admin.CreateTopics(ctx,
		[]ckafka.TopicSpecification{{
			Topic:             "test-idempotent-topic",
			NumPartitions:     1,
			ReplicationFactor: 1,
		}},
	)
	if err != nil {
		t.Fatalf("second creation failed: %v", err)
	}
	if code := results[0].Error.Code(); code != ckafka.ErrTopicAlreadyExists && code != ckafka.ErrNoError {
		t.Errorf("unexpected error code on duplicate creation: %d", code)
	}
	t.Logf("Second creation result: %s (code=%d)", results[0].Topic, results[0].Error.Code())
}

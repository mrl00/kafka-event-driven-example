package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/mrl00/kafka-event-driven-example/internal/retry"
)

type kafkaError struct {
	err       error
	retryable bool
}

func (e kafkaError) Error() string {
	return e.err.Error()
}

func (e kafkaError) Retryable() bool {
	return e.retryable
}

func wrapRetryable(err error) error {
	if err == nil {
		return nil
	}
	return kafkaError{err: err, retryable: true}
}

type KafkaConfig struct {
	Brokers           []string
	Topic             string
	GroupID           string
	NumOfPartitions   int
	ReplicationFactor int
}

func (c KafkaConfig) GetBrokers() string {
	if len(c.Brokers) == 0 {
		return ""
	}
	var brokers strings.Builder
	for i, b := range c.Brokers {
		if i > 0 {
			brokers.WriteString(",")
		}
		brokers.WriteString(b)
	}
	return brokers.String()
}

type OrderEvent struct {
	OrderID    string    `json:"order_id"`
	CustomerID string    `json:"customer_id"`
	Amount     float64   `json:"amount"`
	CreatedAt  time.Time `json:"created_at"`
}

func NewProducer(cfg KafkaConfig) (*kafka.Producer, error) {
	var p *kafka.Producer
	retryCfg := retry.DefaultConfig()

	err := retry.Do(context.Background(), retryCfg, func(ctx context.Context) error {
		var err error
		p, err = kafka.NewProducer(&kafka.ConfigMap{
			"bootstrap.servers":  cfg.GetBrokers(),
			"acks":               "all",
			"retries":            5,
			"enable.idempotence": true,
		})

		if err != nil {
			return wrapRetryable(fmt.Errorf("failed to create producer: %w", err))
		}
		return nil
	})

	return p, err
}

func ProduceOrder(ctx context.Context, producer *kafka.Producer, topic string, order OrderEvent) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	message, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	retryCfg := retry.DefaultConfig()
	retryCfg.MaxRetries = 3

	return retry.Do(ctx, retryCfg, func(c context.Context) error {
		deliveryChan := make(chan kafka.Event)
		defer close(deliveryChan)

		err = producer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Value:          message,
		}, deliveryChan)

		if err != nil {
			return wrapRetryable(fmt.Errorf("failed to produce message: %w", err))
		}

		select {
		case <-c.Done():
			return c.Err()
		case ev := <-deliveryChan:
			m := ev.(*kafka.Message)
			if m.TopicPartition.Error != nil {
				// Se for um erro de "Leader Not Available", o retry vai ajudar!
				return wrapRetryable(fmt.Errorf("delivery failed: %w", m.TopicPartition.Error))
			}
			slog.InfoContext(c, "Mensagem enviada com sucesso",
				"partition", m.TopicPartition.Partition,
				"offset", m.TopicPartition.Offset)
			return nil
		}
	})
}

func NewConsumer(ctx context.Context, cfg KafkaConfig) (*kafka.Consumer, error) {
	var c *kafka.Consumer
	retryCfg := retry.DefaultConfig()

	err := retry.Do(ctx, retryCfg, func(ctx context.Context) error {
		var err error
		c, err = kafka.NewConsumer(&kafka.ConfigMap{
			"bootstrap.servers":  cfg.GetBrokers(),
			"group.id":           cfg.GroupID,
			"auto.offset.reset":  "earliest",
			"enable.auto.commit": true})
		if err != nil {
			return wrapRetryable(fmt.Errorf("cannot create consumer: %w", err))
		}

		err = c.SubscribeTopics([]string{cfg.Topic}, nil)
		if err != nil {
			c.Close()
			return wrapRetryable(fmt.Errorf("failed to subscribe topic %s: %w", cfg.Topic, err))
		}

		return nil
	})

	return c, err
}

func ConsumeOrders(ctx context.Context, consumer *kafka.Consumer) error {
	slog.InfoContext(ctx, "Iniciando loop de consumo de ordens")

	for {
		select {
		case <-ctx.Done():
			slog.InfoContext(ctx, "Parando consumo (contexto cancelado)")
			return ctx.Err()
		default:
			msg, err := consumer.ReadMessage(1 * time.Second)
			if err != nil {
				if kerr, ok := err.(kafka.Error); ok && kerr.Code() == kafka.ErrTimedOut {
					continue
				}
				slog.ErrorContext(ctx, "Erro ao ler mensagem", "error", err)
				continue
			}

			var order OrderEvent
			if err := json.Unmarshal(msg.Value, &order); err != nil {
				slog.WarnContext(ctx, "Mensagem inválida ignorada", "error", err)
				continue
			}

			slog.InfoContext(ctx, "Ordem consumida", "order_id", order.OrderID)
		}
	}
}

func EnsureTopic(ctx context.Context, cfg KafkaConfig) error {
	retryCfg := retry.DefaultConfig()

	return retry.Do(ctx, retryCfg, func(ctx context.Context) error {

		admin, err := kafka.NewAdminClient(&kafka.ConfigMap{
			"bootstrap.servers": cfg.GetBrokers(),
		})
		if err != nil {
			return wrapRetryable(fmt.Errorf("failed to create admin client: %w", err))
		}
		defer admin.Close()

		maxDur := 30 * time.Second

		topicSpec := []kafka.TopicSpecification{
			{
				Topic:             cfg.Topic,
				NumPartitions:     cfg.NumOfPartitions,
				ReplicationFactor: cfg.ReplicationFactor,
			},
		}

		results, err := admin.CreateTopics(ctx, topicSpec, kafka.SetAdminOperationTimeout(maxDur))
		if err != nil {
			return fmt.Errorf("failed to create topic: %w", err)
		}

		for _, result := range results {
			if result.Error.Code() != kafka.ErrNoError && result.Error.Code() != kafka.ErrTopicAlreadyExists {
				return fmt.Errorf("topic creation error for %s: %v", result.Topic, result.Error)
			}
		}

		return nil
	})
}

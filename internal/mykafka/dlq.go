package mykafka

import (
	"context"
	"fmt"
	"strconv"
	"time"

	ckafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type DLQProducer struct {
	producer *ckafka.Producer
	topic    string
}

func NewDLQProducer(cfg KafkaConfig) (*DLQProducer, error) {
	p, err := NewProducer(cfg)
	if err != nil {
		return nil, err
	}

	topic := cfg.Topic + ".dlq"
	return &DLQProducer{
		producer: p,
		topic:    topic,
	}, nil
}

func (d *DLQProducer) Send(ctx context.Context, originalMsg *ckafka.Message, errStr string, errorType string) error {
	headers := []ckafka.Header{
		{Key: "dlq.error", Value: []byte(errStr)},
		{Key: "dlq.error_type", Value: []byte(errorType)},
		{Key: "dlq.original_topic", Value: []byte(*originalMsg.TopicPartition.Topic)},
		{Key: "dlq.original_partition", Value: []byte(strconv.Itoa(int(originalMsg.TopicPartition.Partition)))},
		{Key: "dlq.original_offset", Value: []byte(originalMsg.TopicPartition.Offset.String())},
		{Key: "dlq.timestamp", Value: []byte(time.Now().Format(time.RFC3339))},
	}

	dlqMsg := &ckafka.Message{
		TopicPartition: ckafka.TopicPartition{
			Topic:     &d.topic,
			Partition: ckafka.PartitionAny,
		},
		Key:     originalMsg.Key,
		Value:   originalMsg.Value,
		Headers: headers,
	}

	deliveryChan := make(chan ckafka.Event)
	defer close(deliveryChan)

	err := d.producer.Produce(dlqMsg, deliveryChan)
	if err != nil {
		return fmt.Errorf("failed to send message to dlq: %w", err)
	}

	ev := <-deliveryChan
	m := ev.(*ckafka.Message)
	return m.TopicPartition.Error
}

func (d *DLQProducer) Close() {
	if d.producer != nil {
		d.producer.Close()
	}
}

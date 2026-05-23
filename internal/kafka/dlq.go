package kafka

import (
	"context"
	"fmt"
	"strconv"
	"time"

	ckafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type DLQProducer struct {
	producer Producer
	topic    string
}

func NewDLQProducer(cfg Config) (*DLQProducer, error) {
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

func NewDLQProducerWithProducer(producer Producer, topic string) *DLQProducer {
	return &DLQProducer{producer: producer, topic: topic + ".dlq"}
}

func (d *DLQProducer) buildHeaders(originalMsg *ckafka.Message, errStr string, errorType string, retryCount int) []ckafka.Header {
	topic := ""
	if originalMsg.TopicPartition.Topic != nil {
		topic = *originalMsg.TopicPartition.Topic
	}
	return []ckafka.Header{
		{Key: "dlq.error", Value: []byte(errStr)},
		{Key: "dlq.error_type", Value: []byte(errorType)},
		{Key: "dlq.original_topic", Value: []byte(topic)},
		{Key: "dlq.original_partition", Value: []byte(strconv.Itoa(int(originalMsg.TopicPartition.Partition)))},
		{Key: "dlq.original_offset", Value: []byte(originalMsg.TopicPartition.Offset.String())},
		{Key: "dlq.timestamp", Value: []byte(time.Now().Format(time.RFC3339))},
		{Key: "dlq.retry_count", Value: []byte(strconv.Itoa(retryCount))},
	}
}

func (d *DLQProducer) Send(ctx context.Context, originalMsg *ckafka.Message, errStr string, errorType string, retryCount int) error {
	headers := d.buildHeaders(originalMsg, errStr, errorType, retryCount)

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

package mykafka

import (
	"testing"
)

func TestDLQTopicNaming(t *testing.T) {
	cfg := KafkaConfig{Topic: "orders"}
	dlq, _ := NewDLQProducer(cfg)
	// O fechamento do producer aqui falharia pois não há broker real,
	// mas testamos apenas a lógica do nome do tópico
	if dlq.topic != "orders.dlq" {
		t.Errorf("esperava orders.dlq, recebeu %s", dlq.topic)
	}
}

package kafkaexporter

import (
	"github.com/IBM/sarama"
)

type MessageSender interface {
	SendMessages(msgs []*sarama.ProducerMessage) error
}

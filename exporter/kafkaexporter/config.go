package kafkaexporter

import (
	"github.com/IBM/sarama"
)

type Settings struct {
	Topic     string
	Addresses []string
	*sarama.Config
}

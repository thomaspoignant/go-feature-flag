package kafkaexporter

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/IBM/sarama"
	"github.com/thomaspoignant/go-feature-flag/exporter"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

const (
	formatJSON = "json"
)

// MessageSender is a Kafka producer that implements the SendMessages method
type MessageSender interface {
	SendMessages(msgs []*sarama.ProducerMessage) error
}

// Settings contains Kafka-specific configurations needed for message creation
type Settings struct {
	Topic     string
	Addresses []string
	*sarama.Config
}

// Exporter sends events to a Kafka topic using a synchronous producer
type Exporter struct {
	// Format is the output format for the message value.
	// The only available format right now is JSON, and this field provided for future usage.
	// Default: JSON
	Format string

	// Settings contains the Kafka producer's configuration. The Topic and Addresses fields are required. If
	// no sarama.Config is provided a sensible default will be used.
	Settings Settings

	init   sync.Once
	sender MessageSender
	// dialer will create the producer. This field is added for dependency injection during testing as sarama
	// has the annoying tendency to dial as soon as a producer is created.
	dialer func(addrs []string, config *sarama.Config) (MessageSender, error)
}

// Export will produce a message to the Kafka topic. The message's value will contain the event encoded in the
// selected format. Messages are publish synchronously and will error immediately on failure.
func (e *Exporter) Export(_ context.Context, logger *log.Logger, featureEvents []exporter.FeatureEvent) error {
	if e.sender == nil {
		err := e.initializeProducer()
		if err != nil {
			return fmt.Errorf("writer: %w", err)
		}
	}

	messages := make([]*sarama.ProducerMessage, len(featureEvents))
	for _, event := range featureEvents {
		data, err := e.formatMessage(event)
		if err != nil {
			return fmt.Errorf("format: %w", err)
		}

		messages = append(messages, &sarama.ProducerMessage{
			Topic: e.Settings.Topic,
			Key:   sarama.StringEncoder(event.UserKey),
			Value: sarama.ByteEncoder(data),
		})
	}

	err := e.sender.SendMessages(messages)
	if err != nil {
		return fmt.Errorf("send: %w", err)
	}

	fflog.Printf(logger, "info: [KafkaExporter] sent %d messages", len(messages))
	return nil
}

// IsBulk reports if the producer can handle bulk messages. Will always return false for this exporter.
func (e *Exporter) IsBulk() bool {
	return false
}

// initializeProducer runs only once and creates a new producer from the dialer. If the config is not populated a new
// one will be created with sensible defaults.
func (e *Exporter) initializeProducer() error {
	var err error
	e.init.Do(func() {
		if e.Settings.Config == nil {
			e.Settings.Config = sarama.NewConfig()
			e.Settings.Config.Producer.Return.Successes = true // Needs to be true for sync producers
		}

		if e.dialer == nil {
			e.dialer = func(addrs []string, config *sarama.Config) (MessageSender, error) {
				// Adapter for the function to comply with the MessageSender interface return
				return sarama.NewSyncProducer(addrs, config)
			}
		}

		e.sender, err = e.dialer(e.Settings.Addresses, e.Settings.Config)
		if err != nil {
			err = fmt.Errorf("producer: %w", err)
			return
		}
	})

	return err
}

// formatMessage returns the event encoded in the selected format. Will always use JSON for now.
func (e *Exporter) formatMessage(event exporter.FeatureEvent) ([]byte, error) {
	switch e.Format {
	case formatJSON:
		fallthrough
	default:
		return json.Marshal(event)
	}
}

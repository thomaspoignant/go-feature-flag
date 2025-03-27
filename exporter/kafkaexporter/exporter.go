package kafkaexporter

import (
	"context"
	"encoding/json"
	"fmt"

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

type AsyncMessageSender struct {
	sarama.AsyncProducer
	logger *fflog.FFLogger
}

func (a *AsyncMessageSender) SendMessages(msgs []*sarama.ProducerMessage) error {
	for len(msgs) > 0 {
		select {
		case err := <-a.AsyncProducer.Errors():
			a.logger.Warn("Failed to produce message: %w", err)
		case a.AsyncProducer.Input() <- msgs[0]:
			msgs = msgs[1:]
		}
	}
	return nil
}

// Settings contains Kafka-specific configurations needed for message creation
type Settings struct {
	Topic     string   `json:"topic"`
	Addresses []string `json:"addresses"`
	Async     bool     `json:"async"`
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

	sender MessageSender
	// dialer will create the producer. This field is added for dependency injection during testing as sarama
	// has the annoying tendency to dial as soon as a producer is created.
	dialer func(addrs []string, config *sarama.Config) (MessageSender, error)

	logger *fflog.FFLogger
}

// Export will produce a message to the Kafka topic. The message's value will contain the event encoded in the
// selected format. Messages are published synchronously and will error immediately on failure.
func (e *Exporter) Export(_ context.Context, logger *fflog.FFLogger, featureEvents []exporter.FeatureEvent) error {
	if e.sender == nil {
		err := e.initializeProducer(logger)
		if err != nil {
			return fmt.Errorf("writer: %w", err)
		}
	}

	messages := make([]*sarama.ProducerMessage, 0, len(featureEvents))
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
	return nil
}

// IsBulk reports if the producer can handle bulk messages. Will always return false for this exporter.
func (e *Exporter) IsBulk() bool {
	return false
}

// initializeProducer runs only once and creates a new producer from the dialer. If the config is not populated a new
// one will be created with sensible defaults.
func (e *Exporter) initializeProducer(logger *fflog.FFLogger) error {
	if e.Settings.Config == nil {
		e.Settings.Config = sarama.NewConfig()
		e.Settings.Config.Producer.Return.Successes = true // Needs to be true for sync producers
	}

	if e.dialer == nil {
		e.dialer = func(addrs []string, config *sarama.Config) (MessageSender, error) {
			// Adapter for the function to comply with the MessageSender interface return
			if e.Settings.Async {
				asyncProducer, err := sarama.NewAsyncProducer(addrs, config)
				if err != nil {
					return nil, err
				}
				return &AsyncMessageSender{AsyncProducer: asyncProducer, logger: logger}, nil //TODO Close should be called on shutdown
			}
			return sarama.NewSyncProducer(addrs, config)
		}
	}

	var err error
	e.sender, err = e.dialer(e.Settings.Addresses, e.Settings.Config)
	if err != nil {
		err = fmt.Errorf("producer: %w", err)
		return err
	}

	return nil
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

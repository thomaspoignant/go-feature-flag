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

type Exporter struct {
	// Format is the output format you want in your exported file.
	// Available format are JSON, CSV and Parquet.
	// Default: JSON
	Format string

	Settings Settings

	init   sync.Once
	sender MessageSender
	dialer func(addrs []string, config *sarama.Config) (MessageSender, error)
}

// Export is saving a collection of events in a file.
func (e *Exporter) Export(_ context.Context, logger *log.Logger, featureEvents []exporter.FeatureEvent) error {
	if e.sender == nil {
		err := e.initializeWriter()
		if err != nil {
			return fmt.Errorf("writer: %w", err)
		}
	}

	var messages []*sarama.ProducerMessage
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

func (e *Exporter) IsBulk() bool {
	return false
}

func (e *Exporter) initializeWriter() error {
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

func (e *Exporter) formatMessage(event exporter.FeatureEvent) ([]byte, error) {
	switch e.Format {
	case formatJSON:
		fallthrough
	default:
		return json.Marshal(event)
	}
}

package kafkaexporter

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"text/template"

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

	sender      MessageSender
	init        sync.Once
	csvTemplate *template.Template
}

// Export is saving a collection of events in a file.
func (f *Exporter) Export(_ context.Context, logger *log.Logger, featureEvents []exporter.FeatureEvent) error {
	if f.sender == nil {
		err := f.initializeWriter()
		if err != nil {
			return fmt.Errorf("writer: %w", err)
		}
	}

	var messages []*sarama.ProducerMessage
	for _, event := range featureEvents {
		data, err := f.formatMessage(event)
		if err != nil {
			return fmt.Errorf("format: %w", err)
		}

		messages = append(messages, &sarama.ProducerMessage{
			Topic: f.Settings.Topic,
			Key:   sarama.StringEncoder(event.UserKey),
			Value: sarama.ByteEncoder(data),
		})
	}

	err := f.sender.SendMessages(messages)
	if err != nil {
		return fmt.Errorf("send: %w", err)
	}

	fflog.Printf(logger, "info: [KafkaExporter] sent %d messages", len(messages))
	return nil
}

func (f *Exporter) IsBulk() bool {
	return false
}

func (f *Exporter) initializeWriter() error {
	var err error
	f.init.Do(func() {
		if f.Settings.Config == nil {
			err = errors.New("writer configuration not provided")
			return
		}

		f.sender, err = sarama.NewSyncProducer(f.Settings.Addresses, f.Settings.Config)
		if err != nil {
			err = fmt.Errorf("producer: %w", err)
			return
		}
	})

	return err
}

func (f *Exporter) formatMessage(event exporter.FeatureEvent) ([]byte, error) {
	switch f.Format {
	case formatJSON:
		fallthrough
	default:
		return exporter.FormatEventInJSON(event)
	}
}

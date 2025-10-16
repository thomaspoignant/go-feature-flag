package pubsubexporter

import (
	"context"
	"encoding/json"

	"cloud.google.com/go/pubsub" //nolint:staticcheck
	"github.com/thomaspoignant/go-feature-flag/exporter"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
	"google.golang.org/api/option"
)

// Exporter publishes events on a PubSub topic.
//
// Deprecated: Use pubsubexporterv2.Exporter instead. This exporter uses the legacy
// cloud.google.com/go/pubsub v1 library. The v2 library provides improved performance
// and additional features. This exporter will be removed in a future version.
type Exporter struct {
	// ProjectID is a project to which the PubSub topic belongs.
	ProjectID string

	// Topic is the name of a topic on which messages will be published.
	Topic string

	// Options are Google Cloud API options to connect to PubSub.
	Options []option.ClientOption

	// PublishSettings controls the bundling of published messages.
	// If not set pubsub.DefaultPublishSettings are used.
	PublishSettings *pubsub.PublishSettings

	// EnableMessageOrdering enables the delivery of ordered keys.
	EnableMessageOrdering bool

	// newClientFunc is used only for unit testing purposes.
	newClientFunc func(context.Context, string, ...option.ClientOption) (*pubsub.Client, error)

	// publisher facilitates publishing messages on a PubSub topic.
	publisher *pubsub.Topic
}

// Export publishes a PubSub message for each exporter.FeatureEvent received.
func (e *Exporter) Export(
	ctx context.Context,
	_ *fflog.FFLogger,
	events []exporter.ExportableEvent,
) error {
	if e.publisher == nil {
		if err := e.initPublisher(ctx); err != nil {
			return err
		}
	}

	for _, event := range events {
		messageBody, err := json.Marshal(event)
		if err != nil {
			return err
		}

		_, err = e.publisher.Publish(ctx, &pubsub.Message{
			Data:       messageBody,
			Attributes: map[string]string{"emitter": "GO Feature Flag"},
		}).Get(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

// IsBulk always returns false as PubSub exporter sends each exporter.FeatureEvent as a separate message.
func (e *Exporter) IsBulk() bool {
	return false
}

// initPublisher inits PubSub topic publisher according to the provided configuration.
func (e *Exporter) initPublisher(ctx context.Context) error {
	if e.newClientFunc == nil {
		e.newClientFunc = pubsub.NewClient
	}

	client, err := e.newClientFunc(ctx, e.ProjectID, e.Options...)
	if err != nil {
		return err
	}

	topic := client.Topic(e.Topic)
	if e.PublishSettings != nil {
		topic.PublishSettings = *e.PublishSettings
	}
	topic.EnableMessageOrdering = e.EnableMessageOrdering

	e.publisher = topic
	return nil
}

package pubsubexporter

import (
	"context"
	"encoding/json"
	"log"

	"cloud.google.com/go/pubsub"
	"github.com/thomaspoignant/go-feature-flag/exporter"
	"google.golang.org/api/option"
)

// Exporter sends events to a PubSub topic.
type Exporter struct {
	// ProjectID is a project in which the PubSub topic exists.
	ProjectID string

	// Topic is a name of topic to which messages will be sent.
	Topic string

	// Options are Google Cloud Api options to connect to PubSub.
	Options []option.ClientOption

	// PublishSettings control the bundling of published messages.
	// If not set pubsub.DefaultPublishSettings are used.
	PublishSettings *pubsub.PublishSettings

	// EnableMessageOrdering enables delivery of ordered keys.
	EnableMessageOrdering bool

	// newClientFunc  used only for unit testing purposes.
	newClientFunc func(context.Context, string, ...option.ClientOption) (*pubsub.Client, error)

	// publisher allows for publishing messages on PubSub topic.
	publisher *pubsub.Topic
}

// Export sends PubSub message for each exporter.FeatureEvent received.
func (e *Exporter) Export(ctx context.Context, _ *log.Logger, featureEvents []exporter.FeatureEvent) error {
	if e.publisher == nil {
		if err := e.initPublisher(ctx); err != nil {
			return err
		}
	}

	for _, event := range featureEvents {
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

// IsBulk always returns false as PubSub exporter sends each exporter.FeatureEvent as separate message.
func (e *Exporter) IsBulk() bool {
	return false
}

// initPublisher inits PubSub topic publisher according to provided configuration.
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

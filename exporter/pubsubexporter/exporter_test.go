package pubsubexporter

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/pubsub/pstest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/exporter"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestExporter_Export(t *testing.T) {
	const (
		projectID = "fake-project"
		topic     = "fake-topic"
	)

	ctx := context.TODO()
	logger := &fflog.FFLogger{LeveledLogger: slog.Default()}

	server := pstest.NewServer()
	t.Cleanup(func() { server.Close() })

	conn, err := grpc.NewClient(
		server.Addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	client, err := pubsub.NewClient(ctx, projectID, option.WithGRPCConn(conn))
	require.NoError(t, err)
	t.Cleanup(func() { client.Close() })

	_, err = client.CreateTopic(ctx, topic)
	require.NoError(t, err)

	defaultNewClientFunc := func(_ context.Context, _ string, _ ...option.ClientOption) (*pubsub.Client, error) {
		return client, nil
	}

	type fields struct {
		projectID             string
		topic                 string
		options               []option.ClientOption
		publishSettings       *pubsub.PublishSettings
		enableMessageOrdering bool
		newClientFunc         func(context.Context, string, ...option.ClientOption) (*pubsub.Client, error)
	}
	tests := []struct {
		name    string
		fields  fields
		events  []exporter.ExportableEvent
		wantErr bool
	}{
		{
			name: "should publish a single message with the feature event",
			fields: fields{
				topic:         topic,
				newClientFunc: defaultNewClientFunc,
			},
			events: []exporter.ExportableEvent{
				exporter.FeatureEvent{
					Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false,
				},
			},
		},
		{
			name: "should publish multiple messages with feature events",
			fields: fields{
				topic:         topic,
				newClientFunc: defaultNewClientFunc,
			},
			events: []exporter.ExportableEvent{
				exporter.FeatureEvent{
					Kind: "feature1", ContextKind: "anonymousUser1", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key1",
					Variation: "Default", Value: "YO", Default: false,
				},
				exporter.FeatureEvent{
					Kind: "feature2", ContextKind: "anonymousUser2", UserKey: "ABCDEF", CreationDate: 1617970527, Key: "random-key2",
					Variation: "Default", Value: "YO", Default: true,
				},
			},
		},
		{
			name: "should use the provided client options when creating a client",
			fields: fields{
				topic:   topic,
				options: []option.ClientOption{option.WithAPIKey("some-api-key")},
				newClientFunc: func(_ context.Context, _ string, opts ...option.ClientOption) (*pubsub.Client, error) {
					if len(opts) != 1 {
						return nil, errors.New("not expected number of options")
					} else if opts[0] != option.WithAPIKey("some-api-key") {
						return nil, errors.New("unexpected option provided")
					}
					return client, nil
				},
			},
			events: []exporter.ExportableEvent{
				exporter.FeatureEvent{
					Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false,
				},
			},
		},
		{
			name: "should use the provided publisher settings for the PubSub topic",
			fields: fields{
				topic:           topic,
				newClientFunc:   defaultNewClientFunc,
				publishSettings: &pubsub.PublishSettings{CountThreshold: 123},
			},
			events: []exporter.ExportableEvent{
				exporter.FeatureEvent{
					Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false,
				},
			},
		},
		{
			name: "should enable message ordering if the configuration is set",
			fields: fields{
				topic:                 topic,
				newClientFunc:         defaultNewClientFunc,
				enableMessageOrdering: true,
			},
			events: []exporter.ExportableEvent{
				exporter.FeatureEvent{
					Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false,
				},
			},
		},
		{
			name: "should return an error if there is a problem with creating a PubSub client",
			fields: fields{
				topic: topic,
				newClientFunc: func(_ context.Context, _ string, _ ...option.ClientOption) (*pubsub.Client, error) {
					return nil, errors.New("errored client")
				},
			},
			wantErr: true,
		},
		{
			name: "should return an error if publishing a message fails",
			fields: fields{
				topic:         "not-existing-topic",
				newClientFunc: defaultNewClientFunc,
			},
			events: []exporter.ExportableEvent{
				exporter.FeatureEvent{
					Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(server.ClearMessages)

			e := &Exporter{
				ProjectID:             tt.fields.projectID,
				Topic:                 tt.fields.topic,
				Options:               tt.fields.options,
				PublishSettings:       tt.fields.publishSettings,
				EnableMessageOrdering: tt.fields.enableMessageOrdering,
				newClientFunc:         tt.fields.newClientFunc,
			}
			err = e.Export(ctx, logger, tt.events)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assertMessages(t, tt.events, server.Messages())
			assertPublisherSettings(t, tt.fields.publishSettings, e.publisher)
			assert.Equal(t, tt.fields.enableMessageOrdering, e.publisher.EnableMessageOrdering)
		})
	}
}

func TestExporter_IsBulk(t *testing.T) {
	t.Parallel()

	e := &Exporter{}

	assert.False(t, e.IsBulk(), "PubSub exporter is not a bulk one")
}

func assertMessages(
	t *testing.T,
	expectedEvents []exporter.ExportableEvent,
	messages []*pstest.Message,
) {
	events := make([]exporter.FeatureEvent, len(messages))
	for i, message := range messages {
		assert.Equal(t, map[string]string{"emitter": "GO Feature Flag"}, message.Attributes,
			"message should have associated emitter attribute")

		var event exporter.FeatureEvent
		err := json.Unmarshal(message.Data, &event)
		assert.NoError(t, err)

		events[i] = event
	}
	assert.ElementsMatchf(t, expectedEvents, events, "events should match in any order")
}

func assertPublisherSettings(
	t *testing.T,
	expectedSettings *pubsub.PublishSettings,
	publisher *pubsub.Topic,
) {
	if expectedSettings != nil {
		assert.Equal(t, *expectedSettings, publisher.PublishSettings)
	} else {
		assert.Equal(t, pubsub.DefaultPublishSettings, publisher.PublishSettings)
	}
}

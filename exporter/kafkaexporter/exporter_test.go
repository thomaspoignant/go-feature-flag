package kafkaexporter

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/exporter"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

type messageSenderMock struct {
	messages []*sarama.ProducerMessage
	error    error
}

func (s *messageSenderMock) SendMessages(msgs []*sarama.ProducerMessage) error {
	s.messages = append(s.messages, msgs...)
	return s.error
}

func TestExporter_IsBulk(t *testing.T) {
	exp := Exporter{}
	assert.False(t, exp.IsBulk(), "DeprecatedExporterV1 is not a bulk exporter")
}

func TestExporter_Export(t *testing.T) {
	const mockTopic = "mockTopic"

	tests := []struct {
		name     string
		format   string
		dialer   func(addrs []string, config *sarama.Config) (MessageSender, error)
		events   []exporter.ExportableEvent
		wantErr  bool
		settings Settings
	}{
		{
			name:    "should receive an error if dial failed",
			format:  "json",
			wantErr: true,
			settings: Settings{
				Topic:     mockTopic,
				Addresses: []string{"addr1", "addr2"},
			},
			dialer: func(_ []string, _ *sarama.Config) (MessageSender, error) {
				return nil, errors.New("dial error")
			},
		},
		{
			name:    "should use default dialer when none provided",
			format:  "json",
			wantErr: true, // The default dialer should error
			settings: Settings{
				Topic:     mockTopic,
				Addresses: []string{"addr1", "addr2"},
			},
			dialer: nil,
		},
		{
			name:    "should receive an event with a valid feature event",
			format:  "json",
			wantErr: false,
			events: []exporter.ExportableEvent{
				exporter.FeatureEvent{
					Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false,
				},
				exporter.FeatureEvent{
					Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCDEF", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false,
				},
			},
			settings: Settings{
				Topic:     mockTopic,
				Addresses: []string{"addr1", "addr2"},
			},
			dialer: func(_ []string, _ *sarama.Config) (MessageSender, error) {
				return &messageSenderMock{}, nil
			},
		},
		{
			name:    "should default to JSON format if none provided",
			format:  "", // Should default to JSON and generate a valid message
			wantErr: false,
			events: []exporter.ExportableEvent{
				exporter.FeatureEvent{
					Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false,
				},
				exporter.FeatureEvent{
					Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCDEF", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false,
				},
			},
			settings: Settings{
				Topic:     mockTopic,
				Addresses: []string{"addr1", "addr2"},
			},
			dialer: func(_ []string, _ *sarama.Config) (MessageSender, error) {
				return &messageSenderMock{}, nil
			},
		},
		{
			name:    "should return an error if the publisher is returning an error",
			format:  "json",
			wantErr: true,
			events: []exporter.ExportableEvent{
				exporter.FeatureEvent{
					Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false,
				},
				exporter.FeatureEvent{
					Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCDEF", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false,
				},
			},
			settings: Settings{
				Topic:     mockTopic,
				Addresses: []string{"addr1", "addr2"},
				Config:    &sarama.Config{},
			},
			dialer: func(_ []string, _ *sarama.Config) (MessageSender, error) {
				return &messageSenderMock{
					error: errors.New("failure to send message: datacenter on fire"),
				}, nil
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			exp := &Exporter{
				Format:   tt.format,
				Settings: tt.settings,
				dialer:   tt.dialer,
			}

			logger := &fflog.FFLogger{LeveledLogger: slog.Default()}
			err := exp.Export(context.Background(), logger, tt.events)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			want := make([]*sarama.ProducerMessage, len(tt.events))
			for index, event := range tt.events {
				messageBody, _ := json.Marshal(event)
				want[index] = &sarama.ProducerMessage{
					Topic: mockTopic,
					Key:   sarama.StringEncoder(event.GetUserKey()),
					Value: sarama.ByteEncoder(messageBody),
				}
			}

			assert.Equal(t, want, (exp.sender).(*messageSenderMock).messages)
		})
	}
}

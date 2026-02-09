package exporter_test

import (
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/exporter"
)

func TestTrackingEvent_FormatInCSV(t *testing.T) {
	tests := []struct {
		name          string
		trackingEvent *exporter.TrackingEvent
		template      string
		want          string
		wantErr       assert.ErrorAssertionFunc
	}{
		{
			name: "Should return a marshalled JSON string of the tracking event",
			trackingEvent: &exporter.TrackingEvent{
				Kind:              "tracking",
				ContextKind:       "anonymousUser",
				UserKey:           "ABCD",
				CreationDate:      1617970547,
				Key:               "random-key",
				EvaluationContext: map[string]any{"targetingKey": "ABCD"},
				TrackingDetails: map[string]any{
					"event": "123",
				},
			},
			template: `{{ .Kind}};{{ .ContextKind}};{{ .UserKey}};{{ .CreationDate}};{{ .EvaluationContext}};{{ .TrackingDetails}}`,
			want:     `tracking;anonymousUser;ABCD;1617970547;map[targetingKey:ABCD];map[event:123]`,
			wantErr:  assert.NoError,
		},
		{
			name: "Should return a marshalled JSON string of the tracking event with evaluation context attributes",
			trackingEvent: &exporter.TrackingEvent{
				Kind:              "tracking",
				ContextKind:       "anonymousUser",
				UserKey:           "ABCD",
				CreationDate:      1617970547,
				Key:               "random-key",
				EvaluationContext: map[string]any{"targetingKey": "ABCD", "toto": 123},
				TrackingDetails: map[string]any{
					"event": "123",
				},
			},
			template: `{{ .Kind}};{{ .ContextKind}};{{ .UserKey}};{{ .CreationDate}};{{ .EvaluationContext}};{{ .TrackingDetails}}`,
			want:     `tracking;anonymousUser;ABCD;1617970547;map[targetingKey:ABCD toto:123];map[event:123]`,
			wantErr:  assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			csvTemplate, err := template.New("test").Parse(tt.template)
			assert.NoError(t, err)
			got, err := tt.trackingEvent.FormatInCSV(csvTemplate)
			tt.wantErr(t, err)
			if err == nil {
				assert.Equal(t, tt.want, string(got))
			}
		})
	}
}

func TestTrackingEvent_FormatInJSON(t *testing.T) {
	tests := []struct {
		name          string
		trackingEvent *exporter.TrackingEvent
		want          string
		wantErr       assert.ErrorAssertionFunc
	}{
		{
			name: "Should return a marshalled JSON string of the tracking event",
			trackingEvent: &exporter.TrackingEvent{
				Kind:              "tracking",
				ContextKind:       "anonymousUser",
				UserKey:           "ABCD",
				CreationDate:      1617970547,
				Key:               "random-key",
				EvaluationContext: map[string]any{"targetingKey": "ABCD"},
				TrackingDetails: map[string]any{
					"event": "123",
				},
			},
			want:    `{"kind":"tracking","contextKind":"anonymousUser","userKey":"ABCD","creationDate":1617970547,"key":"random-key","evaluationContext":{"targetingKey":"ABCD"},"trackingEventDetails":{"event":"123"}}`,
			wantErr: assert.NoError,
		},
		{
			name: "Should return a marshalled JSON string of the tracking event with evaluation context attributes",
			trackingEvent: &exporter.TrackingEvent{
				Kind:              "tracking",
				ContextKind:       "anonymousUser",
				UserKey:           "ABCD",
				CreationDate:      1617970547,
				Key:               "random-key",
				EvaluationContext: map[string]any{"targetingKey": "ABCD", "toto": 123},
				TrackingDetails: map[string]any{
					"event": "123",
				},
			},
			want:    `{"kind":"tracking","contextKind":"anonymousUser","userKey":"ABCD","creationDate":1617970547,"key":"random-key","evaluationContext":{"targetingKey":"ABCD","toto":123},"trackingEventDetails":{"event":"123"}}`,
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.trackingEvent.FormatInJSON()
			tt.wantErr(t, err)
			if err == nil {
				assert.JSONEq(t, tt.want, string(got))
			}
		})
	}
}

func TestTrackingEvent_GetKey(t *testing.T) {
	tests := []struct {
		name          string
		trackingEvent *exporter.TrackingEvent
		want          string
	}{
		{
			name: "return existing key",
			trackingEvent: &exporter.TrackingEvent{
				Kind:         "tracking",
				ContextKind:  "anonymousUser",
				UserKey:      "ABCD",
				CreationDate: 1617970547,
				Key:          "random-key",
			},
			want: "random-key",
		},
		{
			name: "empty key",
			trackingEvent: &exporter.TrackingEvent{
				Kind:         "tracking",
				ContextKind:  "anonymousUser",
				UserKey:      "ABCD",
				CreationDate: 1617970547,
				Key:          "",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.trackingEvent.GetKey())
		})
	}
}

func TestTrackingEvent_GetUserKey(t *testing.T) {
	tests := []struct {
		name          string
		trackingEvent *exporter.TrackingEvent
		want          string
	}{
		{
			name: "return existing key",
			trackingEvent: &exporter.TrackingEvent{
				Kind:         "tracking",
				ContextKind:  "anonymousUser",
				UserKey:      "ABCD",
				CreationDate: 1617970547,
				Key:          "random-key",
			},
			want: "ABCD",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.trackingEvent.GetUserKey())
		})
	}
}

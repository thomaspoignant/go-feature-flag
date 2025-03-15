package exporter_test

import (
	"fmt"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/exporter"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
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
				EvaluationContext: ffcontext.NewEvaluationContextBuilder("ABCD").Build(),
				TrackingDetails: map[string]interface{}{
					"event": "123",
				},
			},
			template: `{{ .Kind}};{{ .ContextKind}};{{ .UserKey}};{{ .CreationDate}};{{ .EvaluationContext}};{{ .TrackingDetails}}`,
			want:     `tracking;anonymousUser;ABCD;1617970547;{ABCD map[]};map[event:123]`,
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
				EvaluationContext: ffcontext.NewEvaluationContextBuilder("ABCD").AddCustom("toto", 123).Build(),
				TrackingDetails: map[string]interface{}{
					"event": "123",
				},
			},
			template: `{{ .Kind}};{{ .ContextKind}};{{ .UserKey}};{{ .CreationDate}};{{ .EvaluationContext}};{{ .TrackingDetails}}`,
			want:     `tracking;anonymousUser;ABCD;1617970547;{ABCD map[toto:123]};map[event:123]`,
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
				EvaluationContext: ffcontext.NewEvaluationContextBuilder("ABCD").Build(),
				TrackingDetails: map[string]interface{}{
					"event": "123",
				},
			},
			want:    `{"kind":"tracking","contextKind":"anonymousUser","userKey":"ABCD","creationDate":1617970547,"key":"random-key","evaluationContext":{"targetingKey":"ABCD","attributes":{}},"trackingEventDetails":{"event":"123"}}`,
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
				EvaluationContext: ffcontext.NewEvaluationContextBuilder("ABCD").AddCustom("toto", 123).Build(),
				TrackingDetails: map[string]interface{}{
					"event": "123",
				},
			},
			want:    `{"kind":"tracking","contextKind":"anonymousUser","userKey":"ABCD","creationDate":1617970547,"key":"random-key","evaluationContext":{"targetingKey":"ABCD","attributes":{"toto":123}},"trackingEventDetails":{"event":"123"}}`,
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.trackingEvent.FormatInJSON()
			tt.wantErr(t, err)
			if err == nil {
				fmt.Println(string(got))
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

package exporter_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/exporter"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
)

func TestNewFeatureEvent(t *testing.T) {
	type args struct {
		user      ffcontext.Context
		flagKey   string
		value     interface{}
		variation string
		failed    bool
		version   string
		source    string
	}
	tests := []struct {
		name string
		args args
		want exporter.FeatureEvent
	}{
		{
			name: "anonymous user",
			args: args{
				user:      ffcontext.NewEvaluationContextBuilder("ABCD").AddCustom("anonymous", true).Build(),
				flagKey:   "random-key",
				value:     "YO",
				variation: "Default",
				failed:    false,
				version:   "",
				source:    "SERVER",
			},
			want: exporter.FeatureEvent{
				Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: time.Now().Unix(), Key: "random-key",
				Variation: "Default", Value: "YO", Default: false, Source: "SERVER",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, exporter.NewFeatureEvent(tt.args.user, tt.args.flagKey, tt.args.value, tt.args.variation, tt.args.failed, tt.args.version, tt.args.source), "NewFeatureEvent(%v, %v, %v, %v, %v, %v, %V)", tt.args.user, tt.args.flagKey, tt.args.value, tt.args.variation, tt.args.failed, tt.args.version, tt.args.source)
		})
	}
}

func TestFeatureEvent_MarshalInterface(t *testing.T) {
	tests := []struct {
		name         string
		featureEvent *exporter.FeatureEvent
		want         *exporter.FeatureEvent
		wantErr      bool
	}{
		{
			name: "happy path",
			featureEvent: &exporter.FeatureEvent{
				Kind:         "feature",
				ContextKind:  "anonymousUser",
				UserKey:      "ABCD",
				CreationDate: 1617970547,
				Key:          "random-key",
				Variation:    "Default",
				Value: map[string]interface{}{
					"string": "string",
					"bool":   true,
					"float":  1.23,
					"int":    1,
				},
				Default: false,
			},
			want: &exporter.FeatureEvent{
				Kind:         "feature",
				ContextKind:  "anonymousUser",
				UserKey:      "ABCD",
				CreationDate: 1617970547,
				Key:          "random-key",
				Variation:    "Default",
				Value:        `{"bool":true,"float":1.23,"int":1,"string":"string"}`,
				Default:      false,
			},
		},
		{
			name: "marshal failed",
			featureEvent: &exporter.FeatureEvent{
				Kind:         "feature",
				ContextKind:  "anonymousUser",
				UserKey:      "ABCD",
				CreationDate: 1617970547,
				Key:          "random-key",
				Variation:    "Default",
				Value:        make(chan int),
				Default:      false,
			},
			wantErr: true,
		},
		{
			name:         "nil featureEvent",
			featureEvent: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.featureEvent.MarshalInterface(); (err != nil) != tt.wantErr {
				t.Errorf("FeatureEvent.MarshalInterface() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil {
				assert.Equal(t, tt.want, tt.featureEvent)
			}
		})
	}
}

func TestFeatureEvent_MarshalJSON(t *testing.T) {
	tests := []struct {
		name         string
		featureEvent *exporter.FeatureEvent
		want         string
		wantErr      assert.ErrorAssertionFunc
	}{
		{
			name: "Should not return a metadata field if metadata is empty",
			featureEvent: &exporter.FeatureEvent{
				Kind:         "feature",
				ContextKind:  "anonymousUser",
				UserKey:      "ABCD",
				CreationDate: 1617970547,
				Key:          "random-key",
				Variation:    "Default",
				Value: map[string]interface{}{
					"string": "string",
					"bool":   true,
					"float":  1.23,
					"int":    1,
				},
				Default:  false,
				Metadata: map[string]interface{}{},
			},
			want:    `{"kind":"feature","contextKind":"anonymousUser","userKey":"ABCD","creationDate":1617970547,"key":"random-key","variation":"Default","value":{"string":"string","bool":true,"float":1.23,"int":1},"default":false}`,
			wantErr: assert.NoError,
		},
		{
			name: "Should not return a metadata field if metadata is nil",
			featureEvent: &exporter.FeatureEvent{
				Kind:         "feature",
				ContextKind:  "anonymousUser",
				UserKey:      "ABCD",
				CreationDate: 1617970547,
				Key:          "random-key",
				Variation:    "Default",
				Value: map[string]interface{}{
					"string": "string",
					"bool":   true,
					"float":  1.23,
					"int":    1,
				},
				Default: false,
			},
			want:    `{"kind":"feature","contextKind":"anonymousUser","userKey":"ABCD","creationDate":1617970547,"key":"random-key","variation":"Default","value":{"string":"string","bool":true,"float":1.23,"int":1},"default":false}`,
			wantErr: assert.NoError,
		},
		{
			name: "Should return a metadata field if metadata is not empty",
			featureEvent: &exporter.FeatureEvent{
				Kind:         "feature",
				ContextKind:  "anonymousUser",
				UserKey:      "ABCD",
				CreationDate: 1617970547,
				Key:          "random-key",
				Variation:    "Default",
				Value: map[string]interface{}{
					"string": "string",
					"bool":   true,
					"float":  1.23,
					"int":    1,
				},
				Default: false,
				Metadata: map[string]interface{}{
					"metadata1": "metadata1",
					"metadata2": 24,
					"metadata3": true,
				},
			},
			want:    `{"kind":"feature","contextKind":"anonymousUser","userKey":"ABCD","creationDate":1617970547,"key":"random-key","variation":"Default","value":{"string":"string","bool":true,"float":1.23,"int":1},"default":false,"metadata":{"metadata1":"metadata1","metadata2":24,"metadata3":true}}`,
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.featureEvent)
			tt.wantErr(t, err)
			if err != nil {
				assert.JSONEq(t, tt.want, string(got))
			}
		})
	}
}

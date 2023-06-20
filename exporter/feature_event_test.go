package exporter_test

import (
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/exporter"
)

func TestNewFeatureEvent(t *testing.T) {
	type args struct {
		user      ffcontext.Context
		flagKey   string
		value     interface{}
		variation string
		failed    bool
		version   string
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
			},
			want: exporter.FeatureEvent{
				Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: time.Now().Unix(), Key: "random-key",
				Variation: "Default", Value: "YO", Default: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, exporter.NewFeatureEvent(tt.args.user, tt.args.flagKey, tt.args.value, tt.args.variation, tt.args.failed, tt.args.version), "NewFeatureEvent(%v, %v, %v, %v, %v, %v)", tt.args.user, tt.args.flagKey, tt.args.value, tt.args.variation, tt.args.failed, tt.args.version)
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

package exporter_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/exporter"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
)

func TestNewFeatureEvent(t *testing.T) {
	type args struct {
		user             ffcontext.Context
		flagKey          string
		value            any
		variation        string
		failed           bool
		version          string
		source           string
		exporterMetadata exporter.FeatureEventMetadata
	}
	tests := []struct {
		name string
		args args
		want exporter.FeatureEvent
	}{
		{
			name: "anonymous user",
			args: args{
				user: ffcontext.NewEvaluationContextBuilder("ABCD").
					AddCustom("anonymous", true).
					Build(),
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
			assert.Equalf(
				t,
				tt.want,
				exporter.NewFeatureEvent(
					tt.args.user,
					tt.args.flagKey,
					tt.args.value,
					tt.args.variation,
					tt.args.failed,
					tt.args.version,
					tt.args.source,
					tt.args.exporterMetadata,
				),
				"NewFeatureEvent(%v, %v, %v, %v, %v, %v, %V)",
				tt.args.user,
				tt.args.flagKey,
				tt.args.value,
				tt.args.variation,
				tt.args.failed,
				tt.args.version,
				tt.args.source,
			)
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
				Value: map[string]any{
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
			name: "nil value",
			featureEvent: &exporter.FeatureEvent{
				Kind:         "feature",
				ContextKind:  "anonymousUser",
				UserKey:      "ABCD",
				CreationDate: 1617970547,
				Key:          "random-key",
				Variation:    "Default",
				Value:        nil,
				Default:      false,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := tt.featureEvent.ConvertValueForParquet()
			if (err != nil) != tt.wantErr {
				t.Errorf("FeatureEvent.MarshalInterface() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil {
				assert.Equal(t, tt.want.Value, val)
			}
		})
	}
}

func TestFeatureEvent_FormatInJSON(t *testing.T) {
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
				Value: map[string]any{
					"string": "string",
					"bool":   true,
					"float":  1.23,
					"int":    1,
				},
				Default:  false,
				Metadata: map[string]any{},
			},
			want:    `{"kind":"feature","contextKind":"anonymousUser","userKey":"ABCD","creationDate":1617970547,"key":"random-key","variation":"Default","value":{"bool":true,"float":1.23,"int":1,"string":"string"},"default":false,"version":"","source":""}`,
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
				Value: map[string]any{
					"string": "string",
					"bool":   true,
					"float":  1.23,
					"int":    1,
				},
				Default: false,
			},
			want:    `{"kind":"feature","contextKind":"anonymousUser","userKey":"ABCD","creationDate":1617970547,"key":"random-key","variation":"Default","value":{"bool":true,"float":1.23,"int":1,"string":"string"},"default":false,"version":"","source":""}`,
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
				Value: map[string]any{
					"string": "string",
					"bool":   true,
					"float":  1.23,
					"int":    1,
				},
				Default: false,
				Metadata: map[string]any{
					"metadata1": "metadata1",
					"metadata2": 24,
					"metadata3": true,
				},
			},
			want:    `{"kind":"feature","contextKind":"anonymousUser","userKey":"ABCD","creationDate":1617970547,"key":"random-key","variation":"Default","value":{"bool":true,"float":1.23,"int":1,"string":"string"},"default":false,"version":"","source":"","metadata":{"metadata1":"metadata1","metadata2":24,"metadata3":true}}`,
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.featureEvent.FormatInJSON()
			tt.wantErr(t, err)
			if err == nil {
				assert.JSONEq(t, tt.want, string(got))
			}
		})
	}
}

func TestFeatureEvent_GetKey(t *testing.T) {
	tests := []struct {
		name         string
		featureEvent *exporter.FeatureEvent
		want         string
	}{
		{
			name: "return existing key",
			featureEvent: &exporter.FeatureEvent{
				Kind:         "feature",
				ContextKind:  "anonymousUser",
				UserKey:      "ABCD",
				CreationDate: 1617970547,
				Key:          "random-key",
				Variation:    "Default",
				Value: map[string]any{
					"string": "string",
					"bool":   true,
					"float":  1.23,
					"int":    1,
				},
				Default: false,
			},
			want: "random-key",
		},
		{
			name: "empty key",
			featureEvent: &exporter.FeatureEvent{
				Kind:         "feature",
				ContextKind:  "anonymousUser",
				UserKey:      "ABCD",
				CreationDate: 1617970547,
				Key:          "",
				Variation:    "Default",
				Value:        nil,
				Default:      false,
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.featureEvent.GetKey())
		})
	}
}

func TestFeatureEvent_GetUserKey(t *testing.T) {
	tests := []struct {
		name         string
		featureEvent *exporter.FeatureEvent ``
		want         string
	}{
		{
			name: "return existing key",
			featureEvent: &exporter.FeatureEvent{
				Kind:         "feature",
				ContextKind:  "anonymousUser",
				UserKey:      "ABCD",
				CreationDate: 1617970547,
				Key:          "random-key",
				Variation:    "Default",
				Value: map[string]any{
					"string": "string",
					"bool":   true,
					"float":  1.23,
					"int":    1,
				},
				Default: false,
			},
			want: "ABCD",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.featureEvent.GetUserKey())
		})
	}
}

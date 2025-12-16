package model_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/modules/core/model"
)

func TestVariationResultToJsonStr(t *testing.T) {
	tests := []struct {
		name     string
		result   any
		wantJSON string
	}{
		{
			name: "bool variation result with all fields",
			result: model.VariationResult[bool]{
				TrackEvents:   true,
				VariationType: flag.VariationSDKDefault,
				Failed:        false,
				Version:       "1.0.0",
				Reason:        flag.ReasonDefault,
				ErrorCode:     flag.ErrorCodeGeneral,
				ErrorDetails:  "test error",
				Value:         true,
				Cacheable:     true,
				Metadata:      map[string]any{"key": "value"},
			},
			wantJSON: `{"trackEvents":true,"variationType":"SdkDefault","failed":false,"version":"1.0.0","reason":"DEFAULT","errorCode":"GENERAL","errorDetails":"test error","value":true,"cacheable":true,"metadata":{"key":"value"}}`,
		},
		{
			name: "string variation result",
			result: model.VariationResult[string]{
				TrackEvents:   false,
				VariationType: "enabled",
				Failed:        false,
				Version:       "2.0.0",
				Reason:        flag.ReasonTargetingMatch,
				ErrorCode:     flag.ErrorCodeFlagNotFound,
				Value:         "test-value",
				Cacheable:     false,
			},
			wantJSON: `{"trackEvents":false,"variationType":"enabled","failed":false,"version":"2.0.0","reason":"TARGETING_MATCH","errorCode":"FLAG_NOT_FOUND","value":"test-value","cacheable":false}`,
		},
		{
			name: "int variation result",
			result: model.VariationResult[int]{
				TrackEvents:   true,
				VariationType: "disabled",
				Failed:        true,
				Version:       "3.0.0",
				Reason:        flag.ReasonError,
				ErrorCode:     flag.ErrorCodeTypeMismatch,
				ErrorDetails:  "type error",
				Value:         42,
				Cacheable:     true,
			},
			wantJSON: `{"trackEvents":true,"variationType":"disabled","failed":true,"version":"3.0.0","reason":"ERROR","errorCode":"TYPE_MISMATCH","errorDetails":"type error","value":42,"cacheable":true}`,
		},
		{
			name: "float64 variation result",
			result: model.VariationResult[float64]{
				TrackEvents:   false,
				VariationType: flag.VariationSDKDefault,
				Failed:        false,
				Version:       "4.0.0",
				Reason:        flag.ReasonSplit,
				ErrorCode:     flag.ErrorCodeParseError,
				Value:         3.14,
				Cacheable:     false,
			},
			wantJSON: `{"trackEvents":false,"variationType":"SdkDefault","failed":false,"version":"4.0.0","reason":"SPLIT","errorCode":"PARSE_ERROR","value":3.14,"cacheable":false}`,
		},
		{
			name: "map variation result",
			result: model.VariationResult[map[string]any]{
				TrackEvents:   true,
				VariationType: "enabled",
				Failed:        false,
				Version:       "5.0.0",
				Reason:        flag.ReasonTargetingMatchSplit,
				ErrorCode:     flag.ErrorCodeInvalidContext,
				Value:         map[string]any{"nested": "value", "number": 123},
				Cacheable:     true,
				Metadata:      map[string]any{"meta": "data"},
			},
			wantJSON: `{"trackEvents":true,"variationType":"enabled","failed":false,"version":"5.0.0","reason":"TARGETING_MATCH_SPLIT","errorCode":"INVALID_CONTEXT","value":{"nested":"value","number":123},"cacheable":true,"metadata":{"meta":"data"}}`,
		},
		{
			name: "slice variation result",
			result: model.VariationResult[[]any]{
				TrackEvents:   false,
				VariationType: "disabled",
				Failed:        false,
				Version:       "6.0.0",
				Reason:        flag.ReasonStatic,
				ErrorCode:     flag.ErrorCodeTargetingKeyMissing,
				Value:         []any{"item1", "item2", 123},
				Cacheable:     true,
			},
			wantJSON: `{"trackEvents":false,"variationType":"disabled","failed":false,"version":"6.0.0","reason":"STATIC","errorCode":"TARGETING_KEY_MISSING","value":["item1","item2",123],"cacheable":true}`,
		},
		{
			name: "variation result without optional fields",
			result: model.VariationResult[string]{
				TrackEvents:   true,
				VariationType: flag.VariationSDKDefault,
				Failed:        false,
				Version:       "",
				Reason:        flag.ReasonDisabled,
				ErrorCode:     flag.ErrorCodeProviderNotReady,
				Value:         "default",
				Cacheable:     false,
			},
			wantJSON: `{"trackEvents":true,"variationType":"SdkDefault","failed":false,"version":"","reason":"DISABLED","errorCode":"PROVIDER_NOT_READY","value":"default","cacheable":false}`,
		},
		{
			name: "variation result with empty metadata",
			result: model.VariationResult[bool]{
				TrackEvents:   true,
				VariationType: "enabled",
				Failed:        false,
				Version:       "1.0.0",
				Reason:        flag.ReasonOffline,
				ErrorCode:     flag.ErrorFlagConfiguration,
				Value:         false,
				Cacheable:     true,
				Metadata:      map[string]any{},
			},
			wantJSON: `{"trackEvents":true,"variationType":"enabled","failed":false,"version":"1.0.0","reason":"OFFLINE","errorCode":"FLAG_CONFIG","value":false,"cacheable":true}`,
		},
		{
			name: "variation result with unknown reason",
			result: model.VariationResult[int]{
				TrackEvents:   false,
				VariationType: flag.VariationSDKDefault,
				Failed:        true,
				Version:       "1.0.0",
				Reason:        flag.ReasonUnknown,
				ErrorCode:     flag.ErrorCodeGeneral,
				Value:         0,
				Cacheable:     false,
			},
			wantJSON: `{"trackEvents":false,"variationType":"SdkDefault","failed":true,"version":"1.0.0","reason":"UNKNOWN","errorCode":"GENERAL","value":0,"cacheable":false}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var jsonStr string
			switch v := tt.result.(type) {
			case model.VariationResult[bool]:
				jsonStr = v.ToJsonStr()
			case model.VariationResult[string]:
				jsonStr = v.ToJsonStr()
			case model.VariationResult[int]:
				jsonStr = v.ToJsonStr()
			case model.VariationResult[float64]:
				jsonStr = v.ToJsonStr()
			case model.VariationResult[map[string]any]:
				jsonStr = v.ToJsonStr()
			case model.VariationResult[[]any]:
				jsonStr = v.ToJsonStr()
			}

			// Verify JSON is valid and matches expected
			var gotJSON map[string]any
			err := json.Unmarshal([]byte(jsonStr), &gotJSON)
			assert.NoError(t, err, "JSON should be valid")

			var wantJSON map[string]any
			err = json.Unmarshal([]byte(tt.wantJSON), &wantJSON)
			assert.NoError(t, err, "Expected JSON should be valid")

			assert.Equal(t, wantJSON, gotJSON, "JSON should match expected")
		})
	}
}

func TestRawVarResultJSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		result   model.RawVarResult
		wantJSON string
	}{
		{
			name: "raw var result with all fields",
			result: model.RawVarResult{
				TrackEvents:   true,
				VariationType: flag.VariationSDKDefault,
				Failed:        false,
				Version:       "1.0.0",
				Reason:        flag.ReasonDefault,
				ErrorCode:     flag.ErrorCodeGeneral,
				ErrorDetails:  "test error",
				Value:         "test-value",
				Cacheable:     true,
				Metadata:      map[string]any{"key": "value"},
			},
			wantJSON: `{"trackEvents":true,"variationType":"SdkDefault","failed":false,"version":"1.0.0","reason":"DEFAULT","errorCode":"GENERAL","errorDetails":"test error","value":"test-value","cacheable":true,"metadata":{"key":"value"}}`,
		},
		{
			name: "raw var result with bool value",
			result: model.RawVarResult{
				TrackEvents:   false,
				VariationType: "enabled",
				Failed:        false,
				Version:       "2.0.0",
				Reason:        flag.ReasonTargetingMatch,
				ErrorCode:     flag.ErrorCodeFlagNotFound,
				Value:         true,
				Cacheable:     false,
			},
			wantJSON: `{"trackEvents":false,"variationType":"enabled","failed":false,"version":"2.0.0","reason":"TARGETING_MATCH","errorCode":"FLAG_NOT_FOUND","value":true,"cacheable":false}`,
		},
		{
			name: "raw var result with int value",
			result: model.RawVarResult{
				TrackEvents:   true,
				VariationType: "disabled",
				Failed:        true,
				Version:       "3.0.0",
				Reason:        flag.ReasonError,
				ErrorCode:     flag.ErrorCodeTypeMismatch,
				ErrorDetails:  "type error",
				Value:         42,
				Cacheable:     true,
			},
			wantJSON: `{"trackEvents":true,"variationType":"disabled","failed":true,"version":"3.0.0","reason":"ERROR","errorCode":"TYPE_MISMATCH","errorDetails":"type error","value":42,"cacheable":true}`,
		},
		{
			name: "raw var result with float64 value",
			result: model.RawVarResult{
				TrackEvents:   false,
				VariationType: flag.VariationSDKDefault,
				Failed:        false,
				Version:       "4.0.0",
				Reason:        flag.ReasonSplit,
				ErrorCode:     flag.ErrorCodeParseError,
				Value:         3.14,
				Cacheable:     false,
			},
			wantJSON: `{"trackEvents":false,"variationType":"SdkDefault","failed":false,"version":"4.0.0","reason":"SPLIT","errorCode":"PARSE_ERROR","value":3.14,"cacheable":false}`,
		},
		{
			name: "raw var result with map value",
			result: model.RawVarResult{
				TrackEvents:   true,
				VariationType: "enabled",
				Failed:        false,
				Version:       "5.0.0",
				Reason:        flag.ReasonTargetingMatchSplit,
				ErrorCode:     flag.ErrorCodeInvalidContext,
				Value:         map[string]any{"nested": "value", "number": 123},
				Cacheable:     true,
				Metadata:      map[string]any{"meta": "data"},
			},
			wantJSON: `{"trackEvents":true,"variationType":"enabled","failed":false,"version":"5.0.0","reason":"TARGETING_MATCH_SPLIT","errorCode":"INVALID_CONTEXT","value":{"nested":"value","number":123},"cacheable":true,"metadata":{"meta":"data"}}`,
		},
		{
			name: "raw var result with slice value",
			result: model.RawVarResult{
				TrackEvents:   false,
				VariationType: "disabled",
				Failed:        false,
				Version:       "6.0.0",
				Reason:        flag.ReasonStatic,
				ErrorCode:     flag.ErrorCodeTargetingKeyMissing,
				Value:         []any{"item1", "item2", 123},
				Cacheable:     true,
			},
			wantJSON: `{"trackEvents":false,"variationType":"disabled","failed":false,"version":"6.0.0","reason":"STATIC","errorCode":"TARGETING_KEY_MISSING","value":["item1","item2",123],"cacheable":true}`,
		},
		{
			name: "raw var result without optional fields",
			result: model.RawVarResult{
				TrackEvents:   true,
				VariationType: flag.VariationSDKDefault,
				Failed:        false,
				Version:       "",
				Reason:        flag.ReasonDisabled,
				ErrorCode:     flag.ErrorCodeProviderNotReady,
				Value:         "default",
				Cacheable:     false,
			},
			wantJSON: `{"trackEvents":true,"variationType":"SdkDefault","failed":false,"version":"","reason":"DISABLED","errorCode":"PROVIDER_NOT_READY","value":"default","cacheable":false}`,
		},
		{
			name: "raw var result with nil value",
			result: model.RawVarResult{
				TrackEvents:   false,
				VariationType: flag.VariationSDKDefault,
				Failed:        true,
				Version:       "1.0.0",
				Reason:        flag.ReasonUnknown,
				ErrorCode:     flag.ErrorCodeGeneral,
				Value:         nil,
				Cacheable:     false,
			},
			wantJSON: `{"trackEvents":false,"variationType":"SdkDefault","failed":true,"version":"1.0.0","reason":"UNKNOWN","errorCode":"GENERAL","value":null,"cacheable":false}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON serialization
			jsonBytes, err := json.Marshal(tt.result)
			assert.NoError(t, err, "JSON marshaling should succeed")

			// Verify JSON matches expected
			var gotJSON map[string]any
			err = json.Unmarshal(jsonBytes, &gotJSON)
			assert.NoError(t, err, "JSON should be valid")

			var wantJSON map[string]any
			err = json.Unmarshal([]byte(tt.wantJSON), &wantJSON)
			assert.NoError(t, err, "Expected JSON should be valid")

			assert.Equal(t, wantJSON, gotJSON, "JSON should match expected")

			// Test JSON deserialization
			// Note: JSON unmarshaling converts numbers to float64, so we compare JSON strings instead
			var deserialized model.RawVarResult
			err = json.Unmarshal(jsonBytes, &deserialized)
			assert.NoError(t, err, "JSON unmarshaling should succeed")

			// Re-marshal both to compare JSON representation (handles float64 vs int conversion)
			deserializedJSON, _ := json.Marshal(deserialized)
			originalJSON, _ := json.Marshal(tt.result)
			assert.Equal(t, string(originalJSON), string(deserializedJSON), "Deserialized JSON should match original")
		})
	}
}

func TestVariationResultJSONDeserialization(t *testing.T) {
	tests := []struct {
		name    string
		jsonStr string
		want    model.VariationResult[string]
		wantErr bool
	}{
		{
			name:    "deserialize valid JSON",
			jsonStr: `{"trackEvents":true,"variationType":"SdkDefault","failed":false,"version":"1.0.0","reason":"DEFAULT","errorCode":"GENERAL","value":"test","cacheable":true}`,
			want: model.VariationResult[string]{
				TrackEvents:   true,
				VariationType: flag.VariationSDKDefault,
				Failed:        false,
				Version:       "1.0.0",
				Reason:        flag.ReasonDefault,
				ErrorCode:     flag.ErrorCodeGeneral,
				Value:         "test",
				Cacheable:     true,
			},
			wantErr: false,
		},
		{
			name:    "deserialize JSON with metadata",
			jsonStr: `{"trackEvents":false,"variationType":"enabled","failed":false,"version":"2.0.0","reason":"TARGETING_MATCH","errorCode":"FLAG_NOT_FOUND","value":"value","cacheable":false,"metadata":{"key":"value"}}`,
			want: model.VariationResult[string]{
				TrackEvents:   false,
				VariationType: "enabled",
				Failed:        false,
				Version:       "2.0.0",
				Reason:        flag.ReasonTargetingMatch,
				ErrorCode:     flag.ErrorCodeFlagNotFound,
				Value:         "value",
				Cacheable:     false,
				Metadata:      map[string]any{"key": "value"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got model.VariationResult[string]
			err := json.Unmarshal([]byte(tt.jsonStr), &got)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

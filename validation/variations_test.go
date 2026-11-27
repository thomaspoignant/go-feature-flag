package validation_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/modules/core/dto"
	"github.com/thomaspoignant/go-feature-flag/validation"
)

func TestValidateVariations(t *testing.T) {
	t.Parallel()

	const rewardSchema = `{
		"type": "object",
		"required": ["campaign"],
		"properties": {
			"campaign": {
				"type": "object",
				"required": ["reward"],
				"properties": {"reward": {"type": "integer"}}
			}
		}
	}`
	step := 1
	tests := []struct {
		name       string
		schema     string
		definition string
		want       []validation.VariationError
	}{
		{
			name:       "valid object",
			schema:     rewardSchema,
			definition: `{"variations":{"baseline":{"campaign":{"reward":3000}}}}`,
		},
		{
			name:       "valid array with an internal reference",
			schema:     `{"type":"array","items":{"$ref":"#/$defs/reward"},"$defs":{"reward":{"type":"integer"}}}`,
			definition: `{"variations":{"baseline":[3000,5000]}}`,
		},
		{
			name:       "invalid nested property",
			schema:     rewardSchema,
			definition: `{"variations":{"baseline":{"campaign":{"reward":3000}},"weekend":{"campaign":{"reward":"5000"}}}}`,
			want: []validation.VariationError{
				{FlagKey: "storefront-banner", Variation: "weekend", InstanceLocation: "/campaign/reward"},
			},
		},
		{
			name:       "invalid nested array item",
			schema:     `{"type":"array","items":{"type":"integer"}}`,
			definition: `{"variations":{"weekend":[3000,"5000"]}}`,
			want: []validation.VariationError{
				{FlagKey: "storefront-banner", Variation: "weekend", InstanceLocation: "/1"},
			},
		},
		{
			name:   "invalid scheduled changes",
			schema: rewardSchema,
			definition: `{
				"variations": {"baseline":{"campaign":{"reward":3000}}},
				"scheduledRollout": [
					{"date":"2026-12-01T10:00:00Z","variations":{"baseline":{"campaign":{"reward":5000}}}},
					{"date":"2026-12-01T11:00:00Z","variations":{
						"weekend":{"campaign":{"reward":true}},
						"baseline":{"campaign":{"reward":"5000"}}
					}}
				]
			}`,
			want: []validation.VariationError{
				{FlagKey: "storefront-banner", Variation: "baseline", InstanceLocation: "/campaign/reward", ScheduledStep: &step},
				{FlagKey: "storefront-banner", Variation: "weekend", InstanceLocation: "/campaign/reward", ScheduledStep: &step},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var definition dto.DTO
			require.NoError(t, json.Unmarshal([]byte(tt.definition), &definition))
			got, err := validation.ValidateVariations("storefront-banner", definition, []byte(tt.schema))
			require.NoError(t, err)
			require.Len(t, got, len(tt.want))
			for i := range got {
				require.NotEmpty(t, got[i].Message)
				got[i].Message = ""
			}
			require.Equal(t, tt.want, got)
		})
	}
}

func TestValidateVariationsInvalidSchema(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		schema      string
		wantMessage string
	}{
		{name: "malformed JSON", schema: `{"type":`, wantMessage: "parse JSON Schema"},
		{name: "invalid type", schema: `{"type":"invalid"}`, wantMessage: "compile JSON Schema"},
		{
			name:        "external reference",
			schema:      `{"$ref":"https://schema.example.invalid/reward.json"}`,
			wantMessage: "compile JSON Schema",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := validation.ValidateVariations("storefront-banner", dto.DTO{}, []byte(tt.schema))
			require.ErrorContains(t, err, tt.wantMessage)
			require.Empty(t, got)
		})
	}
}

package validation_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/configfile"
	"github.com/thomaspoignant/go-feature-flag/validation"
)

func TestValidateVariationsTOMLDate(t *testing.T) {
	t.Parallel()

	const definition = `
[storefront-banner.variations.baseline]
starts_at = 2026-12-01T09:00:00Z
`
	const schema = `{"properties":{"starts_at":{"type":"string","pattern":"^2026-12-01T09:00:00Z$"}}}`
	flags, err := configfile.ParseConfig([]byte(definition), "toml")
	require.NoError(t, err)

	violations, err := validation.ValidateVariations("storefront-banner", flags["storefront-banner"], []byte(schema))
	require.NoError(t, err)
	require.Empty(t, violations)
}

func TestValidateVariationsNonJSONValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		definition string
		wantError  string
	}{
		{
			name: "base variation",
			definition: `
[storefront-banner.variations.baseline]
reward = nan
`,
			wantError: `flag "storefront-banner" variation "baseline"`,
		},
		{
			name: "scheduled variation",
			definition: `
[[storefront-banner.scheduledRollout]]
date = 2026-12-01T10:00:00Z
[storefront-banner.scheduledRollout.variations.weekend]
reward = nan
`,
			wantError: `scheduledRollout[0]: validate flag "storefront-banner" variation "weekend"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			flags, err := configfile.ParseConfig([]byte(tt.definition), "toml")
			require.NoError(t, err)
			violations, err := validation.ValidateVariations(
				"storefront-banner", flags["storefront-banner"], []byte(`{"properties":{"reward":{"type":"integer"}}}`),
			)
			require.ErrorContains(t, err, tt.wantError)
			require.Empty(t, violations)
		})
	}
}

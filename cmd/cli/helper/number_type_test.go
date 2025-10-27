package helper_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/cmd/cli/generate/manifest/model"
	"github.com/thomaspoignant/go-feature-flag/cmd/cli/helper"
	"github.com/thomaspoignant/go-feature-flag/modules/core/testutils/testconvert"
)

func TestGetFlagTypeFromVariations(t *testing.T) {
	tests := []struct {
		name       string
		variations map[string]*interface{}
		expected   model.FlagType
		expectErr  bool
	}{
		{
			name: "single boolean type",
			variations: map[string]*interface{}{
				"var1": func() *interface{} { v := interface{}(true); return &v }(),
			},
			expected:  model.FlagTypeBoolean,
			expectErr: false,
		},
		{
			name: "single string type",
			variations: map[string]*interface{}{
				"var1": func() *interface{} { v := interface{}("test"); return &v }(),
			},
			expected:  model.FlagTypeString,
			expectErr: false,
		},
		{
			name: "single integer type",
			variations: map[string]*interface{}{
				"var1": func() *interface{} { v := interface{}(42); return &v }(),
			},
			expected:  model.FlagTypeInteger,
			expectErr: false,
		},
		{
			name: "single float type",
			variations: map[string]*interface{}{
				"var1": func() *interface{} { v := interface{}(42.0); return &v }(),
			},
			expected:  model.FlagTypeFloat,
			expectErr: false,
		},
		{
			name: "single object type",
			variations: map[string]*interface{}{
				"var1": func() *interface{} { v := interface{}(map[string]interface{}{"key": "value"}); return &v }(),
			},
			expected:  model.FlagTypeObject,
			expectErr: false,
		},
		{
			name: "ignore nil values type",
			variations: map[string]*interface{}{
				"var1": nil,
				"var2": testconvert.Interface(map[string]interface{}{"toto": "titi"}),
			},
			expected:  model.FlagTypeObject,
			expectErr: false,
		},
		{
			name: "mixed integer and float (with .0) types",
			variations: map[string]*interface{}{
				"var1": func() *interface{} { v := interface{}(42); return &v }(),
				"var2": func() *interface{} { v := interface{}(42.0); return &v }(),
			},
			expected:  model.FlagTypeInteger,
			expectErr: false,
		},
		{
			name: "mixed integer and float types",
			variations: map[string]*interface{}{
				"var1": func() *interface{} { v := interface{}(42); return &v }(),
				"var2": func() *interface{} { v := interface{}(42.2); return &v }(),
			},
			expected:  model.FlagTypeFloat,
			expectErr: false,
		},
		{
			name: "unknown type",
			variations: map[string]*interface{}{
				"var1": func() *interface{} { v := interface{}([]int{1, 2, 3}); return &v }(),
			},
			expected:  "",
			expectErr: true,
		},
		{
			name:       "nil variations",
			variations: nil,
			expected:   "",
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := helper.GetFlagTypeFromVariations(tt.variations)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

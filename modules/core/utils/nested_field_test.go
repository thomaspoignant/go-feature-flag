package utils_test

import (
    "testing"

    "github.com/stretchr/testify/require"
    "github.com/thomaspoignant/go-feature-flag/modules/core/utils"
)

func TestGetNestedFieldValue(t *testing.T) {
	tests := []struct {
		name    string
		ctx     map[string]interface{}
		key     string
		want    interface{}
		wantErr bool
	}{
		{
			name: "simple nested string",
			ctx: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "alice",
				},
			},
			key:  "user.name",
			want: "alice",
		},
		{
			name: "nested number",
			ctx: map[string]interface{}{
				"metrics": map[string]interface{}{
					"score": 42,
				},
			},
			key:  "metrics.score",
			want: float64(42), // JSON numbers become float64
		},
		{
			name: "nested bool",
			ctx: map[string]interface{}{
				"flags": map[string]interface{}{
					"beta": true,
				},
			},
			key:  "flags.beta",
			want: true,
		},
		{
			name: "missing key returns error",
			ctx: map[string]interface{}{
				"a": map[string]interface{}{"b": 1},
			},
			key:     "a.c",
			wantErr: true,
		},
		{
			name:    "empty key returns error",
			ctx:     map[string]interface{}{"a": 1},
			key:     "",
			wantErr: true,
		},
		{
			name:    "nil context returns error",
			ctx:     nil,
			key:     "a.b",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
        got, err := utils.GetNestedFieldValue(tt.ctx, tt.key)
        if tt.wantErr {
            require.Error(t, err)
            return
        }
        require.NoError(t, err)
        require.EqualValues(t, tt.want, got)
		})
	}
}

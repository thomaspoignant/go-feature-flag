package flag_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
)

func TestContext_AddIntoEvaluationContextEnrichment(t *testing.T) {
	type args struct {
		key   string
		value interface{}
	}
	tests := []struct {
		name                        string
		EvaluationContextEnrichment map[string]interface{}
		args                        args
		expected                    interface{}
	}{
		{
			name:                        "Add a new key to a nil map",
			EvaluationContextEnrichment: nil,
			args: args{
				key:   "env",
				value: "prod",
			},
			expected: "prod",
		},
		{
			name:                        "Add a new key to an existing map",
			EvaluationContextEnrichment: map[string]interface{}{"john": "doe"},
			args: args{
				key:   "env",
				value: "prod",
			},
			expected: "prod",
		},
		{
			name:                        "Override an existing key",
			EvaluationContextEnrichment: map[string]interface{}{"env": "dev"},
			args: args{
				key:   "env",
				value: "prod",
			},
			expected: "prod",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &flag.Context{
				EvaluationContextEnrichment: tt.EvaluationContextEnrichment,
			}
			s.AddIntoEvaluationContextEnrichment(tt.args.key, tt.args.value)
			assert.Equal(t, tt.expected, s.EvaluationContextEnrichment[tt.args.key])
		})
	}
}

package ffcontext

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUser(t *testing.T) {
	tests := []struct {
		name string
		got  EvaluationContext
		want EvaluationContext
	}{
		{
			name: "Builder with only targetingKey",
			got:  NewEvaluationContextBuilder("random-targetingKey").Build(),
			want: EvaluationContext{
				targetingKey: "random-targetingKey",
				attributes:   map[string]interface{}{},
			},
		},
		{
			name: "Builder with attributes attribute",
			got: NewEvaluationContextBuilder("random-targetingKey").
				AddCustom("test", "attributes").
				Build(),
			want: EvaluationContext{
				targetingKey: "random-targetingKey",
				attributes: map[string]interface{}{
					"test": "attributes",
				},
			},
		},
		{
			name: "Builder with attributes attribute",
			got: NewEvaluationContextBuilder("random-targetingKey").
				Anonymous(true).
				AddCustom("test", "attributes").
				Build(),
			want: EvaluationContext{
				targetingKey: "random-targetingKey",
				attributes: map[string]interface{}{
					"test":      "attributes",
					"anonymous": true,
				},
			},
		},
		{
			name: "NewUser with targetingKey",
			got:  NewEvaluationContext("random-targetingKey"),
			want: EvaluationContext{
				targetingKey: "random-targetingKey",
				attributes:   map[string]interface{}{},
			},
		},
		{
			name: "NewUser without targetingKey",
			got:  NewEvaluationContext(""),
			want: EvaluationContext{
				targetingKey: "",
				attributes:   map[string]interface{}{},
			},
		},
		{
			name: "NewAnonymousUser with targetingKey",
			got:  NewAnonymousEvaluationContext("random-targetingKey"),
			want: EvaluationContext{
				targetingKey: "random-targetingKey",
				attributes: map[string]interface{}{
					"anonymous": true,
				},
			},
		},
		{
			name: "NewAnonymousUser without targetingKey",
			got:  NewAnonymousEvaluationContext(""),
			want: EvaluationContext{
				targetingKey: "",
				attributes: map[string]interface{}{
					"anonymous": true,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.got)
			assert.Equal(t, tt.want.IsAnonymous(), tt.got.IsAnonymous())
			assert.Equal(t, tt.want.GetKey(), tt.got.GetKey())
			assert.Equal(t, tt.want.GetCustom(), tt.got.GetCustom())
		})
	}
}

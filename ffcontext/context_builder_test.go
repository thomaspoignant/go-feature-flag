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
			name: "Builder with only key",
			got:  NewEvaluationContextBuilder("random-key").Build(),
			want: EvaluationContext{
				key:    "random-key",
				custom: map[string]interface{}{},
			},
		},
		{
			name: "Builder with custom attribute",
			got: NewEvaluationContextBuilder("random-key").
				AddCustom("test", "custom").
				Build(),
			want: EvaluationContext{
				key: "random-key",
				custom: map[string]interface{}{
					"test": "custom",
				},
			},
		},
		{
			name: "Builder with custom attribute",
			got: NewEvaluationContextBuilder("random-key").
				Anonymous(true).
				AddCustom("test", "custom").
				Build(),
			want: EvaluationContext{
				key: "random-key",
				custom: map[string]interface{}{
					"test":      "custom",
					"anonymous": true,
				},
			},
		},
		{
			name: "NewUser with key",
			got:  NewEvaluationContext("random-key"),
			want: EvaluationContext{
				key:    "random-key",
				custom: map[string]interface{}{},
			},
		},
		{
			name: "NewUser without key",
			got:  NewEvaluationContext(""),
			want: EvaluationContext{
				key:    "",
				custom: map[string]interface{}{},
			},
		},
		{
			name: "NewAnonymousUser with key",
			got:  NewAnonymousEvaluationContext("random-key"),
			want: EvaluationContext{
				key: "random-key",
				custom: map[string]interface{}{
					"anonymous": true,
				},
			},
		},
		{
			name: "NewAnonymousUser without key",
			got:  NewAnonymousEvaluationContext(""),
			want: EvaluationContext{
				key: "",
				custom: map[string]interface{}{
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

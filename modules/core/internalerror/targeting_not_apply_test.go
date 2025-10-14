package internalerror

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/modules/core/ffcontext"
)

func TestRuleNotApply_Error(t *testing.T) {
	type fields struct {
		Context ffcontext.Context
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "Test RuleNotApply_Error",
			fields: fields{Context: ffcontext.NewEvaluationContext("test")},
			want:   "Rule does not apply for this user test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &RuleNotApplyError{
				Context: tt.fields.Context,
			}
			assert.EqualError(t, m, tt.want)
		})
	}
}

package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/modules/core/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/modules/core/utils"
)

func TestUserToMap(t *testing.T) {
	tests := []struct {
		name string
		u    ffcontext.Context
		want map[string]interface{}
	}{
		{
			name: "complete user",
			u: ffcontext.NewEvaluationContextBuilder("key").
				AddCustom("anonymous", false).
				AddCustom("email", "contact@gofeatureflag.org").
				Build(),
			want: map[string]interface{}{
				"key":       "key",
				"anonymous": false,
				"email":     "contact@gofeatureflag.org",
			},
		},
		{
			name: "anonymous user",
			u: ffcontext.NewEvaluationContextBuilder("key").
				AddCustom("anonymous", true).
				Build(),
			want: map[string]interface{}{
				"key":       "key",
				"anonymous": true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, utils.ContextToMap(tt.u))
		})
	}
}

package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/ffuser"
	"github.com/thomaspoignant/go-feature-flag/internal/utils"
)

func TestUserToMap(t *testing.T) {
	tests := []struct {
		name string
		u    ffuser.User
		want map[string]interface{}
	}{
		{
			name: "complete user",
			u:    ffuser.NewUserBuilder("key").Anonymous(false).AddCustom("email", "contact@gofeatureflag.org").Build(),
			want: map[string]interface{}{
				"key":       "key",
				"anonymous": false,
				"email":     "contact@gofeatureflag.org",
			},
		},
		{
			name: "anonymous user",
			u:    ffuser.NewAnonymousUser("key"),
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

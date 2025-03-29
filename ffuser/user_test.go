package ffuser_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/ffuser"
)

func TestUser_AddCustomAttribute(t *testing.T) {
	type args struct {
		name  string
		value any
	}
	tests := []struct {
		name string
		user ffuser.User
		args args
		want map[string]any
	}{
		{
			name: "trying to add nil value",
			user: ffuser.NewUser("123"),
			args: args{},
			want: map[string]any{},
		},
		{
			name: "add valid element",
			user: ffuser.NewUser("123"),
			args: args{
				name:  "test",
				value: "test",
			},
			want: map[string]any{
				"test": "test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.user.AddCustomAttribute(tt.args.name, tt.args.value)
			assert.Equal(t, tt.want, tt.user.GetCustom())
		})
	}
}

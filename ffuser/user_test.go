package ffuser_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/ffuser"
	"testing"
)

func TestUser_AddCustomAttribute(t *testing.T) {
	type args struct {
		name  string
		value interface{}
	}
	tests := []struct {
		name string
		user ffuser.User
		args args
		want map[string]interface{}
	}{
		{
			name: "trying to add nil value",
			user: ffuser.NewUser("123"),
			args: args{},
			want: map[string]interface{}{},
		},
		{
			name: "add valid element",
			user: ffuser.NewUser("123"),
			args: args{
				name:  "test",
				value: "test",
			},
			want: map[string]interface{}{
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

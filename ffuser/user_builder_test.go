package ffuser

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewUser(t *testing.T) {
	tests := []struct {
		name string
		got  User
		want User
	}{
		{
			name: "Builder with only key",
			got:  NewUserBuilder("random-key").Build(),
			want: User{
				key:    "random-key",
				custom: map[string]interface{}{},
			},
		},
		{
			name: "Builder with custom attribute",
			got: NewUserBuilder("random-key").
				AddCustom("test", "custom").
				Build(),
			want: User{
				key: "random-key",
				custom: map[string]interface{}{
					"test": "custom",
				},
			},
		},
		{
			name: "Builder with custom attribute",
			got: NewUserBuilder("random-key").
				Anonymous(true).
				AddCustom("test", "custom").
				Build(),
			want: User{
				key:       "random-key",
				anonymous: true,
				custom: map[string]interface{}{
					"test": "custom",
				},
			},
		},
		{
			name: "NewUser with key",
			got:  NewUser("random-key"),
			want: User{
				key:       "random-key",
				anonymous: false,
				custom:    map[string]interface{}{},
			},
		},
		{
			name: "NewUser without key",
			got:  NewUser(""),
			want: User{
				key:       "",
				anonymous: false,
				custom:    map[string]interface{}{},
			},
		},
		{
			name: "NewAnonymousUser with key",
			got:  NewAnonymousUser("random-key"),
			want: User{
				key:       "random-key",
				anonymous: true,
				custom:    map[string]interface{}{},
			},
		},
		{
			name: "NewAnonymousUser without key",
			got:  NewAnonymousUser(""),
			want: User{
				key:       "",
				anonymous: true,
				custom:    map[string]interface{}{},
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

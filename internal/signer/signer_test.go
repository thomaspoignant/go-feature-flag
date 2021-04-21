package signer

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSign(t *testing.T) {
	type args struct {
		payloadBody []byte
		secretToken []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "sign valid",
			args: args{
				payloadBody: []byte("this is a test"),
				secretToken: []byte("secret"),
			},
			want: "sha256=08abbecc4779c9260cc85a017eb9db8babb5308a614cc9f13a4b9976af6b7cee",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Sign(tt.args.payloadBody, tt.args.secretToken)
			assert.Equal(t, tt.want, got)
		})
	}
}

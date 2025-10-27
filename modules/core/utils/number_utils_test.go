package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/modules/core/utils"
)

func TestIsIntegral(t *testing.T) {
	type args struct {
		val float64
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "1.0 is an integer",
			args: args{
				val: 1.0,
			},
			want: true,
		},
		// check that 1.1 is not an integer
		{
			name: "1.1 is not an integer",
			args: args{
				val: 1.1,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, utils.IsIntegral(tt.args.val), "IsIntegral(%v)", tt.args.val)
		})
	}
}

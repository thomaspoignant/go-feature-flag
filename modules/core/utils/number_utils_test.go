package utils_test

import (
	"encoding/json"
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

func TestToFloat(t *testing.T) {
	tests := []struct {
		name   string
		input  any
		want   float64
		wantOk bool
	}{
		{name: "int", input: 3, want: 3, wantOk: true},
		{name: "int32", input: int32(3), want: 3, wantOk: true},
		{name: "int64", input: int64(3), want: 3, wantOk: true},
		{name: "float32", input: float32(1.5), want: 1.5, wantOk: true},
		{name: "float64", input: float64(1.5), want: 1.5, wantOk: true},
		{name: "json.Number", input: json.Number("2.5"), want: 2.5, wantOk: true},
		{name: "invalid json.Number", input: json.Number("abc"), want: 0, wantOk: false},
		{name: "string is not numeric", input: "1", want: 0, wantOk: false},
		{name: "bool is not numeric", input: true, want: 0, wantOk: false},
		{name: "nil is not numeric", input: nil, want: 0, wantOk: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := utils.ToFloat(tt.input)
			assert.Equal(t, tt.wantOk, ok)
			if tt.wantOk {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

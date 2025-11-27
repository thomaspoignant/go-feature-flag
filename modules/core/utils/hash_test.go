package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/modules/core/utils"
)

func TestHash(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want uint32
	}{
		{
			name: "standard hash",
			args: args{s: "flagNameUserKey"},
			want: 3946001934,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.Hash(tt.args.s); got != tt.want {
				t.Errorf("Hash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildHash(t *testing.T) {
	type args struct {
		flagKey       string
		bucketingKey  string
		maxPercentage uint32
	}
	tests := []struct {
		name string
		args args
		want uint32
	}{
		{
			name: "all fields",
			args: args{
				flagKey:       "my-flag",
				bucketingKey:  "e56f628e-9817-498f-ae38-4961e9c2bb21",
				maxPercentage: 100000,
			},
			want: 70272,
		},
		{
			name: "empty flag key",
			args: args{
				flagKey:       "",
				bucketingKey:  "e56f628e-9817-498f-ae38-4961e9c2bb21",
				maxPercentage: 100000,
			},
			want: 74237,
		},
		{
			name: "empty flag key and bucketing key",
			args: args{
				flagKey:       "",
				bucketingKey:  "",
				maxPercentage: 100000,
			},
			want: 36261,
		},
		{
			name: "empty flag key and bucketing key",
			args: args{
				flagKey:       "",
				bucketingKey:  "",
				maxPercentage: 0,
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.BuildHash(tt.args.flagKey, tt.args.bucketingKey, tt.args.maxPercentage)
			assert.Equal(t, tt.want, got)
		})
	}
}

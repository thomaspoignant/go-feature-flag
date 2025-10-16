package utils_test

import (
	"testing"

	"github.com/thomaspoignant/go-feature-flag/modules/core/utils"
)

func TestContains(t *testing.T) {
	type args struct {
		s   []string
		str string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Should contains an element",
			args: args{
				s:   []string{"aa", "ab", "abc", "abcd", "abcde"},
				str: "aa",
			},
			want: true,
		},
		{
			name: "Should not contains an element",
			args: args{
				s:   []string{"aa", "ab", "abc", "abcd", "abcde"},
				str: "a",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.Contains(tt.args.s, tt.args.str); got != tt.want {
				t.Errorf("Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

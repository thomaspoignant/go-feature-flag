package utils

import "testing"

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
			if got := Hash(tt.args.s); got != tt.want {
				t.Errorf("Hash() = %v, want %v", got, tt.want)
			}
		})
	}
}

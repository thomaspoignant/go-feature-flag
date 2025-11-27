package internalerror

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNestedKeyNotFoundError_Error(t *testing.T) {
	type fields struct {
		Key string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "simple key error message",
			fields: fields{Key: "teamId"},
			want:   "nested key not found: teamId",
		},
		{
			name:   "nested key error message",
			fields: fields{Key: "company.id"},
			want:   "nested key not found: company.id",
		},
		{
			name:   "deep nested key error message",
			fields: fields{Key: "user.profile.role"},
			want:   "nested key not found: user.profile.role",
		},
		{
			name:   "empty key error message",
			fields: fields{Key: ""},
			want:   "nested key not found: ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &NestedKeyNotFoundError{
				Key: tt.fields.Key,
			}
			assert.EqualError(t, e, tt.want)
		})
	}
}

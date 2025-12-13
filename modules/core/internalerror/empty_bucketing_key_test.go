package internalerror

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmptyBucketingKeyErrorError(t *testing.T) {
	type fields struct {
		Message string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "error with standard message",
			fields: fields{Message: "Empty bucketing key"},
			want:   "Error: Empty bucketing key",
		},
		{
			name:   "error with key-specific message",
			fields: fields{Message: "Empty user key"},
			want:   "Error: Empty user key",
		},
		{
			name:   "error with custom key message",
			fields: fields{Message: "Empty sessionId key"},
			want:   "Error: Empty sessionId key",
		},
		{
			name:   "error with empty message",
			fields: fields{Message: ""},
			want:   "Error: ",
		},
		{
			name:   "error with detailed message",
			fields: fields{Message: "Empty bucketing key: user context is missing required field"},
			want:   "Error: Empty bucketing key: user context is missing required field",
		},
		{
			name:   "error with special characters in message",
			fields: fields{Message: "Empty key: test@123!$%"},
			want:   "Error: Empty key: test@123!$%",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &EmptyBucketingKeyError{
				Message: tt.fields.Message,
			}
			assert.EqualError(t, e, tt.want)
		})
	}
}

func TestEmptyBucketingKeyErrorImplementsErrorInterface(t *testing.T) {
	var err error
	e := &EmptyBucketingKeyError{
		Message: "test error",
	}
	err = e
	assert.NotNil(t, err, "EmptyBucketingKeyError should implement error interface")
	assert.Equal(t, "Error: test error", err.Error())
}

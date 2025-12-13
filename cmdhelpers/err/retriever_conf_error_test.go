package err

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRetrieverConfError(t *testing.T) {
	tests := []struct {
		name     string
		property string
		kind     string
	}{
		{
			name:     "simple property and kind",
			property: "bucket",
			kind:     "s3",
		},
		{
			name:     "camelCase property",
			property: "bucketName",
			kind:     "s3",
		},
		{
			name:     "empty strings",
			property: "",
			kind:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewRetrieverConfError(tt.property, tt.kind)
			require.NotNil(t, err, "expected non-nil error")
			assert.Equal(t, tt.property, err.property, "property should match")
			assert.Equal(t, tt.kind, err.kind, "kind should match")
		})
	}
}

func TestRetrieverConfError_Error(t *testing.T) {
	tests := []struct {
		name     string
		property string
		kind     string
		expected string
	}{
		{
			name:     "simple property and kind",
			property: "bucket",
			kind:     "s3",
			expected: `invalid retriever: no "bucket" property found for kind "s3"`,
		},
		{
			name:     "camelCase property - not converted",
			property: "bucketName",
			kind:     "s3",
			expected: `invalid retriever: no "bucketName" property found for kind "s3"`,
		},
		{
			name:     "multiple words in camelCase",
			property: "someLongPropertyName",
			kind:     "http",
			expected: `invalid retriever: no "someLongPropertyName" property found for kind "http"`,
		},
		{
			name:     "empty property and kind",
			property: "",
			kind:     "",
			expected: `invalid retriever: no "" property found for kind ""`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewRetrieverConfError(tt.property, tt.kind)
			got := err.Error()
			assert.Equal(t, tt.expected, got, "Error() message should match expected format")
		})
	}
}

func TestRetrieverConfError_CliErrorMessage(t *testing.T) {
	tests := []struct {
		name     string
		property string
		kind     string
		expected string
	}{
		{
			name:     "simple lowercase property",
			property: "bucket",
			kind:     "s3",
			expected: `invalid retriever: no "bucket" property found for kind "s3"`,
		},
		{
			name:     "camelCase to kebab-case",
			property: "bucketName",
			kind:     "s3",
			expected: `invalid retriever: no "bucket-name" property found for kind "s3"`,
		},
		{
			name:     "multiple camelCase words",
			property: "someLongPropertyName",
			kind:     "http",
			expected: `invalid retriever: no "some-long-property-name" property found for kind "http"`,
		},
		{
			name:     "single uppercase letter",
			property: "a",
			kind:     "test",
			expected: `invalid retriever: no "a" property found for kind "test"`,
		},
		{
			name:     "uppercase property",
			property: "BUCKET",
			kind:     "s3",
			expected: `invalid retriever: no "bucket" property found for kind "s3"`,
		},
		{
			name:     "mixed case with numbers",
			property: "bucket1Name",
			kind:     "s3",
			expected: `invalid retriever: no "bucket1-name" property found for kind "s3"`,
		},
		{
			name:     "already kebab-case",
			property: "bucket-name",
			kind:     "s3",
			expected: `invalid retriever: no "bucket-name" property found for kind "s3"`,
		},
		{
			name:     "starts with uppercase",
			property: "BucketName",
			kind:     "s3",
			expected: `invalid retriever: no "bucket-name" property found for kind "s3"`,
		},
		{
			name:     "consecutive uppercase letters",
			property: "HTTPSEndpoint",
			kind:     "http",
			expected: `invalid retriever: no "httpsendpoint" property found for kind "http"`,
		},
		{
			name:     "empty property and kind",
			property: "",
			kind:     "",
			expected: `invalid retriever: no "" property found for kind ""`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewRetrieverConfError(tt.property, tt.kind)
			got := err.CliErrorMessage()
			assert.Equal(t, tt.expected, got, "CliErrorMessage() should convert camelCase to kebab-case")
		})
	}
}

func TestRetrieverConfError_ImplementsError(t *testing.T) {
	var _ error = &RetrieverConfError{}
	err := NewRetrieverConfError("test", "test")
	require.NotNil(t, err, "expected non-nil error")
	assert.Equal(t, "test", err.property, "property should match")
	assert.Equal(t, "test", err.kind, "kind should match")
}

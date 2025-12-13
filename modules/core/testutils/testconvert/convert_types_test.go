package testconvert_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/modules/core/testutils/testconvert"
)

func TestBool(t *testing.T) {
	tests := []struct {
		name  string
		input bool
		want  bool
	}{
		{
			name:  "true value",
			input: true,
			want:  true,
		},
		{
			name:  "false value",
			input: false,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testconvert.Bool(tt.input)
			assert.NotNil(t, result, "Result should not be nil")
			assert.Equal(t, tt.want, *result, "Value should match input")
		})
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "non-empty string",
			input: "test-value",
			want:  "test-value",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "string with special characters",
			input: "test@123!$%",
			want:  "test@123!$%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testconvert.String(tt.input)
			assert.NotNil(t, result, "Result should not be nil")
			assert.Equal(t, tt.want, *result, "Value should match input")
		})
	}
}

func TestInt(t *testing.T) {
	tests := []struct {
		name  string
		input int
		want  int
	}{
		{
			name:  "positive integer",
			input: 42,
			want:  42,
		},
		{
			name:  "zero",
			input: 0,
			want:  0,
		},
		{
			name:  "negative integer",
			input: -100,
			want:  -100,
		},
		{
			name:  "large integer",
			input: 999999999,
			want:  999999999,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testconvert.Int(tt.input)
			assert.NotNil(t, result, "Result should not be nil")
			assert.Equal(t, tt.want, *result, "Value should match input")
		})
	}
}

func TestFloat64(t *testing.T) {
	tests := []struct {
		name  string
		input float64
		want  float64
	}{
		{
			name:  "positive float",
			input: 3.14,
			want:  3.14,
		},
		{
			name:  "zero",
			input: 0.0,
			want:  0.0,
		},
		{
			name:  "negative float",
			input: -2.5,
			want:  -2.5,
		},
		{
			name:  "large float",
			input: 123456.789,
			want:  123456.789,
		},
		{
			name:  "small float",
			input: 0.000001,
			want:  0.000001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testconvert.Float64(tt.input)
			assert.NotNil(t, result, "Result should not be nil")
			assert.Equal(t, tt.want, *result, "Value should match input")
		})
	}
}

func TestTime(t *testing.T) {
	now := time.Now()
	zeroTime := time.Time{}
	fixedTime := time.Date(2023, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name  string
		input time.Time
		want  time.Time
	}{
		{
			name:  "current time",
			input: now,
			want:  now,
		},
		{
			name:  "zero time",
			input: zeroTime,
			want:  zeroTime,
		},
		{
			name:  "fixed time",
			input: fixedTime,
			want:  fixedTime,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testconvert.Time(tt.input)
			assert.NotNil(t, result, "Result should not be nil")
			assert.Equal(t, tt.want, *result, "Value should match input")
		})
	}
}

func TestInterface(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  interface{}
	}{
		{
			name:  "string value",
			input: "test-string",
			want:  "test-string",
		},
		{
			name:  "int value",
			input: 42,
			want:  42,
		},
		{
			name:  "bool value true",
			input: true,
			want:  true,
		},
		{
			name:  "bool value false",
			input: false,
			want:  false,
		},
		{
			name:  "float64 value",
			input: 3.14,
			want:  3.14,
		},
		{
			name:  "nil value",
			input: nil,
			want:  nil,
		},
		{
			name:  "map value",
			input: map[string]interface{}{"key": "value"},
			want:  map[string]interface{}{"key": "value"},
		},
		{
			name:  "slice value",
			input: []interface{}{1, 2, 3},
			want:  []interface{}{1, 2, 3},
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "zero int",
			input: 0,
			want:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testconvert.Interface(tt.input)
			assert.NotNil(t, result, "Result should not be nil")
			assert.Equal(t, tt.want, *result, "Value should match input")
		})
	}
}

func TestPointerIndependence(t *testing.T) {
	t.Run("Bool pointer independence", func(t *testing.T) {
		original := true
		ptr := testconvert.Bool(original)
		original = false
		assert.True(t, *ptr, "Pointer value should not change when original variable changes")
	})

	t.Run("String pointer independence", func(t *testing.T) {
		original := "original"
		ptr := testconvert.String(original)
		original = "changed"
		assert.Equal(t, "original", *ptr, "Pointer value should not change when original variable changes")
	})

	t.Run("Int pointer independence", func(t *testing.T) {
		original := 42
		ptr := testconvert.Int(original)
		original = 100
		assert.Equal(t, 42, *ptr, "Pointer value should not change when original variable changes")
	})

	t.Run("Float64 pointer independence", func(t *testing.T) {
		original := 3.14
		ptr := testconvert.Float64(original)
		original = 2.71
		assert.Equal(t, 3.14, *ptr, "Pointer value should not change when original variable changes")
	})
}


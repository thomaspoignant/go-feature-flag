package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/modules/core/testutils/testconvert"
)

func TestOfrepEventStream_IsValid(t *testing.T) {
	tests := []struct {
		name        string
		eventStream config.OfrepEventStream
		wantErr     assert.ErrorAssertionFunc
	}{
		{
			name:        "default configuration",
			eventStream: config.OfrepEventStream{},
			wantErr:     assert.NoError,
		},
		{
			name: "valid base URL",
			eventStream: config.OfrepEventStream{
				BaseURL:            "https://gofeatureflag.example.com/goff/",
				InactivityDelaySec: 60,
			},
			wantErr: assert.NoError,
		},
		{
			name:        "base URL without scheme",
			eventStream: config.OfrepEventStream{BaseURL: "gofeatureflag.example.com"},
			wantErr:     assert.Error,
		},
		{
			name:        "relative base URL",
			eventStream: config.OfrepEventStream{BaseURL: "/goff"},
			wantErr:     assert.Error,
		},
		{
			name:        "base URL with an unsupported scheme",
			eventStream: config.OfrepEventStream{BaseURL: "ftp://gofeatureflag.example.com"},
			wantErr:     assert.Error,
		},
		{
			name:        "base URL with a query",
			eventStream: config.OfrepEventStream{BaseURL: "https://gofeatureflag.example.com?apiKey=xxx"},
			wantErr:     assert.Error,
		},
		{
			name:        "invalid base URL",
			eventStream: config.OfrepEventStream{BaseURL: "https://gofeatureflag.example.com/%zz"},
			wantErr:     assert.Error,
		},
		{
			name:        "negative inactivity delay",
			eventStream: config.OfrepEventStream{InactivityDelaySec: -1},
			wantErr:     assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, tt.eventStream.IsValid())
		})
	}
}

func TestOfrepEventStream_IsEnabled(t *testing.T) {
	tests := []struct {
		name    string
		enabled *bool
		want    bool
	}{
		{name: "enabled by default", enabled: nil, want: true},
		{name: "explicitly enabled", enabled: testconvert.Bool(true), want: true},
		{name: "explicitly disabled", enabled: testconvert.Bool(false), want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, config.OfrepEventStream{Enabled: tt.enabled}.IsEnabled())
		})
	}
}

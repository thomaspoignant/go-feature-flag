package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
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
		baseURL string
		want    bool
	}{
		{name: "empty baseURL", baseURL: "", want: false},
		{name: "valid baseURL", baseURL: "https://gofeatureflag.example.com", want: true},
		{name: "invalid baseURL", baseURL: "://badurl", want: false},
		{name: "unsupported scheme baseURL", baseURL: "ftp://gofeatureflag.example.com", want: false},
		{name: "relative baseURL", baseURL: "/relative/path", want: false},
		{name: "baseURL with query", baseURL: "https://gofeatureflag.example.com?something=1", want: false},
		{name: "baseURL with percent-encoding error", baseURL: "https://gofeatureflag.example.com/%zz", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, config.OfrepEventStream{BaseURL: tt.baseURL}.IsEnabled())
		})
	}
}

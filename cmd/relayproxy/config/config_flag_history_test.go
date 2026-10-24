package config_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
)

func boolPtr(b bool) *bool { return &b }

func TestFlagHistory_IsEnabled(t *testing.T) {
	tests := []struct {
		name string
		conf config.FlagHistory
		want bool
	}{
		{name: "default without database is inactive", conf: config.FlagHistory{}, want: false},
		{name: "on by default when a database is configured", conf: config.FlagHistory{URI: "postgres://x"}, want: true},
		{name: "explicitly enabled", conf: config.FlagHistory{Enabled: boolPtr(true), URI: "postgres://x"}, want: true},
		{name: "explicitly disabled", conf: config.FlagHistory{Enabled: boolPtr(false), URI: "postgres://x"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.conf.IsEnabled())
		})
	}
}

func TestFlagHistory_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		conf    config.FlagHistory
		wantErr bool
	}{
		{name: "nothing configured is valid (no database, history inactive)", conf: config.FlagHistory{}},
		{name: "explicitly disabled without uri", conf: config.FlagHistory{Enabled: boolPtr(false)}},
		{name: "uri only", conf: config.FlagHistory{URI: "postgres://localhost/goff"}},
		{name: "explicitly enabled with uri", conf: config.FlagHistory{Enabled: boolPtr(true), URI: "postgres://x"}},
		{name: "explicitly enabled without uri", conf: config.FlagHistory{Enabled: boolPtr(true)}, wantErr: true},
		{
			name:    "negative queue size",
			conf:    config.FlagHistory{URI: "postgres://x", QueueSize: -1},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.conf.IsValid()
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestFlagHistory_RetentionDuration(t *testing.T) {
	tests := []struct {
		name      string
		retention time.Duration
		want      time.Duration
	}{
		{name: "default is one year", retention: 0, want: 365 * 24 * time.Hour},
		{name: "custom", retention: 48 * time.Hour, want: 48 * time.Hour},
		{name: "negative keeps forever", retention: -time.Second, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, config.FlagHistory{Retention: tt.retention}.RetentionDuration())
		})
	}
}

func TestConfig_IsValid_flagHistory(t *testing.T) {
	c := &config.Config{FlagHistory: config.FlagHistory{Enabled: boolPtr(true)}}
	assert.ErrorContains(t, c.IsValid(), "flagHistory.uri")
}

func TestConfig_IsValid_flagHistoryDefaultDoesNotBreakExistingSetups(t *testing.T) {
	c := &config.Config{}
	assert.NoError(t, c.FlagHistory.IsValid())
}

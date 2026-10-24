package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
)

func TestWithInheritedDefaultFlagTTL(t *testing.T) {
	tests := []struct {
		name        string
		flagsetTTL  string
		globalTTL   string
		wantFlagTTL string
	}{
		{name: "flagset overrides global", flagsetTTL: "7d", globalTTL: "90d", wantFlagTTL: "7d"},
		{name: "flagset inherits global", globalTTL: "30d", wantFlagTTL: "30d"},
		{name: "none configured"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := config.FlagSet{CommonFlagSet: config.CommonFlagSet{DefaultFlagTTL: tt.flagsetTTL}}
			got := withInheritedDefaultFlagTTL(fs, tt.globalTTL)
			assert.Equal(t, tt.wantFlagTTL, got.DefaultFlagTTL)
		})
	}
}

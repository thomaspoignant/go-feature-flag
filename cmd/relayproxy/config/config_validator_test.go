package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/retrieverconf"
)

func TestConfig_IsValid_DefaultFlagTTL(t *testing.T) {
	retriever := &retrieverconf.RetrieverConf{Kind: "file", Path: "../testdata/config/valid-file.yaml"}
	tests := []struct {
		name    string
		config  *config.Config
		wantErr string
	}{
		{name: "empty ttl", config: &config.Config{CommonFlagSet: config.CommonFlagSet{Retriever: retriever}}},
		{name: "valid global ttl", config: &config.Config{CommonFlagSet: config.CommonFlagSet{Retriever: retriever, DefaultFlagTTL: "90d"}}},
		{
			name:    "invalid global ttl",
			config:  &config.Config{CommonFlagSet: config.CommonFlagSet{Retriever: retriever, DefaultFlagTTL: "soon"}},
			wantErr: `defaultFlagTTL: invalid ttl "soon"`,
		},
		{
			name: "invalid flagset ttl",
			config: &config.Config{FlagSets: []config.FlagSet{{
				Name:          "fs",
				APIKeys:       []string{"key"},
				CommonFlagSet: config.CommonFlagSet{Retriever: retriever, DefaultFlagTTL: "-3d"},
			}}},
			wantErr: `defaultFlagTTL: invalid ttl "-3d"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.IsValid()
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

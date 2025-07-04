package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
)

func TestFlagSet_New(t *testing.T) {
	tests := []struct {
		name     string
		flagSet  config.FlagSet
		expected config.FlagSet
	}{
		{
			name: "empty flagset",
			flagSet: config.FlagSet{
				Name:       "",
				ApiKey:     "",
				Retrievers: nil,
				Notifiers:  nil,
				Exporters:  nil,
			},
			expected: config.FlagSet{
				Name:       "",
				ApiKey:     "",
				Retrievers: nil,
				Notifiers:  nil,
				Exporters:  nil,
			},
		},
		{
			name: "flagset with all fields",
			flagSet: config.FlagSet{
				Name:       "test-flagset",
				ApiKey:     "test-api-key",
				Retrievers: &[]config.RetrieverConf{},
				Notifiers:  &[]config.NotifierConf{},
				Exporters:  &[]config.ExporterConf{},
			},
			expected: config.FlagSet{
				Name:       "test-flagset",
				ApiKey:     "test-api-key",
				Retrievers: &[]config.RetrieverConf{},
				Notifiers:  &[]config.NotifierConf{},
				Exporters:  &[]config.ExporterConf{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.flagSet)
		})
	}
}

func TestFlagSet_FieldValidation(t *testing.T) {
	tests := []struct {
		name    string
		flagSet config.FlagSet
		valid   bool
	}{
		{
			name: "valid flagset with name and api key",
			flagSet: config.FlagSet{
				Name:   "test-flagset",
				ApiKey: "test-api-key",
			},
			valid: true,
		},
		{
			name: "valid flagset with empty name",
			flagSet: config.FlagSet{
				Name:   "",
				ApiKey: "test-api-key",
			},
			valid: true,
		},
		{
			name: "valid flagset with empty api key",
			flagSet: config.FlagSet{
				Name:   "test-flagset",
				ApiKey: "",
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Since the struct doesn't have any validation logic yet,
			// we're just testing that the struct can be created with these values
			assert.NotNil(t, tt.flagSet)
		})
	}
}

func TestFlagSet_WithRetrievers(t *testing.T) {
	retrievers := &[]config.RetrieverConf{}
	flagSet := config.FlagSet{
		Name:       "test-flagset",
		Retrievers: retrievers,
	}

	assert.NotNil(t, flagSet.Retrievers)
	assert.Equal(t, retrievers, flagSet.Retrievers)
}

func TestFlagSet_WithNotifiers(t *testing.T) {
	notifiers := &[]config.NotifierConf{}
	flagSet := config.FlagSet{
		Name:      "test-flagset",
		Notifiers: notifiers,
	}

	assert.NotNil(t, flagSet.Notifiers)
	assert.Equal(t, notifiers, flagSet.Notifiers)
}

func TestFlagSet_WithExporters(t *testing.T) {
	exporters := &[]config.ExporterConf{}
	flagSet := config.FlagSet{
		Name:      "test-flagset",
		Exporters: exporters,
	}

	assert.NotNil(t, flagSet.Exporters)
	assert.Equal(t, exporters, flagSet.Exporters)
}

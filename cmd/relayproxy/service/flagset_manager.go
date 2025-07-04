package service

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"go.uber.org/zap"
)

const defaultFlagSetName = "default"

type flagsetManagerMode string

const (
	flagsetManagerModeDefault  flagsetManagerMode = "default"
	flagsetManagerModeFlagsets flagsetManagerMode = "flagsets"
)

// FlagsetManager is the manager of the flagsets.
// It is used to retrieve the flagset linked to the API Key.
type FlagsetManager struct {
	// DefaultFlagSet is the flagset used when no API Key is provider.
	// It is the legacy way to handle feature flags in GO Feature Flag.
	// This is used only if no flag set is configured in the configuration file.
	DefaultFlagSet *ffclient.GoFeatureFlag

	// FlagSets is a map that stores the different instances of GoFeatureFlag (one per flagset)
	// It is used to retrieve the flagset linked to the API Key.
	FlagSets map[string]*ffclient.GoFeatureFlag

	// APIKeysToFlagSet is a map that stores the API Key linked to the flagset name.
	// It is used to retrieve the flagset linked to the API Key.
	APIKeysToFlagSet map[string]string

	// Config is the configuration of the relay proxy.
	// It is used to retrieve the configuration of the relay proxy.
	config *config.Config

	// Mode is the mode of the flagset manager.
	mode flagsetManagerMode
}

// NewFlagsetManager is creating a new FlagsetManager.
// It is used to retrieve the flagset linked to the API Key.
func NewFlagsetManager(config *config.Config, logger *zap.Logger) (FlagsetManager, error) {
	if config == nil {
		return FlagsetManager{}, fmt.Errorf("configuration is nil")
	}

	if len(config.FlagSets) == 0 {
		// in case you are using the relay proxy with flagsets, we create the flagsets and map them to the APIKeys.
		// note that the default configuration is ignored in this case.
		return newFlagsetManagerWithDefaultConfig(config, logger)
	}

	flagsetMngr, err := newFlagsetManagerWithFlagsets(config, logger)
	if err != nil {
		return newFlagsetManagerWithDefaultConfig(config, logger)
	}
	return flagsetMngr, nil
}

// newFlagsetManagerWithDefaultConfig is creating a new FlagsetManager with the default configuration.
// The default configuration is the top level configuration of the relay proxy.
func newFlagsetManagerWithDefaultConfig(c *config.Config, logger *zap.Logger) (FlagsetManager, error) {
	defaultFlagSet := config.FlagSet{
		Name: "default",
		CommonFlagSet: config.CommonFlagSet{
			Retrievers:                      c.Retrievers,
			Notifiers:                       c.Notifiers,
			Exporters:                       c.Exporters,
			FileFormat:                      c.FileFormat,
			PollingInterval:                 c.PollingInterval,
			StartWithRetrieverError:         c.StartWithRetrieverError,
			EnablePollingJitter:             c.EnablePollingJitter,
			DisableNotifierOnInit:           c.DisableNotifierOnInit,
			EvaluationContextEnrichment:     c.EvaluationContextEnrichment,
			PersistentFlagConfigurationFile: c.PersistentFlagConfigurationFile,
		},
	}
	client, err := NewGoFeatureFlagClient(&defaultFlagSet, logger, []notifier.Notifier{})
	if err != nil {
		return FlagsetManager{}, err
	}
	return FlagsetManager{
		DefaultFlagSet: &client,
		config:         c,
		mode:           flagsetManagerModeDefault,
	}, nil
}

// newFlagsetManagerWithFlagsets is creating a new FlagsetManager with flagsets.
// It is used to create the flagsets and map them to the APIKeys.
func newFlagsetManagerWithFlagsets(config *config.Config, logger *zap.Logger) (FlagsetManager, error) {
	flagsets := make(map[string]*ffclient.GoFeatureFlag)
	apiKeysToFlagSet := make(map[string]string)

	for _, flagset := range config.FlagSets {
		client, err := NewGoFeatureFlagClient(&flagset, logger, []notifier.Notifier{})
		if err != nil {
			logger.Error("failed to create goff client for flagset", zap.String("flagset", flagset.Name), zap.Error(err))
			continue
		}

		flagSetName := flagset.Name
		if flagSetName == "" || flagSetName == defaultFlagSetName {
			// generating a default flagset name if not provided or equals to default
			flagSetName = uuid.New().String()
		}

		flagsets[flagSetName] = &client
		for _, apiKey := range flagset.ApiKeys {
			apiKeysToFlagSet[apiKey] = flagSetName
		}
	}

	if len(flagsets) == 0 {
		return FlagsetManager{}, errors.New("no flagset configured")
	}

	return FlagsetManager{
		FlagSets:         flagsets,
		APIKeysToFlagSet: apiKeysToFlagSet,
		config:           config,
		mode:             flagsetManagerModeFlagsets,
	}, nil
}

// GetFlagSet is returning the flag set linked to the API Key
func (m *FlagsetManager) GetFlagSet(apiKey string) *ffclient.GoFeatureFlag {
	switch m.mode {
	case flagsetManagerModeFlagsets:
		return m.FlagSets[m.APIKeysToFlagSet[apiKey]]
	default:
		return m.DefaultFlagSet
	}
}

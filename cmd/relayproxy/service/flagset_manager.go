package service

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"github.com/thomaspoignant/go-feature-flag/utils"
	"go.uber.org/zap"
)

type flagsetManagerMode string

const (
	flagsetManagerModeDefault  flagsetManagerMode = "default"
	flagsetManagerModeFlagsets flagsetManagerMode = "flagsets"
)

// FlagsetManager is the interface for managing flagsets.
// It is used to retrieve the flagset linked to the API Key.
type FlagsetManager interface {
	// FlagSet returns the flag set linked to the API Key
	FlagSet(apiKey string) (*ffclient.GoFeatureFlag, error)
	// FlagSetName returns the name of the flagset linked to the API Key
	FlagSetName(apiKey string) (string, error)
	// AllFlagSets returns all flag sets of the flagset manager
	AllFlagSets() (map[string]*ffclient.GoFeatureFlag, error)
	// Default returns the default flagset
	Default() *ffclient.GoFeatureFlag
	// IsDefaultFlagSet returns true if the manager is in default mode (no flagsets configured)
	IsDefaultFlagSet() bool
	// Close closes the flagset manager
	Close()
}

// flagsetManagerImpl is the internal implementation of FlagsetManager
type flagsetManagerImpl struct {
	// DefaultFlagSet is the flagset used when no API Key is provider.
	// It is the legacy way to handle feature flags in GO Feature Flag.
	// This is used only if no flag set is configured in the configuration file.
	DefaultFlagSet *ffclient.GoFeatureFlag

	// FlagSets is a map that stores the different instances of GoFeatureFlag (one per flagset)
	// It is used to retrieve the flagset linked to the API Key.
	FlagSets map[string]*ffclient.GoFeatureFlag

	// APIKeysToFlagSetName is a map that stores the API Key linked to the flagset name.
	// It is used to retrieve the flagset linked to the API Key.
	APIKeysToFlagSetName map[string]string

	// Config is the configuration of the relay proxy.
	// It is used to retrieve the configuration of the relay proxy.
	config *config.Config

	// Mode is the mode of the flagset manager.
	mode flagsetManagerMode
}

// NewFlagsetManager is creating a new FlagsetManager.
// It is used to retrieve the flagset linked to the API Key.
func NewFlagsetManager(
	config *config.Config, logger *zap.Logger, notifiers []notifier.Notifier) (FlagsetManager, error) {
	if config == nil {
		return nil, fmt.Errorf("configuration is nil")
	}

	if len(config.FlagSets) == 0 {
		// in case you are using the relay proxy with flagsets, we create the flagsets and map them to the APIKeys.
		// note that the default configuration is ignored in this case.
		return newFlagsetManagerWithDefaultConfig(config, logger, notifiers)
	}

	flagsetMngr, err := newFlagsetManagerWithFlagsets(config, logger, notifiers)
	if err != nil {
		return nil, err
	}
	return flagsetMngr, nil
}

// newFlagsetManagerWithDefaultConfig is creating a new FlagsetManager with the default configuration.
// The default configuration is the top level configuration of the relay proxy.
func newFlagsetManagerWithDefaultConfig(
	c *config.Config, logger *zap.Logger, notifiers []notifier.Notifier) (FlagsetManager, error) {
	defaultFlagSet := config.FlagSet{
		Name: utils.DefaultFlagSetName,
		CommonFlagSet: config.CommonFlagSet{
			Retriever:                       c.Retriever,
			Retrievers:                      c.Retrievers,
			Notifiers:                       c.Notifiers,
			Exporter:                        c.Exporter,
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
	client, err := NewGoFeatureFlagClient(&defaultFlagSet, logger, notifiers)
	if err != nil {
		return nil, err
	}
	return &flagsetManagerImpl{
		DefaultFlagSet: client,
		config:         c,
		mode:           flagsetManagerModeDefault,
	}, nil
}

// newFlagsetManagerWithFlagsets is creating a new FlagsetManager with flagsets.
// It is used to create the flagsets and map them to the APIKeys.
func newFlagsetManagerWithFlagsets(
	config *config.Config, logger *zap.Logger, notifiers []notifier.Notifier) (FlagsetManager, error) {
	flagsets := make(map[string]*ffclient.GoFeatureFlag)
	apiKeysToFlagSet := make(map[string]string)

	for index, flagset := range config.FlagSets {
		client, err := NewGoFeatureFlagClient(&flagset, logger, notifiers)
		if err != nil {
			logger.Error(
				"failed to create goff client for flagset",
				zap.Int("flagset_index", index),
				zap.String("flagset", flagset.Name),
				zap.Error(err),
			)
			continue
		}

		flagSetName := flagset.Name
		if flagSetName == "" || flagSetName == utils.DefaultFlagSetName {
			// generating a default flagset name if not provided or equals to default
			flagSetName = uuid.New().String()
		}

		flagsets[flagSetName] = client
		for _, apiKey := range flagset.APIKeys {
			apiKeysToFlagSet[apiKey] = flagSetName
		}
	}

	if len(flagsets) == 0 {
		return nil, errors.New("no flagset configured")
	}

	return &flagsetManagerImpl{
		FlagSets:             flagsets,
		APIKeysToFlagSetName: apiKeysToFlagSet,
		config:               config,
		mode:                 flagsetManagerModeFlagsets,
	}, nil
}

// FlagSet is returning the flag set linked to the API Key
func (m *flagsetManagerImpl) FlagSet(apiKey string) (*ffclient.GoFeatureFlag, error) {
	switch m.mode {
	case flagsetManagerModeFlagsets:
		if apiKey == "" {
			return nil, fmt.Errorf("no API key provided")
		}

		flagsetName, exists := m.APIKeysToFlagSetName[apiKey]
		if !exists {
			return nil, fmt.Errorf("flagset not found for API key")
		}
		flagset, exists := m.FlagSets[flagsetName]
		if !exists {
			return nil, fmt.Errorf("impossible to find the flagset with the name %s", flagsetName)
		}
		return flagset, nil
	default:
		if m.DefaultFlagSet == nil {
			return nil, fmt.Errorf("no configured flagset")
		}
		return m.DefaultFlagSet, nil
	}
}

// FlagSetName returns the name of the flagset linked to the API Key
func (m *flagsetManagerImpl) FlagSetName(apiKey string) (string, error) {
	switch m.mode {
	case flagsetManagerModeFlagsets:
		if name, ok := m.APIKeysToFlagSetName[apiKey]; ok {
			return name, nil
		}
		return "", fmt.Errorf("no flag set associated to the API key")
	default:
		return utils.DefaultFlagSetName, nil
	}
}

// AllFlagSets returns the flag sets of the flagset manager.
func (m *flagsetManagerImpl) AllFlagSets() (map[string]*ffclient.GoFeatureFlag, error) {
	switch m.mode {
	case flagsetManagerModeFlagsets:
		if len(m.FlagSets) == 0 {
			return nil, fmt.Errorf("no flagsets configured")
		}
		return m.FlagSets, nil
	default:
		if m.DefaultFlagSet == nil {
			return nil, fmt.Errorf("no default flagset configured")
		}
		return map[string]*ffclient.GoFeatureFlag{
			utils.DefaultFlagSetName: m.DefaultFlagSet,
		}, nil
	}
}

// Default returns the default flagset
func (m *flagsetManagerImpl) Default() *ffclient.GoFeatureFlag {
	return m.DefaultFlagSet
}

// IsDefaultFlagSet returns true if the manager is in default mode (no flagsets configured)
func (m *flagsetManagerImpl) IsDefaultFlagSet() bool {
	return m.mode == flagsetManagerModeDefault
}

func (m *flagsetManagerImpl) Close() {
	if m.DefaultFlagSet != nil {
		m.DefaultFlagSet.Close()
	}
	for _, flagset := range m.FlagSets {
		flagset.Close()
	}
}

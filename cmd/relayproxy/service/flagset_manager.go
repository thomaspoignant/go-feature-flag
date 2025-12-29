package service

import (
	"errors"
	"fmt"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
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
	// OnConfigChange is called when the configuration changes
	OnConfigChange(newConfig *config.Config)
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

	// Logger is the logger for the flagset manager.
	logger *zap.Logger
}

// NewFlagsetManager is creating a new FlagsetManager.
// It is used to retrieve the flagset linked to the API Key.
func NewFlagsetManager(
	config *config.Config, logger *zap.Logger, notifiers []notifier.Notifier) (FlagsetManager, error) {
	if config == nil {
		return nil, fmt.Errorf("configuration is nil")
	}

	var flagsetMngr FlagsetManager
	var err error
	if config.IsUsingFlagsets() {
		// flagsets mode: create flagsets based on the `flagsets` array in the configuration.
		// The top-level retriever/exporter/etc. configuration is ignored in this mode.
		flagsetMngr, err = newFlagsetManagerWithFlagsets(config, logger, notifiers)
	} else {
		// default mode: use the top-level configuration to create a single default flagset.
		flagsetMngr, err = newFlagsetManagerWithDefaultConfig(config, logger, notifiers)
	}
	if err != nil {
		return nil, err
	}
	// Attach a callback to the flagset manager to be called when the configuration changes
	config.AttachConfigChangeCallback(flagsetMngr.OnConfigChange)
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
		logger:         logger,
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

			startLog := "no flagset name provided"
			if flagset.Name == utils.DefaultFlagSetName {
				startLog = "using 'default' as a flagset name"
			}
			logMessage := startLog + ", generating a default flagset name. This is not recommended. " +
				"Not having a flagset name will not allow you to change API Keys associated to the flagset during runtime."
			logger.Warn(logMessage, zap.String("flagset", flagSetName))
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
		logger:               logger,
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

// Close closes the flagset manager
func (m *flagsetManagerImpl) Close() {
	if m.DefaultFlagSet != nil {
		m.DefaultFlagSet.Close()
	}
	for _, flagset := range m.FlagSets {
		flagset.Close()
	}
}

// OnConfigChange is called when the configuration changes
func (m *flagsetManagerImpl) OnConfigChange(newConfig *config.Config) {
	// dont allow to switch from default to flagsets mode (or the opposite) during runtime
	if (newConfig.IsUsingFlagsets() && m.mode == flagsetManagerModeDefault) ||
		(!newConfig.IsUsingFlagsets() && m.mode == flagsetManagerModeFlagsets) {
		m.logger.Error("switching from default to flagsets mode (or the opposite) is not supported during runtime")
		return
	}

	switch m.mode {
	case flagsetManagerModeDefault:
		m.onConfigChangeWithDefault(newConfig)
	case flagsetManagerModeFlagsets:
		m.logger.Debug("flagsets mode is not supported yet")
		// m.onConfigChangeWithFlagsets(newConfig)
	}
}

// func (m *flagsetManagerImpl) onConfigChangeWithFlagsets(newConfig *config.Config) {
// 	// TODO: implement the logic to change the flagsets
// }

// onConfigChangeWithDefault is called when the configuration changes in default mode.
// The only configuration that can be changed is the API Keys and the AuthorizedKeys.
// All the other configuration changes are not supported.
func (m *flagsetManagerImpl) onConfigChangeWithDefault(newConfig *config.Config) {
	reloadAPIKeys := false
	// on default mode, we can only change the API Keys, all the other configuration changes are not supported
	// We need to read the current values with proper locking to avoid data races
	currentAuthorizedKeys := m.config.GetAuthorizedKeys()
	currentAPIKeys := m.config.GetAPIKeys()
	newAuthorizedKeys := newConfig.GetAuthorizedKeys()
	newAPIKeys := newConfig.GetAPIKeys()

	if !cmp.Equal(m.config, newConfig,
		cmpopts.IgnoreUnexported(config.Config{}), cmpopts.IgnoreFields(config.Config{}, "APIKeys", "AuthorizedKeys")) {
		m.logger.Warn("Configuration changed not supported: only API Keys and AuthorizedKeys can be " +
			"changed during runtime in default mode")
	}

	if !cmp.Equal(currentAuthorizedKeys, newAuthorizedKeys, cmpopts.IgnoreUnexported(config.APIKeys{})) {
		m.logger.Info("Configuration changed: reloading the AuthorizedKeys")
		m.config.SetAuthorizedKeys(newAuthorizedKeys)
		reloadAPIKeys = true
	}

	// nolint: staticcheck
	if !cmp.Equal(currentAPIKeys, newAPIKeys) {
		m.logger.Info("Configuration changed: reloading the APIKeys")
		// nolint: staticcheck
		m.config.SetAPIKeys(newAPIKeys)
		reloadAPIKeys = true
	}

	if reloadAPIKeys {
		m.config.ForceReloadAPIKeys()
	}
}

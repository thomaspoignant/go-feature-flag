package service

import (
	"errors"
	"fmt"
	"reflect"
	"sync"

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
	// GetFlagSet returns the flag set linked to the API Key
	GetFlagSet(apiKey string) (*ffclient.GoFeatureFlag, error)
	// GetFlagSetName returns the name of the flagset linked to the API Key
	GetFlagSetName(apiKey string) (string, error)
	// GetFlagSets returns all flag sets of the flagset manager
	GetFlagSets() (map[string]*ffclient.GoFeatureFlag, error)
	// GetDefaultFlagSet returns the default flagset
	GetDefaultFlagSet() *ffclient.GoFeatureFlag
	// IsDefaultFlagSet returns true if the manager is in default mode (no flagsets configured)
	IsDefaultFlagSet() bool
	// ReloadFlagsets reloads flagsets from the new configuration.
	// It validates that existing flagsets haven't been modified, and adds/removes flagsets as needed.
	// Returns an error if any existing flagset has been modified.
	ReloadFlagsets(newConfig *config.Config, logger *zap.Logger, notifiers []notifier.Notifier) error
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

	// mu protects concurrent access to flagsets
	mu sync.RWMutex
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

// GetFlagSet is returning the flag set linked to the API Key
func (m *flagsetManagerImpl) GetFlagSet(apiKey string) (*ffclient.GoFeatureFlag, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

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

// GetFlagSetName returns the name of the flagset linked to the API Key
func (m *flagsetManagerImpl) GetFlagSetName(apiKey string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

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

// GetFlagSets returns the flag sets of the flagset manager.
func (m *flagsetManagerImpl) GetFlagSets() (map[string]*ffclient.GoFeatureFlag, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	switch m.mode {
	case flagsetManagerModeFlagsets:
		if len(m.FlagSets) == 0 {
			return nil, fmt.Errorf("no flagsets configured")
		}
		// Return a copy to prevent external modifications
		result := make(map[string]*ffclient.GoFeatureFlag, len(m.FlagSets))
		for k, v := range m.FlagSets {
			result[k] = v
		}
		return result, nil
	default:
		if m.DefaultFlagSet == nil {
			return nil, fmt.Errorf("no default flagset configured")
		}
		return map[string]*ffclient.GoFeatureFlag{
			utils.DefaultFlagSetName: m.DefaultFlagSet,
		}, nil
	}
}

// GetDefaultFlagSet returns the default flagset
func (m *flagsetManagerImpl) GetDefaultFlagSet() *ffclient.GoFeatureFlag {
	return m.DefaultFlagSet
}

// IsDefaultFlagSet returns true if the manager is in default mode (no flagsets configured)
func (m *flagsetManagerImpl) IsDefaultFlagSet() bool {
	return m.mode == flagsetManagerModeDefault
}

// ReloadFlagsets reloads flagsets from the new configuration.
// It validates that existing flagsets haven't been modified, and adds/removes flagsets as needed.
// Returns an error if any existing flagset has been modified.
func (m *flagsetManagerImpl) ReloadFlagsets(newConfig *config.Config, logger *zap.Logger, notifiers []notifier.Notifier) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.validateReloadPreconditions(newConfig); err != nil {
		return err
	}

	currentMappings := m.buildCurrentFlagsetMappings()

	if err := m.validateFlagsetChanges(newConfig, currentMappings); err != nil {
		return err
	}

	newFlagsets, newAPIKeysToFlagSetName, err := m.createNewFlagsets(newConfig, logger, notifiers)
	if err != nil {
		return err
	}

	m.closeRemovedFlagsets(newFlagsets, logger)
	m.updateFlagsets(newFlagsets, newAPIKeysToFlagSetName, newConfig, logger)

	return nil
}

// validateReloadPreconditions checks if reload is allowed
func (m *flagsetManagerImpl) validateReloadPreconditions(newConfig *config.Config) error {
	if m.mode == flagsetManagerModeDefault {
		return fmt.Errorf("cannot reload flagsets in default mode")
	}
	if len(newConfig.FlagSets) == 0 {
		return fmt.Errorf("cannot reload: new configuration has no flagsets")
	}
	return nil
}

type flagsetMappings struct {
	apiKeyToConfig map[string]*config.FlagSet
	apiKeyToName   map[string]string
}

// buildCurrentFlagsetMappings builds maps of API keys to flagset configurations and names
func (m *flagsetManagerImpl) buildCurrentFlagsetMappings() flagsetMappings {
	mappings := flagsetMappings{
		apiKeyToConfig: make(map[string]*config.FlagSet),
		apiKeyToName:   make(map[string]string),
	}

	for i := range m.config.FlagSets {
		flagset := &m.config.FlagSets[i]
		flagSetName := normalizeFlagsetName(flagset.Name)
		for _, apiKey := range flagset.APIKeys {
			mappings.apiKeyToConfig[apiKey] = flagset
			mappings.apiKeyToName[apiKey] = flagSetName
		}
	}
	return mappings
}

// normalizeFlagsetName returns the flagset name or generates a UUID if empty/default
func normalizeFlagsetName(name string) string {
	if name == "" || name == utils.DefaultFlagSetName {
		return uuid.New().String()
	}
	return name
}

// validateFlagsetChanges validates that existing flagsets haven't been modified
func (m *flagsetManagerImpl) validateFlagsetChanges(newConfig *config.Config, currentMappings flagsetMappings) error {
	for _, newFlagset := range newConfig.FlagSets {
		if err := m.validateSingleFlagset(&newFlagset, currentMappings); err != nil {
			return err
		}
	}
	return nil
}

// validateSingleFlagset validates a single flagset hasn't been modified
func (m *flagsetManagerImpl) validateSingleFlagset(newFlagset *config.FlagSet, currentMappings flagsetMappings) error {
	existingConfig, existingFlagsetName := m.findMatchingFlagset(newFlagset, currentMappings)
	if existingConfig == nil {
		return nil // New flagset, no validation needed
	}

	if !flagsetConfigsEqual(existingConfig, newFlagset) {
		return fmt.Errorf("flagset '%s' has been modified, reload rejected", existingFlagsetName)
	}

	if err := m.validateFlagsetNameChange(existingConfig, newFlagset, existingFlagsetName); err != nil {
		return err
	}

	return m.validateAPIKeyMovements(newFlagset, currentMappings)
}

// findMatchingFlagset finds an existing flagset that matches the new one by API key
func (m *flagsetManagerImpl) findMatchingFlagset(newFlagset *config.FlagSet, currentMappings flagsetMappings) (*config.FlagSet, string) {
	for _, apiKey := range newFlagset.APIKeys {
		if existing, exists := currentMappings.apiKeyToConfig[apiKey]; exists {
			return existing, currentMappings.apiKeyToName[apiKey]
		}
	}
	return nil, ""
}

// validateFlagsetNameChange validates that flagset names haven't changed
func (m *flagsetManagerImpl) validateFlagsetNameChange(existingConfig, newFlagset *config.FlagSet, existingFlagsetName string) error {
	existingHasRealName := hasRealFlagsetName(existingConfig.Name)
	newHasRealName := hasRealFlagsetName(newFlagset.Name)

	if existingHasRealName && newHasRealName && existingFlagsetName != normalizeFlagsetName(newFlagset.Name) {
		return fmt.Errorf("flagset configuration changed (name changed from '%s' to '%s'), reload rejected", existingFlagsetName, normalizeFlagsetName(newFlagset.Name))
	}
	return nil
}

// hasRealFlagsetName checks if a flagset has a real (non-empty, non-default) name
func hasRealFlagsetName(name string) bool {
	return name != "" && name != utils.DefaultFlagSetName
}

// validateAPIKeyMovements validates that API keys haven't moved between flagsets
func (m *flagsetManagerImpl) validateAPIKeyMovements(newFlagset *config.FlagSet, currentMappings flagsetMappings) error {
	newFlagSetName := normalizeFlagsetName(newFlagset.Name)
	newHasRealName := hasRealFlagsetName(newFlagset.Name)

	for _, apiKey := range newFlagset.APIKeys {
		if oldFlagsetConfig, exists := currentMappings.apiKeyToConfig[apiKey]; exists {
			oldHasRealName := hasRealFlagsetName(oldFlagsetConfig.Name)
			if oldHasRealName && newHasRealName {
				oldRealName := oldFlagsetConfig.Name
				if oldRealName != newFlagSetName {
					return fmt.Errorf("API key moved from flagset '%s' to '%s', reload rejected", oldRealName, newFlagSetName)
				}
			}
		}
	}
	return nil
}

// createNewFlagsets creates new flagset clients from the configuration
func (m *flagsetManagerImpl) createNewFlagsets(newConfig *config.Config, logger *zap.Logger, notifiers []notifier.Notifier) (map[string]*ffclient.GoFeatureFlag, map[string]string, error) {
	newFlagsets := make(map[string]*ffclient.GoFeatureFlag)
	newAPIKeysToFlagSetName := make(map[string]string)

	for index, flagset := range newConfig.FlagSets {
		client, err := NewGoFeatureFlagClient(&flagset, logger, notifiers)
		if err != nil {
			logger.Error(
				"failed to create goff client for flagset during reload",
				zap.Int("flagset_index", index),
				zap.String("flagset", flagset.Name),
				zap.Error(err),
			)
			return nil, nil, fmt.Errorf("failed to create flagset '%s': %w", flagset.Name, err)
		}

		flagSetName := normalizeFlagsetName(flagset.Name)
		newFlagsets[flagSetName] = client
		for _, apiKey := range flagset.APIKeys {
			newAPIKeysToFlagSetName[apiKey] = flagSetName
		}
	}

	return newFlagsets, newAPIKeysToFlagSetName, nil
}

// closeRemovedFlagsets closes flagsets that are no longer in the configuration
func (m *flagsetManagerImpl) closeRemovedFlagsets(newFlagsets map[string]*ffclient.GoFeatureFlag, logger *zap.Logger) {
	for name, flagset := range m.FlagSets {
		if _, exists := newFlagsets[name]; !exists {
			logger.Info("closing removed flagset", zap.String("flagset", name))
			flagset.Close()
		}
	}
}

// updateFlagsets updates the manager with the new flagsets
func (m *flagsetManagerImpl) updateFlagsets(newFlagsets map[string]*ffclient.GoFeatureFlag, newAPIKeysToFlagSetName map[string]string, newConfig *config.Config, logger *zap.Logger) {
	m.FlagSets = newFlagsets
	m.APIKeysToFlagSetName = newAPIKeysToFlagSetName
	m.config = newConfig

	logger.Info("flagsets reloaded successfully",
		zap.Int("flagsets_count", len(newFlagsets)),
	)
}

// flagsetConfigsEqual compares two flagset configurations, excluding APIKeys
func flagsetConfigsEqual(a, b *config.FlagSet) bool {
	// Create copies without APIKeys for comparison
	aCopy := *a
	bCopy := *b
	aCopy.APIKeys = nil
	bCopy.APIKeys = nil

	return reflect.DeepEqual(aCopy, bCopy)
}

func (m *flagsetManagerImpl) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.DefaultFlagSet != nil {
		m.DefaultFlagSet.Close()
	}
	for _, flagset := range m.FlagSets {
		flagset.Close()
	}
}

package service

import (
	"errors"
	"fmt"
	"sync"

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

	// apiKeysMutex protects concurrent access to APIKeysToFlagSetName and FlagSets map
	apiKeysMutex sync.RWMutex

	// configChangeMutex protects concurrent execution of onConfigChangeWithFlagsets
	// to ensure configuration changes are processed sequentially
	configChangeMutex sync.Mutex

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

		m.apiKeysMutex.RLock()
		flagsetName, exists := m.APIKeysToFlagSetName[apiKey]
		if !exists {
			m.apiKeysMutex.RUnlock()
			return nil, fmt.Errorf("flagset not found for API key")
		}
		flagset, exists := m.FlagSets[flagsetName]
		m.apiKeysMutex.RUnlock()
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
		m.apiKeysMutex.RLock()
		name, ok := m.APIKeysToFlagSetName[apiKey]
		m.apiKeysMutex.RUnlock()
		if ok {
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
		m.apiKeysMutex.RLock()
		defer m.apiKeysMutex.RUnlock()
		if len(m.FlagSets) == 0 {
			return nil, fmt.Errorf("no flagsets configured")
		}
		// Return a copy to avoid external modifications
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
	m.apiKeysMutex.RLock()
	flagsets := make([]*ffclient.GoFeatureFlag, 0, len(m.FlagSets))
	for _, flagset := range m.FlagSets {
		flagsets = append(flagsets, flagset)
	}
	m.apiKeysMutex.RUnlock()
	for _, flagset := range flagsets {
		flagset.Close()
	}
}

// OnConfigChange is called when the configuration changes
func (m *flagsetManagerImpl) OnConfigChange(newConfig *config.Config) {
	if err := newConfig.IsValid(); err != nil {
		m.logger.Error("the new configuration is invalid, it will not be applied", zap.Error(err))
		m.logger.Debug("invalid configuration:", zap.Error(err), zap.Any("newConfig", newConfig))
		return
	}

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
		m.onConfigChangeWithFlagsets(newConfig)
	}
}

// onConfigChangeWithFlagsets is called when the configuration changes in flagsets mode.
// It handles additions, removals, modifications, and API key changes for flagsets.
// This method is synchronized to ensure configuration changes are processed sequentially.
func (m *flagsetManagerImpl) onConfigChangeWithFlagsets(newConfig *config.Config) {
	m.configChangeMutex.Lock()
	defer m.configChangeMutex.Unlock()

	currentFlagsets := m.buildCurrentFlagsetsMap()
	processedFlagsets := m.processNewFlagsets(newConfig.FlagSets, currentFlagsets)
	m.removeDeletedFlagsets(currentFlagsets, processedFlagsets)
	m.config.ForceReloadAPIKeys()
}

// buildCurrentFlagsetsMap builds a map of current named flagsets.
func (m *flagsetManagerImpl) buildCurrentFlagsetsMap() map[string]config.FlagSet {
	currentFlagsets := make(map[string]config.FlagSet)
	for _, fs := range m.config.GetFlagSets() {
		if fs.Name != "" && fs.Name != utils.DefaultFlagSetName {
			currentFlagsets[fs.Name] = fs
		}
	}
	return currentFlagsets
}

// processNewFlagsets processes new flagsets from the config and returns a set of processed flagset names.
func (m *flagsetManagerImpl) processNewFlagsets(
	newFlagsets []config.FlagSet,
	currentFlagsets map[string]config.FlagSet,
) map[string]bool {
	processedFlagsets := make(map[string]bool)
	for _, newFS := range newFlagsets {
		if m.rejectUnnamedFlagset(newFS) {
			continue
		}
		processedFlagsets[newFS.Name] = true
		m.processFlagsetChange(newFS, currentFlagsets)
	}
	return processedFlagsets
}

// rejectUnnamedFlagset rejects unnamed flagsets and logs an error. Returns true if rejected.
func (m *flagsetManagerImpl) rejectUnnamedFlagset(flagset config.FlagSet) bool {
	if flagset.Name == "" || flagset.Name == utils.DefaultFlagSetName {
		m.logger.Error("Configuration change rejected: unnamed flagsets cannot be added or modified dynamically",
			zap.String("flagset", flagset.Name))
		return true
	}
	return false
}

// processFlagsetChange processes a single flagset change (add, modify, or API key update).
func (m *flagsetManagerImpl) processFlagsetChange(newFS config.FlagSet, currentFlagsets map[string]config.FlagSet) {
	currentFS, exists := currentFlagsets[newFS.Name]
	if !exists {
		m.addFlagset(newFS)
		return
	}

	if m.hasCommonFlagSetChanged(currentFS, newFS) {
		m.logger.Error("Configuration change rejected: flagset modification not allowed",
			zap.String("flagset", newFS.Name),
			zap.String("reason", "only API key changes are allowed for existing flagsets"))
		return
	}

	m.processFlagsetAPIKeyChange(newFS)
}

// removeDeletedFlagsets removes flagsets that no longer exist in the new config.
func (m *flagsetManagerImpl) removeDeletedFlagsets(
	currentFlagsets map[string]config.FlagSet,
	processedFlagsets map[string]bool,
) {
	for name := range currentFlagsets {
		if !processedFlagsets[name] {
			m.removeFlagset(name)
		}
	}
}

// processFlagsetAPIKeyChange handles API key changes for a single flagset.
func (m *flagsetManagerImpl) processFlagsetAPIKeyChange(newConfigFlagset config.FlagSet) {
	flagsetName := newConfigFlagset.Name
	if flagsetName == "" || flagsetName == utils.DefaultFlagSetName {
		return
	}
	currentAPIKeys, err := m.config.GetFlagSetAPIKeys(flagsetName)
	if err != nil {
		return
	}
	if cmp.Equal(currentAPIKeys, newConfigFlagset.APIKeys) {
		return
	}

	m.logger.Info("Configuration changed: updating the APIKeys for flagset",
		zap.String("flagset", flagsetName))

	if err = m.config.SetFlagSetAPIKeys(flagsetName, newConfigFlagset.APIKeys); err != nil {
		m.logger.Error("failed to update the APIKeys for flagset", zap.Error(err))
		return
	}
	m.updateAPIKeysMapping(flagsetName, currentAPIKeys, newConfigFlagset.APIKeys)
}

// updateAPIKeysMapping updates the APIKeysToFlagSetName map with the new API keys.
func (m *flagsetManagerImpl) updateAPIKeysMapping(flagsetName string, oldKeys, newKeys []string) {
	m.apiKeysMutex.Lock()
	defer m.apiKeysMutex.Unlock()

	for _, apiKey := range oldKeys {
		delete(m.APIKeysToFlagSetName, apiKey)
	}
	for _, apiKey := range newKeys {
		m.APIKeysToFlagSetName[apiKey] = flagsetName
	}
}

// hasCommonFlagSetChanged checks if the CommonFlagSet configuration has changed.
// This is used to detect forbidden modifications (everything except APIKeys).
func (m *flagsetManagerImpl) hasCommonFlagSetChanged(old, newFlagset config.FlagSet) bool {
	return !cmp.Equal(old.CommonFlagSet, newFlagset.CommonFlagSet,
		cmpopts.IgnoreUnexported(config.CommonFlagSet{}))
}

// addFlagset creates a new flagset dynamically and starts it.
func (m *flagsetManagerImpl) addFlagset(flagset config.FlagSet) {
	// Create the GoFeatureFlag client (this automatically starts polling)
	client, err := NewGoFeatureFlagClient(&flagset, m.logger, nil)
	if err != nil {
		m.logger.Error("failed to create goff client for new flagset",
			zap.String("flagset", flagset.Name),
			zap.Error(err))
		return
	}

	// Update config struct first (before adding to maps)
	if err := m.config.AddFlagSet(flagset); err != nil {
		m.logger.Error("failed to add flagset to config",
			zap.String("flagset", flagset.Name),
			zap.Error(err))
		client.Close()
		return
	}

	// Add to FlagSets map and update API keys mapping (both protected by mutex)
	m.apiKeysMutex.Lock()
	m.FlagSets[flagset.Name] = client
	for _, apiKey := range flagset.APIKeys {
		m.APIKeysToFlagSetName[apiKey] = flagset.Name
	}
	m.apiKeysMutex.Unlock()

	m.logger.Info("Configuration changed: added new flagset",
		zap.String("flagset", flagset.Name))
}

// removeFlagset removes a flagset dynamically and gracefully stops it.
// The order of operations ensures atomicity: config is updated first, then runtime state.
func (m *flagsetManagerImpl) removeFlagset(flagsetName string) {
	// Step 1: Get the client and API keys before making any changes
	m.apiKeysMutex.RLock()
	client, exists := m.FlagSets[flagsetName]
	apiKeysToRemove, err := m.config.GetFlagSetAPIKeys(flagsetName)
	m.apiKeysMutex.RUnlock()

	if !exists {
		m.logger.Warn("flagset not found for removal",
			zap.String("flagset", flagsetName))
		return
	}

	if err != nil {
		m.logger.Error("failed to get API keys for flagset to remove, API key mapping may be inconsistent",
			zap.String("flagset", flagsetName),
			zap.Error(err))
		apiKeysToRemove = []string{}
	}

	// Step 2: Update the source of truth (config) first
	// If this fails, abort the operation to maintain consistency
	if err := m.config.RemoveFlagSet(flagsetName); err != nil {
		m.logger.Error("failed to remove flagset from config",
			zap.String("flagset", flagsetName),
			zap.Error(err))
		return
	}

	// Step 3: Update runtime state (maps) - protected by mutex
	m.apiKeysMutex.Lock()
	delete(m.FlagSets, flagsetName)
	for _, apiKey := range apiKeysToRemove {
		delete(m.APIKeysToFlagSetName, apiKey)
	}
	m.apiKeysMutex.Unlock()

	// Step 4: Gracefully stop the client (stops polling, flushes exports, closes notifiers)
	// Do this after releasing the lock to avoid blocking other operations
	client.Close()

	m.logger.Info("Configuration changed: removed flagset",
		zap.String("flagset", flagsetName))
}

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

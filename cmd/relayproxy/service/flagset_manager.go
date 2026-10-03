package service

import (
	"errors"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"sync"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/proxynotifier"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service/stream"
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

	// flagsetsMutex protects concurrent access to FlagSets and APIKeysToFlagSetName.
	// Both maps can be mutated at runtime when flagsets are added or removed, so every
	// read or write of these maps must be done while holding this lock.
	flagsetsMutex sync.RWMutex

	// notifiers are the relay-proxy level notifiers (e.g. prometheus, websocket) passed at
	// construction time. They are stored so we can re-create a GoFeatureFlag client when a
	// flagset is added at runtime.
	notifiers []notifier.Notifier

	// sseService is used to create a per-flagset SSE notifier when a flagset is added at
	// runtime. It can be nil.
	sseService stream.SSEService

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
// sseService is optional: when non-nil a per-flagset SSE notifier is created
// so that flag-change events are scoped to the correct flagset.
func NewFlagsetManager(
	config *config.Config, logger *zap.Logger, notifiers []notifier.Notifier,
	sseService stream.SSEService,
) (FlagsetManager, error) {
	if config == nil {
		return nil, fmt.Errorf("configuration is nil")
	}

	var flagsetMngr FlagsetManager
	var err error
	if config.IsUsingFlagsets() {
		flagsetMngr, err = newFlagsetManagerWithFlagsets(config, logger, notifiers, sseService)
	} else {
		flagsetMngr, err = newFlagsetManagerWithDefaultConfig(config, logger, notifiers, sseService)
	}
	if err != nil {
		return nil, err
	}
	config.AttachConfigChangeCallback(flagsetMngr.OnConfigChange)
	return flagsetMngr, nil
}

func newFlagsetManagerWithDefaultConfig(
	c *config.Config, logger *zap.Logger, notifiers []notifier.Notifier,
	sseService stream.SSEService,
) (FlagsetManager, error) {
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
	allNotifiers := appendSSENotifier(notifiers, sseService, utils.DefaultFlagSetName)
	client, err := NewGoFeatureFlagClient(&defaultFlagSet, logger, allNotifiers)
	if err != nil {
		return nil, err
	}
	return &flagsetManagerImpl{
		DefaultFlagSet: client,
		notifiers:      notifiers,
		sseService:     sseService,
		config:         c,
		mode:           flagsetManagerModeDefault,
		logger:         logger,
	}, nil
}

func newFlagsetManagerWithFlagsets(
	config *config.Config, logger *zap.Logger, notifiers []notifier.Notifier,
	sseService stream.SSEService,
) (FlagsetManager, error) {
	flagsets := make(map[string]*ffclient.GoFeatureFlag)
	apiKeysToFlagSet := make(map[string]string)

	for index, flagset := range config.FlagSets {
		flagSetName := flagset.Name
		if !isNamedFlagset(flagset) {
			flagSetName = uuid.New().String()

			startLog := "no flagset name provided"
			if flagset.Name == utils.DefaultFlagSetName {
				startLog = "using 'default' as a flagset name"
			}
			logMessage := startLog + ", generating a default flagset name. This is not recommended. " +
				"Not having a flagset name will not allow you to change API Keys associated to the flagset during runtime."
			logger.Warn(logMessage, zap.String("flagset", flagSetName))
		}

		allNotifiers := appendSSENotifier(notifiers, sseService, flagSetName)
		client, err := NewGoFeatureFlagClient(&flagset, logger, allNotifiers)
		if err != nil {
			logger.Error(
				"failed to create goff client for flagset",
				zap.Int("flagset_index", index),
				zap.String("flagset", flagset.Name),
				zap.Error(err),
			)
			continue
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
		notifiers:            notifiers,
		sseService:           sseService,
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

		m.flagsetsMutex.RLock()
		flagsetName, exists := m.APIKeysToFlagSetName[apiKey]
		flagset, flagsetExists := m.FlagSets[flagsetName]
		m.flagsetsMutex.RUnlock()
		if !exists {
			return nil, fmt.Errorf("flagset not found for API key")
		}
		if !flagsetExists {
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
		m.flagsetsMutex.RLock()
		name, ok := m.APIKeysToFlagSetName[apiKey]
		m.flagsetsMutex.RUnlock()
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
		m.flagsetsMutex.RLock()
		defer m.flagsetsMutex.RUnlock()
		if len(m.FlagSets) == 0 {
			return nil, fmt.Errorf("no flagsets configured")
		}
		// Return a shallow copy so callers can iterate safely even if flagsets are
		// added or removed concurrently.
		out := make(map[string]*ffclient.GoFeatureFlag, len(m.FlagSets))
		maps.Copy(out, m.FlagSets)
		return out, nil
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
	// Snapshot the clients under the lock, then close them outside of it. Close() flushes the
	// exporters and can block, so holding the lock for every flush would needlessly block a
	// concurrent config reload for the whole teardown. This mirrors removeFlagset.
	m.flagsetsMutex.RLock()
	clients := make([]*ffclient.GoFeatureFlag, 0, len(m.FlagSets))
	for _, flagset := range m.FlagSets {
		clients = append(clients, flagset)
	}
	m.flagsetsMutex.RUnlock()

	for _, client := range clients {
		client.Close()
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

// onConfigChangeWithFlagsets reconciles the running flagsets with the new configuration.
// In flagsets mode we support, without a restart:
//   - adding a new flagset (a new GoFeatureFlag client is created),
//   - removing a flagset (its GoFeatureFlag client is closed),
//   - changing the API keys of an existing flagset.
//
// Any other modification of an existing flagset (retriever, exporter, notifier, polling
// interval, ...) is NOT supported and is rejected with an error: the change is ignored at
// runtime and will only be applied the next time the relay proxy is restarted.
//
// Flagsets are tracked by their explicit name. Flagsets without a name (or named "default")
// get a generated name at startup and therefore cannot be added or removed at runtime.
func (m *flagsetManagerImpl) onConfigChangeWithFlagsets(newConfig *config.Config) {
	currentFlagsets := m.config.GetFlagSets()
	currentByName := indexNamedFlagsets(currentFlagsets)
	newByName := indexNamedFlagsets(newConfig.FlagSets)

	// 1. Remove the flagsets that are no longer in the configuration.
	for name, currentFlagset := range currentByName {
		if _, stillPresent := newByName[name]; !stillPresent {
			m.removeFlagset(currentFlagset)
		}
	}

	// 2. Add the flagsets that are new in the configuration.
	for name, newFlagset := range newByName {
		if _, alreadyExists := currentByName[name]; !alreadyExists {
			m.addFlagset(newFlagset)
		}
	}

	// 3. Reconcile the flagsets that are present in both configurations: only their API keys
	// can be changed at runtime, every other modification is rejected.
	for name, newFlagset := range newByName {
		if currentFlagset, ok := currentByName[name]; ok {
			m.reconcileExistingFlagset(currentFlagset, newFlagset)
		}
	}

	// Rebuild the API-key -> flagset routing for every named flagset from the reconciled
	// configuration. Doing it in one pass (instead of incremental per-flagset updates) makes
	// the result independent of the order in which the flagsets above were processed, which
	// matters when an API key is moved between flagsets in a single reload.
	m.rebuildNamedAPIKeysMapping(currentByName, newByName)

	m.warnIfUnnamedFlagsetsChanged(currentFlagsets, newConfig.FlagSets)
	m.config.ForceReloadAPIKeys()
}

// isNamedFlagset returns true if the flagset has an explicit, stable name that allows it to be
// managed at runtime. Flagsets without a name (or named "default") get a generated name at
// startup and cannot be added, removed or have their API keys changed at runtime.
func isNamedFlagset(flagset config.FlagSet) bool {
	return flagset.Name != "" && flagset.Name != utils.DefaultFlagSetName
}

// indexNamedFlagsets indexes the given flagsets by their name, ignoring flagsets that do not
// have an explicit name (empty or "default"): those cannot be tracked across reloads because
// they are assigned a generated name at startup.
func indexNamedFlagsets(flagsets []config.FlagSet) map[string]config.FlagSet {
	out := make(map[string]config.FlagSet, len(flagsets))
	for _, flagset := range flagsets {
		if !isNamedFlagset(flagset) {
			continue
		}
		out[flagset.Name] = flagset
	}
	return out
}

// addFlagset creates a new GoFeatureFlag client for the given flagset and registers it so it
// can serve evaluations immediately.
func (m *flagsetManagerImpl) addFlagset(flagset config.FlagSet) {
	allNotifiers := appendSSENotifier(m.notifiers, m.sseService, flagset.Name)
	client, err := NewGoFeatureFlagClient(&flagset, m.logger, allNotifiers)
	if err != nil {
		m.logger.Error(
			"failed to create the flagset added at runtime, it will not be available until the "+
				"error is fixed",
			zap.String("flagset", flagset.Name),
			zap.Error(err),
		)
		return
	}

	// Register in the authoritative config first. The API-key routing is rebuilt from the config
	// by rebuildNamedAPIKeysMapping, so a client that is not in the config would run without any
	// routing (unreachable). If the config update fails, close the client and skip rather than
	// leaving an orphan running flagset.
	if err := m.config.AddFlagSet(flagset); err != nil {
		m.logger.Error("failed to register the new flagset in the configuration; it will not be started",
			zap.String("flagset", flagset.Name), zap.Error(err))
		client.Close()
		return
	}

	// The API-key routing is (re)built in one pass by rebuildNamedAPIKeysMapping once all flagsets
	// have been reconciled.
	m.flagsetsMutex.Lock()
	m.FlagSets[flagset.Name] = client
	m.flagsetsMutex.Unlock()

	m.logger.Info("Configuration changed: flagset added at runtime",
		zap.String("flagset", flagset.Name))
}

// removeFlagset closes the GoFeatureFlag client of the given flagset and unregisters it.
func (m *flagsetManagerImpl) removeFlagset(flagset config.FlagSet) {
	// Stop routing to this flagset (remove its client and API keys) before closing it; the
	// remaining named flagsets' routing is rebuilt afterwards by rebuildNamedAPIKeysMapping.
	m.flagsetsMutex.Lock()
	client := m.FlagSets[flagset.Name]
	delete(m.FlagSets, flagset.Name)
	for _, apiKey := range flagset.APIKeys {
		delete(m.APIKeysToFlagSetName, apiKey)
	}
	m.flagsetsMutex.Unlock()

	// Close the client outside of the lock: Close() waits for the exporters to flush and may
	// block, we don't want to hold the lock (and block evaluation requests) while it happens.
	if client != nil {
		client.Close()
	}
	if err := m.config.RemoveFlagSet(flagset.Name); err != nil {
		m.logger.Error("failed to remove the flagset from the configuration",
			zap.String("flagset", flagset.Name), zap.Error(err))
	}
	m.logger.Info("Configuration changed: flagset removed at runtime",
		zap.String("flagset", flagset.Name))
}

// reconcileExistingFlagset reconciles a flagset that is present in both the current and the new
// configuration. Only the API keys can be changed at runtime: any other modification is
// rejected with an error and ignored until the next restart of the relay proxy.
func (m *flagsetManagerImpl) reconcileExistingFlagset(current, newFlagset config.FlagSet) {
	// slices.Equal is panic-safe on []string and treats nil and empty slices as equal.
	apiKeysChanged := !slices.Equal(current.APIKeys, newFlagset.APIKeys)

	// reflect.DeepEqual is used on purpose instead of cmp.Equal: it never panics on the
	// unexported fields that can be reached through the flagset configuration (e.g. the Kafka
	// exporter's *sarama.Config). Both values come from the same koanf parsing so their
	// representation is consistent.
	if !reflect.DeepEqual(current.CommonFlagSet, newFlagset.CommonFlagSet) {
		m.logger.Error(
			"modifying a flagset is not supported at runtime: the change is ignored and will "+
				"only be applied the next time the relay proxy is restarted. Only the API keys of a "+
				"flagset can be changed without a restart.",
			zap.String("flagset", newFlagset.Name),
		)
		if apiKeysChanged {
			m.logger.Error(
				"the API keys change for this flagset is also ignored because it is bundled with an "+
					"unsupported flagset modification: revert the other changes to apply the API keys "+
					"change at runtime, otherwise it will be applied on the next restart.",
				zap.String("flagset", newFlagset.Name),
			)
		}
		return
	}

	if apiKeysChanged {
		m.logger.Info("Configuration changed: updating the APIKeys for flagset",
			zap.String("flagset", newFlagset.Name))
		// Update only the authoritative configuration here; the routing map is rebuilt from it
		// by rebuildNamedAPIKeysMapping once all flagsets have been reconciled.
		if err := m.config.SetFlagSetAPIKeys(newFlagset.Name, newFlagset.APIKeys); err != nil {
			m.logger.Error("failed to update the APIKeys for flagset", zap.Error(err))
			return
		}
	}
}

// warnIfUnnamedFlagsetsChanged logs a warning when the number of unnamed (or "default" named)
// flagsets differs between the current and the new configuration: those flagsets cannot be
// added or removed at runtime and the change requires a restart.
func (m *flagsetManagerImpl) warnIfUnnamedFlagsetsChanged(current, newFlagsets []config.FlagSet) {
	if countUnnamedFlagsets(current) != countUnnamedFlagsets(newFlagsets) {
		m.logger.Warn(
			"adding or removing a flagset without an explicit name is not supported at runtime: " +
				"the change will only be applied the next time the relay proxy is restarted. " +
				"Give your flagsets a unique name to manage them at runtime.",
		)
	}
}

// countUnnamedFlagsets counts the flagsets that do not have an explicit name.
func countUnnamedFlagsets(flagsets []config.FlagSet) int {
	count := 0
	for _, flagset := range flagsets {
		if !isNamedFlagset(flagset) {
			count++
		}
	}
	return count
}

// rebuildNamedAPIKeysMapping rebuilds, in a single pass, the API-key -> flagset routing for
// every named flagset involved in the reload, from the reconciled (authoritative) configuration.
//
// This is done after the add/remove/reconcile passes so the resulting routing is independent of
// the order in which those passes processed the flagsets — in particular when an API key is moved
// from one flagset to another in the same reload.
//
// Routing entries that target unnamed (UUID-named) flagsets are preserved: those flagsets cannot
// be managed at runtime, so their mapping never changes here.
func (m *flagsetManagerImpl) rebuildNamedAPIKeysMapping(currentByName, newByName map[string]config.FlagSet) {
	// Authoritative API keys per named flagset after reconciliation. Reading from the config
	// (not from newByName) preserves the previous keys of flagsets whose modification was
	// rejected.
	reconciled := indexNamedFlagsets(m.config.GetFlagSets())

	m.flagsetsMutex.Lock()
	m.dropManagedAPIKeyRouting(currentByName, newByName)
	assigned, collisions := resolveNamedAPIKeyRouting(reconciled, m.FlagSets)
	maps.Copy(m.APIKeysToFlagSetName, assigned)
	m.flagsetsMutex.Unlock()

	// Log collisions outside the lock to keep the critical section minimal (the logger sink may
	// be slow or contended).
	logAPIKeyCollisions(m.logger, collisions)
}

// dropManagedAPIKeyRouting removes every routing entry that targets a named flagset involved in
// this reload (added, removed or kept). Entries targeting unnamed flagsets are left untouched.
// The caller must hold flagsetsMutex.
func (m *flagsetManagerImpl) dropManagedAPIKeyRouting(currentByName, newByName map[string]config.FlagSet) {
	for apiKey, target := range m.APIKeysToFlagSetName {
		_, current := currentByName[target]
		_, added := newByName[target]
		if current || added {
			delete(m.APIKeysToFlagSetName, apiKey)
		}
	}
}

// resolveNamedAPIKeyRouting computes the API-key -> flagset-name routing for the reconciled named
// flagsets that have a running client. It returns the unambiguous assignments and the keys that
// are configured on more than one flagset.
//
// A key can collide here even though newConfig.IsValid() deduped the incoming config: when an
// API-key move onto flagset A is accepted while a forbidden modification on flagset B (which still
// holds that key) is rejected, the reconciled config ends up with the same key on both A and B.
// Such a key must not route non-deterministically, so we fail closed — it is left out of the
// returned assignments (resolves to no flagset) and reported as a collision.
func resolveNamedAPIKeyRouting(
	reconciled map[string]config.FlagSet,
	running map[string]*ffclient.GoFeatureFlag,
) (assigned map[string]string, collisions map[string][2]string) {
	assigned = make(map[string]string)
	collisions = make(map[string][2]string)
	for name, flagset := range reconciled {
		if _, ok := running[name]; !ok {
			continue
		}
		for _, apiKey := range flagset.APIKeys {
			if prev, ok := assigned[apiKey]; ok && prev != name {
				collisions[apiKey] = [2]string{prev, name}
				continue
			}
			assigned[apiKey] = name
		}
	}
	for apiKey := range collisions {
		delete(assigned, apiKey)
	}
	return assigned, collisions
}

// logAPIKeyCollisions reports keys configured on multiple flagsets. It does not log the API key
// itself (it is a secret); it logs the flagsets the key collides between.
func logAPIKeyCollisions(logger *zap.Logger, collisions map[string][2]string) {
	for _, pair := range collisions {
		logger.Error(
			"an API key is configured on multiple flagsets after reconciliation; its routing is "+
				"disabled until the configuration is fixed. This can happen when an API-key move is "+
				"bundled with a rejected flagset modification.",
			zap.String("flagset", pair[0]), zap.String("otherFlagset", pair[1]),
		)
	}
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

// appendSSENotifier returns a copy of notifiers with an SSE notifier appended
// when sseService is non-nil.
func appendSSENotifier(
	notifiers []notifier.Notifier, sseService stream.SSEService, flagsetName string,
) []notifier.Notifier {
	if sseService == nil {
		return notifiers
	}
	out := make([]notifier.Notifier, len(notifiers), len(notifiers)+1)
	copy(out, notifiers)
	return append(out, proxynotifier.NewNotifierSSE(sseService, flagsetName))
}

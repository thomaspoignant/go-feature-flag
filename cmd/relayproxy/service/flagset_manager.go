package service

import (
	"fmt"

	"github.com/google/uuid"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"go.uber.org/zap"
)

const defaultFlagSetName = "default"

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
}

// NewFlagsetManager is creating a new FlagsetManager.
// It is used to retrieve the flagset linked to the API Key.
func NewFlagsetManager(config *config.Config, logger *zap.Logger) (FlagsetManager, error) {
	if config == nil {
		return FlagsetManager{}, fmt.Errorf("configuration is nil")
	}

	// in case you are using the relay proxy without any flagset, we use the default configuration.
	if len(config.FlagSets) == 0 {
		client, err := NewGoFeatureFlagClient(&flagset, logger, []notifier.Notifier{})
		if err != nil {
			return FlagsetManager{}, err
		}
		return FlagsetManager{
			DefaultFlagSet: &client,
			config:         config,
		}, nil
	}

	// in case you are using the relay proxy with flagsets, we create the flagsets and map them to the APIKeys.
	// note that the default configuration is ignored in this case.
	flagsets := make(map[string]*ffclient.GoFeatureFlag)
	apiKeysToFlagSet := make(map[string]string)

	for _, flagset := range config.FlagSets {
		client, err := NewGoFeatureFlagClient(&flagset, logger, []notifier.Notifier{})
		if err != nil {
			logger.Error("failed to create goff client", zap.Error(err))
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

	// add default flagset
	defaultFlagSetConfig := prepareDefaultFlagSet(config)
	defaultFlagSet, err := NewGoFeatureFlagClient(&defaultFlagSetConfig, logger, []notifier.Notifier{})
	if err != nil {
		logger.Error("faild to create default flagset")
	}
	flagsets[defaultFlagSetName] = &defaultFlagSet
	return FlagsetManager{
		FlagSets:         flagsets,
		APIKeysToFlagSet: apiKeysToFlagSet,
		config:           config,
	}, nil
}

// GetFlagSet is returning the flag set linked to the API Key
func (m *FlagsetManager) GetFlagSet(apiKey string) *ffclient.GoFeatureFlag {
	keyType := m.config.GetAPIKeyType(apiKey)
	switch keyType {
	case config.FlagSetKeyType:
		return m.FlagSets[m.APIKeysToFlagSet[apiKey]]
	default:
		return m.FlagSets[defaultFlagSetName]
	}
}

func prepareDefaultFlagSet(proxyConf *config.Config) config.FlagSet {
	return config.FlagSet{
		Name: "default",
		CommonFlagSet: config.CommonFlagSet{
			Retrievers:                      proxyConf.Retrievers,
			Notifiers:                       proxyConf.Notifiers,
			Exporters:                       proxyConf.Exporters,
			FileFormat:                      proxyConf.FileFormat,
			PollingInterval:                 proxyConf.PollingInterval,
			StartWithRetrieverError:         proxyConf.StartWithRetrieverError,
			EnablePollingJitter:             proxyConf.EnablePollingJitter,
			DisableNotifierOnInit:           proxyConf.DisableNotifierOnInit,
			EvaluationContextEnrichment:     proxyConf.EvaluationContextEnrichment,
			PersistentFlagConfigurationFile: proxyConf.PersistentFlagConfigurationFile,
		},
	}
}

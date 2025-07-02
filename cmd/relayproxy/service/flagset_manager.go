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

type FlagsetManager struct {
	FlagSets         map[string]*ffclient.GoFeatureFlag
	APIKeysToFlagSet map[string]string
	config           *config.Config
}

func NewFlagsetManager(config *config.Config, logger *zap.Logger) (FlagsetManager, error) {
	if config == nil {
		return FlagsetManager{}, fmt.Errorf("configuration is nil")
	}

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

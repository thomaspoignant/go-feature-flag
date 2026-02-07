package config

import (
	"fmt"

	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/retrieverconf"
)

// FlagSet is the configuration for a flag set.
// A flag set is a collection of flags that are used to evaluate features.
// It is used to group flags together and to apply the same configuration to them.
// It is also used to apply the same API key to all the flags in the flag set.
type FlagSet struct {
	CommonFlagSet `mapstructure:",inline" koanf:",squash"`
	// APIKeys is the api keys for the flag set.
	// This will add a new API keys to the list of authorizedKeys.evaluation keys.
	// This property is madatory for every flagset, we will use it to filter the flag available.
	APIKeys []string `mapstructure:"apiKeys" koanf:"apikeys"`

	// Name of the flagset.
	// This allow to identify the flagset.
	// Default: generated value
	Name string `mapstructure:"name,omitempty" koanf:"name"`
}

type CommonFlagSet struct {
	// Retriever is the configuration on how to retrieve the file
	Retriever *retrieverconf.RetrieverConf `mapstructure:"retriever" koanf:"retriever"`

	// Retrievers is the exact same things than Retriever but allows to give more than 1 retriever at the time.
	// We are dealing with config files in order, if you have the same flag name in multiple files it will be override
	// based of the order of the retrievers in the slice.
	//
	// Note: If both Retriever and Retrievers are set, we will start by calling the Retriever and,
	// after we will use the order of Retrievers.
	Retrievers *[]retrieverconf.RetrieverConf `mapstructure:"retrievers" koanf:"retrievers"`

	// Notifiers is the configuration on where to notify a flag change
	Notifiers []NotifierConf `mapstructure:"notifiers" koanf:"notifiers"`

	// FixNotifiers, Before version v1.46.0, the notifier was called "notifier" instead of "notifiers".
	// This is a backward compatibility fix to allow to use the old configuration.
	// This will be removed in a future version.
	//
	// Deprecated: use Notifiers instead.
	FixNotifiers []NotifierConf `mapstructure:"notifier" koanf:"notifier"`

	// Exporter is the configuration on how to export data
	Exporter *ExporterConf `mapstructure:"exporter" koanf:"exporter"`

	// Exporters is the exact same things than Exporter but allows to give more than 1 exporter at the time.
	Exporters *[]ExporterConf `mapstructure:"exporters" koanf:"exporters"`

	// FileFormat (optional) is the format of the file to retrieve (available YAML, TOML and JSON)
	// Default: YAML
	FileFormat string `mapstructure:"fileFormat" koanf:"fileformat"`

	// PollingInterval (optional) Poll every X time
	// The minimum possible is 1 second
	// Default: 60 seconds
	PollingInterval int `mapstructure:"pollingInterval" koanf:"pollinginterval"`

	// StartWithRetrieverError (optional) If true, the relay proxy will start even if we did not get any flags from
	// the retriever. It will serve only default values until the retriever returns the flags.
	// The init method will not return any error if the flag file is unreachable.
	// Default: false
	StartWithRetrieverError bool `mapstructure:"startWithRetrieverError" koanf:"startwithretrievererror"`

	// EnablePollingJitter (optional) set to true if you want to avoid having true periodicity when
	// retrieving your flags. It is useful to avoid having spike on your flag configuration storage
	// in case your application is starting multiple instance at the same time.
	// We ensure a deviation that is maximum + or - 10% of your polling interval.
	// Default: false
	EnablePollingJitter bool `mapstructure:"enablePollingJitter" koanf:"enablepollingjitter"`

	// DisableNotifierOnInit (optional) set to true if you do not want to call any notifier
	// when the flags are loaded.
	// This is useful if you do not want a Slack/Webhook notification saying that
	// the flags have been added every time you start the application.
	// Default is set to false for backward compatibility.
	// Default: false
	DisableNotifierOnInit bool `mapstructure:"disableNotifierOnInit" koanf:"disablenotifieroninit"`

	// EvaluationContextEnrichment is the flag to enable the evaluation context enrichment.
	EvaluationContextEnrichment map[string]any `mapstructure:"evaluationContextEnrichment" koanf:"evaluationcontextenrichment"` //nolint: lll

	// PersistentFlagConfigurationFile is the flag to enable the persistent flag configuration file.
	PersistentFlagConfigurationFile string `mapstructure:"persistentFlagConfigurationFile" koanf:"persistentflagconfigurationfile"` //nolint: lll

	// Environment is the environment of the flag set.
	Environment string `mapstructure:"environment" koanf:"environment"`
}

func (c *Config) SetFlagSetAPIKeys(flagsetName string, apiKeys []string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	index, err := c.getFlagSetIndexFromName(flagsetName)
	if err != nil {
		return err
	}
	c.FlagSets[index].APIKeys = apiKeys
	return nil
}

func (c *Config) GetFlagSetAPIKeys(flagsetName string) ([]string, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	index, err := c.getFlagSetIndexFromName(flagsetName)
	if err != nil {
		return nil, err
	}
	return c.FlagSets[index].APIKeys, nil
}

// GetFlagSets returns a copy of all flagsets in the config.
// This method is thread-safe.
func (c *Config) GetFlagSets() []FlagSet {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	// Return a copy to avoid external modifications
	result := make([]FlagSet, len(c.FlagSets))
	copy(result, c.FlagSets)
	return result
}

// getFlagSetIndexFromName returns the index of the flagset in the FlagSets array.
// If the flagset is not found, it returns -1.
// This function is not thread safe, it is expected to be called with the mutex locked.
func (c *Config) getFlagSetIndexFromName(flagsetName string) (int, error) {
	for index, flagset := range c.FlagSets {
		if flagset.Name == flagsetName {
			return index, nil
		}
	}
	return -1, fmt.Errorf("flagset %s not found", flagsetName)
}

// AddFlagSet adds a new flagset to the config.
// This method is thread-safe and should be used when dynamically adding flagsets.
func (c *Config) AddFlagSet(flagset FlagSet) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Check if flagset already exists
	if _, err := c.getFlagSetIndexFromName(flagset.Name); err == nil {
		return fmt.Errorf("flagset %s already exists", flagset.Name)
	}

	c.FlagSets = append(c.FlagSets, flagset)
	return nil
}

// RemoveFlagSet removes a flagset from the config by name.
// This method is thread-safe and should be used when dynamically removing flagsets.
func (c *Config) RemoveFlagSet(flagsetName string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	index, err := c.getFlagSetIndexFromName(flagsetName)
	if err != nil {
		return err
	}

	// Remove the flagset by creating a new slice without it
	c.FlagSets = append(c.FlagSets[:index], c.FlagSets[index+1:]...)
	return nil
}

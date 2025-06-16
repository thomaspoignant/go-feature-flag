package config

// FlagSet is the configuration for a flag set.
// A flag set is a collection of flags that are used to evaluate features.
// It is used to group flags together and to apply the same configuration to them.
// It is also used to apply the same API key to all the flags in the flag set.
type FlagSet struct {
	// Name is the name of the flag set.
	Name string `mapstructure:"name,omitempty" koanf:"name"`

	// ApiKey is the api key for the flag set.
	// This will add a new API key to the list of authorizedKeys.evaluation keys.
	ApiKey string `mapstructure:"apiKey,omitempty" koanf:"apikey"`

	// Retrievers is the list of retrievers for the flag set.
	Retrievers *[]RetrieverConf `mapstructure:"retrievers" koanf:"retrievers"`

	// Notifiers is the list of notifiers for the flag set.
	Notifiers *[]NotifierConf `mapstructure:"notifiers" koanf:"notifiers"`

	// Exporters is the list of exporters for the flag set.
	Exporters *[]ExporterConf `mapstructure:"exporters" koanf:"exporters"`

	// FileFormat is the format of the file to use for the flag set.
	FileFormat string `mapstructure:"fileFormat" koanf:"fileformat"`

	// PollingInterval is the interval in milliseconds to poll the flag set.
	PollingInterval int `mapstructure:"pollingInterval" koanf:"pollinginterval"`

	// StartWithRetrieverError is the flag to start with the retriever error.
	StartWithRetrieverError bool `mapstructure:"startWithRetrieverError" koanf:"startwithretrievererror"`

	// EnablePollingJitter is the flag to enable the polling jitter.
	EnablePollingJitter bool `mapstructure:"enablePollingJitter" koanf:"enablepollingjitter"`

	// DisableNotifierOnInit is the flag to disable the notifier on init.
	DisableNotifierOnInit bool `mapstructure:"disableNotifierOnInit" koanf:"disablenotifieroninit"`

	// EvaluationContextEnrichment is the flag to enable the evaluation context enrichment.
	EvaluationContextEnrichment map[string]interface{} `mapstructure:"evaluationContextEnrichment" koanf:"evaluationcontextenrichment"`

	// PersistentFlagConfigurationFile is the flag to enable the persistent flag configuration file.
	PersistentFlagConfigurationFile string `mapstructure:"persistentFlagConfigurationFile" koanf:"persistentflagconfigurationfile"`
}

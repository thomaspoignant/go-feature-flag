package config

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var DefaultRetriever = struct {
	Timeout      time.Duration
	HTTPMethod   string
	GithubBranch string
}{
	Timeout:      10 * time.Second,
	HTTPMethod:   http.MethodGet,
	GithubBranch: "main",
}

var DefaultExporter = struct {
	Format           string
	LogFormat        string
	FileName         string
	CsvFormat        string
	FlushInterval    time.Duration
	MaxEventInMemory int64
}{
	Format:    "JSON",
	LogFormat: "[{{ .FormattedDate}}] user=\"{{ .UserKey}}\", flag=\"{{ .Key}}\", value=\"{{ .Value}}\"",
	FileName:  "flag-variation-{{ .Hostname}}-{{ .Timestamp}}.{{ .Format}}",
	CsvFormat: "{ .Kind}};{{ .ContextKind}};{{ .UserKey}};{{ .CreationDate}};{{ .Key}};{{ .Variation}};" +
		"{{ .Value}};{{ .Default}}\\n",
	FlushInterval:    60000 * time.Millisecond,
	MaxEventInMemory: 100000,
}

// ParseConfig is reading the configuration file
func ParseConfig(log *zap.Logger, version string) (*Config, error) {
	viper.Set("version", version)

	errBindFlag := viper.BindPFlags(pflag.CommandLine)
	if errBindFlag != nil {
		log.Fatal("impossible to parse flag command line", zap.Error(errBindFlag))
	}

	// Read config file
	configFile := viper.GetString("config")
	if configFile != "" {
		log.Info("reading config from file", zap.String("fileLocation", configFile))
		viper.SetConfigFile(configFile)
	} else {
		log.Info("reading config from default directories")
		viper.SetConfigName("goff-proxy")
		viper.AddConfigPath("./")
		viper.AddConfigPath("/goff/")
		viper.AddConfigPath("/etc/opt/goff/")
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer(`.`, `_`))
	viper.AutomaticEnv()
	setViperDefault()

	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}
	proxyConf := &Config{}
	err = viper.Unmarshal(proxyConf)
	if err != nil {
		return nil, err
	}
	return proxyConf, nil
}

// setViperDefault will set default values for the configuration
func setViperDefault() {
	viper.SetDefault("listen", "1031")
	viper.SetDefault("host", "localhost")
	viper.SetDefault("fileFormat", "yaml")
	viper.SetDefault("pollingInterval", 60000)
	viper.SetDefault("restApiTimeout", 5000)
}

type Config struct {
	// ListenPort (optional) is the port we are using to start the proxy
	ListenPort int `mapstructure:"listen"`

	// HideBanner (optional) if true, we don't display the go-feature-flag relay proxy banner
	HideBanner bool `mapstructure:"hideBanner"`

	// EnableSwagger (optional) to have access to the swagger
	EnableSwagger bool `mapstructure:"enableSwagger"`

	// Host should be set if you are using swagger (default is localhost)
	Host string `mapstructure:"host"`

	// Debug (optional) if true, go-feature-flag relay proxy will run on debug mode, with more logs and custom responses
	Debug bool `mapstructure:"debug"`

	// PollingInterval (optional) Poll every X time
	// The minimum possible is 1 second
	// Default: 60 seconds
	PollingInterval int `mapstructure:"pollingInterval"`

	// FileFormat (optional) is the format of the file to retrieve (available YAML, TOML and JSON)
	// Default: YAML
	FileFormat string `mapstructure:"fileFormat"`

	// StartWithRetrieverError (optional) If true, the relay proxy will start even if we did not get any flags from
	// the retriever. It will serve only default values until the retriever returns the flags.
	// The init method will not return any error if the flag file is unreachable.
	// Default: false
	StartWithRetrieverError bool `mapstructure:"startWithRetrieverError"`

	// Retriever is the configuration on how to retrieve the file
	Retriever *RetrieverConf `mapstructure:"retriever"`

	// Retrievers is the exact same things than Retriever but allows to give more than 1 retriever at the time.
	// We are dealing with config files in order, if you have the same flag name in multiple files it will be override
	// based of the order of the retrievers in the slice.
	//
	// Note: If both Retriever and Retrievers are set, we will start by calling the Retriever and,
	// after we will use the order of Retrievers.
	Retrievers *[]RetrieverConf `mapstructure:"retrievers"`

	// Exporter is the configuration on how to export data
	Exporter *ExporterConf `mapstructure:"exporter"`

	// Notifiers is the configuration on where to notify a flag change
	Notifiers []NotifierConf `mapstructure:"notifier"`

	// RestAPITimeout is the timeout on the API.
	RestAPITimeout int `mapstructure:"restApiTimeout"`

	// Version is the version of the relay-proxy
	Version string

	// APIKeys list of API keys that authorized to use endpoints
	APIKeys []string `mapstructure:"apiKeys"`

	// ---- private fields

	// apiKeySet is the internal representation of the list of api keys configured
	// we store them in a set to be
	apiKeysSet map[string]interface{}
}

// APIKeyExists is checking if an API Key exist in the relay proxy configuration
func (c *Config) APIKeyExists(apiKey string) bool {
	if c.apiKeysSet == nil {
		apiKeySet := make(map[string]interface{})
		for _, currentAPIKey := range c.APIKeys {
			apiKeySet[currentAPIKey] = new(interface{})
		}
		c.apiKeysSet = apiKeySet
	}

	_, ok := c.apiKeysSet[apiKey]
	return ok
}

// IsValid contains all the validation of the configuration.
func (c *Config) IsValid() error {
	if c.ListenPort == 0 {
		return fmt.Errorf("invalid port %d", c.ListenPort)
	}

	if c.Retriever == nil && c.Retrievers == nil {
		return fmt.Errorf("no retriever available in the configuration")
	}

	if c.Retriever != nil {
		if err := c.Retriever.IsValid(); err != nil {
			return err
		}
	}

	if c.Retrievers != nil {
		for _, retriever := range *c.Retrievers {
			if err := retriever.IsValid(); err != nil {
				return err
			}
		}
	}

	// Exporter is optional
	if c.Exporter != nil {
		if err := c.Exporter.IsValid(); err != nil {
			return err
		}
	}

	if c.Notifiers != nil {
		for _, notif := range c.Notifiers {
			if err := notif.IsValid(); err != nil {
				return err
			}
		}
	}

	return nil
}

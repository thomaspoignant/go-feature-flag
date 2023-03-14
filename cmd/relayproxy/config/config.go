package config

import (
	"fmt"
	"github.com/spf13/pflag"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

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

	// retriever
	viper.SetDefault("retriever.timeout", int64(10*time.Second/time.Millisecond))
	viper.SetDefault("retriever.method", http.MethodGet)
	viper.SetDefault("retriever.body", "")

	// exporter
	viper.SetDefault("exporter.format", "JSON")
	viper.SetDefault("exporter.logFormat",
		"[{{ .FormattedDate}}] user=\"{{ .UserKey}}\", flag=\"{{ .Key}}\", value=\"{{ .Value}}\"")
	viper.SetDefault("exporter.filename", "flag-variation-{{ .Hostname}}-{{ .Timestamp}}.{{ .Format}}")
	viper.SetDefault("exporter.csvTemplate",
		"{ .Kind}};{{ .ContextKind}};{{ .UserKey}};{{ .CreationDate}};{{ .Key}};{{ .Variation}};{{ .Value}};{{ .Default}}\\n")
	viper.SetDefault("exporter.flushInterval", 60000)
	viper.SetDefault("exporter.maxEventInMemory", 100000)
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
}

// IsValid contains all the validation of the configuration.
func (c *Config) IsValid() error {
	if c.ListenPort == 0 {
		return fmt.Errorf("invalid port %d", c.ListenPort)
	}

	if err := c.Retriever.IsValid(); err != nil {
		return err
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

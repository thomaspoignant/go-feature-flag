package config

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/pflag"
	"github.com/xitongsys/parquet-go/parquet"
	"go.uber.org/zap"
)

var k = koanf.New(".")
var DefaultRetriever = struct {
	Timeout    time.Duration
	HTTPMethod string
	GitBranch  string
}{
	Timeout:    10 * time.Second,
	HTTPMethod: http.MethodGet,
	GitBranch:  "main",
}

var DefaultExporter = struct {
	Format                  string
	LogFormat               string
	FileName                string
	CsvFormat               string
	FlushInterval           time.Duration
	MaxEventInMemory        int64
	ParquetCompressionCodec string
}{
	Format:    "JSON",
	LogFormat: "[{{ .FormattedDate}}] user=\"{{ .UserKey}}\", flag=\"{{ .Key}}\", value=\"{{ .Value}}\"",
	FileName:  "flag-variation-{{ .Hostname}}-{{ .Timestamp}}.{{ .Format}}",
	CsvFormat: "{{ .Kind}};{{ .ContextKind}};{{ .UserKey}};{{ .CreationDate}};{{ .Key}};{{ .Variation}};" +
		"{{ .Value}};{{ .Default}}\n",
	FlushInterval:           60000 * time.Millisecond,
	MaxEventInMemory:        100000,
	ParquetCompressionCodec: parquet.CompressionCodec_SNAPPY.String(),
}

// New is reading the configuration file
func New(flagSet *pflag.FlagSet, log *zap.Logger, version string) (*Config, error) {
	k.Delete("")

	// Default values
	_ = k.Load(confmap.Provider(map[string]interface{}{
		"listen":          "1031",
		"host":            "localhost",
		"fileFormat":      "yaml",
		"restApiTimeout":  5000,
		"pollingInterval": 60000,
	}, "."), nil)

	// mapping command line parameters to koanf
	if errBindFlag := k.Load(posflag.Provider(flagSet, ".", k), nil); errBindFlag != nil {
		log.Fatal("impossible to parse flag command line", zap.Error(errBindFlag))
	}

	// Read config file
	configFileLocation, errFileLocation := locateConfigFile(k.String("config"))
	if errFileLocation != nil {
		log.Info("not using any configuration file", zap.Error(errFileLocation))
	} else {
		ext := filepath.Ext(configFileLocation)
		var parser koanf.Parser
		switch strings.ToLower(ext) {
		case ".toml":
			parser = toml.Parser()
			break
		case ".json":
			parser = json.Parser()
			break
		default:
			parser = yaml.Parser()
		}

		if errBindFile := k.Load(file.Provider(configFileLocation), parser); errBindFile != nil {
			log.Error("error loading file", zap.Error(errBindFile))
		}
	}

	// Map environment variables
	_ = k.Load(env.Provider("", ".", func(s string) string {
		return strings.ReplaceAll(strings.ToLower(s), "_", ".")
	}), nil)

	_ = k.Set("version", version)

	proxyConf := &Config{}
	errUnmarshal := k.Unmarshal("", &proxyConf)
	if errUnmarshal != nil {
		return nil, errUnmarshal
	}
	return proxyConf, nil
}

type Config struct {
	// ListenPort (optional) is the port we are using to start the proxy
	ListenPort int `mapstructure:"listen" koanf:"listen"`

	// HideBanner (optional) if true, we don't display the go-feature-flag relay proxy banner
	HideBanner bool `mapstructure:"hideBanner" koanf:"hidebanner"`

	// EnableSwagger (optional) to have access to the swagger
	EnableSwagger bool `mapstructure:"enableSwagger" koanf:"enableswagger"`

	// Host should be set if you are using swagger (default is localhost)
	Host string `mapstructure:"host" koanf:"host"`

	// Debug (optional) if true, go-feature-flag relay proxy will run on debug mode, with more logs and custom responses
	Debug bool `mapstructure:"debug" koanf:"debug"`

	// PollingInterval (optional) Poll every X time
	// The minimum possible is 1 second
	// Default: 60 seconds
	PollingInterval int `koanf:"pollinginterval" mapstructure:"pollingInterval"`

	// FileFormat (optional) is the format of the file to retrieve (available YAML, TOML and JSON)
	// Default: YAML
	FileFormat string `mapstructure:"fileFormat" koanf:"fileformat"`

	// StartWithRetrieverError (optional) If true, the relay proxy will start even if we did not get any flags from
	// the retriever. It will serve only default values until the retriever returns the flags.
	// The init method will not return any error if the flag file is unreachable.
	// Default: false
	StartWithRetrieverError bool `mapstructure:"startWithRetrieverError" koanf:"startwithretrievererror"`

	// Retriever is the configuration on how to retrieve the file
	Retriever *RetrieverConf `mapstructure:"retriever" koanf:"retriever"`

	// Retrievers is the exact same things than Retriever but allows to give more than 1 retriever at the time.
	// We are dealing with config files in order, if you have the same flag name in multiple files it will be override
	// based of the order of the retrievers in the slice.
	//
	// Note: If both Retriever and Retrievers are set, we will start by calling the Retriever and,
	// after we will use the order of Retrievers.
	Retrievers *[]RetrieverConf `mapstructure:"retrievers" koanf:"retrievers"`

	// Exporter is the configuration on how to export data
	Exporter *ExporterConf `mapstructure:"exporter" koanf:"exporter"`

	// Notifiers is the configuration on where to notify a flag change
	Notifiers []NotifierConf `mapstructure:"notifier" koanf:"notifier"`

	// RestAPITimeout is the timeout on the API.
	RestAPITimeout int `mapstructure:"restApiTimeout" koanf:"restapitimeout"`

	// Version is the version of the relay-proxy
	Version string `mapstructure:"version" koanf:"version"`

	// APIKeys list of API keys that authorized to use endpoints
	APIKeys []string `mapstructure:"apiKeys" koanf:"apikeys"`

	// StartAsAwsLambda (optional) if true the relay proxy will start ready to be launch as AWS Lambda
	StartAsAwsLambda bool `mapstructure:"startAsAwsLambda" koanf:"startasawslambda"`

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

// locateConfigFile is selecting the configuration file we will use.
func locateConfigFile(inputFilePath string) (string, error) {
	filename := "goff-proxy"
	defaultLocations := []string{
		"./",
		"/goff/",
		"/etc/opt/goff/",
	}
	supportedExtensions := []string{
		"yaml",
		"toml",
		"json",
		"yml",
	}

	if inputFilePath != "" {
		if _, err := os.Stat(inputFilePath); err != nil {
			return "", fmt.Errorf("impossible to find config file %s", inputFilePath)
		}
		return inputFilePath, nil
	}
	for _, location := range defaultLocations {
		for _, ext := range supportedExtensions {
			configFile := fmt.Sprintf("%s%s.%s", location, filename, ext)
			if _, err := os.Stat(configFile); err == nil {
				return configFile, nil
			}
		}
	}
	return "", fmt.Errorf(
		"impossible to find config file in the default locations [%s]", strings.Join(defaultLocations, ","))
}

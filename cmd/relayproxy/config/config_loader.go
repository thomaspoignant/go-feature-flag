package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/pflag"
	"github.com/thomaspoignant/go-feature-flag/utils"
	"go.uber.org/zap"
)

type ConfigLoader struct {
	// input parameters
	cmdLineFlagSet *pflag.FlagSet
	log            *zap.Logger
	version        string
	watchChanges   bool

	// internal state
	k            *koanf.Koanf
	fileProvider *file.File

	// callbacks to be called when the configuration changes
	callbacks []func(newConfig *Config)
}

// NewConfigLoader creates a new ConfigLoader.
func NewConfigLoader(
	cmdLineFlagSet *pflag.FlagSet, log *zap.Logger, version string, watchChanges bool) *ConfigLoader {
	configLoader := ConfigLoader{
		cmdLineFlagSet: cmdLineFlagSet,
		log:            log,
		version:        version,
		watchChanges:   watchChanges,
		k:              koanf.New("."),
	}

	// load the configuration from the command line, the configuration file, the environment variables and the version
	configLoader.loadConfig()

	// start watching for changes
	configLoader.startWatchChanges()
	return &configLoader
}

// loadConfig loads the configuration from
// - the command line
// - the configuration file
// - the environment variables
func (c *ConfigLoader) loadConfig() {
	_ = c.k.Load(confmap.Provider(map[string]any{
		"fileFormat":      "yaml",
		"pollingInterval": 60000,
		"logLevel":        DefaultLogLevel,
	}, "."), nil)

	c.loadPosflag(c.cmdLineFlagSet)
	c.loadConfigFile()
	c.loadEnvVariables()
	_ = c.k.Set("version", c.version)
}

// AddConfigChangeCallback adds a callback to be called when the configuration changes
func (c *ConfigLoader) AddConfigChangeCallback(callback func(newConfig *Config)) {
	c.callbacks = append(c.callbacks, callback)
}

func (c *ConfigLoader) startWatchChanges() {
	if c.fileProvider == nil || !c.watchChanges {
		return
	}
	_ = c.fileProvider.Watch(func(event any, err error) {
		if err != nil {
			c.log.Error("error watching for configuration changes (error from file provider)", zap.Error(err))
			return
		}

		newConfig := &ConfigLoader{
			cmdLineFlagSet: c.cmdLineFlagSet,
			log:            c.log,
			version:        c.version,
			watchChanges:   false,
			k:              koanf.New("."),
		}
		newConfig.loadConfig()          // load the new configuration
		c2, err := newConfig.ToConfig() // unmarshal the new configuration
		if err != nil {
			c.log.Error("error loading new config", zap.Error(err))
			return
		}

		for _, callback := range c.callbacks {
			callback(c2)
		}
	})
}

// ToConfig returns the configuration.
func (c *ConfigLoader) ToConfig() (*Config, error) {
	proxyConf := &Config{}
	errUnmarshal := c.k.Unmarshal("", proxyConf)
	if errUnmarshal != nil {
		return nil, errUnmarshal
	}
	processExporters(proxyConf)
	return proxyConf, nil
}

// StopWatchChanges stops watching for changes in the configuration file
func (c *ConfigLoader) StopWatchChanges() error {
	if c.fileProvider == nil || !c.watchChanges {
		return nil
	}
	return c.fileProvider.Unwatch()
}

// loadPosflag loads the flags from the command line
func (c *ConfigLoader) loadPosflag(cmdLineFlagSet *pflag.FlagSet) {
	posflagProvider := posflag.Provider(cmdLineFlagSet, ".", c.k)
	errBindFlag := c.k.Load(posflagProvider, nil)
	if errBindFlag != nil {
		c.log.Fatal("impossible to parse flag command line", zap.Error(errBindFlag))
	}
}

// loadConfigFile loads the configuration file
func (c *ConfigLoader) loadConfigFile() {
	configFileLocation, errFileLocation := locateConfigFile(c.k.String("config"))
	if errFileLocation != nil {
		c.log.Info("not using any configuration file", zap.Error(errFileLocation))
		return
	}

	parser := selectParserForFile(configFileLocation)
	c.fileProvider = file.Provider(configFileLocation)
	if errBindFile := c.k.Load(c.fileProvider, parser); errBindFile != nil {
		c.log.Error("error loading file", zap.Error(errBindFile))
	}
}

// loadEnvVariables loads the environment variables
func (c *ConfigLoader) loadEnvVariables() {
	_ = c.k.Load(c.mapEnvVariablesProvider(c.k.String("envVariablePrefix"), c.log), nil)
}

// locateConfigFile is selecting the configuration file we will use.
func locateConfigFile(inputFilePath string) (string, error) {
	filename := "goff-proxy"
	defaultLocations := []string{"./", "/goff/", "/etc/opt/goff/"}
	supportedExtensions := []string{"yaml", "toml", "json", "yml"}

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
		"impossible to find config file in the default locations [%s]",
		strings.Join(defaultLocations, ","),
	)
}

// selectParserForFile returns the appropriate parser based on file extension
func selectParserForFile(configFileLocation string) koanf.Parser {
	ext := filepath.Ext(configFileLocation)
	switch strings.ToLower(ext) {
	case ".toml":
		return toml.Parser()
	case ".json":
		return json.Parser()
	default:
		return yaml.Parser()
	}
}

// processExporters handles the post-processing of exporters configuration
func processExporters(proxyConf *Config) {
	if proxyConf.Exporters == nil {
		return
	}

	for i := range *proxyConf.Exporters {
		addresses := (*proxyConf.Exporters)[i].Kafka.Addresses
		if len(addresses) == 0 || (len(addresses) == 1 && strings.Contains(addresses[0], ",")) {
			(*proxyConf.Exporters)[i].Kafka.Addresses = utils.StringToArray(addresses)
		}
	}
}

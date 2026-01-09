package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

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
	k              *koanf.Koanf
	fileProvider   *file.File
	configFilePath string // stored config file path for reload

	// callbacks to be called when the configuration changes
	callbacks []func(newConfig *Config)

	// callbacksMutex is used to protect the callbacks slice
	callbacksMutex sync.RWMutex

	// channel-based event processing for file watcher
	// Using a buffered channel of size 1 coalesces multiple rapid events into one
	eventChan chan struct{} // receives file change events
	stopChan  chan struct{} // signals the event processor goroutine to stop
	stopOnce  sync.Once     // ensures stopChan is only closed once
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

	// Initialize channels only if we're watching for changes
	if watchChanges {
		configLoader.eventChan = make(chan struct{}, 1) // buffered to avoid blocking
		configLoader.stopChan = make(chan struct{})
	}

	// load the configuration from the command line, the configuration file, the environment variables and the version
	configLoader.loadConfig()

	// start watching for changes
	if watchChanges {
		configLoader.startWatchChanges()
	}
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
	c.callbacksMutex.Lock()
	defer c.callbacksMutex.Unlock()
	c.callbacks = append(c.callbacks, callback)
}

// startWatchChanges starts watching for changes in the configuration file
func (c *ConfigLoader) startWatchChanges() {
	if c.fileProvider == nil || !c.watchChanges {
		return
	}

	// Start the event processor goroutine
	// This approach is used to avoid blocking the fsnotify goroutine
	// The event processor goroutine is responsible for processing the events and reloading the configuration
	// Check https://github.com/knadh/koanf/issues/12#issuecomment-2637665148 to understand why this is necessary
	go c.processConfigChangeEvents()

	// The file watcher callback just sends to the channel and returns immediately
	// This prevents blocking the fsnotify goroutine
	errWatch := c.fileProvider.Watch(func(event any, err error) {
		if err != nil {
			c.log.Error("error watching for configuration changes (error from file provider)", zap.Error(err))
			return
		}

		// Non-blocking send to the event channel
		// If the channel is full (event already pending), we skip this event
		// since the pending event will trigger a config reload anyway
		select {
		case c.eventChan <- struct{}{}:
			// Event sent successfully
		default:
			// Channel full, event already pending - skip this one
		}
	})
	if errWatch != nil {
		c.log.Error("error watching for configuration changes (error from file provider)", zap.Error(errWatch))
	}
}

// processConfigChangeEvents processes file change events from the channel.
// The buffered channel (size 1) naturally coalesces rapid events - if an event
// is already pending, new events are dropped since they would trigger the same reload.
func (c *ConfigLoader) processConfigChangeEvents() {
	for {
		select {
		case <-c.stopChan:
			return
		case <-c.eventChan:
			c.reloadConfigAndNotify()
		}
	}
}

// reloadConfigAndNotify reloads the configuration and notifies all callbacks
func (c *ConfigLoader) reloadConfigAndNotify() {
	// Create a new ConfigLoader with the stored config file path
	newConfig := &ConfigLoader{
		cmdLineFlagSet: c.cmdLineFlagSet,
		log:            c.log,
		version:        c.version,
		watchChanges:   false,
		k:              koanf.New("."),
		configFilePath: c.configFilePath,
	}
	newConfig.loadConfig()
	modifiedConfig, err := newConfig.ToConfig()
	if err != nil {
		c.log.Error("error loading new config", zap.Error(err))
		return
	}

	c.callbacksMutex.RLock()
	defer c.callbacksMutex.RUnlock()
	for _, callback := range c.callbacks {
		callback(modifiedConfig)
	}
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

	// Signal the event processor goroutine to stop (only once)
	if c.stopChan != nil {
		c.stopOnce.Do(func() {
			close(c.stopChan)
		})
	}

	return c.fileProvider.Unwatch()
}

// loadPosflag loads the flags from the command line
func (c *ConfigLoader) loadPosflag(cmdLineFlagSet *pflag.FlagSet) {
	posflagProvider := posflag.Provider(cmdLineFlagSet, ".", c.k)
	errBindFlag := c.k.Load(posflagProvider, nil)
	if errBindFlag != nil {
		c.log.Error("impossible to parse flag command line", zap.Error(errBindFlag))
	}
}

// loadConfigFile loads the configuration file
func (c *ConfigLoader) loadConfigFile() {
	var errFileLocation error

	// Use stored config file path if available (for reload), otherwise read from flag set
	if c.configFilePath == "" {
		c.configFilePath, errFileLocation = locateConfigFile(c.k.String("config"))
		if errFileLocation != nil {
			c.log.Info("not using any configuration file", zap.Error(errFileLocation))
			return
		}
	}

	parser := selectParserForFile(c.configFilePath)
	c.fileProvider = file.Provider(c.configFilePath)
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

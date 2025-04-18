package config

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/xitongsys/parquet-go/parquet"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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

const DefaultLogLevel = "info"

var DefaultExporter = struct {
	Format                  string
	LogFormat               string
	FileName                string
	CsvFormat               string
	FlushInterval           time.Duration
	MaxEventInMemory        int64
	ParquetCompressionCodec string
	LogLevel                string
	ExporterEventType       ffclient.ExporterEventType
}{
	Format:    "JSON",
	LogFormat: "[{{ .FormattedDate}}] user=\"{{ .UserKey}}\", flag=\"{{ .Key}}\", value=\"{{ .Value}}\"",
	FileName:  "flag-variation-{{ .Hostname}}-{{ .Timestamp}}.{{ .Format}}",
	CsvFormat: "{{ .Kind}};{{ .ContextKind}};{{ .UserKey}};{{ .CreationDate}};{{ .Key}};{{ .Variation}};" +
		"{{ .Value}};{{ .Default}};{{ .Source}}\n",
	FlushInterval:           60000 * time.Millisecond,
	MaxEventInMemory:        100000,
	ParquetCompressionCodec: parquet.CompressionCodec_SNAPPY.String(),
	LogLevel:                DefaultLogLevel,
	ExporterEventType:       ffclient.FeatureEventExporter,
}

// New is reading the configuration file
func New(flagSet *pflag.FlagSet, log *zap.Logger, version string) (*Config, error) {
	k.Delete("")

	// Default values
	_ = k.Load(confmap.Provider(map[string]interface{}{
		"listen":          "1031",
		"host":            "localhost",
		"fileFormat":      "yaml",
		"pollingInterval": 60000,
		"logLevel":        DefaultLogLevel,
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
		case ".json":
			parser = json.Parser()
		default:
			parser = yaml.Parser()
		}

		if errBindFile := k.Load(file.Provider(configFileLocation), parser); errBindFile != nil {
			log.Error("error loading file", zap.Error(errBindFile))
		}
	}

	// Map environment variables
	_ = k.Load(mapEnvVariablesProvider(k.String("envVariablePrefix"), log), nil)
	_ = k.Set("version", version)

	proxyConf := &Config{}
	errUnmarshal := k.Unmarshal("", &proxyConf)
	if errUnmarshal != nil {
		return nil, errUnmarshal
	}

	if proxyConf.Exporters != nil {
		for i := range *proxyConf.Exporters {
			(*proxyConf.Exporters)[i].Kafka.Addresses = stringToArray(
				(*proxyConf.Exporters)[i].Kafka.Addresses,
			)
		}
	}

	return proxyConf, nil
}

func mapEnvVariablesProvider(prefix string, log *zap.Logger) koanf.Provider {
	return env.ProviderWithValue(prefix, ".", func(key string, v string) (string, interface{}) {
		key = strings.TrimPrefix(key, prefix)
		if strings.HasPrefix(key, "RETRIEVERS") ||
			strings.HasPrefix(key, "NOTIFIERS") ||
			strings.HasPrefix(key, "EXPORTERS") {
			configMap := k.Raw()
			err := loadArrayEnv(key, v, configMap)
			if err != nil {
				log.Error(
					"config: error loading array env",
					zap.String("key", key),
					zap.String("value", v),
					zap.Error(err),
				)
				return key, v
			}
			return key, v
		}

		if strings.HasPrefix(key, "EXPORTER_KAFKA_ADDRESSES") {
			return "exporter.kafka.addresses", strings.Split(v, ",")
		}

		if strings.HasPrefix(key, "AUTHORIZEDKEYS_EVALUATION") {
			return "authorizedKeys.evaluation", strings.Split(v, ",")
		}
		if strings.HasPrefix(key, "AUTHORIZEDKEYS_ADMIN") {
			return "authorizedKeys.admin", strings.Split(v, ",")
		}

		if key == "OTEL_RESOURCE_ATTRIBUTES" {
			parseOtelResourceAttributes(v, log)
			return key, v
		}

		return strings.ReplaceAll(strings.ToLower(key), "_", "."), v
	})
}

func stringToArray(item []string) []string {
	if len(item) > 0 {
		return strings.Split(item[0], ",")
	}
	return item
}

func parseOtelResourceAttributes(attributes string, log *zap.Logger) {
	configMap := k.Raw()
	otel, ok := configMap["otel"].(map[string]interface{})
	if !ok {
		configMap["otel"] = make(map[string]interface{})
		otel = configMap["otel"].(map[string]interface{})
	}

	resource, ok := otel["resource"].(map[string]interface{})
	if !ok {
		otel["resource"] = make(map[string]interface{})
		resource = otel["resource"].(map[string]interface{})
	}

	attrs, ok := resource["attributes"].(map[string]interface{})
	if !ok {
		resource["attributes"] = make(map[string]interface{})
		attrs = resource["attributes"].(map[string]interface{})
	}

	for _, attr := range strings.Split(attributes, ",") {
		k, v, found := strings.Cut(attr, "=")
		if !found {
			log.Error("config: error loading OTEL_RESOURCE_ATTRIBUTES - incorrect format",
				zap.String("key", k), zap.String("value", v))
			continue
		}

		attrs[k] = v
	}

	_ = k.Set("otel", otel)
}

type Config struct {
	// ListenPort (optional) is the port we are using to start the proxy
	ListenPort int `mapstructure:"listen" koanf:"listen"`

	// HideBanner (optional) if true, we don't display the go-feature-flag relay proxy banner
	HideBanner bool `mapstructure:"hideBanner" koanf:"hidebanner"`

	// EnablePprof (optional) if true, go-feature-flag relay proxy will start
	// the pprof endpoints on the same port as the monitoring.
	// Default: false
	EnablePprof bool `mapstructure:"enablePprof" koanf:"enablepprof"`

	// EnableSwagger (optional) to have access to the swagger
	EnableSwagger bool `mapstructure:"enableSwagger" koanf:"enableswagger"`

	// Host should be set if you are using swagger (default is localhost)
	Host string `mapstructure:"host" koanf:"host"`

	// LogLevel (optional) sets the verbosity for logging,
	// Possible values: debug, info, warn, error, dpanic, panic, fatal
	// If level debug go-feature-flag relay proxy will run on debug mode, with more logs and custom responses
	// Default: debug
	LogLevel string `mapstructure:"logLevel" koanf:"loglevel"`

	// LogFormat (optional) sets the log message format
	// Possible values: json, logfmt
	// Default: json
	LogFormat string `mapstructure:"logFormat" koanf:"logformat"`

	// PollingInterval (optional) Poll every X time
	// The minimum possible is 1 second
	// Default: 60 seconds
	PollingInterval int `mapstructure:"pollingInterval" koanf:"pollinginterval"`

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
	DisableNotifierOnInit bool `mapstructure:"DisableNotifierOnInit" koanf:"DisableNotifierOnInit"`

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

	// Exporters is the exact same things than Exporter but allows to give more than 1 exporter at the time.
	Exporters *[]ExporterConf `mapstructure:"exporters" koanf:"exporters"`

	// ExporterCleanQueueInterval (optional) is the duration between each cleaning of the queue by the thread in charge
	// of removing the old events.
	// Default: 1 minute
	ExporterCleanQueueInterval time.Duration `mapstructure:"exporterCleanQueueInterval" koanf:"exportercleanqueueinterval"`

	// Notifiers is the configuration on where to notify a flag change
	Notifiers []NotifierConf `mapstructure:"notifier" koanf:"notifier"`

	// Version is the version of the relay-proxy
	Version string `mapstructure:"version" koanf:"version"`

	// Disable x-gofeatureflag-version header in the relay-proxy HTTP response
	// Default: false
	DisableVersionHeader bool `mapstructure:"disableVersionHeader" koanf:"disableversionheader"`

	// Deprecated: use AuthorizedKeys instead
	// APIKeys list of API keys that authorized to use endpoints
	APIKeys []string `mapstructure:"apiKeys" koanf:"apikeys"`

	// AuthorizedKeys list of API keys that authorized to use endpoints
	AuthorizedKeys APIKeys `mapstructure:"authorizedKeys" koanf:"authorizedkeys"`

	// StartAsAwsLambda (optional) if true, the relay proxy will start ready to be launched as AWS Lambda
	StartAsAwsLambda bool `mapstructure:"startAsAwsLambda" koanf:"startasawslambda"`

	// AwsLambdaAdapter (optional) is the adapter to use when the relay proxy is started as an AWS Lambda.
	// Possible values are "APIGatewayV1", "APIGatewayV2" and "ALB"
	// Default: "APIGatewayV2"
	AwsLambdaAdapter string `mapstructure:"awsLambdaAdapter" koanf:"awslambdaadapter"`

	// EvaluationContextEnrichment (optional) will be merged with the evaluation context sent during the evaluation.
	// It is useful to add common attributes to all the evaluations, such as a server version, environment, ...
	//
	// All those fields will be included in the custom attributes of the evaluation context,
	// if in the evaluation context you have a field with the same name,
	// it will be overridden by the evaluationContextEnrichment.
	// Default: nil
	EvaluationContextEnrichment map[string]interface{} `mapstructure:"evaluationContextEnrichment" koanf:"evaluationcontextenrichment"` //nolint: lll

	// OpenTelemetryOtlpEndpoint (optional) is the endpoint of the OpenTelemetry collector
	// Default: ""
	OpenTelemetryOtlpEndpoint string `mapstructure:"openTelemetryOtlpEndpoint" koanf:"opentelemetryotlpendpoint"`

	// MonitoringPort (optional) is the port we are using to expose the metrics and healthchecks
	// If not set we will use the same port as the proxy
	MonitoringPort int `mapstructure:"monitoringPort" koanf:"monitoringport"`

	// PersistentFlagConfigurationFile (optional) if set GO Feature Flag will store flags configuration in this file
	//  to be able to serve the flags even if none of the retrievers is available during starting time.
	//
	// By default, the flag configuration is not persisted and stays on the retriever system. By setting a file here,
	// you ensure that GO Feature Flag will always start with a configuration but which can be out-dated.
	PersistentFlagConfigurationFile string `mapstructure:"persistentFlagConfigurationFile" koanf:"persistentflagconfigurationfile"` //nolint: lll

	// OtelConfig is the configuration for the OpenTelemetry part of the relay proxy
	OtelConfig OpenTelemetryConfiguration `mapstructure:"otel" koanf:"otel"`

	// JaegerConfig is the configuration for the Jaeger sampling of the relay proxy
	JaegerConfig JaegerSamplerConfiguration `mapstructure:"jaeger" koanf:"jaeger"`

	// EnvVariablePrefix (optional) is the prefix we are using to load the environment variables
	// By default we have no prefix
	EnvVariablePrefix string `mapstructure:"envVariablePrefix" koanf:"envvariableprefix"`

	// ---- private fields

	// apiKeySet is the internal representation of an API keys list configured
	// we store them in a set to be
	apiKeysSet map[string]interface{}

	// adminAPIKeySet is the internal representation of an admin API keys list configured
	// we store them in a set to be
	adminAPIKeySet map[string]interface{}
}

// OpenTelemetryConfiguration is the configuration for the OpenTelemetry part of the relay proxy
// It is used to configure the OpenTelemetry SDK and the OpenTelemetry Exporter
// Most of the time this configuration is set using environment variables.
type OpenTelemetryConfiguration struct {
	SDK struct {
		Disabled bool `mapstructure:"disabled" koanf:"disabled"`
	} `mapstructure:"sdk"      koanf:"sdk"`
	Exporter OtelExporter `mapstructure:"exporter" koanf:"exporter"`
	Service  struct {
		Name string `mapstructure:"name" koanf:"name"`
	} `mapstructure:"service"  koanf:"service"`
	Traces struct {
		Sampler string `mapstructure:"sampler" koanf:"sampler"`
	} `mapstructure:"traces"   koanf:"traces"`
	Resource OtelResource `mapstructure:"resource" koanf:"resource"`
}

type OtelExporter struct {
	Otlp OtelExporterOtlp `mapstructure:"otlp" koanf:"otlp"`
}

type OtelExporterOtlp struct {
	Endpoint string `mapstructure:"endpoint" koanf:"endpoint"`
	Protocol string `mapstructure:"protocol" koanf:"protocol"`
}

type OtelResource struct {
	Attributes map[string]string `mapstructure:"attributes" koanf:"attributes"`
}

// JaegerSamplerConfiguration is the configuration object to configure the sampling.
// Most of the time this configuration is set using environment variables.
type JaegerSamplerConfiguration struct {
	Sampler struct {
		Manager struct {
			Host struct {
				Port string `mapstructure:"port" koanf:"port"`
			} `mapstructure:"host" koanf:"host"`
		} `mapstructure:"manager" koanf:"manager"`
		Refresh struct {
			Interval string `mapstructure:"interval" koanf:"interval"`
		} `mapstructure:"refresh" koanf:"refresh"`
		Max struct {
			Operations int `mapstructure:"operations" koanf:"operations"`
		} `mapstructure:"max" koanf:"max"`
	} `mapstructure:"sampler" koanf:"sampler"`
}

// APIKeysAdminExists is checking if an admin API Key exist in the relay proxy configuration
func (c *Config) APIKeysAdminExists(apiKey string) bool {
	if c.adminAPIKeySet == nil {
		adminAPIKeySet := make(map[string]interface{})
		for _, currentAPIKey := range c.AuthorizedKeys.Admin {
			adminAPIKeySet[currentAPIKey] = new(interface{})
		}
		c.adminAPIKeySet = adminAPIKeySet
	}

	_, ok := c.adminAPIKeySet[apiKey]
	return ok
}

// APIKeyExists is checking if an API Key exist in the relay proxy configuration
func (c *Config) APIKeyExists(apiKey string) bool {
	if c.APIKeysAdminExists(apiKey) {
		return true
	}
	if c.apiKeysSet == nil {
		apiKeySet := make(map[string]interface{})

		// Remove this part when the APIKeys field is removed
		for _, currentAPIKey := range c.APIKeys {
			apiKeySet[currentAPIKey] = new(interface{})
		}
		// end of remove

		for _, currentAPIKey := range c.AuthorizedKeys.Evaluation {
			apiKeySet[currentAPIKey] = new(interface{})
		}
		c.apiKeysSet = apiKeySet
	}

	_, ok := c.apiKeysSet[apiKey]
	return ok
}

// IsValid contains all the validation of the configuration.
func (c *Config) IsValid() error {
	if c == nil {
		return fmt.Errorf("empty config")
	}
	if c.ListenPort == 0 {
		return fmt.Errorf("invalid port %d", c.ListenPort)
	}
	if c.LogLevel != "" {
		if _, err := zapcore.ParseLevel(c.LogLevel); err != nil {
			return err
		}
	}
	if err := c.validateRetrievers(); err != nil {
		return err
	}

	if err := c.validateExporters(); err != nil {
		return err
	}

	if err := c.validateNotifiers(); err != nil {
		return err
	}

	// log format validation
	switch strings.ToLower(c.LogFormat) {
	case "json", "logfmt", "":
		break
	default:
		return fmt.Errorf("invalid log format %s", c.LogFormat)
	}

	return nil
}

func (c *Config) validateRetrievers() error {
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
	return nil
}

func (c *Config) validateExporters() error {
	if c.Exporter != nil {
		if err := c.Exporter.IsValid(); err != nil {
			return err
		}
	}
	if c.Exporters != nil {
		for _, exporter := range *c.Exporters {
			if err := exporter.IsValid(); err != nil {
				return err
			}
		}
	}
	return nil
}
func (c *Config) validateNotifiers() error {
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
		"impossible to find config file in the default locations [%s]",
		strings.Join(defaultLocations, ","),
	)
}

// Load the ENV Like:RETRIEVERS_0_HEADERS_AUTHORIZATION
func loadArrayEnv(s string, v string, configMap map[string]interface{}) error {
	paths := strings.Split(s, "_")
	for i, str := range paths {
		paths[i] = strings.ToLower(str)
	}
	prefixKey := paths[0]
	if configArray, ok := configMap[prefixKey].([]interface{}); ok {
		index, err := strconv.Atoi(paths[1])
		if err != nil {
			return err
		}
		var configItem map[string]interface{}
		outRange := index > len(configArray)-1
		if outRange {
			configItem = make(map[string]interface{})
		} else {
			configItem = configArray[index].(map[string]interface{})
		}

		keys := paths[2:]
		currentMap := configItem
		for i, key := range keys {
			hasKey := false
			lowerKey := key
			for y := range currentMap {
				if y != lowerKey {
					continue
				}
				if nextMap, ok := currentMap[y].(map[string]interface{}); ok {
					currentMap = nextMap
					hasKey = true
					break
				}
			}
			if !hasKey && i != len(keys)-1 {
				newMap := make(map[string]interface{})
				currentMap[lowerKey] = newMap
				currentMap = newMap
			}
		}
		lastKey := keys[len(keys)-1]
		currentMap[lastKey] = v
		if outRange {
			blank := index - len(configArray) + 1
			for i := 0; i < blank; i++ {
				configArray = append(configArray, make(map[string]interface{}))
			}
			configArray[index] = configItem
		} else {
			configArray[index] = configItem
		}
		_ = k.Set(prefixKey, configArray)
	}
	return nil
}

func (c *Config) IsDebugEnabled() bool {
	if c == nil {
		return false
	}
	return strings.ToLower(c.LogLevel) == "debug"
}

func (c *Config) ZapLogLevel() zapcore.Level {
	if c == nil {
		return zapcore.InvalidLevel
	}
	level, err := zapcore.ParseLevel(c.LogLevel)
	if err != nil {
		return zapcore.InvalidLevel
	}
	return level
}

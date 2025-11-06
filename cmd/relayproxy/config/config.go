package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
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

type Config struct {
	CommonFlagSet `mapstructure:",inline" koanf:",squash"`
	// ListenPort (optional) is the port we are using to start the proxy
	//
	// Deprecated: use Server.Port instead
	ListenPort int `mapstructure:"listen" koanf:"listen"`

	// MonitoringPort (optional) is the port we are using to expose the metrics and healthchecks
	// If not set we will use the same port as the proxy
	//
	// Deprecated: use Server.MonitoringPort instead
	MonitoringPort int `mapstructure:"monitoringPort" koanf:"monitoringport"`

	// Server is the server configuration, including host, port, and unix socket
	Server Server `mapstructure:"server" koanf:"server"`

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

	// ExporterCleanQueueInterval (optional) is the duration between each cleaning of the queue by the thread in charge
	// of removing the old events.
	// Default: 1 minute
	ExporterCleanQueueInterval time.Duration `mapstructure:"exporterCleanQueueInterval" koanf:"exportercleanqueueinterval"`

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
	//
	// Deprecated: use `Server.Mode = lambda` instead
	StartAsAwsLambda bool `mapstructure:"startAsAwsLambda" koanf:"startasawslambda"`

	// AwsLambdaAdapter (optional) is the adapter to use when the relay proxy is started as an AWS Lambda.
	// Possible values are "APIGatewayV1", "APIGatewayV2" and "ALB"
	// Default: "APIGatewayV2"
	//
	// Deprecated: use `Server.LambdaAdapter` instead
	AwsLambdaAdapter string `mapstructure:"awsLambdaAdapter" koanf:"awslambdaadapter"`

	// AwsApiGatewayBasePath (optional) is the base path prefix for AWS API Gateway deployments.
	// This is useful when deploying behind a non-root path like "/api" or "/dev/feature-flags".
	// The relay proxy will strip this base path from incoming requests before processing.
	// Example: if set to "/api/feature-flags", requests to "/api/feature-flags/health" will be processed as "/health"
	// Default: ""
	//
	// Deprecated: use `Server.AwsApiGatewayBasePath` instead
	AwsApiGatewayBasePath string `mapstructure:"awsApiGatewayBasePath" koanf:"awsapigatewaybasepath"`

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

	// FlagSets is the list of flag sets configured.
	// A flag set is a group of flags that can be used to configure the relay proxy.
	// Each flag set can have its own API key, retrievers, notifiers and exporters.
	// There is no inheritance between flag sets.
	FlagSets []FlagSet `mapstructure:"flagsets" koanf:"flagsets"`
	// ---- private fields

	// apiKeySet is the internal representation of an API keys list configured
	// we store them in a set to be
	apiKeysSet map[string]ApiKeyType

	// apiKeyPreload is used to be sure that the apiKeysSet is loaded only once.
	apiKeyPreload sync.Once

	// forceAuthenticatedRequests is true if we have at least 1 AuthorizedKey.Evaluation key set.
	forceAuthenticatedRequests bool
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
	loadConfigFile(log)

	// Map environment variables
	_ = k.Load(mapEnvVariablesProvider(k.String("envVariablePrefix"), log), nil)
	_ = k.Set("version", version)

	proxyConf := &Config{}
	errUnmarshal := k.Unmarshal("", &proxyConf)
	if errUnmarshal != nil {
		return nil, errUnmarshal
	}

	processExporters(proxyConf)

	return proxyConf, nil
}

// loadConfigFile handles the loading of configuration files
func loadConfigFile(log *zap.Logger) {
	configFileLocation, errFileLocation := locateConfigFile(k.String("config"))
	if errFileLocation != nil {
		log.Info("not using any configuration file", zap.Error(errFileLocation))
		return
	}

	parser := getParserForFile(configFileLocation)
	if errBindFile := k.Load(file.Provider(configFileLocation), parser); errBindFile != nil {
		log.Error("error loading file", zap.Error(errBindFile))
	}
}

// getParserForFile returns the appropriate parser based on file extension
func getParserForFile(configFileLocation string) koanf.Parser {
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

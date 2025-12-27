package config

import (
	"strings"
	"sync"
	"time"

	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	// CommonFlagSet is the common flag set for the relay proxy
	CommonFlagSet `mapstructure:",inline" koanf:",squash"`

	// Server is the server configuration, including host, port, and unix socket
	Server Server `mapstructure:"server" koanf:"server"`

	// Swagger is the swagger configuration
	Swagger Swagger `mapstructure:"swagger" koanf:"swagger"`

	// HideBanner (optional) if true, we don't display the go-feature-flag relay proxy banner
	HideBanner bool `mapstructure:"hideBanner" koanf:"hidebanner"`

	// EnablePprof (optional) if true, go-feature-flag relay proxy will start
	// the pprof endpoints on the same port as the monitoring.
	// Default: false
	EnablePprof bool `mapstructure:"enablePprof" koanf:"enablepprof"`

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

	// AuthorizedKeys list of API keys that authorized to use endpoints
	AuthorizedKeys APIKeys `mapstructure:"authorizedKeys" koanf:"authorizedkeys"`

	// EvaluationContextEnrichment (optional) will be merged with the evaluation context sent during the evaluation.
	// It is useful to add common attributes to all the evaluations, such as a server version, environment, ...
	//
	// All those fields will be included in the custom attributes of the evaluation context,
	// if in the evaluation context you have a field with the same name,
	// it will be overridden by the evaluationContextEnrichment.
	// Default: nil
	EvaluationContextEnrichment map[string]any `mapstructure:"evaluationContextEnrichment" koanf:"evaluationcontextenrichment"` //nolint: lll

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

	// EnableBulkMetricFlagNames (optional) enables per-flag metrics for bulk evaluation endpoints.
	// This adds flag_name labels to the all_flags_evaluations_total_with_flag metric.
	// Default: false
	EnableBulkMetricFlagNames bool `mapstructure:"enableBulkMetricFlagNames" koanf:"enablebulkmetricflagnames"`

	// FlagSets is the list of flag sets configured.
	// A flag set is a group of flags that can be used to configure the relay proxy.
	// Each flag set can have its own API key, retrievers, notifiers and exporters.
	// There is no inheritance between flag sets.
	FlagSets []FlagSet `mapstructure:"flagsets" koanf:"flagsets"`

	// ---------- Deprecated fields ----------
	// ListenPort (optional) is the port we are using to start the proxy
	//
	// Deprecated: use Server.Port instead
	ListenPort int `mapstructure:"listen" koanf:"listen"`

	// MonitoringPort (optional) is the port we are using to expose the metrics and healthchecks
	// If not set we will use the same port as the proxy
	//
	// Deprecated: use Server.MonitoringPort instead
	MonitoringPort int `mapstructure:"monitoringPort" koanf:"monitoringport"`

	// Deprecated: use Swagger.Enabled instead
	EnableSwagger bool `mapstructure:"enableSwagger" koanf:"enableswagger"`

	// Deprecated: use Swagger.Host instead
	Host string `mapstructure:"host" koanf:"host"`

	// Deprecated: use AuthorizedKeys instead
	// APIKeys list of API keys that authorized to use endpoints
	APIKeys []string `mapstructure:"apiKeys" koanf:"apikeys"`

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
	// ---------- End of deprecated fields ----------

	// ---------- Private fields ----------

	// apiKeySet is the internal representation of an API keys list configured
	// we store them in a set to be
	apiKeysSet map[string]ApiKeyType

	// apiKeyPreload is used to be sure that the apiKeysSet is loaded only once.
	apiKeyPreload sync.Once

	// forceAuthenticatedRequests is true if we have at least 1 AuthorizedKey.Evaluation key set.
	forceAuthenticatedRequests bool

	// configLoader is the service in charge of loading the configuration.
	configLoader *ConfigLoader

	// logger is the logger for the relay proxy
	logger *zap.Logger
	// ---------- End of private fields ----------
}

// New is reading the configuration file
func New(cmdLineFlagSet *pflag.FlagSet, log *zap.Logger, version string) (*Config, error) {
	// Map environment variables
	configLoader := NewConfigLoader(cmdLineFlagSet, log, version, true)
	proxyConf, errUnmarshal := configLoader.ToConfig()
	if errUnmarshal != nil {
		return nil, errUnmarshal
	}
	proxyConf.configLoader = configLoader
	proxyConf.logger = log
	return proxyConf, nil
}

// IsDebugEnabled returns true if the log level is debug
func (c *Config) IsDebugEnabled() bool {
	if c == nil {
		return false
	}
	return strings.ToLower(c.LogLevel) == "debug"
}

// ZapLogLevel returns the zap core level for the log level
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

// IsUsingFlagsets returns true if the configuration is using flagsets
func (c *Config) IsUsingFlagsets() bool {
	return len(c.FlagSets) > 0
}

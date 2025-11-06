package config

import (
	"errors"

	"go.uber.org/zap"
)

type ServerMode = string

const (
	// ServerModeHTTP is the HTTP server mode.
	ServerModeHTTP ServerMode = "http"
	// ServerModeLambda is the AWS Lambda server mode.
	ServerModeLambda ServerMode = "lambda"
	// ServerModeUnixSocket is the Unix Socket server mode.
	ServerModeUnixSocket ServerMode = "unixsocket"
)

type LambdaAdapter = string

const (
	LambdaAdapterAPIGatewayV1 LambdaAdapter = "APIGatewayV1"
	LambdaAdapterAPIGatewayV2 LambdaAdapter = "APIGatewayV2"
	LambdaAdapterALB          LambdaAdapter = "ALB"
)

type Server struct {
	// Mode is the server mode.
	// default: http
	Mode ServerMode `mapstructure:"mode" koanf:"mode"`

	// Host is the server host.
	// default: 0.0.0.0
	Host string `mapstructure:"host" koanf:"host"`

	// Port is the server port.
	// default: 1031
	Port int `mapstructure:"port" koanf:"port"`

	// MonitoringPort is the monitoring port. It can be specified only if the server mode is HTTP.
	// default: none, it will use the same as server port.
	MonitoringPort int `mapstructure:"monitoringPort" koanf:"monitoringport"`

	// UnixSocket is the server unix socket path.
	UnixSocketPath string `mapstructure:"unixSocketPath" koanf:"unixsocketpath"`

	// AWS Lambda configuration
	// LambdaAdapter is the adapter to use when the relay proxy is started as an AWS Lambda.
	// default: APIGatewayV2
	LambdaAdapter LambdaAdapter `mapstructure:"awsLambdaAdapter" koanf:"awsLambdaAdapter"`

	// AwsApiGatewayBasePath (optional) is the base path prefix for AWS API Gateway deployments.
	// This is useful when deploying behind a non-root path like "/api" or "/dev/feature-flags".
	// The relay proxy will strip this base path from incoming requests before processing.
	// Example: if set to "/api/feature-flags", requests to "/api/feature-flags/health" will be processed as "/health"
	// Default: ""
	AwsApiGatewayBasePath string `mapstructure:"awsApiGatewayBasePath" koanf:"awsapigatewaybasepath"`
}

func (s *Server) Validate() error {
	switch s.Mode {
	case ServerModeUnixSocket:
		if s.UnixSocketPath == "" {
			return errors.New("unixSocket must be set when server mode is unixsocket")
		}
		return nil
	default:
		return nil
	}
}

// GetMonitoringPort returns the monitoring port, checking first the top-level config
// and then the server config.
func (c *Config) GetMonitoringPort(logger *zap.Logger) int {
	if c.Server.MonitoringPort != 0 {
		return c.Server.MonitoringPort
	}
	if c.MonitoringPort != 0 {
		if logger != nil {
			logger.Warn("The monitoring port is set using `monitoringPort`, this option is deprecated, please migrate to `server.monitoringPort`")
		}
		return c.MonitoringPort
	}
	return 0
}

// GetServerHost returns the server host, defaulting to "0.0.0.0" if not set.
func (c *Config) GetServerHost() string {
	if c.Server.Host != "" {
		return c.Server.Host
	}
	return "0.0.0.0"
}

// GetServerPort returns the server port, checking first the server config
// and then the top-level config, defaulting to 1031 if not set.
func (c *Config) GetServerPort(logger *zap.Logger) int {
	if c.Server.Port != 0 {
		return c.Server.Port
	}

	if c.ListenPort != 0 {
		if logger != nil {
			logger.Warn("The server port is set using `port`, this option is deprecated, please migrate to `server.port`")
		}
		return c.ListenPort
	}
	return 1031
}

// GetServerMode returns the server mode, checking first the server config
// and then the top-level config, defaulting to HTTP if not set.
func (c *Config) GetServerMode(logger *zap.Logger) ServerMode {
	if c.Server.Mode != "" {
		return c.Server.Mode
	}

	if c.StartAsAwsLambda {
		if logger != nil {
			logger.Warn("The server mode is set using `startAsAwsLambda`, this option is deprecated, please migrate to `server.mode`")
		}
		return ServerModeLambda
	}

	return ServerModeHTTP
}

// GetLambdaAdapter returns the lambda adapter, checking first the server config
// and then the top-level config, defaulting to APIGatewayV2 if not set.
func (c *Config) GetLambdaAdapter(logger *zap.Logger) LambdaAdapter {
	if c.Server.LambdaAdapter != "" {
		return c.Server.LambdaAdapter
	}

	if c.AwsLambdaAdapter != "" {
		if logger != nil {
			logger.Warn("The lambda adapter is set using `awsLambdaAdapter`, this option is deprecated, please migrate to `server.awsLambdaAdapter`")
		}
		return LambdaAdapter(c.AwsLambdaAdapter)
	}

	return LambdaAdapterAPIGatewayV2
}

// GetAwsApiGatewayBasePath returns the AWS API Gateway base path, checking first the server config
// and then the top-level config, defaulting to empty string if not set.
func (c *Config) GetAwsApiGatewayBasePath(logger *zap.Logger) string {
	if c.Server.AwsApiGatewayBasePath != "" {
		return c.Server.AwsApiGatewayBasePath
	}

	if c.AwsApiGatewayBasePath != "" {
		if logger != nil {
			zap.L().Warn("The AWS API Gateway base path is set using `awsApiGatewayBasePath`, this option is deprecated, please migrate to `server.awsApiGatewayBasePath`")
		}
		return c.AwsApiGatewayBasePath
	}

	return ""
}

// GetUnixSocketPath returns the unix socket path.
func (c *Config) GetUnixSocketPath() string {
	return c.Server.UnixSocketPath
}

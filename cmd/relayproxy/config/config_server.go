package config

type Server struct {
	Host           string `mapstructure:"host" koanf:"host"`
	Port           int    `mapstructure:"port" koanf:"port"`
	UnixSocket     string `mapstructure:"unixSocket" koanf:"unixsocket"`
	MonitoringPort int    `mapstructure:"monitoringPort" koanf:"monitoringport"`
}

// GetMonitoringPort returns the monitoring port, checking first the top-level config
// and then the server config.
func (c *Config) GetMonitoringPort() int {
	if c.MonitoringPort != 0 {
		return c.MonitoringPort
	}
	return c.Server.MonitoringPort
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
func (c *Config) GetServerPort() int {
	if c.Server.Port != 0 {
		return c.Server.Port
	}

	if c.ListenPort != 0 {
		return c.ListenPort
	}
	return 1031
}

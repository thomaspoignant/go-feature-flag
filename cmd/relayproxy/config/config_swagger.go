package config

type Swagger struct {
	// Enabled is the flag to enable the swagger.
	Enabled bool `mapstructure:"enabled" koanf:"enabled"`

	// Host is the host to use for the swagger.
	Host string `mapstructure:"host" koanf:"host"`
}

// IsSwaggerEnabled is the function to check if the swagger is enabled.
// If the Swagger.Enabled is true, we return true.
// If the EnableSwagger is true, we return true.
// If both are false, we return false.
func (c *Config) IsSwaggerEnabled() bool {
	if c.Swagger.Enabled {
		return true
	}
	return c.EnableSwagger
}

// SwaggerHost returns the swagger host.
// If the Swagger.Host is set, we return it.
// If the Host is set, we return it.
// If both are not set, we return "localhost".
func (c *Config) SwaggerHost() string {
	if c.Swagger.Host != "" {
		return c.Swagger.Host
	}
	if c.Host != "" {
		return c.Host
	}
	return "localhost"
}

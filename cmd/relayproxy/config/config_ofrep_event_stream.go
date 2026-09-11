package config

// OfrepEventStream holds the configuration used to advertise the flag-change SSE
// endpoint in the OFREP bulk evaluation response (OpenFeature ADR-0008).
type OfrepEventStream struct {
	// Enabled turns on the eventStreams field in the OFREP bulk evaluation response.
	Enabled bool `mapstructure:"enabled" koanf:"enabled"`

	// Endpoint is the public base URL clients should use to reach the relay proxy SSE
	// endpoint, for example https://gofeatureflag.example.com. The path of the SSE
	// endpoint is appended automatically. It is required when Enabled is true.
	Endpoint string `mapstructure:"endpoint" koanf:"endpoint"`

	// InactivityDelaySec is advertised to the client as the delay in seconds after which
	// it should consider the stream inactive. When 0 the field is omitted.
	InactivityDelaySec int `mapstructure:"inactivityDelaySec" koanf:"inactivitydelaysec"`
}

package config

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

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
	Sampler JaegerSampler `mapstructure:"sampler" koanf:"sampler"`
}

// JaegerSampler is the configuration object to configure the sampling.
type JaegerSampler struct {
	Manager JaegerSamplerManager `mapstructure:"manager" koanf:"manager"`
	Refresh JaegerSamplerRefresh `mapstructure:"refresh" koanf:"refresh"`
	Max     JaegerSamplerMax     `mapstructure:"max" koanf:"max"`
}

// JaegerSamplerManager is the configuration object to configure the manager of the sampling.
type JaegerSamplerManager struct {
	Host JaegerSamplerManagerHost `mapstructure:"host" koanf:"host"`
}

// JaegerSamplerManagerHost is the configuration object to configure the host of the manager of the sampling.
type JaegerSamplerManagerHost struct {
	Port string `mapstructure:"port" koanf:"port"`
}

// JaegerSamplerRefresh is the configuration object to configure the refresh of the sampling.
type JaegerSamplerRefresh struct {
	Interval string `mapstructure:"interval" koanf:"interval"`
}

// JaegerSamplerMax is the configuration object to configure the max of the sampling.
type JaegerSamplerMax struct {
	Operations int `mapstructure:"operations" koanf:"operations"`
}

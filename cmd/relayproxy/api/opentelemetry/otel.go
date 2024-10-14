package opentelemetry

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/contrib/samplers/jaegerremote"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
)

type OtelService struct {
	otelTraceProvider *sdktrace.TracerProvider
	otelExporter      sdktrace.SpanExporter
}

func NewOtelService() OtelService {
	return OtelService{}
}

// Init the OpenTelemetry service
func (s *OtelService) Init(ctx context.Context, zapLog *zap.Logger, config config.Config) error {
	// OTEL_SDK_DISABLED is not supported by the Go SDK, but is a standard env
	// var defined by the OTel spec. We'll use it to disable the trace provider.
	if disabled, _ := strconv.ParseBool(os.Getenv("OTEL_SDK_DISABLED")); disabled {
		otel.SetTracerProvider(noop.NewTracerProvider())
		return nil
	}

	// support the openTelemetryOtlpEndpoint config element
	if config.OpenTelemetryOtlpEndpoint != "" &&
		os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") == "" {
		os.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", config.OpenTelemetryOtlpEndpoint)
	}

	exporter, err := autoexport.NewSpanExporter(ctx)
	if err != nil {
		return fmt.Errorf("initializing OTel exporter: %w", err)
	}

	serviceName := "go-feature-flag"
	if v := os.Getenv("OTEL_SERVICE_NAME"); v != "" {
		serviceName = v
	}

	sampler, err := initSampler(serviceName)
	if err != nil {
		return fmt.Errorf("initializing OTel sampler: %w", err)
	}

	resource, err := initResource(ctx, serviceName, config.Version)
	if err != nil {
		return fmt.Errorf("initializing OTel resources: %w", err)
	}

	s.otelExporter = exporter
	s.otelTraceProvider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithSampler(sampler),
		sdktrace.WithResource(resource),
	)

	otel.SetTracerProvider(s.otelTraceProvider)

	// log OTel errors to zap rather than the default log package
	otel.SetErrorHandler(otelErrHandler(func(err error) {
		zapLog.Error("OTel error", zap.Error(err))
	}))

	return nil
}

type otelErrHandler func(err error)

func (o otelErrHandler) Handle(err error) {
	o(err)
}

var _ otel.ErrorHandler = otelErrHandler(nil)

func initResource(ctx context.Context, serviceName, version string) (*resource.Resource, error) {
	return resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithProcessPID(),
		resource.WithProcessExecutableName(),
		resource.WithProcessExecutablePath(),
		resource.WithProcessOwner(),
		resource.WithProcessRuntimeName(),
		resource.WithProcessRuntimeVersion(),
		resource.WithProcessRuntimeDescription(),
		resource.WithHost(),
		resource.WithTelemetrySDK(),
		resource.WithOS(),
		resource.WithContainer(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(version),
		),
	)
}

// initSampler determines which sampling strategy to use. If OTEL_TRACES_SAMPLER
// is unset, we'll always sample.
// If it's set to jaeger_remote, we'll use the Jaeger sampling server (supports
// JAEGER_SAMPLER_MANAGER_HOST_PORT, JAEGER_SAMPLER_REFRESH_INTERVAL, and
// JAEGER_SAMPLER_MAX_OPERATIONS).
// If it's set to any other value, we return nil and sdktrace.NewTracerProvider
// will set up the initSampler from the environment.
func initSampler(serviceName string) (sdktrace.Sampler, error) {
	sampler, ok := os.LookupEnv("OTEL_TRACES_SAMPLER")
	if !ok {
		return sdktrace.AlwaysSample(), nil
	}

	if sampler != "jaeger_remote" {
		return nil, nil
	}

	samplerURL, samplerRefreshInterval, maxOperations, err := jaegerRemoteSamplerOpts()
	if err != nil {
		return nil, err
	}

	return jaegerremote.New(
		serviceName,
		jaegerremote.WithSamplingServerURL(samplerURL),
		jaegerremote.WithSamplingRefreshInterval(samplerRefreshInterval),
		jaegerremote.WithMaxOperations(maxOperations),
		jaegerremote.WithInitialSampler(sdktrace.AlwaysSample()),
	), nil
}

const (
	defaultSamplerURL              = "http://localhost:5778/sampling"
	defaultSamplingRefreshInterval = 1 * time.Minute
	defaultSamplingMaxOperations   = 256
)

func jaegerRemoteSamplerOpts() (string, time.Duration, int, error) {
	samplerURL := defaultSamplerURL
	if v := os.Getenv("JAEGER_SAMPLER_MANAGER_HOST_PORT"); v != "" {
		samplerURL = v
	}

	samplerRefreshInterval := defaultSamplingRefreshInterval
	if v := os.Getenv("JAEGER_SAMPLER_REFRESH_INTERVAL"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return "", 0, 0, fmt.Errorf("parsing JAEGER_SAMPLER_REFRESH_INTERVAL: %w", err)
		}
		samplerRefreshInterval = d
	}

	maxOperations := defaultSamplingMaxOperations
	if v := os.Getenv("JAEGER_SAMPLER_MAX_OPERATIONS"); v != "" {
		i, err := strconv.Atoi(v)
		if err != nil {
			return "", 0, 0, fmt.Errorf("parsing JAEGER_SAMPLER_MAX_OPERATIONS: %w", err)
		}
		maxOperations = i
	}

	return samplerURL, samplerRefreshInterval, maxOperations, nil
}

// Stop the OpenTelemetry service
func (s *OtelService) Stop(ctx context.Context) error {
	if s.otelExporter != nil {
		err := s.otelExporter.Shutdown(ctx)
		if err != nil {
			return err
		}
	}

	if s.otelTraceProvider != nil {
		err := s.otelTraceProvider.Shutdown(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

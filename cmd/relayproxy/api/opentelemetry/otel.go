package opentelemetry

import (
	"context"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"net/url"
)

type OtelService struct {
	otelTraceProvider *sdktrace.TracerProvider
	otelExporter      *otlptrace.Exporter
}

func NewOtelService() OtelService {
	return OtelService{}
}

// Init the OpenTelemetry service
func (s *OtelService) Init(ctx context.Context, config config.Config) error {
	// parsing the OpenTelemetry endpoint
	u, err := url.Parse(config.OpenTelemetryOtlpEndpoint)
	if err != nil {
		return err
	}

	var opts []otlptracehttp.Option
	if u.Scheme == "http" {
		opts = append(opts, otlptracehttp.WithInsecure())
	}
	opts = append(opts, otlptracehttp.WithEndpoint(u.Host))
	client := otlptracehttp.NewClient(opts...)

	s.otelExporter, err = otlptrace.New(ctx, client)
	if err != nil {
		return err
	}

	s.otelTraceProvider = sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(s.otelExporter),
		sdktrace.WithResource(resource.NewSchemaless(
			attribute.String("service.name", "go-feature-flag"),
			attribute.String("service.version", config.Version),
		)),
	)
	otel.SetTracerProvider(s.otelTraceProvider)
	return nil
}

// Stop the OpenTelemetry service
func (s *OtelService) Stop() error {
	if s.otelExporter != nil {
		err := s.otelExporter.Shutdown(context.Background())
		if err != nil {
			return err
		}
	}
	if s.otelTraceProvider != nil {
		err := s.otelTraceProvider.Shutdown(context.Background())
		if err != nil {
			return err
		}
	}
	return nil
}

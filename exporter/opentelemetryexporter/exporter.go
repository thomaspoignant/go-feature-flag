package opentelemetryexporter

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/thomaspoignant/go-feature-flag/exporter"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"

	"log"
)

const (
	serviceName            = "go-feature-flag"
	ServiceVersion         = "0.0.1"
	instrumentationName    = "github.com/thomaspoignant/go-feature-flag"
	instrumentationVersion = "0.0.1"
)

var tracer = otel.GetTracerProvider().Tracer(
	instrumentationName,
	trace.WithInstrumentationVersion(instrumentationVersion),
	trace.WithSchemaURL(semconv.SchemaURL),
)

type Exporter struct {
	resource   *resource.Resource
	processors []*sdktrace.SpanProcessor
}

type ExporterOption func(*Exporter)

func NewExporter(opts ...ExporterOption) *Exporter {
	exporter := Exporter{}
	for _, opt := range opts {
		opt(&exporter)
	}
	return &exporter
}

func WithResource(customResource *resource.Resource) ExporterOption {
	return func(exp *Exporter) {
		mergedResource, err := resource.Merge(customResource, defaultResource())
		if err != nil {
			panic("Unable to merge resources")
		}
		exp.resource = mergedResource
	}
}

func WithBatchSpanProcessors(processors ...*sdktrace.SpanProcessor) ExporterOption {
	return func(exp *Exporter) {
		exp.processors = processors
	}
}

func defaultResource() *resource.Resource {
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(serviceName),
		semconv.ServiceVersion(ServiceVersion),
	)
}

func otelExporter(uri string, opts ...grpc.DialOption) (*otlptrace.Exporter, error) {
	// TODO creds

	if len(opts) == 0 {
		return nil, errors.New("need credentials option")
	}

	conn, err := grpc.NewClient(uri, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	// Set up a trace exporter
	traceExporter, err := otlptracegrpc.New(context.Background(), otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}
	return traceExporter, nil
}

func OtelCollectorBatchSpanProcessor(uri string, opts ...grpc.DialOption) (sdktrace.SpanProcessor, error) {
	otelExporter, err := otelExporter(uri, opts...)
	if err != nil {
		return nil, err
	}

	return sdktrace.NewBatchSpanProcessor(otelExporter), nil
}

func newstdoutExporter() (*stdouttrace.Exporter, error) {
	exp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, fmt.Errorf("failed to initialize stdouttrace exporter: %w", err)
	}
	return exp, nil
}

func stdoutBatchSpanProcessor() (sdktrace.SpanProcessor, error) {
	inMemoryExporter, err := newstdoutExporter()
	if err != nil {
		return nil, err
	}

	return sdktrace.NewBatchSpanProcessor(inMemoryExporter), nil
}

func valueToAttributes(data interface{}, parentName string, maxDepth int, recursionDepth int) []attribute.KeyValue {
	parentName = strings.ToLower(parentName)
	reflectedAttributes := make([]attribute.KeyValue, 0)

	if recursionDepth > maxDepth {
		return reflectedAttributes
	}

	targetType := reflect.TypeOf(data)
	targetValue := reflect.ValueOf(data)
	kind := targetValue.Kind()

	switch kind {
	case reflect.Float32, reflect.Float64:
		reflectedAttributes = append(reflectedAttributes, attribute.Float64(parentName, targetValue.Float()))
	case reflect.Int, reflect.Int8, reflect.Int16,
		reflect.Int32, reflect.Int64:
		reflectedAttributes = append(reflectedAttributes, attribute.Int64(parentName, targetValue.Int()))
	case reflect.Bool:
		reflectedAttributes = append(reflectedAttributes, attribute.Bool(parentName, targetValue.Bool()))
	case reflect.String:
		reflectedAttributes = append(reflectedAttributes, attribute.String(parentName, targetValue.String()))

	case reflect.Struct:
		for i := 0; i < targetType.NumField(); i++ {
			name := targetType.Field(i).Name
			fv := targetValue.Field(i)

			if !fv.CanInterface() {
				continue
			}

			subAttributes := valueToAttributes(fv.Interface(), parentName+"."+name, maxDepth, recursionDepth+1)
			reflectedAttributes = append(reflectedAttributes, subAttributes...)
		}

	case reflect.Invalid:
	default:
	}

	return reflectedAttributes
}

func featureEventToAttributes(featureEvent exporter.FeatureEvent) []attribute.KeyValue {
	// https://opentelemetry.io/docs/specs/semconv/feature-flags/feature-flags-spans/

	attributes := make([]attribute.KeyValue, 0)
	attributes = append(attributes, attribute.String("kind", featureEvent.Kind),
		attribute.String("contextKind", featureEvent.ContextKind),
		attribute.String("userKey", featureEvent.UserKey),
		attribute.Int64("creationDate", featureEvent.CreationDate),
		attribute.String("key", featureEvent.Key),
		attribute.String("variation", featureEvent.Variation),
		attribute.Bool("default", featureEvent.Default),
		attribute.String("version", featureEvent.Version),
		attribute.String("source", featureEvent.Source))

	valueAttrs := valueToAttributes(featureEvent.Value, "value", 2, 0)
	attributes = append(attributes, valueAttrs...)

	return attributes
}

func initProvider(exp *Exporter) (func(context.Context) error, error) {
	// The default resource will win on merge

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(exp.resource),
	)

	if len(exp.processors) == 0 {
		return nil, errors.New("no processors provided")
	}

	for _, spanProcessor := range exp.processors {
		tracerProvider.RegisterSpanProcessor(*spanProcessor)
	}

	otel.SetTracerProvider(tracerProvider)

	// set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.TraceContext{})

	// Shutdown will flush any remaining spans and shut down the exporter.
	return func(ctx context.Context) error {
		err := tracerProvider.ForceFlush(ctx)
		if err != nil {
			return err
		}
		return tracerProvider.Shutdown(ctx)
	}, nil
}

func eventToSpan(ctx context.Context, featureEvent exporter.FeatureEvent) {
	attributes := featureEventToAttributes(featureEvent)
	_, span := tracer.Start(ctx, featureEvent.Kind)
	defer span.End()
	span.SetAttributes(attributes...)
	// How can we detect feature-flag evaluation failure?
	span.SetStatus(codes.Ok, "n/a")
}
func eventsToSpans(ctx context.Context, featureEvents []exporter.FeatureEvent) {
	for _, featureEvent := range featureEvents {
		eventToSpan(ctx, featureEvent)
	}
}

func (exporter *Exporter) Export(ctx context.Context, _ *log.Logger, featureEvents []exporter.FeatureEvent) error {
	shutdown, err := initProvider(exporter)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			log.Fatal("failed to shutdown TracerProvider: %w", err)
		}
	}()

	ctx, span := tracer.Start(ctx, "feature-flag-evaluation")
	defer span.End()
	eventsToSpans(ctx, featureEvents)

	return nil
}

package opentelemetryexporter

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/thomaspoignant/go-feature-flag/exporter"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	defaultTracerName = "go-feature-flag"
	spanName          = "feature_flag.evaluate"
)

type Exporter struct {
	// TracerName allows overriding the OpenTelemetry tracer name.
	// Default: go-feature-flag
	TracerName string

	tracer trace.Tracer
	once   sync.Once
}

func (e *Exporter) init() {
	e.once.Do(func() {
		if e.TracerName == "" {
			e.TracerName = defaultTracerName
		}
		e.tracer = otel.Tracer(e.TracerName)
	})
}

func (e *Exporter) Export(
	ctx context.Context,
	logger *fflog.FFLogger,
	events []exporter.ExportableEvent,
) error {
	e.init()

	for _, evt := range events {
		featureEvent, ok := evt.(exporter.FeatureEvent)
		if !ok {
			continue
		}

		e.createSpan(ctx, featureEvent, logger)
	}

	return nil
}

func (e *Exporter) createSpan(ctx context.Context, f exporter.FeatureEvent, logger *fflog.FFLogger) {
	startTime := time.Unix(f.CreationDate, 0)

	_, span := e.tracer.Start(
		ctx,
		spanName,
		trace.WithTimestamp(startTime),
	)
	defer span.End()

	// Collect all attributes into a single slice, then set them once on the span.
	attributes := make([]attribute.KeyValue, 0, 8+len(f.Metadata))
	attributes = append(attributes,
		attribute.String("feature_flag.key", f.Key),
		attribute.String("feature_flag.user_key", f.UserKey),
		attribute.String("feature_flag.context_kind", f.ContextKind),
		attribute.String("feature_flag.variation", f.Variation),
		attribute.Bool("feature_flag.default", f.Default),
		attribute.String("feature_flag.version", f.Version),
		attribute.String("feature_flag.source", f.Source),
	)
	logger.Debug("Exporting feature flag evaluation")

	// Value (safe stringification)
	if f.Value != nil {
		if b, err := json.Marshal(f.Value); err == nil {
			attributes = append(attributes, attribute.String("feature_flag.value", string(b)))
		}
	}

	// Metadata
	for k, v := range f.Metadata {
		key := "feature_flag.metadata." + k
		switch value := v.(type) {
		case bool:
			attributes = append(attributes, attribute.Bool(key, value))
		case int:
			attributes = append(attributes, attribute.Int(key, value))
		case int64:
			attributes = append(attributes, attribute.Int64(key, value))
		case float64:
			attributes = append(attributes, attribute.Float64(key, value))
		case string:
			attributes = append(attributes, attribute.String(key, value))
		default:
			attributes = append(attributes, attribute.String(key, fmt.Sprint(v)))
		}
	}

	logger.Debug("Setting span attributes for feature flag evaluation")
	// Apply all attributes in a single call to reduce overhead.
	if len(attributes) > 0 {
		span.SetAttributes(attributes...)
	}

	// Status
	if f.Default {
		span.SetStatus(codes.Error, "default value used")
	} else {
		span.SetStatus(codes.Ok, "evaluation successful")
	}
	logger.Debug("Feature flag evaluation span status set")
}

func (e *Exporter) IsBulk() bool {
	return false
}

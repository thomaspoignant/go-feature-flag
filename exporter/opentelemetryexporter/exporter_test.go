package opentelemetryexporter

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/exporter"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestOpenTelemetryIsBulk(t *testing.T) {
	exp := Exporter{}
	assert.False(t, exp.IsBulk(), "OpenTelemetry exporter is not a bulk exporter")
}

func TestExporterCreatesOneSpanPerEvent(t *testing.T) {
	// Arrange
	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(recorder),
	)
	otel.SetTracerProvider(tp)

	exp := &Exporter{
		TracerName: "test-tracer",
	}

	events := []exporter.ExportableEvent{
		exporter.FeatureEvent{
			Kind:         "feature",
			ContextKind:  "user",
			UserKey:      "user-1",
			CreationDate: 1617970547,
			Key:          "flag-1",
			Variation:    "A",
			Value:        true,
			Default:      false,
			Version:      "v1",
			Source:       "SERVER",
		},
		exporter.FeatureEvent{
			Kind:         "feature",
			ContextKind:  "anonymousUser",
			UserKey:      "user-2",
			CreationDate: 1617970548,
			Key:          "flag-2",
			Variation:    "B",
			Value:        "on",
			Default:      true,
			Version:      "v2",
			Source:       "PROVIDER_CACHE",
		},
	}

	logger := &fflog.FFLogger{}

	// Act
	err := exp.Export(context.Background(), logger, events)

	// Assert
	assert.NoError(t, err)

	spans := recorder.Ended()
	assert.Len(t, spans, 2, "should create one span per feature event")

	assert.Equal(t, spanName, spans[0].Name())
	assert.Equal(t, spanName, spans[1].Name())
}

func TestExporterSpanAttributes(t *testing.T) {
	// Arrange
	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(recorder),
	)
	otel.SetTracerProvider(tp)

	exp := &Exporter{}

	event := exporter.FeatureEvent{
		Kind:         "feature",
		ContextKind:  "user",
		UserKey:      "user-123",
		CreationDate: 1617970547,
		Key:          "my-flag",
		Variation:    "on",
		Value:        true,
		Default:      false,
		Version:      "v1",
		Source:       "SERVER",
	}

	// Act
	err := exp.Export(context.Background(), &fflog.FFLogger{}, []exporter.ExportableEvent{event})

	// Assert
	assert.NoError(t, err)

	spans := recorder.Ended()
	assert.Len(t, spans, 1)

	attrs := spans[0].Attributes()

	assertAttribute(t, attrs, "feature_flag.key", "my-flag")
	assertAttribute(t, attrs, "feature_flag.user_key", "user-123")
	assertAttribute(t, attrs, "feature_flag.context_kind", "user")
	assertAttribute(t, attrs, "feature_flag.variation", "on")
	assertAttribute(t, attrs, "feature_flag.version", "v1")
	assertAttribute(t, attrs, "feature_flag.source", "SERVER")
}

func assertAttribute(t *testing.T, attrs []attribute.KeyValue, key string, expected any) {
	t.Helper()
	for _, attr := range attrs {
		if string(attr.Key) == key {
			if attr.Value.AsInterface() != expected {
				t.Fatalf(
					"attribute %s mismatch: expected=%v got=%v",
					key,
					expected,
					attr.Value.AsInterface(),
				)
			}
			return
		}
	}
	t.Fatalf("attribute %s not found", key)
}

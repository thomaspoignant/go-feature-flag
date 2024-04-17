package opentelemetryexporter

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/exporter"
)

func TestFeatureEventsToAttributes(t *testing.T) {

	// TODO: Find various kinds of events
	featureEvents := []exporter.FeatureEvent{
		{
			Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
			Variation: "Default", Value: "YO", Default: false,
		},
		{
			Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCDEF", CreationDate: 1617970547, Key: "random-key",
			Variation: "Default", Value: "YO", Default: false,
		},
	}

	for _, featureEvent := range featureEvents {
		attributes := featureEventToAttributes(featureEvent)
		assert.Len(t, attributes, 9)

	}

}

func TestResource(t *testing.T) {

	resource := Resource()
	assert.NotNil(t, resource)
	assert.NotNil(t, resource.SchemaURL())

	attributes := resource.Attributes()
	assert.Len(t, attributes, 2)

}

func TestExporterBuilding(t *testing.T) {

	resource := Resource()
	batchOptions := sdktrace.BatchSpanProcessorOptions{MaxQueueSize: 10, MaxExportBatchSize: 100, BatchTimeout: time.Millisecond * 100}
	exporter := NewExporter(
		WithBatchSpanProcessorOption(batchOptions),
		WithResource(resource),
	)
	assert.NotNil(t, exporter)
	assert.NotNil(t, exporter.Resource)
	assert.NotNil(t, exporter.BatchSpanProcessorOptions)
	assert.Equal(t, exporter.BatchSpanProcessorOptions, batchOptions)

}

func TestInitProvider(t *testing.T) {

	_, err := initProvider(&Exporter{})
	assert.Nil(t, err)
}

func TestPersistentInMemoryExporter(t *testing.T) {

	ctx := context.Background()

	spanExporter := PersistentInMemoryExporter{}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(&spanExporter)))
	tracer := tp.Tracer("tracer")
	_, span := tracer.Start(ctx, "span")
	span.End()

	err := tp.ForceFlush(ctx)
	assert.NoError(t, err)

	assert.Len(t, spanExporter.GetSpans(), 1)

}

func TestExportWithMultipleProcessors(t *testing.T) {

	featureEvents := []exporter.FeatureEvent{
		{
			Kind: "feature1", ContextKind: "anonymousUser1", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
			Variation: "Default", Value: "YO", Default: false,
		},
		{
			Kind: "feature2", ContextKind: "anonymousUser2", UserKey: "ABCDEF", CreationDate: 1617970547, Key: "random-key",
			Variation: "Default", Value: "YO", Default: false,
		},
	}

	ctx := context.Background()
	logger := log.New(os.Stdout, "", 0)

	spanExporter := PersistentInMemoryExporter{}
	inMemoryProcessor := sdktrace.NewBatchSpanProcessor(&spanExporter)
	stdoutProcessor, err := stdoutBatchSpanProcessor()
	assert.Nil(t, err)
	resource := Resource()
	batchOptions := sdktrace.BatchSpanProcessorOptions{MaxQueueSize: 10, MaxExportBatchSize: 100, BatchTimeout: time.Millisecond * 100}
	processors := make([]*sdktrace.SpanProcessor, 0)
	processors = append(processors, &inMemoryProcessor)
	processors = append(processors, &stdoutProcessor)
	exp := NewExporter(
		WithBatchSpanProcessorOption(batchOptions),
		WithResource(resource),
		WithBatchSpanProcessors(processors),
	)
	exp.Export(ctx, logger, featureEvents)
	assert.Len(t, spanExporter.GetSpans(), 3)
}

func TestExportToOtelCollector(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping integration test")
	}

	featureEvents := []exporter.FeatureEvent{
		{
			Kind: "feature1", ContextKind: "anonymousUser1", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
			Variation: "Default", Value: "YO", Default: false,
		},
		{
			Kind: "feature2", ContextKind: "anonymousUser2", UserKey: "ABCDEF", CreationDate: 1617970547, Key: "random-key",
			Variation: "Default", Value: "YO", Default: false,
		},
	}

	ctx := context.Background()
	logger := log.New(os.Stdout, "", 0)

	consumer := SliceLogConsumer{}
	otelC, err := setupotelCollectorContainer(ctx, &consumer)
	if err != nil {
		t.Fatal(err)
	}

	otelProcessor, err := otelCollectorBatchSpanProcessor(otelC.URI)
	assert.Nil(t, err)
	resource := Resource()
	batchOptions := sdktrace.BatchSpanProcessorOptions{MaxQueueSize: 10, MaxExportBatchSize: 100, BatchTimeout: time.Millisecond * 100}
	processors := make([]*sdktrace.SpanProcessor, 0)
	processors = append(processors, &otelProcessor)
	exp := NewExporter(
		WithBatchSpanProcessorOption(batchOptions),
		WithResource(resource),
		WithBatchSpanProcessors(processors),
	)
	exp.Export(ctx, logger, featureEvents)

	time.Sleep(2 * time.Second)
	assert.GreaterOrEqual(t, len(consumer.logs), 1)
	assert.True(t, consumer.Exists("feature1"))
	assert.True(t, consumer.Exists("feature2"))

	//Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := otelC.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

}

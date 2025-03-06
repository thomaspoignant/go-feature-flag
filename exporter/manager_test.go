package exporter_test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thejerf/slogassert"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/exporter"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
	"github.com/thomaspoignant/go-feature-flag/testutils/mock"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

func TestDataExporterManager_flushWithTime(t *testing.T) {
	mockExporter := mock.Exporter{Bulk: true}
	dataExporterMock := []exporter.Config{
		{
			FlushInterval:    10 * time.Millisecond,
			MaxEventInMemory: 1000,
			Exporter:         &mockExporter,
		},
	}
	dc := exporter.NewManager[exporter.FeatureEvent](context.Background(), dataExporterMock, nil)
	go dc.StartDaemon()
	defer dc.Close()

	// Initialize inputEvents slice
	inputEvents := []exporter.FeatureEvent{
		exporter.NewFeatureEvent(
			ffcontext.NewEvaluationContextBuilder("ABCD").AddCustom("anonymous", true).Build(), "random-key",
			"YO", "defaultVar", false, "", "SERVER", nil),
	}

	for _, event := range inputEvents {
		dc.AddEvent(event)
	}

	time.Sleep(500 * time.Millisecond)
	assert.Equal(t, inputEvents, mockExporter.GetExportedEvents())
}

func TestDataExporterManager_flushWithNumberOfEvents(t *testing.T) {
	mockExporter := mock.Exporter{Bulk: true}
	dataExporterMock := []exporter.Config{
		{
			FlushInterval:    10 * time.Millisecond,
			MaxEventInMemory: 100,
			Exporter:         &mockExporter,
		},
	}
	dc := exporter.NewManager[exporter.FeatureEvent](context.Background(), dataExporterMock, nil)
	go dc.StartDaemon()
	defer dc.Close()

	// Initialize inputEvents slice
	var inputEvents []exporter.FeatureEvent
	for i := 0; i <= 100; i++ {
		inputEvents = append(inputEvents, exporter.NewFeatureEvent(
			ffcontext.NewEvaluationContextBuilder("ABCD").AddCustom("anonymous", true).Build(),
			"random-key", "YO", "defaultVar", false, "", "SERVER", nil))
	}
	for _, event := range inputEvents {
		dc.AddEvent(event)
	}
	assert.Equal(t, inputEvents[:100], mockExporter.GetExportedEvents())
}

func TestDataExporterManager_defaultFlush(t *testing.T) {
	mockExporter := mock.Exporter{Bulk: true}
	dataExporterMock := []exporter.Config{
		{
			FlushInterval:    0,
			MaxEventInMemory: 0,
			Exporter:         &mockExporter,
		},
	}
	dc := exporter.NewManager[exporter.FeatureEvent](context.Background(), dataExporterMock, nil)
	go dc.StartDaemon()
	defer dc.Close()

	// Initialize inputEvents slice
	var inputEvents []exporter.FeatureEvent
	for i := 0; i <= 100000; i++ {
		inputEvents = append(inputEvents, exporter.NewFeatureEvent(
			ffcontext.NewEvaluationContextBuilder("ABCD").AddCustom("anonymous", true).Build(),
			"random-key", "YO", "defaultVar", false, "", "SERVER", nil))
	}
	for _, event := range inputEvents {
		dc.AddEvent(event)
	}
	assert.Equal(t, inputEvents[:100000], mockExporter.GetExportedEvents())
}

func TestDataExporterManager_exporterReturnError(t *testing.T) {
	mockExporter := mock.Exporter{Err: errors.New("random err"), ExpectedNumberErr: 1, Bulk: true}
	dataExporterMock := []exporter.Config{
		{
			FlushInterval:    10 * time.Minute,
			MaxEventInMemory: 100,
			Exporter:         &mockExporter,
		},
	}
	file, _ := os.CreateTemp("", "log")
	defer file.Close()
	defer os.Remove(file.Name())
	handler := slogassert.New(t, slog.LevelInfo, nil)
	logger := slog.New(handler)
	dc := exporter.NewManager[exporter.FeatureEvent](context.Background(), dataExporterMock, &fflog.FFLogger{LeveledLogger: logger})
	go dc.StartDaemon()
	defer dc.Close()

	// Initialize inputEvents slice
	var inputEvents []exporter.FeatureEvent
	for i := 0; i <= 200; i++ {
		inputEvents = append(inputEvents, exporter.NewFeatureEvent(
			ffcontext.NewEvaluationContextBuilder("ABCD").AddCustom("anonymous", true).Build(),
			"random-key", "YO", "defaultVar", false, "", "SERVER", nil))
	}
	for _, event := range inputEvents {
		dc.AddEvent(event)
	}
	// check that the first 100 events are exported
	assert.Equal(t, inputEvents[:100], mockExporter.GetExportedEvents()[:100])
	handler.AssertMessage("error while exporting data: random err")
}

func TestDataExporterManager_nonBulkExporter(t *testing.T) {
	mockExporter := mock.Exporter{Bulk: false}
	dataExporterMock := []exporter.Config{
		{
			FlushInterval:    0,
			MaxEventInMemory: 0,
			Exporter:         &mockExporter,
		},
	}
	dc := exporter.NewManager[exporter.FeatureEvent](context.Background(), dataExporterMock, nil)
	defer dc.Close()

	// Initialize inputEvents slice
	var inputEvents []exporter.FeatureEvent
	for i := 0; i < 100; i++ {
		inputEvents = append(inputEvents, exporter.NewFeatureEvent(
			ffcontext.NewEvaluationContextBuilder("ABCD").AddCustom("anonymous", true).Build(),
			"random-key", "YO", "defaultVar", false, "", "SERVER", nil))
	}
	for _, event := range inputEvents {
		dc.AddEvent(event)
		// we have to wait because we are opening a new thread to slow down the flag evaluation.
		time.Sleep(1 * time.Millisecond)
	}

	assert.Equal(t, inputEvents[:100], mockExporter.GetExportedEvents())
}

func TestAddExporterMetadataFromContextToExporter(t *testing.T) {
	tests := []struct {
		name string
		ctx  ffcontext.EvaluationContext
		want map[string]interface{}
	}{
		{
			name: "extract exporter metadata from context",
			ctx: ffcontext.NewEvaluationContextBuilder("targeting-key").AddCustom("gofeatureflag", map[string]interface{}{
				"exporterMetadata": map[string]interface{}{
					"key1": "value1",
					"key2": 123,
					"key3": true,
					"key4": 123.45,
				},
			}).Build(),
			want: map[string]interface{}{
				"key1": "value1",
				"key2": 123,
				"key3": true,
				"key4": 123.45,
			},
		},
		{
			name: "no exporter metadata in the context",
			ctx:  ffcontext.NewEvaluationContextBuilder("targeting-key").Build(),
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExporter := &mock.Exporter{}
			config := ffclient.Config{
				Retriever: &fileretriever.Retriever{Path: "../testdata/flag-config.yaml"},
				DataExporter: ffclient.DataExporter{
					Exporter:      mockExporter,
					FlushInterval: 100 * time.Millisecond,
				},
			}
			goff, err := ffclient.New(config)
			assert.NoError(t, err)

			_, err = goff.BoolVariation("test-flag", tt.ctx, false)
			assert.NoError(t, err)

			time.Sleep(120 * time.Millisecond)
			assert.Equal(t, 1, len(mockExporter.GetExportedEvents()))
			got := mockExporter.GetExportedEvents()[0].Metadata
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDataExporterManager_multipleExporters(t *testing.T) {
	mockExporter1 := mock.Exporter{Bulk: false}
	mockExporter2 := mock.Exporter{Bulk: true}
	dataExporterMock := []exporter.Config{
		{
			FlushInterval:    0,
			MaxEventInMemory: 0,
			Exporter:         &mockExporter1,
		},
		{
			FlushInterval:    200 * time.Millisecond,
			MaxEventInMemory: 200,
			Exporter:         &mockExporter2,
		},
	}
	dc := exporter.NewManager[exporter.FeatureEvent](context.Background(), dataExporterMock, nil)
	go dc.StartDaemon()
	defer dc.Close()

	// Initialize inputEvents slice
	var inputEvents []exporter.FeatureEvent
	for i := 0; i < 100; i++ {
		inputEvents = append(inputEvents, exporter.NewFeatureEvent(
			ffcontext.NewEvaluationContextBuilder("ABCD").AddCustom("anonymous", true).Build(),
			"random-key", "YO", "defaultVar", false, "", "SERVER", nil))
	}
	for _, event := range inputEvents {
		dc.AddEvent(event)
		// we have to wait because we are opening a new thread to slow down the flag evaluation.
		time.Sleep(1 * time.Millisecond)
	}

	assert.Equal(t, inputEvents[:100], mockExporter1.GetExportedEvents())
	assert.Equal(t, 0, len(mockExporter2.GetExportedEvents()))
	time.Sleep(250 * time.Millisecond)
	assert.Equal(t, inputEvents[:100], mockExporter2.GetExportedEvents())
}

func TestDataExporterManager_multipleExportersWithDifferentFlushInterval(t *testing.T) {
	mockExporter1 := mock.Exporter{Bulk: true}
	mockExporter2 := mock.Exporter{Bulk: true}
	dataExporterMock := []exporter.Config{
		{
			FlushInterval:    50 * time.Millisecond,
			MaxEventInMemory: 0,
			Exporter:         &mockExporter1,
		},
		{
			FlushInterval:    0 * time.Millisecond,
			MaxEventInMemory: 100,
			Exporter:         &mockExporter2,
		},
	}
	dc := exporter.NewManager[exporter.FeatureEvent](context.Background(), dataExporterMock, nil)
	go dc.StartDaemon()
	defer dc.Close()

	// Initialize inputEvents slice
	var inputEvents []exporter.FeatureEvent
	for i := 0; i < 100; i++ {
		inputEvents = append(inputEvents, exporter.NewFeatureEvent(
			ffcontext.NewEvaluationContextBuilder("ABCD").AddCustom("anonymous", true).Build(),
			"random-key", "YO", "defaultVar", false, "", "SERVER", nil))
	}
	go func(dc exporter.Manager[exporter.FeatureEvent]) {
		for _, event := range inputEvents {
			dc.AddEvent(event)
			// we have to wait because we are opening a new thread to slow down the flag evaluation.
			time.Sleep(1 * time.Millisecond)
		}
	}(dc)

	assert.Equal(t, 0, len(mockExporter2.GetExportedEvents()))
	assert.Equal(t, 0, len(mockExporter1.GetExportedEvents()))
	time.Sleep(70 * time.Millisecond)
	assert.True(t, len(mockExporter1.GetExportedEvents()) > 0)
	assert.True(t, len(mockExporter2.GetExportedEvents()) == 0)
	time.Sleep(200 * time.Millisecond)
	assert.True(t, len(mockExporter1.GetExportedEvents()) > 0)
	assert.True(t, len(mockExporter2.GetExportedEvents()) > 0)
}

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

func TestDataExporterScheduler_flushWithTime(t *testing.T) {
	mockExporter := mock.Exporter{Bulk: true}
	dc := exporter.NewScheduler(
		context.Background(), 10*time.Millisecond, 1000, &mockExporter, nil)
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

func TestDataExporterScheduler_flushWithNumberOfEvents(t *testing.T) {
	mockExporter := mock.Exporter{Bulk: true}
	dc := exporter.NewScheduler(
		context.Background(), 10*time.Minute, 100, &mockExporter, nil)
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

func TestDataExporterScheduler_defaultFlush(t *testing.T) {
	mockExporter := mock.Exporter{Bulk: true}
	dc := exporter.NewScheduler(
		context.Background(), 0, 0, &mockExporter, nil)
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

func TestDataExporterScheduler_exporterReturnError(t *testing.T) {
	mockExporter := mock.Exporter{Err: errors.New("random err"), ExpectedNumberErr: 1, Bulk: true}

	file, _ := os.CreateTemp("", "log")
	defer file.Close()
	defer os.Remove(file.Name())
	handler := slogassert.New(t, slog.LevelInfo, nil)
	logger := slog.New(handler)

	dc := exporter.NewScheduler(
		context.Background(), 0, 100, &mockExporter, &fflog.FFLogger{LeveledLogger: logger})
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
	assert.Equal(t, inputEvents[:201], mockExporter.GetExportedEvents())
	handler.AssertMessage("error while exporting data: random err")
}

func TestDataExporterScheduler_nonBulkExporter(t *testing.T) {
	mockExporter := mock.Exporter{Bulk: false}
	dc := exporter.NewScheduler(
		context.Background(), 0, 0, &mockExporter, nil)
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

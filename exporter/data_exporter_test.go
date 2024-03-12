package exporter_test

import (
	"context"
	"errors"
	"log"
	"os"
	"testing"
	"time"

	"github.com/thomaspoignant/go-feature-flag/exporter"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/testutils/mock"
)

func TestDataExporterScheduler_flushWithTime(t *testing.T) {
	mockExporter := mock.Exporter{Bulk: true}
	dc := exporter.NewScheduler(
		context.Background(), 10*time.Millisecond, 1000, &mockExporter, log.New(os.Stdout, "", 0))
	go dc.StartDaemon()
	defer dc.Close()

	// Initialize inputEvents slice
	inputEvents := []exporter.FeatureEvent{
		exporter.NewFeatureEvent(
			ffcontext.NewEvaluationContextBuilder("ABCD").AddCustom("anonymous", true).Build(), "random-key",
			"YO", "defaultVar", false, "", "SERVER"),
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
		context.Background(), 10*time.Minute, 100, &mockExporter, log.New(os.Stdout, "", 0))
	go dc.StartDaemon()
	defer dc.Close()

	// Initialize inputEvents slice
	var inputEvents []exporter.FeatureEvent
	for i := 0; i <= 100; i++ {
		inputEvents = append(inputEvents, exporter.NewFeatureEvent(
			ffcontext.NewEvaluationContextBuilder("ABCD").AddCustom("anonymous", true).Build(),
			"random-key", "YO", "defaultVar", false, "", "SERVER"))
	}
	for _, event := range inputEvents {
		dc.AddEvent(event)
	}
	assert.Equal(t, inputEvents[:100], mockExporter.GetExportedEvents())
}

func TestDataExporterScheduler_defaultFlush(t *testing.T) {
	mockExporter := mock.Exporter{Bulk: true}
	dc := exporter.NewScheduler(
		context.Background(), 0, 0, &mockExporter, log.New(os.Stdout, "", 0))
	go dc.StartDaemon()
	defer dc.Close()

	// Initialize inputEvents slice
	var inputEvents []exporter.FeatureEvent
	for i := 0; i <= 100000; i++ {
		inputEvents = append(inputEvents, exporter.NewFeatureEvent(
			ffcontext.NewEvaluationContextBuilder("ABCD").AddCustom("anonymous", true).Build(),
			"random-key", "YO", "defaultVar", false, "", "SERVER"))
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
	logger := log.New(file, "", 0)

	dc := exporter.NewScheduler(
		context.Background(), 0, 100, &mockExporter, logger)
	go dc.StartDaemon()
	defer dc.Close()

	// Initialize inputEvents slice
	var inputEvents []exporter.FeatureEvent
	for i := 0; i <= 200; i++ {
		inputEvents = append(inputEvents, exporter.NewFeatureEvent(
			ffcontext.NewEvaluationContextBuilder("ABCD").AddCustom("anonymous", true).Build(),
			"random-key", "YO", "defaultVar", false, "", "SERVER"))
	}
	for _, event := range inputEvents {
		dc.AddEvent(event)
	}
	assert.Equal(t, inputEvents[:201], mockExporter.GetExportedEvents())

	// read log
	logs, _ := os.ReadFile(file.Name())
	assert.Regexp(t, "\\["+testutils.RFC3339Regex+"\\] error while exporting data: random err\n", string(logs))
}

func TestDataExporterScheduler_nonBulkExporter(t *testing.T) {
	mockExporter := mock.Exporter{Bulk: false}
	dc := exporter.NewScheduler(
		context.Background(), 0, 0, &mockExporter, log.New(os.Stdout, "", 0))
	defer dc.Close()

	// Initialize inputEvents slice
	var inputEvents []exporter.FeatureEvent
	for i := 0; i < 100; i++ {
		inputEvents = append(inputEvents, exporter.NewFeatureEvent(
			ffcontext.NewEvaluationContextBuilder("ABCD").AddCustom("anonymous", true).Build(),
			"random-key", "YO", "defaultVar", false, "", "SERVER"))
	}
	for _, event := range inputEvents {
		dc.AddEvent(event)
		// we have to wait because we are opening a new thread to slow down the flag evaluation.
		time.Sleep(1 * time.Millisecond)
	}

	assert.Equal(t, inputEvents[:100], mockExporter.GetExportedEvents())
}

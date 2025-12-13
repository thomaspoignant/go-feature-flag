package exporter_test

import (
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/exporter"
	"github.com/thomaspoignant/go-feature-flag/testutils"
	"github.com/thomaspoignant/go-feature-flag/testutils/mock"
	"github.com/thomaspoignant/go-feature-flag/testutils/slogutil"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

func TestDataExporterFlush_TriggerError(t *testing.T) {
	evStore := mock.NewEventStore[exporter.FeatureEvent]()
	for i := 0; i < 100; i++ {
		evStore.Add(exporter.FeatureEvent{
			Kind: "feature",
		})
	}

	logFile, _ := os.CreateTemp("", "")
	textHandler := slogutil.MessageOnlyHandler{Writer: logFile}
	logger := &fflog.FFLogger{LeveledLogger: slog.New(&textHandler)}
	defer func() { _ = os.Remove(logFile.Name()) }()

	exporterMock := mock.Exporter{}
	exp := exporter.NewDataExporter[exporter.FeatureEvent](exporter.Config{
		Exporter:         &exporterMock,
		FlushInterval:    0,
		MaxEventInMemory: 0,
	}, "error", &evStore, logger)

	exp.Flush()
	// flush should error and not return any event
	assert.Equal(t, 0, len(exporterMock.GetExportedEvents()))
	logContent, _ := os.ReadFile(logFile.Name())
	assert.Equal(t, "error\n", string(logContent))
}

func TestDataExporterFlush_TriggerErrorIfNotKnowType(t *testing.T) {
	tests := []struct {
		name        string
		exporter    mock.ExporterMock
		expectedLog string
	}{
		{
			name:        "deprecated exporter",
			exporter:    &mock.ExporterDeprecated{},
			expectedLog: "trying to send unknown object to the exporter (deprecated)\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evStore := mock.NewEventStore[testutils.ExportableMockEvent]()
			for i := 0; i < 100; i++ {
				evStore.Add(testutils.NewExportableMockEvent("feature"))
			}

			logFile, _ := os.CreateTemp("", "")
			textHandler := slogutil.MessageOnlyHandler{Writer: logFile}
			logger := &fflog.FFLogger{LeveledLogger: slog.New(&textHandler)}
			defer func() { _ = os.Remove(logFile.Name()) }()

			exporterMock := tt.exporter
			exp := exporter.NewDataExporter[testutils.ExportableMockEvent](
				exporter.Config{
					Exporter:         exporterMock,
					FlushInterval:    0,
					MaxEventInMemory: 0,
				},
				"id-consumer",
				&evStore,
				logger,
			)

			exp.Flush()
			// flush should error and not return any event
			assert.Equal(t, 0, len(exporterMock.GetExportedEvents()))
			logContent, _ := os.ReadFile(logFile.Name())
			assert.Equal(t, tt.expectedLog, string(logContent))
		})
	}
}

func TestDataExporterFlush_TriggerErrorIfExporterFail(t *testing.T) {
	tests := []struct {
		name        string
		exporter    mock.ExporterMock
		expectedLog string
	}{
		{
			name:        "classic exporter",
			exporter:    &mock.Exporter{Err: fmt.Errorf("error"), ExpectedNumberErr: 1},
			expectedLog: "error while exporting data: error\n",
		},
		{
			name:        "deprecated exporter",
			exporter:    &mock.ExporterDeprecated{Err: fmt.Errorf("error"), ExpectedNumberErr: 1},
			expectedLog: "error while exporting data (deprecated): error\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evStore := mock.NewEventStore[exporter.FeatureEvent]()
			for i := 0; i < 100; i++ {
				evStore.Add(exporter.FeatureEvent{Kind: "feature"})
			}

			logFile, _ := os.CreateTemp("", "")
			textHandler := slogutil.MessageOnlyHandler{Writer: logFile}
			logger := &fflog.FFLogger{LeveledLogger: slog.New(&textHandler)}
			defer func() { _ = os.Remove(logFile.Name()) }()

			exporterMock := tt.exporter
			exp := exporter.NewDataExporter[exporter.FeatureEvent](
				exporter.Config{
					Exporter:         exporterMock,
					FlushInterval:    0,
					MaxEventInMemory: 0,
				}, "id-consumer", &evStore, logger)

			exp.Flush()
			// flush should error and not return any event
			assert.Equal(t, 100, len(exporterMock.GetExportedEvents()))
			logContent, _ := os.ReadFile(logFile.Name())
			assert.Equal(t, tt.expectedLog, string(logContent))
		})
	}
}

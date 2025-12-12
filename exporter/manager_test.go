package exporter_test

import (
	"context"
	"errors"
	"fmt"
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
	tests := []struct {
		name         string
		mockExporter mock.ExporterMock
	}{
		{
			name:         "flushTime: classic exporter",
			mockExporter: &mock.Exporter{Bulk: true},
		},
		{
			name:         "flushTime: deprecated exporter",
			mockExporter: &mock.ExporterDeprecated{Bulk: true},
		},
		{
			name:         "flushTime: deprecated exporter v2",
			mockExporter: &mock.ExporterDeprecatedV2{Bulk: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dataExporterMock := []exporter.Config{
				{
					FlushInterval:    10 * time.Millisecond,
					MaxEventInMemory: 1000,
					Exporter:         tt.mockExporter,
				},
			}
			dc := exporter.NewManager[exporter.FeatureEvent](
				dataExporterMock,
				exporter.DefaultExporterCleanQueueInterval,
				nil,
			)
			go dc.Start()
			defer dc.Stop()

			// Initialize inputEvents slice
			inputEvents := []exporter.FeatureEvent{
				exporter.NewFeatureEvent(
					ffcontext.NewEvaluationContextBuilder("ABCD").
						AddCustom("anonymous", true).
						Build(),
					"random-key",
					"YO",
					"defaultVar",
					false,
					"",
					"SERVER",
					nil,
				),
			}

			want := make([]exporter.ExportableEvent, len(inputEvents))
			for i, event := range inputEvents {
				dc.AddEvent(event)
				want[i] = event
			}

			time.Sleep(500 * time.Millisecond)
			assert.Equal(t, want, tt.mockExporter.GetExportedEvents())
		})
	}
}

func TestDataExporterManager_flushWithNumberOfEvents(t *testing.T) {
	tests := []struct {
		name         string
		mockExporter mock.ExporterMock
	}{
		{
			name:         "flushWithNumberOfEvents: classic exporter",
			mockExporter: &mock.Exporter{Bulk: true},
		},
		{
			name:         "flushWithNumberOfEvents: deprecated exporter",
			mockExporter: &mock.ExporterDeprecated{Bulk: true},
		},
		{
			name:         "flushWithNumberOfEvents: deprecated exporter v2",
			mockExporter: &mock.ExporterDeprecatedV2{Bulk: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dataExporterMock := []exporter.Config{
				{
					FlushInterval:    10 * time.Millisecond,
					MaxEventInMemory: 100,
					Exporter:         tt.mockExporter,
				},
			}
			dc := exporter.NewManager[exporter.FeatureEvent](
				dataExporterMock,
				exporter.DefaultExporterCleanQueueInterval,
				nil,
			)
			go dc.Start()
			defer dc.Stop()

			// Initialize inputEvents slice
			var inputEvents []exporter.FeatureEvent
			for i := 0; i <= 100; i++ {
				inputEvents = append(inputEvents, exporter.NewFeatureEvent(
					ffcontext.NewEvaluationContextBuilder("ABCD").
						AddCustom("anonymous", true).
						Build(),
					"random-key",
					"YO",
					"defaultVar",
					false,
					"",
					"SERVER",
					nil,
				))
			}
			want := make([]exporter.ExportableEvent, len(inputEvents))
			for i, event := range inputEvents {
				dc.AddEvent(event)
				want[i] = event
			}
			assert.Equal(t, want[:100], tt.mockExporter.GetExportedEvents())
		})
	}
}

func TestDataExporterManager_defaultFlush(t *testing.T) {
	tests := []struct {
		name         string
		mockExporter mock.ExporterMock
	}{
		{
			name:         "classic exporter",
			mockExporter: &mock.Exporter{Bulk: true},
		},
		{
			name:         "deprecated exporter",
			mockExporter: &mock.ExporterDeprecated{Bulk: true},
		},
		{
			name:         "deprecated exporter v2",
			mockExporter: &mock.ExporterDeprecatedV2{Bulk: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dataExporterMock := []exporter.Config{
				{
					FlushInterval:    0,
					MaxEventInMemory: 0,
					Exporter:         tt.mockExporter,
				},
			}
			dc := exporter.NewManager[exporter.FeatureEvent](
				dataExporterMock, exporter.DefaultExporterCleanQueueInterval, nil)
			go dc.Start()
			defer dc.Stop()

			// Initialize inputEvents slice
			var inputEvents []exporter.FeatureEvent
			for i := 0; i <= 100000; i++ {
				inputEvents = append(inputEvents, exporter.NewFeatureEvent(
					ffcontext.NewEvaluationContextBuilder("ABCD").
						AddCustom("anonymous", true).
						Build(),
					"random-key",
					"YO",
					"defaultVar",
					false,
					"",
					"SERVER",
					nil,
				))
			}
			want := make([]exporter.ExportableEvent, len(inputEvents))
			for i, event := range inputEvents {
				dc.AddEvent(event)
				want[i] = event
			}
			assert.Equal(t, want[:100000], tt.mockExporter.GetExportedEvents())
		})
	}
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
	defer func() { _ = file.Close() }()
	defer func() { _ = os.Remove(file.Name()) }()
	handler := slogassert.New(t, slog.LevelInfo, nil)
	logger := slog.New(handler)
	dc := exporter.NewManager[exporter.FeatureEvent](dataExporterMock,
		exporter.DefaultExporterCleanQueueInterval, &fflog.FFLogger{LeveledLogger: logger})
	go dc.Start()
	defer dc.Stop()

	// Initialize inputEvents slice
	var inputEvents []exporter.FeatureEvent
	for i := 0; i <= 200; i++ {
		inputEvents = append(inputEvents, exporter.NewFeatureEvent(
			ffcontext.NewEvaluationContextBuilder("ABCD").AddCustom("anonymous", true).Build(),
			"random-key", "YO", "defaultVar", false, "", "SERVER", nil))
	}
	want := make([]exporter.ExportableEvent, len(inputEvents))
	for i, event := range inputEvents {
		dc.AddEvent(event)
		want[i] = event
	}
	// check that the first 100 events are exported
	assert.Equal(t, want[:100], mockExporter.GetExportedEvents()[:100])
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
	dc := exporter.NewManager[exporter.FeatureEvent](
		dataExporterMock, exporter.DefaultExporterCleanQueueInterval, nil)
	defer dc.Stop()

	// Initialize inputEvents slice
	var inputEvents []exporter.FeatureEvent
	for i := 0; i < 100; i++ {
		inputEvents = append(inputEvents, exporter.NewFeatureEvent(
			ffcontext.NewEvaluationContextBuilder("ABCD").AddCustom("anonymous", true).Build(),
			"random-key", "YO", "defaultVar", false, "", "SERVER", nil))
	}
	want := make([]exporter.ExportableEvent, len(inputEvents))
	for i, event := range inputEvents {
		dc.AddEvent(event)
		want[i] = event
		// we have to wait because we are opening a new thread to slow down the flag evaluation.
		time.Sleep(1 * time.Millisecond)
	}

	assert.Equal(t, want[:100], mockExporter.GetExportedEvents())
}

func TestAddExporterMetadataFromContextToExporter(t *testing.T) {
	tests := []struct {
		name string
		ctx  ffcontext.EvaluationContext
		want map[string]interface{}
	}{
		{
			name: "extract exporter metadata from context",
			ctx: ffcontext.NewEvaluationContextBuilder("targeting-key").
				AddCustom("gofeatureflag", map[string]interface{}{
					"exporterMetadata": map[string]interface{}{
						"key1": "value1",
						"key2": 123,
						"key3": true,
						"key4": 123.45,
					},
				}).
				Build(),
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

			switch val := mockExporter.GetExportedEvents()[0].(type) {
			case exporter.FeatureEvent:
				assert.Equal(t, tt.want, val.Metadata)
				break
			default:
				assert.Fail(t, "The exported event is not a FeatureEvent")
			}
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
	dc := exporter.NewManager[exporter.FeatureEvent](
		dataExporterMock, exporter.DefaultExporterCleanQueueInterval, nil)
	go dc.Start()
	defer dc.Stop()

	// Initialize inputEvents slice
	var inputEvents []exporter.FeatureEvent
	for i := 0; i < 100; i++ {
		inputEvents = append(inputEvents, exporter.NewFeatureEvent(
			ffcontext.NewEvaluationContextBuilder("ABCD").AddCustom("anonymous", true).Build(),
			"random-key", "YO", "defaultVar", false, "", "SERVER", nil))
	}
	want := make([]exporter.ExportableEvent, len(inputEvents))
	for i, event := range inputEvents {
		dc.AddEvent(event)
		want[i] = event
		// we have to wait because we are opening a new thread to slow down the flag evaluation.
		time.Sleep(1 * time.Millisecond)
	}

	assert.Equal(t, want[:100], mockExporter1.GetExportedEvents())
	assert.Equal(t, 0, len(mockExporter2.GetExportedEvents()))
	time.Sleep(250 * time.Millisecond)
	assert.Equal(t, want[:100], mockExporter2.GetExportedEvents())
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
	dc := exporter.NewManager[exporter.FeatureEvent](
		dataExporterMock, exporter.DefaultExporterCleanQueueInterval, nil)
	go dc.Start()
	defer dc.Stop()

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

func TestDataExporterManager_ValidateNumberOfEvents(t *testing.T) {
	// We are running the same test multiple times to have more chance to have a race condition.
	for i := 1; i < 20; i++ {
		t.Run(fmt.Sprintf("ValidateNumberOfEvents #%d", i), func(t *testing.T) {
			mockExporter := mock.Exporter{Bulk: true}
			// Init ffclient with a file retriever.
			err := ffclient.Init(ffclient.Config{
				PollingInterval: 10 * time.Second,
				LeveledLogger:   slog.Default(),
				Context:         context.Background(),
				Retriever: &fileretriever.Retriever{
					Path: "../testdata/flag-config.yaml",
				},
				DataExporter: ffclient.DataExporter{
					FlushInterval:    150 * time.Millisecond,
					MaxEventInMemory: 100,
					Exporter:         &mockExporter,
				},
			})
			assert.NoError(t, err)

			// create users
			user1 := ffcontext.
				NewEvaluationContextBuilder("aea2fdc1-b9a0-417a-b707-0c9083de68e3").
				AddCustom("anonymous", true).
				Build()
			user2 := ffcontext.NewEvaluationContext("332460b9-a8aa-4f7a-bc5d-9cc33632df9a")

			_, _ = ffclient.BoolVariation("test-flag", user1, false)
			_, _ = ffclient.BoolVariation("test-flag", user2, false)
			_, _ = ffclient.StringVariation("test-flag2", user1, "defaultValue")
			_, _ = ffclient.JSONVariation(
				"test-flag2",
				user1,
				map[string]interface{}{"test": "toto"},
			)
			time.Sleep(300 * time.Millisecond)
			assert.Equal(t, 4, len(mockExporter.GetExportedEvents()))

			// Wait 2 seconds to have a second file
			_, _ = ffclient.BoolVariation("test-flag", user1, false)
			_, _ = ffclient.BoolVariation("test-flag", user2, false)
			ffclient.Close() // a flush is triggered here
			assert.Equal(t, 6, len(mockExporter.GetExportedEvents()))
		})
	}
}

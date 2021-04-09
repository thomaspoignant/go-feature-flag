package exporter_test

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/thomaspoignant/go-feature-flag/ffuser"
	"github.com/thomaspoignant/go-feature-flag/internal/exporter"
	"github.com/thomaspoignant/go-feature-flag/internal/model"
	"github.com/thomaspoignant/go-feature-flag/testutil"
)

type mockExporter struct {
	exportedEvents    []exporter.FeatureEvent
	err               error
	expectedNumberErr int
	currentNumberErr  int
	mutex             sync.Mutex
}

func (m *mockExporter) Export(logger *log.Logger, events []exporter.FeatureEvent) error {
	m.mutex.Lock()
	m.exportedEvents = append(m.exportedEvents, events...)
	m.mutex.Unlock()

	if m.err != nil {
		if m.expectedNumberErr > m.currentNumberErr {
			m.currentNumberErr++
			return m.err
		}
	}
	return nil
}

func (m *mockExporter) getExportedEvents() []exporter.FeatureEvent {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.exportedEvents
}

func TestDataExporterScheduler_flushWithTime(t *testing.T) {
	mockExporter := mockExporter{mutex: sync.Mutex{}}
	dc := exporter.NewDataExporterScheduler(
		10*time.Millisecond, 1000, &mockExporter, log.New(os.Stdout, "", 0))
	go dc.StartDaemon()
	defer dc.Close()

	inputEvents := []exporter.FeatureEvent{
		exporter.NewFeatureEvent(ffuser.NewAnonymousUser("ABCD"), "random-key",
			model.Flag{Percentage: 100}, "YO", model.VariationDefault, false),
	}

	for _, event := range inputEvents {
		dc.AddEvent(event)
	}

	time.Sleep(10 * time.Millisecond * 2)
	assert.Equal(t, inputEvents, mockExporter.getExportedEvents())
}

func TestDataExporterScheduler_flushWithNumberOfEvents(t *testing.T) {
	mockExporter := mockExporter{mutex: sync.Mutex{}}
	dc := exporter.NewDataExporterScheduler(
		10*time.Minute, 100, &mockExporter, log.New(os.Stdout, "", 0))
	go dc.StartDaemon()
	defer dc.Close()

	var inputEvents []exporter.FeatureEvent
	for i := 0; i <= 100; i++ {
		inputEvents = append(inputEvents, exporter.NewFeatureEvent(ffuser.NewAnonymousUser("ABCD"),
			"random-key", model.Flag{Percentage: 100}, "YO", model.VariationDefault, false))
	}
	for _, event := range inputEvents {
		dc.AddEvent(event)
	}
	assert.Equal(t, inputEvents[:100], mockExporter.getExportedEvents())
}

func TestDataExporterScheduler_defaultFlush(t *testing.T) {
	mockExporter := mockExporter{mutex: sync.Mutex{}}
	dc := exporter.NewDataExporterScheduler(
		0, 0, &mockExporter, log.New(os.Stdout, "", 0))
	go dc.StartDaemon()
	defer dc.Close()

	var inputEvents []exporter.FeatureEvent
	for i := 0; i <= 100000; i++ {
		inputEvents = append(inputEvents, exporter.NewFeatureEvent(ffuser.NewAnonymousUser("ABCD"),
			"random-key", model.Flag{Percentage: 100}, "YO", model.VariationDefault, false))
	}
	for _, event := range inputEvents {
		dc.AddEvent(event)
	}
	assert.Equal(t, inputEvents[:100000], mockExporter.getExportedEvents())
}

func TestDataExporterScheduler_exporterReturnError(t *testing.T) {
	mockExporter := mockExporter{err: errors.New("random err"), expectedNumberErr: 1, mutex: sync.Mutex{}}

	file, _ := ioutil.TempFile("", "log")
	defer file.Close()
	defer os.Remove(file.Name())
	logger := log.New(file, "", 0)

	dc := exporter.NewDataExporterScheduler(
		0, 100, &mockExporter, logger)
	go dc.StartDaemon()
	defer dc.Close()

	var inputEvents []exporter.FeatureEvent
	for i := 0; i <= 200; i++ {
		inputEvents = append(inputEvents, exporter.NewFeatureEvent(ffuser.NewAnonymousUser("ABCD"),
			"random-key", model.Flag{Percentage: 100}, "YO", model.VariationDefault, false))
	}
	for _, event := range inputEvents {
		dc.AddEvent(event)
	}
	assert.Equal(t, inputEvents[:201], mockExporter.getExportedEvents())

	// read log
	logs, _ := ioutil.ReadFile(file.Name())
	assert.Regexp(t, "\\["+testutil.RFC3339Regex+"\\] error while exporting data: random err\\n", string(logs))
}

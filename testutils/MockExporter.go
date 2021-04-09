package testutils

import (
	"log"
	"sync"

	"github.com/thomaspoignant/go-feature-flag/internal/exporter"
)

type MockExporter struct {
	ExportedEvents    []exporter.FeatureEvent
	Err               error
	ExpectedNumberErr int
	CurrentNumberErr  int
	Mutex             sync.Mutex
}

func (m *MockExporter) Export(logger *log.Logger, events []exporter.FeatureEvent) error {
	m.Mutex.Lock()
	m.ExportedEvents = append(m.ExportedEvents, events...)
	m.Mutex.Unlock()

	if m.Err != nil {
		if m.ExpectedNumberErr > m.CurrentNumberErr {
			m.CurrentNumberErr++
			return m.Err
		}
	}
	return nil
}

func (m *MockExporter) GetExportedEvents() []exporter.FeatureEvent {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	return m.ExportedEvents
}

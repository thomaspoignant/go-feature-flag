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
	Bulk              bool

	mutex sync.Mutex
	once  sync.Once
}

func (m *MockExporter) Export(logger *log.Logger, events []exporter.FeatureEvent) error {
	m.once.Do(m.initMutex)
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.ExportedEvents = append(m.ExportedEvents, events...)
	if m.Err != nil {
		if m.ExpectedNumberErr > m.CurrentNumberErr {
			m.CurrentNumberErr++
			return m.Err
		}
	}
	return nil
}

func (m *MockExporter) GetExportedEvents() []exporter.FeatureEvent {
	m.once.Do(m.initMutex)
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.ExportedEvents
}

func (m *MockExporter) IsBulk() bool {
	return m.Bulk
}

func (m *MockExporter) initMutex() {
	m.mutex = sync.Mutex{}
}

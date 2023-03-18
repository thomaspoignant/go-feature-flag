package mock

import (
	"context"
	"log"
	"sync"

	"github.com/thomaspoignant/go-feature-flag/exporter"
)

type Exporter struct {
	ExportedEvents    []exporter.FeatureEvent
	Err               error
	ExpectedNumberErr int
	CurrentNumberErr  int
	Bulk              bool

	mutex sync.Mutex
	once  sync.Once
}

func (m *Exporter) Export(ctx context.Context, logger *log.Logger, events []exporter.FeatureEvent) error {
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

func (m *Exporter) GetExportedEvents() []exporter.FeatureEvent {
	m.once.Do(m.initMutex)
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.ExportedEvents
}

func (m *Exporter) IsBulk() bool {
	return m.Bulk
}

func (m *Exporter) initMutex() {
	m.mutex = sync.Mutex{}
}

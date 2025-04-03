package mock

import (
	"context"
	"log"
	"sync"

	"github.com/thomaspoignant/go-feature-flag/exporter"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

type ExporterMock interface {
	exporter.CommonExporter
	GetExportedEvents() []exporter.ExportableEvent
}
type Exporter struct {
	ExportedEvents    []exporter.ExportableEvent
	Err               error
	ExpectedNumberErr int
	CurrentNumberErr  int
	Bulk              bool

	mutex sync.Mutex
	once  sync.Once
}

func (m *Exporter) Export(
	_ context.Context,
	_ *fflog.FFLogger,
	events []exporter.ExportableEvent,
) error {
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

func (m *Exporter) GetExportedEvents() []exporter.ExportableEvent {
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

// ExporterDeprecated -----
type ExporterDeprecated struct {
	ExportedEvents    []exporter.FeatureEvent
	Err               error
	ExpectedNumberErr int
	CurrentNumberErr  int
	Bulk              bool

	mutex sync.Mutex
	once  sync.Once
}

func (m *ExporterDeprecated) Export(
	_ context.Context,
	_ *log.Logger,
	events []exporter.FeatureEvent,
) error {
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

func (m *ExporterDeprecated) GetExportedEvents() []exporter.ExportableEvent {
	m.once.Do(m.initMutex)
	m.mutex.Lock()
	defer m.mutex.Unlock()

	exportableEvents := make([]exporter.ExportableEvent, len(m.ExportedEvents))
	for index, event := range m.ExportedEvents {
		exportableEvents[index] = event
	}
	return exportableEvents
}

func (m *ExporterDeprecated) IsBulk() bool {
	return m.Bulk
}

func (m *ExporterDeprecated) initMutex() {
	m.mutex = sync.Mutex{}
}

// ExporterDeprecatedV2 -----
type ExporterDeprecatedV2 struct {
	ExportedEvents    []exporter.FeatureEvent
	Err               error
	ExpectedNumberErr int
	CurrentNumberErr  int
	Bulk              bool

	mutex sync.Mutex
	once  sync.Once
}

func (m *ExporterDeprecatedV2) Export(
	_ context.Context,
	_ *fflog.FFLogger,
	events []exporter.FeatureEvent,
) error {
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

func (m *ExporterDeprecatedV2) GetExportedEvents() []exporter.ExportableEvent {
	m.once.Do(m.initMutex)
	m.mutex.Lock()
	defer m.mutex.Unlock()

	exportableEvents := make([]exporter.ExportableEvent, len(m.ExportedEvents))
	for index, event := range m.ExportedEvents {
		exportableEvents[index] = event
	}
	return exportableEvents
}

func (m *ExporterDeprecatedV2) IsBulk() bool {
	return m.Bulk
}

func (m *ExporterDeprecatedV2) initMutex() {
	m.mutex = sync.Mutex{}
}

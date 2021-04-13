package exporter

import (
	"log"
	"sync"
	"time"

	"github.com/thomaspoignant/go-feature-flag/internal/fflog"
)

const defaultFlushInterval = 60 * time.Second
const defaultMaxEventInMemory = int64(100000)

// NewDataExporterScheduler allows to create a new instance of DataExporterScheduler ready to be used to export data.
func NewDataExporterScheduler(flushInterval time.Duration, maxEventInMemory int64,
	exporter Exporter, logger *log.Logger) *DataExporterScheduler {
	if flushInterval == 0 {
		flushInterval = defaultFlushInterval
	}

	if maxEventInMemory == 0 {
		maxEventInMemory = defaultMaxEventInMemory
	}

	return &DataExporterScheduler{
		localCache:      make([]FeatureEvent, 0),
		mutex:           sync.Mutex{},
		maxEventInCache: maxEventInMemory,
		exporter:        exporter,
		daemonChan:      make(chan struct{}),
		ticker:          time.NewTicker(flushInterval),
		logger:          logger,
	}
}

// Exporter is an interface to describe how a exporter looks like.
type Exporter interface {
	// Export will send the data to the exporter.
	Export(*log.Logger, []FeatureEvent) error

	// IsBulk return false if we should directly send the data as soon as it is produce
	// and true if we collect the data to send them in bulk.
	IsBulk() bool
}

// DataExporterScheduler is the struct that handle the data collection.
type DataExporterScheduler struct {
	localCache      []FeatureEvent
	mutex           sync.Mutex
	daemonChan      chan struct{}
	ticker          *time.Ticker
	maxEventInCache int64
	exporter        Exporter
	logger          *log.Logger
}

// AddEvent allow to add an event to the local cache and to call the exporter if we reach
// the maximum number of events that can be present in the cache.
func (dc *DataExporterScheduler) AddEvent(event FeatureEvent) {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()

	if !dc.exporter.IsBulk() {
		// if we are not in bulk we are directly flushing the data
		dc.localCache = append(dc.localCache, event)
		dc.flush()
		return
	}

	if int64(len(dc.localCache)) >= dc.maxEventInCache {
		dc.flush()
	}
	dc.localCache = append(dc.localCache, event)
}

// StartDaemon will start a goroutine to check every X seconds if we should send the data.
// The daemon is started only if we have a bulk exporter.
func (dc *DataExporterScheduler) StartDaemon() {
	for {
		select {
		case <-dc.ticker.C:
			// send data and clear local cache
			dc.mutex.Lock()
			dc.flush()
			dc.mutex.Unlock()
		case <-dc.daemonChan:
			// stop the daemon
			return
		}
	}
}

// Close will stop the daemon and send the data still in the cache
func (dc *DataExporterScheduler) Close() {
	// Close the daemon
	dc.ticker.Stop()
	close(dc.daemonChan)

	// Send the data still in the cache
	dc.mutex.Lock()
	dc.flush()
	dc.mutex.Unlock()
}

// flush will call the data exporter and clear the cache
// this method should be always called with a mutex
func (dc *DataExporterScheduler) flush() {
	if len(dc.localCache) > 0 {
		err := dc.exporter.Export(dc.logger, dc.localCache)
		if err != nil {
			fflog.Printf(dc.logger, "[%v] error while exporting data: %v\n", time.Now().Format(time.RFC3339), err)
			return
		}
	}
	// Clear the cache
	dc.localCache = make([]FeatureEvent, 0)
}

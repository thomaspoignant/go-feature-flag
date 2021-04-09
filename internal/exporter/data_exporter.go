package exporter

import (
	"log"
	"sync"
	"time"
)

const defaultFlushInterval = 60 * time.Second
const defaultMaxEventInCache = int64(100000)

// NewDataExporter allows to create a new instance of DataExporter ready to be used to export data.
func NewDataExporter(flushInterval time.Duration, maxEventInCache int64,
	collector Exporter, logger *log.Logger) *DataExporter {
	if flushInterval == 0 {
		flushInterval = defaultFlushInterval
	}

	if maxEventInCache == 0 {
		maxEventInCache = defaultMaxEventInCache
	}

	return &DataExporter{
		localCache:      make([]FeatureEvent, 0),
		mutex:           sync.Mutex{},
		maxEventInCache: maxEventInCache,
		exporter:        collector,
		daemonChan:      make(chan struct{}),
		ticker:          time.NewTicker(flushInterval),
		logger:          logger,
	}
}

// Exporter is an interface to describe how a exporter looks like.
type Exporter interface {
	Export(*log.Logger, []FeatureEvent) error
}

// DataExporter is the struct that handle the data collection.
type DataExporter struct {
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
func (dc *DataExporter) AddEvent(event FeatureEvent) {
	dc.mutex.Lock()
	if int64(len(dc.localCache)) >= dc.maxEventInCache {
		dc.sendData()
	}
	dc.localCache = append(dc.localCache, event)
	dc.mutex.Unlock()
}

// StartDaemon will start a goroutine to check every X seconds if we should send the data.
func (dc *DataExporter) StartDaemon() {
	for {
		select {
		case <-dc.ticker.C:
			// send data and clear local cache
			dc.mutex.Lock()
			dc.sendData()
			dc.mutex.Unlock()
		case <-dc.daemonChan:
			// stop the daemon
			return
		}
	}
}

// Close will stop the daemon and send the data still in the cache
func (dc *DataExporter) Close() {
	// Close the daemon
	dc.ticker.Stop()
	close(dc.daemonChan)

	// Send the data still in the cache
	dc.mutex.Lock()
	dc.sendData()
	dc.mutex.Unlock()
}

// sendData will call the data exporter and clear the cache
// this method should be always called with a mutex
func (dc *DataExporter) sendData() {
	if len(dc.localCache) > 0 {
		err := dc.exporter.Export(dc.logger, dc.localCache)
		if err != nil {
			if dc.logger != nil {
				dc.logger.Printf("[%v] error while exporting data: %v\n", time.Now().Format(time.RFC3339), err)
			}
		}
	}
	// Clear the cache
	dc.localCache = make([]FeatureEvent, 0)
}

package exporter

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"sync"
	"time"

	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

const (
	defaultFlushInterval    = 60 * time.Second
	defaultMaxEventInMemory = int64(100000)
)

// NewScheduler allows creating a new instance of Scheduler ready to be used to export data.
func NewScheduler(ctx context.Context, flushInterval time.Duration, maxEventInMemory int64,
	exp CommonExporter, logger *fflog.FFLogger,
) *Scheduler {
	if ctx == nil {
		ctx = context.Background()
	}

	if flushInterval == 0 {
		flushInterval = defaultFlushInterval
	}

	if maxEventInMemory == 0 {
		maxEventInMemory = defaultMaxEventInMemory
	}

	return &Scheduler{
		localCache:      make([]FeatureEvent, 0),
		mutex:           sync.Mutex{},
		maxEventInCache: maxEventInMemory,
		exporter:        exp,
		daemonChan:      make(chan struct{}),
		ticker:          time.NewTicker(flushInterval),
		logger:          logger,
		ctx:             ctx,
	}
}

// Scheduler is the struct that handle the data collection.
type Scheduler struct {
	localCache      []FeatureEvent
	mutex           sync.Mutex
	daemonChan      chan struct{}
	ticker          *time.Ticker
	maxEventInCache int64
	exporter        CommonExporter
	logger          *fflog.FFLogger
	ctx             context.Context
}

// AddEvent allow adding an event to the local cache and to call the exporter if we reach
// the maximum number of events that can be present in the cache.
func (dc *Scheduler) AddEvent(event FeatureEvent) {
	if !dc.exporter.IsBulk() {
		err := sendEvents(dc.ctx, dc.exporter, dc.logger, []FeatureEvent{event})
		if err != nil {
			dc.logger.Error(err.Error())
		}
		return
	}

	dc.mutex.Lock()
	defer dc.mutex.Unlock()
	if int64(len(dc.localCache)) >= dc.maxEventInCache {
		dc.flush()
	}
	dc.localCache = append(dc.localCache, event)
}

// StartDaemon will start a goroutine to check every X seconds if we should send the data.
// The daemon is started only if we have a bulk exporter.
func (dc *Scheduler) StartDaemon() {
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
func (dc *Scheduler) Close() {
	// Close the daemon
	dc.ticker.Stop()
	close(dc.daemonChan)

	// Send the data still in the cache
	dc.mutex.Lock()
	dc.flush()
	dc.mutex.Unlock()
}

// GetLogger will return the logger used by the scheduler
func (dc *Scheduler) GetLogger(level slog.Level) *log.Logger {
	if dc.logger == nil {
		return nil
	}
	return dc.logger.GetLogLogger(level)
}

// flush will call the data exporter and clear the cache
func (dc *Scheduler) flush() {
	err := sendEvents(dc.ctx, dc.exporter, dc.logger, dc.localCache)
	if err != nil {
		dc.logger.Error(err.Error())
		return
	}
	dc.localCache = make([]FeatureEvent, 0)
}

func sendEvents(ctx context.Context, exporter CommonExporter, logger *fflog.FFLogger, events []FeatureEvent) error {
	if len(events) == 0 {
		return nil
	}

	switch exp := exporter.(type) {
	case DeprecatedExporter:
		var legacyLogger *log.Logger
		if logger != nil {
			legacyLogger = logger.GetLogLogger(slog.LevelError)
		}
		// use dc exporter as a DeprecatedExporter
		err := exp.Export(ctx, legacyLogger, events)
		slog.Warn("You are using an exporter with the old logger."+
			"Please update your custom exporter to comply to the new Exporter interface.",
			slog.Any("err", err))
		if err != nil {
			return fmt.Errorf("error while exporting data: %w", err)
		}
		break
	case Exporter:
		err := exp.Export(ctx, logger, events)
		if err != nil {
			return fmt.Errorf("error while exporting data: %w", err)
		}
		break
	default:
		return fmt.Errorf("this is not a valid exporter")
	}
	return nil
}

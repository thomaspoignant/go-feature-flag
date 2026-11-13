package exporter

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

const (
	defaultFlushInterval    = 60 * time.Second
	defaultMaxEventInMemory = int64(100000)
	// minRetryDelay is the time we wait before retrying the first time an export in error.
	minRetryDelay = 1 * time.Second
)

type DataExporter[T ExportableEvent] interface {
	// Start is launching the daemon in charge of flushing the data
	Start()
	// Stop is stopping the daemon
	Stop()
	// Flush is sending the data to the exporter
	Flush()
	// RequestFlush is asking the daemon to flush the data as soon as possible, it never blocks the caller.
	RequestFlush()
	// IsBulk return false if we should directly send the data as soon as it is produce
	IsBulk() bool
	// GetConsumerID return the consumer ID used in the event store
	GetConsumerID() string
	// GetMaxEventInMemory return the maximum number of event you keep in the cache before calling Flush()
	GetMaxEventInMemory() int64
}
type Config struct {
	Exporter         CommonExporter
	FlushInterval    time.Duration
	MaxEventInMemory int64
}

type dataExporterImpl[T ExportableEvent] struct {
	consumerID string
	eventStore EventStore[T]
	logger     *fflog.FFLogger
	exporter   Config

	daemonChan chan struct{}
	// flushChan is used to ask the daemon to flush, it has a buffer of 1 because we only need to know
	// that at least 1 flush was requested since the last one.
	flushChan chan struct{}
	ticker    *time.Ticker
	// stopOnce guarantees that we stop the daemon only once: Close() can be called several times
	// on a GO Feature Flag client, and closing an already closed channel panics.
	stopOnce sync.Once
	// daemonStarted guarantees that we never have more than 1 daemon for an exporter.
	daemonStarted atomic.Bool
}

// NewDataExporter create a new DataExporter with the given exporter and his consumer information to consume the data
// from the shared event store.
func NewDataExporter[T ExportableEvent](exporter Config, consumerID string,
	eventStore EventStore[T], logger *fflog.FFLogger) DataExporter[T] {
	if exporter.FlushInterval == 0 {
		exporter.FlushInterval = defaultFlushInterval
	}

	if exporter.MaxEventInMemory == 0 {
		exporter.MaxEventInMemory = defaultMaxEventInMemory
	}

	return &dataExporterImpl[T]{
		consumerID: consumerID,
		eventStore: eventStore,
		logger:     logger,
		exporter:   exporter,
		daemonChan: make(chan struct{}),
		flushChan:  make(chan struct{}, 1),
		ticker:     time.NewTicker(exporter.FlushInterval),
	}
}

// Start is launching the daemon in charge of flushing the data, and it returns when the daemon is stopped.
// It returns directly if the daemon is already running.
func (d *dataExporterImpl[T]) Start() {
	if !d.daemonStarted.CompareAndSwap(false, true) {
		return
	}
	d.daemon()
}

// daemon is the only one to call the exporter while GO Feature Flag is running, it flushes every time
// a flush is requested, and also periodically if we are in bulk mode.
//
// If an export is in error we retry it with an exponential backoff that is never higher than the flush interval,
// and we ignore the flush requests until the next retry to avoid calling for every event an exporter that is down.
func (d *dataExporterImpl[T]) daemon() {
	var flushTicker <-chan time.Time
	if d.IsBulk() {
		// we don't flush periodically if we are not in bulk mode
		flushTicker = d.ticker.C
	}

	var retryDelay time.Duration
	var retryTimer *time.Timer
	// retryChan is nil when the last export was in success, reading a nil channel is blocking forever.
	var retryChan <-chan time.Time
	flush := func() {
		if err := d.flush(); err != nil {
			retryDelay = d.nextRetryDelay(retryDelay)
			retryTimer = time.NewTimer(retryDelay)
			retryChan = retryTimer.C
			return
		}
		retryDelay = 0
		retryChan = nil
	}

	for {
		select {
		case <-flushTicker:
			if retryChan == nil {
				flush()
			}
		case <-d.flushChan:
			if retryChan == nil {
				flush()
			}
		case <-retryChan:
			flush()
		case <-d.daemonChan:
			// stop the daemon
			if retryTimer != nil {
				retryTimer.Stop()
			}
			return
		}
	}
}

// nextRetryDelay returns the time to wait before retrying an export in error.
func (d *dataExporterImpl[T]) nextRetryDelay(previousDelay time.Duration) time.Duration {
	return min(max(previousDelay*2, minRetryDelay), d.exporter.FlushInterval)
}

// Stop is stopping the daemon and flushing the data.
// It is safe to call it several times.
func (d *dataExporterImpl[T]) Stop() {
	d.stopOnce.Do(func() {
		d.ticker.Stop()
		close(d.daemonChan)
	})
	d.Flush()
}

// Flush is sending the data to the exporter
func (d *dataExporterImpl[T]) Flush() {
	_ = d.flush()
}

// RequestFlush is asking the daemon to flush the data as soon as possible, it never blocks the caller.
func (d *dataExporterImpl[T]) RequestFlush() {
	select {
	case d.flushChan <- struct{}{}:
	default:
		// a flush is already requested, the daemon will take all the pending events when it does it.
	}
	// An exporter that is not in bulk mode was exporting the events without having to call Start(),
	// so we start the daemon here if it is not running to keep it working.
	if d.daemonStarted.CompareAndSwap(false, true) {
		go d.daemon()
	}
}

// flush is sending the data to the exporter, and it returns an error if the export has failed.
func (d *dataExporterImpl[T]) flush() error {
	err := d.eventStore.ProcessPendingEvents(d.consumerID, d.sendEvents)
	if err != nil && d.logger != nil {
		d.logger.Error(err.Error())
	}
	return err
}

// IsBulk return false if we should directly send the data as soon as it is produce
func (d *dataExporterImpl[T]) IsBulk() bool {
	return d.exporter.Exporter.IsBulk()
}

// GetConsumerID return the consumer ID used in the event store
func (d *dataExporterImpl[T]) GetConsumerID() string {
	return d.consumerID
}

// GetMaxEventInMemory return the maximum number of event you keep in the cache before calling Flush()
func (d *dataExporterImpl[T]) GetMaxEventInMemory() int64 {
	return d.exporter.MaxEventInMemory
}

// sendEvents is sending the events to the exporter.
func (d *dataExporterImpl[T]) sendEvents(ctx context.Context, events []T) error {
	if len(events) == 0 {
		return nil
	}
	switch exp := d.exporter.Exporter.(type) {
	case DeprecatedExporterV1:
		var legacyLogger *log.Logger
		if d.logger != nil {
			legacyLogger = d.logger.GetLogLogger(slog.LevelError)
		}
		switch events := any(events).(type) {
		case []FeatureEvent:
			// use dc exporter as a DeprecatedExporterV1
			err := exp.Export(ctx, legacyLogger, events)
			slog.WarnContext(ctx, "You are using an exporter with the old logger."+
				"Please update your custom exporter to comply to the new Exporter interface.")
			if err != nil {
				return fmt.Errorf("error while exporting data (deprecated): %w", err)
			}
		default:
			return fmt.Errorf("trying to send unknown object to the exporter (deprecated)")
		}
	case DeprecatedExporterV2:
		switch events := any(events).(type) {
		case []FeatureEvent:
			err := exp.Export(ctx, d.logger, events)
			if err != nil {
				return fmt.Errorf("error while exporting data: %w", err)
			}
		default:
			return fmt.Errorf("trying to send unknown object to the exporter")
		}
	case Exporter:
		exportableEvents := make([]ExportableEvent, len(events))
		for i, event := range events {
			exportableEvents[i] = ExportableEvent(event)
		}
		err := exp.Export(ctx, d.logger, exportableEvents)
		if err != nil {
			return fmt.Errorf("error while exporting data: %w", err)
		}
	default:
		return fmt.Errorf("this is not a valid exporter")
	}
	return nil
}

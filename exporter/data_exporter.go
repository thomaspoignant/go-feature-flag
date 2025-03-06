package exporter

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

const (
	defaultFlushInterval    = 60 * time.Second
	defaultMaxEventInMemory = int64(100000)
)

type Config struct {
	Exporter         CommonExporter
	FlushInterval    time.Duration
	MaxEventInMemory int64
}

type dataExporterImpl[T any] struct {
	ctx        context.Context
	consumerID string
	eventStore *EventStore[T]
	logger     *fflog.FFLogger
	exporter   Config

	daemonChan chan struct{}
	ticker     *time.Ticker
}

func NewDataExporter[T any](ctx context.Context, exporter Config, consumerID string, eventStore *EventStore[T], logger *fflog.FFLogger) dataExporterImpl[T] {
	if ctx == nil {
		ctx = context.Background()
	}

	if exporter.FlushInterval == 0 {
		exporter.FlushInterval = defaultFlushInterval
	}

	if exporter.MaxEventInMemory == 0 {
		exporter.MaxEventInMemory = defaultMaxEventInMemory
	}

	return dataExporterImpl[T]{
		ctx:        ctx,
		consumerID: consumerID,
		eventStore: eventStore,
		logger:     logger,
		exporter:   exporter,
		daemonChan: make(chan struct{}),
		ticker:     time.NewTicker(exporter.FlushInterval),
	}
}

func (d *dataExporterImpl[T]) Start() {
	for {
		select {
		case <-d.ticker.C:
			d.Flush()
		case <-d.daemonChan:
			// stop the daemon
			return
		}
	}
}

func (d *dataExporterImpl[T]) Stop() {
	d.ticker.Stop()
	close(d.daemonChan)
	d.Flush()
}

func (d *dataExporterImpl[T]) Flush() {
	store := *d.eventStore
	eventList, err := store.FetchPendingEvents(d.consumerID)
	if err != nil {
		// log something here
		d.logger.Error("error while fetching pending events", err)
		return
	}
	err = d.sendEvents(d.ctx, eventList.Events)
	if err != nil {
		d.logger.Error(err.Error())
		return
	}
	err = store.UpdateConsumerOffset(d.consumerID, eventList.NewOffset)
	if err != nil {
		d.logger.Error("error while updating offset", err.Error())
	}
}

func (d *dataExporterImpl[T]) sendEvents(ctx context.Context, events []T) error {
	if len(events) == 0 {
		return nil
	}
	switch exp := d.exporter.Exporter.(type) {
	case DeprecatedExporter:
		var legacyLogger *log.Logger
		if d.logger != nil {
			legacyLogger = d.logger.GetLogLogger(slog.LevelError)
		}
		switch events := any(events).(type) {
		case []FeatureEvent:
			// use dc exporter as a DeprecatedExporter
			err := exp.Export(ctx, legacyLogger, events)
			slog.Warn("You are using an exporter with the old logger."+
				"Please update your custom exporter to comply to the new Exporter interface.",
				slog.Any("err", err))
			if err != nil {
				return fmt.Errorf("error while exporting data: %w", err)
			}
		default:
			return fmt.Errorf("trying to send unknown object to the exporter (deprecated)")
		}
		break
	case Exporter:
		switch events := any(events).(type) {
		case []FeatureEvent:
			err := exp.Export(ctx, d.logger, events)
			if err != nil {
				return fmt.Errorf("error while exporting data: %w", err)
			}
		default:
			return fmt.Errorf("trying to send unknown object to the exporter")
		}
		break
	default:
		return fmt.Errorf("this is not a valid exporter")
	}
	return nil
}

package exporter

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

type ManagerImpl[T any] struct {
	logger     *fflog.FFLogger
	consumers  []dataExporterImpl[T]
	eventStore *EventStore[T]
}

func NewManager[T any](ctx context.Context, exporters []DataExporter, logger *fflog.FFLogger) Manager[T] {
	if ctx == nil {
		ctx = context.Background()
	}

	evStore := NewEventStore[T](30 * time.Second)
	consumers := make([]dataExporterImpl[T], len(exporters))
	for index, exporter := range exporters {
		consumerId := uuid.New().String()
		exp := NewDataExporter[T](ctx, exporter, consumerId, &evStore, logger)
		consumers[index] = exp
		evStore.AddConsumer(consumerId)
	}
	return &ManagerImpl[T]{
		logger:     logger,
		consumers:  consumers,
		eventStore: &evStore,
	}
}

func (m *ManagerImpl[T]) AddEvent(event T) {
	store := *m.eventStore
	store.Add(event)
	for _, consumer := range m.consumers {
		if !consumer.exporter.Exporter.IsBulk() {
			consumer.Flush()
			continue
		}

		count, err := store.GetPendingEventCount(consumer.consumerId)
		if err != nil {
			m.logger.Error("error while fetching pending events", err)
			continue
		}

		if count >= consumer.exporter.MaxEventInMemory {
			consumer.Flush()
			continue
		}
	}
}

func (m *ManagerImpl[T]) StartDaemon() {
	for _, consumer := range m.consumers {
		go consumer.Start()
	}
}

func (m *ManagerImpl[T]) Close() {
	for _, consumer := range m.consumers {
		consumer.Stop()
	}
}

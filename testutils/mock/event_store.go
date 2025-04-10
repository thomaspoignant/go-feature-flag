package mock

import (
	"context"
	"fmt"

	"github.com/thomaspoignant/go-feature-flag/exporter"
)

const consumerNameError = "error"

type implMockEventStore[T exporter.ExportableEvent] struct {
	store []T
}

func NewEventStore[T exporter.ExportableEvent]() exporter.EventStore[T] {
	store := &implMockEventStore[T]{}
	return store
}

func (e *implMockEventStore[T]) AddConsumer(_ string) {
	// nothing to do
}

func (e *implMockEventStore[T]) Add(data T) {
	e.store = append(e.store, data)
}

func (e *implMockEventStore[T]) ProcessPendingEvents(consumerID string,
	processFunc func(context.Context, []T) error) error {
	if consumerID == consumerNameError {
		return fmt.Errorf("error")
	}
	if err := processFunc(context.TODO(), e.store); err != nil {
		return err
	}
	return nil
}

func (e *implMockEventStore[T]) GetPendingEventCount(consumerName string) (int64, error) {
	if consumerName == consumerNameError {
		return 0, fmt.Errorf("error")
	}
	return int64(len(e.store)), nil
}

func (e *implMockEventStore[T]) GetTotalEventCount() int64 {
	return int64(len(e.store))
}

func (e *implMockEventStore[T]) Stop() {
	// nothing to do
}

package mock

import (
	"fmt"

	"github.com/thomaspoignant/go-feature-flag/exporter"
)

type implMockEventStore[T any] struct {
	store []T
}

func NewEventStore[T any]() exporter.EventStore[T] {
	store := &implMockEventStore[T]{}
	return store
}

func (e *implMockEventStore[T]) AddConsumer(_ string) {
	// nothing to do
}

func (e *implMockEventStore[T]) Add(data T) {
	e.store = append(e.store, data)
}

func (e *implMockEventStore[T]) FetchPendingEvents(consumerName string) (*exporter.EventList[T], error) {
	if consumerName == "error" {
		return nil, fmt.Errorf("error")
	}
	return &exporter.EventList[T]{
		InitialOffset: 0,
		NewOffset:     1,
		Events:        e.store,
	}, nil
}

func (e *implMockEventStore[T]) GetPendingEventCount(consumerName string) (int64, error) {
	if consumerName == "error" {
		return 0, fmt.Errorf("error")
	}
	return int64(len(e.store)), nil
}

func (e *implMockEventStore[T]) GetTotalEventCount() int64 {
	return int64(len(e.store))
}

func (e *implMockEventStore[T]) UpdateConsumerOffset(consumerName string, offset int64) error {
	if consumerName == "error" || consumerName == "error_update" {
		return fmt.Errorf("error")
	}
	return nil
}

func (e *implMockEventStore[T]) Stop() {
	// nothing to do
}

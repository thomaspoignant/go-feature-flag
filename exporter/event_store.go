package exporter

import (
	"fmt"
	"math"
	"sync"
	"time"
)

const minOffset = int64(math.MinInt64)

type eventStoreImpl[T any] struct {
	// events is a list of events to store
	events []Event[T]
	// mutex to protect the events and consumers
	mutex sync.RWMutex
	// consumers is a map of consumers with their name as key
	consumers map[string]*consumer
	// lastOffset is the last offset used for the Event store
	lastOffset int64
	// stopPeriodicCleaning is a channel to stop the periodic cleaning goroutine
	stopPeriodicCleaning chan struct{}
	// cleanQueueInterval is the duration between each cleaning
	cleanQueueInterval time.Duration
}

func NewEventStore[T any](cleanQueueInterval time.Duration) EventStore[T] {
	store := &eventStoreImpl[T]{
		events:               make([]Event[T], 0),
		mutex:                sync.RWMutex{},
		lastOffset:           minOffset,
		stopPeriodicCleaning: make(chan struct{}),
		cleanQueueInterval:   cleanQueueInterval,
		consumers:            make(map[string]*consumer),
	}
	go store.periodicCleanQueue()
	return store
}

type EventList[T any] struct {
	Events        []T
	InitialOffset int64
	NewOffset     int64
}

type EventStore[T any] interface {
	// AddConsumer is adding a new consumer to the Event store.
	// note that you can't add a consumer after the Event store has been started.
	AddConsumer(consumerName string)

	// Add is adding item of type T in the Event store.
	Add(data T)

	// FetchPendingEvents is returning all the available item in the Event store for this consumer.
	FetchPendingEvents(consumerName string) (*EventList[T], error)

	// GetPendingEventCount is returning the number items available in the Event store for this consumer.
	GetPendingEventCount(consumerName string) (int64, error)

	// GetTotalEventCount returns the total number of events in the store.
	GetTotalEventCount() int64

	// UpdateConsumerOffset updates the offset of the consumer to the new offset.
	UpdateConsumerOffset(consumerName string, offset int64) error

	// Stop is closing the Event store and stop the periodic cleaning.
	Stop()
}

type Event[T any] struct {
	Offset int64
	Data   T
}

type consumer struct {
	Offset int64
}

func (e *eventStoreImpl[T]) AddConsumer(consumerName string) {
	e.consumers[consumerName] = &consumer{Offset: e.lastOffset}
}

// GetTotalEventCount returns the total number of events in the store.
func (e *eventStoreImpl[T]) GetTotalEventCount() int64 {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	return int64(len(e.events))
}

// GetPendingEventCount is returning the number items available in the Event store for this consumer.
func (e *eventStoreImpl[T]) GetPendingEventCount(consumerName string) (int64, error) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	consumer, ok := e.consumers[consumerName]
	if !ok {
		return 0, fmt.Errorf("consumer with name %s not found", consumerName)
	}
	return e.lastOffset - consumer.Offset, nil
}

// Add is adding item of type T in the Event store.
func (e *eventStoreImpl[T]) Add(data T) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.lastOffset++
	e.events = append(e.events, Event[T]{Offset: e.lastOffset, Data: data})
}

// FetchPendingEvents is returning all the available item in the Event store for this consumer.
func (e *eventStoreImpl[T]) FetchPendingEvents(consumerName string) (*EventList[T], error) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	currentConsumer, ok := e.consumers[consumerName]
	if !ok {
		return nil, fmt.Errorf("consumer with name %s not found", consumerName)
	}
	events := make([]T, 0)
	for _, event := range e.events {
		if event.Offset > currentConsumer.Offset {
			events = append(events, event.Data)
		}
	}
	return &EventList[T]{Events: events, InitialOffset: currentConsumer.Offset, NewOffset: e.lastOffset}, nil
}

// UpdateConsumerOffset updates the offset of the consumer to the new offset.
func (e *eventStoreImpl[T]) UpdateConsumerOffset(consumerName string, offset int64) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	if offset > e.lastOffset {
		return fmt.Errorf("invalid offset: offset %d is greater than the last offset %d", offset, e.lastOffset)
	}
	if _, ok := e.consumers[consumerName]; !ok {
		return fmt.Errorf("invalid offset consumerName %s", consumerName)
	}
	e.consumers[consumerName].Offset = e.lastOffset
	return nil
}

// cleanQueue removes all events that have been consumed by all consumers
func (e *eventStoreImpl[T]) cleanQueue() {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	if len(e.events) == 0 {
		// nothing to remove
		return
	}
	consumerMinOffset := minOffset
	for _, currentConsumer := range e.consumers {
		if consumerMinOffset == minOffset || currentConsumer.Offset < consumerMinOffset {
			consumerMinOffset = currentConsumer.Offset
		}
	}
	if consumerMinOffset <= minOffset {
		// nothing to remove
		return
	}

	for i, event := range e.events {
		if event.Offset == consumerMinOffset {
			e.events = e.events[i+1:]
			break
		}
	}
}

// periodicCleanQueue periodically cleans the queue
func (e *eventStoreImpl[T]) periodicCleanQueue() {
	ticker := time.NewTicker(e.cleanQueueInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			e.cleanQueue()
		case <-e.stopPeriodicCleaning:
			return
		}
	}
}

// Stop is closing the Event store and stop the periodic cleaning.
func (e *eventStoreImpl[T]) Stop() {
	close(e.stopPeriodicCleaning)
}

package exporter_test

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/exporter"
	"github.com/thomaspoignant/go-feature-flag/testutils"
)

const defaultTestCleanQueueDuration = 100 * time.Millisecond

func Test_ConsumerNameInvalid(t *testing.T) {
	t.Run(
		"GetPendingEventCount: should return an error if the consumer name is invalid",
		func(t *testing.T) {
			eventStore := exporter.NewEventStore[testutils.ExportableMockEvent](
				defaultTestCleanQueueDuration,
			)
			eventStore.AddConsumer("consumer1")
			defer eventStore.Stop()
			_, err := eventStore.GetPendingEventCount("wrong name")
			assert.NotNil(t, err)
		},
	)
	t.Run(
		"ProcessPendingEvents: should return an error if the consumer name is invalid",
		func(t *testing.T) {
			eventStore := exporter.NewEventStore[testutils.ExportableMockEvent](
				defaultTestCleanQueueDuration,
			)
			eventStore.AddConsumer("consumer1")
			defer eventStore.Stop()
			err := eventStore.ProcessPendingEvents(
				"wrong name",
				func(ctx context.Context, events []testutils.ExportableMockEvent) error { return nil },
			)
			assert.NotNil(t, err)
		},
	)
}

func Test_SingleConsumer(t *testing.T) {
	consumerName := "consumer1"
	eventStore := exporter.NewEventStore[testutils.ExportableMockEvent](
		defaultTestCleanQueueDuration,
	)
	eventStore.AddConsumer(consumerName)
	defer eventStore.Stop()
	got, _ := eventStore.GetPendingEventCount(consumerName)
	assert.Equal(t, int64(0), got)
	assert.Equal(t, int64(0), eventStore.GetTotalEventCount())

	// start producer
	ctx, cancel := context.WithCancel(context.Background())
	go startEventProducer(ctx, eventStore, 100, false)
	time.Sleep(50 * time.Millisecond)
	got, _ = eventStore.GetPendingEventCount(consumerName)
	assert.Equal(t, int64(100), got)
	cancel() // stop producing

	// Consume
	err := eventStore.ProcessPendingEvents(consumerName,
		func(ctx context.Context, events []testutils.ExportableMockEvent) error {
			assert.Equal(t, 100, len(events))
			return nil
		})
	assert.Nil(t, err)
	got, _ = eventStore.GetPendingEventCount(consumerName)
	assert.Equal(t, int64(0), got)

	// restart producing
	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel2()
	go startEventProducer(ctx2, eventStore, 91, false)
	time.Sleep(50 * time.Millisecond)
	got, _ = eventStore.GetPendingEventCount(consumerName)
	assert.Equal(t, int64(91), got)

	err = eventStore.ProcessPendingEvents(consumerName,
		func(ctx context.Context, events []testutils.ExportableMockEvent) error {
			assert.Equal(t, 91, len(events))
			return nil
		})
	assert.Nil(t, err)

	time.Sleep(120 * time.Millisecond) // to wait until garbage collector remove the events
	assert.Equal(t, int64(0), eventStore.GetTotalEventCount())
}

func Test_MultipleConsumersSingleThread(t *testing.T) {
	consumerNames := []string{"consumer1", "consumer2"}
	eventStore := exporter.NewEventStore[testutils.ExportableMockEvent](
		defaultTestCleanQueueDuration,
	)
	for _, name := range consumerNames {
		eventStore.AddConsumer(name)
	}
	defer eventStore.Stop()
	// start producer
	ctx, cancelProducer1 := context.WithCancel(context.Background())
	defer cancelProducer1()
	startEventProducer(ctx, eventStore, 1000, false)
	cancelProducer1()
	assert.Equal(t, int64(1000), eventStore.GetTotalEventCount())

	// Consume with Consumer1 only
	consumer1Size, err := eventStore.GetPendingEventCount(consumerNames[0])
	assert.Nil(t, err)
	assert.Equal(t, int64(1000), consumer1Size)
	err = eventStore.ProcessPendingEvents(consumerNames[0],
		func(ctx context.Context, events []testutils.ExportableMockEvent) error {
			assert.Equal(t, 1000, len(events))
			return nil
		})
	assert.Nil(t, err)

	// Produce a second time
	ctx, cancelProducer2 := context.WithCancel(context.Background())
	defer cancelProducer2()
	startEventProducer(ctx, eventStore, 1000, false)
	cancelProducer2()

	// Check queue size
	assert.Equal(t, int64(2000), eventStore.GetTotalEventCount())
	consumer1Size, err = eventStore.GetPendingEventCount(consumerNames[0])
	assert.Nil(t, err)
	assert.Equal(t, int64(1000), consumer1Size)
	consumer2Size, err := eventStore.GetPendingEventCount(consumerNames[1])
	assert.Nil(t, err)
	assert.Equal(t, int64(2000), consumer2Size)

	// Consumer with Consumer1 and Consumer2
	err = eventStore.ProcessPendingEvents(consumerNames[0],
		func(ctx context.Context, events []testutils.ExportableMockEvent) error {
			assert.Equal(t, 1000, len(events))
			return nil
		})
	assert.Nil(t, err)

	err = eventStore.ProcessPendingEvents(consumerNames[1],
		func(ctx context.Context, events []testutils.ExportableMockEvent) error {
			assert.Equal(t, 2000, len(events))
			return nil
		})
	assert.Nil(t, err)

	// Check garbage collector
	time.Sleep(120 * time.Millisecond)
	assert.Equal(t, int64(0), eventStore.GetTotalEventCount())
}

func Test_MultipleConsumersMultipleGORoutines(t *testing.T) {
	consumerNames := []string{"consumer1", "consumer2"}
	eventStore := exporter.NewEventStore[testutils.ExportableMockEvent](
		defaultTestCleanQueueDuration,
	)
	for _, name := range consumerNames {
		eventStore.AddConsumer(name)
	}
	defer eventStore.Stop()
	// start producer
	ctx, cancelProducer1 := context.WithCancel(context.Background())
	defer cancelProducer1()
	go startEventProducer(ctx, eventStore, 100000, true)
	time.Sleep(50 * time.Millisecond)
	wg := &sync.WaitGroup{}

	consumeFunc := func(eventStore exporter.EventStore[testutils.ExportableMockEvent], consumerName string, eventCounters *map[string]int) {
		defer wg.Done()
		err := eventStore.ProcessPendingEvents(consumerName,
			func(ctx context.Context, events []testutils.ExportableMockEvent) error {
				assert.True(t, len(events) > 0)
				return nil
			})
		assert.Nil(t, err)
		time.Sleep(
			50 * time.Millisecond,
		) // we wait to be sure that the producer has produce new events

		err = eventStore.ProcessPendingEvents(consumerName,
			func(ctx context.Context, events []testutils.ExportableMockEvent) error {
				if eventCounters != nil {
					(*eventCounters)[consumerName] = len(events)
				}
				return nil
			})
		assert.Nil(t, err)
	}

	wg.Add(2)
	eventCounters := map[string]int{}
	go consumeFunc(eventStore, consumerNames[0], &eventCounters)
	go consumeFunc(eventStore, consumerNames[1], &eventCounters)
	wg.Wait()

	assert.Greater(t, eventCounters[consumerNames[0]], 0)
	assert.Greater(t, eventCounters[consumerNames[1]], 0)
}

func Test_ProcessPendingEventInError(t *testing.T) {
	consumerName := "consumer1"
	eventStore := exporter.NewEventStore[testutils.ExportableMockEvent](
		defaultTestCleanQueueDuration,
	)
	eventStore.AddConsumer(consumerName)
	defer eventStore.Stop()
	// start producer
	startEventProducer(context.TODO(), eventStore, 1000, false)
	assert.Equal(t, int64(1000), eventStore.GetTotalEventCount())

	consumer1Size, err := eventStore.GetPendingEventCount(consumerName)
	assert.Equal(t, 1000, int(consumer1Size))
	assert.Nil(t, err)

	// process is in error, so we are not able to update the offset
	err = eventStore.ProcessPendingEvents(
		consumerName,
		func(ctx context.Context, events []testutils.ExportableMockEvent) error {
			assert.Equal(t, 1000, len(events))
			return fmt.Errorf("error")
		},
	)
	assert.NotNil(t, err)

	// We still have the same number of items waiting for next process
	consumer1Size, err = eventStore.GetPendingEventCount(consumerName)
	assert.Equal(t, 1000, int(consumer1Size))
	assert.Nil(t, err)

	// process is not in error anymore
	err = eventStore.ProcessPendingEvents(
		consumerName,
		func(ctx context.Context, events []testutils.ExportableMockEvent) error {
			assert.Equal(t, 1000, len(events))
			return nil
		},
	)
	assert.Nil(t, err)

	// we have consumed all the items
	consumer1Size, err = eventStore.GetPendingEventCount(consumerName)
	assert.Equal(t, 0, int(consumer1Size))
	assert.Nil(t, err)
}

func Test_WaitForEmptyClean(t *testing.T) {
	consumerNames := []string{"consumer1"}
	eventStore := exporter.NewEventStore[testutils.ExportableMockEvent](
		defaultTestCleanQueueDuration,
	)
	for _, name := range consumerNames {
		eventStore.AddConsumer(name)
	}
	defer eventStore.Stop()

	// start producer
	ctx := context.Background()
	startEventProducer(ctx, eventStore, 100, false)
	err := eventStore.ProcessPendingEvents(
		consumerNames[0],
		func(ctx context.Context, events []testutils.ExportableMockEvent) error {
			assert.Equal(t, 100, len(events))
			return nil
		},
	)
	assert.Nil(t, err)
	assert.True(t, eventStore.GetTotalEventCount() > 0)
	time.Sleep(3 * defaultTestCleanQueueDuration)
	assert.Equal(t, int64(0), eventStore.GetTotalEventCount())
}

func startEventProducer(
	ctx context.Context,
	eventStore exporter.EventStore[testutils.ExportableMockEvent],
	produceMax int,
	randomizeProducingTime bool,
) {
	for i := 0; i < produceMax; i++ {
		select {
		case <-ctx.Done():
			return
		default:
			if randomizeProducingTime {
				randomNumber := rand.Intn(10) + 1
				time.Sleep(time.Duration(randomNumber) * time.Millisecond)
			}
			eventStore.Add(testutils.NewExportableMockEvent("Hello"))
		}
	}
}

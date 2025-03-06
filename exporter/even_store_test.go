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
)

const defaultCleanQueueDuration = 100 * time.Millisecond

func Test_ConsumerNameInvalid(t *testing.T) {
	t.Run("GetPendingEventCount: should return an error if the consumer name is invalid", func(t *testing.T) {
		eventStore := exporter.NewEventStore[string](defaultCleanQueueDuration)
		eventStore.AddConsumer("consumer1")
		defer eventStore.Stop()
		_, err := eventStore.GetPendingEventCount("wrong name")
		assert.NotNil(t, err)
	})
	t.Run("GetPendingEventCount: should return an error if the consumer name is invalid", func(t *testing.T) {
		eventStore := exporter.NewEventStore[string](defaultCleanQueueDuration)
		eventStore.AddConsumer("consumer1")
		defer eventStore.Stop()
		_, err := eventStore.FetchPendingEvents("wrong name")
		assert.NotNil(t, err)
	})
}

func Test_SingleConsumer(t *testing.T) {
	consumerName := "consumer1"
	eventStore := exporter.NewEventStore[string](defaultCleanQueueDuration)
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
	events, _ := eventStore.FetchPendingEvents(consumerName)
	assert.Equal(t, 100, len(events.Events))
	err := eventStore.UpdateConsumerOffset(consumerName, events.NewOffset)
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
	events, _ = eventStore.FetchPendingEvents(consumerName)
	err = eventStore.UpdateConsumerOffset(consumerName, events.NewOffset)
	assert.Nil(t, err)
	assert.Equal(t, 91, len(events.Events))

	time.Sleep(120 * time.Millisecond) // to wait until garbage collector remove the events
	assert.Equal(t, int64(0), eventStore.GetTotalEventCount())
}

func Test_MultipleConsumersSingleThread(t *testing.T) {
	consumerNames := []string{"consumer1", "consumer2"}
	eventStore := exporter.NewEventStore[string](defaultCleanQueueDuration)
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
	eventsConsumer1, err := eventStore.FetchPendingEvents(consumerNames[0])
	assert.Nil(t, err)
	assert.Equal(t, 1000, len(eventsConsumer1.Events))
	err = eventStore.UpdateConsumerOffset(consumerNames[0], eventsConsumer1.NewOffset)
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
	eventsConsumer1, err = eventStore.FetchPendingEvents(consumerNames[0])
	assert.Nil(t, err)
	assert.Equal(t, 1000, len(eventsConsumer1.Events))
	err = eventStore.UpdateConsumerOffset(consumerNames[0], eventsConsumer1.NewOffset)
	assert.Nil(t, err)
	eventsConsumer2, err := eventStore.FetchPendingEvents(consumerNames[1])
	assert.Nil(t, err)
	assert.Equal(t, 2000, len(eventsConsumer2.Events))
	err = eventStore.UpdateConsumerOffset(consumerNames[1], eventsConsumer1.NewOffset)
	assert.Nil(t, err)

	// Check garbage collector
	time.Sleep(120 * time.Millisecond)
	assert.Equal(t, int64(0), eventStore.GetTotalEventCount())
}

func Test_MultipleConsumersMultipleGORoutines(t *testing.T) {
	consumerNames := []string{"consumer1", "consumer2"}
	eventStore := exporter.NewEventStore[string](defaultCleanQueueDuration)
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

	consumFunc := func(eventStore exporter.EventStore[string], consumerName string) {
		wg.Add(1)
		defer wg.Done()
		events, err := eventStore.FetchPendingEvents(consumerName)
		assert.Nil(t, err)
		err = eventStore.UpdateConsumerOffset(consumerName, events.NewOffset)
		assert.Nil(t, err)

		assert.True(t, len(events.Events) > 0)
		time.Sleep(50 * time.Millisecond) // we wait to be sure that the producer has produce new events
		events, err = eventStore.FetchPendingEvents(consumerName)
		assert.Nil(t, err)
		err = eventStore.UpdateConsumerOffset(consumerName, events.NewOffset)
		assert.Nil(t, err)
		assert.True(t, len(events.Events) > 0)
	}

	go consumFunc(eventStore, consumerNames[0])
	go consumFunc(eventStore, consumerNames[1])
	wg.Wait()
}

func Test_MultipleGetEventsWithoutSettingOffset(t *testing.T) {
	consumerNames := []string{"consumer1"}
	eventStore := exporter.NewEventStore[string](defaultCleanQueueDuration)
	for _, name := range consumerNames {
		eventStore.AddConsumer(name)
	}
	defer eventStore.Stop()

	// start producer
	ctx := context.Background()
	startEventProducer(ctx, eventStore, 100, false)

	firstCall, err := eventStore.FetchPendingEvents(consumerNames[0])
	assert.Nil(t, err)
	secondCall, err := eventStore.FetchPendingEvents(consumerNames[0])
	assert.Nil(t, err)
	assert.Equal(t, firstCall, secondCall)
}

func Test_UpdateWithInvalidOffset(t *testing.T) {
	consumerNames := []string{"consumer1"}
	eventStore := exporter.NewEventStore[string](defaultCleanQueueDuration)
	for _, name := range consumerNames {
		eventStore.AddConsumer(name)
	}
	defer eventStore.Stop()

	// start producer
	ctx := context.Background()
	startEventProducer(ctx, eventStore, 100, false)

	eventList, err := eventStore.FetchPendingEvents(consumerNames[0])
	assert.Nil(t, err)
	errUpdate := eventStore.UpdateConsumerOffset(consumerNames[0], eventList.NewOffset+100)
	assert.NotNil(t, errUpdate)
}

func Test_UpdateWithInvalidConsumerName(t *testing.T) {
	t.Run("should error if calling FetchPendingEvents with invalid consumer name", func(t *testing.T) {
		consumerNames := []string{"consumer1"}
		eventStore := exporter.NewEventStore[string](defaultCleanQueueDuration)
		for _, name := range consumerNames {
			eventStore.AddConsumer(name)
		}
		defer eventStore.Stop()

		// start producer
		ctx := context.Background()
		startEventProducer(ctx, eventStore, 100, false)

		_, err := eventStore.FetchPendingEvents("wrong consumer name")
		assert.NotNil(t, err)
	})

	t.Run("should error if calling UpdateConsumerOffset with invalid consumer name", func(t *testing.T) {
		consumerNames := []string{"consumer1"}
		eventStore := exporter.NewEventStore[string](defaultCleanQueueDuration)
		for _, name := range consumerNames {
			eventStore.AddConsumer(name)
		}
		defer eventStore.Stop()

		// start producer
		ctx := context.Background()
		startEventProducer(ctx, eventStore, 100, false)

		eventList, err := eventStore.FetchPendingEvents(consumerNames[0])
		assert.Nil(t, err)
		errUpdate := eventStore.UpdateConsumerOffset("wrong consumer name", eventList.NewOffset)
		assert.NotNil(t, errUpdate)
	})
}

func Test_WaitForEmptyClean(t *testing.T) {
	consumerNames := []string{"consumer1"}
	eventStore := exporter.NewEventStore[string](defaultCleanQueueDuration)
	for _, name := range consumerNames {
		eventStore.AddConsumer(name)
	}
	defer eventStore.Stop()

	// start producer
	ctx := context.Background()
	startEventProducer(ctx, eventStore, 100, false)
	list, _ := eventStore.FetchPendingEvents(consumerNames[0])
	_ = eventStore.UpdateConsumerOffset(consumerNames[0], list.NewOffset)
	time.Sleep(3 * defaultCleanQueueDuration)
	assert.Equal(t, int64(0), eventStore.GetTotalEventCount())
}

func startEventProducer(ctx context.Context, eventStore exporter.EventStore[string], produceMax int, randomizeProducingTime bool) {
	for i := 0; i < produceMax; i++ {
		select {
		case <-ctx.Done():
			fmt.Println("Goroutine stopped")
			return
		default:
			if randomizeProducingTime {
				randomNumber := rand.Intn(10) + 1
				time.Sleep(time.Duration(randomNumber) * time.Millisecond)
			}
			eventStore.Add("Hello")
		}
	}
}

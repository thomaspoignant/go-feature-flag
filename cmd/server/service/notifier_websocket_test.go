package service_test

import (
	"github.com/thomaspoignant/go-feature-flag/cmd/server/service"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	"github.com/thomaspoignant/go-feature-flag/testutils/testconvert"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/thomaspoignant/go-feature-flag/notifier"
)

type mockWebsocketService struct {
	internalDiff notifier.DiffCache
	nbConnection int
}

func (m *mockWebsocketService) Register(c service.WebsocketConn) {
	m.nbConnection++
}

func (m *mockWebsocketService) Deregister(c service.WebsocketConn) {
	m.nbConnection--
}

func (m *mockWebsocketService) Close() {
	m.nbConnection = 0
}

func (m *mockWebsocketService) BroadcastFlagChanges(diff notifier.DiffCache) {
	m.internalDiff = diff
}

func TestNotify(t *testing.T) {
	// Create a mock WebsocketService
	mockService := &mockWebsocketService{}

	// Create the websocketNotifier instance with the mock service
	n := service.NewWebsocketNotifier(mockService)

	// Prepare the input data
	diff := notifier.DiffCache{
		Deleted: map[string]flag.Flag{
			"flag-1": &flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"A": testconvert.Interface(true),
					"B": testconvert.Interface(false),
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("A"),
				},
			},
		},
		Added: map[string]flag.Flag{
			"flag-2": &flag.InternalFlag{
				Variations: &map[string]*interface{}{
					"A": testconvert.Interface(true),
					"B": testconvert.Interface(false),
				},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("A"),
				},
			},
		},
		Updated: map[string]notifier.DiffUpdated{
			"my-flag": notifier.DiffUpdated{
				Before: &flag.InternalFlag{
					Variations: &map[string]*interface{}{
						"A": testconvert.Interface(true),
						"B": testconvert.Interface(false),
					},
					DefaultRule: &flag.Rule{
						VariationResult: testconvert.String("A"),
					},
				},
				After: &flag.InternalFlag{
					Variations: &map[string]*interface{}{
						"A": testconvert.Interface(true),
						"B": testconvert.Interface(false),
					},
					DefaultRule: &flag.Rule{
						VariationResult: testconvert.String("B"),
					},
				},
			},
		},
	}

	// Create a WaitGroup to ensure the goroutine completes before asserting the results
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)

	// Call the Notify function
	err := n.Notify(diff, waitGroup)

	// Wait for the goroutine to complete
	waitGroup.Wait()

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, diff, mockService.internalDiff)
}

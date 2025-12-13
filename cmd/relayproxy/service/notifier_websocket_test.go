package service_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/modules/core/testutils/testconvert"
	"github.com/thomaspoignant/go-feature-flag/notifier"
)

type mockWebsocketService struct {
	internalDiff notifier.DiffCache
	nbConnection int
}

func (m *mockWebsocketService) Register(_ service.WebsocketConnector) {
	m.nbConnection++
}

func (m *mockWebsocketService) Deregister(_ service.WebsocketConnector) {
	m.nbConnection--
}

func (m *mockWebsocketService) Close() {
	m.nbConnection = 0
}

func (m *mockWebsocketService) WaitForCleanup(_ time.Duration) error {
	// Mock implementation - just return nil immediately
	return nil
}

func (m *mockWebsocketService) BroadcastFlagChanges(diff notifier.DiffCache) {
	m.internalDiff = diff
}

func TestNotify(t *testing.T) {
	// Create a mock WebsocketService
	mockService := &mockWebsocketService{}

	// Create the notifierWebsocket instance with the mock service
	n := service.NewNotifierWebsocket(mockService)

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
			"my-flag": {
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

	// Call the Notify function
	err := n.Notify(diff)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, diff, mockService.internalDiff)
}

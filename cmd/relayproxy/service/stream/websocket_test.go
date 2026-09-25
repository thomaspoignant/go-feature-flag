package stream_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service/stream"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/modules/core/testutils/testconvert"
	"github.com/thomaspoignant/go-feature-flag/notifier"
)

type mockConn struct {
	writeJSONFunc func(v any) error
	throwError    bool
}

func (m *mockConn) WriteJSON(v any) error {
	if m.throwError {
		return fmt.Errorf("error websocket connection")
	}
	if m.writeJSONFunc != nil {
		return m.writeJSONFunc(v)
	}
	return nil
}

func sensitiveDiffCache() notifier.DiffCache {
	beforeQuery := `user.email eq "before@example.com"`
	afterQuery := `user.email eq "after@example.com"`
	enabled := "enabled"
	return notifier.DiffCache{
		Deleted: map[string]flag.Flag{
			"deleted-flag": &flag.InternalFlag{
				Variations: &map[string]*any{"enabled": testconvert.Interface(true)},
				Rules: &[]flag.Rule{
					{Query: &beforeQuery, VariationResult: &enabled},
				},
			},
		},
		Added: map[string]flag.Flag{
			"added-flag": &flag.InternalFlag{
				Variations: &map[string]*any{"enabled": testconvert.Interface(true)},
				Rules: &[]flag.Rule{
					{Query: &afterQuery, VariationResult: &enabled},
				},
			},
		},
		Updated: map[string]notifier.DiffUpdated{
			"updated-flag": {
				Before: &flag.InternalFlag{Rules: &[]flag.Rule{{Query: &beforeQuery}}},
				After:  &flag.InternalFlag{Rules: &[]flag.Rule{{Query: &afterQuery}}},
			},
		},
	}
}

func TestBroadcastFlagChangesPayload(t *testing.T) {
	diff := sensitiveDiffCache()
	tests := []struct {
		name     string
		options  []stream.Option
		expected any
	}{
		{
			name:     "includes flag details by default",
			expected: diff,
		},
		{
			name:    "omits flag details when disabled",
			options: []stream.Option{stream.WithFlagDetails(false)},
			expected: map[string]map[string]struct{}{
				"deleted": {"deleted-flag": {}},
				"added":   {"added-flag": {}},
				"updated": {"updated-flag": {}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			websocketService := stream.NewWebsocketService(tt.options...)
			conn := &mockConn{}
			var received any
			conn.writeJSONFunc = func(v any) error {
				received = v
				return nil
			}
			websocketService.Register(conn)

			websocketService.BroadcastFlagChanges(diff)

			actualJSON, err := json.Marshal(received)
			require.NoError(t, err)
			expectedJSON, err := json.Marshal(tt.expected)
			require.NoError(t, err)
			assert.JSONEq(t, string(expectedJSON), string(actualJSON))
		})
	}
}

func TestBroadcastFlagChanges(t *testing.T) {
	// Create the websocketService instance
	websocketService := stream.NewWebsocketService()

	// Prepare the input data
	diff := notifier.DiffCache{} // You need to define an appropriate DiffCache

	// Create mock connections
	conn1 := &mockConn{}
	conn2 := &mockConn{}

	// Register the mock connections
	websocketService.Register(conn1)
	websocketService.Register(conn2)

	// Set up a flag to track if the WriteJSON function is called on the connections
	conn1WriteJSONCalled := false
	conn2WriteJSONCalled := false

	// Set the function to be executed when WriteJSON is called on the connections
	conn1.writeJSONFunc = func(v any) error {
		conn1WriteJSONCalled = true
		return nil
	}

	conn2.writeJSONFunc = func(v any) error {
		conn2WriteJSONCalled = true
		return nil
	}

	// Call the BroadcastFlagChanges function
	websocketService.BroadcastFlagChanges(diff)

	// Allow some time for the WriteJSON functions to be executed
	time.Sleep(time.Millisecond)

	// Assertions
	assert.True(t, conn1WriteJSONCalled, "WriteJSON should be called on conn1")
	assert.True(t, conn2WriteJSONCalled, "WriteJSON should be called on conn2")
}

func TestDeregister(t *testing.T) {
	// Create the websocketService instance
	websocketService := stream.NewWebsocketService()

	// Create a mock connection
	conn := &mockConn{}

	// Set the function to be executed when WriteJSON is called on the connections
	conn1WriteJSONCalled := false
	conn.writeJSONFunc = func(v any) error {
		conn1WriteJSONCalled = true
		return nil
	}

	// Register the mock connection
	websocketService.Register(conn)

	// Call the Deregister function
	websocketService.Deregister(conn)

	// Call the BroadcastFlagChanges function after deregistering the connection
	diff := notifier.DiffCache{} // You need to define an appropriate DiffCache
	websocketService.BroadcastFlagChanges(diff)

	// Allow some time for the WriteJSON function to be executed
	time.Sleep(time.Millisecond)

	// Assertions
	assert.False(t, conn1WriteJSONCalled, "WriteJSON should not be called after deregistering")
}

func TestClose(t *testing.T) {
	// Create the websocketService instance
	websocketService := stream.NewWebsocketService()

	// Create mock connections
	conn1 := &mockConn{}
	conn2 := &mockConn{}

	// Set up a flag to track if the WriteJSON function is called on the connections
	conn1WriteJSONCalled := false
	conn2WriteJSONCalled := false

	// Set the function to be executed when WriteJSON is called on the connections
	conn1.writeJSONFunc = func(v any) error {
		conn1WriteJSONCalled = true
		return nil
	}

	conn2.writeJSONFunc = func(v any) error {
		conn2WriteJSONCalled = true
		return nil
	}

	// Register the mock connections
	websocketService.Register(conn1)
	websocketService.Register(conn2)

	// Call the Close function
	websocketService.Close()

	// Call the BroadcastFlagChanges function after closing the connections
	diff := notifier.DiffCache{} // You need to define an appropriate DiffCache
	websocketService.BroadcastFlagChanges(diff)

	// Allow some time for the WriteJSON functions to be executed
	time.Sleep(time.Millisecond)

	// Assertions
	assert.False(t, conn1WriteJSONCalled, "WriteJSON should not be called after closing")
	assert.False(t, conn2WriteJSONCalled, "WriteJSON should not be called after closing")
}

func TestBroadcastFlagChangesDeadLock(t *testing.T) {
	// Create the websocketService instance
	websocketService := stream.NewWebsocketService()
	diff := notifier.DiffCache{} // You need to define an appropriate DiffCache
	conn1 := &mockConn{}

	// the mock will return an error when WriteJSON is called
	// this will trigger the deregister of the connection, and the BroadcastFlagChanges will try to lock the mutex
	// in the past we had a deadlock here (see: https://github.com/thomaspoignant/go-feature-flag/issues/3144)
	conn2 := &mockConn{throwError: true}
	websocketService.Register(conn1)
	websocketService.Register(conn2)
	conn1WriteJSONCalled := false
	conn1.writeJSONFunc = func(v any) error {
		conn1WriteJSONCalled = true
		return nil
	}
	websocketService.BroadcastFlagChanges(diff)
	assert.True(t, conn1WriteJSONCalled, "WriteJSON should be called on conn1")

	// We are not testing the error here, we are testing that the function does not deadlock
}

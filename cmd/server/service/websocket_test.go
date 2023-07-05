package service_test

import (
	"github.com/thomaspoignant/go-feature-flag/cmd/server/service"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/thomaspoignant/go-feature-flag/notifier"
)

type mockConn struct {
	writeJSONFunc func(v interface{}) error
}

func (m *mockConn) WriteJSON(v interface{}) error {
	if m.writeJSONFunc != nil {
		return m.writeJSONFunc(v)
	}
	return nil
}

func TestBroadcastFlagChanges(t *testing.T) {
	// Create the websocketService instance
	websocketService := service.NewWebsocketService()

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
	conn1.writeJSONFunc = func(v interface{}) error {
		conn1WriteJSONCalled = true
		return nil
	}

	conn2.writeJSONFunc = func(v interface{}) error {
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
	websocketService := service.NewWebsocketService()

	// Create a mock connection
	conn := &mockConn{}

	// Set the function to be executed when WriteJSON is called on the connections
	conn1WriteJSONCalled := false
	conn.writeJSONFunc = func(v interface{}) error {
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
	websocketService := service.NewWebsocketService()

	// Create mock connections
	conn1 := &mockConn{}
	conn2 := &mockConn{}

	// Set up a flag to track if the WriteJSON function is called on the connections
	conn1WriteJSONCalled := false
	conn2WriteJSONCalled := false

	// Set the function to be executed when WriteJSON is called on the connections
	conn1.writeJSONFunc = func(v interface{}) error {
		conn1WriteJSONCalled = true
		return nil
	}

	conn2.writeJSONFunc = func(v interface{}) error {
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

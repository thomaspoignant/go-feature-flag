package service_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
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
	websocketService := service.NewWebsocketService()

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
	websocketService := service.NewWebsocketService()

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
	websocketService := service.NewWebsocketService()
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

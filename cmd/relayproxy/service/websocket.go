package service

import (
	"context"
	"sync"
	"time"

	"github.com/thomaspoignant/go-feature-flag/notifier"
)

// WebsocketConnector is an interface to be able to mock websocket.Conn
type WebsocketConnector interface {
	WriteJSON(v any) error
}

// WebsocketService is the service interface that handle the websocket connections
// This service is able to broadcast a notification to all the open websockets
type WebsocketService interface {
	// Register is adding the connection to the list of open connection.
	Register(c WebsocketConnector)
	// Deregister is removing the connection from the list of open connection.
	Deregister(c WebsocketConnector)
	// BroadcastFlagChanges is sending the diff cache struct to the client.
	BroadcastFlagChanges(diff notifier.DiffCache)
	// Close deregister all open connections.
	Close()
	// WaitForCleanup waits for all connections to be properly closed
	WaitForCleanup(timeout time.Duration) error
}

// NewWebsocketService is a constructor to create a new WebsocketService.
func NewWebsocketService() WebsocketService {
	return &websocketServiceImpl{
		clients: map[WebsocketConnector]any{},
		mutex:   &sync.RWMutex{},
		closed:  make(chan struct{}),
	}
}

// websocketServiceImpl is the implementation of the interface.
type websocketServiceImpl struct {
	clients map[WebsocketConnector]any
	mutex   *sync.RWMutex
	closed  chan struct{}
}

// BroadcastFlagChanges is sending a string to all the open connection.
func (w *websocketServiceImpl) BroadcastFlagChanges(diff notifier.DiffCache) {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	for c := range w.clients {
		err := c.WriteJSON(diff)
		if err != nil {
			w.mutex.RUnlock()
			w.Deregister(c)
			w.mutex.RLock()
		}
	}
}

// Register is adding the connection to the list of open connection.
func (w *websocketServiceImpl) Register(c WebsocketConnector) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.clients[c] = struct{}{}
}

// Deregister is removing the connection from the list of open connection.
func (w *websocketServiceImpl) Deregister(c WebsocketConnector) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	delete(w.clients, c)
}

// Close deregister all open connections.
func (w *websocketServiceImpl) Close() {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	for c := range w.clients {
		delete(w.clients, c)
	}
	close(w.closed)
}

// WaitForCleanup waits for all connections to be properly closed
func (w *websocketServiceImpl) WaitForCleanup(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	select {
	case <-w.closed:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

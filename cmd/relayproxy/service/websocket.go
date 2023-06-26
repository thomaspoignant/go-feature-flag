package service

import (
	"github.com/gorilla/websocket"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"sync"
)

// WebsocketService is the service interface that handle the websocket connections
// This service is able to broadcast a notification to all the open websockets
type WebsocketService interface {
	// Register is adding the connection to the list of open connection.
	Register(c *websocket.Conn)
	// Deregister is removing the connection from the list of open connection.
	Deregister(c *websocket.Conn)
	// BroadcastFlagChanges is sending the diff cache struct to the client.
	BroadcastFlagChanges(diff notifier.DiffCache)
	// Close deregister all open connections.
	Close()
}

// NewWebsocketService is a constructor to create a new WebsocketService.
func NewWebsocketService() WebsocketService {
	return &websocketServiceImpl{
		clients: map[*websocket.Conn]interface{}{},
		mutex:   &sync.RWMutex{},
	}
}

// websocketServiceImpl is the implementation of the interface.
type websocketServiceImpl struct {
	clients map[*websocket.Conn]interface{}
	mutex   *sync.RWMutex
}

// BroadcastFlagChanges is sending a string to all the open connection.
func (w *websocketServiceImpl) BroadcastFlagChanges(diff notifier.DiffCache) {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	for c := range w.clients {
		err := c.WriteJSON(diff)
		if err != nil {
			w.Deregister(c)
		}
	}
}

// Register is adding the connection to the list of open connection.
func (w *websocketServiceImpl) Register(c *websocket.Conn) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.clients[c] = struct{}{}
}

// Deregister is removing the connection from the list of open connection.
func (w *websocketServiceImpl) Deregister(c *websocket.Conn) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	delete(w.clients, c)
}

// Close deregister all open connections.
func (w *websocketServiceImpl) Close() {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	for c := range w.clients {
		w.Deregister(c)
	}
}

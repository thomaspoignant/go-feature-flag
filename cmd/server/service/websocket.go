package service

import (
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"sync"
)

// WebsocketConn is an interface to be able to mock websocket.Conn
type WebsocketConn interface {
	WriteJSON(v interface{}) error
}

// WebsocketService is the service interface that handle the websocket connections
// This service is able to broadcast a notification to all the open websockets
type WebsocketService interface {
	// Register is adding the connection to the list of open connection.
	Register(c WebsocketConn)
	// Deregister is removing the connection from the list of open connection.
	Deregister(c WebsocketConn)
	// BroadcastFlagChanges is sending the diff cache struct to the client.
	BroadcastFlagChanges(diff notifier.DiffCache)
	// Close deregister all open connections.
	Close()
}

// NewWebsocketService is a constructor to create a new WebsocketService.
func NewWebsocketService() WebsocketService {
	return &websocketServiceImpl{
		clients: map[WebsocketConn]interface{}{},
		mutex:   &sync.RWMutex{},
	}
}

// websocketServiceImpl is the implementation of the interface.
type websocketServiceImpl struct {
	clients map[WebsocketConn]interface{}
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
func (w *websocketServiceImpl) Register(c WebsocketConn) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.clients[c] = struct{}{}
}

// Deregister is removing the connection from the list of open connection.
func (w *websocketServiceImpl) Deregister(c WebsocketConn) {
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
}

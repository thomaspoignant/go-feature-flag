package service

import (
	"github.com/gorilla/websocket"
)

// WebsocketService is the service interface that handle the websocket connections
// This service is able to broadcast a notification to all the open websockets
type WebsocketService interface {
	// Register is adding the connection to the list of open connection.
	Register(c *websocket.Conn)
	// Deregister is removing the connection from the list of open connection.
	Deregister(c *websocket.Conn)
	// BroadcastText is sending a string to all the open connection.
	BroadcastText(s string)
}

// NewWebsocketService is a constructor to create a new WebsocketService.
func NewWebsocketService() WebsocketService {
	return &websocketServiceImpl{
		clients: map[*websocket.Conn]interface{}{},
	}
}

// websocketServiceImpl is the implementation of the interface.
type websocketServiceImpl struct {
	clients map[*websocket.Conn]interface{}
}

// BroadcastText is sending a string to all the open connection.
func (w *websocketServiceImpl) BroadcastText(_ string) {
	for c := range w.clients {
		err := c.WriteMessage(websocket.TextMessage, []byte("toto"))
		if err != nil {
			w.Deregister(c)
		}
	}
}

// Register is adding the connection to the list of open connection.
func (w *websocketServiceImpl) Register(c *websocket.Conn) {
	w.clients[c] = struct{}{}
}

// Deregister is removing the connection from the list of open connection.
func (w *websocketServiceImpl) Deregister(c *websocket.Conn) {
	delete(w.clients, c)
}

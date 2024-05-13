package api

import (
	custommiddleware "github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/api/middleware"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/controller"
)

func (s *Server) addWebsocketRoutes() {
	// initWebsocketsEndpoints initialize the websocket endpoints
	cFlagReload := controller.NewWsFlagChange(s.services.WebsocketService, s.zapLog)
	wsV1 := s.apiEcho.Group("/ws/v1")
	wsV1.Use(custommiddleware.WebsocketAuthorizer(s.config))
	wsV1.GET("/flag/change", cFlagReload.Handler)
}

package api

import (
	"github.com/labstack/echo/v4"
	custommiddleware "github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/api/middleware"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/controller"
)

func (s *Server) InitWebsocketRoutes(echoInstance *echo.Echo) {
	// initWebsocketsEndpoints initialize the websocket endpoints
	cFlagReload := controller.NewWsFlagChange(s.services.WebsocketService, s.zapLog)
	wsV1 := echoInstance.Group("/ws/v1")
	wsV1.Use(custommiddleware.WebsocketAuthorizer(s.config))
	wsV1.GET("/flag/change", cFlagReload.Handler)
}

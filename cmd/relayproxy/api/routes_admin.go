package api

import (
	"github.com/labstack/echo/v4"
	controller "github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/handler/goff"
)

func (s *Server) addAdminRoutes(cRetrieverRefresh controller.Controller, authMiddleware echo.MiddlewareFunc) {
	adminGrp := s.apiEcho.Group("/admin/v1")
	adminGrp.Use(authMiddleware)
	adminGrp.POST("/retriever/refresh", cRetrieverRefresh.Handler)
}

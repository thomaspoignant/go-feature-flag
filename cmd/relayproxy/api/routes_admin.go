package api

import (
	"github.com/labstack/echo/v4"
	middleware2 "github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/api/middleware"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/controller"
)

func (s *Server) addAdminRoutes(cRetrieverRefresh controller.Controller) {
	adminGrp := s.apiEcho.Group("/admin/v1")
	adminGrp.Use(middleware2.KeyAuthExtended(middleware2.KeyAuthExtendedConfig{
		Validator: func(key string, _ echo.Context) (bool, error) {
			return s.config.APIKeysAdminExists(key), nil
		},
	}))
	adminGrp.POST("/retriever/refresh", cRetrieverRefresh.Handler)
}

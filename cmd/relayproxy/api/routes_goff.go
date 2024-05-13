package api

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/controller"
)

func (s *Server) addGOFFRoutes(
	cAllFlags controller.Controller,
	cFlagEval controller.Controller,
	cEvalDataCollector controller.Controller) {
	// Grouping the routes
	v1 := s.apiEcho.Group("/v1")
	// nolint: staticcheck
	if len(s.config.AuthorizedKeys.Evaluation) > 0 || len(s.config.APIKeys) > 0 {
		v1.Use(middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
			Validator: func(key string, _ echo.Context) (bool, error) {
				return s.config.APIKeyExists(key), nil
			},
		}))
	}
	v1.POST("/allflags", cAllFlags.Handler)
	v1.POST("/feature/:flagKey/eval", cFlagEval.Handler)
	v1.POST("/data/collector", cEvalDataCollector.Handler)

	// Swagger - only available if option is enabled
	if s.config.EnableSwagger {
		s.apiEcho.GET("/swagger/*", echoSwagger.WrapHandler)
	}
}

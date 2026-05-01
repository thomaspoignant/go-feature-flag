package api

import (
	"github.com/labstack/echo/v4"
	etag "github.com/pablor21/echo-etag/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
	controller "github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/handler/goff"
)

func (s *Server) addGOFFRoutes(
	cAllFlags,
	cFlagEval,
	cEvalDataCollector,
	cFlagChange,
	cFlagConfiguration controller.Controller,
	authMiddleware echo.MiddlewareFunc,
) {
	// Grouping the routes
	v1 := s.apiEcho.Group("/v1")
	v1.Use(authMiddleware)
	v1.Use(etag.WithConfig(etag.Config{
		Skipper: func(c echo.Context) bool {
			switch c.Path() {
			case
				"/v1/flag/change",
				"/v1/flag/configuration":
				return false
			default:
				return true
			}
		},
		Weak: false,
	}))

	v1.POST("/allflags", cAllFlags.Handler)
	v1.POST("/feature/:flagKey/eval", cFlagEval.Handler)
	v1.POST("/data/collector", cEvalDataCollector.Handler)
	v1.GET("/flag/change", cFlagChange.Handler)
	v1.POST("/flag/configuration", cFlagConfiguration.Handler)

	// Swagger - only available if option is enabled
	if s.config.IsSwaggerEnabled() {
		s.apiEcho.GET("/swagger/*", echoSwagger.WrapHandler)
	}
}

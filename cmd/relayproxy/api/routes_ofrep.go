package api

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	etag "github.com/pablor21/echo-etag/v4"
	middleware2 "github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/api/middleware"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/ofrep"
)

func (s *Server) addOFREPRoutes(cFlagEvalOFREP ofrep.EvaluateCtrl) {
	ofrepGroup := s.apiEcho.Group("/ofrep/v1")
	ofrepGroup.Use(etag.WithConfig(etag.Config{
		Skipper: func(c echo.Context) bool {
			switch c.Path() {
			case "/ofrep/v1/evaluate/flags", "/ofrep/v1/configuration":
				return false
			default:
				return true
			}
		},
		Weak: false,
	}))

	// nolint: staticcheck
	if s.config.IsAuthenticationEnabled() {
		ofrepGroup.Use(middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
			Validator: func(key string, _ echo.Context) (bool, error) {
				return s.config.APIKeyExists(key), nil
			},
			ErrorHandler: middleware2.AuthMiddlewareErrHandler,
		}))
	}
	ofrepGroup.POST("/evaluate/flags", cFlagEvalOFREP.BulkEvaluate)
	ofrepGroup.POST("/evaluate/flags/:flagKey", cFlagEvalOFREP.Evaluate)
}

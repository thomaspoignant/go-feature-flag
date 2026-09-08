package api

import (
	"github.com/labstack/echo/v4"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/handler/ofrep"
	helpermiddleware "github.com/thomaspoignant/go-feature-flag/cmdhelpers/api/middleware"
)

func (s *Server) addOFREPRoutes(cFlagEvalOFREP ofrep.EvaluateCtrl, authMiddleware echo.MiddlewareFunc) {
	ofrepGroup := s.apiEcho.Group("/ofrep/v1")
	ofrepGroup.Use(helpermiddleware.EtagWithConfig(helpermiddleware.EtagConfig{
		Skipper: func(c echo.Context) bool {
			switch c.Path() {
			case "/ofrep/v1/evaluate/flags":
				return false
			default:
				return true
			}
		},
		Weak: false,
	}))

	ofrepGroup.Use(authMiddleware)
	ofrepGroup.POST("/evaluate/flags", cFlagEvalOFREP.BulkEvaluate)
	ofrepGroup.POST("/evaluate/flags/:flagKey", cFlagEvalOFREP.Evaluate)
}

package api

import (
	"github.com/labstack/echo/v4"
	etag "github.com/pablor21/echo-etag/v4"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/handler/ofrep"
)

func (s *Server) addOFREPRoutes(cFlagEvalOFREP ofrep.EvaluateCtrl, authMiddleware echo.MiddlewareFunc) {
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

	ofrepGroup.Use(authMiddleware)
	ofrepGroup.POST("/evaluate/flags", cFlagEvalOFREP.BulkEvaluate)
	ofrepGroup.POST("/evaluate/flags/:flagKey", cFlagEvalOFREP.Evaluate)
}

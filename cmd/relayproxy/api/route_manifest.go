package api

import (
	"github.com/labstack/echo/v4"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/handler/manifest"
)

func (s *Server) addManifestRoutes(cManifest manifest.ManifestCtrl, authMiddleware echo.MiddlewareFunc) {
	manifestGroup := s.apiEcho.Group("/openfeature/v0")
	manifestGroup.Use(authMiddleware)
	manifestGroup.GET("/manifest", cManifest.GetManifest)
}

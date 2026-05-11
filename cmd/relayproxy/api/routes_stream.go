package api

import (
	"github.com/labstack/echo/v4"
	custommiddleware "github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/api/middleware"
	controller "github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/handler/goff"
)

// addStreamRoutes registers all transports that push flag-change events to
// clients (websocket + SSE) under the /stream/v1 group. The legacy
// /ws/v1/flag/change route is kept for backward compatibility but marked
// deprecated via response headers.
func (s *Server) addStreamRoutes() {
	authorize := custommiddleware.StreamAuthorizer(s.config)

	cWsFlagChange := controller.NewWsFlagChange(s.services.WebsocketService, s.zapLog)
	cSSEFlagChange := controller.NewSSEFlagChange(s.services.SSEService, s.services.FlagsetManager, s.zapLog)

	streamV1 := s.apiEcho.Group("/stream/v1", authorize)
	streamV1.GET("/ws/flag/change", cWsFlagChange.Handler)
	streamV1.GET("/sse/flag/change", cSSEFlagChange.Handler)

	// Legacy alias - kept for backward compatibility, marked deprecated.
	s.apiEcho.GET(
		"/ws/v1/flag/change",
		cWsFlagChange.LegacyHandler,
		authorize,
		deprecatedAlias("/stream/v1/ws/flag/change"),
	)
}

// deprecatedAlias adds RFC 8594 Deprecation + Link headers pointing clients to
// the new endpoint. We do not set Sunset until the removal date is decided.
func deprecatedAlias(replacement string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Deprecation", "true")
			c.Response().Header().Set(
				"Link",
				`<`+replacement+`>; rel="successor-version"`,
			)
			return next(c)
		}
	}
}

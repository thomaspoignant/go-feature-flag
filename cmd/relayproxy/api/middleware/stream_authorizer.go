package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
)

// StreamAuthorizer is a middleware that checks the apiKey query param to authorize
// streaming connections (websocket / SSE).
func StreamAuthorizer(config *config.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			apiKey := c.QueryParam("apiKey")
			if config.IsAuthenticationEnabled() && !config.APIKeyExists(apiKey) {
				return echo.ErrUnauthorized
			}
			return next(c)
		}
	}
}

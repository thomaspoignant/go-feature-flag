package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
)

// WebsocketAuthorizer is a middleware that checks in the params if we have the needed parameter for authorization
func WebsocketAuthorizer(config *config.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if len(config.APIKeys) > 0 {
				apiKey := c.QueryParam("apiKey")
				if !config.APIKeyExists(apiKey) {
					return echo.ErrUnauthorized
				}
			}
			return next(c)
		}
	}
}

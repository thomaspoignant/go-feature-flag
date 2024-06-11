package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
)

// VersionHeader is a middleware that adds the version of the relayproxy in the header
func VersionHeader(config *config.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("X-GOFEATUREFLAG-VERSION", config.Version)
			return next(c)
		}
	}
}

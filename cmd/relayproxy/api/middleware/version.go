package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
)

// VersionHeaderConfig defines the configuration for the middleware.
type VersionHeaderConfig struct {
	Skipper          middleware.Skipper
	RelayProxyConfig *config.Config
}

// VersionHeader is a middleware that adds the version of the relayproxy in the header.
func VersionHeader(cfg VersionHeaderConfig) echo.MiddlewareFunc {
	// Use provided skipper or fallback to the default skipper.
	skipper := cfg.Skipper
	if skipper == nil {
		skipper = DefaultVersionHeaderSkipper
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if skipper(c) {
				return next(c)
			}

			c.Response().Header().Set("X-GOFEATUREFLAG-VERSION", cfg.RelayProxyConfig.Version)
			return next(c)
		}
	}
}

func DefaultVersionHeaderSkipper(_ echo.Context) bool {
	return false
}

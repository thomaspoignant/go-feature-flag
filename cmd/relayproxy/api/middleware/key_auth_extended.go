package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// KeyAuthExtendedConfig defines the configuration for the extended key auth middleware.
// It extends middleware.KeyAuthConfig to also check the X-API-Key header.
type KeyAuthExtendedConfig struct {
	// Validator is a function to validate the API key.
	// It receives the key and the context, and returns true if the key is valid.
	Validator middleware.KeyAuthValidator

	// ErrorHandler is a function to handle errors when authentication fails.
	ErrorHandler middleware.KeyAuthErrorHandler

	// Skipper defines a function to skip middleware.
	Skipper middleware.Skipper

	// KeyLookup is a string in the form of "<source>:<name>" that is used
	// to extract key from the request. Optional. Default value "header:Authorization".
	// Possible values:
	// - "header:<name>" or "header:<name>:<cut-prefix>" - extracts key from the request header
	// - "query:<name>" - extracts key from the query string
	// - "form:<name>" - extracts key from the form
	// - "cookie:<name>" - extracts key from the cookie
	KeyLookup string
}

// KeyAuthExtended is a middleware that extends middleware.KeyAuthWithConfig
// to also check the X-API-Key header in addition to the standard Authorization header.
func KeyAuthExtended(config KeyAuthExtendedConfig) echo.MiddlewareFunc {
	// Set default values
	if config.Validator == nil {
		panic("echo: key auth extended middleware requires a validator function")
	}
	if config.ErrorHandler == nil {
		config.ErrorHandler = middleware.DefaultKeyAuthConfig.ErrorHandler
	}
	if config.Skipper == nil {
		config.Skipper = middleware.DefaultKeyAuthConfig.Skipper
	}
	if config.KeyLookup == "" {
		config.KeyLookup = middleware.DefaultKeyAuthConfig.KeyLookup
	}

	// Create the standard KeyAuth middleware for Authorization header
	keyAuthMiddleware := middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
		Validator:    config.Validator,
		ErrorHandler: config.ErrorHandler,
		Skipper:      config.Skipper,
		KeyLookup:    config.KeyLookup,
	})

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Check if we should skip this middleware
			if config.Skipper(c) {
				return next(c)
			}

			// First, check X-API-Key header
			xAPIKey := c.Request().Header.Get("X-API-Key")
			if xAPIKey != "" {
				valid, err := config.Validator(xAPIKey, c)
				if err != nil {
					return config.ErrorHandler(err, c)
				}
				if valid {
					return next(c)
				}
				// If X-API-Key is present but invalid, return error
				return config.ErrorHandler(echo.ErrUnauthorized, c)
			}

			// If X-API-Key is not present, fall back to standard KeyAuth behavior
			return keyAuthMiddleware(next)(c)
		}
	}
}

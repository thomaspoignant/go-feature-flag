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

// setDefaults sets default values for the KeyAuthExtendedConfig.
func setDefaults(config *KeyAuthExtendedConfig) {
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
}

// validateXAPIKey validates the X-API-Key header if present.
// Returns (handled, error) where handled indicates if X-API-Key was processed.
// If handled is true and error is nil, the request should continue to next handler.
// If handled is true and error is not nil, the error should be returned.
// If handled is false, fall back to standard KeyAuth middleware.
func validateXAPIKey(c echo.Context, config KeyAuthExtendedConfig, next echo.HandlerFunc) (bool, error) {
	xAPIKey := c.Request().Header.Get("X-API-Key")
	if xAPIKey == "" {
		return false, nil // X-API-Key not present, fall back to standard middleware
	}

	valid, err := config.Validator(xAPIKey, c)
	if err != nil {
		return true, config.ErrorHandler(err, c) // X-API-Key present but validation error
	}
	if !valid {
		return true, config.ErrorHandler(echo.ErrUnauthorized, c) // X-API-Key present but invalid
	}

	// X-API-Key is valid, continue to next handler
	return true, next(c)
}

// KeyAuthExtended is a middleware that extends middleware.KeyAuthWithConfig
// to also check the X-API-Key header in addition to the standard Authorization header.
func KeyAuthExtended(config KeyAuthExtendedConfig) echo.MiddlewareFunc {
	// Set default values
	setDefaults(&config)

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
			handled, err := validateXAPIKey(c, config, next)
			if handled {
				return err
			}

			// If X-API-Key is not present, fall back to standard KeyAuth behavior
			return keyAuthMiddleware(next)(c)
		}
	}
}

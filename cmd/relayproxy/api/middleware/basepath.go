package middleware

import (
	"strings"

	"github.com/labstack/echo/v4"
)

// BasePathConfig holds the configuration for the base path middleware
type BasePathConfig struct {
	// Skipper defines a function to skip middleware
	Skipper func(c echo.Context) bool

	// BasePath is the base path to strip from incoming requests
	BasePath string
}

// BasePathMiddleware creates a middleware that strips a base path from incoming requests.
// This is useful for AWS API Gateway deployments where requests come with a base path prefix.
func BasePathMiddleware(config BasePathConfig) echo.MiddlewareFunc {
	// If no base path is configured, return a no-op middleware
	if config.BasePath == "" {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return next
		}
	}

	// Normalize the base path - ensure it starts with / and doesn't end with /
	basePath := strings.TrimSuffix(config.BasePath, "/")
	if !strings.HasPrefix(basePath, "/") {
		basePath = "/" + basePath
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip middleware if skipper returns true
			if config.Skipper != nil && config.Skipper(c) {
				return next(c)
			}

			req := c.Request()
			path := req.URL.Path

			// If the request path starts with the base path, strip it
			if strings.HasPrefix(path, basePath) {
				// Strip the base path from the URL path
				strippedPath := strings.TrimPrefix(path, basePath)

				// Ensure the stripped path starts with /
				if strippedPath == "" || !strings.HasPrefix(strippedPath, "/") {
					strippedPath = "/" + strippedPath
				}

				// Update the request URL path
				req.URL.Path = strippedPath
			}

			return next(c)
		}
	}
}

package middleware

import (
	"fmt"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// DefaultSkipper is what we use as a default.
// Some endpoints are excluded from the logs to avoid flooding the logs and
// because they are not bringing a lot of value.
func DefaultSkipper(c echo.Context) bool {
	skipperURL := []string{"/health", "/info", "/metrics"}
	for _, ignoredPath := range skipperURL {
		if strings.HasPrefix(ignoredPath, c.Request().URL.String()) {
			return true
		}
	}
	return false
}

// DebugSkipper is the skipper used in debug mode, we log everything.
func DebugSkipper(_ echo.Context) bool {
	return false
}

// ZapLogger is a middleware and zap to provide an "access log" like logging for each request.
func ZapLogger(log *zap.Logger, cfg *config.Config) echo.MiddlewareFunc {
	// select the right skipper
	skipper := DefaultSkipper
	if cfg != nil && cfg.IsDebugEnabled() {
		skipper = DebugSkipper
	}

	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		Skipper: skipper,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			req := c.Request()
			res := c.Response()

			fields := make([]zapcore.Field, 0, 8)
			fields = append(fields,
				zap.String("remote_ip", c.RealIP()),
				zap.String("latency", time.Since(v.StartTime).String()),
				zap.String("host", req.Host),
				zap.String("request", fmt.Sprintf("%s %s", req.Method, req.RequestURI)),
				zap.Int("status", res.Status),
				zap.Int64("size", res.Size),
				zap.String("user_agent", req.UserAgent()),
			)

			id := req.Header.Get(echo.HeaderXRequestID)
			if id == "" {
				id = res.Header().Get(echo.HeaderXRequestID)
			}
			fields = append(fields, zap.String("request_id", id))

			n := res.Status
			switch {
			case n >= 500:
				log.With(zap.Error(v.Error)).Error("Server error", fields...)
			case n >= 400:
				log.With(zap.Error(v.Error)).Warn("Client error", fields...)
			case n >= 300:
				log.Debug("Redirection", fields...)
			default:
				log.Debug("Success", fields...)
			}
			return nil
		},
	})
}

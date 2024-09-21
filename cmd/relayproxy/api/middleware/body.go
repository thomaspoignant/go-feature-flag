package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"go.uber.org/zap"
)

func BodyLogger(log *zap.Logger, cfg *config.Config) echo.MiddlewareFunc {
	skipper := func(_ echo.Context) bool { return !cfg.IsDebugEnabled() }

	return middleware.BodyDumpWithConfig(middleware.BodyDumpConfig{
		Skipper: skipper,
		Handler: func(c echo.Context, req []byte, _ []byte) {
			log.Debug("Request", zap.ByteString("body", req))
		},
	})
}

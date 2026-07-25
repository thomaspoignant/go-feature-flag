package api

import (
	"github.com/labstack/echo/v4"
	controller "github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/handler/goff"
)

// addFlagManagementRoutes registers the /v1/flags flag-management API. It reuses the Admin
// auth middleware: writes are further restricted at the handler level to flagsets backed by
// exactly one writable (PostgreSQL) retriever.
func (s *Server) addFlagManagementRoutes(
	cListFlags controller.Controller,
	cCreateFlag controller.Controller,
	cGetFlag controller.Controller,
	cReplaceFlag controller.Controller,
	cPatchFlag controller.Controller,
	cDeleteFlag controller.Controller,
	cSetFlagState controller.Controller,
	authMiddleware echo.MiddlewareFunc,
) {
	flagsGrp := s.apiEcho.Group("/v1/flags")
	flagsGrp.Use(authMiddleware)
	flagsGrp.GET("", cListFlags.Handler)
	flagsGrp.POST("", cCreateFlag.Handler)
	flagsGrp.GET("/:flag_key", cGetFlag.Handler)
	flagsGrp.PUT("/:flag_key", cReplaceFlag.Handler)
	flagsGrp.PATCH("/:flag_key", cPatchFlag.Handler)
	flagsGrp.DELETE("/:flag_key", cDeleteFlag.Handler)
	flagsGrp.PATCH("/:flag_key/state", cSetFlagState.Handler)
}

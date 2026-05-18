package api

import (
	"github.com/labstack/echo/v4"

	"github.com/thomaspoignant/go-feature-flag/cmd/management/api/middleware"
	"github.com/thomaspoignant/go-feature-flag/cmd/management/model"
)

func registerVersionRoutes(g *echo.Group, h Handlers, s Services) {
	v := g.Group("/flags/:id/versions")
	v.GET("", h.Versions.List, middleware.RequireFlagRole(s.Teams, s.Flagsets, s.Flags, "id", model.RoleViewer))
	v.GET("/:n", h.Versions.Get, middleware.RequireFlagRole(s.Teams, s.Flagsets, s.Flags, "id", model.RoleViewer))
	v.POST("/:n/rollback", h.Versions.Rollback, middleware.RequireFlagRole(s.Teams, s.Flagsets, s.Flags, "id", model.RoleEditor))
}

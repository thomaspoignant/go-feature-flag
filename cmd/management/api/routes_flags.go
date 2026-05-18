package api

import (
	"github.com/labstack/echo/v4"

	"github.com/thomaspoignant/go-feature-flag/cmd/management/api/middleware"
	"github.com/thomaspoignant/go-feature-flag/cmd/management/model"
)

func registerFlagRoutes(g *echo.Group, h Handlers, s Services) {
	g.GET("/flagsets/:flagsetId/flags", h.Flags.List,
		middleware.RequireFlagsetRole(s.Teams, s.Flagsets, "flagsetId", model.RoleViewer))
	g.POST("/flagsets/:flagsetId/flags", h.Flags.Create,
		middleware.RequireFlagsetRole(s.Teams, s.Flagsets, "flagsetId", model.RoleEditor))

	f := g.Group("/flags/:id")
	f.GET("", h.Flags.Get, middleware.RequireFlagRole(s.Teams, s.Flagsets, s.Flags, "id", model.RoleViewer))
	f.PUT("", h.Flags.Update, middleware.RequireFlagRole(s.Teams, s.Flagsets, s.Flags, "id", model.RoleEditor))
	f.POST("/disable", h.Flags.Disable, middleware.RequireFlagRole(s.Teams, s.Flagsets, s.Flags, "id", model.RoleEditor))
	f.DELETE("", h.Flags.Delete, middleware.RequireFlagRole(s.Teams, s.Flagsets, s.Flags, "id", model.RoleAdmin))
}

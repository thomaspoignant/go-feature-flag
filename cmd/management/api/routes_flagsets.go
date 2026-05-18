package api

import (
	"github.com/labstack/echo/v4"

	"github.com/thomaspoignant/go-feature-flag/cmd/management/api/middleware"
	"github.com/thomaspoignant/go-feature-flag/cmd/management/model"
)

func registerFlagsetRoutes(g *echo.Group, h Handlers, s Services) {
	g.GET("/teams/:teamId/flagsets", h.Flagsets.ListByTeam,
		middleware.RequireTeamRole(s.Teams, "teamId", model.RoleViewer))
	g.POST("/teams/:teamId/flagsets", h.Flagsets.Create,
		middleware.RequireTeamRole(s.Teams, "teamId", model.RoleEditor))

	fs := g.Group("/flagsets/:id")
	fs.GET("", h.Flagsets.Get, middleware.RequireFlagsetRole(s.Teams, s.Flagsets, "id", model.RoleViewer))
	fs.PATCH("", h.Flagsets.Update, middleware.RequireFlagsetRole(s.Teams, s.Flagsets, "id", model.RoleEditor))
	fs.DELETE("", h.Flagsets.Delete, middleware.RequireFlagsetRole(s.Teams, s.Flagsets, "id", model.RoleAdmin))
	fs.POST("/api-keys", h.Flagsets.CreateAPIKey, middleware.RequireFlagsetRole(s.Teams, s.Flagsets, "id", model.RoleAdmin))
	fs.DELETE("/api-keys/:keyHash", h.Flagsets.DeleteAPIKey, middleware.RequireFlagsetRole(s.Teams, s.Flagsets, "id", model.RoleAdmin))
}

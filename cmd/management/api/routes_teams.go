package api

import (
	"github.com/labstack/echo/v4"

	"github.com/thomaspoignant/go-feature-flag/cmd/management/api/middleware"
	"github.com/thomaspoignant/go-feature-flag/cmd/management/model"
)

func registerTeamRoutes(g *echo.Group, h Handlers, s Services) {
	teams := g.Group("/teams")
	teams.GET("", h.Teams.List)
	teams.POST("", h.Teams.Create, middleware.RequireSuperAdmin())

	scoped := teams.Group("/:id")
	scoped.GET("", h.Teams.Get, middleware.RequireTeamRole(s.Teams, "id", model.RoleViewer))
	scoped.PATCH("", h.Teams.Update, middleware.RequireTeamRole(s.Teams, "id", model.RoleAdmin))
	scoped.DELETE("", h.Teams.Delete, middleware.RequireSuperAdmin())

	scoped.GET("/members", h.Teams.ListMembers, middleware.RequireTeamRole(s.Teams, "id", model.RoleViewer))
	scoped.POST("/members", h.Teams.AddMember, middleware.RequireTeamRole(s.Teams, "id", model.RoleAdmin))
	scoped.PATCH("/members/:uid", h.Teams.UpdateMember, middleware.RequireTeamRole(s.Teams, "id", model.RoleAdmin))
	scoped.DELETE("/members/:uid", h.Teams.RemoveMember, middleware.RequireTeamRole(s.Teams, "id", model.RoleAdmin))
}

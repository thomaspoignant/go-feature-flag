package api

import (
	"github.com/labstack/echo/v4"
)

func registerOnboardingRoutes(g *echo.Group, h Handlers) {
	ob := g.Group("/onboarding")
	ob.POST("/team", h.Onboarding.CreateTeam)
}

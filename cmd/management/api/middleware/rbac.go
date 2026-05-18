package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/thomaspoignant/go-feature-flag/cmd/management/model"
	"github.com/thomaspoignant/go-feature-flag/cmd/management/service"
)

func RequireSuperAdmin() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims := MustClaims(c)
			if claims == nil || !claims.IsSuperAdmin {
				return c.JSON(http.StatusForbidden, model.APIResponse{Success: false, Message: "super admin required"})
			}
			return next(c)
		}
	}
}

// RequireTeamRole gates routes that have :teamId or :id (interpreted as team) in path.
// paramName is the path-param holding the team id.
func RequireTeamRole(teams *service.TeamService, paramName string, min model.Role) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims := MustClaims(c)
			if claims == nil {
				return c.JSON(http.StatusUnauthorized, model.APIResponse{Success: false, Message: "unauthenticated"})
			}
			if claims.IsSuperAdmin {
				return next(c)
			}
			teamID := c.Param(paramName)
			if teamID == "" {
				return c.JSON(http.StatusBadRequest, model.APIResponse{Success: false, Message: "missing team id"})
			}
			role, err := teams.Role(c.Request().Context(), teamID, claims.UserID)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, model.APIResponse{Success: false, Message: "failed to load membership"})
			}
			if role == "" || !role.AtLeast(min) {
				return c.JSON(http.StatusForbidden, model.APIResponse{Success: false, Message: "insufficient role"})
			}
			return next(c)
		}
	}
}

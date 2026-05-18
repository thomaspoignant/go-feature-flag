package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/thomaspoignant/go-feature-flag/cmd/management/model"
	"github.com/thomaspoignant/go-feature-flag/cmd/management/service"
)

// RequireFlagsetRole resolves the team owning a flagset (param `paramName`) and checks role.
func RequireFlagsetRole(teams *service.TeamService, flagsets *service.FlagsetService, paramName string, min model.Role) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims := MustClaims(c)
			if claims == nil {
				return c.JSON(http.StatusUnauthorized, model.APIResponse{Success: false, Message: "unauthenticated"})
			}
			if claims.IsSuperAdmin {
				return next(c)
			}
			id := c.Param(paramName)
			fs, err := flagsets.Get(c.Request().Context(), id)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, model.APIResponse{Success: false, Message: "failed to load flagset"})
			}
			if fs == nil {
				return c.JSON(http.StatusNotFound, model.APIResponse{Success: false, Message: "flagset not found"})
			}
			role, err := teams.Role(c.Request().Context(), fs.TeamID, claims.UserID)
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

// RequireFlagRole resolves flag → flagset → team.
func RequireFlagRole(teams *service.TeamService, flagsets *service.FlagsetService, flags *service.FlagService, paramName string, min model.Role) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims := MustClaims(c)
			if claims == nil {
				return c.JSON(http.StatusUnauthorized, model.APIResponse{Success: false, Message: "unauthenticated"})
			}
			if claims.IsSuperAdmin {
				return next(c)
			}
			flagID := c.Param(paramName)
			fsID, err := flags.FlagsetID(c.Request().Context(), flagID)
			if err != nil {
				return c.JSON(http.StatusNotFound, model.APIResponse{Success: false, Message: "flag not found"})
			}
			fs, err := flagsets.Get(c.Request().Context(), fsID)
			if err != nil || fs == nil {
				return c.JSON(http.StatusNotFound, model.APIResponse{Success: false, Message: "flagset not found"})
			}
			role, err := teams.Role(c.Request().Context(), fs.TeamID, claims.UserID)
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

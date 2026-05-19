package handler

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/thomaspoignant/go-feature-flag/cmd/management/api/middleware"
	"github.com/thomaspoignant/go-feature-flag/cmd/management/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/management/model"
	"github.com/thomaspoignant/go-feature-flag/cmd/management/repository"
	"github.com/thomaspoignant/go-feature-flag/cmd/management/service"
)

const stateCookie = "goff_mgmt_oidc_state"

type AuthHandler struct {
	cfg   config.Config
	auth  *service.AuthService
	users *repository.UserRepo
	teams *service.TeamService
}

func NewAuthHandler(cfg config.Config, auth *service.AuthService, users *repository.UserRepo, teams *service.TeamService) *AuthHandler {
	return &AuthHandler{cfg: cfg, auth: auth, users: users, teams: teams}
}

// Login godoc
// @Summary  Start OIDC login
// @Tags     auth
// @Success  302
// @Router   /auth/login [get]
func (h *AuthHandler) Login(c echo.Context) error {
	state, err := h.auth.RandomState()
	if err != nil {
		return internal(c, "failed to generate state")
	}
	c.SetCookie(&http.Cookie{
		Name:     stateCookie,
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cfg.Auth.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(10 * time.Minute),
	})
	return c.Redirect(http.StatusFound, h.auth.AuthURL(state))
}

// Callback godoc
// @Summary  OIDC callback
// @Tags     auth
// @Param    code  query string true  "OIDC authorization code"
// @Param    state query string true  "OIDC state"
// @Success  302
// @Router   /auth/callback [get]
func (h *AuthHandler) Callback(c echo.Context) error {
	state := c.QueryParam("state")
	code := c.QueryParam("code")
	cookie, err := c.Cookie(stateCookie)
	if err != nil || cookie.Value == "" || cookie.Value != state {
		return badRequest(c, "invalid state")
	}
	c.SetCookie(&http.Cookie{Name: stateCookie, Value: "", Path: "/", MaxAge: -1})
	_, jwtStr, err := h.auth.Exchange(c.Request().Context(), code)
	if err != nil {
		return badRequest(c, "exchange failed: "+err.Error())
	}
	c.SetCookie(&http.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    jwtStr,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cfg.Auth.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		Domain:   h.cfg.Auth.CookieDomain,
		Expires:  time.Now().Add(h.cfg.Auth.SessionMaxAge),
	})
	dest := h.cfg.Auth.PostLoginRedirect
	if dest == "" {
		dest = "/"
	}
	return c.Redirect(http.StatusFound, dest)
}

// Logout godoc
// @Summary  Logout
// @Tags     auth
// @Success  200 {object} model.APIResponse
// @Router   /auth/logout [post]
func (h *AuthHandler) Logout(c echo.Context) error {
	c.SetCookie(&http.Cookie{
		Name: middleware.SessionCookieName, Value: "", Path: "/", MaxAge: -1,
		Secure: h.cfg.Auth.CookieSecure, HttpOnly: true,
	})
	return ok(c, nil)
}

// Me godoc
// @Summary  Get current user + memberships
// @Tags     auth
// @Success  200 {object} model.APIResponse{data=model.MeResponse}
// @Router   /auth/me [get]
func (h *AuthHandler) Me(c echo.Context) error {
	claims := middleware.MustClaims(c)
	u, err := h.users.GetByID(c.Request().Context(), claims.UserID)
	if err != nil {
		return internal(c, "failed to load user")
	}
	if u == nil {
		return notFound(c, "user not found")
	}
	mem, err := h.teams.Memberships(c.Request().Context(), claims.UserID)
	if err != nil {
		return internal(c, "failed to load membership")
	}
	return ok(c, model.MeResponse{User: *u, Membership: mem})
}

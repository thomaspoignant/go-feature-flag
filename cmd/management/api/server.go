package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
	emw "github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	"go.uber.org/zap"

	"github.com/thomaspoignant/go-feature-flag/cmd/management/api/middleware"
	"github.com/thomaspoignant/go-feature-flag/cmd/management/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/management/handler"
	"github.com/thomaspoignant/go-feature-flag/cmd/management/model"
	"github.com/thomaspoignant/go-feature-flag/cmd/management/service"
)

type Handlers struct {
	Auth       *handler.AuthHandler
	Teams      *handler.TeamHandler
	Flagsets   *handler.FlagsetHandler
	Flags      *handler.FlagHandler
	Versions   *handler.VersionHandler
	Audit      *handler.AuditHandler
	Onboarding *handler.OnboardingHandler
}

type Services struct {
	Auth     *service.AuthService
	Teams    *service.TeamService
	Flagsets *service.FlagsetService
	Flags    *service.FlagService
}

func New(cfg config.Config, log *zap.Logger, h Handlers, s Services) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Use(emw.Recover())
	// AllowCredentials requires a specific origin (the CORS spec forbids wildcard
	// with credentials), so echo the request Origin back. Suitable for local dev;
	// tighten to an explicit allow-list in production.
	e.Use(emw.CORSWithConfig(emw.CORSConfig{
		AllowOriginFunc:  func(origin string) (bool, error) { return true, nil },
		AllowCredentials: true,
		AllowMethods: []string{
			http.MethodGet, http.MethodHead, http.MethodPost,
			http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions,
		},
		AllowHeaders: []string{"Content-Type", "Accept", "Authorization"},
	}))
	e.Use(middleware.ZapLogger(log))

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, model.APIResponse{Success: true, Data: map[string]string{"status": "ok"}})
	})

	if cfg.Server.EnableSwagger {
		e.GET("/swagger/*", echoSwagger.WrapHandler)
	}

	e.GET("/auth/login", h.Auth.Login)
	e.GET("/auth/callback", h.Auth.Callback)
	e.POST("/auth/logout", h.Auth.Logout)

	auth := middleware.RequireAuth(s.Auth)

	api := e.Group("/api/v1", auth)
	api.GET("/auth/me", h.Auth.Me)

	registerOnboardingRoutes(api, h)
	registerTeamRoutes(api, h, s)
	registerFlagsetRoutes(api, h, s)
	registerFlagRoutes(api, h, s)
	registerVersionRoutes(api, h, s)
	api.GET("/audit", h.Audit.List)

	return e
}

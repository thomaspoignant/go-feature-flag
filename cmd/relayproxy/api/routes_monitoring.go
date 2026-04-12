package api

import (
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo-contrib/pprof"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	custommiddleware "github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/api/middleware"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/controller"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	helpermiddleware "github.com/thomaspoignant/go-feature-flag/cmdhelpers/api/middleware"
)

func (s *Server) addMonitoringRoutes() {
	if s.config.EffectiveMonitoringPort(s.zapLog) != 0 {
		s.monitoringEcho = echo.New()
		s.monitoringEcho.HideBanner = true
		s.monitoringEcho.HidePort = true
		s.monitoringEcho.Debug = s.config.IsDebugEnabled()
		s.monitoringEcho.Use(helpermiddleware.ZapLogger(s.zapLog, s.config.IsDebugEnabled()))
		s.monitoringEcho.Use(middleware.CORSWithConfig(middleware.DefaultCORSConfig))
		s.apiEcho.Use(custommiddleware.VersionHeader(custommiddleware.VersionHeaderConfig{
			RelayProxyConfig: s.config,
		}))
		s.monitoringEcho.Use(middleware.Recover())
		s.initMonitoringEndpoint(s.monitoringEcho)
	} else {
		s.initMonitoringEndpoint(s.apiEcho)
	}
}

// initMonitoringEndpoint initialize the monitoring endpoints and associate them to the correct echo instance.
func (s *Server) initMonitoringEndpoint(echoInstance *echo.Echo) {
	if s.services.Metrics != (metric.Metrics{}) {
		echoInstance.GET("/metrics", echoprometheus.NewHandlerWithConfig(
			echoprometheus.HandlerConfig{Gatherer: s.services.Metrics.Registry}))
	}

	// Init controllers
	cHealth := controller.NewHealth(s.services.MonitoringService)
	cInfo := controller.NewInfo(s.services.MonitoringService)

	// health Routes
	echoInstance.GET("/health", cHealth.Handler)
	echoInstance.GET("/info", cInfo.Handler)

	if s.config.IsDebugEnabled() || s.config.EnablePprof {
		pprof.Register(echoInstance)
	}
}

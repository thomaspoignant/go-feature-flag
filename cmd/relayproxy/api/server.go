package api

import (
	"errors"
	"fmt"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	custommiddleware "github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/api/middleware"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/api/opentelemetry"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/controller"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/ofrep"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"
)

// New is used to create a new instance of the API server
func New(config *config.Config,
	services service.Services,
	zapLog *zap.Logger,
) Server {
	s := Server{
		config:      config,
		services:    services,
		zapLog:      zapLog,
		otelService: opentelemetry.NewOtelService(),
	}
	s.apiEcho = echo.New()
	s.initRoutes()
	return s
}

// Server is the struct that represents the API server
type Server struct {
	config         *config.Config
	apiEcho        *echo.Echo
	monitoringEcho *echo.Echo
	services       service.Services
	zapLog         *zap.Logger
	otelService    opentelemetry.OtelService
}

// initRoutes initialize the API endpoints that contain business logic and specificity for the relay proxy
func (s *Server) initRoutes() {
	s.apiEcho.HideBanner = true
	s.apiEcho.HidePort = true
	s.apiEcho.Debug = s.config.Debug
	if s.services.Metrics != (metric.Metrics{}) {
		s.apiEcho.Use(echoprometheus.NewMiddlewareWithConfig(echoprometheus.MiddlewareConfig{
			Subsystem:  metric.GOFFSubSystem,
			Registerer: s.services.Metrics.Registry,
		}))
	}
	s.apiEcho.Use(otelecho.Middleware("go-feature-flag"))
	s.apiEcho.Use(custommiddleware.ZapLogger(s.zapLog, s.config))
	s.apiEcho.Use(middleware.CORSWithConfig(middleware.DefaultCORSConfig))
	s.apiEcho.Use(middleware.Recover())
	s.apiEcho.Use(middleware.TimeoutWithConfig(
		middleware.TimeoutConfig{
			Skipper: func(c echo.Context) bool {
				// ignore websocket in the timeout
				return strings.HasPrefix(c.Request().URL.String(), "/ws")
			},
			Timeout: time.Duration(s.config.RestAPITimeout) * time.Millisecond,
		}),
	)

	// Init controllers
	cAllFlags := controller.NewAllFlags(s.services.GOFeatureFlagService, s.services.Metrics)
	cFlagEval := controller.NewFlagEval(s.services.GOFeatureFlagService, s.services.Metrics)
	cFlagEvalOFREP := ofrep.NewOFREPEvaluate(s.services.GOFeatureFlagService, s.services.Metrics)
	cEvalDataCollector := controller.NewCollectEvalData(s.services.GOFeatureFlagService, s.services.Metrics)
	cRetrieverRefresh := controller.NewForceFlagsRefresh(s.services.GOFeatureFlagService, s.services.Metrics)

	// Init routes
	s.addGOFFRoutes(cAllFlags, cFlagEval, cEvalDataCollector)
	s.addOFREPRoutes(cFlagEvalOFREP)
	s.addWebsocketRoutes()
	s.addMonitoringRoutes()
	s.addAdminRoutes(cRetrieverRefresh)
}

// Start launch the API server
func (s *Server) Start() {
	// starting the monitoring server on a different port if configured
	if s.monitoringEcho != nil {
		go func() {
			addressMonitoring := fmt.Sprintf("0.0.0.0:%d", s.config.MonitoringPort)
			s.zapLog.Info(
				"Starting monitoring",
				zap.String("address", addressMonitoring))
			err := s.monitoringEcho.Start(addressMonitoring)
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				s.zapLog.Fatal("Error starting monitoring", zap.Error(err))
			}
		}()
		defer func() { _ = s.monitoringEcho.Close() }()
	}

	// starting the main application
	if s.config.ListenPort == 0 {
		s.config.ListenPort = 1031
	}
	address := fmt.Sprintf("0.0.0.0:%d", s.config.ListenPort)
	s.zapLog.Info(
		"Starting go-feature-flag relay proxy ...",
		zap.String("address", address),
		zap.String("version", s.config.Version))

	err := s.apiEcho.Start(address)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.zapLog.Fatal("Error starting relay proxy", zap.Error(err))
	}
}

// StartAwsLambda is starting the relay proxy as an AWS Lambda
func (s *Server) StartAwsLambda() {
	adapter := newAwsLambdaHandler(s.apiEcho)
	adapter.Start()
}

// Stop shutdown the API server
func (s *Server) Stop() {
	err := s.otelService.Stop()
	if err != nil {
		s.zapLog.Error("impossible to stop otel", zap.Error(err))
	}

	if s.monitoringEcho != nil {
		err = s.monitoringEcho.Close()
		if err != nil {
			s.zapLog.Fatal("impossible to stop monitoring", zap.Error(err))
		}
	}

	err = s.apiEcho.Close()
	if err != nil {
		s.zapLog.Fatal("impossible to stop go-feature-flag relay proxy", zap.Error(err))
	}
}

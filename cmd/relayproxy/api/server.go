package api

import (
	"errors"
	"fmt"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	custommiddleware "github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/api/middleware"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/api/opentelemetry"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/controller"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
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
	s.init()
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

// init initialize the configuration of our API server (using echo)
func (s *Server) init() {
	s.apiEcho = echo.New()
	s.initAPIEndpoint(s.apiEcho)
	if s.config.MonitoringPort != 0 {
		s.monitoringEcho = echo.New()
		s.monitoringEcho.HideBanner = true
		s.monitoringEcho.HidePort = true
		s.monitoringEcho.Debug = s.config.Debug
		s.monitoringEcho.Use(custommiddleware.ZapLogger(s.zapLog, s.config))
		s.monitoringEcho.Use(middleware.CORSWithConfig(middleware.DefaultCORSConfig))
		s.monitoringEcho.Use(middleware.Recover())
		s.initMonitoringEndpoint(s.monitoringEcho)
	} else {
		s.initMonitoringEndpoint(s.apiEcho)
	}
}

// initAPIEndpoint initialize the API endpoints that contain business logic and specificity for the relay proxy
func (s *Server) initAPIEndpoint(echoInstance *echo.Echo) {
	echoInstance.HideBanner = true
	echoInstance.HidePort = true
	echoInstance.Debug = s.config.Debug
	if s.services.Metrics != (metric.Metrics{}) {
		s.apiEcho.Use(echoprometheus.NewMiddlewareWithConfig(echoprometheus.MiddlewareConfig{
			Subsystem:  metric.GOFFSubSystem,
			Registerer: s.services.Metrics.Registry,
		}))
	}
	echoInstance.Use(otelecho.Middleware("go-feature-flag"))
	echoInstance.Use(custommiddleware.ZapLogger(s.zapLog, s.config))
	echoInstance.Use(middleware.CORSWithConfig(middleware.DefaultCORSConfig))
	echoInstance.Use(middleware.Recover())
	echoInstance.Use(middleware.TimeoutWithConfig(
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
	cEvalDataCollector := controller.NewCollectEvalData(s.services.GOFeatureFlagService, s.services.Metrics)

	// Init routes
	v1 := echoInstance.Group("/v1")
	if len(s.config.APIKeys) > 0 {
		v1.Use(middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
			Validator: func(key string, c echo.Context) (bool, error) {
				return s.config.APIKeyExists(key), nil
			},
		}))
	}
	v1.POST("/allflags", cAllFlags.Handler)
	v1.POST("/feature/:flagKey/eval", cFlagEval.Handler)
	v1.POST("/data/collector", cEvalDataCollector.Handler)

	// Swagger - only available if option is enabled
	if s.config.EnableSwagger {
		echoInstance.GET("/swagger/*", echoSwagger.WrapHandler)
	}

	// initWebsocketsEndpoints initialize the websocket endpoints
	cFlagReload := controller.NewWsFlagChange(s.services.WebsocketService, s.zapLog)
	wsV1 := echoInstance.Group("/ws/v1")
	wsV1.Use(custommiddleware.WebsocketAuthorizer(s.config))
	wsV1.GET("/flag/change", cFlagReload.Handler)
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

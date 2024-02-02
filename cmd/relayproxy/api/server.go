package api

import (
	"context"
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
	config                 *config.Config
	proxyEchoInstance      *echo.Echo
	monitoringEchoInstance *echo.Echo
	services               service.Services
	zapLog                 *zap.Logger
	otelService            opentelemetry.OtelService
}

func (s *Server) addMetricRoutes(instance *echo.Echo) {
	if s.services.Metrics != (metric.Metrics{}) {
		instance.GET("/metrics", echoprometheus.NewHandlerWithConfig(
			echoprometheus.HandlerConfig{Gatherer: s.services.Metrics.Registry}))
	}

	cHealth := controller.NewHealth(s.services.MonitoringService)
	cInfo := controller.NewInfo(s.services.MonitoringService)

	// health Routes
	instance.GET("/health", cHealth.Handler)
	instance.GET("/info", cInfo.Handler)
}

// init initialize the configuration of our API server (using echo)
func (s *Server) init() {
	s.proxyEchoInstance = s.newDefaultEchoInstance()
	if s.config.OpenTelemetryOtlpEndpoint != "" {
		err := s.otelService.Init(context.Background(), *s.config)
		if err != nil {
			s.zapLog.Error("error while initializing Otel", zap.Error(err))
			// we can continue because otel is not mandatory to start the server
		}
	}

	// Global Middlewares
	if s.services.Metrics != (metric.Metrics{}) {
		s.proxyEchoInstance.Use(echoprometheus.NewMiddlewareWithConfig(echoprometheus.MiddlewareConfig{
			Subsystem:  metric.GOFFSubSystem,
			Registerer: s.services.Metrics.Registry,
		}))
	}
	s.proxyEchoInstance.Use(otelecho.Middleware("go-feature-flag"))

	if s.config.MonitoringPort != 0 {
		s.monitoringEchoInstance = s.newDefaultEchoInstance()
		s.addMetricRoutes(s.monitoringEchoInstance)
	} else {
		s.addMetricRoutes(s.proxyEchoInstance)
	}

	// Swagger - only available if option is enabled
	if s.config.EnableSwagger {
		s.proxyEchoInstance.GET("/swagger/*", echoSwagger.WrapHandler)
	}

	// endpoints configuration
	s.initAPIEndpoints()
	s.initWebsocketsEndpoints()
}

// initAPIEndpoints initialize the API endpoints
func (s *Server) initAPIEndpoints() {
	// Init controllers
	cAllFlags := controller.NewAllFlags(s.services.GOFeatureFlagService, s.services.Metrics)
	cFlagEval := controller.NewFlagEval(s.services.GOFeatureFlagService, s.services.Metrics)
	cEvalDataCollector := controller.NewCollectEvalData(s.services.GOFeatureFlagService, s.services.Metrics)

	// Init routes
	v1 := s.proxyEchoInstance.Group("/v1")
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
}

// initWebsocketsEndpoints initialize the websocket endpoints
func (s *Server) initWebsocketsEndpoints() {
	cFlagReload := controller.NewWsFlagChange(s.services.WebsocketService, s.zapLog)
	v1 := s.proxyEchoInstance.Group("/ws/v1")
	v1.Use(custommiddleware.WebsocketAuthorizer(s.config))
	v1.GET("/flag/change", cFlagReload.Handler)
}

// Start launch the API server
func (s *Server) Start() {
	if s.config.ListenPort == 0 {
		s.config.ListenPort = 1031
	}
	address := fmt.Sprintf("0.0.0.0:%d", s.config.ListenPort)
	s.zapLog.Info(
		"Starting go-feature-flag relay proxy ...",
		zap.String("address", address),
		zap.String("version", s.config.Version))

	if s.monitoringEchoInstance != nil {
		go func() {
			addressMonitoring := fmt.Sprintf("0.0.0.0:%d", s.config.MonitoringPort)
			s.zapLog.Info(
				"Starting monitoring",
				zap.String("address", addressMonitoring))
			err := s.monitoringEchoInstance.Start(addressMonitoring)
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				s.zapLog.Fatal("Error starting monitoring", zap.Error(err))
			}
		}()
		defer func() { _ = s.monitoringEchoInstance.Close() }()
	}

	err := s.proxyEchoInstance.Start(address)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.zapLog.Fatal("Error starting relay proxy", zap.Error(err))
	}
}

// StartAwsLambda is starting the relay proxy as an AWS Lambda
func (s *Server) StartAwsLambda() {
	adapter := newAwsLambdaHandler(s.proxyEchoInstance)
	adapter.Start()
}

// Stop shutdown the API server
func (s *Server) Stop() {
	err := s.otelService.Stop()
	if err != nil {
		s.zapLog.Error("impossible to stop otel", zap.Error(err))
	}

	err = s.proxyEchoInstance.Shutdown(context.Background())
	if err != nil {
		s.zapLog.Fatal("impossible to stop go-feature-flag relay proxy", zap.Error(err))
	}
}

func (s *Server) newDefaultEchoInstance() *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Debug = s.config.Debug
	e.Use(custommiddleware.ZapLogger(s.zapLog, s.config))
	e.Use(middleware.CORSWithConfig(middleware.DefaultCORSConfig))
	e.Use(middleware.TimeoutWithConfig(
		middleware.TimeoutConfig{
			Skipper: func(c echo.Context) bool {
				// ignore websocket in the timeout
				return strings.HasPrefix(c.Request().URL.String(), "/ws")
			},
			Timeout: time.Duration(s.config.RestAPITimeout) * time.Millisecond,
		}),
	)
	e.Use(middleware.Recover())
	return e
}

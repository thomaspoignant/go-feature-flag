package api

import (
	"fmt"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	custommiddleware "github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/api/middleware"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/controller"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"go.uber.org/zap"
	"strings"
	"time"
)

// New is used to create a new instance of the API server
func New(config *config.Config,
	services service.Services,
	zapLog *zap.Logger,
) Server {
	s := Server{
		config:   config,
		services: services,
		zapLog:   zapLog,
	}
	s.init()
	return s
}

// Server is the struct that represent the API server
type Server struct {
	config       *config.Config
	echoInstance *echo.Echo
	services     service.Services
	zapLog       *zap.Logger
}

// init initialize the configuration of our API server (using echo)
func (s *Server) init() {
	s.echoInstance = echo.New()
	s.echoInstance.HideBanner = true
	s.echoInstance.HidePort = true
	s.echoInstance.Debug = s.config.Debug

	// Global Middlewares
	metrics := metric.NewMetrics()
	prom := prometheus.NewPrometheus("gofeatureflag", nil, metrics.MetricList())
	prom.Use(s.echoInstance)
	s.echoInstance.Use(metrics.AddCustomMetricsMiddleware)
	s.echoInstance.Use(custommiddleware.ZapLogger(s.zapLog, s.config))
	s.echoInstance.Use(middleware.Recover())
	s.echoInstance.Use(middleware.TimeoutWithConfig(
		middleware.TimeoutConfig{
			Skipper: func(c echo.Context) bool {
				// ignore websocket in the timeout
				return strings.HasPrefix(c.Request().URL.String(), "/ws")
			},
			Timeout: time.Duration(s.config.RestAPITimeout) * time.Millisecond,
		}),
	)

	// endpoints configuration
	s.initAPIEndpoints()
	s.initPublicEndpoints()
	s.initWebsocketsEndpoints()
}

// initAPIEndpoints initialize the API endpoints
func (s *Server) initAPIEndpoints() {
	// Init controllers
	cAllFlags := controller.NewAllFlags(s.services.GOFeatureFlagService)
	cFlagEval := controller.NewFlagEval(s.services.GOFeatureFlagService)
	cEvalDataCollector := controller.NewCollectEvalData(s.services.GOFeatureFlagService)

	// Init routes
	v1 := s.echoInstance.Group("/v1")
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

// initPublicEndpoints initialize the public endpoints to monitor the application
func (s *Server) initPublicEndpoints() {
	// Init controllers
	cHealth := controller.NewHealth(s.services.MonitoringService)
	cInfo := controller.NewInfo(s.services.MonitoringService)

	// health Routes
	s.echoInstance.GET("/health", cHealth.Handler)
	s.echoInstance.GET("/info", cInfo.Handler)

	// Swagger - only available if option is enabled
	if s.config.EnableSwagger {
		s.echoInstance.GET("/swagger/*", echoSwagger.WrapHandler)
	}
}

// initWebsocketsEndpoints initialize the websocket endpoints
func (s *Server) initWebsocketsEndpoints() {
	cFlagReload := controller.NewWsFlagChange(s.services.WebsocketService, s.zapLog)
	v1 := s.echoInstance.Group("/ws/v1")
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

	err := s.echoInstance.Start(address)
	if err != nil {
		s.zapLog.Fatal("impossible to start the proxy", zap.Error(err))
	}
}

// StartAwsLambda is starting the relay proxy as an AWS Lambda
func (s *Server) StartAwsLambda() {
	adapter := newAwsLambdaHandler(s.echoInstance)
	adapter.Start()
}

// Stop shutdown the API server
func (s *Server) Stop() {
	err := s.echoInstance.Close()
	if err != nil {
		s.zapLog.Fatal("impossible to stop go-feature-flag relay proxy", zap.Error(err))
	}
}

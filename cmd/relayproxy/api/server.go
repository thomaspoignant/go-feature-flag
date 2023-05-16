package api

import (
	"fmt"
	custommiddleware "github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/api/middleware"
	"strings"
	"time"

	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/controller"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"go.uber.org/zap"
)

// New is used to create a new instance of the API server
func New(config *config.Config,
	monitoringService service.Monitoring,
	goFF *ffclient.GoFeatureFlag,
	zapLog *zap.Logger,
) Server {
	s := Server{
		config:            config,
		monitoringService: monitoringService,
		goFF:              goFF,
		zapLog:            zapLog,
	}
	s.init()
	return s
}

// Server is the struct that represent the API server
type Server struct {
	config            *config.Config
	echoInstance      *echo.Echo
	monitoringService service.Monitoring
	goFF              *ffclient.GoFeatureFlag
	zapLog            *zap.Logger
}

// init initialize the configuration of our API server (using echo)
func (s *Server) init() {
	s.echoInstance = echo.New()
	s.echoInstance.HideBanner = true
	s.echoInstance.HidePort = true
	s.echoInstance.Debug = s.config.Debug

	// Prometheus
	metrics := metric.NewMetrics()
	prometheus := prometheus.NewPrometheus("gofeatureflag", nil, metrics.MetricList())
	prometheus.Use(s.echoInstance)
	s.echoInstance.Use(metrics.AddCustomMetricsMiddleware)

	// Middlewares
	s.echoInstance.Use(custommiddleware.ZapLogger(s.zapLog, s.config))
	s.echoInstance.Use(middleware.Recover())
	s.echoInstance.Use(middleware.TimeoutWithConfig(
		middleware.TimeoutConfig{Timeout: time.Duration(s.config.RestAPITimeout) * time.Millisecond}),
	)
	if len(s.config.APIKeys) > 0 {
		s.echoInstance.Use(middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
			Skipper: func(c echo.Context) bool {
				_, ok := map[string]struct{}{
					"/health":  {},
					"/info":    {},
					"/metrics": {},
				}[c.Path()]
				return ok || strings.HasPrefix(c.Path(), "/swagger/")
			},
			Validator: func(key string, c echo.Context) (bool, error) {
				return s.config.APIKeyExists(key), nil
			},
		}))
	}

	// Init controllers
	cHealth := controller.NewHealth(s.monitoringService)
	cInfo := controller.NewInfo(s.monitoringService)
	cAllFlags := controller.NewAllFlags(s.goFF)
	cFlagEval := controller.NewFlagEval(s.goFF)
	cEvalDataCollector := controller.NewCollectEvalData(s.goFF)

	// health Routes
	s.echoInstance.GET("/health", cHealth.Handler)
	s.echoInstance.GET("/info", cInfo.Handler)

	// Swagger - only available if option is enabled
	if s.config.EnableSwagger {
		s.echoInstance.GET("/swagger/*", echoSwagger.WrapHandler)
	}

	// GO feature flags routes
	v1 := s.echoInstance.Group("/v1")
	v1.POST("/allflags", cAllFlags.Handler)
	v1.POST("/feature/:flagKey/eval", cFlagEval.Handler)
	v1.POST("/data/collector", cEvalDataCollector.Handler)
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

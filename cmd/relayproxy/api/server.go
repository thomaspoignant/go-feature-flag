package api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
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
	unixListener   net.Listener
}

// initRoutes initialize the API endpoints that contain business logic and specificity for the relay proxy
func (s *Server) initRoutes() {
	s.apiEcho.HideBanner = true
	s.apiEcho.HidePort = true
	s.apiEcho.Debug = s.config.IsDebugEnabled()
	s.apiEcho.Use(otelecho.Middleware("go-feature-flag"))
	s.apiEcho.Use(custommiddleware.ZapLogger(s.zapLog, s.config))
	s.apiEcho.Use(middleware.BodyDumpWithConfig(middleware.BodyDumpConfig{
		Skipper: func(c echo.Context) bool {
			isSwagger := strings.HasPrefix(c.Request().URL.String(), "/swagger")
			return isSwagger || !s.zapLog.Core().Enabled(zap.DebugLevel)
		},
		Handler: func(_ echo.Context, reqBody []byte, _ []byte) {
			s.zapLog.Debug("Request info", zap.ByteString("request_body", reqBody))
		},
	}))
	if s.services.Metrics != (metric.Metrics{}) {
		s.apiEcho.Use(echoprometheus.NewMiddlewareWithConfig(echoprometheus.MiddlewareConfig{
			Subsystem:  metric.GOFFSubSystem,
			Registerer: s.services.Metrics.Registry,
		}))
	}
	s.apiEcho.Use(middleware.CORSWithConfig(middleware.DefaultCORSConfig))

	s.apiEcho.Use(custommiddleware.VersionHeader(custommiddleware.VersionHeaderConfig{
		Skipper: func(_ echo.Context) bool {
			return s.config.DisableVersionHeader
		},
		RelayProxyConfig: s.config,
	}))

	s.apiEcho.Use(middleware.Recover())

	// Init controllers
	cAllFlags := controller.NewAllFlags(s.services.FlagsetManager, s.services.Metrics)
	cFlagEval := controller.NewFlagEval(s.services.FlagsetManager, s.services.Metrics)
	cFlagEvalOFREP := ofrep.NewOFREPEvaluate(s.services.FlagsetManager, s.services.Metrics)
	cEvalDataCollector := controller.NewCollectEvalData(
		s.services.FlagsetManager,
		s.services.Metrics,
		s.zapLog,
	)
	cRetrieverRefresh := controller.NewForceFlagsRefresh(
		s.services.FlagsetManager,
		s.services.Metrics,
	)
	cFlagChangeAPI := controller.NewAPIFlagChange(
		s.services.FlagsetManager,
		s.services.Metrics,
	)
	cFlagConfiguration := controller.NewAPIFlagConfiguration(
		s.services.FlagsetManager,
		s.services.Metrics,
	)

	// Init routes
	s.addGOFFRoutes(cAllFlags, cFlagEval, cEvalDataCollector, cFlagChangeAPI, cFlagConfiguration)
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

	// start the OpenTelemetry tracing service
	err := s.otelService.Init(context.Background(), s.zapLog, s.config)
	if err != nil {
		s.zapLog.Error(
			"error while initializing OTel, continuing without tracing enabled",
			zap.Error(err),
		)
		// we can continue because otel is not mandatory to start the server
	}

	// starting the main application
	// Start Unix socket listener if configured
	if s.config.UnixSocket != "" {
		go s.startUnixSocket()
		defer s.cleanupUnixSocket()
	}

	// Start TCP listener if port is configured
	if s.config.ListenPort != 0 {
		address := fmt.Sprintf("0.0.0.0:%d", s.config.ListenPort)
		s.zapLog.Info(
			"Starting go-feature-flag relay proxy ...",
			zap.String("address", address),
			zap.String("version", s.config.Version))

		err = s.apiEcho.Start(address)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.zapLog.Fatal("Error starting relay proxy", zap.Error(err))
		}
	} else if s.config.UnixSocket != "" {
		// If only Unix socket is configured, wait for shutdown
		select {}
	}
}

// StartAwsLambda is starting the relay proxy as an AWS Lambda
func (s *Server) StartAwsLambda() {
	lambda.Start(s.getLambdaHandler())
}

func (s *Server) getLambdaHandler() interface{} {
	handlerMngr := newAwsLambdaHandlerManager(s.apiEcho, s.config.AwsApiGatewayBasePath)
	return handlerMngr.GetAdapter(s.config.AwsLambdaAdapter)
}

// Stop shutdown the API server
func (s *Server) Stop(ctx context.Context) {
	err := s.otelService.Stop(ctx)
	if err != nil {
		s.zapLog.Error("impossible to stop otel", zap.Error(err))
	}

	if s.monitoringEcho != nil {
		err = s.monitoringEcho.Close()
		if err != nil {
			s.zapLog.Fatal("impossible to stop monitoring", zap.Error(err))
		}
	}

	// Close Unix socket listener if it exists
	s.cleanupUnixSocket()

	err = s.apiEcho.Close()
	if err != nil {
		s.zapLog.Fatal("impossible to stop go-feature-flag relay proxy", zap.Error(err))
	}
}

// startUnixSocket starts the API server on a Unix socket
func (s *Server) startUnixSocket() {
	// Remove existing socket file if it exists
	_ = os.Remove(s.config.UnixSocket)

	listener, err := net.Listen("unix", s.config.UnixSocket)
	if err != nil {
		s.zapLog.Fatal("Error creating Unix socket listener", zap.Error(err))
		return
	}

	s.unixListener = listener
	s.zapLog.Info(
		"Starting go-feature-flag relay proxy on Unix socket ...",
		zap.String("socket", s.config.UnixSocket),
		zap.String("version", s.config.Version))

	err = s.apiEcho.Server.Serve(listener)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.zapLog.Fatal("Error starting relay proxy on Unix socket", zap.Error(err))
	}
}

// cleanupUnixSocket closes the Unix socket listener and removes the socket file
func (s *Server) cleanupUnixSocket() {
	if s.unixListener != nil {
		err := s.unixListener.Close()
		if err != nil {
			s.zapLog.Error("Error closing Unix socket listener", zap.Error(err))
		}
		s.unixListener = nil
	}

	if s.config.UnixSocket != "" {
		err := os.Remove(s.config.UnixSocket)
		if err != nil && !os.IsNotExist(err) {
			s.zapLog.Error("Error removing Unix socket file", zap.Error(err))
		}
	}
}

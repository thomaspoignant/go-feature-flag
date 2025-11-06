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

func (s *Server) StartWithContext(ctx context.Context) {
	// start the OpenTelemetry tracing service
	err := s.otelService.Init(ctx, s.zapLog, s.config)
	if err != nil {
		s.zapLog.Error(
			"error while initializing OTel, continuing without tracing enabled",
			zap.Error(err),
		)
		// we can continue because otel is not mandatory to start the server
	}

	switch s.config.GetServerMode(s.zapLog) {
	case config.ServerModeLambda:
		s.startAwsLambda()
	case config.ServerModeUnixSocket:
		s.startUnixSocketServer(ctx)
	default:
		s.startAsHTTPServer()
	}
}

// startUnixSocketServer launch the API server as a unix socket.
func (s *Server) startUnixSocketServer(ctx context.Context) {
	socketPath := s.config.GetUnixSocketPath()

	// Clean up the old socket file if it exists (important for graceful restarts)
	if _, err := os.Stat(socketPath); err == nil {
		if err := os.Remove(socketPath); err != nil {
			s.zapLog.Fatal("Could not remove old socket file", zap.String("path", socketPath), zap.Error(err))
		}
	}

	lc := net.ListenConfig{}
	listener, err := lc.Listen(ctx, "unix", socketPath)
	if err != nil {
		s.zapLog.Fatal("Error creating Unix listener", zap.Error(err))
	}

	defer func() {
		if err := listener.Close(); err != nil {
			s.zapLog.Error("error closing unix socket listener", zap.Error(err))
		}
	}()
	s.apiEcho.Listener = listener

	s.zapLog.Info(
		"Starting go-feature-flag relay proxy as unix socket...",
		zap.String("socket", socketPath),
		zap.String("version", s.config.Version))

	err = s.apiEcho.StartServer(new(http.Server))
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.zapLog.Fatal("Error starting relay proxy as unix socket", zap.Error(err))
	}
}

// startAsHTTPServer launch the API server
func (s *Server) startAsHTTPServer() {
	// starting the monitoring server on a different port if configured
	if s.monitoringEcho != nil {
		go func() {
			addressMonitoring := fmt.Sprintf("%s:%d", s.config.GetServerHost(), s.config.GetMonitoringPort(s.zapLog))
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

	address := fmt.Sprintf("%s:%d", s.config.GetServerHost(), s.config.GetServerPort(s.zapLog))
	s.zapLog.Info(
		"Starting go-feature-flag relay proxy ...",
		zap.String("address", address),
		zap.String("version", s.config.Version))

	err := s.apiEcho.Start(address)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.zapLog.Fatal("Error starting relay proxy", zap.Error(err))
	}
}

// startAwsLambda is starting the relay proxy as an AWS Lambda
func (s *Server) startAwsLambda() {
	lambda.Start(s.getLambdaHandler())
}

// getLambdaHandler returns the appropriate lambda handler based on the configuration.
// We need a dedicated function because it is called from tests as well, this is the
// reason why we can't merged it in startAwsLambda.
func (s *Server) getLambdaHandler() interface{} {
	handlerMngr := newAwsLambdaHandlerManager(s.apiEcho, s.config.GetAwsApiGatewayBasePath(s.zapLog))
	return handlerMngr.GetAdapter(s.config.GetLambdaAdapter(s.zapLog))
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

	if s.apiEcho != nil {
		err = s.apiEcho.Close()
		if err != nil {
			s.zapLog.Fatal("impossible to stop go-feature-flag relay proxy", zap.Error(err))
		}
	}
}

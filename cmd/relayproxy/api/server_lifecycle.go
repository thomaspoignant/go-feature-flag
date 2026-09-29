package api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"go.uber.org/zap"
)

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

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	switch s.config.ServerMode(s.zapLog) {
	case config.ServerModeLambda:
		s.startAwsLambda()
	case config.ServerModeUnixSocket:
		s.startUnixSocketServer(ctx)
	default:
		s.startAsHTTPServer(ctx)
	}
}

// startUnixSocketServer launch the API server as a unix socket.
func (s *Server) startUnixSocketServer(ctx context.Context) {
	socketPath := s.config.UnixSocketPath()

	// Clean up the old socket file if it exists (important for graceful restarts)
	if _, err := os.Stat(socketPath); err == nil {
		if err := os.Remove(socketPath); err != nil {
			s.zapLog.Fatal("Could not remove old socket file", zap.String("path", socketPath), zap.Error(err))
		}
	}

	// Start a http server for monitoring if monitoringport is configured
	if s.isMonitoringPortConfigured() {
		go s.startMonitoringServer()
		defer func() { _ = s.monitoringEcho.Shutdown(ctx) }()
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
func (s *Server) startAsHTTPServer(ctx context.Context) {
	if s.isMonitoringPortConfigured() {
		go s.startMonitoringServer()
		defer func() { _ = s.monitoringEcho.Shutdown(ctx) }()
	}

	address := fmt.Sprintf("%s:%d", s.config.ServerHost(), s.config.ServerPort(s.zapLog))
	s.zapLog.Info(
		"Starting go-feature-flag relay proxy ...",
		zap.String("address", address),
		zap.String("version", s.config.Version))

	shutdownDone := make(chan struct{})
	// nolint:gosec
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.apiEcho.Shutdown(shutdownCtx); err != nil {
			s.zapLog.Error("error shutting down api server", zap.Error(err))
		}
		close(shutdownDone)
	}()

	err := s.apiEcho.Start(address)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.zapLog.Fatal("Error starting relay proxy", zap.Error(err))
	}

	// Wait for Shutdown to finish draining connections before returning.
	<-shutdownDone
}

func (s *Server) startMonitoringServer() {
	addressMonitoring := fmt.Sprintf("%s:%d", s.config.ServerHost(), s.config.EffectiveMonitoringPort(s.zapLog))
	s.zapLog.Info(
		"Starting monitoring",
		zap.String("address", addressMonitoring))
	err := s.monitoringEcho.Start(addressMonitoring)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.zapLog.Fatal("Error starting monitoring", zap.Error(err))
	}
}

// startAwsLambda is starting the relay proxy as an AWS Lambda
func (s *Server) startAwsLambda() {
	lambda.Start(s.lambdaHandler())
}

// lambdaHandler returns the appropriate lambda handler based on the configuration.
// We need a dedicated function because it is called from tests as well, this is the
// reason why we can't merged it in startAwsLambda.
func (s *Server) lambdaHandler() any {
	handlerMngr := newAwsLambdaHandlerManager(s.apiEcho, s.config.EffectiveAwsApiGatewayBasePath(s.zapLog))
	return handlerMngr.SelectAdapter(s.config.LambdaAdapter(s.zapLog))
}

// Stop shutdown the API server
func (s *Server) Stop(ctx context.Context) {
	err := s.otelService.Stop(ctx)
	if err != nil {
		s.zapLog.Error("impossible to stop otel", zap.Error(err))
	}

	if s.monitoringEcho != nil {
		if err = s.monitoringEcho.Shutdown(ctx); err != nil {
			s.zapLog.Error("error stopping monitoring", zap.Error(err))
		}
	}

	if s.apiEcho != nil {
		if err = s.apiEcho.Shutdown(ctx); err != nil {
			s.zapLog.Error("error stopping relay proxy", zap.Error(err))
		}
	}
}

// isMonitoringPortConfigured checks if the monitoring port is configured.
func (s *Server) isMonitoringPortConfigured() bool {
	return s.monitoringEcho != nil && s.config.EffectiveMonitoringPort(s.zapLog) > 0
}

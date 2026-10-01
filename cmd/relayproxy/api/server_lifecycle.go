package api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/labstack/echo/v5"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"go.uber.org/zap"
)

// gracefulShutdownTimeout is the time given to in-flight requests to complete when the servers are stopped.
const gracefulShutdownTimeout = 5 * time.Second

func (s *Server) StartWithContext(ctx context.Context) {
	defer close(s.stopped)
	select {
	case <-s.stopRequested:
		// Stop has been called before the server started.
		return
	default:
	}

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
	// Stop closes s.stopRequested, which cancels ctx and gracefully shuts down the servers started below.
	go func() {
		select {
		case <-s.stopRequested:
			cancel()
		case <-ctx.Done():
		}
	}()

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
		var wg sync.WaitGroup
		wg.Go(func() { s.startMonitoringServer(ctx) })
		defer wg.Wait()
	}

	lc := net.ListenConfig{}
	listener, err := lc.Listen(ctx, "unix", socketPath)
	if err != nil {
		s.zapLog.Fatal("Error creating Unix listener", zap.Error(err))
	}

	defer func() {
		// the graceful shutdown already closes the listener
		if err := listener.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
			s.zapLog.Error("error closing unix socket listener", zap.Error(err))
		}
	}()

	s.zapLog.Info(
		"Starting go-feature-flag relay proxy as unix socket...",
		zap.String("socket", socketPath),
		zap.String("version", s.config.Version))

	err = echo.StartConfig{
		Listener:        listener,
		HideBanner:      true,
		HidePort:        true,
		GracefulTimeout: gracefulShutdownTimeout,
	}.Start(ctx, s.apiEcho)
	if err != nil {
		s.zapLog.Fatal("Error starting relay proxy as unix socket", zap.Error(err))
	}
}

// startAsHTTPServer launch the API server
func (s *Server) startAsHTTPServer(ctx context.Context) {
	if s.isMonitoringPortConfigured() {
		var wg sync.WaitGroup
		wg.Go(func() { s.startMonitoringServer(ctx) })
		defer wg.Wait()
	}

	address := fmt.Sprintf("%s:%d", s.config.ServerHost(), s.config.ServerPort(s.zapLog))
	s.zapLog.Info(
		"Starting go-feature-flag relay proxy ...",
		zap.String("address", address),
		zap.String("version", s.config.Version))

	// Start blocks until ctx is cancelled and the graceful shutdown has drained the connections.
	err := echo.StartConfig{
		Address:         address,
		HideBanner:      true,
		HidePort:        true,
		GracefulTimeout: gracefulShutdownTimeout,
		OnShutdownError: func(err error) {
			s.zapLog.Error("error shutting down api server", zap.Error(err))
		},
	}.Start(ctx, s.apiEcho)
	if err != nil {
		s.zapLog.Fatal("Error starting relay proxy", zap.Error(err))
	}
}

func (s *Server) startMonitoringServer(ctx context.Context) {
	addressMonitoring := fmt.Sprintf("%s:%d", s.config.ServerHost(), s.config.EffectiveMonitoringPort(s.zapLog))
	s.zapLog.Info(
		"Starting monitoring",
		zap.String("address", addressMonitoring))
	err := echo.StartConfig{
		Address:         addressMonitoring,
		HideBanner:      true,
		HidePort:        true,
		GracefulTimeout: gracefulShutdownTimeout,
		OnShutdownError: func(err error) {
			s.zapLog.Error("error stopping monitoring", zap.Error(err))
		},
	}.Start(ctx, s.monitoringEcho)
	if err != nil {
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

// Stop shutdown the API server.
// It waits for StartWithContext to return (servers drained) or for ctx to be done.
func (s *Server) Stop(ctx context.Context) {
	err := s.otelService.Stop(ctx)
	if err != nil {
		s.zapLog.Error("impossible to stop otel", zap.Error(err))
	}

	s.stop()
	select {
	case <-s.stopped:
	case <-ctx.Done():
		s.zapLog.Error("error stopping relay proxy", zap.Error(ctx.Err()))
	}
}

// isMonitoringPortConfigured checks if the monitoring port is configured.
func (s *Server) isMonitoringPortConfigured() bool {
	return s.monitoringEcho != nil && s.config.EffectiveMonitoringPort(s.zapLog) > 0
}

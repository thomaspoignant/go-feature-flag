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
	"github.com/labstack/echo/v5"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"go.uber.org/zap"
)

const (
	// gracefulTimeout is how long a listener waits for in-flight requests to finish
	// before it is forcibly closed.
	gracefulTimeout = 5 * time.Second

	// readHeaderTimeout bounds how long a client may take to send its request headers.
	// It replaces Echo v5's default ReadTimeout, which we clear below.
	readHeaderTimeout = 30 * time.Second
)

// tuneHTTPServer relaxes the timeouts Echo v5 applies by default.
//
// echo.StartConfig sets ReadTimeout to 30s to satisfy gosec G112 (slowloris). That
// deadline covers the whole request and would cut off the long-lived connections behind
// the SSE and websocket flag-change endpoints, which Echo v4 happily kept open. We drop
// ReadTimeout and set ReadHeaderTimeout instead, which keeps the slowloris protection
// without bounding the lifetime of a stream.
func tuneHTTPServer(s *http.Server) error {
	s.ReadTimeout = 0
	s.ReadHeaderTimeout = readHeaderTimeout
	return nil
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
		go s.startMonitoringServer(ctx)
		defer s.stopMonitoringServer(ctx)
	}

	lc := net.ListenConfig{}
	listener, err := lc.Listen(ctx, "unix", socketPath)
	if err != nil {
		s.zapLog.Fatal("Error creating Unix listener", zap.Error(err))
	}

	defer func() {
		if err := listener.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
			s.zapLog.Error("error closing unix socket listener", zap.Error(err))
		}
	}()

	s.zapLog.Info(
		"Starting go-feature-flag relay proxy as unix socket...",
		zap.String("socket", socketPath),
		zap.String("version", s.config.Version))

	ctx, done := s.registerAPIListener(ctx)
	defer close(done)

	err = echo.StartConfig{
		Listener:        listener,
		HideBanner:      true,
		HidePort:        true,
		GracefulTimeout: gracefulTimeout,
		BeforeServeFunc: tuneHTTPServer,
	}.Start(ctx, s.apiEcho)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.zapLog.Fatal("Error starting relay proxy as unix socket", zap.Error(err))
	}
}

// startAsHTTPServer launch the API server
func (s *Server) startAsHTTPServer(ctx context.Context) {
	if s.isMonitoringPortConfigured() {
		go s.startMonitoringServer(ctx)
		defer s.stopMonitoringServer(ctx)
	}

	address := fmt.Sprintf("%s:%d", s.config.ServerHost(), s.config.ServerPort(s.zapLog))
	s.zapLog.Info(
		"Starting go-feature-flag relay proxy ...",
		zap.String("address", address),
		zap.String("version", s.config.Version))

	ctx, done := s.registerAPIListener(ctx)
	defer close(done)

	err := echo.StartConfig{
		Address:         address,
		HideBanner:      true,
		HidePort:        true,
		GracefulTimeout: gracefulTimeout,
		BeforeServeFunc: tuneHTTPServer,
	}.Start(ctx, s.apiEcho)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.zapLog.Fatal("Error starting relay proxy", zap.Error(err))
	}
}

func (s *Server) startMonitoringServer(ctx context.Context) {
	addressMonitoring := fmt.Sprintf("%s:%d", s.config.ServerHost(), s.config.EffectiveMonitoringPort(s.zapLog))
	s.zapLog.Info(
		"Starting monitoring",
		zap.String("address", addressMonitoring))

	ctx, done := s.registerMonitoringListener(ctx)
	defer close(done)

	err := echo.StartConfig{
		Address:         addressMonitoring,
		HideBanner:      true,
		HidePort:        true,
		GracefulTimeout: gracefulTimeout,
		BeforeServeFunc: tuneHTTPServer,
	}.Start(ctx, s.monitoringEcho)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.zapLog.Fatal("Error starting monitoring", zap.Error(err))
	}
}

// registerAPIListener derives a cancellable context for the API listener and records the
// cancel func plus a done channel so Stop can shut it down and wait for it.
func (s *Server) registerAPIListener(ctx context.Context) (context.Context, chan struct{}) {
	ctx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})

	s.mutex.Lock()
	s.apiCancel = cancel
	s.apiDone = done
	s.mutex.Unlock()

	return ctx, done
}

// registerMonitoringListener is registerAPIListener for the monitoring listener.
func (s *Server) registerMonitoringListener(ctx context.Context) (context.Context, chan struct{}) {
	ctx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})

	s.mutex.Lock()
	s.monitoringCancel = cancel
	s.monitoringDone = done
	s.mutex.Unlock()

	return ctx, done
}

// stopMonitoringServer cancels the monitoring listener and waits for it to drain.
func (s *Server) stopMonitoringServer(ctx context.Context) {
	s.mutex.Lock()
	cancel, done := s.monitoringCancel, s.monitoringDone
	s.monitoringCancel, s.monitoringDone = nil, nil
	s.mutex.Unlock()

	waitForListener(ctx, cancel, done)
}

// waitForListener cancels a listener context and waits for the serving goroutine to
// return, giving up if the supplied context is done first.
func waitForListener(ctx context.Context, cancel context.CancelFunc, done chan struct{}) {
	if cancel == nil {
		return
	}
	cancel()
	if done == nil {
		return
	}
	select {
	case <-done:
	case <-ctx.Done():
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

	s.mutex.Lock()
	apiCancel, apiDone := s.apiCancel, s.apiDone
	monitoringCancel, monitoringDone := s.monitoringCancel, s.monitoringDone
	s.apiCancel, s.apiDone = nil, nil
	s.monitoringCancel, s.monitoringDone = nil, nil
	s.mutex.Unlock()

	waitForListener(ctx, monitoringCancel, monitoringDone)
	waitForListener(ctx, apiCancel, apiDone)
}

// isMonitoringPortConfigured checks if the monitoring port is configured.
func (s *Server) isMonitoringPortConfigured() bool {
	return s.monitoringEcho != nil && s.config.EffectiveMonitoringPort(s.zapLog) > 0
}

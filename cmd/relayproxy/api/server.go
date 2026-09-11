package api

import (
	"context"
	"strings"
	"sync"

	"github.com/labstack/echo-contrib/v5/echoprometheus"
	echootel "github.com/labstack/echo-opentelemetry"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	custommiddleware "github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/api/middleware"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/api/opentelemetry"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	controller "github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/handler/goff"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/handler/manifest"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/handler/ofrep"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	helpermiddleware "github.com/thomaspoignant/go-feature-flag/cmdhelpers/api/middleware"
	"go.uber.org/zap"
)

type AuthMiddlewareType = string

const (
	UserAuth  AuthMiddlewareType = "USER"
	AdminAuth AuthMiddlewareType = "ADMIN"
)

// New is used to create a new instance of the API server
func New(config *config.Config,
	services service.Services,
	zapLog *zap.Logger,
) *Server {
	s := &Server{
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

	// Echo v5 drives shutdown through the context passed to StartConfig.Start rather than
	// an Echo.Shutdown method, so the server keeps the cancel func and a done channel per
	// listener in order to implement Stop.
	mutex            sync.Mutex
	apiCancel        context.CancelFunc
	apiDone          chan struct{}
	monitoringCancel context.CancelFunc
	monitoringDone   chan struct{}
}

// initRoutes initialize the API endpoints that contain business logic and specificity for the relay proxy
func (s *Server) initRoutes() {
	// HideBanner/HidePort/Debug were removed from echo.Echo in v5; the first two are now
	// set on echo.StartConfig at start time (see server_lifecycle.go) and Debug has no
	// replacement — debug behaviour here is already driven by the zap logger level.
	s.apiEcho.Use(echootel.NewMiddleware("go-feature-flag"))
	s.apiEcho.Use(helpermiddleware.ZapLogger(s.zapLog, s.config.IsDebugEnabled()))
	s.apiEcho.Use(middleware.BodyDumpWithConfig(middleware.BodyDumpConfig{
		Skipper: func(c *echo.Context) bool {
			isSwagger := strings.HasPrefix(c.Request().URL.String(), "/swagger")
			return isSwagger || !s.zapLog.Core().Enabled(zap.DebugLevel)
		},
		Handler: bodyDumpHandler(s.zapLog),
	}))
	if s.services.Metrics != (metric.Metrics{}) {
		s.apiEcho.Use(echoprometheus.NewMiddlewareWithConfig(echoprometheus.MiddlewareConfig{
			Subsystem:  metric.GOFFSubSystem,
			Registerer: s.services.Metrics.Registry,
			HistogramOptsFunc: func(opts prometheus.HistogramOpts) prometheus.HistogramOpts {
				// Enable native histograms for all middleware histograms
				// This provides higher-fidelity latency distributions while preserving
				// classic histogram buckets for backward compatibility.
				opts.NativeHistogramBucketFactor = 1.1
				opts.NativeHistogramMaxBucketNumber = 160
				return opts
			},
		}))
	}
	s.apiEcho.Use(middleware.CORS("*"))

	s.apiEcho.Use(custommiddleware.VersionHeader(custommiddleware.VersionHeaderConfig{
		Skipper: func(_ *echo.Context) bool {
			return s.config.DisableVersionHeader
		},
		RelayProxyConfig: s.config,
	}))

	s.apiEcho.Use(middleware.Recover())

	// Init controllers
	cAllFlags := controller.NewAllFlags(s.services.FlagsetManager, s.services.Metrics)
	cFlagEval := controller.NewFlagEval(s.services.FlagsetManager, s.services.Metrics)
	cFlagEvalOFREP := ofrep.NewOFREPEvaluate(s.services.FlagsetManager, s.services.Metrics)
	cManifest := manifest.NewManifest(s.services.FlagsetManager, s.services.Metrics, s.zapLog)
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
	userAuth := s.getAuthMiddleware(UserAuth)
	adminAuth := s.getAuthMiddleware(AdminAuth)
	s.addGOFFRoutes(cAllFlags, cFlagEval, cEvalDataCollector, cFlagChangeAPI, cFlagConfiguration, userAuth)
	s.addOFREPRoutes(cFlagEvalOFREP, userAuth)
	s.addStreamRoutes()
	s.addMonitoringRoutes()
	s.addAdminRoutes(cRetrieverRefresh, adminAuth)
	s.addManifestRoutes(cManifest, userAuth)
}

func (s *Server) getAuthMiddleware(middlewareType AuthMiddlewareType) echo.MiddlewareFunc {
	switch middlewareType {
	case AdminAuth:
		return custommiddleware.KeyAuthExtended(custommiddleware.KeyAuthExtendedConfig{
			Validator: func(_ *echo.Context, key string, _ middleware.ExtractorSource) (bool, error) {
				return s.config.APIKeysAdminExists(key), nil
			},
			ErrorHandler: custommiddleware.AuthMiddlewareErrHandler,
		})
	default:
		return custommiddleware.KeyAuthExtended(custommiddleware.KeyAuthExtendedConfig{
			Validator: func(_ *echo.Context, key string, _ middleware.ExtractorSource) (bool, error) {
				return s.config.APIKeyExists(key), nil
			},
			ErrorHandler: custommiddleware.AuthMiddlewareErrHandler,
			Skipper: func(c *echo.Context) bool {
				return !s.config.IsAuthenticationEnabled()
			},
		})
	}
}

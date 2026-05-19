// Package main is the entry point of the GO Feature Flag Management API.
//
// @title       GO Feature Flag Management API
// @version     0.1.0
// @description Source-of-truth API for managing GO Feature Flag flagsets, flags, teams, and audit history.
// @BasePath    /api/v1
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/pflag"
	"go.uber.org/zap"

	"github.com/thomaspoignant/go-feature-flag/cmd/management/api"
	"github.com/thomaspoignant/go-feature-flag/cmd/management/config"
	_ "github.com/thomaspoignant/go-feature-flag/cmd/management/docs"
	"github.com/thomaspoignant/go-feature-flag/cmd/management/handler"
	"github.com/thomaspoignant/go-feature-flag/cmd/management/repository"
	"github.com/thomaspoignant/go-feature-flag/cmd/management/service"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "fatal:", err)
		os.Exit(1)
	}
}

func run() error {
	fs := pflag.NewFlagSet("goff-management", pflag.ExitOnError)
	config.RegisterFlags(fs)
	if err := fs.Parse(os.Args[1:]); err != nil {
		return err
	}

	cfg, err := config.Load(fs)
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return err
	}

	log, err := newLogger(cfg.Log)
	if err != nil {
		return err
	}
	defer func() { _ = log.Sync() }()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	db, err := repository.NewDB(ctx, cfg.DB)
	if err != nil {
		return err
	}
	defer db.Close()

	userRepo := repository.NewUserRepo(db)
	teamRepo := repository.NewTeamRepo(db)
	flagsetRepo := repository.NewFlagsetRepo(db)
	flagRepo := repository.NewFlagRepo(db)
	versionRepo := repository.NewFlagVersionRepo(db)
	auditRepo := repository.NewAuditRepo(db)

	authSvc, err := service.NewAuthService(ctx, *cfg, userRepo)
	if err != nil {
		return err
	}
	teamSvc := service.NewTeamService(teamRepo, userRepo, auditRepo)
	onboardingSvc := service.NewOnboardingService(teamRepo, auditRepo)
	flagsetSvc := service.NewFlagsetService(flagsetRepo, auditRepo)
	flagSvc := service.NewFlagService(db, flagRepo, versionRepo, flagsetRepo, auditRepo)
	versionSvc := service.NewVersionService(db, flagRepo, versionRepo, flagsetRepo, auditRepo)
	auditSvc := service.NewAuditService(auditRepo)

	handlers := api.Handlers{
		Auth:       handler.NewAuthHandler(*cfg, authSvc, userRepo, teamSvc),
		Teams:      handler.NewTeamHandler(teamSvc),
		Flagsets:   handler.NewFlagsetHandler(flagsetSvc),
		Flags:      handler.NewFlagHandler(flagSvc),
		Versions:   handler.NewVersionHandler(versionSvc),
		Audit:      handler.NewAuditHandler(auditSvc),
		Onboarding: handler.NewOnboardingHandler(onboardingSvc),
	}
	services := api.Services{Auth: authSvc, Teams: teamSvc, Flagsets: flagsetSvc, Flags: flagSvc}

	e := api.New(*cfg, log, handlers, services)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{Addr: addr, Handler: e, ReadHeaderTimeout: 10 * time.Second}

	errCh := make(chan error, 1)
	go func() {
		log.Info("starting server", zap.String("addr", addr))
		errCh <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		log.Info("shutdown signal received")
		shutdownCtx, c2 := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
		defer c2()
		return srv.Shutdown(shutdownCtx)
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func newLogger(cfg config.LogConfig) (*zap.Logger, error) {
	zc := zap.NewProductionConfig()
	if cfg.Format == "console" {
		zc = zap.NewDevelopmentConfig()
	}
	if cfg.Level != "" {
		lvl, err := zap.ParseAtomicLevel(cfg.Level)
		if err != nil {
			return nil, err
		}
		zc.Level = lvl
	}
	return zc.Build()
}

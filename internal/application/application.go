package application

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ikondratev/event-service/internal/auth"
	"github.com/ikondratev/event-service/internal/grpcserver"
	"github.com/ikondratev/event-service/internal/httpserver"
	"github.com/ikondratev/event-service/internal/logger"
	eventsvc "github.com/ikondratev/event-service/internal/service/event"
	"github.com/ikondratev/event-service/internal/settings"
	"github.com/ikondratev/event-service/internal/storage/postgres"

	eventrepo "github.com/ikondratev/event-service/internal/repo/postgres"
)

type Application struct {
	logger   *slog.Logger
	settings *settings.Settings
	server   *httpserver.HttpServer
	db       *postgres.Db
	grpc     *grpcserver.Listener
}

func New(env string) (*Application, error) {
	logger := logger.New(env)

	// Ini settings
	settings, err := settings.New(env)
	if err != nil {
		return nil, fmt.Errorf("load settings: %w", err)
	}

	// Init verifier
	verifier, err := auth.NewVerifier(
		settings.Auth.Issuer,
		settings.Auth.Audience,
		settings.Auth.PublicKeyPath,
	)
	if err != nil {
		return nil, fmt.Errorf("auth verifier: %w", err)
	}

	// Init db
	database, err := postgres.New(context.Background(), settings)
	if err != nil {
		return nil, fmt.Errorf("open db error: %w", err)
	}

	// Init repos
	eventRepo := eventrepo.NewEventRepo(database.Adapter, settings)
	idemRepo := eventrepo.NewIdempotencyRepo(database.Adapter)

	// Services
	tx := postgres.NewTransactor(database.Adapter)
	service := eventsvc.New(tx, idemRepo, eventRepo)

	// gRpc
	grpcHandler := grpcserver.New(logger, service, verifier)
	grpcListerner := grpcserver.NewListener(settings, logger, grpcHandler)

	// Init routes
	httpServer := httpserver.NewHTTPServer(settings, logger)

	// Finaly App
	return &Application{
		logger:   logger,
		settings: settings,
		server:   httpServer,
		db:       database,
		grpc:     grpcListerner,
	}, nil
}

func (a *Application) Run() error {
	chErr := make(chan error, 2)
	go a.server.Start(chErr)
	go a.grpc.Start(chErr)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	if err := a.waitShutdown(chErr, quit); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(a.settings.Server.WaitingShutdown)*time.Second,
	)
	defer cancel()

	// Stop gRpc
	if err := a.grpc.Stop(ctx); err != nil {
		a.logger.Error("gRpc error", "err", err)
		return fmt.Errorf("grpc shuddown error: %w", err)
	}

	// Stop http
	if err := a.server.Stop(ctx); err != nil {
		a.logger.Error("http server error", "err", err)
		return fmt.Errorf("server shutdown error: %w", err)
	}

	// Stop DB
	if a.db != nil {
		if err := a.db.Adapter.Close(); err != nil {
			return fmt.Errorf("db close error: %w", err)
		}
		a.logger.Info("DB closed")
	}

	a.logger.Info("Server stopped gracefully")
	return nil
}

func (a *Application) waitShutdown(errCh <-chan error, quit <-chan os.Signal) error {
	select {
	case err := <-errCh:
		if !httpserver.IsStopped(err) && !grpcserver.IsStopped(err) {
			a.logger.Error("stop error", "err", err)
			return fmt.Errorf("server error: %w", err)
		}
		return nil

	case sig := <-quit:
		a.logger.Info("Server receive signal:", "signal", sig)
		return nil
	}
}

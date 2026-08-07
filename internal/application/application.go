package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"microserice/internal/logger"
	"microserice/internal/router"
	"microserice/internal/settings"
	"microserice/internal/storage/postgres"

	eventrepo "microserice/internal/repo/postgres"
)

type Application struct {
	logger   *slog.Logger
	settings *settings.Settings
	server   *http.Server
	db 		 *postgres.Db
}

func New(env string) (*Application, error) {
	logger := logger.New(env)

	// Ini settings
	settings, err := settings.New(env)
	if err != nil {
		return nil, fmt.Errorf("Error: Load settings: %w", err)
	}

	// Init db
	database, err := postgres.New(context.Background(), settings)
	if err != nil {
		return nil, fmt.Errorf("Open db error: %w", err)
	}

	// Init repo
	eventRepo := eventrepo.NewEventRepo(database.Adapter)

	// Init routes
	routes := router.New(logger, eventRepo).RegisterRotes()
	server := &http.Server{
		Addr:              settings.Server.Port,
		Handler:           routes,
		ReadHeaderTimeout: time.Duration(settings.Server.HeaderTimeout) * time.Second,
		ReadTimeout:       time.Duration(settings.Server.ReadTimeout)   * time.Second,
		WriteTimeout:      time.Duration(settings.Server.WriteTimeout)  * time.Second,
		IdleTimeout:       time.Duration(settings.Server.IdleTimeout)   * time.Second,
	}
	
	// Finaly App
	return &Application{
		logger:   logger,
		settings: settings,
		server:   server,
		db: 	  database,
	}, nil
}

func (a *Application) Run() error {
	chErr := make(chan error, 1)
	go a.startServer(chErr)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	if err := a.waitShutdown(chErr, quit); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(a.settings.Server.WaitingShutdown)*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	if a.db != nil {
		if err := a.db.Adapter.Close(); err != nil {
			return fmt.Errorf("db close error: %w", err)
		}
		a.logger.Info("DB closed")
	}

	a.logger.Info("Server stopped gracefully")
	return nil
}

func (a *Application) startServer(errCh chan<- error) {
	a.logger.Info("Server started...", "addr", a.server.Addr)
	errCh <- a.server.ListenAndServe()
}

func (a *Application) waitShutdown(errCh <-chan error, quit <-chan os.Signal) error {
	select {
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("Server error %w", err)
		}
		return nil

	case sig := <-quit:
		a.logger.Info("Server receive signal:", "signal", sig)
		return nil
	}
}

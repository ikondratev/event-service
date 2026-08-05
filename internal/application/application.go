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
)

type Application struct {
	logger   *slog.Logger
	settings *settings.Settings
	server   *http.Server
}

func New(env string) (*Application, error) {
	logger := logger.New(env)

	settings, err := settings.New(env)
	if err != nil {
		return nil, fmt.Errorf("Error: Load settings: %w", err)
	}

	routes := router.New(logger).RegisterRotes()
	server := &http.Server {
		Addr:    		   settings.Port,
		Handler: 		   routes,
		ReadHeaderTimeout: 5  * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,

	}

	return &Application{
		logger:   logger,
		settings: settings,
		server:   server,
	}, nil
}

func (a *Application) startServer(chErr chan <- error) {
	a.logger.Info("Server started", "addr", a.server.Addr)
	chErr <- a.server.ListenAndServe()
}

func (a *Application) waitForShutdown(chErr <- chan error, quit <- chan os.Signal) error {
	select {
	case err := <- chErr:
		if !errors.Is(err, http.ErrServerClosed) {
			a.logger.Error("Server error", "error", err)
			return fmt.Errorf("Server error: %w", err)
		}
		return nil
		
	case sig := <- quit:
		a.logger.Info("Server received signal", "signal", sig)
		return nil
	}
}
 
func (a *Application) Run() error {
	errCh := make(chan error, 1)
	go a.startServer(errCh)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	if err := a.waitForShutdown(errCh, quit); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(a.settings.WaitingShutdown)*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("Server shudwon failed: %w", err)
	}

	a.logger.Info("Server stopped gracefully")
	return nil
}
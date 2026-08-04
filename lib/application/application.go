package application

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"microserice/lib/logger"
	"microserice/lib/router"
	"microserice/lib/settings"
)

type Application struct {
	logger   *slog.Logger
	settings *settings.Settings
	server   *http.Server
}

func New(env string) (*Application, error) {
	logger := logger.New(env)
	settings := settings.New(env)
	router := router.New(logger)

	server := &http.Server {
		Addr:    		   settings.Port,
		Handler: 		   router.Register(),
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

func (a *Application) Run() error {
	if err := a.server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		a.logger.Error("Server error:", "error", err)
		return fmt.Errorf("Error: Application: %w", err)
	}

	return nil
}
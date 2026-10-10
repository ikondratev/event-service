package httpserver

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/ikondratev/event-service/internal/settings"
)

type HttpServer struct {
	engine   *http.Server
	settings *settings.Settings
	logger   *slog.Logger
}

func NewHTTPServer(s *settings.Settings, logger *slog.Logger) *HttpServer {
	router := NewRouter(logger)
	handler := router.RegisterRotes()

	server := &http.Server{
		Addr:              s.Server.Port,
		Handler:           handler,
		ReadHeaderTimeout: time.Duration(s.Server.HeaderTimeout) * time.Second,
		ReadTimeout:       time.Duration(s.Server.ReadTimeout) * time.Second,
		WriteTimeout:      time.Duration(s.Server.WriteTimeout) * time.Second,
		IdleTimeout:       time.Duration(s.Server.IdleTimeout) * time.Second,
	}

	return &HttpServer{
		engine:   server,
		settings: s,
		logger:   logger,
	}
}

func (s *HttpServer) Start(errCh chan error) {
	s.logger.Info("http server started", "addr", s.engine.Addr, "env", s.settings.Environment)
	errCh <- s.engine.ListenAndServe()
}

func (s *HttpServer) Stop(ctx context.Context) error {
	return s.engine.Shutdown(ctx)
}

func IsStopped(err error) bool {
	return errors.Is(err, http.ErrServerClosed)
}

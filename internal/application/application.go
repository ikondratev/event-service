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

	"github.com/ikondratev/event-service/internal/kafka"
	"github.com/ikondratev/event-service/internal/logger"
	"github.com/ikondratev/event-service/internal/router"
	"github.com/ikondratev/event-service/internal/settings"
	"github.com/ikondratev/event-service/internal/storage/postgres"

	worker "github.com/ikondratev/event-service/internal/workers"
	eventrepo "github.com/ikondratev/event-service/internal/repo/postgres"
)

type Application struct {
	logger   	 *slog.Logger
	settings 	 *settings.Settings
	server   	 *http.Server
	db 		 	 *postgres.Db
	kproducer 	 *kafka.Producer
	worker   	 *worker.OutboxWorker
	workerCancel context.CancelFunc
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

	// Init repos
	eventRepo := eventrepo.NewEventRepo(database.Adapter, settings)
	outboxRepo := eventrepo.NewOutboxRepo(database.Adapter)

	// Kafa
	producer := kafka.NewProducer(settings)

	// Workers
	outboxWorker := worker.NewOutboxWorker(logger, outboxRepo, producer, settings)

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
		logger:    logger,
		settings:  settings,
		server:    server,
		db: 	   database,
		kproducer: producer,
		worker:    outboxWorker,
	}, nil
}

func (a *Application) Run() error {
	workerCtx, workerCancel := context.WithCancel(context.Background())
	a.workerCancel = workerCancel

	go a.worker.Run(workerCtx)

	chErr := make(chan error, 1)
	go a.startServer(chErr)

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

	// Stop http
	if err := a.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	// Stop worker
	if a.workerCancel != nil {
		a.workerCancel()
		a.worker.Wait()
		a.logger.Info("Outbox worker canceled")
	}

	// Stop kafka
	if a.kproducer != nil {
		flushCtx, flushCancel := context.WithTimeout(
			ctx, 
			time.Duration(a.settings.Kafka.FlushTimeout)*time.Second,
		)
		if err := a.kproducer.Flush(flushCtx); err != nil {
			a.logger.Error("Flush Kafka error", "error", err)
		}
		flushCancel()

		if err := a.kproducer.Close(); err != nil {
			return fmt.Errorf("Stop producer error: %w", err)
		}

		a.logger.Info("Kafka producer closed")
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

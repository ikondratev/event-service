package application

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ikondratev/event-service/internal/kafka"
	"github.com/ikondratev/event-service/internal/logger"
	"github.com/ikondratev/event-service/internal/settings"
	"github.com/ikondratev/event-service/internal/storage/postgres"

	eventrepo "github.com/ikondratev/event-service/internal/repo/postgres"
	worker "github.com/ikondratev/event-service/internal/workers"
)

type Worker struct {
	logger *slog.Logger
	settings *settings.Settings
	producer *kafka.Producer
	db *postgres.Db
	kfWorker *worker.OutboxWorker
}

func NewWorker(env string) (*Worker, error) {
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
	outboxRepo := eventrepo.NewOutboxRepo(database.Adapter)

	// Kafa
	producer := kafka.NewProducer(settings)

	// Workers
	outboxWorker := worker.NewOutboxWorker(logger, outboxRepo, producer, settings)

	return &Worker {
		logger: logger,
		settings: settings,
		producer: producer,
		db: database,
		kfWorker: outboxWorker,
	}, nil
}

func (w *Worker) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go w.kfWorker.Run(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <- quit
	w.logger.Info("outbox worker received signal", "signal", sig)

	cancel()
	w.kfWorker.Wait()
	w.logger.Info("outbox worker stopped")

	// Stop producer
	if w.producer != nil {
		flushCtx, flushCancel := context.WithTimeout(
			context.Background(),
			time.Duration(w.settings.Kafka.FlushTimeout)*time.Second,
		)
		if err := w.producer.Flush(flushCtx); err != nil {
			w.logger.Error("flush kafka error", "error", err)
		}
		flushCancel()
		if err := w.producer.Close(); err != nil {
			return fmt.Errorf("stop producer error: %w", err)
		}
		w.logger.Info("kafka producer closed")
	}

	// Stop DB
	if w.db != nil {
		if err := w.db.Adapter.Close(); err != nil {
			return fmt.Errorf("db close error: %w", err)
		}
		w.logger.Info("db closed")
	}

	w.logger.Info("worker stopped gracefully")
	return nil
}

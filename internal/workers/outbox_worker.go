package worker

import (
	"context"
	"fmt"
	"log/slog"
	"github.com/ikondratev/event-service/internal/kafka"
	"github.com/ikondratev/event-service/internal/repo/postgres"
	"github.com/ikondratev/event-service/internal/settings"
	"sync"
	"time"
)

type OutboxWorker struct {
	logger 	 *slog.Logger
	repo 	 *postgres.OutboxRepo
	producer kafka.Publisher
	settings *settings.Settings
	done	 sync.WaitGroup
}

func NewOutboxWorker(
	logger 	  *slog.Logger,
	repo 	  *postgres.OutboxRepo,
	publisher kafka.Publisher,
	settings  *settings.Settings,
) *OutboxWorker {
	return &OutboxWorker{
		logger: logger,
		repo: repo,
		producer: publisher,
		settings: settings,
	}
}

func (w *OutboxWorker) Run(ctx context.Context) {
	w.done.Add(1)
	defer w.done.Done()

	ticker := time.NewTicker(time.Duration(w.settings.Kafka.PollInterval)*time.Second)
	defer ticker.Stop()

	w.logger.Info("OutboxWorker strarted", 
		"poll_interval", w.settings.Kafka.PollInterval,
		"batch_size", w.settings.Kafka.BatchSize,
	)

	for {
		select {
		case <- ctx.Done():
			w.logger.Info("OutboxWorker stopped")
			return
		case <- ticker.C:
			w.processBatch(ctx)
		}
	}
}

func (w *OutboxWorker) processBatch(ctx context.Context) {
	msgs, err := w.repo.FetchPending(ctx, w.settings.Kafka.BatchSize)
	if err != nil {
		w.logger.Error("Worker batch:", "error", err)
		return
	}

	for _, msg := range msgs {
		key := []byte(fmt.Sprint(msg.EventID))

		if err := w.producer.Publish(ctx, msg.Topic, key, msg.Payload); err != nil {
			w.logger.Error("kafka publish failed",
				"outbox_id", msg.ID,
				"topic", msg.Topic,
				"error", err,
			)
			continue
		}

		if err := w.repo.MarkSent(ctx, msg.ID); err != nil {
			w.logger.Error("mark sent failed",
				"outbox_id", msg.ID,
				"error", err,
			)
			continue
		}

		w.logger.Info("outbox message sent",
			"outbox_id", msg.ID,
			"event_id", msg.EventID,
			"topic", msg.Topic,
		)
	}
}

func (w *OutboxWorker) Wait() {
	w.done.Wait()
}
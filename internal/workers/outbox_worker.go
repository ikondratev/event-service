package worker

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/ikondratev/event-service/internal/kafka"
	"github.com/ikondratev/event-service/internal/repo/postgres"
	"github.com/ikondratev/event-service/internal/settings"
)

type OutboxWorker struct {
	logger 	   *slog.Logger
	repo 	   *postgres.OutboxRepo
	producer   kafka.Publisher
	settings   *settings.Settings
	done	   sync.WaitGroup
	instanceID string 
}

func NewOutboxWorker(
	logger 	   *slog.Logger,
	repo 	   *postgres.OutboxRepo,
	publisher  kafka.Publisher,
	settings   *settings.Settings,
) *OutboxWorker {
	instanceID := os.Getenv("POD_NAME")
	if instanceID == "" {
		hostname, err := os.Hostname()
		if err != nil {
			instanceID = "unknown"
		} else {
			instanceID = hostname
		}
	}

	return &OutboxWorker{
		logger: logger,
		repo: repo,
		producer: publisher,
		settings: settings,
		instanceID: instanceID,
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
		"instance_id", w.instanceID,
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

func (w *OutboxWorker) Wait() {
	w.done.Wait()
}

func (w *OutboxWorker) processBatch(ctx context.Context) {
	lease := time.Duration(w.settings.Worker.LeaseSeconds) * time.Second
	backoff := time.Duration(w.settings.Worker.BackoffSeconds) * time.Second

	msgs, err := w.repo.Claim(
		ctx, 
		w.instanceID, 
		w.settings.Kafka.BatchSize, 
		lease,
	)
	if err != nil {
		w.logger.Error("Worker batch:", "error", err)
		return
	}

	workCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), time.Duration(w.settings.Worker.TimeoutSeconds)*time.Second)
	defer cancel()

	for _, msg := range msgs {
		key := []byte(fmt.Sprint(msg.EventID))
		if err := w.producer.Publish(
			workCtx, 
			msg.Topic, 
			key, 
			msg.Payload,
		); err != nil {
			w.logger.Error("kafka publish failed",
				"outbox_id", msg.ID,
				"topic", msg.Topic,
				"error", err,
			)
			if nackErr := w.repo.Nack(
				workCtx,
				msg.ID,
				w.instanceID,
				err.Error(),
				w.settings.Worker.MaxAttempts,
				backoff,
			); nackErr != nil {
				w.logger.Error("nack failed",
					"outbox_id", msg.ID,
					"error", nackErr,
				)
			}
			continue
		}

		if err := w.repo.MarkSent(workCtx, msg.ID, w.instanceID); err != nil {
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

package eventsvc

import (
	"context"
	"fmt"

	"github.com/ikondratev/event-service/internal/domain/event"
	"github.com/ikondratev/event-service/internal/domain/idempotency"
)

type Transactor interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context)error) error
}

type Service struct {
	tx     Transactor
	keys   idempotency.IdempotencyRepo
	events event.EventRepo
}

func New(
	tx Transactor, 
	keys idempotency.IdempotencyRepo, 
	events event.EventRepo,
) (*Service) {
	return &Service{tx: tx, keys: keys, events: events}
}

type CreateResult struct {
	Event 	 event.Event
	Replayed bool
}

func (s *Service) Create(
	ctx context.Context,
	key, requestHash string,
	e *event.Event,
) (CreateResult, error) {
	var result CreateResult

	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		inserted, err := s.keys.TryInsert(ctx, key, requestHash)
		if err != nil {
			return err
		}

		if inserted {
			if err := s.events.Create(ctx, e); err != nil {
				return err
			}
			if err := s.keys.AttachEvent(ctx, key, e.ID); err != nil {
				return err
			}
			result = CreateResult{Event: *e, Replayed: false}
			return nil
		}

		existing, err := s.keys.Get(ctx, key)
		if err != nil {
			return err
		}
		if existing.RequestHash != requestHash {
			return idempotency.ErrPayloadMismatch
		}
		if existing.EventID == nil {
			return fmt.Errorf("idempotncy key has no event")
		}

		got, err := s.events.GetByID(ctx, *existing.EventID)
		if err != nil {
			return err
		}
		result = CreateResult{Event: *got, Replayed: true}
		return nil
	})

	return result, err
}

func (s *Service) List(ctx context.Context) ([]event.Event, error) {
	return s.events.List(ctx)
}
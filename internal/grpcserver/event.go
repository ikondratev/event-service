package grpcserver

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/ikondratev/event-service/internal/auth"
	"github.com/ikondratev/event-service/internal/domain/idempotency"
	eventdto "github.com/ikondratev/event-service/internal/dto/event"
	"github.com/ikondratev/event-service/internal/eventpb"

	eventsvc "github.com/ikondratev/event-service/internal/service/event"
)

type Server struct {
	eventpb.UnimplementedEventServiceServer
	log      *slog.Logger
	events   *eventsvc.Service
	verifier *auth.Verifier
}

func New(
	log *slog.Logger,
	events *eventsvc.Service,
	verifier *auth.Verifier,
) *Server {
	return &Server{
		log:      log,
		events:   events,
		verifier: verifier,
	}
}

func (s *Server) CreateEvent(
	ctx context.Context,
	in *eventpb.CreateEventRequest,
) (*eventpb.CreateEventResponse, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	authorization := first(md.Get("authorization"))
	idempotencyKey := first(md.Get("idempotency-key"))
	traceID := first(md.Get("x-trace-id"))

	log := s.log.With("trace_id", traceID)
	token, ok := strings.CutPrefix(authorization, "Bearer")
	if !ok || strings.TrimSpace(token) == "" {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	identity, err := s.verifier.Parse(strings.TrimSpace(token))
	if err != nil || identity.Subject == "" || !identity.HasScope("events:write") {
		if err != nil || identity.Subject == "" {
			return nil, status.Error(codes.Unauthenticated, "unauthorized")
		}
		return nil, status.Error(codes.PermissionDenied, "forbidden")
	}

	if err := idempotency.ValidateKey(idempotencyKey); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	req := toCreateRequest(in)
	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid event")
	}

	hash, err := eventdto.Hash(req)
	if err != nil {
		log.Error("hash event", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	domainEvent, err := eventdto.ToDomain(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid event")
	}

	result, err := s.events.Create(ctx, idempotencyKey, hash, domainEvent)
	if err != nil {
		if errors.Is(err, idempotency.ErrPayloadMismatch) {
			return nil, status.Error(codes.AlreadyExists, "idempotency key conflict")
		}
		log.Error("create event", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	log.Info("event created",
		"event_id", result.Event.ID,
		"replayed", result.Replayed,
	)

	return &eventpb.CreateEventResponse{
		Id:        result.Event.ID,
		Kind:      result.Event.Kind,
		CreatedAt: result.Event.CreatedAt.UTC().Format(time.RFC3339),
	}, nil
}

func toCreateRequest(in *eventpb.CreateEventRequest) eventdto.CreateRequest {
	data := in.GetData()
	payment := data.GetPaymentData()
	location := payment.GetLocation()
	device := data.GetPaymentDevice()

	return eventdto.CreateRequest{
		Kind: in.GetKind(),
		Data: eventdto.EventData{
			PaymentData: eventdto.PaymentData{
				AccountID: payment.GetAccountId(),
				IP:        payment.GetIpAddres(),
				Location: eventdto.Location{
					Lat: location.GetLat(),
					Lon: location.GetLon(),
				},
				Amount:    payment.GetAmount(),
				Currency:  payment.GetCurrency(),
				Timestamp: payment.GetTimestamp(),
				UserEmail: payment.GetUserEmail(),
				UserPhone: payment.GetUserPhone(),
			},
			PayDevice: eventdto.PayDevice{
				Token: device.GetToken(),
				Type:  device.GetType(),
			},
		},
	}
}

func first(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

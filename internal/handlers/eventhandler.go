package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/ikondratev/event-service/internal/domain/idempotency"
	eventdto "github.com/ikondratev/event-service/internal/dto/event"
	"github.com/ikondratev/event-service/internal/service/event"
)

type EventHandler struct {
	logger    *slog.Logger
	service	  *eventsvc.Service
}

func NewEventHandler(logger *slog.Logger, service *eventsvc.Service) * EventHandler {
	return &EventHandler{
		logger:  logger,
		service: service,
	}
}

func (h *EventHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get("Idempotency-Key")
	if err := idempotency.ValidateKey(key); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req eventdto.CreateRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	
	if err := decoder.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := req.Validate(); err != nil {
		if verr, ok := err.(*eventdto.ValidationError); ok {
			writeJSON(w, http.StatusBadRequest, map[string]any{"errors": verr.Errors})
			return
		}
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	hash, err := eventdto.Hash(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	domainEvent, err := eventdto.ToDomain(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid event data")
		return
	}

	result, err := h.service.Create(r.Context(), key, hash, domainEvent)
	if err != nil {
		if errors.Is(err, idempotency.ErrPayloadMismatch) {
			writeError(w, http.StatusConflict, "idempotency key conflict")
			return
		}
		h.logger.Error("failed to create event", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to create event")
		return
	}

	status := http.StatusCreated
	if result.Replayed {
		status = http.StatusOK
	}

	writeJSON(w, status, eventdto.ToResponse(result.Event))
}

func (h *EventHandler) GetEvents(w http.ResponseWriter, r *http.Request) {
	events, err := h.service.List(r.Context())
	if err != nil {
		h.logger.Error("failed to list events", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to get events")
		return
	}

	response := make([]eventdto.Response, 0, len(events))
	for _, e := range events {
		response = append(response, eventdto.ToResponse(e))
	}
	
	writeJSON(w, http.StatusOK, response)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
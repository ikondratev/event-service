package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/ikondratev/event-service/internal/domain/event"
	eventdto "github.com/ikondratev/event-service/internal/dto/event"
)

type EventHandler struct {
	logger    *slog.Logger
	repo      event.EventRepo
}

func NewEventHandler(logger *slog.Logger, eventRepo event.EventRepo) * EventHandler {
	return &EventHandler{
		logger: logger,
		repo: 	eventRepo,
	}
}

func (h *EventHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
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

	domainEvent, err := eventdto.ToDomain(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid event data")
		return
	}

	if err := h.repo.Create(r.Context(), domainEvent); err != nil {
		h.logger.Error("failed to create event", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to create event")
		return
	}

	writeJSON(w, http.StatusCreated, eventdto.ToResponse(*domainEvent))
}

func (h *EventHandler) GetEvents(w http.ResponseWriter, r *http.Request) {
	events, err := h.repo.List(r.Context())
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
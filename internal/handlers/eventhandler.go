package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type Event struct {
	Kind   string `json:"kind"`
	Data   string `json:"data"`
	Status string `json:"status"`
}

type EventHandler struct {
	logger *slog.Logger
}

func NewEventHandler(logger *slog.Logger) * EventHandler {
	return &EventHandler{logger: logger}
}

func (h *EventHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	var event Event

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	
	if err := decoder.Decode(&event); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if event.Kind == "" || event.Data == "" || event.Status == "" {
		http.Error(w, "kind, data, status are requeired", http.StatusBadRequest)
		return
	}

	h.logger.Info("event created", "event", event)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(event)
}

func (h *EventHandler) GetEvents(w http.ResponseWriter, r *http.Request) {
	event := Event{
		Kind:   "LOGO",
		Data:   "Data",
		Status: "Status",
	}
	
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(event)
	if err != nil {
		http.Error(w, "failed to encode", http.StatusInternalServerError)
		return
	}
}
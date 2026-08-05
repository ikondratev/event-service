package router

import (
	"log/slog"
	"microserice/internal/handlers"
	"microserice/internal/middleware"
	"net/http"

	"github.com/gorilla/mux"
)

type Router struct {
	engine *mux.Router
	logger *slog.Logger
}

func New(logger *slog.Logger) *Router {
	router := mux.NewRouter()
	router.Use(middleware.Logging(logger))

	return &Router{
		engine: router,
		logger: logger,
	}
}

func (r *Router) RegisterRotes() *mux.Router {
	eventHandler := handlers.NewEventHandler(r.logger)
	r.engine.HandleFunc("/ping", handlers.Pong).Methods(http.MethodGet)
	api := r.engine.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/event", eventHandler.CreateEvent).Methods(http.MethodPost)
	api.HandleFunc("/events", eventHandler.GetEvents).Methods(http.MethodGet)

	return r.engine
}
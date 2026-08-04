package router

import (
	"log/slog"
	"microserice/lib/handlers"
	"microserice/lib/middleware"
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

func (r *Router) registerHandlers() {
	r.engine.HandleFunc("/ping", handlers.Pong).Methods(http.MethodGet)
}

func (r *Router) Register() *mux.Router {
	r.registerHandlers()
	return r.engine
}
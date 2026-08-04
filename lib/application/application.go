package application

import (
	"log/slog"

	"microserice/lib/logger"
)

type Application struct {
	logger *slog.Logger
	router *router.Router
}

func New(env string) *Application {
	logger := logger.New(env)
	return &Application{
		logger: logger,
	}
}

func Run() {

}
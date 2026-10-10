package main

import (
	"os"

	"github.com/ikondratev/event-service/internal/application"
)

func main() {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev"
	}

	app, err := application.New(env)
	if err != nil {
		os.Exit(1)
	}

	if err := app.Run(); err != nil {
		os.Exit(1)
	}
}

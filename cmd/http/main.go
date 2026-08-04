package main

import (
	"os"
	"log/slog"

	"microserice/lib/application"
)

const env = "DEV"

func main() {
	app := application.New(env)
	if err := app.Run(); err != nil {
		slog.Error("app exited with error", "error", err)
		os.Exit(1)
	}
}
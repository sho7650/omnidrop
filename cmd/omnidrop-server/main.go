package main

import (
	"log/slog"
	"os"

	"omnidrop/internal/app"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	// Logger is configured inside application.Run() (first thing it does),
	// so any error returned below logs through the configured slog default.
	application := app.NewWithVersion(Version, BuildTime)
	if err := application.Run(); err != nil {
		slog.Error("Application failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

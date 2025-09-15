package main

import (
	"log"

	"omnidrop/internal/app"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	application := app.NewWithVersion(Version, BuildTime)
	if err := application.Run(); err != nil {
		log.Fatalf("Application failed: %v", err)
	}
}
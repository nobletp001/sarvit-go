package main

import (
	"log"

	"github.com/nobletp001/sarvit/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatalf("❌ App exited with error: %v", err)
	}
}

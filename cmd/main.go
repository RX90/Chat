package main

import (
	"log"

	"github.com/RX90/Chat/internal/app"
)

func main() {
	a, err := app.NewApp()
	if err != nil {
		log.Fatalf("failed to initialize app: %v", err)
	}
	a.Run()
}

// Package main is the entry point of the application.
package main

import (
	"context"
	"log"

	"github.com/sanchey92/metric-agent/internal/app"
	"github.com/sanchey92/metric-agent/internal/config"
)

func main() {
	cfg := config.New()
	ctx := context.Background()

	a := app.New(cfg)

	if err := a.Run(ctx); err != nil {
		log.Fatal("error")
	}
}

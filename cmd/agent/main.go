// Package main is the entry point of the application.
package main

import (
	"fmt"

	"github.com/sanchey92/metric-agent/internal/config"
)

func main() {
	cfg := config.New()

	fmt.Println(cfg)
}

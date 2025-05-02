// Package config provides configuration management for the application.
// It handles environment variables parsing and provides a structured way
// to access application settings.
package config

import (
	"os"
	"strconv"
	"time"
)

// Config represents the application configuration structure.
// It contains all configurable parameters needed across the application.
type Config struct {
	ServerAddr     string
	PollInterval   time.Duration
	ReportInterval time.Duration
}

// New creates and initializes a new Config instance.
// It reads environment variables with fallback to default values.
// Returns a pointer to the initialized Config.
func New() *Config {
	addr := getEnv("SERVER_ADDRESS", "http://localhost:8080/update")
	pollSec := getEnvAsInt("POLL_INTERVAL", 2)
	reportSec := getEnvAsInt("REPORT_INTERVAL", 10)

	return &Config{
		ServerAddr:     addr,
		PollInterval:   time.Duration(pollSec) * time.Second,
		ReportInterval: time.Duration(reportSec) * time.Second,
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvAsInt(name string, defaultValue int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}

	return defaultValue
}

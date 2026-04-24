// Package config provides centralized configuration for the benchmark servers.
package config

import (
	"os"
	"time"
)

// Config holds all runtime configuration for the application.
type Config struct {
	GRPCAddr        string
	RESTAddr        string
	ShutdownTimeout time.Duration
}

// Default returns a Config populated with sensible defaults.
// Values can be overridden via environment variables:
//
//	GRPC_ADDR  - gRPC listen address (default: ":50051")
//	REST_ADDR  - REST listen address (default: ":8080")
func Default() *Config {
	return &Config{
		GRPCAddr:        getEnv("GRPC_ADDR", ":50051"),
		RESTAddr:        getEnv("REST_ADDR", ":8080"),
		ShutdownTimeout: 5 * time.Second,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

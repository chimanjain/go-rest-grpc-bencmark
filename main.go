// Command go-rest-grpc-benchmark starts both a REST and gRPC benchmark server and waits
// for an OS interrupt signal before shutting down gracefully.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/chimanjain/go-rest-grpc-benchmark/config"
	"github.com/chimanjain/go-rest-grpc-benchmark/server"
)

func main() {
	cfg := config.Default()

	// Start gRPC server.
	grpcSrv, err := server.NewGRPCServer(cfg.GRPCAddr)
	if err != nil {
		slog.Error("failed to start gRPC server", "error", err)
		os.Exit(1)
	}

	// Start REST server.
	restSrv, err := server.NewRESTServer(cfg.RESTAddr)
	if err != nil {
		slog.Error("failed to start REST server", "error", err)
		os.Exit(1)
	}

	slog.Info("servers are ready",
		"grpc", grpcSrv.Addr(),
		"rest", restSrv.Addr(),
	)
	slog.Info("run benchmarks with: go test -bench . -benchmem ./...")

	// Block until an OS termination signal is received.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutdown signal received; draining connections...")

	// Graceful shutdown with a deadline.
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	grpcSrv.Stop()

	if err := restSrv.Shutdown(ctx); err != nil {
		slog.Error("REST server forced shutdown", "error", err)
	}

	slog.Info("all servers stopped cleanly")
}

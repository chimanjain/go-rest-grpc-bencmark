// Package server provides gRPC and REST server implementations for the benchmark.
package server

import (
	"context"
	"log/slog"
	"net"

	pb "github.com/chimanjain/go-rest-grpc-benchmark/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCServer wraps a gRPC server with its dependencies.
type GRPCServer struct {
	srv  *grpc.Server
	addr string
}

// benchmarkServiceServer is the private gRPC service implementation.
type benchmarkServiceServer struct {
	pb.UnimplementedBenchmarkServiceServer
}

// ProcessData implements the BenchmarkService.ProcessData RPC.
// It echoes the request ID and returns a processed result string.
func (s *benchmarkServiceServer) ProcessData(
	ctx context.Context,
	req *pb.ProcessRequest,
) (*pb.ProcessResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request must not be nil")
	}
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "request ID must not be empty")
	}

	return &pb.ProcessResponse{
		Id:     req.Id,
		Result: "Processed: " + req.Payload,
	}, nil
}

// NewGRPCServer creates and starts a new gRPC server bound to addr.
// The caller is responsible for calling Stop when finished.
func NewGRPCServer(addr string) (*GRPCServer, error) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	srv := grpc.NewServer()
	pb.RegisterBenchmarkServiceServer(srv, &benchmarkServiceServer{})

	go func() {
		slog.Info("gRPC server started", "addr", addr)
		if err := srv.Serve(lis); err != nil {
			slog.Error("gRPC server stopped", "error", err)
		}
	}()

	return &GRPCServer{srv: srv, addr: addr}, nil
}

// Stop performs a graceful shutdown of the gRPC server.
func (g *GRPCServer) Stop() {
	slog.Info("stopping gRPC server", "addr", g.addr)
	g.srv.GracefulStop()
}

// Addr returns the address the server is listening on.
func (g *GRPCServer) Addr() string { return g.addr }

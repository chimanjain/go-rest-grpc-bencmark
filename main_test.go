// Package main contains benchmark tests that compare the performance of a
// Gin-based REST API against a gRPC service under three load profiles:
//   - Small payload  (sequential)
//   - Large payload  (sequential, 100 KB)
//   - Parallel load  (goroutine-per-CPU)
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	pb "github.com/chimanjain/go-rest-grpc-benchmark/proto"
	"github.com/chimanjain/go-rest-grpc-benchmark/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	grpcAddr    = ":50051"
	restAddr    = ":8080"
	restBaseURL = "http://127.0.0.1:8080/api/v1/process"

	// smallPayload is a representative short message.
	smallPayload = "benchmarking_payload_data_string_example"
)

// largePayload is a 100 KB string used to stress-test serialisation performance.
var largePayload = strings.Repeat("x", 100*1024)

// suite holds shared test infrastructure initialised once in TestMain.
var suite struct {
	grpcSrv    *server.GRPCServer
	restSrv    *server.RESTServer
	grpcConn   *grpc.ClientConn
	grpcClient pb.BenchmarkServiceClient
	httpClient *http.Client
}

// TestMain sets up shared servers and clients before any benchmark runs and
// tears them down cleanly afterwards.
func TestMain(m *testing.M) {
	var err error

	// ── Start servers ──────────────────────────────────────────────────────────
	suite.grpcSrv, err = server.NewGRPCServer(grpcAddr)
	if err != nil {
		slog.Error("failed to start gRPC server", "error", err)
		os.Exit(1)
	}

	suite.restSrv, err = server.NewRESTServer(restAddr)
	if err != nil {
		slog.Error("failed to start REST server", "error", err)
		os.Exit(1)
	}

	// Allow both servers a moment to be ready before accepting connections.
	time.Sleep(200 * time.Millisecond)

	// ── gRPC client ────────────────────────────────────────────────────────────
	suite.grpcConn, err = grpc.NewClient(
		"127.0.0.1"+grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		slog.Error("failed to create gRPC client", "error", err)
		os.Exit(1)
	}
	suite.grpcClient = pb.NewBenchmarkServiceClient(suite.grpcConn)

	// ── HTTP client (connection-pooled) ────────────────────────────────────────
	suite.httpClient = server.NewRESTClient(1000)

	// ── Run ────────────────────────────────────────────────────────────────────
	code := m.Run()

	// ── Teardown ───────────────────────────────────────────────────────────────
	suite.grpcConn.Close()
	suite.grpcSrv.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	suite.restSrv.Shutdown(ctx) //nolint:errcheck

	os.Exit(code)
}

// TestCorrectness verifies that both servers are operational and returning valid results.
func TestCorrectness(t *testing.T) {
	t.Run("REST", func(t *testing.T) {
		body := restBody(t, "test-rest", "hello rest")
		resp, err := suite.httpClient.Post(restBaseURL, "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("REST request failed: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("gRPC", func(t *testing.T) {
		req := &pb.ProcessRequest{Id: "test-grpc", Payload: "hello grpc"}
		resp, err := suite.grpcClient.ProcessData(context.Background(), req)
		if err != nil {
			t.Fatalf("gRPC request failed: %v", err)
		}
		if resp.Result != "Processed: hello grpc" {
			t.Errorf("unexpected gRPC result: %s", resp.Result)
		}
	})
}

// ─── Helpers ───────────────────────────────────────────────────────────────────

// restBody serialises a ProcessRequest into a JSON byte slice.
func restBody(t testing.TB, id, payload string) []byte {
	t.Helper()
	data := map[string]string{"id": id, "payload": payload}
	b, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	return b
}

// doRESTRequest executes a single POST and returns the response.
func doRESTRequest(b *testing.B, body []byte) {
	b.Helper()
	resp, err := suite.httpClient.Post(restBaseURL, "application/json", bytes.NewReader(body))
	if err != nil {
		b.Fatalf("REST request failed: %v", err)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
}

// doGRPCRequest executes a single ProcessData RPC.
func doGRPCRequest(b *testing.B, req *pb.ProcessRequest) {
	b.Helper()
	_, err := suite.grpcClient.ProcessData(context.Background(), req)
	if err != nil {
		b.Fatalf("gRPC request failed: %v", err)
	}
}

// ─── Benchmarks ────────────────────────────────────────────────────────────────

// BenchmarkREST_SmallPayload measures serial REST throughput for a small body.
func BenchmarkREST_SmallPayload(b *testing.B) {
	body := restBody(b, "bench-1", smallPayload)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doRESTRequest(b, body)
	}
}

// BenchmarkREST_LargePayload measures serial REST throughput with a 100 KB body.
func BenchmarkREST_LargePayload(b *testing.B) {
	body := restBody(b, "bench-1", largePayload)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doRESTRequest(b, body)
	}
}

// BenchmarkREST_Parallel measures REST throughput under GOMAXPROCS goroutines.
func BenchmarkREST_Parallel(b *testing.B) {
	body := restBody(b, "bench-1", smallPayload)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doRESTRequest(b, body)
		}
	})
}

// BenchmarkGRPC_SmallPayload measures serial gRPC throughput for a small body.
func BenchmarkGRPC_SmallPayload(b *testing.B) {
	req := &pb.ProcessRequest{Id: "bench-1", Payload: smallPayload}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doGRPCRequest(b, req)
	}
}

// BenchmarkGRPC_LargePayload measures serial gRPC throughput with a 100 KB body.
func BenchmarkGRPC_LargePayload(b *testing.B) {
	req := &pb.ProcessRequest{Id: "bench-1", Payload: largePayload}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doGRPCRequest(b, req)
	}
}

// BenchmarkGRPC_Parallel measures gRPC throughput under GOMAXPROCS goroutines.
func BenchmarkGRPC_Parallel(b *testing.B) {
	req := &pb.ProcessRequest{Id: "bench-1", Payload: smallPayload}
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doGRPCRequest(b, req)
		}
	})
}

# Go REST vs gRPC Benchmark

This project compares the performance of a REST API (using the Gin framework) against a gRPC service in Go. It measures latency, memory allocation, and CPU efficiency.

## Overview

This project is built using professional Go practices, including structured logging (`slog`), clean architecture, graceful shutdown, and robust benchmark configurations.

The project structure is as follows:
- `proto/`: Contains the Protocol Buffer definitions and generated Go code.
- `config/`: Centralized configuration management with sensible defaults and environment variable overrides.
- `server/`: Houses the encapsulated `GRPCServer` and Gin-based `RESTServer` implementations.
- `main.go`: The application entry point orchestrating server startup and graceful shutdown via OS signals.
- `main_test.go`: A comprehensive benchmark suite comparing both servers under various load profiles (sequential, parallel, large payloads).

## 🛠️ Build and Setup

### 1. Prerequisites
- [Go](https://golang.org/doc/install) (1.21+)
- [Protocol Buffers Compiler (protoc)](https://grpc.io/docs/protoc-installation/)
- Go plugins for protoc:
  ```bash
  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
  ```

### 2. Generate Code and Install Dependencies
```bash
# Generate gRPC Go code
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/benchmark.proto

# Install dependencies
go mod tidy
```

## 📖 Usage

### Running Functional Tests
To verify the system's correctness:
```bash
go test -v ./...
```
These tests cover unit tests for server handlers and end-to-end integration tests.

### Running Benchmarks
To compare the performance of REST and gRPC:
```bash
go test -bench . -benchmem
```
This outputs latency (ns/op), throughput (req/s), memory usage (B/op), and allocations per request.

### Running with Profiling
To generate CPU and memory profiles for deep analysis:
```bash
go test -bench . -benchmem -cpuprofile cpu.out -memprofile mem.out
go tool pprof -http=:8081 cpu.out
```

### Starting Servers Manually
```bash
go run main.go
```
- **REST**: `http://localhost:8080/api/v1/process`
- **gRPC**: `localhost:50051`

## Sample Benchmark Results

Tested on `windows/arm64` with varying loads:

### 1. Small Payload (Default)
| Metric | REST (Gin) | gRPC | Improvement |
| :--- | :--- | :--- | :--- |
| **Latency** | ~352,000 ns/op | ~89,000 ns/op | **3.9x faster** |
| **Throughput** | ~2,840 req/s | ~11,200 req/s | **3.9x higher** |
| **Memory** | ~25.9 KB/op | ~9.7 KB/op | **2.6x more efficient** |
| **Allocations** | 195 allocs/op | 155 allocs/op | **1.2x fewer** |

### 2. Large Payload (100KB)
| Metric | REST (Gin) | gRPC | Improvement |
| :--- | :--- | :--- | :--- |
| **Latency** | ~1,136,000 ns/op | ~372,000 ns/op | **~3.0x faster** |
| **Throughput** | ~880 req/s | ~2,680 req/s | **~3.0x higher** |
| **Memory** | ~806 KB/op | ~450 KB/op | **~1.8x more efficient** |
| **Allocations** | 241 allocs/op | 211 allocs/op | **1.1x fewer** |

### 3. Parallel Load (Concurrency)
| Metric | REST (Gin) | gRPC | Improvement |
| :--- | :--- | :--- | :--- |
| **Latency** | ~112,000 ns/op | ~22,000 ns/op | **~5.0x faster** |
| **Throughput** | ~8,900 req/s | ~44,400 req/s | **~5.0x higher** |
| **Allocations** | 187 allocs/op | 141 allocs/op | **1.3x fewer** |

## Analysis
The benchmark clearly shows that gRPC outperforms REST significantly, especially as payload size increases. This is due to:
1. **Protobuf Binary Format**: Much more compact than JSON.
2. **HTTP/2**: gRPC uses HTTP/2 multiplexing, whereas standard REST (HTTP/1.1 in this test) creates more overhead.
3. **Serialization**: Protobuf serialization/deserialization is highly optimized compared to JSON reflection.

## Profiling

To generate CPU and memory profiles:
```bash
go test -bench . -benchmem -cpuprofile cpu.out -memprofile mem.out
```
View them using `pprof`:
```bash
go tool pprof -http=:8081 cpu.out
```

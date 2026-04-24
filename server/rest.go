package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	pb "github.com/chimanjain/go-rest-grpc-benchmark/proto"
	"github.com/gin-gonic/gin"
)

// RESTServer wraps a standard HTTP server configured with Gin.
type RESTServer struct {
	srv  *http.Server
	addr string
}

// NewRESTServer creates and starts a new Gin-based HTTP server bound to addr.
// The caller is responsible for calling Shutdown when finished.
func NewRESTServer(addr string) (*RESTServer, error) {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Recovery()) // recover from panics without crashing the server

	registerRoutes(router)

	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("REST server started", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("REST server stopped unexpectedly", "error", err)
		}
	}()

	return &RESTServer{srv: srv, addr: addr}, nil
}

// Shutdown performs a graceful shutdown with the provided context deadline.
func (r *RESTServer) Shutdown(ctx context.Context) error {
	slog.Info("stopping REST server", "addr", r.addr)
	return r.srv.Shutdown(ctx)
}

// Addr returns the address the server is listening on.
func (r *RESTServer) Addr() string { return r.addr }

// registerRoutes attaches all API routes to the given router.
func registerRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	{
		api.POST("/process", handleProcess)
	}
}

// processRequest is the request body for the /process endpoint.
type processRequest struct {
	ID      string `json:"id"      binding:"required"`
	Payload string `json:"payload"`
}

// processResponse is the response body for the /process endpoint.
type processResponse struct {
	ID     string `json:"id"`
	Result string `json:"result"`
}

// handleProcess processes an incoming REST request.
// POST /api/v1/process
func handleProcess(c *gin.Context) {
	var req processRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp := processResponse{
		ID:     req.ID,
		Result: "Processed: " + req.Payload,
	}
	c.JSON(http.StatusOK, resp)
}

// NewRESTClient returns a pre-configured HTTP client suitable for benchmarking.
// It reuses connections via a shared transport to avoid per-request dial overhead.
func NewRESTClient(maxIdleConns int) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:          maxIdleConns,
			MaxIdleConnsPerHost:   maxIdleConns,
			MaxConnsPerHost:       maxIdleConns,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   5 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
}

// RESTPayload serialises a ProcessRequest into the REST wire format.
func RESTPayload(req *pb.ProcessRequest) processRequest {
	return processRequest{ID: req.Id, Payload: req.Payload}
}

package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	pb "github.com/chimanjain/go-rest-grpc-benchmark/proto"
	"github.com/gin-gonic/gin"
)

func TestHandleProcess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	registerRoutes(r)

	t.Run("ValidRequest", func(t *testing.T) {
		body := `{"id": "test-1", "payload": "hello"}`
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/process", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		expected := `{"id":"test-1","result":"Processed: hello"}`
		if strings.TrimSpace(w.Body.String()) != expected {
			t.Errorf("expected body %s, got %s", expected, w.Body.String())
		}
	})

	t.Run("InvalidRequest", func(t *testing.T) {
		body := `{"payload": "hello"}` // missing "id"
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/process", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

func TestGRPCProcessData(t *testing.T) {
	s := &benchmarkServiceServer{}
	ctx := context.Background()

	t.Run("ValidRequest", func(t *testing.T) {
		req := &pb.ProcessRequest{Id: "test-1", Payload: "hello"}
		resp, err := s.ProcessData(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Id != "test-1" {
			t.Errorf("expected id test-1, got %s", resp.Id)
		}
		if resp.Result != "Processed: hello" {
			t.Errorf("expected result 'Processed: hello', got %s", resp.Result)
		}
	})

	t.Run("NilRequest", func(t *testing.T) {
		_, err := s.ProcessData(ctx, nil)
		if err == nil {
			t.Fatal("expected error for nil request")
		}
	})

	t.Run("EmptyID", func(t *testing.T) {
		req := &pb.ProcessRequest{Id: "", Payload: "hello"}
		_, err := s.ProcessData(ctx, req)
		if err == nil {
			t.Fatal("expected error for empty ID")
		}
	})
}

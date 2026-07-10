package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	runtimeadapter "capcom/internal/adapters/runtime"
	"capcom/internal/domain"
	"capcom/internal/services"
)

func TestHealthz(t *testing.T) {
	router := NewRouter(RouterConfig{Version: "test"}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if contentType := rec.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("content type = %q, want application/json", contentType)
	}

	var got healthResponse
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	want := healthResponse{
		Status:  "ok",
		Service: "capcom",
		Version: "test",
	}
	if got != want {
		t.Fatalf("response = %#v, want %#v", got, want)
	}
}

func TestNotFound(t *testing.T) {
	router := NewRouter(RouterConfig{Version: "test"}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestCreateRuntimeConnection(t *testing.T) {
	router := NewRouter(RouterConfig{
		Version:            "test",
		RuntimeConnections: fakeRuntimeConnectionService{},
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	body := bytes.NewBufferString(`{
		"name":"local-gantry",
		"runtime_type":"gantry",
		"mode":"read_only",
		"endpoint":"http://127.0.0.1:3000",
		"actor":"test",
		"reason":"setup"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/runtime-connections", body)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var got runtimeConnectionResponse
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.ID != "runtime-1" {
		t.Fatalf("id = %q, want runtime-1", got.ID)
	}
}

func TestRuntimeConnectionEndpointsRequireDatabase(t *testing.T) {
	router := NewRouter(RouterConfig{Version: "test"}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	req := httptest.NewRequest(http.MethodGet, "/v1/runtime-connections", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
}

func TestListRuntimeConnections(t *testing.T) {
	router := NewRouter(RouterConfig{
		Version:            "test",
		RuntimeConnections: fakeRuntimeConnectionService{},
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	req := httptest.NewRequest(http.MethodGet, "/v1/runtime-connections", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var got []runtimeConnectionResponse
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("response length = %d, want 1", len(got))
	}
}

func TestGetRuntimeConnection(t *testing.T) {
	router := NewRouter(RouterConfig{
		Version:            "test",
		RuntimeConnections: fakeRuntimeConnectionService{},
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	req := httptest.NewRequest(http.MethodGet, "/v1/runtime-connections/runtime-1", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestTestRuntimeConnection(t *testing.T) {
	router := NewRouter(RouterConfig{
		Version:            "test",
		RuntimeConnections: fakeRuntimeConnectionService{},
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	req := httptest.NewRequest(http.MethodPost, "/v1/runtime-connections/runtime-1/test", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var got runtimeConnectionTestResponse
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Status != string(domain.RuntimeStatusActive) {
		t.Fatalf("status = %q, want active", got.Status)
	}
}

type fakeRuntimeConnectionService struct{}

func (fakeRuntimeConnectionService) Create(_ context.Context, input services.CreateRuntimeConnectionInput) (domain.RuntimeConnection, error) {
	now := time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC)
	return domain.RuntimeConnection{
		ID:        "runtime-1",
		Name:      input.Name,
		Kind:      input.Kind,
		Mode:      input.Mode,
		Status:    domain.RuntimeStatusPending,
		BaseURL:   input.Endpoint,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (fakeRuntimeConnectionService) Get(context.Context, string) (domain.RuntimeConnection, error) {
	now := time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC)
	return domain.RuntimeConnection{
		ID:        "runtime-1",
		Name:      "local-gantry",
		Kind:      domain.RuntimeKindGantry,
		Mode:      domain.RuntimeModeReadOnly,
		Status:    domain.RuntimeStatusPending,
		BaseURL:   "http://127.0.0.1:3000",
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (fakeRuntimeConnectionService) List(context.Context) ([]domain.RuntimeConnection, error) {
	conn, err := fakeRuntimeConnectionService{}.Get(context.Background(), "runtime-1")
	if err != nil {
		return nil, err
	}
	return []domain.RuntimeConnection{conn}, nil
}

func (fakeRuntimeConnectionService) Test(context.Context, string) (*runtimeadapter.CheckResult, error) {
	return &runtimeadapter.CheckResult{
		Status:  domain.RuntimeStatusActive,
		Message: "ok",
		Capabilities: runtimeadapter.Capabilities{
			ReadAgents:      true,
			ReadAgentAccess: true,
		},
	}, nil
}

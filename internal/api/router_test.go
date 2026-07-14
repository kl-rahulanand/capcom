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

func TestConsoleIsServedWithoutAdminToken(t *testing.T) {
	router := NewRouter(RouterConfig{Version: "test", AdminToken: "test-admin-token"}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("Capcom Console")) {
		t.Fatalf("response did not contain console document")
	}
}

func TestNotFound(t *testing.T) {
	router := NewRouter(RouterConfig{Version: "test", AdminToken: "test-admin-token"}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	req := authenticatedRequest(http.MethodGet, "/missing", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestCreateRuntimeConnection(t *testing.T) {
	router := NewRouter(RouterConfig{
		Version:            "test",
		AdminToken:         "test-admin-token",
		RuntimeConnections: fakeRuntimeConnectionService{},
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	body := bytes.NewBufferString(`{
		"name":"local-gantry",
		"runtime_type":"gantry",
		"mode":"read_only",
		"endpoint":"http://127.0.0.1:3000",
		"auth_ref":"gantry-key",
		"actor":"test",
		"reason":"setup"
	}`)
	req := authenticatedRequest(http.MethodPost, "/v1/runtime-connections", body)
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

func TestCreateSecretDoesNotReturnValue(t *testing.T) {
	router := NewRouter(RouterConfig{Version: "test", AdminToken: "test-admin-token", Secrets: fakeSecretService{}}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	req := authenticatedRequest(http.MethodPost, "/v1/secrets", bytes.NewBufferString(`{
		"name":"gantry-key","value":"top-secret","actor":"tester","reason":"setup"
	}`))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body: %s", rec.Code, rec.Body.String())
	}
	if bytes.Contains(rec.Body.Bytes(), []byte("top-secret")) || bytes.Contains(rec.Body.Bytes(), []byte("value")) {
		t.Fatalf("response exposed secret material: %s", rec.Body.String())
	}
}

func TestRuntimeConnectionEndpointsRequireDatabase(t *testing.T) {
	router := NewRouter(RouterConfig{Version: "test", AdminToken: "test-admin-token"}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	req := authenticatedRequest(http.MethodGet, "/v1/runtime-connections", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
}

func TestListRuntimeConnections(t *testing.T) {
	router := NewRouter(RouterConfig{
		Version:            "test",
		AdminToken:         "test-admin-token",
		RuntimeConnections: fakeRuntimeConnectionService{},
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	req := authenticatedRequest(http.MethodGet, "/v1/runtime-connections", nil)
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
		AdminToken:         "test-admin-token",
		RuntimeConnections: fakeRuntimeConnectionService{},
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	req := authenticatedRequest(http.MethodGet, "/v1/runtime-connections/runtime-1", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestTestRuntimeConnection(t *testing.T) {
	router := NewRouter(RouterConfig{
		Version:            "test",
		AdminToken:         "test-admin-token",
		RuntimeConnections: fakeRuntimeConnectionService{},
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	req := authenticatedRequest(http.MethodPost, "/v1/runtime-connections/runtime-1/test", nil)
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

func TestListRuntimeAgents(t *testing.T) {
	router := NewRouter(RouterConfig{
		Version: "test", AdminToken: "test-admin-token",
		RuntimeConnections: fakeRuntimeConnectionService{},
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	req := authenticatedRequest(http.MethodGet, "/v1/runtime-connections/runtime-1/agents", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var got []runtimeAgentResponse
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(got) != 1 || got[0].RuntimeAgentID != "agent-1" {
		t.Fatalf("response = %#v", got)
	}
}

func TestGetRuntimeAgentAccess(t *testing.T) {
	router := NewRouter(RouterConfig{
		Version: "test", AdminToken: "test-admin-token",
		RuntimeConnections: fakeRuntimeConnectionService{},
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	req := authenticatedRequest(http.MethodGet, "/v1/runtime-connections/runtime-1/agents/agent-1/access", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var got runtimeAgentAccessResponse
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.AgentID != "agent-1" || len(got.Selections) != 1 {
		t.Fatalf("response = %#v", got)
	}
}

func TestListRuntimeAgentSkills(t *testing.T) {
	router := NewRouter(RouterConfig{
		Version: "test", AdminToken: "test-admin-token",
		RuntimeConnections: fakeRuntimeConnectionService{},
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	req := authenticatedRequest(http.MethodGet, "/v1/runtime-connections/runtime-1/agents/agent-1/skills", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var got []runtimeAgentSkillResponse
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(got) != 1 || got[0].RuntimeSkillID != "skill-1" {
		t.Fatalf("response = %#v", got)
	}
}

func TestProtectedEndpointRequiresAdminToken(t *testing.T) {
	router := NewRouter(RouterConfig{Version: "test", AdminToken: "test-admin-token"}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	req := httptest.NewRequest(http.MethodGet, "/v1/runtime-connections", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func authenticatedRequest(method, target string, body io.Reader) *http.Request {
	req := httptest.NewRequest(method, target, body)
	req.Header.Set("Authorization", "Bearer test-admin-token")
	return req
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
		AuthRef:   input.AuthRef,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

type fakeSecretService struct{}

func (fakeSecretService) Create(_ context.Context, input services.StoreSecretInput) (domain.Secret, error) {
	return testSecret(input.Name), nil
}

func (fakeSecretService) Rotate(_ context.Context, input services.StoreSecretInput) (domain.Secret, error) {
	return testSecret(input.Name), nil
}

func testSecret(name string) domain.Secret {
	now := time.Date(2026, 7, 14, 10, 0, 0, 0, time.UTC)
	return domain.Secret{ID: "secret-1", Name: name, CreatedAt: now, UpdatedAt: now}
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

func (fakeRuntimeConnectionService) ListAgents(context.Context, string) ([]domain.AgentSnapshot, error) {
	return []domain.AgentSnapshot{{
		RuntimeAgentID: "agent-1", Kind: domain.AgentKindRegistered, Name: "Research agent", Status: domain.AgentStatusEnabled,
		ObservedAt: time.Date(2026, 7, 14, 10, 0, 0, 0, time.UTC),
	}}, nil
}

func (fakeRuntimeConnectionService) ListAgentSkills(context.Context, string, string) ([]domain.AgentSkillSnapshot, error) {
	return []domain.AgentSkillSnapshot{{
		RuntimeSkillID: "skill-1", Name: "Research", Status: "active",
		ObservedAt: time.Date(2026, 7, 14, 10, 0, 0, 0, time.UTC),
	}}, nil
}

func (fakeRuntimeConnectionService) GetAgentAccess(context.Context, string, string) (*domain.AccessDocument, error) {
	return &domain.AccessDocument{
		AgentID: "agent-1", Source: "gantry",
		ObservedAt: time.Date(2026, 7, 14, 10, 0, 0, 0, time.UTC),
		Selections: []domain.AccessSelection{{Kind: "capability", ID: "web", Name: "web", Allowed: true}},
	}, nil
}

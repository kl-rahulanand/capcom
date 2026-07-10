package gantry

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	runtimeadapter "capcom/internal/adapters/runtime"
	"capcom/internal/domain"
)

func TestClientImplementsRuntimeAdapter(t *testing.T) {
	var _ runtimeadapter.Adapter = Client{}
}

func TestCheck(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/health" {
			t.Fatalf("path = %q, want /v1/health", r.URL.Path)
		}
		writeTestJSON(t, w, map[string]any{"status": "ok"})
	}))
	defer server.Close()

	got, err := NewClient(server.Client()).Check(context.Background(), domain.RuntimeConnection{
		BaseURL: server.URL,
		Mode:    domain.RuntimeModeReadOnly,
	})
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if got.Status != domain.RuntimeStatusActive {
		t.Fatalf("status = %q, want active", got.Status)
	}
	if !got.Capabilities.ReadAgents || !got.Capabilities.ReadAgentAccess {
		t.Fatalf("read capabilities were not set: %#v", got.Capabilities)
	}
	if got.Capabilities.ReplaceAgentAccess {
		t.Fatal("read-only connection should not support replace access")
	}
}

func TestListAgents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/agents" {
			t.Fatalf("path = %q, want /v1/agents", r.URL.Path)
		}
		writeTestJSON(t, w, []map[string]any{
			{"id": "agent:main", "name": "main", "status": "active"},
		})
	}))
	defer server.Close()

	got, err := NewClient(server.Client()).ListAgents(context.Background(), domain.RuntimeConnection{BaseURL: server.URL})
	if err != nil {
		t.Fatalf("ListAgents returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("agent count = %d, want 1", len(got))
	}
	if got[0].RuntimeAgentID != "agent:main" {
		t.Fatalf("agent id = %q", got[0].RuntimeAgentID)
	}
	if got[0].Status != domain.AgentStatusEnabled {
		t.Fatalf("agent status = %q", got[0].Status)
	}
}

func TestGetAgentAccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/agents/agent:main/access" {
			t.Fatalf("path = %q, want /v1/agents/agent:main/access", r.URL.Path)
		}
		writeTestJSON(t, w, map[string]any{
			"agentId": "agent:main",
			"selections": []map[string]any{
				{"id": "browser.use", "version": "builtin"},
			},
		})
	}))
	defer server.Close()

	got, err := NewClient(server.Client()).GetAgentAccess(context.Background(), domain.RuntimeConnection{BaseURL: server.URL}, "agent:main")
	if err != nil {
		t.Fatalf("GetAgentAccess returned error: %v", err)
	}
	if got.AgentID != "agent:main" {
		t.Fatalf("agent id = %q", got.AgentID)
	}
	if len(got.Selections) != 1 || got.Selections[0].ID != "browser.use" {
		t.Fatalf("selections = %#v", got.Selections)
	}
	if got.Raw["agentId"] != "agent:main" {
		t.Fatalf("raw payload not preserved: %#v", got.Raw)
	}
}

func writeTestJSON(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Fatalf("write json: %v", err)
	}
}

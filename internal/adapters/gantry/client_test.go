package gantry

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

	got, err := NewClient(server.Client(), staticCredentialResolver{"gantry-key": "test-token"}).Check(context.Background(), domain.RuntimeConnection{
		BaseURL: server.URL,
		Mode:    domain.RuntimeModeReadOnly,
		AuthRef: "gantry-key",
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

	got, err := NewClient(server.Client(), staticCredentialResolver{"gantry-key": "test-token"}).ListAgents(context.Background(), domain.RuntimeConnection{BaseURL: server.URL, AuthRef: "gantry-key"})
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
	if got[0].Kind != domain.AgentKindRegistered {
		t.Fatalf("agent kind = %q, want registered", got[0].Kind)
	}
}

func TestListAgentsAcceptsEnvelope(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeTestJSON(t, w, map[string]any{
			"agents": []map[string]any{{"id": "agent:main", "name": "main", "status": "active"}},
		})
	}))
	defer server.Close()

	got, err := NewClient(server.Client(), staticCredentialResolver{"gantry-key": "test-token"}).ListAgents(
		context.Background(), domain.RuntimeConnection{BaseURL: server.URL, AuthRef: "gantry-key"},
	)
	if err != nil {
		t.Fatalf("ListAgents returned error: %v", err)
	}
	if len(got) != 1 || got[0].RuntimeAgentID != "agent:main" {
		t.Fatalf("agents = %#v", got)
	}
}

func TestListAgentSkills(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/agents/agent:main_agent/skills":
			writeTestJSON(t, w, map[string]any{"bindings": []map[string]any{{
				"id": "binding-1", "agentId": "agent:main_agent", "skillId": "skill:research",
				"status": "active", "configVersionId": "config-2",
			}}})
		case "/v1/skills":
			writeTestJSON(t, w, map[string]any{"skills": []map[string]any{{
				"id": "skill:research", "name": "Research", "description": "Search trusted sources",
				"source": "bundled", "status": "installed", "toolIds": []string{"tool:web"},
				"workflowRefs": []string{"workflow:report"},
			}}})
		default:
			t.Fatalf("path = %q", r.URL.Path)
		}
	}))
	defer server.Close()

	got, err := NewClient(server.Client(), staticCredentialResolver{"gantry-key": "test-token"}).ListAgentSkills(
		context.Background(), domain.RuntimeConnection{BaseURL: server.URL, AuthRef: "gantry-key"}, "agent:main_agent",
	)
	if err != nil {
		t.Fatalf("ListAgentSkills returned error: %v", err)
	}
	if len(got) != 1 || got[0].Name != "Research" || got[0].Description != "Search trusted sources" || len(got[0].ToolIDs) != 1 {
		t.Fatalf("skills = %#v", got)
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

	got, err := NewClient(server.Client(), staticCredentialResolver{"gantry-key": "test-token"}).GetAgentAccess(context.Background(), domain.RuntimeConnection{BaseURL: server.URL, AuthRef: "gantry-key"}, "agent:main")
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

func TestCollectSnapshotReturnsCompleteNormalizedState(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/health":
			writeTestJSON(t, w, map[string]any{"status": "ok"})
		case "/v1/agents":
			writeTestJSON(t, w, []map[string]any{{"id": "agent:main_agent", "name": "gantry", "status": "active"}})
		case "/v1/agents/agent:main_agent/skills":
			writeTestJSON(t, w, map[string]any{"bindings": []map[string]any{{"skillId": "skill:research", "status": "active"}}})
		case "/v1/skills":
			writeTestJSON(t, w, map[string]any{"skills": []map[string]any{{"id": "skill:research", "name": "Research Brief"}}})
		case "/v1/agents/agent:main_agent/access":
			writeTestJSON(t, w, map[string]any{"agentId": "agent:main_agent", "selections": []map[string]any{{"id": "browser.use"}}})
		case "/v1/runs":
			writeTestJSON(t, w, map[string]any{"runs": []any{}})
		case "/v1/jobs":
			writeTestJSON(t, w, map[string]any{"jobs": []any{}})
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer server.Close()

	got, err := NewClient(server.Client(), staticCredentialResolver{"gantry-key": "test-token"}).CollectSnapshot(
		context.Background(), domain.RuntimeConnection{BaseURL: server.URL, AuthRef: "gantry-key"},
	)
	if err != nil {
		t.Fatalf("CollectSnapshot returned error: %v", err)
	}
	if len(got.Agents) != 1 || got.Agents[0].Agent.Kind != domain.AgentKindMain {
		t.Fatalf("agents = %#v", got.Agents)
	}
	if len(got.Agents[0].Skills) != 1 || got.Agents[0].Skills[0].Name != "Research Brief" {
		t.Fatalf("skills = %#v", got.Agents[0].Skills)
	}
	if len(got.Agents[0].Access.Selections) != 1 {
		t.Fatalf("access = %#v", got.Agents[0].Access)
	}
}

func TestNormalizeDelegatedExecutions(t *testing.T) {
	now := time.Now().UTC()
	events := []gantryRunEvent{
		{ID: "event-1", CreatedAt: now.Format(time.RFC3339Nano), Payload: map[string]any{"taskId": "task-1", "taskKind": "delegated_agent", "subagentType": "researcher", "description": "Investigate"}},
		{ID: "event-2", CreatedAt: now.Add(time.Second).Format(time.RFC3339Nano), Payload: map[string]any{"taskId": "task-1", "status": "completed", "summary": "Done"}},
		{ID: "event-3", CreatedAt: now.Format(time.RFC3339Nano), Payload: map[string]any{"taskId": "command-1", "taskKind": "async_command"}},
	}
	events[0].Metadata.RuntimeEventType = "task.started"
	events[1].Metadata.RuntimeEventType = "task.notification"
	got := normalizeDelegatedExecutions(gantryRun{RunID: "run-1", JobID: "job-1"}, "agent:main_agent", events, now)
	if len(got) != 1 || got[0].RuntimeExecutionID != "task-1" || got[0].RuntimeAgentID != "agent:main_agent" || got[0].Status != "completed" {
		t.Fatalf("executions = %#v", got)
	}
}

func TestClientSendsBearerCredential(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Fatalf("Authorization = %q", got)
		}
		writeTestJSON(t, w, map[string]any{"status": "ok"})
	}))
	defer server.Close()

	_, err := NewClient(server.Client(), staticCredentialResolver{"gantry-key": "test-token"}).Check(
		context.Background(),
		domain.RuntimeConnection{BaseURL: server.URL, AuthRef: "gantry-key", Mode: domain.RuntimeModeReadOnly},
	)
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
}

type staticCredentialResolver map[string]string

func (r staticCredentialResolver) Resolve(_ context.Context, ref string) (string, error) {
	return r[ref], nil
}

func writeTestJSON(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Fatalf("write json: %v", err)
	}
}

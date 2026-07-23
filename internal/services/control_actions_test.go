package services

import (
	"context"
	"database/sql"
	"testing"

	runtimeadapter "capcom/internal/adapters/runtime"
	"capcom/internal/domain"
)

type statusRuntimeRepository struct {
	RuntimeConnectionRepository
	connection domain.RuntimeConnection
}

func (r statusRuntimeRepository) Get(context.Context, string) (domain.RuntimeConnection, error) {
	return r.connection, nil
}

type statusAgentRepository struct {
	RuntimeSyncRepository
	detail domain.PersistedAgentDetail
}

func (r statusAgentRepository) GetPersistedAgent(context.Context, string) (domain.PersistedAgentDetail, error) {
	return r.detail, nil
}

type statusActionRepository struct {
	actions map[string]domain.ControlAction
}

func (r *statusActionRepository) Create(_ context.Context, action domain.ControlAction) (domain.ControlAction, error) {
	action.ID = "action-1"
	r.actions[action.IdempotencyKey] = action
	return action, nil
}

func (r *statusActionRepository) FindByIdempotencyKey(_ context.Context, key string) (domain.ControlAction, error) {
	action, ok := r.actions[key]
	if !ok {
		return domain.ControlAction{}, sql.ErrNoRows
	}
	return action, nil
}

func (r *statusActionRepository) Update(_ context.Context, action domain.ControlAction, _, response map[string]any, _ string) (domain.ControlAction, error) {
	action.Result = response
	r.actions[action.IdempotencyKey] = action
	return action, nil
}

type statusAdapter struct {
	runtimeadapter.Adapter
	called bool
}

func (a *statusAdapter) Kind() domain.RuntimeKind { return domain.RuntimeKindGantry }

func (a *statusAdapter) Check(context.Context, domain.RuntimeConnection) (*runtimeadapter.CheckResult, error) {
	return &runtimeadapter.CheckResult{Capabilities: runtimeadapter.Capabilities{SetAgentStatus: true}}, nil
}

func (a *statusAdapter) SetAgentStatus(_ context.Context, _ domain.RuntimeConnection, runtimeAgentID string, status domain.AgentStatus) (*domain.AgentSnapshot, error) {
	a.called = true
	return &domain.AgentSnapshot{RuntimeAgentID: runtimeAgentID, Status: status}, nil
}

type statusSyncer struct{ called bool }

func (s *statusSyncer) Sync(context.Context, SyncRuntimeInput) (domain.RuntimeSyncRun, error) {
	s.called = true
	return domain.RuntimeSyncRun{Status: domain.SyncStatusSucceeded}, nil
}

func TestSetAgentStatusSucceedsAndRunsVerificationSync(t *testing.T) {
	actions := &statusActionRepository{actions: map[string]domain.ControlAction{}}
	adapter := &statusAdapter{}
	syncer := &statusSyncer{}
	service := NewControlActionService(
		statusRuntimeRepository{connection: domain.RuntimeConnection{ID: "runtime-1", Kind: domain.RuntimeKindGantry, Mode: domain.RuntimeModeControlEnabled}},
		statusAgentRepository{detail: domain.PersistedAgentDetail{Agent: domain.PersistedAgent{Agent: domain.Agent{ID: "agent-1", Status: domain.AgentStatusEnabled}, RuntimeConnectionID: "runtime-1", RuntimeAgentID: "gantry-agent-1"}}},
		actions,
		nil,
		syncer,
	).WithAdapter(adapter)

	action, err := service.SetAgentStatus(context.Background(), SetAgentStatusInput{
		AgentID: "agent-1", Status: domain.AgentStatusDisabled, Actor: "test", Reason: "maintenance", IdempotencyKey: "status-1",
	})
	if err != nil {
		t.Fatalf("SetAgentStatus returned error: %v", err)
	}
	if action.Status != domain.ControlActionSucceeded || action.Type != "disable_agent" {
		t.Fatalf("unexpected action: %#v", action)
	}
	if !adapter.called || !syncer.called {
		t.Fatalf("expected adapter and verification sync calls, adapter=%v sync=%v", adapter.called, syncer.called)
	}
}

func TestSetAgentStatusDryRunDoesNotMutateRuntime(t *testing.T) {
	actions := &statusActionRepository{actions: map[string]domain.ControlAction{}}
	adapter := &statusAdapter{}
	service := NewControlActionService(
		statusRuntimeRepository{connection: domain.RuntimeConnection{ID: "runtime-1", Kind: domain.RuntimeKindGantry, Mode: domain.RuntimeModeControlEnabled}},
		statusAgentRepository{detail: domain.PersistedAgentDetail{Agent: domain.PersistedAgent{Agent: domain.Agent{ID: "agent-1", Status: domain.AgentStatusEnabled}, RuntimeConnectionID: "runtime-1", RuntimeAgentID: "gantry-agent-1"}}},
		actions,
		nil,
		nil,
	).WithAdapter(adapter)

	action, err := service.SetAgentStatus(context.Background(), SetAgentStatusInput{
		AgentID: "agent-1", Status: domain.AgentStatusDisabled, Actor: "test", Reason: "preview", IdempotencyKey: "status-dry", DryRun: true,
	})
	if err != nil {
		t.Fatalf("SetAgentStatus returned error: %v", err)
	}
	if action.Status != domain.ControlActionSucceeded || adapter.called {
		t.Fatalf("expected successful dry run without adapter mutation: %#v", action)
	}
}

func TestSetAgentStatusRejectsReadOnlyRuntime(t *testing.T) {
	actions := &statusActionRepository{actions: map[string]domain.ControlAction{}}
	adapter := &statusAdapter{}
	service := NewControlActionService(
		statusRuntimeRepository{connection: domain.RuntimeConnection{ID: "runtime-1", Kind: domain.RuntimeKindGantry, Mode: domain.RuntimeModeReadOnly}},
		statusAgentRepository{detail: domain.PersistedAgentDetail{Agent: domain.PersistedAgent{Agent: domain.Agent{ID: "agent-1", Status: domain.AgentStatusEnabled}, RuntimeConnectionID: "runtime-1", RuntimeAgentID: "gantry-agent-1"}}},
		actions,
		nil,
		nil,
	).WithAdapter(adapter)

	action, err := service.SetAgentStatus(context.Background(), SetAgentStatusInput{
		AgentID: "agent-1", Status: domain.AgentStatusDisabled, Actor: "test", Reason: "maintenance", IdempotencyKey: "status-read-only",
	})
	if err == nil {
		t.Fatal("expected read-only runtime rejection")
	}
	if action.Status != domain.ControlActionRejected || adapter.called {
		t.Fatalf("expected rejected action without adapter mutation: %#v", action)
	}
}

package services

import (
	"context"
	"testing"

	runtimeadapter "capcom/internal/adapters/runtime"
	"capcom/internal/domain"
)

func TestRuntimeConnectionServiceCreateValidatesInput(t *testing.T) {
	service := NewRuntimeConnectionService(fakeRuntimeRepo{}, nil).WithCredentialResolver(fakeCredentialResolver{})

	_, err := service.Create(context.Background(), CreateRuntimeConnectionInput{
		Name:     "",
		Kind:     domain.RuntimeKindGantry,
		Mode:     domain.RuntimeModeReadOnly,
		Endpoint: "http://127.0.0.1:3000",
		AuthRef:  "gantry-key",
	})
	if err == nil {
		t.Fatal("Create returned nil error")
	}
}

func TestRuntimeConnectionServiceCreate(t *testing.T) {
	service := NewRuntimeConnectionService(fakeRuntimeRepo{}, fakeAuditRepo{}).WithCredentialResolver(fakeCredentialResolver{})

	conn, err := service.Create(context.Background(), CreateRuntimeConnectionInput{
		Name:     "local-gantry",
		Kind:     domain.RuntimeKindGantry,
		Mode:     domain.RuntimeModeReadOnly,
		Endpoint: "http://127.0.0.1:3000",
		AuthRef:  "gantry-key",
		Actor:    "test",
		Reason:   "integration setup",
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if conn.Status != domain.RuntimeStatusPending {
		t.Fatalf("status = %q, want %q", conn.Status, domain.RuntimeStatusPending)
	}
}

func TestRuntimeConnectionServiceTest(t *testing.T) {
	service := NewRuntimeConnectionService(fakeRuntimeRepo{}, nil).WithCredentialResolver(fakeCredentialResolver{}).WithAdapter(fakeAdapter{})

	got, err := service.Test(context.Background(), "runtime-1")
	if err != nil {
		t.Fatalf("Test returned error: %v", err)
	}
	if got.Status != domain.RuntimeStatusActive {
		t.Fatalf("status = %q, want active", got.Status)
	}
}

func TestRuntimeConnectionServiceReadsAgentsThroughAdapter(t *testing.T) {
	service := NewRuntimeConnectionService(fakeRuntimeRepo{}, nil).WithAdapter(fakeAdapter{})

	agents, err := service.ListAgents(context.Background(), "runtime-1")
	if err != nil {
		t.Fatalf("ListAgents returned error: %v", err)
	}
	if len(agents) != 1 || agents[0].RuntimeAgentID != "agent-1" {
		t.Fatalf("agents = %#v", agents)
	}

	access, err := service.GetAgentAccess(context.Background(), "runtime-1", "agent-1")
	if err != nil {
		t.Fatalf("GetAgentAccess returned error: %v", err)
	}
	if access.AgentID != "agent-1" {
		t.Fatalf("agent id = %q", access.AgentID)
	}

	skills, err := service.ListAgentSkills(context.Background(), "runtime-1", "agent-1")
	if err != nil {
		t.Fatalf("ListAgentSkills returned error: %v", err)
	}
	if len(skills) != 1 || skills[0].RuntimeSkillID != "skill-1" {
		t.Fatalf("skills = %#v", skills)
	}
}

type fakeRuntimeRepo struct{}

func (fakeRuntimeRepo) Create(_ context.Context, conn domain.RuntimeConnection) (domain.RuntimeConnection, error) {
	conn.ID = "runtime-1"
	return conn, nil
}

func (fakeRuntimeRepo) Get(_ context.Context, id string) (domain.RuntimeConnection, error) {
	return domain.RuntimeConnection{ID: id, Name: "runtime", Kind: domain.RuntimeKindGantry, AuthRef: "gantry-key"}, nil
}

func (fakeRuntimeRepo) List(context.Context) ([]domain.RuntimeConnection, error) {
	return []domain.RuntimeConnection{{ID: "runtime-1", Name: "runtime"}}, nil
}

type fakeAuditRepo struct{}

func (fakeAuditRepo) Create(_ context.Context, event domain.AuditEvent) (domain.AuditEvent, error) {
	event.ID = "audit-1"
	return event, nil
}

type fakeAdapter struct{}

type fakeCredentialResolver struct{}

func (fakeCredentialResolver) Resolve(context.Context, string) (string, error) {
	return "token", nil
}

func (fakeAdapter) Kind() domain.RuntimeKind {
	return domain.RuntimeKindGantry
}

func (fakeAdapter) Check(context.Context, domain.RuntimeConnection) (*runtimeadapter.CheckResult, error) {
	return &runtimeadapter.CheckResult{Status: domain.RuntimeStatusActive}, nil
}

func (fakeAdapter) ListAgents(context.Context, domain.RuntimeConnection) ([]domain.AgentSnapshot, error) {
	return []domain.AgentSnapshot{{RuntimeAgentID: "agent-1"}}, nil
}

func (fakeAdapter) ListAgentSkills(context.Context, domain.RuntimeConnection, string) ([]domain.AgentSkillSnapshot, error) {
	return []domain.AgentSkillSnapshot{{RuntimeSkillID: "skill-1"}}, nil
}

func (fakeAdapter) GetAgentAccess(context.Context, domain.RuntimeConnection, string) (*domain.AccessDocument, error) {
	return &domain.AccessDocument{AgentID: "agent-1"}, nil
}

func (fakeAdapter) ReplaceAgentAccess(context.Context, domain.RuntimeConnection, string, domain.AccessDocument) (*domain.AccessDocument, error) {
	return nil, nil
}

func (fakeAdapter) CollectSnapshot(context.Context, domain.RuntimeConnection) (*domain.RuntimeSnapshot, error) {
	return &domain.RuntimeSnapshot{}, nil
}

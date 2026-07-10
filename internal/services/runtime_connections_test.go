package services

import (
	"context"
	"testing"

	runtimeadapter "capcom/internal/adapters/runtime"
	"capcom/internal/domain"
)

func TestRuntimeConnectionServiceCreateValidatesInput(t *testing.T) {
	service := NewRuntimeConnectionService(fakeRuntimeRepo{}, nil)

	_, err := service.Create(context.Background(), CreateRuntimeConnectionInput{
		Name:     "",
		Kind:     domain.RuntimeKindGantry,
		Mode:     domain.RuntimeModeReadOnly,
		Endpoint: "http://127.0.0.1:3000",
	})
	if err == nil {
		t.Fatal("Create returned nil error")
	}
}

func TestRuntimeConnectionServiceCreate(t *testing.T) {
	service := NewRuntimeConnectionService(fakeRuntimeRepo{}, fakeAuditRepo{})

	conn, err := service.Create(context.Background(), CreateRuntimeConnectionInput{
		Name:     "local-gantry",
		Kind:     domain.RuntimeKindGantry,
		Mode:     domain.RuntimeModeReadOnly,
		Endpoint: "http://127.0.0.1:3000",
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
	service := NewRuntimeConnectionService(fakeRuntimeRepo{}, nil).WithAdapter(fakeAdapter{})

	got, err := service.Test(context.Background(), "runtime-1")
	if err != nil {
		t.Fatalf("Test returned error: %v", err)
	}
	if got.Status != domain.RuntimeStatusActive {
		t.Fatalf("status = %q, want active", got.Status)
	}
}

type fakeRuntimeRepo struct{}

func (fakeRuntimeRepo) Create(_ context.Context, conn domain.RuntimeConnection) (domain.RuntimeConnection, error) {
	conn.ID = "runtime-1"
	return conn, nil
}

func (fakeRuntimeRepo) Get(_ context.Context, id string) (domain.RuntimeConnection, error) {
	return domain.RuntimeConnection{ID: id, Name: "runtime", Kind: domain.RuntimeKindGantry}, nil
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

func (fakeAdapter) Kind() domain.RuntimeKind {
	return domain.RuntimeKindGantry
}

func (fakeAdapter) Check(context.Context, domain.RuntimeConnection) (*runtimeadapter.CheckResult, error) {
	return &runtimeadapter.CheckResult{Status: domain.RuntimeStatusActive}, nil
}

func (fakeAdapter) ListAgents(context.Context, domain.RuntimeConnection) ([]domain.AgentSnapshot, error) {
	return nil, nil
}

func (fakeAdapter) GetAgentAccess(context.Context, domain.RuntimeConnection, string) (*domain.AccessDocument, error) {
	return nil, nil
}

func (fakeAdapter) ReplaceAgentAccess(context.Context, domain.RuntimeConnection, string, domain.AccessDocument) (*domain.AccessDocument, error) {
	return nil, nil
}

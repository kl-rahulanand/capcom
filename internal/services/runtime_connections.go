package services

import (
	"context"
	"fmt"
	"strings"

	runtimeadapter "capcom/internal/adapters/runtime"
	"capcom/internal/domain"
)

type RuntimeConnectionRepository interface {
	Create(ctx context.Context, conn domain.RuntimeConnection) (domain.RuntimeConnection, error)
	Get(ctx context.Context, id string) (domain.RuntimeConnection, error)
	List(ctx context.Context) ([]domain.RuntimeConnection, error)
}

type AuditRepository interface {
	Create(ctx context.Context, event domain.AuditEvent) (domain.AuditEvent, error)
}

type RuntimeConnectionService struct {
	runtimes RuntimeConnectionRepository
	audit    AuditRepository
	secrets  runtimeadapter.CredentialResolver
	adapters map[domain.RuntimeKind]runtimeadapter.Adapter
}

type CreateRuntimeConnectionInput struct {
	Name        string
	Kind        domain.RuntimeKind
	Mode        domain.RuntimeMode
	Endpoint    string
	Actor       string
	Reason      string
	Description string
	AuthRef     string
}

func (s RuntimeConnectionService) WithCredentialResolver(resolver runtimeadapter.CredentialResolver) RuntimeConnectionService {
	s.secrets = resolver
	return s
}

func NewRuntimeConnectionService(runtimes RuntimeConnectionRepository, audit AuditRepository) RuntimeConnectionService {
	return RuntimeConnectionService{
		runtimes: runtimes,
		audit:    audit,
		adapters: map[domain.RuntimeKind]runtimeadapter.Adapter{},
	}
}

func (s RuntimeConnectionService) WithAdapter(adapter runtimeadapter.Adapter) RuntimeConnectionService {
	if adapter == nil {
		return s
	}
	if s.adapters == nil {
		s.adapters = map[domain.RuntimeKind]runtimeadapter.Adapter{}
	}
	s.adapters[adapter.Kind()] = adapter
	return s
}

func (s RuntimeConnectionService) Create(ctx context.Context, input CreateRuntimeConnectionInput) (domain.RuntimeConnection, error) {
	if s.runtimes == nil {
		return domain.RuntimeConnection{}, fmt.Errorf("runtime repository is required")
	}
	if err := validateCreateRuntimeConnection(input); err != nil {
		return domain.RuntimeConnection{}, err
	}
	if s.secrets == nil {
		return domain.RuntimeConnection{}, fmt.Errorf("credential resolver is required")
	}
	if _, err := s.secrets.Resolve(ctx, input.AuthRef); err != nil {
		return domain.RuntimeConnection{}, fmt.Errorf("resolve auth_ref: %w", err)
	}

	conn, err := s.runtimes.Create(ctx, domain.RuntimeConnection{
		Name:    strings.TrimSpace(input.Name),
		Kind:    input.Kind,
		Mode:    input.Mode,
		Status:  domain.RuntimeStatusPending,
		BaseURL: strings.TrimSpace(input.Endpoint),
		AuthRef: strings.TrimSpace(input.AuthRef),
		Metadata: map[string]any{
			"description": strings.TrimSpace(input.Description),
		},
	})
	if err != nil {
		return domain.RuntimeConnection{}, err
	}

	if s.audit != nil && strings.TrimSpace(input.Actor) != "" && strings.TrimSpace(input.Reason) != "" {
		if _, err := s.audit.Create(ctx, domain.AuditEvent{
			RuntimeConnectionID: conn.ID,
			Actor:               strings.TrimSpace(input.Actor),
			EventType:           "runtime_connection.created",
			TargetType:          "runtime_connection",
			TargetID:            conn.ID,
			Reason:              strings.TrimSpace(input.Reason),
			Result:              "succeeded",
			After: map[string]any{
				"id":           conn.ID,
				"name":         conn.Name,
				"runtime_type": string(conn.Kind),
				"mode":         string(conn.Mode),
				"status":       string(conn.Status),
				"endpoint":     conn.BaseURL,
				"auth_ref":     conn.AuthRef,
			},
		}); err != nil {
			return domain.RuntimeConnection{}, fmt.Errorf("audit runtime connection creation: %w", err)
		}
	}

	return conn, nil
}

func (s RuntimeConnectionService) Get(ctx context.Context, id string) (domain.RuntimeConnection, error) {
	if strings.TrimSpace(id) == "" {
		return domain.RuntimeConnection{}, fmt.Errorf("id is required")
	}
	return s.runtimes.Get(ctx, strings.TrimSpace(id))
}

func (s RuntimeConnectionService) List(ctx context.Context) ([]domain.RuntimeConnection, error) {
	return s.runtimes.List(ctx)
}

func (s RuntimeConnectionService) Test(ctx context.Context, id string) (*runtimeadapter.CheckResult, error) {
	conn, adapter, err := s.connectionAdapter(ctx, id)
	if err != nil {
		return nil, err
	}

	return adapter.Check(ctx, conn)
}

func (s RuntimeConnectionService) ListAgents(ctx context.Context, id string) ([]domain.AgentSnapshot, error) {
	conn, adapter, err := s.connectionAdapter(ctx, id)
	if err != nil {
		return nil, err
	}

	return adapter.ListAgents(ctx, conn)
}

func (s RuntimeConnectionService) GetAgentAccess(ctx context.Context, id string, runtimeAgentID string) (*domain.AccessDocument, error) {
	if strings.TrimSpace(runtimeAgentID) == "" {
		return nil, fmt.Errorf("runtime agent id is required")
	}
	conn, adapter, err := s.connectionAdapter(ctx, id)
	if err != nil {
		return nil, err
	}

	return adapter.GetAgentAccess(ctx, conn, strings.TrimSpace(runtimeAgentID))
}

func (s RuntimeConnectionService) ListAgentSkills(ctx context.Context, id string, runtimeAgentID string) ([]domain.AgentSkillSnapshot, error) {
	if strings.TrimSpace(runtimeAgentID) == "" {
		return nil, fmt.Errorf("runtime agent id is required")
	}
	conn, adapter, err := s.connectionAdapter(ctx, id)
	if err != nil {
		return nil, err
	}
	return adapter.ListAgentSkills(ctx, conn, strings.TrimSpace(runtimeAgentID))
}

func (s RuntimeConnectionService) connectionAdapter(ctx context.Context, id string) (domain.RuntimeConnection, runtimeadapter.Adapter, error) {
	conn, err := s.Get(ctx, id)
	if err != nil {
		return domain.RuntimeConnection{}, nil, err
	}
	adapter, ok := s.adapters[conn.Kind]
	if !ok {
		return domain.RuntimeConnection{}, nil, fmt.Errorf("runtime adapter %q is not registered", conn.Kind)
	}
	return conn, adapter, nil
}

func validateCreateRuntimeConnection(input CreateRuntimeConnectionInput) error {
	if strings.TrimSpace(input.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if input.Kind != domain.RuntimeKindGantry {
		return fmt.Errorf("runtime_type must be %q", domain.RuntimeKindGantry)
	}
	switch input.Mode {
	case domain.RuntimeModeReadOnly, domain.RuntimeModeControlEnabled:
	default:
		return fmt.Errorf("mode must be one of %q, %q", domain.RuntimeModeReadOnly, domain.RuntimeModeControlEnabled)
	}
	if strings.TrimSpace(input.Endpoint) == "" {
		return fmt.Errorf("endpoint is required")
	}
	if strings.TrimSpace(input.AuthRef) == "" {
		return fmt.Errorf("auth_ref is required")
	}
	if strings.TrimSpace(input.Actor) == "" {
		return fmt.Errorf("actor is required")
	}
	if strings.TrimSpace(input.Reason) == "" {
		return fmt.Errorf("reason is required")
	}
	return nil
}

package runtime

import (
	"context"

	"capcom/internal/domain"
)

type Adapter interface {
	Kind() domain.RuntimeKind
	Check(ctx context.Context, conn domain.RuntimeConnection) (*CheckResult, error)
	ListAgents(ctx context.Context, conn domain.RuntimeConnection) ([]domain.AgentSnapshot, error)
	GetAgentAccess(ctx context.Context, conn domain.RuntimeConnection, runtimeAgentID string) (*domain.AccessDocument, error)
	ReplaceAgentAccess(ctx context.Context, conn domain.RuntimeConnection, runtimeAgentID string, access domain.AccessDocument) (*domain.AccessDocument, error)
}

type CheckResult struct {
	Status       domain.RuntimeStatus
	Message      string
	Capabilities Capabilities
	Metadata     map[string]any
}

type Capabilities struct {
	ReadAgents         bool `json:"read_agents"`
	ReadAgentAccess    bool `json:"read_agent_access"`
	ReplaceAgentAccess bool `json:"replace_agent_access"`
}

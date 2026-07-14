package runtime

import (
	"context"

	"capcom/internal/domain"
)

type CredentialResolver interface {
	Resolve(ctx context.Context, ref string) (string, error)
}

type Adapter interface {
	Kind() domain.RuntimeKind
	Check(ctx context.Context, conn domain.RuntimeConnection) (*CheckResult, error)
	ListAgents(ctx context.Context, conn domain.RuntimeConnection) ([]domain.AgentSnapshot, error)
	ListAgentSkills(ctx context.Context, conn domain.RuntimeConnection, runtimeAgentID string) ([]domain.AgentSkillSnapshot, error)
	GetAgentAccess(ctx context.Context, conn domain.RuntimeConnection, runtimeAgentID string) (*domain.AccessDocument, error)
	ReplaceAgentAccess(ctx context.Context, conn domain.RuntimeConnection, runtimeAgentID string, access domain.AccessDocument) (*domain.AccessDocument, error)
	CollectSnapshot(ctx context.Context, conn domain.RuntimeConnection) (*domain.RuntimeSnapshot, error)
}

type CheckResult struct {
	Status       domain.RuntimeStatus
	Message      string
	Capabilities Capabilities
	Metadata     map[string]any
}

type Capabilities struct {
	ReadAgents             bool `json:"read_agents"`
	ReadAgentHierarchy     bool `json:"read_agent_hierarchy"`
	ReadAgentSkills        bool `json:"read_agent_skills"`
	ReadAgentAccess        bool `json:"read_agent_access"`
	ReplaceAgentAccess     bool `json:"replace_agent_access"`
	ReadSubagentExecutions bool `json:"read_subagent_executions"`
}

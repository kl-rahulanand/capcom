package domain

import "time"

type AgentStatus string
type AgentKind string

const (
	AgentStatusUnknown  AgentStatus = "unknown"
	AgentStatusEnabled  AgentStatus = "enabled"
	AgentStatusDisabled AgentStatus = "disabled"
	AgentStatusStale    AgentStatus = "stale"
)

const (
	AgentKindMain       AgentKind = "main"
	AgentKindRegistered AgentKind = "registered"
	AgentKindSubagent   AgentKind = "subagent"
)

type Agent struct {
	ID        string
	Name      string
	Status    AgentStatus
	Metadata  map[string]any
	CreatedAt time.Time
	UpdatedAt time.Time
}

type AgentSnapshot struct {
	RuntimeAgentID       string
	ParentRuntimeAgentID string
	Kind                 AgentKind
	Name                 string
	Status               AgentStatus
	Metadata             map[string]any
	ObservedAt           time.Time
}

type AgentSkillSnapshot struct {
	RuntimeSkillID string
	Name           string
	Description    string
	Source         string
	Status         string
	Version        string
	ToolIDs        []string
	WorkflowRefs   []string
	Metadata       map[string]any
	ObservedAt     time.Time
}

type AgentBinding struct {
	ID                  string
	AgentID             string
	RuntimeConnectionID string
	RuntimeAgentID      string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

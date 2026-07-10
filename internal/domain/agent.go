package domain

import "time"

type AgentStatus string

const (
	AgentStatusUnknown  AgentStatus = "unknown"
	AgentStatusEnabled  AgentStatus = "enabled"
	AgentStatusDisabled AgentStatus = "disabled"
	AgentStatusStale    AgentStatus = "stale"
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
	RuntimeAgentID string
	Name           string
	Status         AgentStatus
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

package domain

import "time"

type AuditEvent struct {
	ID                  string
	RuntimeConnectionID string
	AgentID             string
	ControlActionID     string
	Actor               string
	EventType           string
	TargetType          string
	TargetID            string
	Reason              string
	Before              map[string]any
	After               map[string]any
	Result              string
	Metadata            map[string]any
	CreatedAt           time.Time
}

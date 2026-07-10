package domain

import "time"

type ControlActionStatus string

const (
	ControlActionQueued    ControlActionStatus = "queued"
	ControlActionRunning   ControlActionStatus = "running"
	ControlActionSucceeded ControlActionStatus = "succeeded"
	ControlActionFailed    ControlActionStatus = "failed"
	ControlActionRejected  ControlActionStatus = "rejected"
)

type ControlAction struct {
	ID                  string
	RuntimeConnectionID string
	AgentID             string
	Type                string
	Status              ControlActionStatus
	Actor               string
	Reason              string
	IdempotencyKey      string
	Before              map[string]any
	After               map[string]any
	Result              map[string]any
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

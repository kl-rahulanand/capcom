package domain

import "time"

type RuntimeKind string

const (
	RuntimeKindGantry RuntimeKind = "gantry"
)

type RuntimeMode string

const (
	RuntimeModeReadOnly       RuntimeMode = "read_only"
	RuntimeModeControlEnabled RuntimeMode = "control_enabled"
)

type RuntimeStatus string

const (
	RuntimeStatusPending  RuntimeStatus = "pending"
	RuntimeStatusActive   RuntimeStatus = "active"
	RuntimeStatusDegraded RuntimeStatus = "degraded"
	RuntimeStatusDisabled RuntimeStatus = "disabled"
	RuntimeStatusFailed   RuntimeStatus = "failed"
)

type RuntimeConnection struct {
	ID           string
	Name         string
	Kind         RuntimeKind
	Mode         RuntimeMode
	Status       RuntimeStatus
	BaseURL      string
	Metadata     map[string]any
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastSyncedAt *time.Time
}

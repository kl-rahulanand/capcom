package domain

import "time"

type DriftSeverity string

const (
	DriftSeverityInfo     DriftSeverity = "info"
	DriftSeverityWarning  DriftSeverity = "warning"
	DriftSeverityCritical DriftSeverity = "critical"
)

type DriftStatus string

const (
	DriftStatusOpen     DriftStatus = "open"
	DriftStatusResolved DriftStatus = "resolved"
)

type DriftFinding struct {
	ID          string
	AgentID     string
	Kind        string
	Severity    DriftSeverity
	Status      DriftStatus
	Expected    map[string]any
	Actual      map[string]any
	FirstSeenAt time.Time
	LastSeenAt  time.Time
	ResolvedAt  *time.Time
}

package domain

import "time"

type AccessDocument struct {
	AgentID      string
	Selections   []AccessSelection
	Raw          map[string]any
	ObservedAt   time.Time
	DesiredAt    *time.Time
	Source       string
	SourceRef    string
	SourceSHA256 string
}

type AccessSelection struct {
	Kind       string
	ID         string
	Name       string
	Allowed    bool
	Attributes map[string]any
}

package domain

import "time"

type RuntimeDiagnosticSnapshot struct {
	CheckID    string
	Status     string
	Message    string
	ObservedAt time.Time
	Metadata   map[string]any
	Raw        map[string]any
}

type RuntimeInventorySnapshot struct {
	RuntimeItemID string
	Kind          string
	Name          string
	Status        string
	Provider      string
	Source        string
	ObservedAt    time.Time
	Metadata      map[string]any
	Raw           map[string]any
}

type RuntimeCapabilitySnapshot struct {
	RuntimeCapabilityID string
	Version             string
	Name                string
	Category            string
	Risk                string
	Can                 string
	Cannot              string
	Source              string
	ObservedAt          time.Time
	Metadata            map[string]any
	Raw                 map[string]any
}

type PersistedRuntimeDiagnostic struct {
	ID                  string
	RuntimeConnectionID string
	RuntimeDiagnosticSnapshot
}

type PersistedRuntimeInventory struct {
	ID                  string
	RuntimeConnectionID string
	RuntimeInventorySnapshot
}

type PersistedRuntimeCapability struct {
	ID                  string
	RuntimeConnectionID string
	RuntimeCapabilitySnapshot
}

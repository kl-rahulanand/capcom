package domain

import "time"

type SyncTrigger string
type SyncStatus string
type FreshnessStatus string

const (
	SyncTriggerManual     SyncTrigger = "manual"
	SyncTriggerScheduled  SyncTrigger = "scheduled"
	SyncTriggerPostAction SyncTrigger = "post_action"
)

const (
	SyncStatusRunning   SyncStatus = "running"
	SyncStatusSucceeded SyncStatus = "succeeded"
	SyncStatusFailed    SyncStatus = "failed"
	SyncStatusSkipped   SyncStatus = "skipped"
)

const (
	FreshnessLive   FreshnessStatus = "live"
	FreshnessCached FreshnessStatus = "cached"
	FreshnessStale  FreshnessStatus = "stale"
)

type RuntimeSnapshot struct {
	ObservedAt         time.Time
	Agents             []SnapshotAgent
	Diagnostics        []RuntimeDiagnosticSnapshot
	Inventory          []RuntimeInventorySnapshot
	CapabilityCatalog  []RuntimeCapabilitySnapshot
	AgentDelegations   []AgentDelegationSnapshot
	Executions         []RuntimeExecutionSnapshot
	SubagentExecutions []SubagentExecutionSnapshot
	Metadata           map[string]any
	Capabilities       map[string]bool
}

type SnapshotAgent struct {
	Agent  AgentSnapshot
	Skills []AgentSkillSnapshot
	Access AccessDocument
}

type RuntimeSyncRun struct {
	ID                  string
	RuntimeConnectionID string
	Trigger             SyncTrigger
	Status              SyncStatus
	StartedAt           time.Time
	FinishedAt          *time.Time
	DurationMS          int64
	AgentsSeen          int
	SkillsSeen          int
	BindingsSeen        int
	AccessDocumentsSeen int
	ExecutionsSeen      int
	DiagnosticsSeen     int
	InventorySeen       int
	CapabilitiesSeen    int
	DelegationsSeen     int
	ErrorCode           string
	ErrorMessage        string
	Metadata            map[string]any
}

// AgentDelegationSnapshot is a durable directed relationship advertised by a
// runtime. It is separate from parent-child hierarchy because one delegate may
// be callable by multiple orchestrators.
type AgentDelegationSnapshot struct {
	OrchestratorRuntimeAgentID string
	DelegateRuntimeAgentID     string
	DelegateRef                string
	ToolName                   string
	DisplayName                string
	Persona                    string
	Configured                 bool
	Resolved                   bool
	Revision                   int
	Status                     string
	ObservedAt                 time.Time
	Metadata                   map[string]any
	Raw                        map[string]any
}

type PersistedAgentDelegation struct {
	ID                  string
	RuntimeConnectionID string
	AgentDelegationSnapshot
}

type RuntimeExecutionSnapshot struct {
	RuntimeExecutionID       string
	ParentRuntimeExecutionID string
	RuntimeAgentID           string
	Kind                     string
	Status                   string
	StartedAt                *time.Time
	EndedAt                  *time.Time
	ObservedAt               time.Time
	Metadata                 map[string]any
	Raw                      map[string]any
}

type PersistedRuntimeExecution struct {
	ID                  string
	RuntimeConnectionID string
	RuntimeExecutionSnapshot
}

type PersistedAgent struct {
	Agent
	RuntimeConnectionID  string
	RuntimeAgentID       string
	ParentRuntimeAgentID string
	Kind                 AgentKind
	Freshness            FreshnessStatus
	ObservedAt           time.Time
	LastSuccessfulSyncAt *time.Time
	RuntimeStatus        RuntimeStatus
}

type PersistedAgentDetail struct {
	Agent  PersistedAgent
	Skills []AgentSkillSnapshot
	Access AccessDocument
}

// SubagentExecutionSnapshot is an ephemeral delegated execution observed in a
// runtime. It is deliberately separate from the durable agent inventory.
type SubagentExecutionSnapshot struct {
	RuntimeExecutionID string
	ParentRunID        string
	RuntimeAgentID     string
	SubagentType       string
	Status             string
	Description        string
	Summary            string
	StartedAt          *time.Time
	EndedAt            *time.Time
	ObservedAt         time.Time
	Metadata           map[string]any
	Raw                map[string]any
}

type PersistedSubagentExecution struct {
	ID                  string
	RuntimeConnectionID string
	SubagentExecutionSnapshot
}

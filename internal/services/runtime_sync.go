package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	runtimeadapter "capcom/internal/adapters/runtime"
	"capcom/internal/domain"
)

var ErrSyncConflict = errors.New("runtime sync already in progress")

type RuntimeSyncRepository interface {
	TryLock(ctx context.Context, runtimeID string) (release func(), acquired bool, err error)
	CreateRun(ctx context.Context, run domain.RuntimeSyncRun) (domain.RuntimeSyncRun, error)
	FailRun(ctx context.Context, run domain.RuntimeSyncRun, code, message string) (domain.RuntimeSyncRun, error)
	PersistSnapshot(ctx context.Context, run domain.RuntimeSyncRun, snapshot domain.RuntimeSnapshot, missingThreshold int) (domain.RuntimeSyncRun, error)
	ListRuns(ctx context.Context, runtimeID string, limit int) ([]domain.RuntimeSyncRun, error)
	GetRun(ctx context.Context, runtimeID, runID string) (domain.RuntimeSyncRun, error)
	ListPersistedAgents(ctx context.Context, runtimeID string) ([]domain.PersistedAgent, error)
	GetPersistedAgent(ctx context.Context, agentID string) (domain.PersistedAgentDetail, error)
	ListSubagentExecutions(ctx context.Context, runtimeID, agentID string) ([]domain.PersistedSubagentExecution, error)
}

type RuntimeSyncService struct {
	runtimes         RuntimeConnectionRepository
	store            RuntimeSyncRepository
	audit            AuditRepository
	adapters         map[domain.RuntimeKind]runtimeadapter.Adapter
	missingThreshold int
}

type SyncRuntimeInput struct {
	RuntimeConnectionID string
	Trigger             domain.SyncTrigger
	Actor               string
	Reason              string
}

func NewRuntimeSyncService(runtimes RuntimeConnectionRepository, store RuntimeSyncRepository, audit AuditRepository, missingThreshold int) RuntimeSyncService {
	if missingThreshold <= 0 {
		missingThreshold = 3
	}
	return RuntimeSyncService{runtimes: runtimes, store: store, audit: audit, adapters: map[domain.RuntimeKind]runtimeadapter.Adapter{}, missingThreshold: missingThreshold}
}

func (s RuntimeSyncService) WithAdapter(adapter runtimeadapter.Adapter) RuntimeSyncService {
	if adapter != nil {
		s.adapters[adapter.Kind()] = adapter
	}
	return s
}

func (s RuntimeSyncService) Sync(ctx context.Context, input SyncRuntimeInput) (domain.RuntimeSyncRun, error) {
	if strings.TrimSpace(input.RuntimeConnectionID) == "" {
		return domain.RuntimeSyncRun{}, fmt.Errorf("runtime connection id is required")
	}
	if input.Trigger == "" {
		input.Trigger = domain.SyncTriggerManual
	}
	if input.Trigger == domain.SyncTriggerManual && (strings.TrimSpace(input.Actor) == "" || strings.TrimSpace(input.Reason) == "") {
		return domain.RuntimeSyncRun{}, fmt.Errorf("actor and reason are required")
	}
	conn, err := s.runtimes.Get(ctx, input.RuntimeConnectionID)
	if err != nil {
		return domain.RuntimeSyncRun{}, err
	}
	adapter, ok := s.adapters[conn.Kind]
	if !ok {
		return domain.RuntimeSyncRun{}, fmt.Errorf("runtime adapter %q is not registered", conn.Kind)
	}
	release, acquired, err := s.store.TryLock(ctx, conn.ID)
	if err != nil {
		return domain.RuntimeSyncRun{}, err
	}
	if !acquired {
		return domain.RuntimeSyncRun{}, ErrSyncConflict
	}
	defer release()

	run, err := s.store.CreateRun(ctx, domain.RuntimeSyncRun{RuntimeConnectionID: conn.ID, Trigger: input.Trigger})
	if err != nil {
		return domain.RuntimeSyncRun{}, err
	}
	snapshot, collectErr := adapter.CollectSnapshot(ctx, conn)
	if collectErr != nil {
		failed, failErr := s.store.FailRun(ctx, run, "adapter_error", sanitizeRuntimeError(collectErr))
		s.auditSync(ctx, failed, input, "runtime.sync_failed", "failed")
		if failErr != nil {
			return failed, errors.Join(collectErr, failErr)
		}
		return failed, fmt.Errorf("collect runtime snapshot: %w", collectErr)
	}
	if err := validateRuntimeSnapshot(*snapshot); err != nil {
		failed, failErr := s.store.FailRun(ctx, run, "invalid_snapshot", err.Error())
		s.auditSync(ctx, failed, input, "runtime.sync_failed", "failed")
		if failErr != nil {
			return failed, errors.Join(err, failErr)
		}
		return failed, err
	}
	completed, err := s.store.PersistSnapshot(ctx, run, *snapshot, s.missingThreshold)
	if err != nil {
		failed, failErr := s.store.FailRun(ctx, run, "persistence_error", sanitizeRuntimeError(err))
		s.auditSync(ctx, failed, input, "runtime.sync_failed", "failed")
		if failErr != nil {
			return failed, errors.Join(err, failErr)
		}
		return failed, err
	}
	s.auditSync(ctx, completed, input, "runtime.sync_succeeded", "succeeded")
	return completed, nil
}

func (s RuntimeSyncService) ListRuns(ctx context.Context, runtimeID string, limit int) ([]domain.RuntimeSyncRun, error) {
	return s.store.ListRuns(ctx, runtimeID, limit)
}

func (s RuntimeSyncService) GetRun(ctx context.Context, runtimeID, runID string) (domain.RuntimeSyncRun, error) {
	return s.store.GetRun(ctx, runtimeID, runID)
}

func (s RuntimeSyncService) ListAgents(ctx context.Context, runtimeID string) ([]domain.PersistedAgent, error) {
	return s.store.ListPersistedAgents(ctx, runtimeID)
}

func (s RuntimeSyncService) GetAgent(ctx context.Context, agentID string) (domain.PersistedAgentDetail, error) {
	if strings.TrimSpace(agentID) == "" {
		return domain.PersistedAgentDetail{}, fmt.Errorf("agent id is required")
	}
	return s.store.GetPersistedAgent(ctx, agentID)
}

func (s RuntimeSyncService) ListSubagentExecutions(ctx context.Context, runtimeID, agentID string) ([]domain.PersistedSubagentExecution, error) {
	return s.store.ListSubagentExecutions(ctx, runtimeID, agentID)
}

func validateRuntimeSnapshot(snapshot domain.RuntimeSnapshot) error {
	seen := map[string]struct{}{}
	for _, item := range snapshot.Agents {
		id := strings.TrimSpace(item.Agent.RuntimeAgentID)
		if id == "" {
			return fmt.Errorf("snapshot contains an agent without runtime id")
		}
		if _, ok := seen[id]; ok {
			return fmt.Errorf("snapshot contains duplicate runtime agent id %q", id)
		}
		seen[id] = struct{}{}
		if item.Access.AgentID != "" && item.Access.AgentID != id {
			return fmt.Errorf("access document agent id %q does not match %q", item.Access.AgentID, id)
		}
		skills := map[string]struct{}{}
		for _, skill := range item.Skills {
			if strings.TrimSpace(skill.RuntimeSkillID) == "" {
				return fmt.Errorf("agent %q has skill without runtime id", id)
			}
			if _, ok := skills[skill.RuntimeSkillID]; ok {
				return fmt.Errorf("agent %q has duplicate skill %q", id, skill.RuntimeSkillID)
			}
			skills[skill.RuntimeSkillID] = struct{}{}
		}
	}
	executions := map[string]struct{}{}
	for _, execution := range snapshot.SubagentExecutions {
		id := strings.TrimSpace(execution.RuntimeExecutionID)
		if id == "" {
			return fmt.Errorf("snapshot contains a subagent execution without runtime id")
		}
		if _, ok := executions[id]; ok {
			return fmt.Errorf("snapshot contains duplicate subagent execution id %q", id)
		}
		executions[id] = struct{}{}
	}
	return nil
}

func (s RuntimeSyncService) auditSync(ctx context.Context, run domain.RuntimeSyncRun, input SyncRuntimeInput, eventType, result string) {
	if s.audit == nil {
		return
	}
	actor := strings.TrimSpace(input.Actor)
	if actor == "" {
		actor = "capcom-sync-worker"
	}
	reason := strings.TrimSpace(input.Reason)
	if reason == "" {
		reason = "scheduled runtime synchronization"
	}
	_, _ = s.audit.Create(ctx, domain.AuditEvent{RuntimeConnectionID: run.RuntimeConnectionID, Actor: actor,
		EventType: eventType, TargetType: "runtime_sync_run", TargetID: run.ID, Reason: reason, Result: result,
		Metadata: map[string]any{"trigger": run.Trigger, "status": run.Status, "agents_seen": run.AgentsSeen,
			"skills_seen": run.SkillsSeen, "duration_ms": run.DurationMS, "error_code": run.ErrorCode}})
}

func sanitizeRuntimeError(err error) string {
	message := strings.TrimSpace(err.Error())
	if len(message) > 1000 {
		message = message[:1000]
	}
	return message
}

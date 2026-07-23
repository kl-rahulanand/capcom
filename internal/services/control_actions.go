package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	runtimeadapter "capcom/internal/adapters/runtime"
	"capcom/internal/domain"
)

type ControlActionRepository interface {
	Create(ctx context.Context, action domain.ControlAction) (domain.ControlAction, error)
	FindByIdempotencyKey(ctx context.Context, key string) (domain.ControlAction, error)
	Update(ctx context.Context, action domain.ControlAction, runtimeRequest, runtimeResponse map[string]any, errorText string) (domain.ControlAction, error)
}

type PostActionSyncer interface {
	Sync(ctx context.Context, input SyncRuntimeInput) (domain.RuntimeSyncRun, error)
}

type ControlActionService struct {
	runtimes RuntimeConnectionRepository
	agents   RuntimeSyncRepository
	actions  ControlActionRepository
	audit    AuditRepository
	syncer   PostActionSyncer
	adapters map[domain.RuntimeKind]runtimeadapter.Adapter
}

type ReconcileAccessInput struct {
	AgentID        string
	Access         domain.AccessDocument
	Actor          string
	Reason         string
	IdempotencyKey string
	DryRun         bool
}

type SetAgentStatusInput struct {
	AgentID        string
	Status         domain.AgentStatus
	Actor          string
	Reason         string
	IdempotencyKey string
	DryRun         bool
}

func NewControlActionService(runtimes RuntimeConnectionRepository, agents RuntimeSyncRepository, actions ControlActionRepository, audit AuditRepository, syncer PostActionSyncer) ControlActionService {
	return ControlActionService{runtimes: runtimes, agents: agents, actions: actions, audit: audit, syncer: syncer, adapters: map[domain.RuntimeKind]runtimeadapter.Adapter{}}
}
func (s ControlActionService) WithAdapter(adapter runtimeadapter.Adapter) ControlActionService {
	if adapter != nil {
		s.adapters[adapter.Kind()] = adapter
	}
	return s
}

func (s ControlActionService) ReconcileAccess(ctx context.Context, input ReconcileAccessInput) (domain.ControlAction, error) {
	if strings.TrimSpace(input.AgentID) == "" || strings.TrimSpace(input.Actor) == "" || strings.TrimSpace(input.Reason) == "" || strings.TrimSpace(input.IdempotencyKey) == "" {
		return domain.ControlAction{}, fmt.Errorf("agent id, actor, reason, and idempotency_key are required")
	}
	for _, selection := range input.Access.Selections {
		if strings.TrimSpace(selection.ID) == "" {
			return domain.ControlAction{}, fmt.Errorf("every access selection requires an id")
		}
	}
	if existing, err := s.actions.FindByIdempotencyKey(ctx, input.IdempotencyKey); err == nil {
		return existing, nil
	} else if err != sql.ErrNoRows {
		return domain.ControlAction{}, err
	}
	detail, err := s.agents.GetPersistedAgent(ctx, input.AgentID)
	if err != nil {
		return domain.ControlAction{}, err
	}
	conn, err := s.runtimes.Get(ctx, detail.Agent.RuntimeConnectionID)
	if err != nil {
		return domain.ControlAction{}, err
	}
	before := accessMap(detail.Access)
	requested := accessMap(input.Access)
	action := domain.ControlAction{RuntimeConnectionID: conn.ID, AgentID: input.AgentID, Type: "replace_agent_access", Status: domain.ControlActionQueued,
		Actor: strings.TrimSpace(input.Actor), Reason: strings.TrimSpace(input.Reason), IdempotencyKey: strings.TrimSpace(input.IdempotencyKey), Before: before, After: requested}
	if conn.Mode != domain.RuntimeModeControlEnabled {
		action.Status = domain.ControlActionRejected
		action, err = s.actions.Create(ctx, action)
		if err != nil {
			return action, err
		}
		action, _ = s.actions.Update(ctx, action, requested, nil, "runtime connection is read-only")
		s.auditAction(ctx, action, "control_action.rejected", "rejected", before, requested, map[string]any{"reason": "read_only_runtime"})
		return action, fmt.Errorf("runtime connection is read-only")
	}
	adapter, ok := s.adapters[conn.Kind]
	if !ok {
		action.Status = domain.ControlActionRejected
		action, err = s.actions.Create(ctx, action)
		if err != nil {
			return action, err
		}
		action, _ = s.actions.Update(ctx, action, requested, nil, "runtime adapter is not registered")
		s.auditAction(ctx, action, "control_action.rejected", "rejected", before, requested, map[string]any{"reason": "adapter_unavailable"})
		return action, fmt.Errorf("runtime adapter %q is not registered", conn.Kind)
	}
	action, err = s.actions.Create(ctx, action)
	if err != nil {
		return action, err
	}
	s.auditAction(ctx, action, "control_action.requested", "succeeded", before, requested, nil)
	if input.DryRun {
		action.Status = domain.ControlActionSucceeded
		result := map[string]any{"dry_run": true, "validated": true}
		action, err = s.actions.Update(ctx, action, requested, result, "")
		s.auditAction(ctx, action, "control_action.dry_run_succeeded", "succeeded", before, requested, result)
		return action, err
	}
	action.Status = domain.ControlActionRunning
	action, err = s.actions.Update(ctx, action, requested, nil, "")
	if err != nil {
		return action, err
	}
	observed, callErr := adapter.ReplaceAgentAccess(ctx, conn, detail.Agent.RuntimeAgentID, input.Access)
	if callErr != nil {
		action.Status = domain.ControlActionFailed
		action, _ = s.actions.Update(ctx, action, requested, nil, sanitizeRuntimeError(callErr))
		s.auditAction(ctx, action, "control_action.failed", "failed", before, requested, map[string]any{"error": sanitizeRuntimeError(callErr)})
		return action, fmt.Errorf("replace runtime access: %w", callErr)
	}
	result := accessMap(*observed)
	action.Status = domain.ControlActionSucceeded
	action, err = s.actions.Update(ctx, action, requested, result, "")
	if err != nil {
		return action, err
	}
	if s.syncer != nil {
		_, syncErr := s.syncer.Sync(ctx, SyncRuntimeInput{RuntimeConnectionID: conn.ID, Trigger: domain.SyncTriggerPostAction, Actor: input.Actor, Reason: "verify access reconciliation: " + input.Reason})
		if syncErr != nil {
			result["verification_sync_error"] = sanitizeRuntimeError(syncErr)
		}
	}
	s.auditAction(ctx, action, "control_action.succeeded", "succeeded", before, requested, result)
	return action, nil
}

func (s ControlActionService) SetAgentStatus(ctx context.Context, input SetAgentStatusInput) (domain.ControlAction, error) {
	if strings.TrimSpace(input.AgentID) == "" || strings.TrimSpace(input.Actor) == "" || strings.TrimSpace(input.Reason) == "" || strings.TrimSpace(input.IdempotencyKey) == "" {
		return domain.ControlAction{}, fmt.Errorf("agent id, actor, reason, and idempotency_key are required")
	}
	if input.Status != domain.AgentStatusEnabled && input.Status != domain.AgentStatusDisabled {
		return domain.ControlAction{}, fmt.Errorf("status must be %q or %q", domain.AgentStatusEnabled, domain.AgentStatusDisabled)
	}
	if existing, err := s.actions.FindByIdempotencyKey(ctx, input.IdempotencyKey); err == nil {
		return existing, nil
	} else if err != sql.ErrNoRows {
		return domain.ControlAction{}, err
	}
	detail, err := s.agents.GetPersistedAgent(ctx, input.AgentID)
	if err != nil {
		return domain.ControlAction{}, err
	}
	conn, err := s.runtimes.Get(ctx, detail.Agent.RuntimeConnectionID)
	if err != nil {
		return domain.ControlAction{}, err
	}
	actionType := "enable_agent"
	if input.Status == domain.AgentStatusDisabled {
		actionType = "disable_agent"
	}
	before := map[string]any{"status": detail.Agent.Status}
	after := map[string]any{"status": input.Status}
	action := domain.ControlAction{RuntimeConnectionID: conn.ID, AgentID: input.AgentID, Type: actionType,
		Status: domain.ControlActionQueued, Actor: strings.TrimSpace(input.Actor), Reason: strings.TrimSpace(input.Reason),
		IdempotencyKey: strings.TrimSpace(input.IdempotencyKey), Before: before, After: after}
	if conn.Mode != domain.RuntimeModeControlEnabled {
		return s.rejectStatusAction(ctx, action, before, after, "runtime connection is read-only", "read_only_runtime")
	}
	adapter, ok := s.adapters[conn.Kind]
	if !ok {
		return s.rejectStatusAction(ctx, action, before, after, "runtime adapter is not registered", "adapter_unavailable")
	}
	check, checkErr := adapter.Check(ctx, conn)
	if checkErr != nil {
		return s.rejectStatusAction(ctx, action, before, after, sanitizeRuntimeError(checkErr), "adapter_check_failed")
	}
	if !check.Capabilities.SetAgentStatus {
		return s.rejectStatusAction(ctx, action, before, after, "runtime adapter does not support agent status control", "unsupported_action")
	}
	action, err = s.actions.Create(ctx, action)
	if err != nil {
		return action, err
	}
	s.auditStatusAction(ctx, action, "control_action.requested", "succeeded", before, after, nil)
	if input.DryRun {
		action.Status = domain.ControlActionSucceeded
		result := map[string]any{"dry_run": true, "validated": true, "status": input.Status}
		action, err = s.actions.Update(ctx, action, after, result, "")
		s.auditStatusAction(ctx, action, "control_action.dry_run_succeeded", "succeeded", before, after, result)
		return action, err
	}
	action.Status = domain.ControlActionRunning
	action, err = s.actions.Update(ctx, action, after, nil, "")
	if err != nil {
		return action, err
	}
	observed, callErr := adapter.SetAgentStatus(ctx, conn, detail.Agent.RuntimeAgentID, input.Status)
	if callErr != nil {
		action.Status = domain.ControlActionFailed
		action, _ = s.actions.Update(ctx, action, after, nil, sanitizeRuntimeError(callErr))
		s.auditStatusAction(ctx, action, "control_action.failed", "failed", before, after, map[string]any{"error": sanitizeRuntimeError(callErr)})
		return action, fmt.Errorf("set runtime agent status: %w", callErr)
	}
	result := map[string]any{"status": observed.Status, "runtime_agent_id": observed.RuntimeAgentID}
	action.Status = domain.ControlActionSucceeded
	action, err = s.actions.Update(ctx, action, after, result, "")
	if err != nil {
		return action, err
	}
	if s.syncer != nil {
		_, syncErr := s.syncer.Sync(ctx, SyncRuntimeInput{RuntimeConnectionID: conn.ID, Trigger: domain.SyncTriggerPostAction,
			Actor: input.Actor, Reason: "verify agent status change: " + input.Reason})
		if syncErr != nil {
			result["verification_sync_error"] = sanitizeRuntimeError(syncErr)
		}
	}
	s.auditStatusAction(ctx, action, "control_action.succeeded", "succeeded", before, after, result)
	return action, nil
}

func (s ControlActionService) rejectStatusAction(ctx context.Context, action domain.ControlAction, before, after map[string]any, message, reason string) (domain.ControlAction, error) {
	action.Status = domain.ControlActionRejected
	created, err := s.actions.Create(ctx, action)
	if err != nil {
		return action, err
	}
	created, _ = s.actions.Update(ctx, created, after, nil, message)
	s.auditStatusAction(ctx, created, "control_action.rejected", "rejected", before, after, map[string]any{"reason": reason})
	return created, fmt.Errorf("%s", message)
}

func (s ControlActionService) auditStatusAction(ctx context.Context, action domain.ControlAction, eventType, result string, before, after, metadata map[string]any) {
	if s.audit == nil {
		return
	}
	_, _ = s.audit.Create(ctx, domain.AuditEvent{RuntimeConnectionID: action.RuntimeConnectionID, AgentID: action.AgentID,
		ControlActionID: action.ID, Actor: action.Actor, EventType: eventType, TargetType: "agent_status",
		TargetID: action.AgentID, Reason: action.Reason, Before: before, After: after, Result: result, Metadata: metadata})
}

func accessMap(access domain.AccessDocument) map[string]any {
	data, _ := json.Marshal(access)
	var result map[string]any
	_ = json.Unmarshal(data, &result)
	return result
}
func (s ControlActionService) auditAction(ctx context.Context, action domain.ControlAction, eventType, result string, before, after, metadata map[string]any) {
	if s.audit == nil {
		return
	}
	_, _ = s.audit.Create(ctx, domain.AuditEvent{RuntimeConnectionID: action.RuntimeConnectionID, AgentID: action.AgentID, ControlActionID: action.ID, Actor: action.Actor, EventType: eventType, TargetType: "agent_access", TargetID: action.AgentID, Reason: action.Reason, Before: before, After: after, Result: result, Metadata: metadata})
}

package store

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"capcom/internal/domain"
)

type RuntimeConnectionRepository struct {
	db *sql.DB
}

type AgentRepository struct {
	db *sql.DB
}

type AuditRepository struct {
	db *sql.DB
}

func NewRuntimeConnectionRepository(db *sql.DB) RuntimeConnectionRepository {
	return RuntimeConnectionRepository{db: db}
}

func NewAgentRepository(db *sql.DB) AgentRepository {
	return AgentRepository{db: db}
}

func NewAuditRepository(db *sql.DB) AuditRepository {
	return AuditRepository{db: db}
}

func (r AuditRepository) Create(ctx context.Context, event domain.AuditEvent) (domain.AuditEvent, error) {
	if event.ID == "" {
		id, err := newID()
		if err != nil {
			return domain.AuditEvent{}, err
		}
		event.ID = id
	}
	event.CreatedAt = time.Now().UTC()

	beforeJSON, err := marshalNullableJSON(event.Before)
	if err != nil {
		return domain.AuditEvent{}, fmt.Errorf("marshal audit before: %w", err)
	}
	afterJSON, err := marshalNullableJSON(event.After)
	if err != nil {
		return domain.AuditEvent{}, fmt.Errorf("marshal audit after: %w", err)
	}
	metadataJSON, err := marshalNullableJSON(event.Metadata)
	if err != nil {
		return domain.AuditEvent{}, fmt.Errorf("marshal audit metadata: %w", err)
	}

	if _, err := r.db.ExecContext(ctx, `
INSERT INTO audit_events (
	id, runtime_connection_id, agent_id, control_action_id, actor, event_type,
	target_type, target_id, reason, before_json, after_json, result, metadata_json, created_at
) VALUES ($1, NULLIF($2, '')::uuid, NULLIF($3, '')::uuid, NULLIF($4, '')::uuid, $5, $6, $7, $8, $9, $10::jsonb, $11::jsonb, $12, COALESCE($13::jsonb, '{}'::jsonb), $14)`,
		event.ID,
		event.RuntimeConnectionID,
		event.AgentID,
		event.ControlActionID,
		event.Actor,
		event.EventType,
		event.TargetType,
		event.TargetID,
		event.Reason,
		beforeJSON,
		afterJSON,
		event.Result,
		metadataJSON,
		event.CreatedAt,
	); err != nil {
		return domain.AuditEvent{}, fmt.Errorf("create audit event: %w", err)
	}

	return event, nil
}

func (r RuntimeConnectionRepository) Create(ctx context.Context, conn domain.RuntimeConnection) (domain.RuntimeConnection, error) {
	if conn.ID == "" {
		id, err := newID()
		if err != nil {
			return domain.RuntimeConnection{}, err
		}
		conn.ID = id
	}
	now := time.Now().UTC()
	conn.CreatedAt = now
	conn.UpdatedAt = now

	if _, err := r.db.ExecContext(ctx, `
INSERT INTO runtime_connections (
	id, name, runtime_type, mode, status, endpoint, metadata_json, created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, $6, '{}'::jsonb, $7, $8)`,
		conn.ID, conn.Name, conn.Kind, conn.Mode, conn.Status, conn.BaseURL, conn.CreatedAt, conn.UpdatedAt); err != nil {
		return domain.RuntimeConnection{}, fmt.Errorf("create runtime connection: %w", err)
	}
	return conn, nil
}

func (r RuntimeConnectionRepository) Get(ctx context.Context, id string) (domain.RuntimeConnection, error) {
	var conn domain.RuntimeConnection
	var lastSyncedAt sql.NullTime
	if err := r.db.QueryRowContext(ctx, `
SELECT id, name, runtime_type, mode, status, endpoint, created_at, updated_at, last_sync_at
FROM runtime_connections
WHERE id = $1`, id).Scan(
		&conn.ID,
		&conn.Name,
		&conn.Kind,
		&conn.Mode,
		&conn.Status,
		&conn.BaseURL,
		&conn.CreatedAt,
		&conn.UpdatedAt,
		&lastSyncedAt,
	); err != nil {
		return domain.RuntimeConnection{}, fmt.Errorf("get runtime connection: %w", err)
	}
	if lastSyncedAt.Valid {
		conn.LastSyncedAt = &lastSyncedAt.Time
	}
	return conn, nil
}

func (r RuntimeConnectionRepository) List(ctx context.Context) ([]domain.RuntimeConnection, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT id, name, runtime_type, mode, status, endpoint, created_at, updated_at, last_sync_at
FROM runtime_connections
ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list runtime connections: %w", err)
	}
	defer rows.Close()

	var conns []domain.RuntimeConnection
	for rows.Next() {
		var conn domain.RuntimeConnection
		var lastSyncedAt sql.NullTime
		if err := rows.Scan(
			&conn.ID,
			&conn.Name,
			&conn.Kind,
			&conn.Mode,
			&conn.Status,
			&conn.BaseURL,
			&conn.CreatedAt,
			&conn.UpdatedAt,
			&lastSyncedAt,
		); err != nil {
			return nil, fmt.Errorf("scan runtime connection: %w", err)
		}
		if lastSyncedAt.Valid {
			conn.LastSyncedAt = &lastSyncedAt.Time
		}
		conns = append(conns, conn)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate runtime connections: %w", err)
	}
	return conns, nil
}

func (r AgentRepository) Create(ctx context.Context, agent domain.Agent, binding domain.AgentBinding) (domain.Agent, domain.AgentBinding, error) {
	if agent.ID == "" {
		id, err := newID()
		if err != nil {
			return domain.Agent{}, domain.AgentBinding{}, err
		}
		agent.ID = id
	}
	if binding.ID == "" {
		id, err := newID()
		if err != nil {
			return domain.Agent{}, domain.AgentBinding{}, err
		}
		binding.ID = id
	}
	binding.AgentID = agent.ID
	now := time.Now().UTC()
	agent.CreatedAt = now
	agent.UpdatedAt = now
	binding.CreatedAt = now
	binding.UpdatedAt = now

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Agent{}, domain.AgentBinding{}, fmt.Errorf("begin create agent: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
INSERT INTO agents (id, name, status, metadata_json, created_at, updated_at)
VALUES ($1, $2, $3, '{}'::jsonb, $4, $5)`,
		agent.ID, agent.Name, agent.Status, agent.CreatedAt, agent.UpdatedAt); err != nil {
		return domain.Agent{}, domain.AgentBinding{}, fmt.Errorf("create agent: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
INSERT INTO agent_runtime_bindings (
	id, agent_id, runtime_connection_id, runtime_agent_id, created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, $6)`,
		binding.ID, binding.AgentID, binding.RuntimeConnectionID, binding.RuntimeAgentID, binding.CreatedAt, binding.UpdatedAt); err != nil {
		return domain.Agent{}, domain.AgentBinding{}, fmt.Errorf("create agent binding: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return domain.Agent{}, domain.AgentBinding{}, fmt.Errorf("commit create agent: %w", err)
	}

	return agent, binding, nil
}

func (r AgentRepository) Get(ctx context.Context, id string) (domain.Agent, error) {
	var agent domain.Agent
	if err := r.db.QueryRowContext(ctx, `
SELECT id, name, status, created_at, updated_at
FROM agents
WHERE id = $1`, id).Scan(&agent.ID, &agent.Name, &agent.Status, &agent.CreatedAt, &agent.UpdatedAt); err != nil {
		return domain.Agent{}, fmt.Errorf("get agent: %w", err)
	}
	return agent, nil
}

func (r AgentRepository) List(ctx context.Context) ([]domain.Agent, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT id, name, status, created_at, updated_at
FROM agents
ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list agents: %w", err)
	}
	defer rows.Close()

	var agents []domain.Agent
	for rows.Next() {
		var agent domain.Agent
		if err := rows.Scan(&agent.ID, &agent.Name, &agent.Status, &agent.CreatedAt, &agent.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan agent: %w", err)
		}
		agents = append(agents, agent)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate agents: %w", err)
	}
	return agents, nil
}

func newID() (string, error) {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}
	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80

	encoded := hex.EncodeToString(bytes[:])
	return fmt.Sprintf("%s-%s-%s-%s-%s", encoded[0:8], encoded[8:12], encoded[12:16], encoded[16:20], encoded[20:32]), nil
}

func marshalNullableJSON(value map[string]any) (any, error) {
	if value == nil {
		return nil, nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return string(data), nil
}

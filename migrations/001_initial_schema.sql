CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS runtime_connections (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL UNIQUE,
    runtime_type text NOT NULL,
    mode text NOT NULL CHECK (mode IN ('read_only', 'control_enabled')),
    status text NOT NULL CHECK (status IN ('pending', 'active', 'degraded', 'disabled', 'failed')),
    endpoint text NOT NULL,
    auth_ref text,
    last_sync_at timestamptz,
    last_error text,
    metadata_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_runtime_connections_type_status
    ON runtime_connections (runtime_type, status);

CREATE TABLE IF NOT EXISTS agents (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL,
    status text NOT NULL CHECK (status IN ('unknown', 'enabled', 'disabled', 'stale')),
    owner_business text,
    owner_technical text,
    escalation_contact text,
    purpose text,
    environment text,
    risk_level text NOT NULL DEFAULT 'low' CHECK (risk_level IN ('low', 'medium', 'high', 'critical')),
    metadata_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_agents_status
    ON agents (status);

CREATE INDEX IF NOT EXISTS idx_agents_risk_level
    ON agents (risk_level);

CREATE TABLE IF NOT EXISTS agent_runtime_bindings (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id uuid NOT NULL REFERENCES agents (id) ON DELETE CASCADE,
    runtime_connection_id uuid NOT NULL REFERENCES runtime_connections (id) ON DELETE CASCADE,
    runtime_agent_id text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (runtime_connection_id, runtime_agent_id),
    UNIQUE (agent_id, runtime_connection_id)
);

CREATE INDEX IF NOT EXISTS idx_agent_runtime_bindings_agent
    ON agent_runtime_bindings (agent_id);

CREATE TABLE IF NOT EXISTS access_desired_state (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id uuid NOT NULL REFERENCES agents (id) ON DELETE CASCADE,
    manifest_version text NOT NULL,
    desired_status text NOT NULL,
    access_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    approvals_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    policies_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    manifest_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    source text NOT NULL,
    applied_by text NOT NULL,
    applied_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (agent_id)
);

CREATE INDEX IF NOT EXISTS idx_access_desired_state_access_json
    ON access_desired_state USING gin (access_json);

CREATE TABLE IF NOT EXISTS access_actual_state (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id uuid NOT NULL REFERENCES agents (id) ON DELETE CASCADE,
    runtime_connection_id uuid NOT NULL REFERENCES runtime_connections (id) ON DELETE CASCADE,
    runtime_status text NOT NULL,
    access_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    admin_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    inventory_refs_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    raw_runtime_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    observed_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (agent_id, runtime_connection_id)
);

CREATE INDEX IF NOT EXISTS idx_access_actual_state_access_json
    ON access_actual_state USING gin (access_json);

CREATE TABLE IF NOT EXISTS drift_findings (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id uuid NOT NULL REFERENCES agents (id) ON DELETE CASCADE,
    drift_type text NOT NULL,
    field_path text NOT NULL,
    expected_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    actual_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    severity text NOT NULL CHECK (severity IN ('info', 'warning', 'critical')),
    mode text NOT NULL DEFAULT 'observe' CHECK (mode IN ('observe', 'approval', 'enforce')),
    status text NOT NULL CHECK (status IN ('open', 'acknowledged', 'resolved', 'ignored')),
    first_seen_at timestamptz NOT NULL DEFAULT now(),
    last_seen_at timestamptz NOT NULL DEFAULT now(),
    resolved_at timestamptz
);

CREATE INDEX IF NOT EXISTS idx_drift_findings_agent_status
    ON drift_findings (agent_id, status);

CREATE UNIQUE INDEX IF NOT EXISTS idx_drift_findings_open_key
    ON drift_findings (agent_id, drift_type, field_path)
    WHERE status IN ('open', 'acknowledged');

CREATE TABLE IF NOT EXISTS control_actions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    runtime_connection_id uuid NOT NULL REFERENCES runtime_connections (id),
    agent_id uuid REFERENCES agents (id),
    action_type text NOT NULL,
    requested_by text NOT NULL,
    reason text NOT NULL,
    idempotency_key text,
    parameters_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    status text NOT NULL CHECK (status IN ('queued', 'running', 'succeeded', 'failed', 'rejected')),
    runtime_request_json jsonb,
    runtime_response_json jsonb,
    error text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    finished_at timestamptz,
    UNIQUE (idempotency_key)
);

CREATE INDEX IF NOT EXISTS idx_control_actions_runtime_status
    ON control_actions (runtime_connection_id, status);

CREATE INDEX IF NOT EXISTS idx_control_actions_agent_created
    ON control_actions (agent_id, created_at DESC);

CREATE TABLE IF NOT EXISTS audit_events (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    runtime_connection_id uuid REFERENCES runtime_connections (id),
    agent_id uuid REFERENCES agents (id),
    control_action_id uuid REFERENCES control_actions (id),
    actor text NOT NULL,
    event_type text NOT NULL,
    target_type text NOT NULL,
    target_id text NOT NULL,
    reason text,
    before_json jsonb,
    after_json jsonb,
    result text NOT NULL CHECK (result IN ('succeeded', 'failed', 'rejected')),
    metadata_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_audit_events_target_created
    ON audit_events (target_type, target_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_audit_events_actor_created
    ON audit_events (actor, created_at DESC);

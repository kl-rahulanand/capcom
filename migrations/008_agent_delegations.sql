ALTER TABLE runtime_sync_runs
    ADD COLUMN IF NOT EXISTS delegations_seen integer NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS agent_delegations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    runtime_connection_id uuid NOT NULL REFERENCES runtime_connections (id) ON DELETE CASCADE,
    orchestrator_runtime_agent_id text NOT NULL,
    delegate_key text NOT NULL,
    delegate_runtime_agent_id text NOT NULL DEFAULT '',
    delegate_ref text NOT NULL,
    tool_name text NOT NULL DEFAULT '',
    display_name text NOT NULL DEFAULT '',
    persona text NOT NULL DEFAULT '',
    configured boolean NOT NULL DEFAULT false,
    resolved boolean NOT NULL DEFAULT false,
    revision integer NOT NULL DEFAULT 0,
    status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'stale')),
    observed_at timestamptz NOT NULL,
    last_seen_sync_run_id uuid REFERENCES runtime_sync_runs (id),
    missing_successful_syncs integer NOT NULL DEFAULT 0,
    metadata_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    raw_runtime_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (runtime_connection_id, orchestrator_runtime_agent_id, delegate_key)
);

CREATE INDEX IF NOT EXISTS idx_agent_delegations_orchestrator
    ON agent_delegations (runtime_connection_id, orchestrator_runtime_agent_id, status);

CREATE INDEX IF NOT EXISTS idx_agent_delegations_delegate
    ON agent_delegations (runtime_connection_id, delegate_runtime_agent_id, status);

ALTER TABLE runtime_sync_runs
    ADD COLUMN IF NOT EXISTS executions_seen integer NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS runtime_executions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    runtime_connection_id uuid NOT NULL REFERENCES runtime_connections (id) ON DELETE CASCADE,
    runtime_execution_id text NOT NULL,
    parent_runtime_execution_id text,
    runtime_agent_id text NOT NULL DEFAULT '',
    kind text NOT NULL,
    status text NOT NULL,
    started_at timestamptz,
    ended_at timestamptz,
    observed_at timestamptz NOT NULL,
    metadata_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    raw_runtime_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (runtime_connection_id, kind, runtime_execution_id)
);

CREATE INDEX IF NOT EXISTS idx_runtime_executions_runtime_observed
    ON runtime_executions (runtime_connection_id, observed_at DESC);
CREATE INDEX IF NOT EXISTS idx_runtime_executions_agent_observed
    ON runtime_executions (runtime_connection_id, runtime_agent_id, observed_at DESC);
CREATE INDEX IF NOT EXISTS idx_runtime_executions_parent
    ON runtime_executions (runtime_connection_id, parent_runtime_execution_id);

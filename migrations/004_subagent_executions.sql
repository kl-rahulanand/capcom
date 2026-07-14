CREATE TABLE IF NOT EXISTS subagent_executions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    runtime_connection_id uuid NOT NULL REFERENCES runtime_connections (id) ON DELETE CASCADE,
    runtime_execution_id text NOT NULL,
    parent_run_id text NOT NULL,
    runtime_agent_id text NOT NULL DEFAULT '',
    subagent_type text NOT NULL DEFAULT '',
    status text NOT NULL DEFAULT 'running',
    description text NOT NULL DEFAULT '',
    summary text NOT NULL DEFAULT '',
    started_at timestamptz,
    ended_at timestamptz,
    observed_at timestamptz NOT NULL,
    metadata_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    raw_runtime_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (runtime_connection_id, runtime_execution_id)
);

CREATE INDEX IF NOT EXISTS idx_subagent_executions_runtime_observed
    ON subagent_executions (runtime_connection_id, observed_at DESC);
CREATE INDEX IF NOT EXISTS idx_subagent_executions_agent_observed
    ON subagent_executions (runtime_connection_id, runtime_agent_id, observed_at DESC);

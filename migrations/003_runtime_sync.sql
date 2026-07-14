ALTER TABLE runtime_connections
    ADD COLUMN IF NOT EXISTS sync_enabled boolean NOT NULL DEFAULT true,
    ADD COLUMN IF NOT EXISTS sync_interval_seconds integer NOT NULL DEFAULT 60,
    ADD COLUMN IF NOT EXISTS last_sync_status text,
    ADD COLUMN IF NOT EXISTS last_sync_started_at timestamptz,
    ADD COLUMN IF NOT EXISTS last_sync_finished_at timestamptz,
    ADD COLUMN IF NOT EXISTS last_sync_duration_ms bigint;

ALTER TABLE runtime_connections
    DROP CONSTRAINT IF EXISTS runtime_connections_sync_interval_check;
ALTER TABLE runtime_connections
    ADD CONSTRAINT runtime_connections_sync_interval_check
    CHECK (sync_interval_seconds BETWEEN 15 AND 86400);

CREATE TABLE IF NOT EXISTS runtime_sync_runs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    runtime_connection_id uuid NOT NULL REFERENCES runtime_connections (id) ON DELETE CASCADE,
    trigger text NOT NULL CHECK (trigger IN ('manual', 'scheduled', 'post_action')),
    status text NOT NULL CHECK (status IN ('running', 'succeeded', 'failed', 'skipped')),
    started_at timestamptz NOT NULL DEFAULT now(),
    finished_at timestamptz,
    duration_ms bigint,
    agents_seen integer NOT NULL DEFAULT 0,
    skills_seen integer NOT NULL DEFAULT 0,
    bindings_seen integer NOT NULL DEFAULT 0,
    access_documents_seen integer NOT NULL DEFAULT 0,
    error_code text,
    error_message text,
    metadata_json jsonb NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX IF NOT EXISTS idx_runtime_sync_runs_runtime_started
    ON runtime_sync_runs (runtime_connection_id, started_at DESC);
CREATE INDEX IF NOT EXISTS idx_runtime_sync_runs_status_started
    ON runtime_sync_runs (status, started_at DESC);

ALTER TABLE agent_runtime_bindings
    ADD COLUMN IF NOT EXISTS kind text NOT NULL DEFAULT 'registered',
    ADD COLUMN IF NOT EXISTS parent_runtime_agent_id text,
    ADD COLUMN IF NOT EXISTS last_seen_at timestamptz,
    ADD COLUMN IF NOT EXISTS last_seen_sync_run_id uuid REFERENCES runtime_sync_runs (id),
    ADD COLUMN IF NOT EXISTS missing_successful_syncs integer NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS raw_runtime_json jsonb NOT NULL DEFAULT '{}'::jsonb;

ALTER TABLE agent_runtime_bindings
    DROP CONSTRAINT IF EXISTS agent_runtime_bindings_kind_check;
ALTER TABLE agent_runtime_bindings
    ADD CONSTRAINT agent_runtime_bindings_kind_check
    CHECK (kind IN ('main', 'registered', 'subagent'));

CREATE TABLE IF NOT EXISTS runtime_skills (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    runtime_connection_id uuid NOT NULL REFERENCES runtime_connections (id) ON DELETE CASCADE,
    runtime_skill_id text NOT NULL,
    name text NOT NULL,
    description text NOT NULL DEFAULT '',
    source text NOT NULL DEFAULT '',
    status text NOT NULL DEFAULT 'active',
    version text NOT NULL DEFAULT '',
    tool_ids_json jsonb NOT NULL DEFAULT '[]'::jsonb,
    workflow_refs_json jsonb NOT NULL DEFAULT '[]'::jsonb,
    metadata_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    raw_runtime_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    observed_at timestamptz NOT NULL,
    last_seen_sync_run_id uuid REFERENCES runtime_sync_runs (id),
    missing_successful_syncs integer NOT NULL DEFAULT 0,
    UNIQUE (runtime_connection_id, runtime_skill_id)
);

CREATE INDEX IF NOT EXISTS idx_runtime_skills_runtime_name
    ON runtime_skills (runtime_connection_id, name);

CREATE TABLE IF NOT EXISTS agent_skill_bindings (
    agent_id uuid NOT NULL REFERENCES agents (id) ON DELETE CASCADE,
    runtime_skill_id uuid NOT NULL REFERENCES runtime_skills (id) ON DELETE CASCADE,
    status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'stale')),
    observed_at timestamptz NOT NULL,
    last_seen_sync_run_id uuid REFERENCES runtime_sync_runs (id),
    missing_successful_syncs integer NOT NULL DEFAULT 0,
    PRIMARY KEY (agent_id, runtime_skill_id)
);

CREATE INDEX IF NOT EXISTS idx_agent_skill_bindings_agent_status
    ON agent_skill_bindings (agent_id, status);

ALTER TABLE access_actual_state
    ADD COLUMN IF NOT EXISTS last_seen_sync_run_id uuid REFERENCES runtime_sync_runs (id),
    ADD COLUMN IF NOT EXISTS freshness_status text NOT NULL DEFAULT 'stale';

ALTER TABLE access_actual_state
    DROP CONSTRAINT IF EXISTS access_actual_state_freshness_check;
ALTER TABLE access_actual_state
    ADD CONSTRAINT access_actual_state_freshness_check
    CHECK (freshness_status IN ('live', 'cached', 'stale'));

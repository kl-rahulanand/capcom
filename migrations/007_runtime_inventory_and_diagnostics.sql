ALTER TABLE runtime_sync_runs
    ADD COLUMN IF NOT EXISTS diagnostics_seen integer NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS inventory_seen integer NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS capabilities_seen integer NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS runtime_diagnostics (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    runtime_connection_id uuid NOT NULL REFERENCES runtime_connections (id) ON DELETE CASCADE,
    check_id text NOT NULL,
    status text NOT NULL,
    message text NOT NULL DEFAULT '',
    observed_at timestamptz NOT NULL,
    metadata_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    raw_runtime_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (runtime_connection_id, check_id)
);

CREATE INDEX IF NOT EXISTS idx_runtime_diagnostics_runtime_status
    ON runtime_diagnostics (runtime_connection_id, status);

CREATE TABLE IF NOT EXISTS runtime_inventory_items (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    runtime_connection_id uuid NOT NULL REFERENCES runtime_connections (id) ON DELETE CASCADE,
    runtime_item_id text NOT NULL,
    kind text NOT NULL,
    name text NOT NULL,
    status text NOT NULL DEFAULT 'unknown',
    provider text NOT NULL DEFAULT '',
    source text NOT NULL DEFAULT '',
    observed_at timestamptz NOT NULL,
    metadata_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    raw_runtime_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (runtime_connection_id, kind, runtime_item_id)
);

CREATE INDEX IF NOT EXISTS idx_runtime_inventory_runtime_kind
    ON runtime_inventory_items (runtime_connection_id, kind, status);

CREATE TABLE IF NOT EXISTS runtime_capabilities (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    runtime_connection_id uuid NOT NULL REFERENCES runtime_connections (id) ON DELETE CASCADE,
    runtime_capability_id text NOT NULL,
    version text NOT NULL,
    name text NOT NULL,
    category text NOT NULL DEFAULT '',
    risk text NOT NULL DEFAULT '',
    can_description text NOT NULL DEFAULT '',
    cannot_description text NOT NULL DEFAULT '',
    source text NOT NULL DEFAULT '',
    observed_at timestamptz NOT NULL,
    metadata_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    raw_runtime_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (runtime_connection_id, runtime_capability_id, version)
);

CREATE INDEX IF NOT EXISTS idx_runtime_capabilities_runtime_risk
    ON runtime_capabilities (runtime_connection_id, risk, category);

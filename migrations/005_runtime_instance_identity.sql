ALTER TABLE runtime_connections
    ADD COLUMN IF NOT EXISTS display_name text NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS environment text NOT NULL DEFAULT 'unspecified',
    ADD COLUMN IF NOT EXISTS labels_json jsonb NOT NULL DEFAULT '{}'::jsonb;

UPDATE runtime_connections
SET display_name = name
WHERE display_name = '';

ALTER TABLE runtime_connections
    DROP CONSTRAINT IF EXISTS runtime_connections_environment_check;
ALTER TABLE runtime_connections
    ADD CONSTRAINT runtime_connections_environment_check
    CHECK (environment ~ '^[a-z][a-z0-9_-]{0,31}$');

CREATE UNIQUE INDEX IF NOT EXISTS idx_runtime_connections_unique_endpoint
    ON runtime_connections (runtime_type, lower(rtrim(endpoint, '/')));

CREATE INDEX IF NOT EXISTS idx_runtime_connections_environment_status
    ON runtime_connections (environment, status);

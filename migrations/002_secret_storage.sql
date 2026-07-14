CREATE TABLE IF NOT EXISTS secrets (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL UNIQUE,
    ciphertext bytea NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK (name ~ '^[A-Za-z][A-Za-z0-9._-]{0,127}$')
);

CREATE INDEX IF NOT EXISTS idx_secrets_updated_at
    ON secrets (updated_at DESC);

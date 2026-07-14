# Capcom

Capcom is a runtime-agnostic AgentOps control plane. Gantry is the first runtime adapter, but the core architecture is designed so other agent runtimes can be added behind the same adapter boundary.

## Current Status

Implementation has started with the first backend slices:

- Go module at the repository root.
- `capcom-server` binary with graceful shutdown.
- `capcom` CLI placeholder.
- Environment-based config loading.
- Structured JSON logging.
- `GET /healthz` endpoint.
- Runtime-neutral domain shell.
- Runtime adapter interface for Gantry and future adapters.
- `.env` loading for local development.
- OpenAPI contract at `api/openapi.yaml`.
- Postgres configuration.
- `database/sql` Postgres connection setup through pgx.
- File-based migration runner.
- Initial V1 database schema.
- `capcom migrate up` CLI command.
- `capcom-server` can connect to Postgres through `CAPCOM_DATABASE_URL`.
- Runtime connection REST APIs.
- Gantry runtime adapter read-path health check.
- Runtime connection test endpoint.
- AES-256-GCM encrypted runtime secret storage.
- Audited secret creation and rotation APIs.
- Runtime connections persist secret references rather than credentials.
- Gantry adapter Bearer authentication resolved at request time.
- Embedded Capcom verification console served by the Go binary.
- Live runtime-neutral agent and access inspection through the selected adapter.
- Main/registered/subagent classification and live current-skill inspection.
- Gantry delegated-task ingestion as separate ephemeral subagent executions.
- Selected-agent details with enriched skill descriptions, tools, workflows, and effective access.
- Transactional manual and periodic runtime synchronization.
- Persisted agents, hierarchy, assigned skills, and effective access.
- Live, cached, and stale freshness semantics with last-known-state retention.
- Sync history and database-backed overlap protection.
- Audited, idempotent access reconciliation with read-only rejection and dry-run validation.
- Unit tests for config and API health behavior.

The next implementation slice is desired-state manifest apply followed by drift
detection against the durable runtime snapshots.

## Repository Layout

```text
cmd/
  capcom-server/        # API server entrypoint
  capcom/               # CLI entrypoint
internal/
  adapters/runtime/     # Runtime-neutral adapter interface
  api/                  # HTTP router, handlers, and embedded console
  config/               # Environment config
  domain/               # Runtime-neutral Capcom domain types
  store/                # Postgres connection, migrations, repositories
migrations/             # SQL migrations
api/
  openapi.yaml          # Current REST API contract
docs/
  v1/                   # V1 architecture and implementation docs
  Architecture/         # Architecture diagram assets
```

## Run Locally

For local development, copy the example environment file once:

```powershell
Copy-Item .env.example .env
```

Then run the server:

```powershell
make run
```

Default server address:

```text
:8080
```

Open the verification console at `http://127.0.0.1:8080/`. Enter the value of
`CAPCOM_ADMIN_TOKEN` in the connection dialog; the token is kept in session
storage and is cleared when the browser tab closes.

The Agents view separates durable Gantry agents from ephemeral subagent
executions. Gantry must have emitted a `delegated_agent` task lifecycle event
inside a job run before the lower table contains rows; registered agents are
never relabeled as subagents.

Health check:

```powershell
Invoke-RestMethod http://127.0.0.1:8080/healthz
```

Expected response:

```json
{
  "status": "ok",
  "service": "capcom",
  "version": "dev"
}
```

Run with the local development database:

```powershell
make migrate-up
make run
```

## Configuration

| Variable | Default | Description |
|---|---:|---|
| `CAPCOM_HTTP_ADDR` | `:8080` | HTTP listen address |
| `CAPCOM_HTTP_READ_HEADER_TIMEOUT` | `5s` | HTTP read-header timeout |
| `CAPCOM_HTTP_SHUTDOWN_TIMEOUT` | `10s` | Graceful shutdown timeout |
| `CAPCOM_SERVICE_VERSION` | `dev` | Version reported by health responses |
| `CAPCOM_LOG_LEVEL` | `info` | One of `debug`, `info`, `warn`, `error` |
| `CAPCOM_ADMIN_TOKEN` | empty | Bearer token required by every API except `GET /healthz` |
| `CAPCOM_SECRET_KEY` | empty | Base64-encoded 32-byte AES key; required when Postgres is configured |
| `CAPCOM_DATABASE_URL` | empty | Postgres connection string, required for migrations |
| `CAPCOM_DATABASE_MAX_OPEN_CONNS` | `10` | Maximum open Postgres connections |
| `CAPCOM_DATABASE_MAX_IDLE_CONNS` | `5` | Maximum idle Postgres connections |
| `CAPCOM_DATABASE_CONN_MAX_LIFETIME` | `30m` | Maximum Postgres connection lifetime |
| `CAPCOM_SYNC_WORKER_ENABLED` | `true` | Enables periodic runtime synchronization |
| `CAPCOM_SYNC_WORKER_TICK` | `5s` | Scheduler scan interval |
| `CAPCOM_SYNC_MAX_CONCURRENCY` | `4` | Maximum concurrent runtime syncs |
| `CAPCOM_SYNC_REQUEST_TIMEOUT` | `30s` | Timeout for one scheduled sync |
| `CAPCOM_SYNC_MISSING_THRESHOLD` | `3` | Successful absences before stale marking |

## Development Commands

```powershell
make test
make vet
make tidy
make migrate-up
make run
```

Equivalent direct Go commands:

```powershell
go test ./...
go vet ./...
go mod tidy
go run ./cmd/capcom migrate up
go run ./cmd/capcom-server
```

Repository integration tests run only when `CAPCOM_TEST_DATABASE_URL` points to
a database whose name contains `test`. They clean up records they create and
refuse the development `capcom` database to prevent fixture connections from
appearing in the console.

## Database

Capcom uses Postgres for V1 persistence. Local development values live in `.env`.

For a new checkout:

```powershell
Copy-Item .env.example .env
make migrate-up
```

Current local development database:

```text
Container: pulse-pg
Host port: 5433
Database: capcom
User: capcom
Password: capcom
URL: postgres://capcom:capcom@127.0.0.1:5433/capcom?sslmode=disable
```

For this workspace, `.env.example` already points at the local development database:

`postgres://capcom:capcom@127.0.0.1:5433/capcom?sslmode=disable`

The initial schema creates:

- `runtime_connections`
- `agents`
- `agent_runtime_bindings`
- `access_desired_state`
- `access_actual_state`
- `drift_findings`
- `control_actions`
- `audit_events`
- `secrets`
- `schema_migrations`

## Runtime Connection API

Generate a local Capcom encryption key once and add it to `.env`:

```powershell
$bytes = New-Object byte[] 32
$rng = [Security.Cryptography.RandomNumberGenerator]::Create()
try { $rng.GetBytes($bytes) } finally { $rng.Dispose() }
$key = [Convert]::ToBase64String($bytes)
$content = Get-Content .env | Where-Object { $_ -notmatch '^CAPCOM_SECRET_KEY=' }
Set-Content .env ($content + "CAPCOM_SECRET_KEY=$key")
```

Set a separate high-entropy `CAPCOM_ADMIN_TOKEN` in `.env`, then use it for API
requests:

```powershell
$capcomAdminToken = "<value from CAPCOM_ADMIN_TOKEN>"
$capcomHeaders = @{ Authorization = "Bearer $capcomAdminToken" }
```

Store the Gantry Control API token. The response contains metadata only:

```powershell
Invoke-RestMethod `
  -Method Post `
  -Uri http://127.0.0.1:8080/v1/secrets `
  -ContentType "application/json" `
  -Headers $capcomHeaders `
  -Body (@{
    name = "gantry-control-api-key"
    value = $gantryToken
    actor = "local-dev"
    reason = "configure Gantry authentication"
  } | ConvertTo-Json)
```

Create a Gantry runtime connection:

```powershell
Invoke-RestMethod `
  -Method Post `
  -Uri http://127.0.0.1:8080/v1/runtime-connections `
  -ContentType "application/json" `
  -Headers $capcomHeaders `
  -Body '{
    "name": "local-gantry",
    "runtime_type": "gantry",
    "mode": "read_only",
    "endpoint": "http://127.0.0.1:8787",
    "auth_ref": "gantry-control-api-key",
    "actor": "local-dev",
    "reason": "initial Gantry connection"
  }'
```

List runtime connections:

```powershell
Invoke-RestMethod -Headers $capcomHeaders http://127.0.0.1:8080/v1/runtime-connections
```

Get one runtime connection:

```powershell
Invoke-RestMethod -Headers $capcomHeaders http://127.0.0.1:8080/v1/runtime-connections/<runtime-id>
```

Test a runtime connection through its adapter:

```powershell
Invoke-RestMethod `
  -Method Post `
  -Uri http://127.0.0.1:8080/v1/runtime-connections/<runtime-id>/test `
  -Headers $capcomHeaders
```

Read live agents through the configured adapter:

```powershell
Invoke-RestMethod `
  -Headers $capcomHeaders `
  http://127.0.0.1:8080/v1/runtime-connections/<runtime-id>/agents
```

Read one live agent's canonical access document:

```powershell
Invoke-RestMethod `
  -Headers $capcomHeaders `
  http://127.0.0.1:8080/v1/runtime-connections/<runtime-id>/agents/<runtime-agent-id>/access
```

Read one live agent's current skill bindings:

```powershell
Invoke-RestMethod `
  -Headers $capcomHeaders `
  http://127.0.0.1:8080/v1/runtime-connections/<runtime-id>/agents/<runtime-agent-id>/skills
```

These nested agent endpoints are inspection reads. The upcoming sync loop will
persist normalized agents and access state for drift detection.

Runtime connection APIs return `503 database_not_configured` when the server is started without `CAPCOM_DATABASE_URL`.

Rotate a stored credential without changing runtime connections:

```powershell
Invoke-RestMethod `
  -Method Put `
  -Uri http://127.0.0.1:8080/v1/secrets/gantry-control-api-key `
  -ContentType "application/json" `
  -Headers $capcomHeaders `
  -Body (@{
    value = $newGantryToken
    actor = "local-dev"
    reason = "scheduled credential rotation"
  } | ConvertTo-Json)
```

Capcom never returns secret values. Keep `CAPCOM_SECRET_KEY` stable: changing it
without re-encrypting stored secrets makes existing references undecryptable.

For Gantry runtime connections, `/test` calls Gantry `GET /v1/health` and returns adapter capabilities:

```json
{
  "status": "active",
  "message": "gantry health check succeeded",
  "capabilities": {
    "read_agents": true,
    "read_agent_access": true,
    "replace_agent_access": false
  }
}
```

## API Contract

The current REST contract is maintained in [api/openapi.yaml](C:/Users/caw-dev/Desktop/capcom/api/openapi.yaml).

Use this file as the source of truth for Postman imports, generated clients, and future server-side validation. Do not hand-maintain a separate Postman collection as the primary contract.

## Documentation

The V1 source of truth is [docs/v1/README.md](C:/Users/caw-dev/Desktop/capcom/docs/v1/README.md).

Important docs:

- [Execution implementation plan](C:/Users/caw-dev/Desktop/capcom/docs/v1/13-execution-implementation-plan.md)
- [Development rules](C:/Users/caw-dev/Desktop/capcom/docs/v1/11-development-rules.md)
- [Go coding rulebook](C:/Users/caw-dev/Desktop/capcom/docs/v1/12-go-coding-rulebook.md)
- [Architecture overview](C:/Users/caw-dev/Desktop/capcom/docs/v1/01-architecture-overview.md)
- [Gantry adapter contract](C:/Users/caw-dev/Desktop/capcom/docs/v1/03-gantry-adapter-contract.md)

## Implementation Rule

Update this README whenever the project gains a new runnable command, package, endpoint, configuration variable, setup requirement, or major implementation milestone.

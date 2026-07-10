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
- Postgres configuration.
- `database/sql` Postgres connection setup through pgx.
- File-based migration runner.
- Initial V1 database schema.
- `capcom migrate up` CLI command.
- `capcom-server` can connect to Postgres through `CAPCOM_DATABASE_URL`.
- Runtime connection REST APIs.
- Gantry runtime adapter read-path health check.
- Runtime connection test endpoint.
- Unit tests for config and API health behavior.

The next implementation slice is the manual sync loop that imports Gantry agents and access state.

## Repository Layout

```text
cmd/
  capcom-server/        # API server entrypoint
  capcom/               # CLI entrypoint
internal/
  adapters/runtime/     # Runtime-neutral adapter interface
  api/                  # HTTP router and handlers
  config/               # Environment config
  domain/               # Runtime-neutral Capcom domain types
  store/                # Postgres connection, migrations, repositories
migrations/             # SQL migrations
docs/
  v1/                   # V1 architecture and implementation docs
  Architecture/         # Architecture diagram assets
```

## Run Locally

```powershell
go run ./cmd/capcom-server
```

Default server address:

```text
:8080
```

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
$env:CAPCOM_DATABASE_URL = "postgres://capcom:capcom@localhost:5433/capcom?sslmode=disable"
go run ./cmd/capcom-server
```

## Configuration

| Variable | Default | Description |
|---|---:|---|
| `CAPCOM_HTTP_ADDR` | `:8080` | HTTP listen address |
| `CAPCOM_HTTP_READ_HEADER_TIMEOUT` | `5s` | HTTP read-header timeout |
| `CAPCOM_HTTP_SHUTDOWN_TIMEOUT` | `10s` | Graceful shutdown timeout |
| `CAPCOM_SERVICE_VERSION` | `dev` | Version reported by health responses |
| `CAPCOM_LOG_LEVEL` | `info` | One of `debug`, `info`, `warn`, `error` |
| `CAPCOM_DATABASE_URL` | empty | Postgres connection string, required for migrations |
| `CAPCOM_DATABASE_MAX_OPEN_CONNS` | `10` | Maximum open Postgres connections |
| `CAPCOM_DATABASE_MAX_IDLE_CONNS` | `5` | Maximum idle Postgres connections |
| `CAPCOM_DATABASE_CONN_MAX_LIFETIME` | `30m` | Maximum Postgres connection lifetime |

## Development Commands

```powershell
go test ./...
go vet ./...
go run ./cmd/capcom-server
go run ./cmd/capcom migrate up
```

The `Makefile` provides the same commands for environments with `make`:

```powershell
make test
make vet
make run
make migrate-up
```

## Database

Capcom uses Postgres for V1 persistence. Set `CAPCOM_DATABASE_URL` before running migrations:

```powershell
$env:CAPCOM_DATABASE_URL = "postgres://capcom:capcom@localhost:5432/capcom?sslmode=disable"
go run ./cmd/capcom migrate up
```

Current local development database:

```text
Container: pulse-pg
Host port: 5433
Database: capcom
User: capcom
Password: capcom
URL: postgres://capcom:capcom@localhost:5433/capcom?sslmode=disable
```

For this workspace, run:

```powershell
$env:CAPCOM_DATABASE_URL = "postgres://capcom:capcom@localhost:5433/capcom?sslmode=disable"
go run ./cmd/capcom migrate up
```

The initial schema creates:

- `runtime_connections`
- `agents`
- `agent_runtime_bindings`
- `access_desired_state`
- `access_actual_state`
- `drift_findings`
- `control_actions`
- `audit_events`
- `schema_migrations`

## Runtime Connection API

Create a Gantry runtime connection:

```powershell
Invoke-RestMethod `
  -Method Post `
  -Uri http://127.0.0.1:8080/v1/runtime-connections `
  -ContentType "application/json" `
  -Body '{
    "name": "local-gantry",
    "runtime_type": "gantry",
    "mode": "read_only",
    "endpoint": "http://127.0.0.1:3000",
    "actor": "local-dev",
    "reason": "initial Gantry connection"
  }'
```

List runtime connections:

```powershell
Invoke-RestMethod http://127.0.0.1:8080/v1/runtime-connections
```

Get one runtime connection:

```powershell
Invoke-RestMethod http://127.0.0.1:8080/v1/runtime-connections/<runtime-id>
```

Test a runtime connection through its adapter:

```powershell
Invoke-RestMethod `
  -Method Post `
  -Uri http://127.0.0.1:8080/v1/runtime-connections/<runtime-id>/test
```

Runtime connection APIs return `503 database_not_configured` when the server is started without `CAPCOM_DATABASE_URL`.

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

## Documentation

The V1 source of truth is [docs/v1/README.md](C:/Users/caw-dev/Desktop/capcom/docs/v1/README.md).

Important docs:

- [Execution implementation plan](C:/Users/caw-dev/Desktop/capcom/docs/v1/13-execution-implementation-plan.md)
- [Local dev and API tooling plan](C:/Users/caw-dev/Desktop/capcom/docs/v1/14-local-dev-api-tooling-plan.md)
- [Development rules](C:/Users/caw-dev/Desktop/capcom/docs/v1/11-development-rules.md)
- [Go coding rulebook](C:/Users/caw-dev/Desktop/capcom/docs/v1/12-go-coding-rulebook.md)
- [Architecture overview](C:/Users/caw-dev/Desktop/capcom/docs/v1/01-architecture-overview.md)
- [Gantry adapter contract](C:/Users/caw-dev/Desktop/capcom/docs/v1/03-gantry-adapter-contract.md)

## Implementation Rule

Update this README whenever the project gains a new runnable command, package, endpoint, configuration variable, setup requirement, or major implementation milestone.

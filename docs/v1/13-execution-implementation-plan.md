# 13 - Execution Implementation Plan

## Purpose

This document turns the V1 architecture pack into an implementation sequence that agents can execute without re-litigating the product shape.

V1 should be built as a thin but complete control loop:

1. Start Capcom.
2. Connect to Gantry.
3. Import agents and access state.
4. Store desired state.
5. Detect drift.
6. Execute one explicit audited control action.
7. Show the result through API, CLI, and later a minimal dashboard.

## Execution Rules

- Follow `11-development-rules.md` and `12-go-coding-rulebook.md`.
- Keep Gantry code behind the runtime adapter boundary.
- Do not let API handlers call Gantry directly.
- Do not introduce a UI before the backend loop works through API and CLI.
- Every phase must leave the repo runnable and testable.
- Every mutation must write audit data from the first implementation, not as cleanup.
- Prefer small vertical slices over large horizontal subsystems.

## Target Repository Shape

```text
cmd/
  capcom-server/
    main.go
  capcom/
    main.go
internal/
  api/
  audit/
  auth/
  config/
  controls/
  domain/
  drift/
  manifests/
  secrets/
  services/
  store/
  workers/
  adapters/
    runtime/
    gantry/
migrations/
docs/
  v1/
```

## Phase 0 - Backend Skeleton

Goal: create a minimal Go backend that starts, logs, reads config, and exposes health.

Tasks:

- Create `go.mod`.
- Add `cmd/capcom-server/main.go`.
- Add `internal/config` for environment-driven config.
- Add `internal/api` with router setup.
- Add `GET /healthz`.
- Add structured JSON logging.
- Add graceful shutdown.
- Add `Makefile` or documented commands for test/build/run.

Acceptance:

- `go test ./...` passes.
- `go vet ./...` passes.
- `go run ./cmd/capcom-server` starts locally.
- `GET /healthz` returns `200` with a stable JSON body.

Suggested first health response:

```json
{
  "status": "ok",
  "service": "capcom",
  "version": "dev"
}
```

## Phase 1 - Database And Migrations

Goal: add Postgres persistence and the initial schema required for the control loop.

Tasks:

- Add DB config.
- Add DB connection setup with sane pool settings.
- Add migration runner or migration command.
- Create initial migrations for:
  - `runtime_connections`
  - `agents`
  - `agent_runtime_bindings`
  - `access_desired_state`
  - `access_actual_state`
  - `drift_findings`
  - `control_actions`
  - `audit_events`
- Add repository package under `internal/store`.
- Add repository tests for core CRUD paths.

Acceptance:

- Migrations apply from an empty database.
- Migrations are repeatable in local development.
- Repository tests cover create/list/get for runtime connections and agents.
- No migration depends on Gantry-specific concepts except generic runtime metadata.

## Phase 2 - Domain And Runtime Adapter Boundary

Goal: define Capcom's internal model and keep runtime-specific behavior isolated.

Tasks:

- Add `internal/domain` types:
  - `RuntimeConnection`
  - `RuntimeKind`
  - `Agent`
  - `AgentBinding`
  - `AccessDocument`
  - `DesiredState`
  - `ActualState`
  - `DriftFinding`
  - `ControlAction`
  - `AuditEvent`
- Add `internal/adapters/runtime` interfaces.
- Add compile-time adapter conformance checks.
- Add service layer contracts in `internal/services`.

Core interface shape:

```go
type RuntimeAdapter interface {
    Kind() domain.RuntimeKind
    Check(ctx context.Context, conn domain.RuntimeConnection) (*RuntimeCheck, error)
    ListAgents(ctx context.Context, conn domain.RuntimeConnection) ([]domain.AgentSnapshot, error)
    GetAgentAccess(ctx context.Context, conn domain.RuntimeConnection, runtimeAgentID string) (*domain.AccessDocument, error)
    ReplaceAgentAccess(ctx context.Context, conn domain.RuntimeConnection, runtimeAgentID string, access domain.AccessDocument) (*domain.AccessDocument, error)
}
```

Acceptance:

- Domain package does not import Gantry adapter code.
- Service package depends on runtime interfaces, not concrete Gantry clients.
- Gantry-specific fields live in metadata maps or adapter DTOs, not core domain tables unless they are generic.

## Phase 3 - Runtime Connections API

Goal: allow Capcom to register a Gantry runtime and validate connectivity.

Tasks:

- Add runtime connection create/list/get APIs.
- Add connection test endpoint.
- Add secret handling for runtime tokens or credentials.
- Add read-only vs control-enabled mode.
- Add audit entry for runtime connection creation and credential rotation.
- Add API tests.

Endpoints:

- `POST /v1/runtime-connections`
- `GET /v1/runtime-connections`
- `GET /v1/runtime-connections/{id}`
- `POST /v1/runtime-connections/{id}/test`

Acceptance:

- Invalid Gantry URL or credentials return a useful error.
- Valid Gantry connection becomes active.
- Secrets are not returned by read APIs.
- Runtime mode is persisted.

## Phase 4 - Gantry Adapter Read Path

Goal: import live Gantry state without mutation.

Tasks:

- Add `internal/adapters/gantry` HTTP client.
- Implement `Check`.
- Implement `ListAgents` using `GET /v1/agents`.
- Implement `GetAgentAccess` using `GET /v1/agents/{agentId}/access`.
- Normalize Gantry DTOs into Capcom domain snapshots.
- Add fixture-based contract tests.
- Add timeout, retry, and error classification.

Acceptance:

- Adapter can import agents from Gantry.
- Adapter can import access state from Gantry.
- Gantry outage marks runtime degraded but preserves last known state.
- Tests use recorded or handcrafted Gantry fixture JSON.

## Phase 5 - Manual Sync Loop

Status: implemented, including periodic polling, persisted skills, freshness,
sync history, and last-known-state preservation.

Goal: persist actual runtime state and expose sync results.

Tasks:

- Add sync service.
- Add manual sync endpoint.
- Persist agents and runtime bindings.
- Persist agent kind and parent relationships when the adapter provides them.
- Persist current skill bindings separately from effective access.
- Persist actual access documents.
- Write sync audit events.
- Store runtime degraded status on failures.

Endpoint:

- `POST /v1/runtime-connections/{id}/sync`

Acceptance:

- Manual sync imports Gantry agents.
- Repeated sync is idempotent.
- Removed or missing agents are marked stale instead of hard-deleted.
- Main, registered, and subagent identities remain distinguishable.
- Unsupported hierarchy is reported as unavailable, not as an empty child list.
- Sync failure does not erase last known actual state.

## Phase 6 - Desired State And CLI Apply

Goal: store approved desired state through manifests.

Tasks:

- Add YAML manifest parser.
- Implement `RuntimeConnection` manifest validation.
- Implement `Agent` manifest validation.
- Add apply service.
- Add `capcom apply -f <file>` CLI command.
- Write audit events for desired-state changes.

Acceptance:

- Valid manifest stores desired state.
- Invalid manifest reports field-level errors.
- Applying the same manifest twice is idempotent.
- Desired state stores actor, source, version, and timestamp.

## Phase 7 - Drift Detection

Goal: compare desired state against actual Gantry state.

Tasks:

- Implement drift comparator.
- Compare desired agent enabled status vs actual.
- Compare desired access selections vs actual.
- Write open drift findings.
- Resolve drift when desired and actual match.
- Add drift list/get APIs.
- Add CLI `capcom diff agent`.

Endpoints:

- `GET /v1/drift-findings`
- `GET /v1/agents/{id}/drift`

Acceptance:

- Extra actual capability creates an open drift finding.
- Missing actual capability creates an open drift finding.
- Re-sync after reconciliation resolves the finding.
- Drift records include expected, actual, severity, and first/last observed timestamps.

## Phase 8 - Safe Control Action

Status: backend and console flow implemented for access reconciliation. Live
mutation verification requires a `control_enabled` runtime connection; the local
read-only Gantry connection is covered by the audited rejection path.

Goal: execute one audited runtime mutation through Gantry.

Tasks:

- Add control action service.
- Add dry-run validation.
- Add idempotency key handling.
- Add `ReplaceAgentAccess` Gantry implementation using `PUT /v1/agents/{agentId}/access`.
- Add action status lifecycle:
  - `queued`
  - `running`
  - `succeeded`
  - `failed`
  - `rejected`
- Add before/after audit entries.
- Add read-only runtime rejection.

Endpoint:

- `POST /v1/agents/{id}/actions/reconcile-access`

Acceptance:

- Read-only runtime rejects mutation.
- Control-enabled runtime updates Gantry access.
- Every action stores actor, reason, before, after, result, and runtime response metadata.
- Failed action does not mark drift resolved.
- Successful action followed by sync resolves drift.

## Phase 9 - Minimal API Completion

Goal: expose enough API for a dashboard and demo without adding UI yet.

Tasks:

- Add fleet list endpoint.
- Add agent detail endpoint.
- Add agent access endpoint.
- Add audit list endpoint.
- Add OpenAPI or documented request/response examples.
- Add API smoke tests.

Endpoints:

- `GET /v1/agents`
- `GET /v1/agents/{id}`
- `GET /v1/agents/{id}/access`
- `GET /v1/audit-events`
- `GET /v1/control-actions`
- `GET /v1/control-actions/{id}`

Acceptance:

- API can power the full V1 demo through curl or CLI.
- Responses do not expose secrets.
- Runtime degraded state is visible in agent and runtime responses.

## Phase 10 - Minimal Dashboard

Goal: provide an operator UI once the backend loop is proven.

Screens:

- Runtime Connections
- Fleet View
- Agent Detail
- Drift Findings
- Audit Log

Acceptance:

- Operator can run the V1 demo from the dashboard.
- Dashboard shows runtime degraded state clearly.
- Dashboard does not allow mutation without reason text.
- Dashboard exposes dry-run result before reconcile.

## Phase 11 - Demo Hardening

Goal: make V1 repeatable.

Tasks:

- Add demo manifests.
- Add local setup instructions.
- Add smoke test script.
- Add fixture data.
- Update docs with implementation discoveries.
- Run clean setup twice.

Acceptance:

- Full demo passes twice from a clean local environment.
- Known limitations are documented.
- Post-V1 backlog is updated.

## Recommended First PR

Scope:

- `go.mod`
- `cmd/capcom-server/main.go`
- `internal/config`
- `internal/api`
- `internal/domain`
- `internal/adapters/runtime`
- `GET /healthz`
- tests for config and health

Do not include:

- database schema
- Gantry adapter
- UI
- CLI
- drift logic

Acceptance:

- `go test ./...`
- `go vet ./...`
- `go run ./cmd/capcom-server`
- health endpoint responds locally

## Work Breakdown By Agent

If multiple agents work in parallel, split by boundaries:

- Agent A: backend scaffold, config, HTTP server, logging.
- Agent B: database schema and repositories.
- Agent C: Gantry adapter fixtures and client.
- Agent D: manifest parser and CLI.
- Agent E: drift comparator and tests.

Rules for parallel work:

- No agent changes another agent's package without checking current files first.
- Shared domain changes must be small and reviewed against `02-domain-model-database.md`.
- Adapter changes must not alter service contracts without updating docs.
- Generated files and temporary render/build artifacts must be removed before handoff.

## Risk Register

| Risk | Impact | Mitigation |
|---|---|---|
| Gantry API shape changes | Adapter breaks | Keep fixture-based contract tests and isolate DTOs |
| Desired vs actual state becomes Gantry-specific | Future adapters become hard | Keep domain model runtime-neutral |
| Control actions mutate without audit | Trust failure | Audit before implementing mutation success path |
| Dashboard starts too early | Backend architecture gets distorted | Build API/CLI loop first |
| Runtime outage deletes state | Operators lose last known state | Mark degraded and preserve state |
| Secrets leak through APIs or logs | Security incident | Redact at config, API DTO, and logger boundaries |

## Done Criteria For The Whole V1

- Gantry runtime connection can be created and tested.
- Gantry agents and access state can be synced.
- Desired state can be applied from YAML.
- Drift can be detected and listed.
- One reconcile action can safely update Gantry.
- Read-only mode blocks mutation.
- Every mutation and desired-state change has an audit trail.
- Local demo is repeatable from clean setup.

# 11 - Development Rules

## Goal

Set rules before implementation starts so Capcom V1 stays focused, testable, and runtime-agnostic.

## Architectural Rules

### Gantry Is An Adapter

Gantry is the first runtime adapter. It must not define the Capcom domain model.

Allowed:

```text
internal/adapters/gantry -> imports Gantry-shaped structs
internal/services        -> imports Capcom domain structs
internal/domain          -> imports no adapter packages
```

Forbidden:

```text
internal/domain imports internal/adapters/gantry
internal/services stores raw Gantry responses as first-class domain objects
control actions call Gantry directly instead of going through RuntimeAdapter
```

### Adapter Capability Levels

Every runtime adapter must declare what it can do:

```text
L0 manual registry
L1 telemetry
L2 inventory
L3 drift capable
L4 control capable
L5 enforcement capable
```

Gantry V1 target:

```text
L4 control capable for health, inventory, access import, selected drift, disable/enable, restrict/replace access.
```

The UI and API should expose support level so Capcom does not overpromise control for weaker future adapters.

### Core Interfaces

The runtime adapter interface should stay small:

```go
type RuntimeAdapter interface {
    Capabilities() AdapterCapabilities
    Health(ctx context.Context) (RuntimeHealth, error)
    Doctor(ctx context.Context) (RuntimeDoctor, error)
    ListAgents(ctx context.Context) ([]RuntimeAgent, error)
    GetAgentAdmin(ctx context.Context, externalAgentID string) (RuntimeAgentAdmin, error)
    GetAgentAccess(ctx context.Context, externalAgentID string) (RuntimeAgentAccess, error)
    ListInventory(ctx context.Context) (RuntimeInventory, error)
    ListEvents(ctx context.Context, cursor RuntimeCursor) ([]RuntimeEvent, RuntimeCursor, error)
    ExecuteControlAction(ctx context.Context, action RuntimeControlAction) (RuntimeControlResult, error)
}
```

Do not add runtime-specific methods to this interface unless at least two adapter types need them or the method is represented as a generic capability.

## Backend Rules

Go-specific implementation must follow `12-go-coding-rulebook.md`.

### Handler-Service-Store Split

Handlers:

- authenticate
- parse request
- validate request shape
- call service
- write response

Services:

- own business flow
- validate runtime mode and adapter capability
- call repositories and adapters
- call audit service for mutations

Repositories:

- persist and query data
- do not contain business decisions
- do not call adapters

### Error Shape

All API errors use:

```json
{
  "error": {
    "code": "INVALID_REQUEST",
    "message": "reason is required",
    "details": {}
  }
}
```

Do not leak raw runtime secrets, credentials, or Authorization headers in errors.

### Mutations

Every mutation must have:

- actor
- target
- action
- reason when runtime/control related
- pre-state where available
- post-state or error
- audit log

Runtime mutation without audit is a bug.

## Database Rules

Use typed columns for:

- ids
- runtime type
- runtime mode
- status
- severity
- timestamps
- actor
- action type

Use JSONB for:

- Gantry `/access` documents
- runtime admin payloads
- raw payload snapshots
- desired access documents
- before/after audit payloads

Migration rules:

- migrations must be deterministic
- no destructive migration without explicit backup/transition plan
- indexes are required for list/detail pages before UI depends on them

## Security Rules

- No plaintext runtime keys in manifests.
- No plaintext runtime keys in logs.
- Store runtime keys as encrypted secrets or external secret refs.
- Read-only runtime connections cannot execute control actions.
- Gantry keys should use the smallest practical scope set.
- Audit logs are append-only.

## Drift Rules

V1 drift starts narrow:

- desired status vs actual status
- desired selected capabilities vs actual selected capabilities

Do not add behavioral/eval drift until the first Gantry control loop works.

Drift records are never deleted. They move through statuses:

```text
open
acknowledged
resolved
ignored
```

## Control Action Rules

V1 actions:

- `disable_agent`
- `enable_agent`
- `replace_access`
- `restrict_capability`

Rules:

- validate adapter supports action
- validate runtime mode is `control_enabled`
- validate target still exists
- validate actual state is available
- write pre-action audit
- call adapter
- write post-action audit
- mark actual state stale or trigger sync

No autonomous remediation in V1.

## Testing Rules

Minimum tests before merging a feature:

- unit tests for pure logic
- adapter normalization tests with fixtures
- repository tests for new tables/queries
- API tests for new endpoints
- failure-path tests for runtime unavailable, invalid credentials, and rejected mutation

Live Gantry tests should be optional and env-gated.

## Frontend Rules

Dashboard V1 should be operational, not marketing.

Required screens:

- Runtime Connections
- Fleet View
- Agent Detail
- Drift
- Audit Log

Rules:

- always show runtime status
- always show last sync status/time
- never hide degraded runtime state
- require confirmation and reason for control actions
- show unsupported actions disabled with explanation

## Documentation Rules

Update docs in the same change when any of these changes:

- API route
- database schema
- manifest shape
- adapter capability
- control action behavior
- drift behavior
- security assumption

`docs/v1/` remains canonical until V1 is complete.

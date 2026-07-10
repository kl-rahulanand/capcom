# 06 - Drift Detection Design

## Goal

Detect when live Gantry state differs from Capcom approved desired state.

## V1 Drift Scope

P0 drift:

- agent status
- selected capabilities in `/v1/agents/{agentId}/access`

P1 within V1:

- sources.skills
- sources.mcpServers
- sources.tools
- bound conversations and approvers

Out of V1:

- behavioral drift from traces
- quality/eval drift
- policy-as-code enforcement
- autonomous remediation

## Comparison Rules

### Status Drift

Desired:

```yaml
desiredState:
  status: active
```

Actual:

```json
{ "runtimeStatus": "disabled" }
```

Result:

```text
drift_type=status
field_path=desiredState.status
severity=warning
```

### Capability Drift

Desired selections are compared as a set of `(id, version)`.

Extra actual selection:

```text
desired: [browser.use]
actual: [browser.use, production-db.write]
```

Creates:

```text
drift_type=capability
field_path=access.selections[production-db.write]
severity=critical if high-risk or write/system capability
```

Missing actual selection:

```text
desired: [browser.use, slack.send]
actual: [browser.use]
```

Creates warning by default.

## Severity

| Drift | Severity |
|---|---|
| Extra production/system/write capability on high/critical agent | critical |
| Extra selected capability | warning |
| Missing selected capability | warning |
| Status mismatch | warning |
| Metadata-only mismatch | info |
| Unknown runtime field | info |

## Drift Lifecycle

```text
open -> acknowledged -> resolved
open -> ignored
ignored -> open when desired/actual changes materially
```

Rules:

- If the same drift appears in a later sync, update `last_seen_at`.
- If drift no longer appears, mark `resolved`.
- Do not delete drift records.
- Store desired and actual snapshots for audit/debugging.

## Drift Keys

Stable key:

```text
agent_id + drift_type + field_path
```

This avoids duplicate open records for the same mismatch.

## Trigger Points

Run drift detection after:

- runtime sync
- manifest apply
- control action completion
- manual `POST /v1/runtime-connections/{id}/sync`

## V1 UI Behavior

Fleet view:

- show open drift count per agent
- show highest severity

Agent detail:

- desired vs actual access
- open drift list
- suggested action where deterministic

Drift page:

- filter by severity, runtime, agent, status


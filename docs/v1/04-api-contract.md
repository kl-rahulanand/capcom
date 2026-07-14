# 04 - Capcom REST API Contract

## API Principles

- API is the canonical surface.
- Dashboard and CLI both call the API.
- Every mutation requires an actor.
- Runtime mutations require a reason.
- Responses should include stable Capcom IDs and runtime external IDs where relevant.

## Auth

V1 uses a single admin token:

```text
Authorization: Bearer <CAPCOM_ADMIN_TOKEN>
```

All API endpoints require auth except `GET /healthz`. The static console shell at
`GET /` and `/assets/*` is public, but it contains no control-plane data. The
console sends the admin token for every protected data request and retains it in
browser session storage only.

## System

```text
GET /healthz
GET /readyz
```

## Runtime Connections

```text
POST /v1/runtime-connections
GET /v1/runtime-connections
GET /v1/runtime-connections/{id}
PATCH /v1/runtime-connections/{id}
POST /v1/runtime-connections/{id}/test
POST /v1/runtime-connections/{id}/sync
GET /v1/runtime-connections/{id}/sync-runs
GET /v1/runtime-connections/{id}/agents
GET /v1/runtime-connections/{id}/agents/{runtimeAgentId}/access
GET /v1/runtime-connections/{id}/agents/{runtimeAgentId}/skills
```

Runtime connection creation uses `auth_ref`, which must name a previously
stored secret. Inline runtime credentials are rejected.

The nested runtime-agent endpoints are live adapter inspection reads for
diagnostics. The console uses the persisted fleet endpoints after a successful
sync, so a runtime outage preserves the last known state.

Manual sync requires `actor` and `reason`, returns its completed sync run, and
returns `409` when another sync holds the runtime lock. Failed collection marks
the runtime degraded without replacing imported state.

The runtime-agent skills response is enriched from both the runtime's binding
and catalog views. It identifies assigned skills and explains each skill through
its description, source, tool IDs, workflow references, and permission metadata.

## Secrets

```text
POST /v1/secrets
PUT /v1/secrets/{name}
```

Secret responses contain metadata only. There is no API that returns plaintext,
ciphertext, or the configured encryption key. Create and rotate requests require
`actor` and `reason`.

Create request:

```json
{
  "name": "gantry-local",
  "runtimeType": "gantry",
  "mode": "control_enabled",
  "endpoint": {
    "kind": "base_url",
    "value": "http://127.0.0.1:8787"
  },
  "auth_ref": "gantry-control-api-key",
  "actor": "admin@local",
  "reason": "connect local Gantry"
}
```

Response:

```json
{
  "id": "runtime_...",
  "name": "gantry-local",
  "runtimeType": "gantry",
  "mode": "control_enabled",
  "status": "active",
  "lastSyncAt": null,
  "createdAt": "...",
  "updatedAt": "..."
}
```

## Agents

```text
GET /v1/agents
GET /v1/agents/{id}
GET /v1/agents/{id}/desired-state
PUT /v1/agents/{id}/desired-state
GET /v1/agents/{id}/actual-state
GET /v1/agents/{id}/events
GET /v1/agents/{id}/drift
GET /v1/subagent-executions
```

`GET /v1/subagent-executions` returns ephemeral delegated execution
observations and accepts `runtime_connection_id` and `agent_id` filters. These
records contain parent run and owning agent references but are never returned
as durable agents.

Agent detail response should include:

```json
{
  "agent": {},
  "desiredState": {},
  "actualState": {},
  "openDriftCount": 1,
  "runtimeConnection": {}
}
```

## Manifests

```text
POST /v1/manifests/validate
POST /v1/manifests/apply
POST /v1/manifests/export
```

Apply request:

```json
{
  "source": "yaml",
  "content": "apiVersion: capcom.ai/v1alpha1\nkind: Agent\n...",
  "appliedBy": "admin@local"
}
```

Apply response:

```json
{
  "status": "applied",
  "resources": [
    {
      "kind": "Agent",
      "name": "access-request-agent",
      "id": "agent_..."
    }
  ]
}
```

## Drift

```text
GET /v1/drift
GET /v1/drift/{id}
POST /v1/drift/{id}/acknowledge
POST /v1/drift/{id}/ignore
```

Query parameters:

- `status`
- `severity`
- `agentId`
- `runtimeConnectionId`

## Control Actions

```text
POST /v1/control-actions
GET /v1/control-actions
GET /v1/control-actions/{id}
```

Request:

```json
{
  "targetType": "agent",
  "targetId": "agent_...",
  "actionType": "restrict_capability",
  "requestedBy": "operator@local",
  "reason": "Capability drift detected",
  "parameters": {
    "capabilityId": "browser.use"
  }
}
```

Supported V1 action types:

```text
disable_agent
enable_agent
replace_access
restrict_capability
```

Response:

```json
{
  "id": "control_action_...",
  "status": "succeeded",
  "runtimeResult": {},
  "createdAt": "...",
  "finishedAt": "..."
}
```

## Audit

```text
GET /v1/audit
GET /v1/audit/{id}
```

Query parameters:

- `actor`
- `targetType`
- `targetId`
- `action`
- `from`
- `to`

## Error Shape

```json
{
  "error": {
    "code": "INVALID_REQUEST",
    "message": "reason is required",
    "details": {}
  }
}
```

Common codes:

```text
INVALID_REQUEST
UNAUTHORIZED
FORBIDDEN
NOT_FOUND
CONFLICT
RUNTIME_UNAVAILABLE
RUNTIME_REJECTED
INTERNAL
```

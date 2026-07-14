# Capcom MVP API And Schema Contract

## Goal

Define the Capcom-owned API resources and manifest shapes for the Gantry-first MVP. This contract is intentionally runtime-agnostic, even though Gantry is the first adapter.

## Resource Model

| Resource | Purpose |
|---|---|
| `RuntimeConnection` | Stores runtime endpoint, mode, status, and secret reference |
| `Agent` | Normalized agent registry record |
| `AgentDesiredState` | Approved desired state from UI, API, CLI, or YAML |
| `AgentActualState` | Latest observed state imported from Gantry |
| `AgentEvent` | Normalized runtime event timeline |
| `DriftRecord` | Desired-vs-actual mismatch |
| `ControlAction` | Operator-requested runtime mutation |
| `AuditLog` | Immutable record of mutation, sync, and control outcomes |

## RuntimeConnection Manifest

```yaml
apiVersion: capcom.ai/v1alpha1
kind: RuntimeConnection
metadata:
  name: gantry-local
  labels:
    environment: local
spec:
  type: gantry
  mode: control_enabled
  endpoint:
    baseUrl: http://127.0.0.1:8787
    socketPath: ""
  auth:
    apiKeyRef: gantry-control-api-key
  sync:
    mode: polling
    intervalSeconds: 60
    preserveLastKnownState: true
  webhooks:
    enabled: false
```

Validation rules:

- `apiVersion` must be `capcom.ai/v1alpha1`.
- `kind` must be `RuntimeConnection`.
- `metadata.name` must be unique.
- `spec.type` must be `gantry` for MVP.
- `spec.mode` must be `read_only` or `control_enabled`.
- Exactly one of `endpoint.baseUrl` or `endpoint.socketPath` should be used for a local connection.
- `auth.apiKeyRef` must reference a stored secret, not an inline plaintext key.
- `webhooks.enabled` defaults to `false` for MVP.

## Agent Manifest

```yaml
apiVersion: capcom.ai/v1alpha1
kind: Agent
metadata:
  name: access-request-agent
  labels:
    team: it-platform
    environment: production
spec:
  owner:
    business: it-platform@company.com
    technical: ai-platform@company.com
    escalation: oncall-it-platform@company.com
  purpose: Handles employee access requests
  riskLevel: high
  runtime:
    type: gantry
    connectionRef: gantry-local
    externalAgentId: agent:main_agent
  desiredState:
    status: active
  capabilities:
    allowedTools:
      - servicenow
      - okta
      - slack
    restrictedTools:
      - production-db
    allowedSkills: []
    allowedMcpServers: []
  approvals:
    requiredFor:
      - production_access
      - privilege_escalation
      - new_tool_access
  policies:
    driftMode: observe
    maxFailureRatePercent: 5
    maxDailyCostUsd: 50
```

Validation rules:

- `kind` must be `Agent`.
- `metadata.name` must be unique inside Capcom.
- `spec.runtime.connectionRef` must reference an active or degraded runtime connection.
- `spec.runtime.externalAgentId` must map to a discovered runtime agent before actual-state comparison.
- `spec.riskLevel` must be `low`, `medium`, `high`, or `critical`.
- `spec.policies.driftMode` must be `observe`, `approval`, or `enforce`; MVP default is `observe`.
- `restrictedTools` wins over `allowedTools` if the same capability appears in both lists.
- MVP should reject unsupported manifest versions with a clear error.

## REST API

Runtime connections:

```text
POST   /v1/runtime-connections
GET    /v1/runtime-connections
GET    /v1/runtime-connections/{id}
POST   /v1/runtime-connections/{id}/test
POST   /v1/runtime-connections/{id}/sync
```

Agents:

```text
GET    /v1/agents
GET    /v1/agents/{id}
PUT    /v1/agents/{id}/desired-state
GET    /v1/agents/{id}/actual-state
GET    /v1/agents/{id}/events
GET    /v1/agents/{id}/drift
```

Manifests:

```text
POST   /v1/manifests/validate
POST   /v1/manifests/apply
POST   /v1/manifests/export
```

Control actions:

```text
POST   /v1/control-actions
GET    /v1/control-actions
GET    /v1/control-actions/{id}
```

Audit:

```text
GET    /v1/audit
```

## ControlAction Request Shape

```json
{
  "targetType": "agent",
  "targetId": "agent_123",
  "actionType": "restrict_capability",
  "requestedBy": "operator@company.com",
  "reason": "Capability drift detected during MVP demo",
  "parameters": {
    "capabilityType": "tool",
    "capabilityId": "production-db"
  }
}
```

Rules:

- `requestedBy` and `reason` are required.
- Capcom must write a pre-action audit record before calling Gantry.
- Capcom must write a post-action audit record after success or failure.
- Failed runtime mutations must be visible in `ControlAction.status`.
- Read-only runtime connections must reject mutating control actions.

## Drift Rules

MVP drift comparison includes:

- desired status vs actual runtime status
- desired selected tools vs actual agent access
- desired selected skills vs actual agent access
- desired MCP servers vs actual agent access
- expected conversations and approvers when available from Gantry

Severity:

| Drift | Default Severity |
|---|---|
| Extra production/system capability on high-risk agent | critical |
| Extra write-capable tool | warning |
| Missing expected capability | warning |
| Metadata-only mismatch | info |
| Unknown runtime field | info |

Status:

```text
open
acknowledged
resolved
ignored
```

## Audit Requirements

Every mutation must record:

- actor
- action
- target type and target id
- reason
- before state
- after state
- runtime request id if available
- runtime response or error
- timestamp

Audit logs are immutable for MVP. Corrections should be appended as new audit entries.


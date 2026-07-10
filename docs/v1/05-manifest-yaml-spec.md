# 05 - Manifest YAML Spec

## Goals

- Define approved desired state in a reviewable format.
- Allow UI/API/CLI/YAML to share one state model.
- Stay Kubernetes-compatible in style but not Kubernetes-dependent.

## RuntimeConnection

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
  auth:
    apiKeyRef: gantry-control-api-key
  sync:
    intervalSeconds: 60
    preserveLastKnownState: true
```

Validation:

- `spec.type` must be `gantry` in V1.
- `spec.mode` must be `read_only` or `control_enabled`.
- `auth.apiKeyRef` must reference stored secret material.
- Inline API keys are rejected in YAML.

## Agent

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
  access:
    sources:
      skills: []
      mcpServers: []
      tools: []
    selections:
      - id: browser.use
        version: builtin
  approvals:
    requiredFor:
      - production_access
      - privilege_escalation
  policies:
    driftMode: observe
```

## Access Model

Capcom mirrors Gantry's actual access shape:

```yaml
access:
  sources:
    skills:
      - id: skill_123
        version: "1"
    mcpServers:
      - id: mcp_123
        version: "1"
        tools:
          - issue.create
    tools:
      - id: builtin_browser
        kind: builtin
  selections:
    - id: browser.use
      version: builtin
```

Interpretation:

- `sources` are connected/provisioned inventory.
- `selections` are durable approved authority.
- V1 drift should compare `selections` first.
- Source drift can be included after the first P0 loop.

## Drift Modes

| Mode | V1 Behavior |
|---|---|
| observe | Create drift records only |
| approval | Same as observe in V1; reserved for Phase 2 workflow |
| enforce | Rejected in V1 unless feature flag is enabled |

V1 default is `observe`.

## Required Fields

Agent manifests require:

- `metadata.name`
- `spec.runtime.connectionRef`
- `spec.runtime.externalAgentId`
- `spec.owner.technical`
- `spec.purpose`
- `spec.riskLevel`
- `spec.desiredState.status`
- `spec.access.selections`

## Apply Semantics

Applying an `Agent` manifest:

1. Validates schema and referenced runtime connection.
2. Finds or creates normalized Capcom agent by runtime external id.
3. Stores desired state.
4. Writes audit log.
5. Triggers drift detection.

It does not mutate Gantry automatically in V1.


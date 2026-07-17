# 01 - Architecture Overview

## Goal

Capcom V1 is a standalone control plane for governing production agents, starting with Gantry. It owns approved desired state, imports actual runtime state, detects drift, executes safe control actions through runtime APIs, and records audit history.

## Non-Goals

- Do not run agents.
- Do not fork Gantry.
- Do not read Gantry database tables directly.
- Do not replace LangSmith, Langfuse, Phoenix, Datadog, or other observability tools.
- Do not implement autonomous remediation in V1.
- Do not require Kubernetes in V1.
- Do not require webhooks in V1.

## Component Diagram

```mermaid
flowchart TD
    UI["Next.js Dashboard"]
    CLI["capcom CLI"]
    API["Go API Server"]
    DB["Postgres"]
    Secrets["Secret Store"]
    RuntimeSvc["Runtime Service"]
    AgentSvc["Agent Service"]
    ManifestSvc["Manifest Service"]
    DriftSvc["Drift Service"]
    ControlSvc["Control Service"]
    AuditSvc["Audit Service"]
    Worker["Sync/Reconcile Worker"]
    Adapter["Gantry Adapter"]
    GantryDev["Gantry Development"]
    GantryStage["Gantry Staging"]
    GantryProd["Gantry Production"]

    UI --> API
    CLI --> API
    API --> RuntimeSvc
    API --> AgentSvc
    API --> ManifestSvc
    API --> DriftSvc
    API --> ControlSvc
    API --> AuditSvc
    RuntimeSvc --> DB
    AgentSvc --> DB
    ManifestSvc --> DB
    DriftSvc --> DB
    ControlSvc --> DB
    AuditSvc --> DB
    RuntimeSvc --> Secrets
    ControlSvc --> Adapter
    RuntimeSvc --> Adapter
    Adapter --> GantryDev
    Adapter --> GantryStage
    Adapter --> GantryProd
    Worker --> RuntimeSvc
    Worker --> DriftSvc
```

One stateless adapter implementation serves many runtime instances. Every
adapter call receives a `RuntimeConnection`; the connection selects the
endpoint and secret reference. Agent, skill, run, and execution identities are
always scoped by `runtime_connection_id`, so identical Gantry-native IDs across
instances cannot collide.

## Main Runtime Loop

```mermaid
sequenceDiagram
    participant W as Sync Worker
    participant RS as Runtime Service
    participant GA as Gantry Adapter
    participant DB as Postgres
    participant DS as Drift Service

    W->>RS: SyncRuntime(runtimeId)
    RS->>GA: Health()
    GA-->>RS: ok
    RS->>GA: ListAgents()
    GA-->>RS: runtime agents
    loop each agent
        RS->>GA: GetAgentAdmin(agentId)
        GA-->>RS: agent + conversations + capability view
        RS->>GA: GetAgentAccess(agentId)
        GA-->>RS: sources + selections + tool access
        RS->>DB: Upsert agent + actual state
    end
    RS->>DB: Mark runtime synced
    W->>DS: DetectDrift(runtimeId)
    DS->>DB: Upsert drift records
```

## Control Action Loop

```mermaid
sequenceDiagram
    participant O as Operator
    participant API as Capcom API
    participant CS as Control Service
    participant A as Audit Service
    participant GA as Gantry Adapter
    participant DB as Postgres

    O->>API: POST /v1/control-actions
    API->>CS: Validate actor, reason, runtime mode
    CS->>DB: Create pending action
    CS->>A: Write pre-action audit
    CS->>GA: Execute mutation
    GA-->>CS: Runtime result
    CS->>DB: Update action status
    CS->>A: Write post-action audit
    CS->>DB: Mark actual state stale
```

## Service Responsibilities

| Service | Responsibility |
|---|---|
| Runtime Service | Runtime connections, health checks, sync orchestration |
| Agent Service | Normalized agent registry, actual state, timelines |
| Manifest Service | Validate and apply YAML/API desired state |
| Drift Service | Compare desired and actual state, manage drift records |
| Control Service | Validate and execute safe runtime mutations |
| Audit Service | Immutable mutation and sync audit history |
| Secret Service | Store or reference runtime credentials |
| Gantry Adapter | Translate Capcom calls into Gantry Control API calls |

## Package Layout

```text
cmd/
  capcom-server/
  capcom/

internal/
  api/
  auth/
  config/
  domain/
  store/
  services/
  adapters/
    runtime/
    gantry/
  drift/
  controls/
  audit/
  secrets/
  workers/
  manifests/
```

## Boundary Rule

Gantry-specific request/response structs must stay under `internal/adapters/gantry`. Domain services should operate on Capcom domain types only. Raw Gantry payloads may be stored for debugging, but they should not become the domain model.

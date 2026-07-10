# Capcom Runtime Adapter Strategy

## Core Question

How can Capcom connect not only to Gantry, but to other agent systems in the market?

## Answer

Capcom should use a tiered adapter model. Not every agent framework exposes the same control surface, so Capcom should separate inventory, telemetry, desired state, drift, and control instead of assuming every runtime can do everything.

## Adapter Capability Levels

| Level | Name | What Capcom Can Do | Example Targets |
|---|---|---|---|
| L0 | Manual registry | Store declared agents and owners only | SaaS agents with no API |
| L1 | Telemetry adapter | Ingest traces/events and infer activity | LangSmith, Langfuse, Phoenix, Datadog, OpenTelemetry |
| L2 | Inventory adapter | Import agents, tools, prompts, skills, MCP servers, owners | Gantry, AgentOps, AGNTCY directory, internal catalogs |
| L3 | Desired-state adapter | Store approved desired state and compare actual state | Gantry, custom runtimes, Kubernetes jobs |
| L4 | Control adapter | Execute safe mutations through runtime API | Gantry now; future internal adapters |
| L5 | Enforcement adapter | Enforce policy inline before tool/action execution | Gateway, sidecar, SDK shim, MCP proxy |

Capcom's MVP with Gantry is L4 for a narrow set of actions. Most other systems will start at L1-L2.

## Required Adapter Interface

Every adapter should declare its capabilities:

```yaml
adapter:
  type: gantry
  supports:
    health: true
    inventory: true
    actualState: true
    desiredState: false
    events: true
    controlActions:
      - disable_agent
      - enable_agent
      - replace_access
      - bind_skill
      - unbind_skill
      - bind_mcp_server
      - unbind_mcp_server
      - pause_job
      - resume_job
```

Then implement:

```text
Health(ctx)
DiscoverAgents(ctx)
GetAgent(ctx, externalAgentId)
GetActualState(ctx, externalAgentId)
ListEvents(ctx, cursor)
ListInventory(ctx)
ExecuteControlAction(ctx, action)
```

Optional:

```text
ApplyDesiredState(ctx, manifest)
RegisterWebhook(ctx, callbackUrl)
ExportTelemetry(ctx, format)
```

## How To Connect Common Agent Types

### Gantry

Use native control API. This is the reference adapter.

Capabilities:

- agent inventory
- access inventory
- capability catalog
- skills and MCP bindings
- sessions/events
- jobs/runs
- safe mutations
- audit-friendly API responses

### LangGraph / LangChain Apps

Use a telemetry-first adapter unless the team also controls deployment metadata.

Connection options:

- LangSmith traces for runs, tools, errors, feedback, evaluations, dashboards
- app-side SDK shim to register agents and expose desired/actual state
- deployment metadata from Kubernetes, CI/CD, or service catalog

Capcom likely starts at L1-L2, then reaches L3-L4 only when the customer installs a Capcom SDK/sidecar.

### CrewAI / AutoGen / AG2 / Agno

Use framework instrumentation first.

Connection options:

- AgentOps integrations for traces/sessions
- Langfuse/Phoenix/OpenTelemetry spans
- Capcom SDK wrapper around crew/agent construction
- optional policy gateway around tool calls

Most of these will not expose external control APIs by default, so Capcom should not promise disable/restrict unless deployed through a Capcom wrapper.

### OpenAI Agents SDK

Use tracing and app-side registration.

Connection options:

- OpenAI Agents SDK trace export where available
- OpenTelemetry/OpenInference spans
- Capcom SDK to declare agent identity, tools, handoffs, model, owners
- gateway/proxy if the customer wants inline controls

### Claude Code / Coding Agents

Use host/workspace monitoring and explicit registration.

Connection options:

- tool logs and command traces
- repo-level agent instruction discovery
- CI/CD or IDE integration metadata
- eBPF/host-level monitoring for high-security environments

This is probably L1-L2 unless Capcom controls the execution wrapper.

### SaaS Agents

Use directory and identity integrations.

Connection options:

- Microsoft Agent 365 / Entra-style registry if available
- SaaS admin APIs
- SCIM/identity metadata
- audit logs
- OAuth app inventory

Control may be limited to quarantine, disable app, revoke token, or remove OAuth grant.

### Kubernetes / Internal Jobs

Use platform-native APIs.

Connection options:

- Kubernetes API for deployments/jobs/CRDs
- service catalog ownership metadata
- OpenTelemetry traces
- admission controller or policy controller for enforcement

This can become L4-L5 if Capcom owns a controller or admission webhook, but that should be Phase 2.

### MCP Servers

Use MCP as both inventory and control boundary.

Connection options:

- discover MCP servers and tools
- classify tool risk
- bind/unbind MCP servers to agents where the runtime supports it
- proxy MCP calls through a Capcom gateway for enforcement

This is important because MCP is becoming a common "agent reaches tools" interface.

## Adapter Implementation Order

1. Gantry native adapter.
2. OpenTelemetry/OpenInference telemetry adapter.
3. LangSmith/Langfuse/Phoenix import adapters for traces and evals.
4. Kubernetes/internal service catalog adapter for deployed agents.
5. MCP gateway/proxy adapter.
6. Identity/directory adapter for SaaS agents.

## Product Rule

Capcom should label every connected runtime by control level:

```text
Observe only
Inventory managed
Drift detected
Control enabled
Policy enforced
```

This prevents overpromising. It also makes the product more credible because users can see what Capcom can and cannot do for each runtime.


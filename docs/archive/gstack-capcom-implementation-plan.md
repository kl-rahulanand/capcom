# Capcom MVP Implementation Plan

## Phase 0 - Prep And Contracts

Goal: make the MVP buildable without ambiguity.

Tasks:

- Confirm Gantry local runtime can run after Docker/Postgres is available.
- Fetch Gantry OpenAPI from `/openapi.json` once the runtime is running.
- Generate or hand-write the first Capcom Gantry adapter client against the current SDK/API surface.
- Freeze Capcom MVP schemas for `RuntimeConnection` and `Agent`.
- Move business value metadata and topology metadata to Phase 2 in the canonical MVP doc.
- Use `gstack-capcom-local-runbook.md`, `gstack-capcom-api-schema-contract.md`, and `gstack-capcom-demo-uat-checklist.md` as the Phase 0 handoff pack.

Acceptance:

- Capcom docs identify the exact Gantry endpoints used by the MVP.
- Gantry setup instructions and blockers are documented.
- A local Gantry runtime can be reached by health/doctor, or the remaining external blocker is explicit.
- Demo/UAT acceptance criteria are documented before implementation starts.

## Phase 1 - Core Go Service

Goal: create the Capcom control-plane skeleton.

Tasks:

- Create Go module and service layout.
- Add config loading.
- Add Postgres connection.
- Add migrations for runtime connections, agents, desired state, actual state, events, drift records, control actions, and audit logs.
- Add minimal auth mode: single admin API token for local MVP.
- Add OpenAPI-first REST surface.

Acceptance:

- `capcom-server` starts locally.
- `GET /healthz` returns ok.
- Database migrations apply.
- Runtime connection records can be created and listed.

## Phase 2 - Gantry Adapter

Goal: connect Capcom to Gantry without direct DB reads.

Tasks:

- Implement Gantry adapter with health and doctor checks.
- Support socket or loopback TCP transport.
- Store Gantry API key as an encrypted secret or secret reference.
- Import Gantry agents.
- Import agent access, visible sources, capabilities, conversations, and approvers.
- Preserve last known state on runtime outage.

Acceptance:

- Valid Gantry credentials activate a connection.
- Invalid credentials are rejected.
- Imported agents appear in Capcom registry.
- Runtime outage marks the connection degraded without deleting agents.

## Phase 3 - Desired State And Basic Drift

Goal: prove Capcom is more than a dashboard.

Tasks:

- Add `Agent` manifest validation.
- Store desired selected capabilities.
- Compare desired selected capabilities with actual Gantry access.
- Create drift records for extra or missing capabilities.
- Default drift mode to `observe`.
- Show severity as info, warning, or critical.

Acceptance:

- Applying a manifest creates desired state.
- Extra actual capability creates a drift record.
- Removing drift in Gantry or updating desired state resolves the drift record.

## Phase 4 - Event Ingestion

Goal: show recent runtime activity in Capcom.

Tasks:

- Poll/list Gantry session or run events.
- Add cursor/idempotency by runtime id and external event id.
- Store normalized agent events.
- Update latest observed state timestamps.

Acceptance:

- Agent detail shows recent Gantry events.
- Duplicate event ingestion does not duplicate stored events.

## Phase 5 - Safe Control Actions

Goal: let operators take controlled action.

Tasks:

- Add control action API.
- Require actor and reason.
- Execute Gantry enable/disable if supported by current API.
- Execute access replacement/restriction through `PUT /v1/agents/:agentId/access`.
- Write pre-action and post-action audit logs.

Acceptance:

- Operator can restrict one capability from Capcom.
- Runtime result is stored.
- Audit log shows actor, reason, target, before/after, result, and timestamp.

## Phase 6 - Minimal UI And CLI

Goal: make the MVP demoable.

Dashboard screens:

- Runtime Connections
- Fleet View
- Agent Detail
- Drift
- Audit Log

CLI:

- `capcom runtime connect`
- `capcom apply -f agent.yaml`
- `capcom get agents`
- `capcom describe agent <id>`
- `capcom diff agent <id>`
- `capcom restrict capability <agent> <capability>`

Acceptance:

- Demo can be completed from UI and CLI.
- Same desired state is visible across API, CLI, and UI.

## Demo Script

1. Start Gantry.
2. Start Capcom.
3. Add Gantry runtime connection.
4. Import agents.
5. Apply desired state for one agent.
6. Show actual Gantry access.
7. Introduce one capability drift.
8. Show drift record in Capcom.
9. Restrict the capability from Capcom.
10. Show Gantry access updated.
11. Show audit log.

## Phase 2 Backlog

- Business value metadata.
- Topology metadata.
- Advanced policy drift.
- Approval drift.
- Enforce mode.
- A2A/AGNTCY mapping.
- OpenTelemetry GenAI export.
- Webhook push ingestion.
- Kubernetes CRDs/operator.

# Capcom V1 Architecture Pack

## Purpose

These documents are the source of truth before V1 development starts. They define what Capcom V1 builds, what it deliberately does not build, and how the Gantry-first architecture should be implemented.

## V1 Thesis

Capcom V1 proves one control-plane loop:

1. Connect to Gantry.
2. Import live agent and access state.
3. Store approved desired state.
4. Detect drift.
5. Execute one safe audited control action.
6. Preserve an audit trail.

The product is not an observability dashboard or a new agent runtime. Gantry runs agents. Capcom governs approved state and reconciles actual runtime state against it.

## Canonical Docs

| Order | Document | Purpose |
|---:|---|---|
| 1 | `01-architecture-overview.md` | System boundaries, components, data flows |
| 2 | `02-domain-model-database.md` | Domain entities, database tables, indexes |
| 3 | `03-gantry-adapter-contract.md` | Current Gantry endpoint mapping and adapter behavior |
| 4 | `04-api-contract.md` | Capcom REST API for V1 |
| 5 | `05-manifest-yaml-spec.md` | YAML desired-state schema |
| 6 | `06-drift-detection-design.md` | Drift comparison, severity, resolution |
| 7 | `07-control-action-safety.md` | Safe mutation design and audit semantics |
| 8 | `08-sync-worker-design.md` | Polling, retries, degraded runtime behavior |
| 9 | `09-security-model.md` | Auth, secret storage, scopes, audit retention |
| 10 | `10-v1-build-plan.md` | Milestones, test plan, demo acceptance |
| 11 | `11-development-rules.md` | Engineering rules before implementation starts |
| 12 | `12-go-coding-rulebook.md` | Go-specific coding practices for agents and engineers |
| 13 | `13-execution-implementation-plan.md` | Execution sequence, task breakdown, and acceptance gates |
| 14 | `14-durable-runtime-sync-and-access-control-plan.md` | Detailed plan for durable sync, stale-state handling, worker scheduling, and audited access control |
| 15 | `15-multi-runtime-instances.md` | Multi-instance identity, isolation, API hierarchy, and console contract |

## V1 Decisions

- Backend: Go.
- Database: Postgres.
- Dashboard: dependency-free web console embedded in the Go server for V1.
- CLI: Go CLI in the same repo/module.
- First runtime adapter: Gantry.
- Event mode: polling/control API first.
- Webhooks: Phase 2.
- Enforcement mode: Phase 2.
- Kubernetes operator: Phase 2.
- OpenTelemetry/LangSmith/Langfuse/Phoenix adapters: after Gantry V1 loop works.

## Definition Of V1 Done

- Valid Gantry connection activates.
- Invalid Gantry connection is rejected.
- Gantry agents import into Capcom.
- Agent access imports from `/v1/agents/{agentId}/access`.
- Desired `Agent` manifest can be applied.
- Extra actual capability creates an open drift record.
- Operator can restrict access or disable/enable an agent.
- Every mutation has actor, reason, before, after, result, and timestamp.
- Gantry outage marks runtime degraded and preserves last known state.

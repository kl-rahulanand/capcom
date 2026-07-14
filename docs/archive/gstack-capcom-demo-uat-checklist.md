# Capcom Gantry MVP Demo And UAT Checklist

## Demo Goal

Prove that Capcom can connect to Gantry, import agents, apply desired state, detect capability drift, execute a safe control action, and preserve an audit trail.

## Participants

| Role | Responsibility |
|---|---|
| Demo driver | Runs Gantry, Capcom, CLI, and dashboard |
| Platform engineer reviewer | Validates setup and manifests |
| Operator reviewer | Validates drift and control workflow |
| Security reviewer | Validates access, reason capture, and audit evidence |

## Pre-Demo Checklist

| Check | Expected Result | Evidence |
|---|---|---|
| Docker Desktop running | `docker ps` succeeds | Screenshot or command output |
| Gantry Postgres running | Postgres container healthy | Container status |
| Gantry health passes | `/v1/health` returns ok | Response body |
| Gantry doctor passes | `/v1/doctor` has no blocking failures | Response body |
| Gantry OpenAPI captured | `gantry-openapi.snapshot.json` exists | File path |
| Capcom server starts | `/healthz` returns ok | Response body |
| Capcom DB migrations applied | Required tables exist | Migration output |
| Admin token configured | API rejects unauthenticated request | HTTP 401/403 response |

## Demo Script

1. Start Gantry.
2. Start Capcom.
3. Create a Gantry runtime connection in Capcom.
4. Show connection health and doctor status.
5. Sync/import Gantry agents.
6. Show Fleet View with imported agents.
7. Open one Agent Detail page.
8. Show actual tools, skills, MCP servers, conversations, and approvers.
9. Apply an `Agent` manifest for that agent.
10. Show desired state beside actual state.
11. Introduce one extra Gantry capability outside the manifest.
12. Trigger Capcom sync or wait for polling.
13. Show the drift record.
14. Execute `restrict_capability` with actor and reason.
15. Show Gantry access updated after sync.
16. Show audit log entries for manifest apply, drift detection, and control action.

## Acceptance Criteria

| ID | Scenario | Pass Criteria |
|---|---|---|
| UAT-01 | Valid runtime connection | Capcom activates connection only after Gantry health and doctor checks |
| UAT-02 | Invalid credentials | Capcom rejects connection and does not mark it active |
| UAT-03 | Agent import | Imported agent appears with runtime id, owner fields, status, and timestamps |
| UAT-04 | Capability import | Agent detail shows actual Gantry tools, skills, MCP servers, conversations, and approvers where available |
| UAT-05 | Desired state apply | Applying manifest creates or updates `AgentDesiredState` |
| UAT-06 | Drift detection | Extra runtime capability creates an open drift record |
| UAT-07 | Drift resolution | Restricting capability resolves or supersedes the drift record after sync |
| UAT-08 | Control guardrails | Control action requires actor, reason, and compatible runtime mode |
| UAT-09 | Audit trail | Mutating actions include actor, reason, before/after, result, and timestamp |
| UAT-10 | Runtime outage | Capcom marks runtime degraded and preserves last known imported agents |

## Evidence To Capture

- Runtime connection creation response.
- Gantry health and doctor responses.
- Agent import response or dashboard screenshot.
- Desired manifest used for the demo.
- Actual Gantry access before drift.
- Actual Gantry access after drift.
- Capcom drift record JSON.
- Control action request and response.
- Audit log rows for apply, drift, and restrict.
- Runtime outage behavior if tested.

## Failure Handling

| Failure | Demo Response |
|---|---|
| Gantry unavailable | Show Capcom degraded status and last known state preservation |
| Event ingestion unavailable | Use manual sync or polling fallback and document event blocker |
| Control action unsupported by Gantry | Show failed control action with audit entry and no hidden mutation |
| OpenAPI changed | Freeze snapshot, update adapter contract, and rerun schema checks |

## Definition Of Done

The MVP demo is done when:

- One Gantry runtime is connected.
- One agent is imported.
- One desired-state manifest is applied.
- One capability drift is detected.
- One safe control action is executed.
- One audit trail proves who did what, why, and what changed.


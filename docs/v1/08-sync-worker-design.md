# 08 - Sync Worker Design

## Goal

Keep Capcom's actual state close enough to Gantry runtime state for drift detection and operator decisions.

## Sync Modes

V1 supports:

- manual sync
- periodic polling

V1 does not require:

- webhooks
- streaming
- distributed worker fleet

Post-V1 adds signed Gantry webhook ingestion as an acceleration path. A webhook
is durably deduplicated and schedules a coalesced targeted adapter sync; it does
not mutate Gantry directly or replace periodic full reconciliation. The detailed
receiver, signature, replay-protection, and inbox design is defined in
`16-adapter-roadmap-and-webhook-plan.md`.

## Sync Inputs

Runtime connection:

- endpoint
- auth ref
- mode
- interval
- last cursor values

## Sync Steps

1. Mark sync run `running`.
2. Call Gantry `GET /v1/health`.
3. Call Gantry `GET /v1/doctor`.
4. List agents.
5. For each agent:
   - read agent detail
   - read agent admin
   - read agent access
   - upsert normalized agent
   - upsert actual state
6. List inventory/capabilities.
7. Optionally list recent runs/events.
8. Mark runtime active and sync run succeeded.
9. Run drift detection.

The persisted sync run records `diagnostics_seen`, `inventory_seen`, and
`capabilities_seen`. Inventory and capabilities are instance-scoped and retain
the last successful observation when the runtime later becomes unavailable.

## Degraded Runtime Behavior

If health/doctor fails:

- mark runtime `degraded`
- record `last_error`
- create failed sync run
- preserve last known agents and actual state
- do not delete missing agents
- do not auto-resolve drift

## Missing Agents

If an agent disappears from Gantry:

- V1 should mark it `not_seen` only after repeated successful syncs where it is absent.
- Do not delete agent rows in V1.

Suggested threshold:

```text
missing_after_successful_syncs = 3
```

## Event Ingestion

V1 minimum:

- store run/job events when reachable
- dedupe by `(runtime_connection_id, external_event_id)`

If event IDs are absent, derive a stable hash from:

```text
runtime + event type + occurredAt + agent/run/job id + payload hash
```

## Scheduling

Default interval:

```text
60 seconds
```

Allow per-runtime override:

```yaml
sync:
  intervalSeconds: 60
```

## Retries

- Retry transient HTTP/network failures with short backoff inside one sync run.
- Do not retry mutations in sync worker.
- Do not run overlapping syncs for the same runtime.

## Observability

Expose:

- last sync status
- last sync duration
- agents seen
- delegation edges seen
- events seen
- last error
- runtime status

Dashboard Runtime Connections page should show these fields.

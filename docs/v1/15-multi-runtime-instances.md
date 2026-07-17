# 15 - Multi-Runtime Instances

## Hierarchy

```text
adapter kind -> runtime instance -> durable agents -> ephemeral subagent executions
```

A Gantry installation is one runtime instance. The Gantry adapter is stateless
and reusable; it is not itself an installation or an agent.

## Required Identity

Each instance has a unique stable `name`, mutable `display_name`, environment,
labels, endpoint, mode, and encrypted `auth_ref`. Operators must never rely on
color alone to distinguish instances.

Changing mutable identity uses `PATCH /v1/runtime-instances/{id}` and requires
an `actor` and `reason`. The instance UUID, stable key, endpoint, and credential
reference are not changed by this operation.

Example:

| Key | Display name | Environment | Endpoint |
|---|---|---|---|
| `gantry-development` | Gantry Development | `development` | `http://127.0.0.1:8787` |
| `gantry-staging` | Gantry Staging | `staging` | `http://127.0.0.1:8788` |
| `gantry-production` | Gantry Production | `production` | `http://127.0.0.1:8789` |

## Deployment Isolation

Run installations with different Compose project names, host ports, Postgres
databases or volumes, and Control API keys. No Gantry source change is needed.
When Capcom runs in Docker, use reachable container DNS names or
`host.docker.internal` instead of loopback host URLs.

## Failure Isolation

- Sync locks and sync schedules are per runtime instance UUID.
- A failed instance preserves its own last-known state and cannot degrade peers.
- Identical Gantry agent, skill, run, and task IDs remain distinct because all
  unique keys include `runtime_connection_id`.
- Control actions resolve the owning agent binding before selecting an endpoint
  and secret.

## Console Contract

Selectors and instance cards show display name, environment, endpoint, adapter,
status, and stable key. Agent and subagent views retain that context. Global
search results must include the owning instance before fleet search is added.

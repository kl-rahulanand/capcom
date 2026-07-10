# 02 - Domain Model And Database Design

## Entity Overview

| Entity | Purpose |
|---|---|
| RuntimeConnection | Connected agent runtime and credential reference |
| RuntimeSyncRun | One runtime sync attempt and result |
| Agent | Normalized agent registry record |
| AgentDesiredState | Approved Capcom state for an agent |
| AgentActualState | Latest observed runtime state |
| AgentEvent | Normalized runtime event timeline |
| DriftRecord | Desired-vs-actual mismatch |
| ControlAction | Operator-requested runtime mutation |
| AuditLog | Immutable record of sync and mutation outcomes |
| SecretRef | Local encrypted or external secret reference |

## ID Strategy

- Use Capcom UUIDs for internal primary keys.
- Preserve runtime IDs as `external_id`.
- Runtime identity is always `(runtime_connection_id, external_id)`.
- Agent names are display identifiers, not stable keys.

## Tables

### runtime_connections

| Column | Type | Notes |
|---|---|---|
| id | uuid pk | Internal id |
| name | text unique | Human name |
| runtime_type | text | `gantry` in V1 |
| mode | text | `read_only`, `control_enabled` |
| endpoint_kind | text | `base_url`, `socket` |
| endpoint | text | URL or socket path |
| auth_ref | text | Secret ref |
| status | text | `pending`, `active`, `degraded`, `disabled`, `failed` |
| last_sync_at | timestamptz null | Last successful sync |
| last_error | text null | Latest failure |
| created_at | timestamptz |  |
| updated_at | timestamptz |  |

Indexes:

- unique `(name)`
- index `(runtime_type, status)`

### runtime_sync_runs

| Column | Type | Notes |
|---|---|---|
| id | uuid pk |  |
| runtime_connection_id | uuid fk |  |
| status | text | `running`, `succeeded`, `failed` |
| started_at | timestamptz |  |
| finished_at | timestamptz null |  |
| agents_seen | integer |  |
| events_seen | integer |  |
| error | text null |  |

### agents

| Column | Type | Notes |
|---|---|---|
| id | uuid pk | Capcom agent id |
| runtime_connection_id | uuid fk |  |
| external_id | text | Gantry agent id |
| name | text | Runtime display name |
| runtime_status | text | Latest imported runtime status |
| owner_business | text null | Desired/metadata |
| owner_technical | text null | Desired/metadata |
| escalation_contact | text null | Desired/metadata |
| purpose | text null | Desired/metadata |
| environment | text null | Desired/metadata |
| risk_level | text | `low`, `medium`, `high`, `critical` |
| first_seen_at | timestamptz |  |
| last_seen_at | timestamptz |  |
| created_at | timestamptz |  |
| updated_at | timestamptz |  |

Indexes:

- unique `(runtime_connection_id, external_id)`
- index `(runtime_status)`
- index `(risk_level)`

### agent_desired_states

| Column | Type | Notes |
|---|---|---|
| id | uuid pk |  |
| agent_id | uuid fk |  |
| manifest_version | text | `capcom.ai/v1alpha1` |
| desired_status | text | `active`, `disabled` |
| desired_access_json | jsonb | Sources and selected capabilities |
| approvals_json | jsonb | Approval requirements |
| policies_json | jsonb | Drift mode, limits |
| manifest_json | jsonb | Full normalized manifest |
| source | text | `api`, `cli`, `dashboard`, `yaml` |
| applied_by | text | Actor |
| applied_at | timestamptz |  |

Indexes:

- unique `(agent_id)`
- gin `(desired_access_json)`

### agent_actual_states

| Column | Type | Notes |
|---|---|---|
| id | uuid pk |  |
| agent_id | uuid fk |  |
| runtime_status | text |  |
| access_json | jsonb | Normalized Gantry `/access` response |
| admin_json | jsonb | Normalized admin detail |
| inventory_refs_json | jsonb | Referenced capabilities/sources |
| raw_runtime_json | jsonb | Raw payload for debugging |
| observed_at | timestamptz |  |

Indexes:

- unique `(agent_id)`
- gin `(access_json)`

### agent_events

| Column | Type | Notes |
|---|---|---|
| id | uuid pk |  |
| runtime_connection_id | uuid fk |  |
| agent_id | uuid fk null | May be unknown |
| external_event_id | text | Runtime event id |
| event_type | text | Normalized type |
| severity | text | `info`, `warning`, `critical` |
| payload_json | jsonb | Raw/normalized event |
| occurred_at | timestamptz |  |
| ingested_at | timestamptz |  |

Indexes:

- unique `(runtime_connection_id, external_event_id)`
- index `(agent_id, occurred_at desc)`

### drift_records

| Column | Type | Notes |
|---|---|---|
| id | uuid pk |  |
| agent_id | uuid fk |  |
| drift_type | text | `status`, `capability`, `source`, `approval` |
| field_path | text | Human/debug path |
| desired_json | jsonb | Desired value |
| actual_json | jsonb | Actual value |
| severity | text | `info`, `warning`, `critical` |
| mode | text | `observe`, `approval`, `enforce` |
| status | text | `open`, `acknowledged`, `resolved`, `ignored` |
| first_seen_at | timestamptz |  |
| last_seen_at | timestamptz |  |
| resolved_at | timestamptz null |  |

Indexes:

- index `(agent_id, status)`
- unique open drift key: `(agent_id, drift_type, field_path)` where status in open/acknowledged

### control_actions

| Column | Type | Notes |
|---|---|---|
| id | uuid pk |  |
| runtime_connection_id | uuid fk |  |
| agent_id | uuid fk null |  |
| action_type | text | `disable_agent`, `enable_agent`, `replace_access`, `restrict_capability` |
| requested_by | text | Actor |
| reason | text | Required |
| parameters_json | jsonb | Request |
| status | text | `pending`, `running`, `succeeded`, `failed`, `cancelled` |
| runtime_request_json | jsonb null | Sent request |
| runtime_response_json | jsonb null | Runtime result |
| error | text null |  |
| created_at | timestamptz |  |
| finished_at | timestamptz null |  |

### audit_logs

| Column | Type | Notes |
|---|---|---|
| id | uuid pk |  |
| actor | text |  |
| action | text |  |
| target_type | text |  |
| target_id | text |  |
| reason | text null |  |
| before_json | jsonb null |  |
| after_json | jsonb null |  |
| result | text | `succeeded`, `failed` |
| metadata_json | jsonb | Runtime ids, request ids |
| created_at | timestamptz |  |

Audit logs are append-only in application code.

## JSON Shape Guidance

Use JSONB for runtime-specific or frequently changing payloads:

- `agent_actual_states.access_json`
- `agent_actual_states.admin_json`
- `agent_desired_states.desired_access_json`
- `drift_records.desired_json`
- `drift_records.actual_json`

Do not put core query fields only in JSON. Runtime status, risk, mode, and timestamps should be typed columns.


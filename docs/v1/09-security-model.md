# 09 - Security Model

## V1 Security Goals

- Avoid storing plaintext runtime keys in manifests.
- Separate read-only and control-enabled runtime modes.
- Require actor and reason for mutations.
- Preserve immutable audit logs.
- Keep the first version simple enough to build.

## Capcom Auth

V1 uses single admin token:

```text
CAPCOM_ADMIN_TOKEN
```

All API requests except `GET /healthz` require:

```text
Authorization: Bearer <token>
```

Future:

- users
- SSO
- RBAC
- per-action permissions

## Runtime Credentials

Runtime API keys should be stored as:

- encrypted local secret in Postgres, or
- external secret reference

V1 acceptable implementation:

- `secrets` table
- AES-GCM encrypted value
- encryption key from `CAPCOM_SECRET_KEY`

Implemented V1 details:

- `CAPCOM_SECRET_KEY` is standard base64 encoding of exactly 32 random bytes.
- Secret values are encrypted with AES-256-GCM before repository writes.
- The secret name is AES-GCM associated data.
- Ciphertext includes an internal format version and random nonce.
- Create and rotate require actor and reason and write append-only audit events.
- Gantry credentials are resolved immediately before an adapter request and sent as a Bearer token.
- Secret values, ciphertext, and Authorization headers are never returned by APIs or written to audit metadata.

Manifests reference secrets by name:

```yaml
auth:
  apiKeyRef: gantry-control-api-key
```

Inline secrets in YAML are rejected.

## Runtime Modes

| Mode | Behavior |
|---|---|
| read_only | sync/import/drift only; no runtime mutation |
| control_enabled | allows approved control actions |

Control service must check mode before every mutation.

## Gantry Scopes

Read-like V1 connection:

```text
sessions:read
agents:admin
jobs:read
skills:read
mcp:read
```

Control-enabled:

```text
sessions:read
agents:admin
jobs:read
jobs:write
skills:read
skills:admin
mcp:read
mcp:admin
```

## Audit Retention

V1:

- keep audit logs indefinitely
- keep runtime events at least 30 days
- do not hard-delete audit rows

## Sensitive Data

Avoid logging:

- runtime API keys
- secret values
- full credential payloads
- raw Authorization headers

Raw runtime payload storage should redact obvious secret fields.

## Threats And Mitigations

| Threat | Mitigation |
|---|---|
| Stolen Capcom admin token | Local MVP risk; rotate token, move to SSO/RBAC later |
| Over-scoped Gantry key | Support read-only mode and document scopes |
| Accidental destructive action | No delete-agent action in V1; require reason and confirmation |
| Silent mutation | Audit before and after every control action |
| Runtime drift hidden by outage | Preserve last known state and mark runtime degraded |

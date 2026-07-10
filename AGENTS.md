# Capcom Development Rules

These rules apply to the whole repository.

## Product Boundary

- Capcom is a runtime-agnostic AgentOps control plane.
- Gantry is the first adapter, not the product boundary.
- Core domain code must not import Gantry-specific request/response types.
- Adapter-specific code belongs under `internal/adapters/<runtime>`.
- Store raw runtime payloads only for debugging/audit; normalize before domain use.

## V1 Source Of Truth

- Treat `docs/v1/` as the implementation contract.
- If implementation discovers a mismatch, update the relevant V1 doc in the same change.
- Do not expand scope beyond V1 without adding it to the explicit post-V1 backlog.

## Backend Rules

- Use Go for the core server, services, workers, repositories, adapters, and CLI.
- Follow `docs/v1/12-go-coding-rulebook.md` for Go-specific coding style.
- Keep handlers thin: parse/auth/validate, then call services.
- Keep services runtime-neutral.
- Keep repositories free of business logic.
- Every mutating API must record audit.
- Runtime mutations must require actor and reason.
- Read-only runtime connections must reject control actions.

## Adapter Rules

- All adapters must declare capability level and supported actions.
- Control actions must validate support before execution.
- Gantry V1 should use `/v1/agents/{agentId}/access` as the canonical access document.
- Capcom must not read Gantry database tables directly.
- Capcom must preserve last known state when a runtime is unavailable.

## Data Rules

- Use typed columns for fields we query often.
- Use JSONB for runtime-specific payloads and snapshots.
- Do not store plaintext runtime keys.
- Audit logs are append-only.
- Do not delete imported agents on a single missing sync.

## Testing Rules

- Add focused unit tests for domain logic, drift logic, validators, and adapter normalization.
- Add repository/API integration tests for risky paths.
- Use recorded Gantry fixtures for adapter contract tests.
- Do not require a live Gantry runtime for normal unit tests.

## Documentation Rules

- Update docs when architecture, schema, API, or runtime contract changes.
- Keep diagrams in `docs/Architecture/`.
- Keep implementation docs in `docs/v1/`.

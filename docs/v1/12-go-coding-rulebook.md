# 12 - Go Coding Rulebook

## Purpose

This rulebook tells coding agents and engineers how to write Go for Capcom. It adapts official Go guidance and established production style guides to Capcom's V1 architecture.

Primary sources:

- Effective Go: https://go.dev/doc/effective_go
- Go Code Review Comments: https://go.dev/wiki/CodeReviewComments
- Organizing a Go module: https://go.dev/doc/modules/layout
- Go context patterns: https://go.dev/blog/context
- Go error handling: https://go.dev/blog/error-handling-and-go
- Go database connection management: https://go.dev/doc/database/manage-connections
- Go fuzzing tutorial: https://go.dev/doc/tutorial/fuzz
- Go vulnerability management: https://go.dev/doc/security/vuln/
- Google Go Style Guide: https://google.github.io/styleguide/go/
- Uber Go Style Guide: https://github.com/uber-go/guide/blob/master/style.md

## Non-Negotiable Rules

- Run `gofmt` on all Go files.
- Run `go test ./...` before claiming backend work is complete.
- Run `go vet ./...` before merging non-trivial backend changes.
- Do not use `panic` for expected errors.
- Do not use `unsafe`.
- Do not introduce global mutable state for business logic.
- Do not let Gantry types leak into `internal/domain` or runtime-neutral services.
- All external calls must accept `context.Context`.
- All runtime mutations must be audited.

## Module And Package Layout

Capcom V1 uses one Go module at repo root unless implementation proves a multi-module setup is needed.

Recommended layout:

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

migrations/
```

Rules:

- Put binaries under `cmd/<name>`.
- Put private application code under `internal`.
- Keep package names short, lowercase, and single-word where possible.
- Do not create catch-all packages named `common`, `utils`, `helpers`, or `shared`.
- Package by responsibility, not by technical file type.

Good:

```text
internal/drift
internal/adapters/gantry
internal/services
```

Bad:

```text
internal/common
internal/helpers
internal/models
```

## Dependency Direction

Allowed:

```text
cmd -> api/services/config
api -> services/domain
services -> domain/store/adapters/audit
store -> domain
adapters/gantry -> adapters/runtime/domain
domain -> standard library only or tiny value-object helpers
```

Forbidden:

```text
domain -> adapters
domain -> store
store -> services
api -> adapters/gantry directly
services -> concrete Gantry response structs
```

If a dependency feels hard, add an interface at the consumer boundary instead of importing upward.

## Interfaces

Rules:

- Define interfaces where they are consumed, not where they are implemented.
- Keep interfaces small.
- Prefer concrete types until an abstraction is needed.
- Never use `*interface`.
- Add compile-time interface checks for important implementations.

Example:

```go
var _ runtime.Adapter = (*gantry.Adapter)(nil)
```

The `RuntimeAdapter` interface is architectural. Do not add Gantry-only methods to it.

## Naming

Rules:

- Use Go initialisms consistently: `ID`, `API`, `URL`, `HTTP`, `JSON`, `SQL`.
- Avoid stutter: prefer `runtime.Connection`, not `runtime.RuntimeConnection` if package context is clear.
- Avoid `Get` prefixes unless the operation is not a simple accessor and local style already uses it.
- Keep names specific enough to survive code review.
- Boolean names should read naturally: `isActive`, `hasDrift`, `controlEnabled`.

Package naming:

- lowercase
- no underscores
- no mixed caps
- singular unless plural is the domain term

## Error Handling

Rules:

- Return errors; do not panic for expected failures.
- Add context when returning errors across package boundaries.
- Use `fmt.Errorf("...: %w", err)` for wrapping.
- Use `errors.Is` and `errors.As` for classified errors.
- Handle an error exactly once: return it, wrap it, or log it, but do not log and return repeatedly.
- API handlers should translate service errors into stable API error codes.
- Runtime adapter errors should preserve enough information to classify timeout, auth failure, not found, and rejected mutation.

Good:

```go
agents, err := adapter.ListAgents(ctx)
if err != nil {
    return fmt.Errorf("list gantry agents: %w", err)
}
```

Bad:

```go
if err != nil {
    log.Printf("error: %v", err)
    return err
}
```

Use sentinel/domain errors for expected service outcomes:

```go
var ErrRuntimeReadOnly = errors.New("runtime connection is read-only")
```

## Context

Rules:

- `context.Context` is the first argument of request-scoped functions.
- Do not store contexts in structs.
- Do not pass `nil` contexts.
- Respect cancellation before expensive or remote work.
- External calls must have timeouts.
- Do not put optional function parameters in context.

Good:

```go
func (s *RuntimeService) Sync(ctx context.Context, runtimeID domain.RuntimeID) error
```

Bad:

```go
type RuntimeService struct {
    ctx context.Context
}
```

## HTTP/API Handlers

Rules:

- Handlers authenticate, decode, validate, call service, encode.
- Keep business logic out of handlers.
- Do not write partial responses before service work is complete.
- Use stable JSON error shape from `docs/v1/04-api-contract.md`.
- Decode with size limits.
- Use `http.Server` timeouts.
- Never echo secrets or raw auth headers.

Handler shape:

```go
func (h *Handler) CreateRuntimeConnection(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    // auth, decode, validate
    // call service
    // write response
}
```

## Database

Rules:

- Use `database/sql` or a well-maintained driver/pool consistently.
- Configure connection pool limits explicitly.
- Pass context to every query.
- Use transactions for multi-table mutations.
- Keep SQL in repository packages.
- Repositories do not call services or adapters.
- Avoid long transactions around network calls.
- Never build SQL with string concatenation from user input.

Connection pool values must be config-driven:

```text
DB_MAX_OPEN_CONNS
DB_MAX_IDLE_CONNS
DB_CONN_MAX_LIFETIME
DB_CONN_MAX_IDLE_TIME
```

Transaction rule:

```text
Read current state -> compute mutation -> write DB changes in transaction -> call runtime outside transaction unless atomicity requires another pattern.
```

For control actions, do not hold DB transaction open during Gantry HTTP calls. Store pending action, audit pre-state, call Gantry, then update result.

## JSON And API DTOs

Rules:

- Separate API DTOs from domain types when transport shape differs from domain shape.
- Use explicit JSON tags.
- Avoid exposing DB structs directly through API.
- Use `omitempty` intentionally; do not hide meaningful zero values.
- Validate request DTOs before passing to services.

## Concurrency

Rules:

- Start goroutines only when there is a lifecycle owner.
- Every goroutine must have a cancellation path.
- Do not fire-and-forget from request handlers.
- Prefer worker structs with `Start(ctx)` and `Stop`/context cancellation.
- Use channels for ownership transfer or signaling, not as magical queues.
- Channel buffer size should usually be 0 or 1 unless measured.
- Run `go test -race ./...` before merging concurrency-heavy changes.

Worker pattern:

```go
func (w *SyncWorker) Run(ctx context.Context) error {
    ticker := time.NewTicker(w.interval)
    defer ticker.Stop()
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-ticker.C:
            w.syncOnce(ctx)
        }
    }
}
```

## Time

Rules:

- Use `time.Time` in domain/storage code.
- Store timestamps in UTC.
- Use `time.Now().UTC()` at boundaries.
- Inject a clock in tests when time matters.
- Never compare time strings for business logic.

## Logging

Rules:

- Use structured logging.
- Include request id, runtime id, agent id, control action id where relevant.
- Do not log secrets.
- Log errors at the boundary where they are handled.
- Do not log and return the same error at every layer.

## Configuration

Rules:

- Config loads once at startup.
- Validate config before starting server/workers.
- Keep config structs explicit.
- Do not read environment variables deep inside services.
- Secrets are references or encrypted values, not plain config dumps.

## Testing

Minimum expectations:

- Unit tests for pure domain logic.
- Table-driven tests for validators, drift detection, and control action validation.
- Adapter normalization tests using recorded Gantry fixtures.
- Repository tests for migrations and queries.
- API tests for endpoint behavior and error shape.
- Optional live Gantry tests behind env flag.

Table-driven test shape:

```go
func TestDetectCapabilityDrift(t *testing.T) {
    tests := []struct {
        name string
        desired []CapabilitySelection
        actual []CapabilitySelection
        want int
    }{
        // cases
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // assert
        })
    }
}
```

Use fuzzing for parsers and manifest validation once the basic tests exist.

## Security

Rules:

- Run `govulncheck ./...` before release candidates.
- Do not use `unsafe`.
- Do not log credentials.
- Use constant-time comparison for tokens when practical.
- Redact raw runtime payloads before storing if they may contain secrets.
- Avoid shelling out from the server. If absolutely required, isolate and review.

## Performance

Rules:

- Optimize after measuring.
- Avoid unnecessary allocations in hot loops only when profiling shows a problem.
- Preallocate slices/maps when size is known.
- Use `strconv` instead of `fmt` in hot conversion paths.
- Do not add caching until the invalidation story is clear.

## Comments And Documentation

Rules:

- Exported identifiers need useful doc comments.
- Comments should explain why, invariants, or non-obvious behavior.
- Avoid comments that restate code.
- Package comments are required for packages with exported APIs.

Good:

```go
// RuntimeAdapter is implemented by runtimes Capcom can observe or control.
// Implementations must not expose runtime-specific payloads to services.
type RuntimeAdapter interface { ... }
```

## Capcom-Specific Review Checklist

Before merging Go code, check:

- Does this keep Gantry behind `internal/adapters/gantry`?
- Does this preserve runtime-neutral domain types?
- Does every external call accept context?
- Does every mutation write audit?
- Does read-only mode reject control action?
- Does runtime outage preserve last known state?
- Are errors wrapped with useful context?
- Are secrets redacted?
- Are tests included for success and failure paths?
- Did `gofmt`, `go test ./...`, and `go vet ./...` pass?


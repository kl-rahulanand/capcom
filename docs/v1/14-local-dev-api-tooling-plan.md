# 14 - Local Dev And API Tooling Plan

## Purpose

This plan captures the decisions from the local development and API tooling discussion before we implement them.

Capcom is currently a Go backend. Go should remain the source of truth for backend dependencies and versions. Node/npm should not be introduced for backend command shortcuts unless a real JavaScript or TypeScript frontend/tooling requirement appears.

## Decisions

### Dependency And Version Management

Use Go-native dependency management:

- `go.mod` for module name, Go version, and dependencies.
- `go.sum` for dependency checksums.
- `go get` for adding or upgrading dependencies.
- `go mod tidy` for pruning and normalizing module metadata.
- `go install <tool>@<version>` or tracked Go tool dependencies for Go-based CLI tools.

Do not add `package.json` for the Go backend.

### Local Environment

Use local `.env` files for developer configuration:

- Commit `.env.example`.
- Ignore `.env`.
- Load `.env` automatically in local server and CLI flows.
- Keep production behavior compatible with real environment variables.
- Environment variables from the OS should override `.env` values.

Initial `.env.example` should include:

```env
CAPCOM_HTTP_ADDR=:8080
CAPCOM_SERVICE_VERSION=dev
CAPCOM_LOG_LEVEL=info
CAPCOM_DATABASE_URL=postgres://capcom:capcom@localhost:5433/capcom?sslmode=disable
CAPCOM_DATABASE_MAX_OPEN_CONNS=10
CAPCOM_DATABASE_MAX_IDLE_CONNS=5
CAPCOM_DATABASE_CONN_MAX_LIFETIME=30m
```

### Developer Commands

Keep `Makefile` for now because it already exists and is common in Go projects.

Add Windows-friendly scripts only if `make` is unavailable for the team. Preferred future option is `Taskfile.yml`, because it is cross-platform and does not imply a Node dependency.

Near-term command target list:

```text
make run
make migrate-up
make test
make vet
make tidy
```

Future Taskfile command equivalents:

```text
task run
task migrate
task test
task vet
task tidy
```

### API Documentation

Use OpenAPI as the API contract source of truth.

Add:

- `api/openapi.yaml`
- current REST endpoints
- request and response schemas
- error response schemas
- local server URL

Current endpoints to document first:

- `GET /healthz`
- `POST /v1/runtime-connections`
- `GET /v1/runtime-connections`
- `GET /v1/runtime-connections/{id}`
- `POST /v1/runtime-connections/{id}/test`

### Swagger/Postman

Do not hand-maintain a Postman collection first.

Preferred flow:

1. Maintain `api/openapi.yaml`.
2. Import OpenAPI into Postman when needed.
3. Later serve API docs from the server, for example `/docs`.

Recommended Go/OpenAPI tool direction:

- Prefer `oapi-codegen` for generated Go types/server/client code once the API contract stabilizes.
- Avoid annotation-first Swagger generation for now; it tends to make handler comments become the contract.

## Implementation Sequence

1. Add `.gitignore`.
2. Add `.env.example`.
3. Add a small `.env` loader in `internal/config`.
4. Update tests for `.env` loading and OS override behavior.
5. Add `make tidy`.
6. Add `api/openapi.yaml` for current endpoints.
7. Update `README.md` to prefer `.env` and `make` commands.
8. Re-run:
   - `gofmt`
   - `go test ./...`
   - `go vet ./...`

## Acceptance Criteria

- A new developer can configure local Capcom by copying `.env.example` to `.env`.
- Running the server does not require manually pasting env vars in every shell.
- Backend dependencies remain managed by `go.mod` and `go.sum`.
- README does not imply npm is needed for the Go backend.
- Current API endpoints are represented in OpenAPI.

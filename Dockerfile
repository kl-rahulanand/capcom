# syntax=docker/dockerfile:1

# Build the Go API server and the capcom CLI (used for migrations).
FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/capcom-server ./cmd/capcom-server \
 && CGO_ENABLED=0 GOOS=linux go build -o /out/capcom ./cmd/capcom

# Minimal runtime image (static binaries, non-root).
FROM gcr.io/distroless/static-debian12:nonroot AS runtime
WORKDIR /app
COPY --from=build /out/capcom-server /app/capcom-server
COPY --from=build /out/capcom /app/capcom
# The migrator reads SQL files from ./migrations at runtime (not embedded).
COPY --from=build /src/migrations /app/migrations
ENV CAPCOM_HTTP_ADDR=":8080"
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/capcom-server"]

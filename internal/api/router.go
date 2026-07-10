package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	runtimeadapter "capcom/internal/adapters/runtime"
	"capcom/internal/domain"
	"capcom/internal/services"
)

type RouterConfig struct {
	Version            string
	RuntimeConnections RuntimeConnectionService
}

type RuntimeConnectionService interface {
	Create(ctx context.Context, input services.CreateRuntimeConnectionInput) (domain.RuntimeConnection, error)
	Get(ctx context.Context, id string) (domain.RuntimeConnection, error)
	List(ctx context.Context) ([]domain.RuntimeConnection, error)
	Test(ctx context.Context, id string) (*runtimeadapter.CheckResult, error)
}

func NewRouter(cfg RouterConfig, logger *slog.Logger) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handleHealth(cfg))
	mux.HandleFunc("POST /v1/runtime-connections", handleCreateRuntimeConnection(cfg))
	mux.HandleFunc("GET /v1/runtime-connections", handleListRuntimeConnections(cfg))
	mux.HandleFunc("GET /v1/runtime-connections/{id}", handleGetRuntimeConnection(cfg))
	mux.HandleFunc("POST /v1/runtime-connections/{id}/test", handleTestRuntimeConnection(cfg))
	mux.HandleFunc("/", handleNotFound)

	return requestLogger(mux, logger)
}

func handleHealth(cfg RouterConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, healthResponse{
			Status:  "ok",
			Service: "capcom",
			Version: cfg.Version,
		})
	}
}

func handleNotFound(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotFound, errorResponse{
		Error: "not_found",
	})
}

func handleCreateRuntimeConnection(cfg RouterConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cfg.RuntimeConnections == nil {
			writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "database_not_configured"})
			return
		}

		var req createRuntimeConnectionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid_json"})
			return
		}

		conn, err := cfg.RuntimeConnections.Create(r.Context(), services.CreateRuntimeConnectionInput{
			Name:        req.Name,
			Kind:        domain.RuntimeKind(req.RuntimeType),
			Mode:        domain.RuntimeMode(req.Mode),
			Endpoint:    req.Endpoint,
			Actor:       req.Actor,
			Reason:      req.Reason,
			Description: req.Description,
		})
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
			return
		}

		writeJSON(w, http.StatusCreated, runtimeConnectionResponseFromDomain(conn))
	}
}

func handleListRuntimeConnections(cfg RouterConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cfg.RuntimeConnections == nil {
			writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "database_not_configured"})
			return
		}

		conns, err := cfg.RuntimeConnections.List(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "list_runtime_connections_failed"})
			return
		}

		response := make([]runtimeConnectionResponse, 0, len(conns))
		for _, conn := range conns {
			response = append(response, runtimeConnectionResponseFromDomain(conn))
		}
		writeJSON(w, http.StatusOK, response)
	}
}

func handleGetRuntimeConnection(cfg RouterConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cfg.RuntimeConnections == nil {
			writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "database_not_configured"})
			return
		}

		conn, err := cfg.RuntimeConnections.Get(r.Context(), r.PathValue("id"))
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeJSON(w, http.StatusNotFound, errorResponse{Error: "not_found"})
				return
			}
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "get_runtime_connection_failed"})
			return
		}

		writeJSON(w, http.StatusOK, runtimeConnectionResponseFromDomain(conn))
	}
}

func handleTestRuntimeConnection(cfg RouterConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cfg.RuntimeConnections == nil {
			writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "database_not_configured"})
			return
		}

		result, err := cfg.RuntimeConnections.Test(r.Context(), r.PathValue("id"))
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeJSON(w, http.StatusNotFound, errorResponse{Error: "not_found"})
				return
			}
			writeJSON(w, http.StatusBadGateway, errorResponse{Error: err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, runtimeConnectionTestResponse{
			Status:       string(result.Status),
			Message:      result.Message,
			Capabilities: result.Capabilities,
			Metadata:     result.Metadata,
		})
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		slog.Default().Error("write json response", "error", err)
	}
}

type healthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Version string `json:"version"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type createRuntimeConnectionRequest struct {
	Name        string `json:"name"`
	RuntimeType string `json:"runtime_type"`
	Mode        string `json:"mode"`
	Endpoint    string `json:"endpoint"`
	Actor       string `json:"actor"`
	Reason      string `json:"reason"`
	Description string `json:"description"`
}

type runtimeConnectionResponse struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	RuntimeType  string  `json:"runtime_type"`
	Mode         string  `json:"mode"`
	Status       string  `json:"status"`
	Endpoint     string  `json:"endpoint"`
	LastSyncedAt *string `json:"last_synced_at,omitempty"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

type runtimeConnectionTestResponse struct {
	Status       string                      `json:"status"`
	Message      string                      `json:"message"`
	Capabilities runtimeadapter.Capabilities `json:"capabilities"`
	Metadata     map[string]any              `json:"metadata,omitempty"`
}

func runtimeConnectionResponseFromDomain(conn domain.RuntimeConnection) runtimeConnectionResponse {
	var lastSyncedAt *string
	if conn.LastSyncedAt != nil {
		formatted := conn.LastSyncedAt.UTC().Format("2006-01-02T15:04:05Z07:00")
		lastSyncedAt = &formatted
	}

	return runtimeConnectionResponse{
		ID:           conn.ID,
		Name:         conn.Name,
		RuntimeType:  string(conn.Kind),
		Mode:         string(conn.Mode),
		Status:       string(conn.Status),
		Endpoint:     conn.BaseURL,
		LastSyncedAt: lastSyncedAt,
		CreatedAt:    conn.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    conn.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}
}

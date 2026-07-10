package gantry

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	runtimeadapter "capcom/internal/adapters/runtime"
	"capcom/internal/domain"
)

const defaultTimeout = 10 * time.Second

type Client struct {
	httpClient *http.Client
	timeout    time.Duration
}

func NewClient(httpClient *http.Client) Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return Client{
		httpClient: httpClient,
		timeout:    defaultTimeout,
	}
}

func (c Client) Kind() domain.RuntimeKind {
	return domain.RuntimeKindGantry
}

func (c Client) Check(ctx context.Context, conn domain.RuntimeConnection) (*runtimeadapter.CheckResult, error) {
	var payload map[string]any
	if err := c.doJSON(ctx, conn, http.MethodGet, "/v1/health", nil, &payload); err != nil {
		return nil, err
	}

	return &runtimeadapter.CheckResult{
		Status:  domain.RuntimeStatusActive,
		Message: "gantry health check succeeded",
		Capabilities: runtimeadapter.Capabilities{
			ReadAgents:         true,
			ReadAgentAccess:    true,
			ReplaceAgentAccess: conn.Mode == domain.RuntimeModeControlEnabled,
		},
		Metadata: payload,
	}, nil
}

func (c Client) ListAgents(ctx context.Context, conn domain.RuntimeConnection) ([]domain.AgentSnapshot, error) {
	var agents []gantryAgent
	if err := c.doJSON(ctx, conn, http.MethodGet, "/v1/agents", nil, &agents); err != nil {
		return nil, err
	}

	snapshots := make([]domain.AgentSnapshot, 0, len(agents))
	observedAt := time.Now().UTC()
	for _, agent := range agents {
		snapshots = append(snapshots, domain.AgentSnapshot{
			RuntimeAgentID: agent.ID,
			Name:           agent.DisplayName(),
			Status:         agent.StatusDomain(),
			Metadata: map[string]any{
				"description": agent.Description,
				"raw":         agent.Raw,
			},
			ObservedAt: observedAt,
		})
	}
	return snapshots, nil
}

func (c Client) GetAgentAccess(ctx context.Context, conn domain.RuntimeConnection, runtimeAgentID string) (*domain.AccessDocument, error) {
	path := fmt.Sprintf("/v1/agents/%s/access", url.PathEscape(runtimeAgentID))
	var access gantryAccess
	if err := c.doJSON(ctx, conn, http.MethodGet, path, nil, &access); err != nil {
		return nil, err
	}

	raw := access.Raw
	if raw == nil {
		raw = map[string]any{}
	}

	return &domain.AccessDocument{
		AgentID:    access.AgentID,
		Selections: normalizeSelections(access.Selections),
		Raw:        raw,
		ObservedAt: time.Now().UTC(),
		Source:     "gantry",
	}, nil
}

func (c Client) ReplaceAgentAccess(ctx context.Context, conn domain.RuntimeConnection, runtimeAgentID string, access domain.AccessDocument) (*domain.AccessDocument, error) {
	if conn.Mode != domain.RuntimeModeControlEnabled {
		return nil, fmt.Errorf("runtime connection is read-only")
	}

	path := fmt.Sprintf("/v1/agents/%s/access", url.PathEscape(runtimeAgentID))
	body := map[string]any{
		"selections": access.Selections,
	}

	var response gantryAccess
	if err := c.doJSON(ctx, conn, http.MethodPut, path, body, &response); err != nil {
		return nil, err
	}

	return &domain.AccessDocument{
		AgentID:    response.AgentID,
		Selections: normalizeSelections(response.Selections),
		Raw:        response.Raw,
		ObservedAt: time.Now().UTC(),
		Source:     "gantry",
	}, nil
}

func (c Client) doJSON(ctx context.Context, conn domain.RuntimeConnection, method string, path string, body any, out any) error {
	base, err := normalizeBaseURL(conn.BaseURL)
	if err != nil {
		return err
	}

	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		reader = bytes.NewReader(data)
	}

	requestCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(requestCtx, method, base+path, reader)
	if err != nil {
		return fmt.Errorf("build gantry request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("call gantry %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		limited, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("gantry %s %s returned %d: %s", method, path, resp.StatusCode, strings.TrimSpace(string(limited)))
	}

	if out == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode gantry response: %w", err)
	}
	return nil
}

func normalizeBaseURL(raw string) (string, error) {
	trimmed := strings.TrimRight(strings.TrimSpace(raw), "/")
	if trimmed == "" {
		return "", fmt.Errorf("gantry endpoint is required")
	}
	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", fmt.Errorf("parse gantry endpoint: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("gantry endpoint must use http or https")
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("gantry endpoint host is required")
	}
	return trimmed, nil
}

func normalizeSelections(selections []gantrySelection) []domain.AccessSelection {
	normalized := make([]domain.AccessSelection, 0, len(selections))
	for _, selection := range selections {
		normalized = append(normalized, domain.AccessSelection{
			Kind:    "capability",
			ID:      selection.ID,
			Name:    selection.ID,
			Allowed: true,
			Attributes: map[string]any{
				"version": selection.Version,
			},
		})
	}
	return normalized
}

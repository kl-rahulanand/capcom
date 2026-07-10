package store

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"capcom/internal/config"
	"capcom/internal/domain"
)

func TestRepositoriesIntegration(t *testing.T) {
	databaseURL := os.Getenv("CAPCOM_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("CAPCOM_TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	db, err := OpenPostgres(config.DatabaseConfig{
		URL:          databaseURL,
		MaxOpenConns: 2,
		MaxIdleConns: 1,
	})
	if err != nil {
		t.Fatalf("OpenPostgres returned error: %v", err)
	}
	defer db.Close()

	if _, err := NewMigrator(db, "../../migrations").Up(ctx); err != nil {
		t.Fatalf("migrate up: %v", err)
	}

	runtimes := NewRuntimeConnectionRepository(db)
	suffix := time.Now().UnixNano()
	runtimeConn, err := runtimes.Create(ctx, domain.RuntimeConnection{
		Name:    fmt.Sprintf("test-gantry-%d", suffix),
		Kind:    domain.RuntimeKindGantry,
		Mode:    domain.RuntimeModeReadOnly,
		Status:  domain.RuntimeStatusPending,
		BaseURL: "http://127.0.0.1:3000",
	})
	if err != nil {
		t.Fatalf("create runtime connection: %v", err)
	}

	gotRuntime, err := runtimes.Get(ctx, runtimeConn.ID)
	if err != nil {
		t.Fatalf("get runtime connection: %v", err)
	}
	if gotRuntime.Name != runtimeConn.Name {
		t.Fatalf("runtime name = %q, want %q", gotRuntime.Name, runtimeConn.Name)
	}

	agents := NewAgentRepository(db)
	agent, _, err := agents.Create(ctx, domain.Agent{
		Name:   fmt.Sprintf("test-agent-%d", suffix),
		Status: domain.AgentStatusEnabled,
	}, domain.AgentBinding{
		RuntimeConnectionID: runtimeConn.ID,
		RuntimeAgentID:      fmt.Sprintf("gantry-agent-%d", suffix),
	})
	if err != nil {
		t.Fatalf("create agent: %v", err)
	}

	gotAgent, err := agents.Get(ctx, agent.ID)
	if err != nil {
		t.Fatalf("get agent: %v", err)
	}
	if gotAgent.Name != agent.Name {
		t.Fatalf("agent name = %q, want %q", gotAgent.Name, agent.Name)
	}
}

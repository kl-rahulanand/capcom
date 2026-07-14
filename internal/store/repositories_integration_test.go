package store

import (
	"context"
	"fmt"
	"os"
	"strings"
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
	var databaseName string
	if err := db.QueryRowContext(ctx, `SELECT current_database()`).Scan(&databaseName); err != nil {
		t.Fatalf("read test database name: %v", err)
	}
	if !strings.Contains(strings.ToLower(databaseName), "test") {
		t.Skipf("CAPCOM_TEST_DATABASE_URL points to non-test database %q", databaseName)
	}

	if _, err := NewMigrator(db, "../../migrations").Up(ctx); err != nil {
		t.Fatalf("migrate up: %v", err)
	}

	runtimes := NewRuntimeConnectionRepository(db)
	secrets := NewSecretRepository(db)
	suffix := time.Now().UnixNano()
	secret, err := secrets.Create(ctx, domain.Secret{Name: fmt.Sprintf("test-secret-%d", suffix)}, []byte("encrypted-value"))
	if err != nil {
		t.Fatalf("create secret: %v", err)
	}
	ciphertext, err := secrets.GetCiphertext(ctx, secret.Name)
	if err != nil || string(ciphertext) != "encrypted-value" {
		t.Fatalf("get secret ciphertext = %q, %v", ciphertext, err)
	}
	runtimeConn, err := runtimes.Create(ctx, domain.RuntimeConnection{
		Name:    fmt.Sprintf("test-gantry-%d", suffix),
		Kind:    domain.RuntimeKindGantry,
		Mode:    domain.RuntimeModeReadOnly,
		Status:  domain.RuntimeStatusPending,
		BaseURL: "http://127.0.0.1:3000",
		AuthRef: secret.Name,
	})
	if err != nil {
		t.Fatalf("create runtime connection: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DELETE FROM agents WHERE id IN (
			SELECT agent_id FROM agent_runtime_bindings WHERE runtime_connection_id = $1
		)`, runtimeConn.ID)
		_, _ = db.ExecContext(context.Background(), `DELETE FROM runtime_connections WHERE id = $1`, runtimeConn.ID)
		_, _ = db.ExecContext(context.Background(), `DELETE FROM secrets WHERE id = $1`, secret.ID)
	})

	gotRuntime, err := runtimes.Get(ctx, runtimeConn.ID)
	if err != nil {
		t.Fatalf("get runtime connection: %v", err)
	}
	if gotRuntime.Name != runtimeConn.Name {
		t.Fatalf("runtime name = %q, want %q", gotRuntime.Name, runtimeConn.Name)
	}
	if gotRuntime.AuthRef != secret.Name {
		t.Fatalf("runtime auth ref = %q, want %q", gotRuntime.AuthRef, secret.Name)
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

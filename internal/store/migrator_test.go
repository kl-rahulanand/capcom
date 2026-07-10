package store

import (
	"strings"
	"testing"
)

func TestLoadMigrations(t *testing.T) {
	migrations, err := LoadMigrations("../../migrations")
	if err != nil {
		t.Fatalf("LoadMigrations returned error: %v", err)
	}
	if len(migrations) == 0 {
		t.Fatal("LoadMigrations returned no migrations")
	}
	if migrations[0].Version != "001_initial_schema" {
		t.Fatalf("first migration = %q, want 001_initial_schema", migrations[0].Version)
	}
	if !strings.Contains(migrations[0].SQL, "CREATE TABLE IF NOT EXISTS runtime_connections") {
		t.Fatal("initial migration does not create runtime_connections")
	}
}

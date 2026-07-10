package store

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Migrator struct {
	db  *sql.DB
	dir string
}

type Migration struct {
	Version string
	Path    string
	SQL     string
}

func NewMigrator(db *sql.DB, dir string) Migrator {
	return Migrator{db: db, dir: dir}
}

func (m Migrator) Up(ctx context.Context) ([]string, error) {
	migrations, err := LoadMigrations(m.dir)
	if err != nil {
		return nil, err
	}

	if _, err := m.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS schema_migrations (
	version text PRIMARY KEY,
	applied_at timestamptz NOT NULL DEFAULT now()
)`); err != nil {
		return nil, fmt.Errorf("ensure schema_migrations: %w", err)
	}

	applied := make([]string, 0, len(migrations))
	for _, migration := range migrations {
		ran, err := m.hasRun(ctx, migration.Version)
		if err != nil {
			return nil, err
		}
		if ran {
			continue
		}
		if err := m.apply(ctx, migration); err != nil {
			return nil, err
		}
		applied = append(applied, migration.Version)
	}

	return applied, nil
}

func LoadMigrations(dir string) ([]Migration, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read migrations dir: %w", err)
	}

	migrations := make([]Migration, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}

		version := strings.TrimSuffix(entry.Name(), ".sql")
		migrations = append(migrations, Migration{
			Version: version,
			Path:    path,
			SQL:     string(content),
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func (m Migrator) hasRun(ctx context.Context, version string) (bool, error) {
	var exists bool
	if err := m.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)`, version).Scan(&exists); err != nil {
		return false, fmt.Errorf("check migration %s: %w", version, err)
	}
	return exists, nil
}

func (m Migrator) apply(ctx context.Context, migration Migration) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration %s: %w", migration.Version, err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, migration.SQL); err != nil {
		return fmt.Errorf("apply migration %s: %w", migration.Version, err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, migration.Version); err != nil {
		return fmt.Errorf("record migration %s: %w", migration.Version, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %s: %w", migration.Version, err)
	}
	return nil
}

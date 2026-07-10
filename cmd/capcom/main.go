package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"capcom/internal/config"
	"capcom/internal/store"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "capcom: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return usage()
	}

	switch args[0] {
	case "migrate":
		if len(args) != 2 || args[1] != "up" {
			return usage()
		}
		return migrateUp()
	default:
		return usage()
	}
}

func usage() error {
	fmt.Fprintln(os.Stderr, "usage:")
	fmt.Fprintln(os.Stderr, "  capcom migrate up")
	os.Exit(2)
	return nil
}

func migrateUp() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if cfg.Database.URL == "" {
		return fmt.Errorf("CAPCOM_DATABASE_URL is required")
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))
	db, err := store.OpenPostgres(cfg.Database)
	if err != nil {
		return err
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	applied, err := store.NewMigrator(db, "migrations").Up(ctx)
	if err != nil {
		return err
	}
	logger.Info("migrations complete", "applied", applied)
	return nil
}

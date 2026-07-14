package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"capcom/internal/adapters/gantry"
	"capcom/internal/api"
	"capcom/internal/config"
	secretcipher "capcom/internal/secrets"
	"capcom/internal/services"
	"capcom/internal/store"
	"capcom/internal/workers"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.New(slog.NewJSONHandler(os.Stderr, nil)).Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))
	slog.SetDefault(logger)

	if err := run(context.Background(), cfg, logger); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, cfg config.Config, logger *slog.Logger) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	routerConfig := api.RouterConfig{
		Version:    cfg.Service.Version,
		AdminToken: cfg.Security.AdminToken,
	}
	var syncWorker *workers.RuntimeSyncWorker
	if cfg.Database.URL != "" {
		if cfg.Security.AdminToken == "" {
			return fmt.Errorf("CAPCOM_ADMIN_TOKEN is required when CAPCOM_DATABASE_URL is configured")
		}
		if len(cfg.Secrets.Key) != 32 {
			return fmt.Errorf("CAPCOM_SECRET_KEY is required when CAPCOM_DATABASE_URL is configured")
		}
		db, err := store.OpenPostgres(cfg.Database)
		if err != nil {
			return err
		}
		defer db.Close()

		pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if err := db.PingContext(pingCtx); err != nil {
			return err
		}

		cipher, err := secretcipher.NewCipher(cfg.Secrets.Key)
		if err != nil {
			return err
		}
		auditRepository := store.NewAuditRepository(db)
		secretService := services.NewSecretService(store.NewSecretRepository(db), auditRepository, cipher)
		runtimeRepository := store.NewRuntimeConnectionRepository(db)
		gantryAdapter := gantry.NewClient(nil, secretService)
		runtimeService := services.NewRuntimeConnectionService(runtimeRepository, auditRepository).
			WithCredentialResolver(secretService).WithAdapter(gantryAdapter)
		syncService := services.NewRuntimeSyncService(runtimeRepository, store.NewSyncRepository(db), auditRepository, cfg.Sync.MissingThreshold).
			WithAdapter(gantryAdapter)
		controlService := services.NewControlActionService(runtimeRepository, store.NewSyncRepository(db), store.NewControlActionRepository(db), auditRepository, syncService).
			WithAdapter(gantryAdapter)
		routerConfig.Secrets = secretService
		routerConfig.RuntimeConnections = runtimeService
		routerConfig.RuntimeSync = syncService
		routerConfig.ControlActions = controlService
		if cfg.Sync.WorkerEnabled {
			syncWorker = workers.NewRuntimeSyncWorker(runtimeService, syncService, cfg.Sync.WorkerTick, cfg.Sync.MaxConcurrency, cfg.Sync.RequestTimeout, logger)
		}
		logger.Info("postgres connected")
	} else {
		logger.Warn("database not configured; runtime connection APIs will return service unavailable")
	}
	if syncWorker != nil {
		go syncWorker.Run(ctx)
	}

	srv := &http.Server{
		Addr:              cfg.HTTP.Addr,
		Handler:           api.NewRouter(routerConfig, logger),
		ReadHeaderTimeout: cfg.HTTP.ReadHeaderTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("starting capcom server", "addr", cfg.HTTP.Addr, "version", cfg.Service.Version)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		logger.Info("shutting down capcom server")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}

	select {
	case err := <-errCh:
		return err
	case <-time.After(100 * time.Millisecond):
		return nil
	}
}

package config

import (
	"log/slog"
	"testing"
	"time"
)

func TestLoadFromLookupDefaults(t *testing.T) {
	cfg, err := LoadFromLookup(func(string) (string, bool) {
		return "", false
	})
	if err != nil {
		t.Fatalf("LoadFromLookup returned error: %v", err)
	}

	if cfg.HTTP.Addr != defaultHTTPAddr {
		t.Fatalf("HTTP addr = %q, want %q", cfg.HTTP.Addr, defaultHTTPAddr)
	}
	if cfg.HTTP.ReadHeaderTimeout != defaultReadHeaderTimeout {
		t.Fatalf("read header timeout = %s, want %s", cfg.HTTP.ReadHeaderTimeout, defaultReadHeaderTimeout)
	}
	if cfg.HTTP.ShutdownTimeout != defaultShutdownTimeout {
		t.Fatalf("shutdown timeout = %s, want %s", cfg.HTTP.ShutdownTimeout, defaultShutdownTimeout)
	}
	if cfg.Service.Version != defaultServiceVersion {
		t.Fatalf("service version = %q, want %q", cfg.Service.Version, defaultServiceVersion)
	}
	if cfg.Database.URL != "" {
		t.Fatalf("database URL = %q, want empty", cfg.Database.URL)
	}
	if cfg.Database.MaxOpenConns != 10 {
		t.Fatalf("database max open conns = %d, want 10", cfg.Database.MaxOpenConns)
	}
	if cfg.Database.MaxIdleConns != 5 {
		t.Fatalf("database max idle conns = %d, want 5", cfg.Database.MaxIdleConns)
	}
	if cfg.LogLevel != slog.LevelInfo {
		t.Fatalf("log level = %s, want info", cfg.LogLevel)
	}
}

func TestLoadFromLookupOverrides(t *testing.T) {
	values := map[string]string{
		"CAPCOM_HTTP_ADDR":                "127.0.0.1:9090",
		"CAPCOM_HTTP_READ_HEADER_TIMEOUT": "2s",
		"CAPCOM_HTTP_SHUTDOWN_TIMEOUT":    "3s",
		"CAPCOM_SERVICE_VERSION":          "test-version",
		"CAPCOM_LOG_LEVEL":                "debug",
		"CAPCOM_DATABASE_URL":             "postgres://capcom:capcom@localhost:5432/capcom?sslmode=disable",
		"CAPCOM_DATABASE_MAX_OPEN_CONNS":  "12",
		"CAPCOM_DATABASE_MAX_IDLE_CONNS":  "6",
	}

	cfg, err := LoadFromLookup(func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	})
	if err != nil {
		t.Fatalf("LoadFromLookup returned error: %v", err)
	}

	if cfg.HTTP.Addr != "127.0.0.1:9090" {
		t.Fatalf("HTTP addr = %q", cfg.HTTP.Addr)
	}
	if cfg.HTTP.ReadHeaderTimeout != 2*time.Second {
		t.Fatalf("read header timeout = %s", cfg.HTTP.ReadHeaderTimeout)
	}
	if cfg.HTTP.ShutdownTimeout != 3*time.Second {
		t.Fatalf("shutdown timeout = %s", cfg.HTTP.ShutdownTimeout)
	}
	if cfg.Service.Version != "test-version" {
		t.Fatalf("service version = %q", cfg.Service.Version)
	}
	if cfg.Database.URL != "postgres://capcom:capcom@localhost:5432/capcom?sslmode=disable" {
		t.Fatalf("database URL = %q", cfg.Database.URL)
	}
	if cfg.Database.MaxOpenConns != 12 {
		t.Fatalf("database max open conns = %d", cfg.Database.MaxOpenConns)
	}
	if cfg.Database.MaxIdleConns != 6 {
		t.Fatalf("database max idle conns = %d", cfg.Database.MaxIdleConns)
	}
	if cfg.LogLevel != slog.LevelDebug {
		t.Fatalf("log level = %s", cfg.LogLevel)
	}
}

func TestLoadFromLookupRejectsInvalidDuration(t *testing.T) {
	_, err := LoadFromLookup(func(key string) (string, bool) {
		if key == "CAPCOM_HTTP_SHUTDOWN_TIMEOUT" {
			return "nope", true
		}
		return "", false
	})
	if err == nil {
		t.Fatal("LoadFromLookup returned nil error")
	}
}

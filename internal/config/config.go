package config

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultHTTPAddr          = ":8080"
	defaultReadHeaderTimeout = 5 * time.Second
	defaultShutdownTimeout   = 10 * time.Second
	defaultServiceVersion    = "dev"
)

type Config struct {
	HTTP     HTTPConfig
	Database DatabaseConfig
	Secrets  SecretConfig
	Security SecurityConfig
	Service  ServiceConfig
	Sync     SyncConfig
	LogLevel slog.Level
}

type SyncConfig struct {
	WorkerEnabled    bool
	WorkerTick       time.Duration
	MaxConcurrency   int
	RequestTimeout   time.Duration
	MissingThreshold int
}

type SecretConfig struct {
	Key []byte
}

type SecurityConfig struct {
	AdminToken string
}

type HTTPConfig struct {
	Addr              string
	ReadHeaderTimeout time.Duration
	ShutdownTimeout   time.Duration
}

type ServiceConfig struct {
	Version string
}

type DatabaseConfig struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func Load() (Config, error) {
	dotEnv, err := loadDotEnv(".env")
	if err != nil {
		return Config{}, err
	}
	return LoadFromLookup(mergedLookup(dotEnv, os.LookupEnv))
}

func LoadFromLookup(lookup func(string) (string, bool)) (Config, error) {
	secretKey, err := secretKeyEnv(lookup, "CAPCOM_SECRET_KEY")
	if err != nil {
		return Config{}, err
	}

	readHeaderTimeout, err := durationEnv(lookup, "CAPCOM_HTTP_READ_HEADER_TIMEOUT", defaultReadHeaderTimeout)
	if err != nil {
		return Config{}, err
	}

	shutdownTimeout, err := durationEnv(lookup, "CAPCOM_HTTP_SHUTDOWN_TIMEOUT", defaultShutdownTimeout)
	if err != nil {
		return Config{}, err
	}

	logLevel, err := logLevelEnv(lookup, "CAPCOM_LOG_LEVEL", slog.LevelInfo)
	if err != nil {
		return Config{}, err
	}

	connMaxLifetime, err := durationEnv(lookup, "CAPCOM_DATABASE_CONN_MAX_LIFETIME", 30*time.Minute)
	if err != nil {
		return Config{}, err
	}

	maxOpenConns, err := intEnv(lookup, "CAPCOM_DATABASE_MAX_OPEN_CONNS", 10)
	if err != nil {
		return Config{}, err
	}

	maxIdleConns, err := intEnv(lookup, "CAPCOM_DATABASE_MAX_IDLE_CONNS", 5)
	if err != nil {
		return Config{}, err
	}
	workerTick, err := durationEnv(lookup, "CAPCOM_SYNC_WORKER_TICK", 5*time.Second)
	if err != nil {
		return Config{}, err
	}
	requestTimeout, err := durationEnv(lookup, "CAPCOM_SYNC_REQUEST_TIMEOUT", 30*time.Second)
	if err != nil {
		return Config{}, err
	}
	maxConcurrency, err := positiveIntEnv(lookup, "CAPCOM_SYNC_MAX_CONCURRENCY", 4)
	if err != nil {
		return Config{}, err
	}
	missingThreshold, err := positiveIntEnv(lookup, "CAPCOM_SYNC_MISSING_THRESHOLD", 3)
	if err != nil {
		return Config{}, err
	}
	workerEnabled, err := boolEnv(lookup, "CAPCOM_SYNC_WORKER_ENABLED", true)
	if err != nil {
		return Config{}, err
	}

	return Config{
		HTTP: HTTPConfig{
			Addr:              stringEnv(lookup, "CAPCOM_HTTP_ADDR", defaultHTTPAddr),
			ReadHeaderTimeout: readHeaderTimeout,
			ShutdownTimeout:   shutdownTimeout,
		},
		Database: DatabaseConfig{
			URL:             stringEnv(lookup, "CAPCOM_DATABASE_URL", ""),
			MaxOpenConns:    maxOpenConns,
			MaxIdleConns:    maxIdleConns,
			ConnMaxLifetime: connMaxLifetime,
		},
		Secrets: SecretConfig{Key: secretKey},
		Security: SecurityConfig{
			AdminToken: stringEnv(lookup, "CAPCOM_ADMIN_TOKEN", ""),
		},
		Service: ServiceConfig{
			Version: stringEnv(lookup, "CAPCOM_SERVICE_VERSION", defaultServiceVersion),
		},
		Sync: SyncConfig{
			WorkerEnabled:    workerEnabled,
			WorkerTick:       workerTick,
			MaxConcurrency:   maxConcurrency,
			RequestTimeout:   requestTimeout,
			MissingThreshold: missingThreshold,
		},
		LogLevel: logLevel,
	}, nil
}

func boolEnv(lookup func(string) (string, bool), key string, fallback bool) (bool, error) {
	value, ok := lookup(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		return false, fmt.Errorf("parse %s: %w", key, err)
	}
	return parsed, nil
}

func positiveIntEnv(lookup func(string) (string, bool), key string, fallback int) (int, error) {
	value, err := intEnv(lookup, key, fallback)
	if err != nil {
		return 0, err
	}
	if value <= 0 {
		return 0, fmt.Errorf("%s must be positive", key)
	}
	return value, nil
}

func secretKeyEnv(lookup func(string) (string, bool), key string) ([]byte, error) {
	value, ok := lookup(key)
	if !ok || strings.TrimSpace(value) == "" {
		return nil, nil
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(value))
	if err != nil {
		return nil, fmt.Errorf("parse %s as base64: %w", key, err)
	}
	if len(decoded) != 32 {
		return nil, fmt.Errorf("%s must decode to exactly 32 bytes", key)
	}
	return decoded, nil
}

func stringEnv(lookup func(string) (string, bool), key string, fallback string) string {
	value, ok := lookup(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func durationEnv(lookup func(string) (string, bool), key string, fallback time.Duration) (time.Duration, error) {
	value, ok := lookup(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback, nil
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	if duration <= 0 {
		return 0, fmt.Errorf("%s must be positive", key)
	}
	return duration, nil
}

func intEnv(lookup func(string) (string, bool), key string, fallback int) (int, error) {
	value, ok := lookup(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback, nil
	}

	var parsed int
	if _, err := fmt.Sscanf(value, "%d", &parsed); err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	if parsed < 0 {
		return 0, fmt.Errorf("%s must be zero or positive", key)
	}
	return parsed, nil
}

func logLevelEnv(lookup func(string) (string, bool), key string, fallback slog.Level) (slog.Level, error) {
	value, ok := lookup(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback, nil
	}

	switch strings.ToLower(strings.TrimSpace(value)) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("%s must be one of debug, info, warn, error", key)
	}
}

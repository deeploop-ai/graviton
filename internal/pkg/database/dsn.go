package database

import (
	"fmt"
	"os"
)

// SourceFromEnv resolves the Postgres DSN from environment variables.
// It prefers GRAVITON_DATA_DATABASE_SOURCE and falls back to POSTGRES_* compose vars.
func SourceFromEnv() string {
	if dsn := os.Getenv("GRAVITON_DATA_DATABASE_SOURCE"); dsn != "" {
		return dsn
	}
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		envOr("POSTGRES_USER", "graviton"),
		envOr("POSTGRES_PASSWORD", "graviton"),
		envOr("POSTGRES_HOST", "127.0.0.1"),
		envOr("POSTGRES_PORT", "5432"),
		envOr("POSTGRES_DB", "graviton"),
	)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

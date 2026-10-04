package database_test

import (
	"os"
	"testing"
	"time"

	"pkg/database"
)

func TestSanitizeDSN(t *testing.T) {
	raw := "postgres://user:pass@localhost:5432/graphfolio?sslmode=disable&search_path=portfolio&x-migrations-table=schema_migrations&x-custom=123"
	clean, err := database.SanitizeDSN(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if clean == raw {
		t.Errorf("SanitizeDSN should have removed x-* query params")
	}
	if clean != "postgres://user:pass@localhost:5432/graphfolio?search_path=portfolio&sslmode=disable" &&
		clean != "postgres://user:pass@localhost:5432/graphfolio?sslmode=disable&search_path=portfolio" {
		t.Errorf("got %q, unexpected stripped query string", clean)
	}
}

func TestConfigFromEnv(t *testing.T) {
	const key = "TEST_PORTFOLIO_DB_URL_VAL"
	os.Setenv(key, "postgres://u:p@localhost:5432/testdb")
	os.Setenv("DB_MAX_CONns", "25")
	os.Setenv("DB_MAX_CONNS", "25")
	os.Setenv("DB_MAX_CONN_LIFETIME", "2h")
	defer func() {
		os.Unsetenv(key)
		os.Unsetenv("DB_MAX_CONNS")
		os.Unsetenv("DB_MAX_CONN_LIFETIME")
	}()

	cfg, err := database.ConfigFromEnv(key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DSN != "postgres://u:p@localhost:5432/testdb" {
		t.Errorf("got DSN %q", cfg.DSN)
	}
	if cfg.MaxConns != 25 {
		t.Errorf("got MaxConns %d, want 25", cfg.MaxConns)
	}
	if cfg.MaxConnLifetime != 2*time.Hour {
		t.Errorf("got MaxConnLifetime %v, want 2h", cfg.MaxConnLifetime)
	}
}

func TestConfigFromEnvMissing(t *testing.T) {
	os.Unsetenv("DEFINITELY_NOT_SET_ENV_VAR")
	os.Unsetenv("DATABASE_URL")

	_, err := database.ConfigFromEnv("DEFINITELY_NOT_SET_ENV_VAR")
	if err == nil {
		t.Error("expected error for missing environment variable, got nil")
	}
}

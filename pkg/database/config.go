package database

import (
	"bufio"
	"bytes"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Config holds settings for a PostgreSQL connection pool.
type Config struct {
	DSN               string
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
	ConnectTimeout    time.Duration
}

// DefaultConfig returns recommended baseline pool settings.
func DefaultConfig(dsn string) Config {
	return Config{
		DSN:               dsn,
		MaxConns:          10,
		MinConns:          2,
		MaxConnLifetime:   time.Hour,
		MaxConnIdleTime:   30 * time.Minute,
		HealthCheckPeriod: time.Minute,
		ConnectTimeout:    5 * time.Second,
	}
}

// ConfigFromEnv constructs a Config using the given environment variable name for the DSN,
// with optional overrides for pool parameters (e.g. DB_MAX_CONNS).
// If the variable is unset, it automatically searches parent directories for a .env file.
func ConfigFromEnv(dsnEnvKey string) (Config, error) {
	dsn := os.Getenv(dsnEnvKey)
	if dsn == "" {
		loadDotEnv()
		dsn = os.Getenv(dsnEnvKey)
	}
	if dsn == "" {
		// Fallback to generic DATABASE_URL
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		return Config{}, fmt.Errorf("database: environment variable %q (or DATABASE_URL) is empty", dsnEnvKey)
	}

	cfg := DefaultConfig(dsn)

	if v := os.Getenv("DB_MAX_CONNS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.MaxConns = int32(n)
		}
	}
	if v := os.Getenv("DB_MIN_CONNS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			cfg.MinConns = int32(n)
		}
	}
	if v := os.Getenv("DB_MAX_CONN_LIFETIME"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.MaxConnLifetime = d
		}
	}
	if v := os.Getenv("DB_MAX_CONN_IDLE_TIME"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.MaxConnIdleTime = d
		}
	}

	return cfg, nil
}

// SanitizeDSN strips migration-tool-specific query parameters (such as `x-*`)
// so that the URL can be cleanly parsed by pgxpool.
func SanitizeDSN(rawDSN string) (string, error) {
	u, err := url.Parse(rawDSN)
	if err != nil {
		return "", fmt.Errorf("database: failed to parse DSN: %w", err)
	}

	q := u.Query()
	for k := range q {
		if strings.HasPrefix(strings.ToLower(k), "x-") {
			q.Del(k)
		}
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// loadDotEnv walks up the directory tree looking for a .env file and sets any unset environment variables.
func loadDotEnv() {
	dir, err := os.Getwd()
	if err != nil {
		return
	}
	for i := 0; i < 5; i++ {
		envPath := filepath.Join(dir, ".env")
		if data, err := os.ReadFile(envPath); err == nil {
			parseAndSetEnv(data)
			return
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
}

func parseAndSetEnv(data []byte) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(parts[1])
			if len(v) >= 2 && ((v[0] == '"' && v[len(v)-1] == '"') || (v[0] == '\'' && v[len(v)-1] == '\'')) {
				v = v[1 : len(v)-1]
			}
			if os.Getenv(k) == "" {
				_ = os.Setenv(k, v)
			}
		}
	}
}

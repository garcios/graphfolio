package database_test

import (
	"context"
	"os"
	"testing"
	"time"

	"pkg/database"

	"github.com/shopspring/decimal"
)

func TestPostgresIntegration(t *testing.T) {
	testDSN := os.Getenv("TEST_PORTFOLIO_DB_URL")
	if testDSN == "" {
		t.Skip("TEST_PORTFOLIO_DB_URL not set; skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg := database.DefaultConfig(testDSN)
	pool, err := database.NewPool(ctx, cfg)
	if err != nil {
		t.Skipf("cannot connect to test postgres (%v); skipping integration test", err)
	}
	defer pool.Close()

	// 1. Verify schema search_path is set to portfolio
	var schema string
	err = pool.QueryRow(ctx, "SELECT current_schema()").Scan(&schema)
	if err != nil {
		t.Fatalf("failed to query current schema: %v", err)
	}
	if schema != "portfolio" {
		t.Errorf("got current_schema() = %q, want 'portfolio'", schema)
	}

	// 2. Verify UUIDv7 function exists and generates valid version 7 UUIDs
	var generatedUUID string
	err = pool.QueryRow(ctx, "SELECT uuidv7()::text").Scan(&generatedUUID)
	if err != nil {
		t.Fatalf("failed to call uuidv7(): %v", err)
	}
	if len(generatedUUID) != 36 {
		t.Errorf("expected 36-char UUID, got %q", generatedUUID)
	}
	// In UUIDv7, char at index 14 is '7'
	if generatedUUID[14] != '7' {
		t.Errorf("expected UUID version 7, got %q", generatedUUID)
	}

	// 3. Verify shopspring decimal round-trip without floating-point inaccuracies
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback(ctx)

	// Clean insert within rolled back tx
	_, err = tx.Exec(ctx, `
		INSERT INTO portfolio.currencies (code, name, minor_units)
		VALUES ('TST', 'Test Currency', 2)
		ON CONFLICT (code) DO NOTHING
	`)
	if err != nil {
		t.Fatalf("failed to insert test currency: %v", err)
	}

	// Test inserting and reading exact decimal
	dIn := decimal.RequireFromString("123456.789012")
	var dOut decimal.Decimal
	err = tx.QueryRow(ctx, "SELECT $1::numeric", dIn).Scan(&dOut)
	if err != nil {
		t.Fatalf("failed to roundtrip decimal: %v", err)
	}
	if !dIn.Equal(dOut) {
		t.Errorf("decimal mismatch: got %s, want %s", dOut, dIn)
	}
}

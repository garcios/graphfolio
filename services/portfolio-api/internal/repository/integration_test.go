package repository_test

import (
	"context"
	"os"
	"testing"
	"time"

	"pkg/database"
	"portfolio-api/internal/repository"
)

func TestPostgresRepositoryQueries(t *testing.T) {
	dsn := os.Getenv("TEST_PORTFOLIO_DB_URL")
	if dsn == "" {
		dsn = os.Getenv("PORTFOLIO_DB_URL")
	}
	if dsn == "" {
		t.Skip("No PORTFOLIO_DB_URL set; skipping repository integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := database.NewPool(ctx, database.DefaultConfig(dsn))
	if err != nil {
		t.Skipf("cannot connect to postgres (%v); skipping", err)
	}
	defer pool.Close()

	repo := repository.NewPostgresRepository(pool)

	// 1. FindPortfolioByUser with demo user "1"
	p, err := repo.FindPortfolioByUser(ctx, "1")
	if err != nil {
		t.Fatalf("failed to find portfolio for user '1': %v", err)
	}
	if p.BaseCurrency != "USD" {
		t.Errorf("got base currency %q, want USD", p.BaseCurrency)
	}

	// 2. GetHoldingsWithMarketData
	holdings, err := repo.GetHoldingsWithMarketData(ctx, p.ID)
	if err != nil {
		t.Fatalf("failed to get holdings: %v", err)
	}
	if len(holdings) == 0 {
		t.Errorf("expected seeded holdings, got 0")
	}

	foundAAPL := false
	for _, h := range holdings {
		if h.Ticker == "AAPL" {
			foundAAPL = true
			if !h.LatestPrice.Equal(h.PrevPrice) && !h.LatestPrice.IsPositive() {
				t.Errorf("expected positive AAPL price, got %s", h.LatestPrice)
			}
		}
	}
	if !foundAAPL {
		t.Errorf("did not find AAPL in holdings")
	}

	// 3. GetCashBalances
	cash, err := repo.GetCashBalances(ctx, p.ID)
	if err != nil {
		t.Fatalf("failed to get cash balances: %v", err)
	}
	if len(cash) == 0 {
		t.Errorf("expected cash balance, got 0")
	}

	// 4. GetLatestValuation
	val, err := repo.GetLatestValuation(ctx, p.ID)
	if err != nil {
		t.Fatalf("failed to get valuation: %v", err)
	}
	if val == nil || !val.TWRIndex.IsPositive() {
		t.Errorf("expected valid TWRIndex in valuation")
	}
}

package repository_test

import (
	"context"
	"os"
	"testing"
	"time"

	"pkg/database"
	"portfolio-api/internal/marketdata"
	"portfolio-api/internal/repository"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
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

	// 5. ListActiveCurrencies
	currencies, err := repo.ListActiveCurrencies(ctx)
	if err != nil {
		t.Fatalf("failed to list active currencies: %v", err)
	}
	if len(currencies) == 0 {
		t.Errorf("expected active currencies, got 0")
	}

	// 6. BatchUpsertInstrumentPrices with Conflict Upsert
	testDate := time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC)
	prices := []marketdata.PriceRecord{
		{
			Symbol:     "AAPL",
			PriceDate:  testDate,
			ClosePrice: decimal.RequireFromString("230.50"),
			Source:     "batch_test",
		},
		{
			Symbol:     "MSFT",
			PriceDate:  testDate,
			ClosePrice: decimal.RequireFromString("435.00"),
			Source:     "batch_test",
		},
	}

	if err := repo.BatchUpsertInstrumentPrices(ctx, prices); err != nil {
		t.Fatalf("initial BatchUpsertInstrumentPrices failed: %v", err)
	}

	// Re-upsert with updated prices to verify idempotent conflict handling
	prices[0].ClosePrice = decimal.RequireFromString("231.00")
	if err := repo.BatchUpsertInstrumentPrices(ctx, prices); err != nil {
		t.Fatalf("second BatchUpsertInstrumentPrices failed: %v", err)
	}

	// 7. BatchUpsertFXRates with Conflict Upsert
	fxRecords := []marketdata.FXRecord{
		{
			BaseCurrency:  "EUR",
			QuoteCurrency: "USD",
			RateDate:      testDate,
			Rate:          decimal.RequireFromString("1.0850000000"),
			Source:        "batch_test",
		},
		{
			BaseCurrency:  "EUR",
			QuoteCurrency: "GBP",
			RateDate:      testDate,
			Rate:          decimal.RequireFromString("0.8550000000"),
			Source:        "batch_test",
		},
	}

	if err := repo.BatchUpsertFXRates(ctx, fxRecords); err != nil {
		t.Fatalf("initial BatchUpsertFXRates failed: %v", err)
	}

	// Re-upsert with updated rates
	fxRecords[0].Rate = decimal.RequireFromString("1.0860000000")
	if err := repo.BatchUpsertFXRates(ctx, fxRecords); err != nil {
		t.Fatalf("second BatchUpsertFXRates failed: %v", err)
	}

	// Verify rate retrieval
	rate, err := repo.GetFXRate(ctx, "EUR", "USD")
	if err != nil {
		t.Fatalf("failed to get updated FX rate: %v", err)
	}
	if !rate.Equal(decimal.RequireFromString("1.0860000000")) {
		t.Errorf("got rate %s, want 1.0860000000", rate.String())
	}

	// 5. HasPricesForRange
	aaplInst, err := repo.FindInstrumentBySymbol(ctx, "AAPL")
	if err != nil {
		t.Fatalf("FindInstrumentBySymbol failed: %v", err)
	}

	hasPrices, err := repo.HasPricesForRange(ctx, aaplInst.ID, testDate.AddDate(0, 0, -1), testDate)
	if err != nil {
		t.Fatalf("HasPricesForRange failed: %v", err)
	}
	if !hasPrices {
		t.Errorf("expected HasPricesForRange to be true for AAPL")
	}

	hasNoPrices, err := repo.HasPricesForRange(ctx, uuid.New(), testDate.AddDate(0, 0, -1), testDate)
	if err != nil {
		t.Fatalf("HasPricesForRange failed for unknown instrument: %v", err)
	}
	if hasNoPrices {
		t.Errorf("expected HasPricesForRange to be false for non-existent instrument")
	}
}

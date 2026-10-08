package repository_test

import (
	"context"
	"os"
	"testing"
	"time"

	"pkg/database"
	"portfolio-api/internal/domain"
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

	// 8. UpsertValuationsBatch
	vDate1 := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	vDate2 := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	vDate3 := time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC)

	batchVals := []domain.PortfolioValuation{
		{
			PortfolioID:     p.ID,
			ValuationDate:   vDate1,
			MarketValueBase: decimal.RequireFromString("10000.00"),
			CashValueBase:   decimal.RequireFromString("2000.00"),
			NetFlowBase:     decimal.Zero,
			DailyReturn:     decimal.Zero,
			TWRIndex:        decimal.RequireFromString("1.000000000000"),
		},
		{
			PortfolioID:     p.ID,
			ValuationDate:   vDate2,
			MarketValueBase: decimal.RequireFromString("10500.00"),
			CashValueBase:   decimal.RequireFromString("2000.00"),
			NetFlowBase:     decimal.Zero,
			DailyReturn:     decimal.RequireFromString("0.041666666667"),
			TWRIndex:        decimal.RequireFromString("1.041666666667"),
		},
		{
			PortfolioID:     p.ID,
			ValuationDate:   vDate3,
			MarketValueBase: decimal.RequireFromString("11000.00"),
			CashValueBase:   decimal.RequireFromString("2000.00"),
			NetFlowBase:     decimal.Zero,
			DailyReturn:     decimal.RequireFromString("0.040000000000"),
			TWRIndex:        decimal.RequireFromString("1.083333333334"),
		},
	}

	if err := repo.UpsertValuationsBatch(ctx, batchVals); err != nil {
		t.Fatalf("UpsertValuationsBatch failed: %v", err)
	}

	// 9. GetLatestValuationBefore
	latestBefore, err := repo.GetLatestValuationBefore(ctx, p.ID, vDate3)
	if err != nil {
		t.Fatalf("GetLatestValuationBefore failed: %v", err)
	}
	if latestBefore == nil {
		t.Fatalf("expected valuation before %s, got nil", vDate3)
	}
	if !latestBefore.ValuationDate.Equal(vDate2) {
		t.Errorf("got valuation date %v, want %v", latestBefore.ValuationDate, vDate2)
	}

	// 10. GetHistoricalPriceMatrix with LOCF across weekend
	priceMatrix, err := repo.GetHistoricalPriceMatrix(ctx, []uuid.UUID{aaplInst.ID}, vDate2, vDate2.AddDate(0, 0, 3))
	if err != nil {
		t.Fatalf("GetHistoricalPriceMatrix failed: %v", err)
	}
	if len(priceMatrix[aaplInst.ID]) < 4 {
		t.Errorf("expected at least 4 daily price points for AAPL with LOCF, got %d", len(priceMatrix[aaplInst.ID]))
	}
	// Saturday and Sunday should have the carried-forward price
	satDateStr := vDate2.AddDate(0, 0, 1).Format("2006-01-02")
	if price, ok := priceMatrix[aaplInst.ID][satDateStr]; !ok || !price.IsPositive() {
		t.Errorf("expected positive carried-forward price on Saturday %s, got %v", satDateStr, price)
	}

	// 11. GetHistoricalFXMatrix with LOCF
	fxMatrix, err := repo.GetHistoricalFXMatrix(ctx, []string{"EUR", "USD"}, "USD", vDate2, vDate2.AddDate(0, 0, 3))
	if err != nil {
		t.Fatalf("GetHistoricalFXMatrix failed: %v", err)
	}
	if usdRate, ok := fxMatrix["USD"][satDateStr]; !ok || !usdRate.Equal(decimal.NewFromInt(1)) {
		t.Errorf("expected USD rate to be 1.0, got %v", usdRate)
	}
	if eurRate, ok := fxMatrix["EUR"][satDateStr]; !ok || !eurRate.IsPositive() {
		t.Errorf("expected positive EUR rate on Saturday %s, got %v", satDateStr, eurRate)
	}

	// 12. DeleteValuationsFromDate
	if err := repo.DeleteValuationsFromDate(ctx, p.ID, vDate2); err != nil {
		t.Fatalf("DeleteValuationsFromDate failed: %v", err)
	}
	afterDelete, err := repo.GetLatestValuationBefore(ctx, p.ID, vDate3)
	if err != nil {
		t.Fatalf("GetLatestValuationBefore after delete failed: %v", err)
	}
	if afterDelete != nil && afterDelete.ValuationDate.Equal(vDate2) {
		t.Errorf("valuation for %v should have been deleted", vDate2)
	}
}

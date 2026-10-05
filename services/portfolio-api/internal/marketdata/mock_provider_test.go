package marketdata_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"portfolio-api/internal/marketdata"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestMockPriceProvider(t *testing.T) {
	ctx := context.Background()
	provider := marketdata.NewMockPriceProvider("TestPriceProvider")

	if provider.Name() != "TestPriceProvider" {
		t.Errorf("expected name 'TestPriceProvider', got: %s", provider.Name())
	}

	inst1 := marketdata.InstrumentRef{
		ID:       uuid.New(),
		Symbol:   "AAPL",
		Exchange: "NASDAQ",
		Currency: "USD",
	}
	inst2 := marketdata.InstrumentRef{
		ID:       uuid.New(),
		Symbol:   "MSFT",
		Exchange: "NASDAQ",
		Currency: "USD",
	}

	tradeDate := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	provider.SetPrice("AAPL", tradeDate, decimal.RequireFromString("228.45"))

	t.Run("fetch closing prices with explicit and default prices", func(t *testing.T) {
		records, err := provider.FetchClosingPrices(ctx, []marketdata.InstrumentRef{inst1, inst2}, tradeDate)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(records) != 2 {
			t.Fatalf("expected 2 records, got: %d", len(records))
		}

		if !records[0].ClosePrice.Equal(decimal.RequireFromString("228.45")) {
			t.Errorf("AAPL price = %s, want 228.45", records[0].ClosePrice.String())
		}
		// MSFT gets default price (100)
		if !records[1].ClosePrice.Equal(decimal.NewFromInt(100)) {
			t.Errorf("MSFT default price = %s, want 100", records[1].ClosePrice.String())
		}
	})

	t.Run("fetch historical prices filters out weekends", func(t *testing.T) {
		// Friday to Tuesday (5 days: Fri, Sat, Sun, Mon, Tue -> 3 weekdays)
		friday := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
		tuesday := time.Date(2026, 10, 6, 0, 0, 0, 0, time.UTC)

		records, err := provider.FetchHistoricalPrices(ctx, inst1, friday, tuesday)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(records) != 3 {
			t.Fatalf("expected 3 weekday records, got %d", len(records))
		}
		for _, r := range records {
			if r.PriceDate.Weekday() == time.Saturday || r.PriceDate.Weekday() == time.Sunday {
				t.Errorf("record contains weekend date: %v", r.PriceDate)
			}
		}
	})

	t.Run("error injection", func(t *testing.T) {
		injectedErr := errors.New("upstream vendor timeout")
		provider.SetError(injectedErr)

		_, err := provider.FetchClosingPrices(ctx, []marketdata.InstrumentRef{inst1}, tradeDate)
		if !errors.Is(err, injectedErr) {
			t.Errorf("expected injected error, got: %v", err)
		}

		_, errHist := provider.FetchHistoricalPrices(ctx, inst1, tradeDate, tradeDate)
		if !errors.Is(errHist, injectedErr) {
			t.Errorf("expected injected error, got: %v", errHist)
		}

		// Clear error
		provider.SetError(nil)
	})
}

func TestMockFXRateProvider(t *testing.T) {
	ctx := context.Background()
	provider := marketdata.NewMockFXRateProvider("TestFXProvider")

	tradeDate := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	provider.SetRate("EUR", "USD", tradeDate, decimal.RequireFromString("1.0850000000"))
	provider.SetRate("EUR", "GBP", tradeDate, decimal.RequireFromString("0.8550000000"))

	t.Run("fetch FX rates with triangulation fallback", func(t *testing.T) {
		pairs := []marketdata.CurrencyPair{
			{Base: "EUR", Quote: "USD"},
			{Base: "USD", Quote: "GBP"}, // Triangulated via EUR
		}

		records, err := provider.FetchFXRates(ctx, pairs, tradeDate)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(records) != 2 {
			t.Fatalf("expected 2 records, got: %d", len(records))
		}

		if !records[0].Rate.Equal(decimal.RequireFromString("1.0850000000")) {
			t.Errorf("EUR/USD rate = %s, want 1.0850000000", records[0].Rate.String())
		}
		expectedCross := decimal.RequireFromString("0.7880184332")
		if !records[1].Rate.Equal(expectedCross) {
			t.Errorf("USD/GBP triangulated rate = %s, want %s", records[1].Rate.String(), expectedCross.String())
		}
	})

	t.Run("fetch historical FX rates iterates over dates", func(t *testing.T) {
		from := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC)

		records, err := provider.FetchHistoricalFXRates(ctx, marketdata.CurrencyPair{Base: "EUR", Quote: "USD"}, from, to)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(records) != 3 {
			t.Fatalf("expected 3 days of rates, got: %d", len(records))
		}
	})

	t.Run("error injection", func(t *testing.T) {
		injectedErr := errors.New("network failure")
		provider.SetError(injectedErr)

		_, err := provider.FetchFXRates(ctx, []marketdata.CurrencyPair{{Base: "EUR", Quote: "USD"}}, tradeDate)
		if !errors.Is(err, injectedErr) {
			t.Errorf("expected injected error, got: %v", err)
		}
	})
}

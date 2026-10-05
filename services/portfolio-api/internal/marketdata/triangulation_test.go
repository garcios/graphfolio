package marketdata_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	"portfolio-api/internal/marketdata"

	"github.com/shopspring/decimal"
)

func TestTriangulationEngine(t *testing.T) {
	engine := marketdata.NewTriangulationEngine("EUR")

	// Ingest official ECB rates quoted against EUR
	rates := []marketdata.FXRecord{
		{BaseCurrency: "EUR", QuoteCurrency: "USD", Rate: decimal.RequireFromString("1.0850000000"), Source: "ecb"},
		{BaseCurrency: "EUR", QuoteCurrency: "GBP", Rate: decimal.RequireFromString("0.8550000000"), Source: "ecb"},
		{BaseCurrency: "EUR", QuoteCurrency: "JPY", Rate: decimal.RequireFromString("162.3000000000"), Source: "ecb"},
		{BaseCurrency: "EUR", QuoteCurrency: "CHF", Rate: decimal.RequireFromString("0.9450000000"), Source: "ecb"},
		{BaseCurrency: "EUR", QuoteCurrency: "AUD", Rate: decimal.RequireFromString("1.6520000000"), Source: "ecb"},
		{BaseCurrency: "EUR", QuoteCurrency: "CAD", Rate: decimal.RequireFromString("1.4850000000"), Source: "ecb"},
	}
	engine.AddRates(rates)

	t.Run("identity rate returns 1.0", func(t *testing.T) {
		currencies := []string{"EUR", "USD", "GBP", "JPY", "CHF"}
		for _, c := range currencies {
			rate, src, err := engine.GetRate(c, c)
			if err != nil {
				t.Fatalf("unexpected error for identity %s: %v", c, err)
			}
			if !rate.Equal(decimal.NewFromInt(1)) {
				t.Errorf("expected 1.0 for %s/%s, got: %s", c, c, rate.String())
			}
			if src != "identity" {
				t.Errorf("expected source 'identity', got: %s", src)
			}
		}
	})

	t.Run("direct rates match ECB quote", func(t *testing.T) {
		rate, src, err := engine.GetRate("EUR", "USD")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := decimal.RequireFromString("1.0850000000")
		if !rate.Equal(expected) {
			t.Errorf("EUR/USD = %s, want %s", rate.String(), expected.String())
		}
		if src != "ecb" {
			t.Errorf("source = %s, want 'ecb'", src)
		}
	})

	t.Run("inverse rates computed accurately", func(t *testing.T) {
		// USD -> EUR = 1 / 1.0850000000 = 0.9216589862
		rate, _, err := engine.GetRate("USD", "EUR")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := decimal.RequireFromString("0.9216589862")
		if !rate.Equal(expected) {
			t.Errorf("USD/EUR = %s, want %s", rate.String(), expected.String())
		}

		// GBP -> EUR = 1 / 0.8550000000 = 1.1695906433
		rateGBP, _, err := engine.GetRate("GBP", "EUR")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expectedGBP := decimal.RequireFromString("1.1695906433")
		if !rateGBP.Equal(expectedGBP) {
			t.Errorf("GBP/EUR = %s, want %s", rateGBP.String(), expectedGBP.String())
		}
	})

	t.Run("cross-rate triangulation via EUR anchor", func(t *testing.T) {
		tests := []struct {
			base  string
			quote string
			want  decimal.Decimal
		}{
			{
				base:  "USD",
				quote: "GBP",
				// 0.8550 / 1.0850 = 0.7880184332
				want: decimal.RequireFromString("0.7880184332"),
			},
			{
				base:  "GBP",
				quote: "USD",
				// 1.0850 / 0.8550 = 1.2690058480
				want: decimal.RequireFromString("1.2690058480"),
			},
			{
				base:  "USD",
				quote: "JPY",
				// 162.30 / 1.0850 = 149.5852534562
				want: decimal.RequireFromString("149.5852534562"),
			},
			{
				base:  "GBP",
				quote: "JPY",
				// 162.30 / 0.8550 = 189.8245614035
				want: decimal.RequireFromString("189.8245614035"),
			},
			{
				base:  "AUD",
				quote: "CAD",
				// 1.4850 / 1.6520 = 0.8989104116
				want: decimal.RequireFromString("0.8989104116"),
			},
		}

		for _, tt := range tests {
			t.Run(tt.base+"/"+tt.quote, func(t *testing.T) {
				rate, _, err := engine.GetRate(tt.base, tt.quote)
				if err != nil {
					t.Fatalf("unexpected triangulation error: %v", err)
				}
				if !rate.Equal(tt.want) {
					t.Errorf("%s/%s rate = %s, want %s", tt.base, tt.quote, rate.String(), tt.want.String())
				}
			})
		}
	})

	t.Run("unresolvable currency pair returns error", func(t *testing.T) {
		_, _, err := engine.GetRate("BRL", "ZAR")
		if err == nil {
			t.Fatal("expected error for unrecorded currency pair BRL/ZAR, got nil")
		}
		if !errors.Is(err, marketdata.ErrUnresolvableFXRate) {
			t.Errorf("expected ErrUnresolvableFXRate, got: %v", err)
		}

		_, _, errUSDToZAR := engine.GetRate("USD", "ZAR")
		if errUSDToZAR == nil {
			t.Fatal("expected error for half-missing currency pair USD/ZAR, got nil")
		}
		if !errors.Is(errUSDToZAR, marketdata.ErrUnresolvableFXRate) {
			t.Errorf("expected ErrUnresolvableFXRate, got: %v", errUSDToZAR)
		}
	})

	t.Run("triangulate batch pairs", func(t *testing.T) {
		reqDate := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
		pairs := []marketdata.CurrencyPair{
			{Base: "USD", Quote: "EUR"},
			{Base: "USD", Quote: "GBP"},
			{Base: "EUR", Quote: "JPY"},
		}

		results, err := engine.TriangulatePairs(pairs, reqDate)
		if err != nil {
			t.Fatalf("TriangulatePairs failed: %v", err)
		}
		if len(results) != 3 {
			t.Fatalf("expected 3 records, got: %d", len(results))
		}
		if results[0].BaseCurrency != "USD" || results[0].QuoteCurrency != "EUR" {
			t.Errorf("unexpected pair in result[0]: %s/%s", results[0].BaseCurrency, results[0].QuoteCurrency)
		}
		if results[0].RateDate != reqDate {
			t.Errorf("unexpected date: %v", results[0].RateDate)
		}
	})

	t.Run("concurrent access safety", func(t *testing.T) {
		var wg sync.WaitGroup
		for i := 0; i < 50; i++ {
			wg.Add(2)
			go func(idx int) {
				defer wg.Done()
				engine.AddRate("EUR", "SEK", decimal.RequireFromString("11.5000000000"), "concurrency_test")
			}(i)
			go func() {
				defer wg.Done()
				_, _, _ = engine.GetRate("USD", "JPY")
				_, _, _ = engine.GetRate("EUR", "SEK")
			}()
		}
		wg.Wait()
	})
}

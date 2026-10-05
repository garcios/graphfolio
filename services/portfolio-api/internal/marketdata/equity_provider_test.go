package marketdata

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const sampleTwelveDataDailyJSON = `{
  "meta": {
    "symbol": "AAPL",
    "interval": "1day",
    "currency": "USD",
    "exchange": "NASDAQ"
  },
  "values": [
    {
      "datetime": "2026-10-02",
      "open": "226.50000",
      "high": "228.15000",
      "low": "225.80000",
      "close": "227.63000",
      "volume": "45123000"
    },
    {
      "datetime": "2026-10-01",
      "open": "225.00000",
      "high": "227.00000",
      "low": "224.50000",
      "close": "226.10000",
      "volume": "41234000"
    }
  ],
  "status": "ok"
}`

const sampleTwelveDataHistoricalJSON = `{
  "meta": {
    "symbol": "AAPL",
    "interval": "1day",
    "currency": "USD",
    "exchange": "NASDAQ"
  },
  "values": [
    {
      "datetime": "2026-10-02",
      "open": "226.50000",
      "high": "228.15000",
      "low": "225.80000",
      "close": "227.63000",
      "volume": "45123000"
    },
    {
      "datetime": "2026-10-01",
      "open": "225.00000",
      "high": "227.00000",
      "low": "224.50000",
      "close": "226.10000",
      "volume": "41234000"
    },
    {
      "datetime": "2026-09-30",
      "open": "224.00000",
      "high": "225.50000",
      "low": "223.80000",
      "close": "225.00000",
      "volume": "39876000"
    }
  ],
  "status": "ok"
}`

const sampleYahooChartJSON = `{
  "chart": {
    "result": [
      {
        "meta": {
          "currency": "USD",
          "symbol": "MSFT"
        },
        "timestamp": [1790860800, 1790947200],
        "indicators": {
          "quote": [
            {
              "close": [430.50, 432.25],
              "open": [429.00, 431.00],
              "high": [431.50, 433.00],
              "low": [428.50, 430.00]
            }
          ]
        }
      }
    ],
    "error": null
  }
}`

func TestTwelveDataProvider(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("symbol") == "FAIL" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"code": 429, "message": "You have run out of API credits", "status": "error"}`))
			return
		}
		if r.URL.Query().Get("start_date") != "" {
			_, _ = w.Write([]byte(sampleTwelveDataHistoricalJSON))
			return
		}
		_, _ = w.Write([]byte(sampleTwelveDataDailyJSON))
	}))
	defer ts.Close()

	instID := uuid.New()
	inst := InstrumentRef{
		ID:       instID,
		Symbol:   "AAPL",
		Exchange: "NASDAQ",
		Currency: "USD",
	}

	provider := NewTwelveDataProvider(TwelveDataConfig{
		BaseURL: ts.URL,
		APIKey:  "test-api-key",
	})

	t.Run("Name returns Twelve Data", func(t *testing.T) {
		if provider.Name() != "Twelve Data" {
			t.Errorf("expected Twelve Data, got %s", provider.Name())
		}
	})

	t.Run("FetchClosingPrices retrieves exact close mark", func(t *testing.T) {
		date, _ := time.Parse("2006-01-02", "2026-10-02")
		records, err := provider.FetchClosingPrices(context.Background(), []InstrumentRef{inst}, date)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(records) != 1 {
			t.Fatalf("expected 1 record, got %d", len(records))
		}

		rec := records[0]
		if rec.Symbol != "AAPL" {
			t.Errorf("expected symbol AAPL, got %s", rec.Symbol)
		}
		expectedClose := decimal.RequireFromString("227.63000")
		if !rec.ClosePrice.Equal(expectedClose) {
			t.Errorf("expected close %s, got %s", expectedClose, rec.ClosePrice)
		}
		if rec.Source != "twelvedata" {
			t.Errorf("expected source twelvedata, got %s", rec.Source)
		}
	})

	t.Run("FetchHistoricalPrices parses and sorts ascending", func(t *testing.T) {
		from, _ := time.Parse("2006-01-02", "2026-09-30")
		to, _ := time.Parse("2006-01-02", "2026-10-02")

		records, err := provider.FetchHistoricalPrices(context.Background(), inst, from, to)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(records) != 3 {
			t.Fatalf("expected 3 records, got %d", len(records))
		}

		if records[0].PriceDate.Format("2006-01-02") != "2026-09-30" {
			t.Errorf("expected first date 2026-09-30, got %s", records[0].PriceDate.Format("2006-01-02"))
		}
		if !records[0].ClosePrice.Equal(decimal.RequireFromString("225.00000")) {
			t.Errorf("expected 225.00000, got %s", records[0].ClosePrice)
		}

		if records[2].PriceDate.Format("2006-01-02") != "2026-10-02" {
			t.Errorf("expected third date 2026-10-02, got %s", records[2].PriceDate.Format("2006-01-02"))
		}
	})

	t.Run("returns error on API quota limit", func(t *testing.T) {
		failInst := InstrumentRef{ID: uuid.New(), Symbol: "FAIL"}
		providerNoRetry := NewTwelveDataProvider(TwelveDataConfig{
			BaseURL: ts.URL,
			RetryCfg: RetryConfig{
				MaxRetries: 0,
			},
		})

		_, err := providerNoRetry.FetchClosingPrices(context.Background(), []InstrumentRef{failInst}, time.Now())
		if err == nil {
			t.Fatalf("expected error on 429 quota, got nil")
		}
	})
}

func TestYahooFinanceProvider(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/v8/finance/chart/UNKNOWN" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"chart":{"result":null,"error":{"code":"Not Found","description":"Symbol not found"}}}`))
			return
		}
		_, _ = w.Write([]byte(sampleYahooChartJSON))
	}))
	defer ts.Close()

	inst := InstrumentRef{
		ID:       uuid.New(),
		Symbol:   "MSFT",
		Exchange: "NASDAQ",
		Currency: "USD",
	}

	provider := NewYahooFinanceProvider(YahooFinanceConfig{
		BaseURL: ts.URL,
	})

	t.Run("Name returns Yahoo Finance", func(t *testing.T) {
		if provider.Name() != "Yahoo Finance" {
			t.Errorf("expected Yahoo Finance, got %s", provider.Name())
		}
	})

	t.Run("FetchClosingPrices preserves exact precision without float drift", func(t *testing.T) {
		records, err := provider.FetchClosingPrices(context.Background(), []InstrumentRef{inst}, time.Time{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(records) != 1 {
			t.Fatalf("expected 1 record, got %d", len(records))
		}

		rec := records[0]
		expected := decimal.RequireFromString("432.25")
		if !rec.ClosePrice.Equal(expected) {
			t.Errorf("expected %s, got %s", expected, rec.ClosePrice)
		}
		if rec.Source != "yahoo" {
			t.Errorf("expected source yahoo, got %s", rec.Source)
		}
	})

	t.Run("FetchClosingPrices reports error on unknown symbol", func(t *testing.T) {
		badInst := InstrumentRef{Symbol: "UNKNOWN"}
		providerNoRetry := NewYahooFinanceProvider(YahooFinanceConfig{
			BaseURL: ts.URL,
			RetryCfg: RetryConfig{
				MaxRetries: 0,
			},
		})

		_, err := providerNoRetry.FetchClosingPrices(context.Background(), []InstrumentRef{badInst}, time.Time{})
		if err == nil {
			t.Fatalf("expected error on unknown symbol, got nil")
		}
	})
}

func TestResilientPriceProvider(t *testing.T) {
	primaryMock := NewMockPriceProvider("Primary")
	primaryMock.SetDefaultPrice(decimal.Zero)
	fallbackMock := NewMockPriceProvider("Fallback")
	fallbackMock.SetDefaultPrice(decimal.Zero)

	inst1 := InstrumentRef{ID: uuid.New(), Symbol: "AAPL"}
	inst2 := InstrumentRef{ID: uuid.New(), Symbol: "MSFT"}
	tradeDate, _ := time.Parse("2006-01-02", "2026-10-02")

	// Primary has AAPL, but NOT MSFT
	primaryMock.SetPrice("AAPL", tradeDate, decimal.RequireFromString("227.63"))
	// Fallback has both
	fallbackMock.SetPrice("AAPL", tradeDate, decimal.RequireFromString("227.50"))
	fallbackMock.SetPrice("MSFT", tradeDate, decimal.RequireFromString("432.25"))

	composite := NewResilientPriceProvider(primaryMock, fallbackMock)

	t.Run("falls back seamlessly for missing symbol", func(t *testing.T) {
		records, err := composite.FetchClosingPrices(context.Background(), []InstrumentRef{inst1, inst2}, tradeDate)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(records) != 2 {
			t.Fatalf("expected 2 records, got %d", len(records))
		}

		var aaplRec, msftRec *PriceRecord
		for i := range records {
			if records[i].Symbol == "AAPL" {
				aaplRec = &records[i]
			}
			if records[i].Symbol == "MSFT" {
				msftRec = &records[i]
			}
		}

		if aaplRec == nil || msftRec == nil {
			t.Fatalf("missing AAPL or MSFT in returned records")
		}

		// AAPL came from primary (227.63)
		if !aaplRec.ClosePrice.Equal(decimal.RequireFromString("227.63")) {
			t.Errorf("expected AAPL close 227.63 from primary, got %s", aaplRec.ClosePrice)
		}

		// MSFT came from fallback (432.25)
		if !msftRec.ClosePrice.Equal(decimal.RequireFromString("432.25")) {
			t.Errorf("expected MSFT close 432.25 from fallback, got %s", msftRec.ClosePrice)
		}
	})

	t.Run("FetchHistoricalPrices falls back when primary returns error or empty", func(t *testing.T) {
		from := tradeDate.AddDate(0, 0, -5)
		to := tradeDate

		// Primary has no historical prices for MSFT
		records, err := composite.FetchHistoricalPrices(context.Background(), inst2, from, to)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(records) != 1 {
			t.Fatalf("expected 1 record from fallback, got %d", len(records))
		}
		if records[0].Symbol != "MSFT" {
			t.Errorf("expected MSFT record, got %s", records[0].Symbol)
		}
	})
}

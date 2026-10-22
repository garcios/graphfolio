package marketdata

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type mockTransport func(req *http.Request) (*http.Response, error)

func (m mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m(req)
}

const sampleYahooQuoteSummaryJSON = `{
  "quoteSummary": {
    "result": [
      {
        "calendarEvents": {
          "exDividendDate": { "raw": 1707436800, "fmt": "2024-02-09" },
          "dividendDate": { "raw": 1707955200, "fmt": "2024-02-15" }
        },
        "defaultKeyStatistics": {
          "lastDividendValue": { "raw": "0.24", "fmt": "0.24" }
        },
        "summaryDetail": {
          "currency": "USD",
          "dividendRate": { "raw": "0.96", "fmt": "0.96" },
          "exDividendDate": { "raw": 1707436800, "fmt": "2024-02-09" }
        }
      }
    ],
    "error": null
  }
}`

const sampleYahooChartWithDividendsJSON = `{
  "chart": {
    "result": [
      {
        "meta": { "currency": "USD", "symbol": "AAPL" },
        "timestamp": [1707436800],
        "indicators": {
          "quote": [{ "close": [188.0] }]
        },
        "events": {
          "dividends": {
            "1699540200": { "amount": "0.24", "date": 1699540200 },
            "1707436800": { "amount": "0.25", "date": 1707436800 }
          }
        }
      }
    ],
    "error": null
  }
}`

const sampleTwelveDataDividendsJSON = `{
  "meta": {
    "symbol": "AAPL",
    "name": "Apple Inc",
    "currency": "USD",
    "exchange": "NASDAQ"
  },
  "dividends": [
    {
      "ex_date": "2024-02-09",
      "payment_date": "2024-02-15",
      "record_date": "2024-02-12",
      "amount": "0.24"
    },
    {
      "ex_date": "2023-11-10",
      "payment_date": "2023-11-16",
      "record_date": "2023-11-13",
      "amount": "0.24"
    }
  ],
  "status": "ok"
}`

func newMockClient(body string, statusCode int) *http.Client {
	return &http.Client{
		Transport: mockTransport(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: statusCode,
				Status:     http.StatusText(statusCode),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(bytes.NewBufferString(body)),
			}, nil
		}),
	}
}

func TestYahooFinanceProvider_GetUpcomingDividends(t *testing.T) {
	client := newMockClient(sampleYahooQuoteSummaryJSON, http.StatusOK)
	provider := NewYahooFinanceProvider(YahooFinanceConfig{
		BaseURL:    "https://query1.finance.yahoo.com",
		HTTPClient: client,
	})

	inst := InstrumentRef{
		ID:       uuid.New(),
		Symbol:   "AAPL",
		Exchange: "NASDAQ",
		Currency: "USD",
	}

	events, err := provider.GetUpcomingDividends(context.Background(), inst)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 dividend event, got %d", len(events))
	}

	ev := events[0]
	if ev.AmountPerShare.Cmp(decimal.NewFromFloat(0.24)) != 0 {
		t.Errorf("expected DPS 0.24, got %v", ev.AmountPerShare)
	}
	expectedEx := time.Unix(1707436800, 0).UTC().Truncate(24 * time.Hour)
	if !ev.ExDate.Equal(expectedEx) {
		t.Errorf("expected ExDate %v, got %v", expectedEx, ev.ExDate)
	}
	expectedPay := time.Unix(1707955200, 0).UTC().Truncate(24 * time.Hour)
	if !ev.PaymentDate.Equal(expectedPay) {
		t.Errorf("expected PaymentDate %v, got %v", expectedPay, ev.PaymentDate)
	}
	if ev.Source != "yahoo" {
		t.Errorf("expected source yahoo, got %s", ev.Source)
	}
}

func TestYahooFinanceProvider_GetUpcomingDividends_ASXFranking(t *testing.T) {
	client := newMockClient(sampleYahooQuoteSummaryJSON, http.StatusOK)
	provider := NewYahooFinanceProvider(YahooFinanceConfig{
		BaseURL:    "https://query1.finance.yahoo.com",
		HTTPClient: client,
	})

	inst := InstrumentRef{
		ID:       uuid.New(),
		Symbol:   "BHP.AX",
		Exchange: "ASX",
		Currency: "AUD",
	}

	events, err := provider.GetUpcomingDividends(context.Background(), inst)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 dividend event, got %d", len(events))
	}
	if events[0].FrankingPct.Cmp(decimal.NewFromInt(100)) != 0 {
		t.Errorf("expected 100%% franking for ASX equity, got %v", events[0].FrankingPct)
	}
}

func TestYahooFinanceProvider_GetHistoricalDividends(t *testing.T) {
	client := newMockClient(sampleYahooChartWithDividendsJSON, http.StatusOK)
	provider := NewYahooFinanceProvider(YahooFinanceConfig{
		BaseURL:    "https://query1.finance.yahoo.com",
		HTTPClient: client,
	})

	inst := InstrumentRef{
		ID:       uuid.New(),
		Symbol:   "AAPL",
		Exchange: "NASDAQ",
		Currency: "USD",
	}

	points, err := provider.GetHistoricalDividends(context.Background(), inst, time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(points) != 2 {
		t.Fatalf("expected 2 historical points, got %d", len(points))
	}

	// Should be sorted chronologically
	if points[0].Amount.Cmp(decimal.NewFromFloat(0.24)) != 0 {
		t.Errorf("expected first amount 0.24, got %v", points[0].Amount)
	}
	if points[1].Amount.Cmp(decimal.NewFromFloat(0.25)) != 0 {
		t.Errorf("expected second amount 0.25, got %v", points[1].Amount)
	}
}

func TestTwelveDataProvider_GetUpcomingDividends(t *testing.T) {
	client := newMockClient(sampleTwelveDataDividendsJSON, http.StatusOK)
	provider := NewTwelveDataProvider(TwelveDataConfig{
		BaseURL:    "https://api.twelvedata.com",
		APIKey:     "test-key",
		HTTPClient: client,
	})

	inst := InstrumentRef{
		ID:       uuid.New(),
		Symbol:   "AAPL",
		Exchange: "NASDAQ",
		Currency: "USD",
	}

	events, err := provider.GetUpcomingDividends(context.Background(), inst)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("expected 2 dividend events, got %d", len(events))
	}

	if events[0].Source != "twelvedata" {
		t.Errorf("expected source twelvedata, got %s", events[0].Source)
	}
}

func TestTwelveDataProvider_GetHistoricalDividends(t *testing.T) {
	client := newMockClient(sampleTwelveDataDividendsJSON, http.StatusOK)
	provider := NewTwelveDataProvider(TwelveDataConfig{
		BaseURL:    "https://api.twelvedata.com",
		APIKey:     "test-key",
		HTTPClient: client,
	})

	inst := InstrumentRef{
		ID:       uuid.New(),
		Symbol:   "AAPL",
		Exchange: "NASDAQ",
		Currency: "USD",
	}

	points, err := provider.GetHistoricalDividends(context.Background(), inst, time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(points) != 2 {
		t.Fatalf("expected 2 historical points, got %d", len(points))
	}

	// Should be sorted ascending by date
	if points[0].Date.After(points[1].Date) {
		t.Errorf("expected points in ascending chronological order")
	}
}

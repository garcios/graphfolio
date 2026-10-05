package marketdata

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

const sampleECBDailyXML = `<?xml version="1.0" encoding="UTF-8"?>
<gesmes:Envelope xmlns:gesmes="http://www.gesmes.org/xml/2002-05-01" xmlns="http://www.ecb.int/vocabulary/2002-08-01/eurofxref">
	<gesmes:subject>Reference rates</gesmes:subject>
	<gesmes:Sender>
		<gesmes:name>European Central Bank</gesmes:name>
	</gesmes:Sender>
	<Cube>
		<Cube time="2026-10-02">
			<Cube currency="USD" rate="1.0850"/>
			<Cube currency="JPY" rate="162.30"/>
			<Cube currency="GBP" rate="0.8350"/>
			<Cube currency="CHF" rate="0.9420"/>
			<Cube currency="AUD" rate="1.6150"/>
		</Cube>
	</Cube>
</gesmes:Envelope>`

const sampleECBHistXML = `<?xml version="1.0" encoding="UTF-8"?>
<gesmes:Envelope xmlns:gesmes="http://www.gesmes.org/xml/2002-05-01" xmlns="http://www.ecb.int/vocabulary/2002-08-01/eurofxref">
	<Cube>
		<Cube time="2026-10-02">
			<Cube currency="USD" rate="1.0850"/>
			<Cube currency="GBP" rate="0.8350"/>
		</Cube>
		<Cube time="2026-10-01">
			<Cube currency="USD" rate="1.0820"/>
			<Cube currency="GBP" rate="0.8330"/>
		</Cube>
		<Cube time="2026-09-30">
			<Cube currency="USD" rate="1.0800"/>
			<Cube currency="GBP" rate="0.8310"/>
		</Cube>
	</Cube>
</gesmes:Envelope>`

func TestParseECBXML(t *testing.T) {
	t.Run("successfully parses daily XML", func(t *testing.T) {
		res, err := ParseECBXML([]byte(sampleECBDailyXML))
		if err != nil {
			t.Fatalf("unexpected parse error: %v", err)
		}

		records, ok := res["2026-10-02"]
		if !ok {
			t.Fatalf("expected date 2026-10-02 in parsed results")
		}

		if len(records) != 5 {
			t.Fatalf("expected 5 rate records, got %d", len(records))
		}

		// Find USD record
		var usdRec *FXRecord
		for i := range records {
			if records[i].QuoteCurrency == "USD" {
				usdRec = &records[i]
				break
			}
		}

		if usdRec == nil {
			t.Fatalf("USD record not found")
		}

		expectedUSD := decimal.RequireFromString("1.0850")
		if !usdRec.Rate.Equal(expectedUSD) {
			t.Errorf("expected USD rate %s, got %s", expectedUSD, usdRec.Rate)
		}
		if usdRec.BaseCurrency != "EUR" {
			t.Errorf("expected base EUR, got %s", usdRec.BaseCurrency)
		}
	})

	t.Run("fails cleanly on invalid XML", func(t *testing.T) {
		_, err := ParseECBXML([]byte("<not-valid-xml>"))
		if err == nil {
			t.Fatalf("expected error on invalid xml, got nil")
		}
	})

	t.Run("fails when no rates are found", func(t *testing.T) {
		emptyXML := `<gesmes:Envelope xmlns:gesmes="http://www.gesmes.org/xml/2002-05-01"><Cube></Cube></gesmes:Envelope>`
		_, err := ParseECBXML([]byte(emptyXML))
		if err == nil {
			t.Fatalf("expected error on empty rates, got nil")
		}
	})
}

func TestECBProvider_FetchFXRates(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(sampleECBDailyXML))
	}))
	defer ts.Close()

	provider := NewECBProvider(ECBProviderConfig{
		DailyURL:   ts.URL,
		Hist90dURL: ts.URL,
		Limiter:    NewRateLimiter(50, 10),
	})

	targetDate, _ := time.Parse("2006-01-02", "2026-10-02")
	targetDate = targetDate.UTC()

	t.Run("fetches direct EUR/USD rate", func(t *testing.T) {
		pairs := []CurrencyPair{{Base: "EUR", Quote: "USD"}}
		rates, err := provider.FetchFXRates(context.Background(), pairs, targetDate)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(rates) != 1 {
			t.Fatalf("expected 1 rate, got %d", len(rates))
		}
		expected := decimal.RequireFromString("1.0850")
		if !rates[0].Rate.Equal(expected) {
			t.Errorf("expected %s, got %s", expected, rates[0].Rate)
		}
	})

	t.Run("resolves inverse USD/EUR rate exactly", func(t *testing.T) {
		pairs := []CurrencyPair{{Base: "USD", Quote: "EUR"}}
		rates, err := provider.FetchFXRates(context.Background(), pairs, targetDate)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(rates) != 1 {
			t.Fatalf("expected 1 rate, got %d", len(rates))
		}

		// 1 / 1.0850 rounded to 10 decimal places = 0.9216589862
		expected := decimal.NewFromInt(1).DivRound(decimal.RequireFromString("1.0850"), 10)
		if !rates[0].Rate.Equal(expected) {
			t.Errorf("expected inverse %s, got %s", expected, rates[0].Rate)
		}
	})

	t.Run("resolves cross-rate USD/GBP via EUR anchor", func(t *testing.T) {
		pairs := []CurrencyPair{{Base: "USD", Quote: "GBP"}}
		rates, err := provider.FetchFXRates(context.Background(), pairs, targetDate)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(rates) != 1 {
			t.Fatalf("expected 1 rate, got %d", len(rates))
		}

		// (EUR/GBP) / (EUR/USD) = 0.8350 / 1.0850 = 0.7695852535
		expected := decimal.RequireFromString("0.8350").DivRound(decimal.RequireFromString("1.0850"), 10)
		if !rates[0].Rate.Equal(expected) {
			t.Errorf("expected cross rate %s, got %s", expected, rates[0].Rate)
		}
	})

	t.Run("carries forward Friday rate for Sunday (LOCF)", func(t *testing.T) {
		sundayDate, _ := time.Parse("2006-01-02", "2026-10-04") // Sunday
		pairs := []CurrencyPair{{Base: "EUR", Quote: "USD"}}
		rates, err := provider.FetchFXRates(context.Background(), pairs, sundayDate)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(rates) != 1 {
			t.Fatalf("expected 1 rate, got %d", len(rates))
		}
		expected := decimal.RequireFromString("1.0850")
		if !rates[0].Rate.Equal(expected) {
			t.Errorf("expected LOCF rate %s, got %s", expected, rates[0].Rate)
		}
	})
}

func TestECBProvider_FetchHistoricalFXRates(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(sampleECBHistXML))
	}))
	defer ts.Close()

	provider := NewECBProvider(ECBProviderConfig{
		DailyURL:   ts.URL,
		Hist90dURL: ts.URL,
	})

	t.Run("fetches date range in ascending order", func(t *testing.T) {
		from, _ := time.Parse("2006-01-02", "2026-10-01")
		to, _ := time.Parse("2006-01-02", "2026-10-02")

		rates, err := provider.FetchHistoricalFXRates(context.Background(), CurrencyPair{Base: "EUR", Quote: "USD"}, from, to)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(rates) != 2 {
			t.Fatalf("expected 2 historical points, got %d", len(rates))
		}

		if rates[0].RateDate.Format("2006-01-02") != "2026-10-01" {
			t.Errorf("expected first point 2026-10-01, got %s", rates[0].RateDate.Format("2006-01-02"))
		}
		if !rates[0].Rate.Equal(decimal.RequireFromString("1.0820")) {
			t.Errorf("expected 1.0820, got %s", rates[0].Rate)
		}

		if rates[1].RateDate.Format("2006-01-02") != "2026-10-02" {
			t.Errorf("expected second point 2026-10-02, got %s", rates[1].RateDate.Format("2006-01-02"))
		}
		if !rates[1].Rate.Equal(decimal.RequireFromString("1.0850")) {
			t.Errorf("expected 1.0850, got %s", rates[1].Rate)
		}
	})

	t.Run("returns error when to is before from", func(t *testing.T) {
		from, _ := time.Parse("2006-01-02", "2026-10-02")
		to, _ := time.Parse("2006-01-02", "2026-10-01")

		_, err := provider.FetchHistoricalFXRates(context.Background(), CurrencyPair{Base: "EUR", Quote: "USD"}, from, to)
		if err == nil {
			t.Fatalf("expected error when to is before from, got nil")
		}
	})
}

func TestECBProvider_CacheAndRetry(t *testing.T) {
	var requestCount int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&requestCount, 1)
		if count == 1 {
			// First call fails with 503
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(sampleECBDailyXML))
	}))
	defer ts.Close()

	provider := NewECBProvider(ECBProviderConfig{
		DailyURL:   ts.URL,
		Hist90dURL: ts.URL,
		CacheTTL:   10 * time.Minute,
		RetryCfg: RetryConfig{
			MaxRetries:     2,
			InitialBackoff: 5 * time.Millisecond,
		},
	})

	date, _ := time.Parse("2006-01-02", "2026-10-02")

	// Call 1: triggers retry on 503, succeeds on request 2
	rates1, err := provider.FetchFXRates(context.Background(), []CurrencyPair{{Base: "EUR", Quote: "USD"}}, date)
	if err != nil {
		t.Fatalf("unexpected error on call 1: %v", err)
	}
	if len(rates1) != 1 {
		t.Fatalf("expected 1 rate, got %d", len(rates1))
	}
	if atomic.LoadInt32(&requestCount) != 2 {
		t.Errorf("expected 2 server requests (1 failure + 1 retry), got %d", atomic.LoadInt32(&requestCount))
	}

	// Call 2: should be served from memory cache without another network request
	rates2, err := provider.FetchFXRates(context.Background(), []CurrencyPair{{Base: "EUR", Quote: "USD"}}, date)
	if err != nil {
		t.Fatalf("unexpected error on call 2: %v", err)
	}
	if len(rates2) != 1 {
		t.Fatalf("expected 1 rate, got %d", len(rates2))
	}
	if atomic.LoadInt32(&requestCount) != 2 {
		t.Errorf("expected requestCount to remain 2 due to cache, got %d", atomic.LoadInt32(&requestCount))
	}
}

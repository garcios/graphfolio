package marketdata

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestDetectFrequency(t *testing.T) {
	t.Run("Default fallback on insufficient history", func(t *testing.T) {
		freq, step := DetectFrequency(nil)
		if freq != "QUARTERLY" || step != 91 {
			t.Errorf("expected QUARTERLY 91, got %s %d", freq, step)
		}
	})

	t.Run("Quarterly equity cadence", func(t *testing.T) {
		history := []DividendHistoryPoint{
			{Date: time.Date(2023, 2, 10, 0, 0, 0, 0, time.UTC), Amount: decimal.NewFromFloat(0.24)},
			{Date: time.Date(2023, 5, 12, 0, 0, 0, 0, time.UTC), Amount: decimal.NewFromFloat(0.24)},
			{Date: time.Date(2023, 8, 11, 0, 0, 0, 0, time.UTC), Amount: decimal.NewFromFloat(0.24)},
			{Date: time.Date(2023, 11, 10, 0, 0, 0, 0, time.UTC), Amount: decimal.NewFromFloat(0.24)},
		}
		freq, step := DetectFrequency(history)
		if freq != "QUARTERLY" || step != 91 {
			t.Errorf("expected QUARTERLY 91, got %s %d", freq, step)
		}
	})

	t.Run("Monthly REIT cadence", func(t *testing.T) {
		history := []DividendHistoryPoint{
			{Date: time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC), Amount: decimal.NewFromFloat(0.10)},
			{Date: time.Date(2023, 2, 15, 0, 0, 0, 0, time.UTC), Amount: decimal.NewFromFloat(0.10)},
			{Date: time.Date(2023, 3, 15, 0, 0, 0, 0, time.UTC), Amount: decimal.NewFromFloat(0.10)},
			{Date: time.Date(2023, 4, 15, 0, 0, 0, 0, time.UTC), Amount: decimal.NewFromFloat(0.10)},
		}
		freq, step := DetectFrequency(history)
		if freq != "MONTHLY" || step != 30 {
			t.Errorf("expected MONTHLY 30, got %s %d", freq, step)
		}
	})

	t.Run("Semi-annual cadence", func(t *testing.T) {
		history := []DividendHistoryPoint{
			{Date: time.Date(2022, 3, 1, 0, 0, 0, 0, time.UTC), Amount: decimal.NewFromFloat(1.50)},
			{Date: time.Date(2022, 9, 1, 0, 0, 0, 0, time.UTC), Amount: decimal.NewFromFloat(1.50)},
			{Date: time.Date(2023, 3, 1, 0, 0, 0, 0, time.UTC), Amount: decimal.NewFromFloat(1.50)},
		}
		freq, step := DetectFrequency(history)
		if freq != "SEMI_ANNUAL" || step != 182 {
			t.Errorf("expected SEMI_ANNUAL 182, got %s %d", freq, step)
		}
	})
}

func TestSynthesizeForwardEstimates(t *testing.T) {
	instID := uuid.New()
	inst := InstrumentRef{
		ID:       instID,
		Symbol:   "AAPL",
		Exchange: "NASDAQ",
		Currency: "USD",
	}

	asOf := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)

	announced := []DividendScheduleEvent{
		{
			InstrumentID:   instID,
			Symbol:         "AAPL",
			ExchangeCode:   "NASDAQ",
			ExDate:         time.Date(2024, 2, 9, 0, 0, 0, 0, time.UTC),
			PaymentDate:    time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC),
			AmountPerShare: decimal.NewFromFloat(0.24),
			CurrencyCode:   "USD",
			Frequency:      "QUARTERLY",
			Status:         "EX_DATE_PASSED",
			IsEstimated:    false,
		},
	}

	history := []DividendHistoryPoint{
		{Date: time.Date(2023, 5, 12, 0, 0, 0, 0, time.UTC), Amount: decimal.NewFromFloat(0.24)},
		{Date: time.Date(2023, 8, 11, 0, 0, 0, 0, time.UTC), Amount: decimal.NewFromFloat(0.24)},
		{Date: time.Date(2023, 11, 10, 0, 0, 0, 0, time.UTC), Amount: decimal.NewFromFloat(0.24)},
		{Date: time.Date(2024, 2, 9, 0, 0, 0, 0, time.UTC), Amount: decimal.NewFromFloat(0.24)},
	}

	results := SynthesizeForwardEstimates(inst, announced, history, asOf)

	if len(results) < 3 {
		t.Fatalf("expected at least 3 forward quarters synthesized, got %d", len(results))
	}

	// First event should be the announced one
	if results[0].IsEstimated {
		t.Errorf("expected first event to not be estimated")
	}

	// Subsequent events should be estimated
	estimatedCount := 0
	for _, ev := range results[1:] {
		if ev.IsEstimated {
			estimatedCount++
		}
		if ev.AmountPerShare.Cmp(decimal.NewFromFloat(0.24)) != 0 {
			t.Errorf("expected estimated amount 0.24, got %v", ev.AmountPerShare)
		}
	}

	if estimatedCount == 0 {
		t.Errorf("expected synthesized estimated events")
	}
}

func TestSynthesizeForwardEstimates_ASXFranking(t *testing.T) {
	instID := uuid.New()
	inst := InstrumentRef{
		ID:       instID,
		Symbol:   "BHP.AX",
		Exchange: "ASX",
		Currency: "AUD",
	}

	asOf := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	history := []DividendHistoryPoint{
		{Date: time.Date(2023, 3, 1, 0, 0, 0, 0, time.UTC), Amount: decimal.NewFromFloat(1.20)},
		{Date: time.Date(2023, 9, 1, 0, 0, 0, 0, time.UTC), Amount: decimal.NewFromFloat(1.20)},
	}

	results := SynthesizeForwardEstimates(inst, nil, history, asOf)
	if len(results) == 0 {
		t.Fatalf("expected forward estimates for BHP.AX")
	}

	for _, ev := range results {
		if ev.FrankingPct.Cmp(decimal.NewFromInt(100)) != 0 {
			t.Errorf("expected 100%% franking for ASX ticker, got %v", ev.FrankingPct)
		}
	}
}

package domain_test

import (
	"testing"
	"time"

	"portfolio-api/internal/domain"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestCalculateInvestment(t *testing.T) {
	t.Run("same currency conversion (USD holding in USD portfolio)", func(t *testing.T) {
		// AAPL mock scenario:
		// Q = 142.5, Latest = 185.92, Prev = 182.02 (delta = 3.90)
		// Cost basis = 22263.50, Value = 142.5 * 185.92 = 26493.60
		// Total return = 26493.60 - 22263.50 = 4230.10
		h := domain.HoldingWithPrice{
			PortfolioID:        uuid.New(),
			InstrumentID:       uuid.New(),
			Ticker:             "AAPL",
			Name:               "Apple Inc.",
			InstrumentCurrency: "USD",
			Quantity:           decimal.RequireFromString("142.5"),
			CostBasis:          decimal.RequireFromString("22263.50"),
			CostBasisBase:      decimal.RequireFromString("22263.50"),
			RealizedPnLBase:    decimal.Zero,
			DividendsBase:      decimal.Zero,
			LatestPrice:        decimal.RequireFromString("185.92"),
			PrevPrice:          decimal.RequireFromString("182.02"),
			FXRateToBase:       decimal.NewFromInt(1),
		}

		inv := domain.CalculateInvestment(h, "USD")

		if inv.Price.CurrencyCode != "USD" {
			t.Errorf("got Price currency %s, want USD", inv.Price.CurrencyCode)
		}
		if inv.TotalValue.Amount.String() != "26493.6" {
			t.Errorf("got TotalValue %s, want 26493.6", inv.TotalValue.Amount)
		}
		if inv.TotalValue.CurrencyCode != "USD" {
			t.Errorf("got TotalValue currency %s, want USD", inv.TotalValue.CurrencyCode)
		}

		// 142.5 * 3.90 = 555.75
		if inv.TodayReturnAmount.Amount.String() != "555.75" {
			t.Errorf("got TodayReturnAmount %s, want 555.75", inv.TodayReturnAmount.Amount)
		}
		if inv.TodayReturnAmount.CurrencyCode != "USD" {
			t.Errorf("got TodayReturnAmount currency %s, want USD", inv.TodayReturnAmount.CurrencyCode)
		}

		// Total Return = 4230.10
		if inv.TotalReturnAmount.Amount.String() != "4230.1" {
			t.Errorf("got TotalReturnAmount %s, want 4230.1", inv.TotalReturnAmount.Amount)
		}
		if inv.TotalReturnAmount.CurrencyCode != "USD" {
			t.Errorf("got TotalReturnAmount currency %s, want USD", inv.TotalReturnAmount.CurrencyCode)
		}

		// 4230.10 / 22263.50 = 19.0%
		if inv.TotalReturnPercent.String() != "19" {
			t.Errorf("got TotalReturnPercent %s, want 19", inv.TotalReturnPercent)
		}

		expectedAvgBuyPrice := decimal.RequireFromString("156.24")
		if !inv.AverageBuyPrice.Amount.Equal(expectedAvgBuyPrice) {
			t.Errorf("got AverageBuyPrice %s, want %s", inv.AverageBuyPrice.Amount, expectedAvgBuyPrice)
		}
		if inv.AverageBuyPrice.CurrencyCode != "USD" {
			t.Errorf("got AverageBuyPrice currency %s, want USD", inv.AverageBuyPrice.CurrencyCode)
		}
	})

	t.Run("foreign currency conversion (USD holding in AUD portfolio)", func(t *testing.T) {
		// AAPL mock scenario with FXRate to AUD = 1.5230:
		// Q = 142.5, Latest = 185.92 USD, Prev = 182.02 USD (delta = 3.90 USD)
		// TotalValue USD = 26493.60 -> TotalValue AUD = 26493.60 * 1.5230 = 40349.75 AUD
		// TodayReturn USD = 555.75 -> TodayReturn AUD = 555.75 * 1.5230 = 846.41 AUD
		// CostBasisBase AUD = 33907.31 -> TotalReturn AUD = 40349.75 - 33907.31 = 6442.44 AUD
		h := domain.HoldingWithPrice{
			PortfolioID:        uuid.New(),
			InstrumentID:       uuid.New(),
			Ticker:             "AAPL",
			Name:               "Apple Inc.",
			InstrumentCurrency: "USD",
			Quantity:           decimal.RequireFromString("142.5"),
			CostBasis:          decimal.RequireFromString("22263.50"),
			CostBasisBase:      decimal.RequireFromString("33907.31"),
			RealizedPnLBase:    decimal.Zero,
			DividendsBase:      decimal.Zero,
			LatestPrice:        decimal.RequireFromString("185.92"),
			PrevPrice:          decimal.RequireFromString("182.02"),
			FXRateToBase:       decimal.RequireFromString("1.5230"),
		}

		inv := domain.CalculateInvestment(h, "AUD")

		// Traded price stays in instrument currency
		if inv.Price.CurrencyCode != "USD" {
			t.Errorf("got Price currency %s, want USD", inv.Price.CurrencyCode)
		}
		if inv.Price.Amount.String() != "185.92" {
			t.Errorf("got Price amount %s, want 185.92", inv.Price.Amount)
		}

		// TotalValue converted to base currency (AUD)
		expectedTotalValue := decimal.RequireFromString("40349.75")
		if !inv.TotalValue.Amount.Equal(expectedTotalValue) {
			t.Errorf("got TotalValue %s, want %s", inv.TotalValue.Amount, expectedTotalValue)
		}
		if inv.TotalValue.CurrencyCode != "AUD" {
			t.Errorf("got TotalValue currency %s, want AUD", inv.TotalValue.CurrencyCode)
		}

		// TodayReturnAmount converted to base currency (AUD)
		expectedTodayReturn := decimal.RequireFromString("846.41")
		if !inv.TodayReturnAmount.Amount.Equal(expectedTodayReturn) {
			t.Errorf("got TodayReturnAmount %s, want %s", inv.TodayReturnAmount.Amount, expectedTodayReturn)
		}
		if inv.TodayReturnAmount.CurrencyCode != "AUD" {
			t.Errorf("got TodayReturnAmount currency %s, want AUD", inv.TodayReturnAmount.CurrencyCode)
		}

		// TotalReturnAmount in base currency (AUD)
		expectedTotalReturn := decimal.RequireFromString("6442.44")
		if !inv.TotalReturnAmount.Amount.Equal(expectedTotalReturn) {
			t.Errorf("got TotalReturnAmount %s, want %s", inv.TotalReturnAmount.Amount, expectedTotalReturn)
		}
		if inv.TotalReturnAmount.CurrencyCode != "AUD" {
			t.Errorf("got TotalReturnAmount currency %s, want AUD", inv.TotalReturnAmount.CurrencyCode)
		}

		expectedAvgBuyPrice := decimal.RequireFromString("156.24")
		if !inv.AverageBuyPrice.Amount.Equal(expectedAvgBuyPrice) {
			t.Errorf("got AverageBuyPrice %s, want %s", inv.AverageBuyPrice.Amount, expectedAvgBuyPrice)
		}
		if inv.AverageBuyPrice.CurrencyCode != "USD" {
			t.Errorf("got AverageBuyPrice currency %s, want USD", inv.AverageBuyPrice.CurrencyCode)
		}
	})
}

func TestCalculatePortfolioSummary(t *testing.T) {
	portfolio := domain.Portfolio{
		ID:              uuid.New(),
		UserID:          uuid.New(),
		Name:            "Main",
		BaseCurrency:    "USD",
		CostBasisMethod: domain.CostBasisMethodAverageCost,
	}

	holdings := []domain.HoldingWithPrice{
		{
			PortfolioID:        portfolio.ID,
			InstrumentID:       uuid.New(),
			Ticker:             "AAPL",
			Name:               "Apple Inc.",
			InstrumentCurrency: "USD",
			Quantity:           decimal.RequireFromString("100"),
			CostBasis:          decimal.RequireFromString("10000"),
			CostBasisBase:      decimal.RequireFromString("10000"),
			LatestPrice:        decimal.RequireFromString("150"),
			PrevPrice:          decimal.RequireFromString("140"),
			FXRateToBase:       decimal.NewFromInt(1),
		},
	}

	cash := []domain.CashBalance{
		{
			PortfolioID:  portfolio.ID,
			CurrencyCode: "USD",
			Balance:      decimal.RequireFromString("5000"),
		},
	}

	val := &domain.PortfolioValuation{
		PortfolioID:     portfolio.ID,
		ValuationDate:   time.Now(),
		MarketValueBase: decimal.RequireFromString("15000"),
		CashValueBase:   decimal.RequireFromString("5000"),
		TWRIndex:        decimal.RequireFromString("1.1420"),
	}

	summary := domain.CalculatePortfolioSummary(portfolio, holdings, cash, val, nil)

	// TotalValue = 15000 + 5000 = 20000
	if summary.TotalValue.Amount.String() != "20000" {
		t.Errorf("got TotalValue %s, want 20000", summary.TotalValue.Amount)
	}

	// Today's return = 100 * (150 - 140) = 1000
	if summary.TodayReturnAmount.Amount.String() != "1000" {
		t.Errorf("got TodayReturnAmount %s, want 1000", summary.TodayReturnAmount.Amount)
	}

	// Annualized return = (1.1420 - 1) * 100 = 14.2%
	if summary.AnnualizedReturnPercent.String() != "14.2" {
		t.Errorf("got AnnualizedReturnPercent %s, want 14.2", summary.AnnualizedReturnPercent)
	}
}

func TestCalculatePortfolioHistory(t *testing.T) {
	t.Run("calculates history with positive returns", func(t *testing.T) {
		pID := uuid.New()
		date1 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		date2 := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)

		vals := []domain.PortfolioValuation{
			{
				PortfolioID:     pID,
				ValuationDate:   date1,
				MarketValueBase: decimal.RequireFromString("95000.00"),
				CashValueBase:   decimal.RequireFromString("5000.00"),
				DailyReturn:     decimal.Zero,
				TWRIndex:        decimal.RequireFromString("1.0000"),
			},
			{
				PortfolioID:     pID,
				ValuationDate:   date2,
				MarketValueBase: decimal.RequireFromString("105000.00"),
				CashValueBase:   decimal.RequireFromString("5000.00"),
				DailyReturn:     decimal.RequireFromString("0.1000"),
				TWRIndex:        decimal.RequireFromString("1.1000"),
			},
		}

		history := domain.CalculatePortfolioHistory("USD", vals)
		if len(history.Points) != 2 {
			t.Fatalf("expected 2 points, got %d", len(history.Points))
		}

		// Start: 100,000, End: 110,000 -> Return: 10,000 (10.00%)
		expectedStart := decimal.RequireFromString("100000.00")
		expectedEnd := decimal.RequireFromString("110000.00")
		expectedReturn := decimal.RequireFromString("10000.00")
		expectedPercent := decimal.RequireFromString("10")

		if !history.StartValue.Amount.Equal(expectedStart) {
			t.Errorf("expected start value %s, got %s", expectedStart, history.StartValue.Amount)
		}
		if !history.EndValue.Amount.Equal(expectedEnd) {
			t.Errorf("expected end value %s, got %s", expectedEnd, history.EndValue.Amount)
		}
		if !history.ReturnAmount.Amount.Equal(expectedReturn) {
			t.Errorf("expected return amount %s, got %s", expectedReturn, history.ReturnAmount.Amount)
		}
		if !history.ReturnPercent.Equal(expectedPercent) {
			t.Errorf("expected return percent %s, got %s", expectedPercent, history.ReturnPercent)
		}
		if history.Points[0].TotalValue.CurrencyCode != "USD" {
			t.Errorf("expected currency USD, got %s", history.Points[0].TotalValue.CurrencyCode)
		}
	})

	t.Run("calculates history with negative returns", func(t *testing.T) {
		pID := uuid.New()
		date1 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		date2 := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)

		vals := []domain.PortfolioValuation{
			{
				PortfolioID:     pID,
				ValuationDate:   date1,
				MarketValueBase: decimal.RequireFromString("100000.00"),
				CashValueBase:   decimal.Zero,
				DailyReturn:     decimal.Zero,
				TWRIndex:        decimal.RequireFromString("1.0000"),
			},
			{
				PortfolioID:     pID,
				ValuationDate:   date2,
				MarketValueBase: decimal.RequireFromString("80000.00"),
				CashValueBase:   decimal.Zero,
				DailyReturn:     decimal.RequireFromString("-0.2000"),
				TWRIndex:        decimal.RequireFromString("0.8000"),
			},
		}

		history := domain.CalculatePortfolioHistory("USD", vals)
		expectedReturn := decimal.RequireFromString("-20000.00")
		expectedPercent := decimal.RequireFromString("-20")

		if !history.ReturnAmount.Amount.Equal(expectedReturn) {
			t.Errorf("expected return amount %s, got %s", expectedReturn, history.ReturnAmount.Amount)
		}
		if !history.ReturnPercent.Equal(expectedPercent) {
			t.Errorf("expected return percent %s, got %s", expectedPercent, history.ReturnPercent)
		}
	})

	t.Run("handles empty valuations slice", func(t *testing.T) {
		history := domain.CalculatePortfolioHistory("USD", nil)
		if len(history.Points) != 0 {
			t.Errorf("expected 0 points, got %d", len(history.Points))
		}
		if !history.StartValue.Amount.IsZero() {
			t.Errorf("expected zero start value, got %s", history.StartValue.Amount)
		}
		if !history.ReturnPercent.IsZero() {
			t.Errorf("expected zero return percent, got %s", history.ReturnPercent)
		}
	})
}

func TestCalculateInvestment_ReturnAttribution(t *testing.T) {
	t.Run("domestic holding without currency exposure", func(t *testing.T) {
		// BHP in AUD portfolio:
		// Q = 100, Latest = 45.00, CostBasis = 4000.00, CostBasisBase = 4000.00
		// TotalValue = 4500.00 AUD
		// CapGainLocal = 500.00, CapGainBase = 500.00, CapGain% = 12.50%
		// CurrencyGain = 0.00, CurrencyGain% = 0.00%, IsInternational = false
		// DividendsBase = 120.00, YieldOnCost = 120 / 4000 = 3.00%
		h := domain.HoldingWithPrice{
			PortfolioID:        uuid.New(),
			InstrumentID:       uuid.New(),
			Ticker:             "BHP",
			Name:               "BHP Group Limited",
			InstrumentCurrency: "AUD",
			Quantity:           decimal.RequireFromString("100"),
			CostBasis:          decimal.RequireFromString("4000.00"),
			CostBasisBase:      decimal.RequireFromString("4000.00"),
			RealizedPnLBase:    decimal.Zero,
			DividendsBase:      decimal.RequireFromString("120.00"),
			LatestPrice:        decimal.RequireFromString("45.00"),
			PrevPrice:          decimal.RequireFromString("44.50"),
			FXRateToBase:       decimal.NewFromInt(1),
		}

		inv := domain.CalculateInvestment(h, "AUD")

		if inv.IsInternational {
			t.Errorf("expected IsInternational to be false for domestic holding")
		}
		if !inv.CapitalGainAmount.Amount.Equal(decimal.RequireFromString("500.00")) {
			t.Errorf("got CapGain %s, want 500.00", inv.CapitalGainAmount.Amount)
		}
		if !inv.CapitalGainPercent.Equal(decimal.RequireFromString("12.50")) {
			t.Errorf("got CapGain%% %s, want 12.50", inv.CapitalGainPercent)
		}
		if !inv.CurrencyGainAmount.Amount.IsZero() {
			t.Errorf("got CurrencyGain %s, want 0.00", inv.CurrencyGainAmount.Amount)
		}
		if !inv.CurrencyGainPercent.IsZero() {
			t.Errorf("got CurrencyGain%% %s, want 0.00", inv.CurrencyGainPercent)
		}
		if !inv.IncomeAmount.Amount.Equal(decimal.RequireFromString("120.00")) {
			t.Errorf("got Income %s, want 120.00", inv.IncomeAmount.Amount)
		}
		if !inv.IncomeYieldPercent.Equal(decimal.RequireFromString("3.00")) {
			t.Errorf("got IncomeYield%% %s, want 3.00", inv.IncomeYieldPercent)
		}

		// Mathematical attribution reconciliation
		unrealizedBase := inv.TotalValue.Amount.Sub(h.CostBasisBase)
		sumAttribution := inv.CapitalGainAmount.Amount.Add(inv.CurrencyGainAmount.Amount)
		if !unrealizedBase.Equal(sumAttribution) {
			t.Errorf("attribution reconciliation failed: unrealizedBase %s != sumAttribution %s", unrealizedBase, sumAttribution)
		}
	})

	t.Run("international holding with price gain and currency tailwind", func(t *testing.T) {
		// MSFT in AUD portfolio:
		// Q = 10, Latest = 400.00 USD, CostBasis = 3500.00 USD (350.00/share)
		// Initial FX = 1.40 (CostBasisBase = 3500 * 1.40 = 4900.00 AUD)
		// Current FX = 1.50 AUD per USD (USD appreciated vs AUD -> currency tailwind)
		// TotalValueLocal = 4000.00 USD
		// TotalValueBase = 4000 * 1.50 = 6000.00 AUD
		// CapGainLocal = 4000 - 3500 = 500.00 USD
		// CapGainBase = 500 * 1.40 = 700.00 AUD (price gain at acquisition FX)
		// CapGain% = 500 / 3500 = 14.29%
		// CurrencyGainBase = 4000 * (1.50 - 1.40) = 400.00 AUD
		// CurrencyGain% = (1.50 - 1.40) / 1.40 = 7.14%
		// DividendsBase = 50.00 AUD
		h := domain.HoldingWithPrice{
			PortfolioID:        uuid.New(),
			InstrumentID:       uuid.New(),
			Ticker:             "MSFT",
			Name:               "Microsoft Corp",
			InstrumentCurrency: "USD",
			Quantity:           decimal.RequireFromString("10"),
			CostBasis:          decimal.RequireFromString("3500.00"),
			CostBasisBase:      decimal.RequireFromString("4900.00"),
			RealizedPnLBase:    decimal.Zero,
			DividendsBase:      decimal.RequireFromString("50.00"),
			LatestPrice:        decimal.RequireFromString("400.00"),
			PrevPrice:          decimal.RequireFromString("395.00"),
			FXRateToBase:       decimal.RequireFromString("1.50"),
		}

		inv := domain.CalculateInvestment(h, "AUD")

		if !inv.IsInternational {
			t.Errorf("expected IsInternational to be true for foreign holding")
		}
		if !inv.CapitalGainAmount.Amount.Equal(decimal.RequireFromString("700.00")) {
			t.Errorf("got CapGain %s, want 700.00", inv.CapitalGainAmount.Amount)
		}
		if !inv.CapitalGainPercent.Equal(decimal.RequireFromString("14.29")) {
			t.Errorf("got CapGain%% %s, want 14.29", inv.CapitalGainPercent)
		}
		if !inv.CurrencyGainAmount.Amount.Equal(decimal.RequireFromString("400.00")) {
			t.Errorf("got CurrencyGain %s, want 400.00", inv.CurrencyGainAmount.Amount)
		}
		if !inv.CurrencyGainPercent.Equal(decimal.RequireFromString("7.14")) {
			t.Errorf("got CurrencyGain%% %s, want 7.14", inv.CurrencyGainPercent)
		}
		if !inv.IncomeAmount.Amount.Equal(decimal.RequireFromString("50.00")) {
			t.Errorf("got Income %s, want 50.00", inv.IncomeAmount.Amount)
		}

		// Exact identity verification:
		// CapGainBase (700) + CurrencyGainBase (400) = 1100 AUD = TotalValueBase (6000) - CostBasisBase (4900)
		unrealizedBase := inv.TotalValue.Amount.Sub(h.CostBasisBase)
		sumAttribution := inv.CapitalGainAmount.Amount.Add(inv.CurrencyGainAmount.Amount)
		if !unrealizedBase.Equal(sumAttribution) {
			t.Errorf("attribution identity failed: unrealizedBase %s != sumAttribution %s", unrealizedBase, sumAttribution)
		}
	})

	t.Run("international holding with price gain but currency headwind (drag)", func(t *testing.T) {
		// NVDA in AUD portfolio:
		// Q = 10, Latest = 120.00 USD, CostBasis = 1000.00 USD
		// Initial FX = 1.60 (CostBasisBase = 1600.00 AUD)
		// Current FX = 1.45 (USD depreciated vs AUD -> currency drag)
		// TotalValueLocal = 1200.00 USD
		// TotalValueBase = 1200 * 1.45 = 1740.00 AUD
		// CapGainLocal = 1200 - 1000 = 200.00 USD
		// CapGainBase = 200 * 1.60 = 320.00 AUD
		// CurrencyGainBase = 1200 * (1.45 - 1.60) = -180.00 AUD
		// Total unrealized = 1740 - 1600 = +140.00 AUD = 320 - 180!
		h := domain.HoldingWithPrice{
			PortfolioID:        uuid.New(),
			InstrumentID:       uuid.New(),
			Ticker:             "NVDA",
			Name:               "Nvidia Corp",
			InstrumentCurrency: "USD",
			Quantity:           decimal.RequireFromString("10"),
			CostBasis:          decimal.RequireFromString("1000.00"),
			CostBasisBase:      decimal.RequireFromString("1600.00"),
			RealizedPnLBase:    decimal.Zero,
			DividendsBase:      decimal.Zero,
			LatestPrice:        decimal.RequireFromString("120.00"),
			PrevPrice:          decimal.RequireFromString("118.00"),
			FXRateToBase:       decimal.RequireFromString("1.45"),
		}

		inv := domain.CalculateInvestment(h, "AUD")

		if !inv.CapitalGainAmount.Amount.Equal(decimal.RequireFromString("320.00")) {
			t.Errorf("got CapGain %s, want 320.00", inv.CapitalGainAmount.Amount)
		}
		if !inv.CurrencyGainAmount.Amount.Equal(decimal.RequireFromString("-180.00")) {
			t.Errorf("got CurrencyGain %s, want -180.00", inv.CurrencyGainAmount.Amount)
		}

		unrealizedBase := inv.TotalValue.Amount.Sub(h.CostBasisBase)
		sumAttribution := inv.CapitalGainAmount.Amount.Add(inv.CurrencyGainAmount.Amount)
		if !unrealizedBase.Equal(sumAttribution) {
			t.Errorf("attribution identity failed: unrealizedBase %s != sumAttribution %s", unrealizedBase, sumAttribution)
		}
	})

	t.Run("zero cost basis handles cleanly without division by zero", func(t *testing.T) {
		h := domain.HoldingWithPrice{
			PortfolioID:        uuid.New(),
			InstrumentID:       uuid.New(),
			Ticker:             "FREE",
			Name:               "Gift Shares",
			InstrumentCurrency: "USD",
			Quantity:           decimal.RequireFromString("10"),
			CostBasis:          decimal.Zero,
			CostBasisBase:      decimal.Zero,
			RealizedPnLBase:    decimal.Zero,
			DividendsBase:      decimal.Zero,
			LatestPrice:        decimal.RequireFromString("10.00"),
			PrevPrice:          decimal.RequireFromString("10.00"),
			FXRateToBase:       decimal.RequireFromString("1.50"),
		}

		inv := domain.CalculateInvestment(h, "AUD")

		if !inv.CapitalGainPercent.IsZero() {
			t.Errorf("expected zero CapGain%%, got %s", inv.CapitalGainPercent)
		}
		if !inv.CurrencyGainPercent.IsZero() {
			t.Errorf("expected zero CurrencyGain%%, got %s", inv.CurrencyGainPercent)
		}
		if !inv.IncomeYieldPercent.IsZero() {
			t.Errorf("expected zero IncomeYield%%, got %s", inv.IncomeYieldPercent)
		}
		if !inv.AverageBuyPrice.Amount.IsZero() {
			t.Errorf("expected zero AverageBuyPrice for zero cost basis, got %s", inv.AverageBuyPrice.Amount)
		}
	})

	t.Run("post-split holding calculates correct split-adjusted average buy price", func(t *testing.T) {
		// e.g. Pre-split 100 shares @ $50 ($5,000 cost basis), 2-for-1 split -> 200 shares @ $25 avg buy price
		h := domain.HoldingWithPrice{
			PortfolioID:        uuid.New(),
			InstrumentID:       uuid.New(),
			Ticker:             "SPLT",
			Name:               "Split Corp",
			InstrumentCurrency: "USD",
			Quantity:           decimal.RequireFromString("200"),
			CostBasis:          decimal.RequireFromString("5000.00"),
			CostBasisBase:      decimal.RequireFromString("5000.00"),
			LatestPrice:        decimal.RequireFromString("30.00"),
			PrevPrice:          decimal.RequireFromString("30.00"),
			FXRateToBase:       decimal.NewFromInt(1),
		}

		inv := domain.CalculateInvestment(h, "USD")
		expectedAvgBuy := decimal.RequireFromString("25")
		if !inv.AverageBuyPrice.Amount.Equal(expectedAvgBuy) {
			t.Errorf("got AverageBuyPrice %s, want %s", inv.AverageBuyPrice.Amount, expectedAvgBuy)
		}
		if inv.AverageBuyPrice.CurrencyCode != "USD" {
			t.Errorf("got AverageBuyPrice currency %s, want USD", inv.AverageBuyPrice.CurrencyCode)
		}
	})

	t.Run("zero quantity holding returns zero AverageBuyPrice without panic", func(t *testing.T) {
		h := domain.HoldingWithPrice{
			PortfolioID:        uuid.New(),
			InstrumentID:       uuid.New(),
			Ticker:             "ZERO",
			Name:               "Zero Shares",
			InstrumentCurrency: "USD",
			Quantity:           decimal.Zero,
			CostBasis:          decimal.Zero,
			CostBasisBase:      decimal.Zero,
			LatestPrice:        decimal.RequireFromString("10.00"),
			PrevPrice:          decimal.RequireFromString("10.00"),
			FXRateToBase:       decimal.NewFromInt(1),
		}

		inv := domain.CalculateInvestment(h, "USD")
		if !inv.AverageBuyPrice.Amount.IsZero() {
			t.Errorf("expected zero AverageBuyPrice for zero quantity, got %s", inv.AverageBuyPrice.Amount)
		}
	})
}

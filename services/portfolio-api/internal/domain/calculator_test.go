package domain_test

import (
	"testing"
	"time"

	"portfolio-api/internal/domain"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestCalculateInvestment(t *testing.T) {
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

	if inv.TotalValue.Amount.String() != "26493.6" {
		t.Errorf("got TotalValue %s, want 26493.6", inv.TotalValue.Amount)
	}

	// 142.5 * 3.90 = 555.75
	if inv.TodayReturnAmount.Amount.String() != "555.75" {
		t.Errorf("got TodayReturnAmount %s, want 555.75", inv.TodayReturnAmount.Amount)
	}

	// Total Return = 4230.10
	if inv.TotalReturnAmount.Amount.String() != "4230.1" {
		t.Errorf("got TotalReturnAmount %s, want 4230.1", inv.TotalReturnAmount.Amount)
	}

	// 4230.10 / 22263.50 = 19.0%
	if inv.TotalReturnPercent.String() != "19" {
		t.Errorf("got TotalReturnPercent %s, want 19", inv.TotalReturnPercent)
	}
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

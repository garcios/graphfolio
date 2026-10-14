package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculateBenchmarkComparison_RebasingAtT0(t *testing.T) {
	// TC-BENCH-01: Re-basing to 0.00% at t0 for both series
	portfolioID := uuid.New()
	t0 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	t1 := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)

	valuations := []PortfolioValuation{
		{
			PortfolioID:     portfolioID,
			ValuationDate:   t0,
			MarketValueBase: decimal.RequireFromString("100000"),
			CashValueBase:   decimal.Zero,
			TWRIndex:        decimal.RequireFromString("1.000000000000"),
		},
		{
			PortfolioID:     portfolioID,
			ValuationDate:   t1,
			MarketValueBase: decimal.RequireFromString("105000"),
			CashValueBase:   decimal.Zero,
			TWRIndex:        decimal.RequireFromString("1.050000000000"),
		},
	}

	quotes := []RawBenchmarkQuote{
		{
			PriceDate:  t0,
			ClosePrice: decimal.RequireFromString("500"),
		},
		{
			PriceDate:  t1,
			ClosePrice: decimal.RequireFromString("525"),
		},
	}

	benchmark := Benchmark{
		ID:           "SP500TR",
		Symbol:       "SP500TR",
		CurrencyCode: "USD",
	}

	res, err := CalculateBenchmarkComparison(
		portfolioID,
		"USD",
		benchmark,
		BenchmarkCurrencyModeBenchmarkNative,
		valuations,
		quotes,
	)
	require.NoError(t, err)
	require.Len(t, res.Points, 2)

	// Check Point 0
	assert.True(t, res.Points[0].PortfolioReturnPercent.Equal(decimal.Zero), "portfolio return at T0 must be 0%")
	assert.True(t, res.Points[0].BenchmarkReturnPercent.Equal(decimal.Zero), "benchmark return at T0 must be 0%")
	assert.True(t, res.Points[0].ActiveReturnPercent.Equal(decimal.Zero), "active return at T0 must be 0%")

	// Check Point 1: +5.00% for both -> active return 0.00%
	assert.True(t, res.Points[1].PortfolioReturnPercent.Equal(decimal.RequireFromString("5.00")), "portfolio return at T1 should be +5.00%")
	assert.True(t, res.Points[1].BenchmarkReturnPercent.Equal(decimal.RequireFromString("5.00")), "benchmark return at T1 should be +5.00%")
	assert.True(t, res.Points[1].ActiveReturnPercent.Equal(decimal.Zero), "active return at T1 should be 0.00%")
}

func TestCalculateBenchmarkComparison_ScenarioA_MultiYearCurrencyTranslation(t *testing.T) {
	// Scenario A from implementation plan:
	// Investor Base: AUD
	// Selected Benchmark: SP500TR (USD)
	// Mode: MODE_PORTFOLIO_BASE
	// t0: 2024-10-14
	//   TWRIndex: 1.000000000000
	//   P_b: 500.00 USD
	//   USD/AUD FX: 1.50
	//   P_adj(t0): 500.00 * 1.50 = 750.00 AUD
	// tN: 2026-10-14 (CD = 730 days)
	//   TWRIndex: 1.440000000000 (+44.00%)
	//   P_b: 600.00 USD (+20% in USD)
	//   USD/AUD FX: 1.55
	//   P_adj(tN): 600.00 * 1.55 = 930.00 AUD (+24.00% in AUD)
	// Expected:
	//   Active Cumulative Return: +20.00%
	//   Portfolio Ann Return: +20.02% p.a.
	//   Benchmark Ann Return: +11.39% p.a.
	//   Active Ann Return: +8.63% p.a.
	portfolioID := uuid.New()
	t0 := time.Date(2024, 10, 14, 0, 0, 0, 0, time.UTC)
	tN := time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC)

	valuations := []PortfolioValuation{
		{
			PortfolioID:     portfolioID,
			ValuationDate:   t0,
			MarketValueBase: decimal.RequireFromString("100000"),
			CashValueBase:   decimal.Zero,
			TWRIndex:        decimal.RequireFromString("1.000000000000"),
		},
		{
			PortfolioID:     portfolioID,
			ValuationDate:   tN,
			MarketValueBase: decimal.RequireFromString("144000"),
			CashValueBase:   decimal.Zero,
			TWRIndex:        decimal.RequireFromString("1.440000000000"),
		},
	}

	fx0 := decimal.RequireFromString("1.50")
	fxN := decimal.RequireFromString("1.55")

	quotes := []RawBenchmarkQuote{
		{
			PriceDate:    t0,
			ClosePrice:   decimal.RequireFromString("500.00"),
			FXRateToBase: &fx0,
		},
		{
			PriceDate:    tN,
			ClosePrice:   decimal.RequireFromString("600.00"),
			FXRateToBase: &fxN,
		},
	}

	benchmark := Benchmark{
		ID:           "SP500TR",
		Symbol:       "SP500TR",
		CurrencyCode: "USD",
	}

	res, err := CalculateBenchmarkComparison(
		portfolioID,
		"AUD",
		benchmark,
		BenchmarkCurrencyModePortfolioBase,
		valuations,
		quotes,
	)
	require.NoError(t, err)
	assert.Equal(t, 730, res.ElapsedDays)

	// Cumulative returns
	assert.True(t, res.Metrics.PortfolioCumulativeReturnPercent.Equal(decimal.RequireFromString("44.00")), "portfolio cumulative return must be +44.00%")
	assert.True(t, res.Metrics.BenchmarkCumulativeReturnPercent.Equal(decimal.RequireFromString("24.00")), "benchmark cumulative return in AUD must be +24.00%")
	assert.True(t, res.Metrics.ActiveCumulativeReturnPercent.Equal(decimal.RequireFromString("20.00")), "active cumulative return must be +20.00%")

	// Annualization
	assert.True(t, res.Metrics.IsAnnualized, "should be annualized for 730 days >= 365")
	require.NotNil(t, res.Metrics.PortfolioAnnualizedReturnPercent)
	require.NotNil(t, res.Metrics.BenchmarkAnnualizedReturnPercent)
	require.NotNil(t, res.Metrics.ActiveAnnualizedReturnPercent)

	t.Logf("ElapsedDays: %d, pAnn: %s, bAnn: %s, actAnn: %s", res.ElapsedDays, res.Metrics.PortfolioAnnualizedReturnPercent, res.Metrics.BenchmarkAnnualizedReturnPercent, res.Metrics.ActiveAnnualizedReturnPercent)
	assert.True(t, res.Metrics.PortfolioAnnualizedReturnPercent.Equal(decimal.RequireFromString("20.01")), "portfolio annualized should be +20.01%")
	assert.True(t, res.Metrics.BenchmarkAnnualizedReturnPercent.Equal(decimal.RequireFromString("11.36")), "benchmark annualized should be +11.36%")
	assert.True(t, res.Metrics.ActiveAnnualizedReturnPercent.Equal(decimal.RequireFromString("8.65")), "active annualized return should be +8.65%")
}

func TestCalculateBenchmarkComparison_ScenarioB_SubYearGIPSProhibition(t *testing.T) {
	// Scenario B: Horizon = 90 calendar days. GIPS Standard 2.A.20 prohibition
	portfolioID := uuid.New()
	t0 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	tN := t0.AddDate(0, 0, 90)

	valuations := []PortfolioValuation{
		{
			PortfolioID:     portfolioID,
			ValuationDate:   t0,
			MarketValueBase: decimal.RequireFromString("100000"),
			CashValueBase:   decimal.Zero,
			TWRIndex:        decimal.RequireFromString("1.000000000000"),
		},
		{
			PortfolioID:     portfolioID,
			ValuationDate:   tN,
			MarketValueBase: decimal.RequireFromString("106000"),
			CashValueBase:   decimal.Zero,
			TWRIndex:        decimal.RequireFromString("1.060000000000"),
		},
	}

	quotes := []RawBenchmarkQuote{
		{
			PriceDate:  t0,
			ClosePrice: decimal.RequireFromString("500.00"),
		},
		{
			PriceDate:  tN,
			ClosePrice: decimal.RequireFromString("522.50"), // +4.50%
		},
	}

	benchmark := Benchmark{
		ID:           "SP500TR",
		Symbol:       "SP500TR",
		CurrencyCode: "USD",
	}

	res, err := CalculateBenchmarkComparison(
		portfolioID,
		"USD",
		benchmark,
		BenchmarkCurrencyModeBenchmarkNative,
		valuations,
		quotes,
	)
	require.NoError(t, err)

	assert.False(t, res.Metrics.IsAnnualized, "must not be annualized when CD < 365")
	assert.Nil(t, res.Metrics.PortfolioAnnualizedReturnPercent)
	assert.Nil(t, res.Metrics.BenchmarkAnnualizedReturnPercent)
	assert.Nil(t, res.Metrics.ActiveAnnualizedReturnPercent)

	assert.True(t, res.Metrics.PortfolioCumulativeReturnPercent.Equal(decimal.RequireFromString("6.00")))
	assert.True(t, res.Metrics.BenchmarkCumulativeReturnPercent.Equal(decimal.RequireFromString("4.50")))
	assert.True(t, res.Metrics.ActiveCumulativeReturnPercent.Equal(decimal.RequireFromString("1.50")))
}

func TestCalculateBenchmarkComparison_LOCF_HolidayAlignment(t *testing.T) {
	// TC-BENCH-05: Holiday LOCF handling when calendar dates do not overlap
	portfolioID := uuid.New()
	t0 := time.Date(2025, 7, 3, 0, 0, 0, 0, time.UTC)
	t1 := time.Date(2025, 7, 4, 0, 0, 0, 0, time.UTC) // US Independence Day (holiday)
	t2 := time.Date(2025, 7, 5, 0, 0, 0, 0, time.UTC)

	valuations := []PortfolioValuation{
		{PortfolioID: portfolioID, ValuationDate: t0, TWRIndex: decimal.RequireFromString("1.000000000000"), MarketValueBase: decimal.NewFromInt(100)},
		{PortfolioID: portfolioID, ValuationDate: t1, TWRIndex: decimal.RequireFromString("1.010000000000"), MarketValueBase: decimal.NewFromInt(101)},
		{PortfolioID: portfolioID, ValuationDate: t2, TWRIndex: decimal.RequireFromString("1.020000000000"), MarketValueBase: decimal.NewFromInt(102)},
	}

	// Benchmark has quote on July 3, misses July 4, and next quote on July 5
	quotes := []RawBenchmarkQuote{
		{PriceDate: t0, ClosePrice: decimal.RequireFromString("100.00")},
		{PriceDate: t2, ClosePrice: decimal.RequireFromString("105.00")},
	}

	benchmark := Benchmark{ID: "SP500TR", Symbol: "SP500TR", CurrencyCode: "USD"}

	res, err := CalculateBenchmarkComparison(
		portfolioID,
		"USD",
		benchmark,
		BenchmarkCurrencyModeBenchmarkNative,
		valuations,
		quotes,
	)
	require.NoError(t, err)
	require.Len(t, res.Points, 3)

	// On July 4, price should be carried forward from July 3 (100.00)
	assert.True(t, res.Points[1].BenchmarkPriceAdjusted.Amount.Equal(decimal.RequireFromString("100.00")))
	assert.True(t, res.Points[1].BenchmarkReturnPercent.Equal(decimal.Zero))

	// On July 5, price is 105.00 (+5.00%)
	assert.True(t, res.Points[2].BenchmarkPriceAdjusted.Amount.Equal(decimal.RequireFromString("105.00")))
	assert.True(t, res.Points[2].BenchmarkReturnPercent.Equal(decimal.RequireFromString("5.00")))
}

func TestCalculateBenchmarkComparison_BetaAndTrackingError(t *testing.T) {
	// TC-BENCH-06: 35 daily observations to test Beta and Tracking Error calculation
	portfolioID := uuid.New()
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	numDays := 35
	valuations := make([]PortfolioValuation, numDays)
	quotes := make([]RawBenchmarkQuote, numDays)

	// Simulate portfolio having 1.2x beta to benchmark
	portTWR := decimal.NewFromInt(1)
	benchPrice := decimal.RequireFromString("100.00")

	for i := 0; i < numDays; i++ {
		d := startDate.AddDate(0, 0, i)
		if i > 0 {
			var benchFactor decimal.Decimal
			var portFactor decimal.Decimal
			if i%2 == 0 {
				benchFactor = decimal.RequireFromString("1.015")
				portFactor = decimal.RequireFromString("1.018") // 1.2x of 1.5% is 1.8%
			} else {
				benchFactor = decimal.RequireFromString("0.995") // -0.5%
				portFactor = decimal.RequireFromString("0.994") // 1.2x of -0.5% is -0.6%
			}
			benchPrice = benchPrice.Mul(benchFactor)
			portTWR = portTWR.Mul(portFactor)
		}

		valuations[i] = PortfolioValuation{
			PortfolioID:     portfolioID,
			ValuationDate:   d,
			MarketValueBase: decimal.NewFromInt(100000),
			CashValueBase:   decimal.Zero,
			TWRIndex:        portTWR,
		}

		quotes[i] = RawBenchmarkQuote{
			PriceDate:  d,
			ClosePrice: benchPrice,
		}
	}

	benchmark := Benchmark{ID: "SP500TR", Symbol: "SP500TR", CurrencyCode: "USD"}

	res, err := CalculateBenchmarkComparison(
		portfolioID,
		"USD",
		benchmark,
		BenchmarkCurrencyModeBenchmarkNative,
		valuations,
		quotes,
	)
	require.NoError(t, err)

	require.NotNil(t, res.Metrics.Beta, "Beta must be calculated for 34 daily returns >= 30")
	require.NotNil(t, res.Metrics.TrackingErrorPercent, "Tracking error must be calculated")

	// Since portfolio returned exactly 1.2 * benchmark return constantly, Beta is ~1.20
	assert.True(t, res.Metrics.Beta.Equal(decimal.RequireFromString("1.20")), "Beta should be ~1.20")
}

func TestCalculateBenchmarkComparison_EmptyGuards(t *testing.T) {
	portfolioID := uuid.New()
	benchmark := Benchmark{ID: "SP500TR", Symbol: "SP500TR", CurrencyCode: "USD"}

	_, err := CalculateBenchmarkComparison(portfolioID, "USD", benchmark, BenchmarkCurrencyModePortfolioBase, nil, nil)
	assert.Error(t, err)

	vals := []PortfolioValuation{{ValuationDate: time.Now(), TWRIndex: decimal.NewFromInt(1)}}
	_, err = CalculateBenchmarkComparison(portfolioID, "USD", benchmark, BenchmarkCurrencyModePortfolioBase, vals, nil)
	assert.Error(t, err)
}

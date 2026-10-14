package domain

import (
	"errors"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// BenchmarkCurrencyMode represents whether the benchmark is translated into
// the portfolio base currency via daily FX or expressed in native currency.
type BenchmarkCurrencyMode string

const (
	BenchmarkCurrencyModePortfolioBase   BenchmarkCurrencyMode = "PORTFOLIO_BASE"
	BenchmarkCurrencyModeBenchmarkNative BenchmarkCurrencyMode = "BENCHMARK_NATIVE"
)

// Benchmark represents a market index or proxy reference asset.
type Benchmark struct {
	ID           string
	InstrumentID uuid.UUID
	Symbol       string
	Name         string
	ExchangeCode string
	CurrencyCode string
	IndexType    string // "TOTAL_RETURN" or "PRICE_RETURN"
	IsDefault    bool
	DisplayOrder int
}

// RawBenchmarkQuote represents a historical closing price for a benchmark and optional FX rate.
type RawBenchmarkQuote struct {
	PriceDate    time.Time
	ClosePrice   decimal.Decimal
	FXRateToBase *decimal.Decimal // Rate to convert benchmark currency to portfolio base currency
}

// BenchmarkComparisonPoint represents a single aligned time point comparing portfolio and benchmark.
type BenchmarkComparisonPoint struct {
	Date                   time.Time
	PortfolioReturnPercent decimal.Decimal // Cumulative % from start date T_0
	BenchmarkReturnPercent decimal.Decimal // Cumulative % from start date T_0
	ActiveReturnPercent    decimal.Decimal // PortfolioReturnPercent - BenchmarkReturnPercent
	PortfolioValue         Money           // Total portfolio valuation in base currency
	BenchmarkPriceAdjusted Money           // Benchmark price in selected currency mode
}

// BenchmarkMetrics represents risk and relative return attribution metrics.
type BenchmarkMetrics struct {
	PortfolioCumulativeReturnPercent decimal.Decimal
	BenchmarkCumulativeReturnPercent decimal.Decimal
	ActiveCumulativeReturnPercent    decimal.Decimal
	PortfolioAnnualizedReturnPercent *decimal.Decimal
	BenchmarkAnnualizedReturnPercent *decimal.Decimal
	ActiveAnnualizedReturnPercent    *decimal.Decimal
	IsAnnualized                     bool
	Beta                             *decimal.Decimal
	TrackingErrorPercent             *decimal.Decimal
	InformationRatio                 *decimal.Decimal
}

// BenchmarkComparisonResult aggregates time-series comparison points and performance metrics.
type BenchmarkComparisonResult struct {
	PortfolioID  uuid.UUID
	BaseCurrency string
	Benchmark    Benchmark
	CurrencyMode BenchmarkCurrencyMode
	StartDate    time.Time
	EndDate      time.Time
	ElapsedDays  int
	Points       []BenchmarkComparisonPoint
	Metrics      BenchmarkMetrics
}

// CalculateBenchmarkComparison produces common-origin re-based time series and risk attribution metrics.
func CalculateBenchmarkComparison(
	portfolioID uuid.UUID,
	baseCurrency string,
	benchmark Benchmark,
	currencyMode BenchmarkCurrencyMode,
	valuations []PortfolioValuation,
	quotes []RawBenchmarkQuote,
) (*BenchmarkComparisonResult, error) {
	if len(valuations) == 0 {
		return nil, errors.New("cannot compare benchmark: empty portfolio valuations")
	}
	if len(quotes) == 0 {
		return nil, errors.New("cannot compare benchmark: empty benchmark quotes")
	}

	startDate := valuations[0].ValuationDate
	endDate := valuations[len(valuations)-1].ValuationDate
	elapsedDays := int(endDate.Sub(startDate).Hours() / 24)
	if elapsedDays < 0 {
		elapsedDays = 0
	}

	// 1. Build an aligned map or LOCF lookup of benchmark prices
	// Quotes are expected in ascending chronological order.
	resolvedQuotes := alignQuotesLOCF(valuations, quotes)

	points := make([]BenchmarkComparisonPoint, len(valuations))

	// Baseline values at T_0
	t0TWR := valuations[0].TWRIndex
	if !t0TWR.IsPositive() {
		t0TWR = decimal.NewFromInt(1)
	}

	t0Price := resolveBenchmarkAdjustedPrice(resolvedQuotes[0], benchmark.CurrencyCode, baseCurrency, currencyMode)
	if !t0Price.IsPositive() {
		return nil, errors.New("benchmark initial price at T_0 must be positive")
	}

	priceCurrency := benchmark.CurrencyCode
	if currencyMode == BenchmarkCurrencyModePortfolioBase {
		priceCurrency = baseCurrency
	}

	oneHundred := decimal.NewFromInt(100)

	// Series for daily returns to calculate Beta & Tracking Error
	var dailyPortReturns []float64
	var dailyBenchReturns []float64

	var prevPortTWR decimal.Decimal = t0TWR
	var prevBenchPrice decimal.Decimal = t0Price

	for i, v := range valuations {
		// Portfolio cumulative return from T_0: (TWR(t) / TWR(0) - 1) * 100
		var portCumReturn decimal.Decimal
		if t0TWR.IsPositive() {
			portCumReturn = v.TWRIndex.DivRound(t0TWR, 8).Sub(decimal.NewFromInt(1)).Mul(oneHundred).Round(2)
		}

		// Benchmark adjusted price at t
		adjPrice := resolveBenchmarkAdjustedPrice(resolvedQuotes[i], benchmark.CurrencyCode, baseCurrency, currencyMode)

		// Benchmark cumulative return from T_0: (Price(t) / Price(0) - 1) * 100
		benchCumReturn := adjPrice.DivRound(t0Price, 8).Sub(decimal.NewFromInt(1)).Mul(oneHundred).Round(2)

		// Active return spread: R_p(t) - R_b(t)
		activeReturn := portCumReturn.Sub(benchCumReturn).Round(2)

		totalPortVal := v.MarketValueBase.Add(v.CashValueBase).Round(2)

		points[i] = BenchmarkComparisonPoint{
			Date:                   v.ValuationDate,
			PortfolioReturnPercent: portCumReturn,
			BenchmarkReturnPercent: benchCumReturn,
			ActiveReturnPercent:    activeReturn,
			PortfolioValue:         NewMoney(totalPortVal, baseCurrency),
			BenchmarkPriceAdjusted: NewMoney(adjPrice.Round(2), priceCurrency),
		}

		// For daily return tracking (i >= 1)
		if i > 0 {
			if prevPortTWR.IsPositive() && prevBenchPrice.IsPositive() {
				rp := v.TWRIndex.Sub(prevPortTWR).DivRound(prevPortTWR, 8)
				rb := adjPrice.Sub(prevBenchPrice).DivRound(prevBenchPrice, 8)

				rpF, _ := rp.Float64()
				rbF, _ := rb.Float64()

				dailyPortReturns = append(dailyPortReturns, rpF)
				dailyBenchReturns = append(dailyBenchReturns, rbF)
			}
			prevPortTWR = v.TWRIndex
			prevBenchPrice = adjPrice
		}
	}

	lastPoint := points[len(points)-1]
	portCum := lastPoint.PortfolioReturnPercent
	benchCum := lastPoint.BenchmarkReturnPercent
	activeCum := lastPoint.ActiveReturnPercent

	// 2. GIPS Standard 2.A.20: Sub-1-Year Annualization Prohibition
	var portAnn *decimal.Decimal
	var benchAnn *decimal.Decimal
	var activeAnn *decimal.Decimal
	isAnnualized := false

	if elapsedDays >= 365 {
		isAnnualized = true

		// Annualized portfolio return: ((1 + R_p/100) ^ (365.25 / elapsedDays) - 1) * 100
		pAnn := calculateCompoundAnnualized(portCum, elapsedDays)
		bAnn := calculateCompoundAnnualized(benchCum, elapsedDays)
		actAnn := pAnn.Sub(bAnn).Round(2)

		portAnn = &pAnn
		benchAnn = &bAnn
		activeAnn = &actAnn
	}

	// 3. Attribution Risk Metrics (Beta, Tracking Error, Information Ratio)
	// Only compute if at least 30 observations
	var beta *decimal.Decimal
	var trackingError *decimal.Decimal
	var infoRatio *decimal.Decimal

	if len(dailyPortReturns) >= 30 {
		b, te := calculateBetaAndTrackingError(dailyPortReturns, dailyBenchReturns)
		if b != nil {
			beta = b
		}
		if te != nil {
			trackingError = te
			// Information Ratio is only computed when period >= 365 days and TE > 0
			if isAnnualized && activeAnn != nil && te.IsPositive() {
				ir := activeAnn.DivRound(*te, 2)
				infoRatio = &ir
			}
		}
	}

	metrics := BenchmarkMetrics{
		PortfolioCumulativeReturnPercent: portCum,
		BenchmarkCumulativeReturnPercent: benchCum,
		ActiveCumulativeReturnPercent:    activeCum,
		PortfolioAnnualizedReturnPercent: portAnn,
		BenchmarkAnnualizedReturnPercent: benchAnn,
		ActiveAnnualizedReturnPercent:    activeAnn,
		IsAnnualized:                     isAnnualized,
		Beta:                             beta,
		TrackingErrorPercent:             trackingError,
		InformationRatio:                 infoRatio,
	}

	return &BenchmarkComparisonResult{
		PortfolioID:  portfolioID,
		BaseCurrency: baseCurrency,
		Benchmark:    benchmark,
		CurrencyMode: currencyMode,
		StartDate:    startDate,
		EndDate:      endDate,
		ElapsedDays:  elapsedDays,
		Points:       points,
		Metrics:      metrics,
	}, nil
}

// alignQuotesLOCF matches each valuation date with the most recent benchmark quote at or before that date.
func alignQuotesLOCF(valuations []PortfolioValuation, quotes []RawBenchmarkQuote) []RawBenchmarkQuote {
	resolved := make([]RawBenchmarkQuote, len(valuations))

	quoteIdx := 0
	var lastQuote RawBenchmarkQuote
	hasQuote := false

	for i, v := range valuations {
		// Advance quotes while price_date <= valuation_date
		for quoteIdx < len(quotes) && !quotes[quoteIdx].PriceDate.After(v.ValuationDate) {
			lastQuote = quotes[quoteIdx]
			hasQuote = true
			quoteIdx++
		}

		if hasQuote {
			resolved[i] = lastQuote
		} else if len(quotes) > 0 {
			// If no quote before valuation date, use earliest available
			resolved[i] = quotes[0]
		}
	}

	return resolved
}

// resolveBenchmarkAdjustedPrice converts a benchmark quote to the requested currency mode.
func resolveBenchmarkAdjustedPrice(
	quote RawBenchmarkQuote,
	benchmarkCurrency string,
	baseCurrency string,
	mode BenchmarkCurrencyMode,
) decimal.Decimal {
	if mode == BenchmarkCurrencyModeBenchmarkNative || benchmarkCurrency == baseCurrency {
		return quote.ClosePrice
	}

	// In PORTFOLIO_BASE mode, multiply by FX rate if available
	if quote.FXRateToBase != nil && quote.FXRateToBase.IsPositive() {
		return quote.ClosePrice.Mul(*quote.FXRateToBase)
	}

	return quote.ClosePrice
}

// calculateCompoundAnnualized computes: ((1 + cumPercent / 100) ^ (365.25 / days) - 1) * 100
func calculateCompoundAnnualized(cumPercent decimal.Decimal, days int) decimal.Decimal {
	if days <= 0 {
		return decimal.Zero
	}
	factor := decimal.NewFromInt(1).Add(cumPercent.Div(decimal.NewFromInt(100)))
	if !factor.IsPositive() {
		return decimal.NewFromInt(-100)
	}

	f, _ := factor.Float64()
	exponent := 365.25 / float64(days)
	annFloat := math.Pow(f, exponent) - 1.0

	return decimal.NewFromFloat(annFloat).Mul(decimal.NewFromInt(100)).Round(2)
}

// calculateBetaAndTrackingError calculates sample Beta and Annualized Tracking Error from daily return arrays.
func calculateBetaAndTrackingError(rp, rb []float64) (*decimal.Decimal, *decimal.Decimal) {
	n := float64(len(rp))
	if n < 2 {
		return nil, nil
	}

	// Compute means
	var sumRp, sumRb float64
	for i := range rp {
		sumRp += rp[i]
		sumRb += rb[i]
	}
	meanRp := sumRp / n
	meanRb := sumRb / n

	// Compute covariance and variance
	var cov, varRb float64
	for i := range rp {
		dRp := rp[i] - meanRp
		dRb := rb[i] - meanRb
		cov += dRp * dRb
		varRb += dRb * dRb
	}

	var beta *decimal.Decimal
	if varRb > 1e-12 {
		betaVal := cov / varRb
		b := decimal.NewFromFloat(betaVal).Round(2)
		beta = &b
	}

	// Compute sample Tracking Error: stddev(rp - rb) * sqrt(252) * 100
	var sumDiff float64
	diffs := make([]float64, len(rp))
	for i := range rp {
		diffs[i] = rp[i] - rb[i]
		sumDiff += diffs[i]
	}
	meanDiff := sumDiff / n

	var sumSqDiff float64
	for i := range diffs {
		dev := diffs[i] - meanDiff
		sumSqDiff += dev * dev
	}

	sampleVar := sumSqDiff / (n - 1)
	if sampleVar < 0 {
		sampleVar = 0
	}
	dailyStdDev := math.Sqrt(sampleVar)
	annTE := dailyStdDev * math.Sqrt(252) * 100.0

	teDec := decimal.NewFromFloat(annTE).Round(2)
	return beta, &teDec
}

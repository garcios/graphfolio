package domain

import (
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type DividendFrequency string

const (
	DividendFreqMonthly    DividendFrequency = "MONTHLY"
	DividendFreqQuarterly  DividendFrequency = "QUARTERLY"
	DividendFreqSemiAnnual DividendFrequency = "SEMI_ANNUAL"
	DividendFreqAnnual     DividendFrequency = "ANNUAL"
	DividendFreqIrregular  DividendFrequency = "IRREGULAR"
)

type DividendStatus string

const (
	DividendStatusAnnounced  DividendStatus = "ANNOUNCED"
	DividendStatusExPassed   DividendStatus = "EX_DATE_PASSED"
	DividendStatusPaid       DividendStatus = "PAID"
	DividendStatusCancelled  DividendStatus = "CANCELLED"
)

// DividendSchedule represents an authoritative or synthesized forward/historical dividend event.
type DividendSchedule struct {
	ID                 uuid.UUID         `json:"id"`
	InstrumentID       uuid.UUID         `json:"instrument_id"`
	ExDate             time.Time         `json:"ex_date"`
	RecordDate         *time.Time        `json:"record_date,omitempty"`
	PaymentDate        time.Time         `json:"payment_date"`
	AmountPerShare     decimal.Decimal   `json:"amount_per_share"`
	CurrencyCode       string            `json:"currency_code"`
	Frequency          DividendFrequency `json:"frequency"`
	Status             DividendStatus    `json:"status"`
	IsEstimated        bool              `json:"is_estimated"`
	IsOverride         bool              `json:"is_override"`
	FrankingPercentage decimal.Decimal   `json:"franking_percentage"`
	WithholdingTaxPct  decimal.Decimal   `json:"withholding_tax_pct"`
	Source             string            `json:"source"`
	CreatedAt          time.Time         `json:"created_at"`
	UpdatedAt          time.Time         `json:"updated_at"`
}

// DividendSyncResult contains telemetry returned after a dividend ingestion run.
type DividendSyncResult struct {
	Success              bool     `json:"success"`
	SchedulesSynced      int      `json:"schedules_synced"`
	PrimarySourceCount   int      `json:"primary_source_count"`
	FallbackSourceCount  int      `json:"fallback_source_count"`
	PlaceholdersUpgraded int      `json:"placeholders_upgraded"`
	Message              string   `json:"message"`
	DurationMs           int64    `json:"duration_ms"`
	FailedSymbols        []string `json:"failed_symbols"`
}

// DividendSyncOptions provides filters for targeted manual dividend ingestion.
type DividendSyncOptions struct {
	Symbols        []string
	ExchangeCode   *string
	SyncHistorical bool
	ForceProvider  string // 'YAHOO', 'TWELVE_DATA', or empty for default cascade
}

// DividendCalendarEvent represents a single ex-date or payment date event on the investor's calendar.
type DividendCalendarEvent struct {
	ID               uuid.UUID         `json:"id"`
	InstrumentID     uuid.UUID         `json:"instrument_id"`
	Symbol           string            `json:"symbol"`
	Name             string            `json:"name"`
	ExDate           time.Time         `json:"ex_date"`
	RecordDate       *time.Time        `json:"record_date,omitempty"`
	PaymentDate      time.Time         `json:"payment_date"`
	SharesHeld       decimal.Decimal   `json:"shares_held"`
	DPSDeclared      Money             `json:"dps_declared"`      // In local instrument currency
	GrossAmountBase  Money             `json:"gross_amount_base"` // In portfolio base currency
	NetAmountBase    Money             `json:"net_amount_base"`   // After withholding tax
	Frequency        DividendFrequency `json:"frequency"`
	Status           DividendStatus    `json:"status"`
	IsUserEntitled   bool              `json:"is_user_entitled"`
	IsEstimated      bool              `json:"is_estimated"`
	FrankingPct      decimal.Decimal   `json:"franking_pct"`
}

// MonthlyDividendProjection aggregates forward cash flows for a specific calendar month.
type MonthlyDividendProjection struct {
	Year            int   `json:"year"`
	Month           int   `json:"month"` // 1..12
	ProjectedIncome Money `json:"projected_income"`
	DividendCount   int   `json:"dividend_count"`
}

// HoldingYieldItem presents forward dividend metrics for a single portfolio holding.
type HoldingYieldItem struct {
	InstrumentID        uuid.UUID         `json:"instrument_id"`
	Symbol              string            `json:"symbol"`
	Name                string            `json:"name"`
	Quantity            decimal.Decimal   `json:"quantity"`
	CurrentPrice        Money             `json:"current_price"`
	AverageBuyPrice     Money             `json:"average_buy_price"`
	CurrentMarketValue  Money             `json:"current_market_value"`
	CostBasis           Money             `json:"cost_basis"`
	AnnualDPS           Money             `json:"annual_dps"`
	ProjectedAnnualCash Money             `json:"projected_annual_cash"`
	CurrentYieldPct     decimal.Decimal   `json:"current_yield_pct"`
	YieldOnCostPct      *decimal.Decimal  `json:"yield_on_cost_pct"` // Nil when cost basis <= 0
	DividendCAGR3Y      *decimal.Decimal  `json:"dividend_cagr_3y"`  // Nil when insufficient history
	Frequency           DividendFrequency `json:"frequency"`
	NextExDate          *time.Time        `json:"next_ex_date,omitempty"`
	NextPaymentDate     *time.Time        `json:"next_payment_date,omitempty"`
}

// DividendYieldAnalytics aggregates top-level passive income metrics across the portfolio.
type DividendYieldAnalytics struct {
	AnnualForwardIncome   Money                       `json:"annual_forward_income"`
	MonthlyAverageIncome  Money                       `json:"monthly_average_income"`
	PortfolioCurrentYield decimal.Decimal             `json:"portfolio_current_yield"`
	PortfolioYieldOnCost  *decimal.Decimal            `json:"portfolio_yield_on_cost"`
	PendingCashInflow     Money                       `json:"pending_cash_inflow"`
	MonthlyProjections    []MonthlyDividendProjection `json:"monthly_projections"`
	HoldingsYield         []HoldingYieldItem          `json:"holdings_yield"`
}

// DRIPSimulationInput provides parameters for compounding simulation.
type DRIPSimulationInput struct {
	HorizonYears           int             `json:"horizon_years"`             // 1..30
	DividendGrowthRatePct  decimal.Decimal `json:"dividend_growth_rate_pct"`  // Default 5.0%
	CapitalAppreciationPct decimal.Decimal `json:"capital_appreciation_pct"` // Default 7.0%
}

// DRIPSimulationPoint captures values at yearly intervals.
type DRIPSimulationPoint struct {
	Year                  int             `json:"year"`
	DRIPPortfolioValue    decimal.Decimal `json:"drip_portfolio_value"`
	CashPortfolioValue    decimal.Decimal `json:"cash_portfolio_value"`
	DRIPAccumulatedShares decimal.Decimal `json:"drip_accumulated_shares"`
	CashAccumulatedCash   decimal.Decimal `json:"cash_accumulated_cash"`
	DRIPAnnualDividend    decimal.Decimal `json:"drip_annual_dividend"`
	CashAnnualDividend    decimal.Decimal `json:"cash_annual_dividend"`
}

// DRIPSimulationResult packages the dual compounding projection curves.
type DRIPSimulationResult struct {
	StartingPortfolioValue decimal.Decimal       `json:"starting_portfolio_value"`
	StartingAnnualDividend decimal.Decimal       `json:"starting_annual_dividend"`
	BaseCurrency           string                `json:"base_currency"`
	HorizonYears           int                   `json:"horizon_years"`
	Points                 []DRIPSimulationPoint `json:"points"`
	TerminalDRIPValue      decimal.Decimal       `json:"terminal_drip_value"`
	TerminalCashValue      decimal.Decimal       `json:"terminal_cash_value"`
	DRIPBenefitAmount      decimal.Decimal       `json:"drip_benefit_amount"`
}

// AnnualizeDPS returns the forward annualized DPS based on distribution frequency.
func AnnualizeDPS(dps decimal.Decimal, freq DividendFrequency) decimal.Decimal {
	switch freq {
	case DividendFreqMonthly:
		return dps.Mul(decimal.NewFromInt(12))
	case DividendFreqQuarterly:
		return dps.Mul(decimal.NewFromInt(4))
	case DividendFreqSemiAnnual:
		return dps.Mul(decimal.NewFromInt(2))
	case DividendFreqAnnual:
		return dps.Mul(decimal.NewFromInt(1))
	default:
		return dps.Mul(decimal.NewFromInt(4))
	}
}

// CalculateForwardYieldAnalytics computes forward dividend metrics across open holdings.
func CalculateForwardYieldAnalytics(
	holdings []HoldingWithPrice,
	latestSchedules map[uuid.UUID]DividendSchedule,
	forward12MSchedules []DividendSchedule,
	historicalSchedules map[uuid.UUID][]DividendSchedule,
	pendingSchedules []DividendSchedule,
	entitledSharesMap map[uuid.UUID]decimal.Decimal,
	fxRates map[string]decimal.Decimal,
	baseCurrency string,
	now time.Time,
) *DividendYieldAnalytics {
	hundred := decimal.NewFromInt(100)
	totalForwardAnnualIncome := decimal.Zero
	totalMarketValueBase := decimal.Zero
	totalCostBasisBase := decimal.Zero
	var yieldItems []HoldingYieldItem

	holdingsMap := make(map[uuid.UUID]HoldingWithPrice)
	for _, h := range holdings {
		if !h.Quantity.IsPositive() {
			continue
		}
		holdingsMap[h.InstrumentID] = h

		fx := h.FXRateToBase
		if !fx.IsPositive() {
			if r, ok := fxRates[h.InstrumentCurrency]; ok && r.IsPositive() {
				fx = r
			} else {
				fx = decimal.NewFromInt(1)
			}
		}

		mvBase := h.Quantity.Mul(h.LatestPrice).Mul(fx)
		cbBase := h.CostBasisBase
		if !cbBase.IsPositive() && h.CostBasis.IsPositive() {
			cbBase = h.CostBasis.Mul(fx)
		}

		totalMarketValueBase = totalMarketValueBase.Add(mvBase)
		totalCostBasisBase = totalCostBasisBase.Add(cbBase)

		sched, hasSched := latestSchedules[h.InstrumentID]
		freq := DividendFreqQuarterly
		dps := decimal.Zero
		var nextEx, nextPay *time.Time
		if hasSched {
			freq = sched.Frequency
			dps = sched.AmountPerShare
			tEx := sched.ExDate
			tPay := sched.PaymentDate
			nextEx = &tEx
			nextPay = &tPay
		}

		divFX := fx
		if hasSched && sched.CurrencyCode != "" && sched.CurrencyCode != h.InstrumentCurrency {
			if r, ok := fxRates[sched.CurrencyCode]; ok && r.IsPositive() {
				divFX = r
			}
		}

		annualDPS := AnnualizeDPS(dps, freq)
		projCashBase := h.Quantity.Mul(annualDPS).Mul(divFX)
		totalForwardAnnualIncome = totalForwardAnnualIncome.Add(projCashBase)

		currYield := decimal.Zero
		if h.LatestPrice.IsPositive() {
			currYield = annualDPS.DivRound(h.LatestPrice, 6).Mul(hundred)
		}

		var yoc *decimal.Decimal
		avgBuyPrice := decimal.Zero
		if h.Quantity.IsPositive() && h.CostBasis.IsPositive() {
			avgBuyPrice = h.CostBasis.DivRound(h.Quantity, 4)
			if avgBuyPrice.IsPositive() {
				yocVal := annualDPS.DivRound(avgBuyPrice, 6).Mul(hundred)
				yoc = &yocVal
			}
		}

		// Calculate 3Y Dividend CAGR if available
		var cagr3y *decimal.Decimal
		if histList, ok := historicalSchedules[h.InstrumentID]; ok && len(histList) > 0 {
			threeYearsAgo := now.AddDate(-3, -2, 0)
			var oldDPS decimal.Decimal
			for _, oldSched := range histList {
				if oldSched.ExDate.After(threeYearsAgo) && oldSched.ExDate.Before(threeYearsAgo.AddDate(0, 4, 0)) && oldSched.AmountPerShare.IsPositive() {
					oldDPS = oldSched.AmountPerShare
					break
				}
			}
			if oldDPS.IsPositive() && dps.IsPositive() {
				ratio := dps.DivRound(oldDPS, 6).InexactFloat64()
				if ratio > 0 {
					// CAGR = ratio^(1/3) - 1
					cagrVal := decimal.NewFromFloat((pow1_3(ratio) - 1.0) * 100).Round(2)
					cagr3y = &cagrVal
				}
			}
		}

		yieldItems = append(yieldItems, HoldingYieldItem{
			InstrumentID:        h.InstrumentID,
			Symbol:              h.Ticker,
			Name:                h.Name,
			Quantity:            h.Quantity,
			CurrentPrice:        NewMoney(h.LatestPrice, h.InstrumentCurrency),
			AverageBuyPrice:     NewMoney(avgBuyPrice, h.InstrumentCurrency),
			CurrentMarketValue:  NewMoney(mvBase, baseCurrency),
			CostBasis:           NewMoney(cbBase, baseCurrency),
			AnnualDPS:           NewMoney(annualDPS, h.InstrumentCurrency),
			ProjectedAnnualCash: NewMoney(projCashBase, baseCurrency),
			CurrentYieldPct:     currYield,
			YieldOnCostPct:      yoc,
			DividendCAGR3Y:      cagr3y,
			Frequency:           freq,
			NextExDate:          nextEx,
			NextPaymentDate:     nextPay,
		})
	}

	monthlyAverageIncome := totalForwardAnnualIncome.DivRound(decimal.NewFromInt(12), 2)

	portfolioCurrentYield := decimal.Zero
	if totalMarketValueBase.IsPositive() {
		portfolioCurrentYield = totalForwardAnnualIncome.DivRound(totalMarketValueBase, 6).Mul(hundred)
	}

	var portfolioYOC *decimal.Decimal
	if totalCostBasisBase.IsPositive() {
		val := totalForwardAnnualIncome.DivRound(totalCostBasisBase, 6).Mul(hundred)
		portfolioYOC = &val
	}

	// 12-Month Rolling Monthly Projections
	monthlyProjections := make([]MonthlyDividendProjection, 12)
	curYear := now.Year()
	curMonth := int(now.Month())
	for i := 0; i < 12; i++ {
		m := curMonth + i
		y := curYear
		for m > 12 {
			m -= 12
			y++
		}
		monthlyProjections[i] = MonthlyDividendProjection{
			Year:            y,
			Month:           m,
			ProjectedIncome: ZeroMoney(baseCurrency),
			DividendCount:   0,
		}
	}

	for _, s := range forward12MSchedules {
		h, exists := holdingsMap[s.InstrumentID]
		if !exists || !h.Quantity.IsPositive() {
			continue
		}
		fx := decimal.NewFromInt(1)
		if h.FXRateToBase.IsPositive() && (s.CurrencyCode == "" || s.CurrencyCode == h.InstrumentCurrency) {
			fx = h.FXRateToBase
		} else if r, ok := fxRates[s.CurrencyCode]; ok && r.IsPositive() {
			fx = r
		}
		amountBase := h.Quantity.Mul(s.AmountPerShare).Mul(fx)

		sYear := s.PaymentDate.Year()
		sMonth := int(s.PaymentDate.Month())
		for idx := range monthlyProjections {
			if monthlyProjections[idx].Year == sYear && monthlyProjections[idx].Month == sMonth {
				monthlyProjections[idx].ProjectedIncome.Amount = monthlyProjections[idx].ProjectedIncome.Amount.Add(amountBase)
				monthlyProjections[idx].DividendCount++
				break
			}
		}
	}

	// Pending Cash Inflow calculation
	pendingInflowBase := decimal.Zero
	for _, s := range pendingSchedules {
		shares, ok := entitledSharesMap[s.ID]
		if !ok || !shares.IsPositive() {
			continue
		}
		h, exists := holdingsMap[s.InstrumentID]
		fx := decimal.NewFromInt(1)
		if exists && h.FXRateToBase.IsPositive() && (s.CurrencyCode == "" || s.CurrencyCode == h.InstrumentCurrency) {
			fx = h.FXRateToBase
		} else if r, ok := fxRates[s.CurrencyCode]; ok && r.IsPositive() {
			fx = r
		}
		whtRate := s.WithholdingTaxPct.DivRound(hundred, 4)
		netDPS := s.AmountPerShare.Mul(decimal.NewFromInt(1).Sub(whtRate))
		netAmountBase := shares.Mul(netDPS).Mul(fx)
		pendingInflowBase = pendingInflowBase.Add(netAmountBase)
	}

	return &DividendYieldAnalytics{
		AnnualForwardIncome:   NewMoney(totalForwardAnnualIncome, baseCurrency),
		MonthlyAverageIncome:  NewMoney(monthlyAverageIncome, baseCurrency),
		PortfolioCurrentYield: portfolioCurrentYield,
		PortfolioYieldOnCost:  portfolioYOC,
		PendingCashInflow:     NewMoney(pendingInflowBase, baseCurrency),
		MonthlyProjections:    monthlyProjections,
		HoldingsYield:         yieldItems,
	}
}

func pow1_3(x float64) float64 {
	if x <= 0 {
		return 0
	}
	return math.Cbrt(x)
}

// SimulateDRIPCompounding models dual compounding strategies: 100% DRIP vs Cash Payout.
func SimulateDRIPCompounding(
	startingPortfolioValue decimal.Decimal,
	startingAnnualDividend decimal.Decimal,
	input DRIPSimulationInput,
	baseCurrency string,
) *DRIPSimulationResult {
	horizon := input.HorizonYears
	if horizon < 1 {
		horizon = 1
	} else if horizon > 30 {
		horizon = 30
	}

	divGrowthPct := input.DividendGrowthRatePct
	if divGrowthPct.IsZero() {
		divGrowthPct = decimal.RequireFromString("5.0")
	}
	capGrowthPct := input.CapitalAppreciationPct
	if capGrowthPct.IsZero() {
		capGrowthPct = decimal.RequireFromString("7.0")
	}

	divGrowth := divGrowthPct.DivRound(decimal.NewFromInt(100), 6).InexactFloat64()
	capGrowth := capGrowthPct.DivRound(decimal.NewFromInt(100), 6).InexactFloat64()

	// Initial synthetic share basis: price P0 = 100, shares Q0 = startingValue / 100
	p0 := 100.0
	startValFloat := startingPortfolioValue.InexactFloat64()
	if startValFloat <= 0 {
		startValFloat = 10000.0
	}
	q0 := startValFloat / p0

	startDivFloat := startingAnnualDividend.InexactFloat64()
	if startDivFloat <= 0 {
		startDivFloat = startValFloat * 0.04 // default 4% yield baseline
	}
	monthlyDPS0 := startDivFloat / (12.0 * q0)

	qDRIP := q0
	cashPool := 0.0

	var points []DRIPSimulationPoint

	totalMonths := horizon * 12
	for m := 1; m <= totalMonths; m++ {
		tYears := float64(m) / 12.0
		// Price appreciation: P_m = P_0 * (1 + g)^t
		pm := p0 * powFloat(1.0+capGrowth, tYears)
		// Dividend growth: DPS_m = DPS_0 * (1 + d)^t
		dpsm := monthlyDPS0 * powFloat(1.0+divGrowth, tYears)

		// 1. DRIP
		monthlyPayoutDRIP := qDRIP * dpsm
		deltaQ := monthlyPayoutDRIP / pm
		qDRIP += deltaQ

		// 2. Cash
		monthlyPayoutCash := q0 * dpsm
		cashPool += monthlyPayoutCash

		if m%12 == 0 {
			yr := m / 12
			dripValue := decimal.NewFromFloat(qDRIP * pm).Round(2)
			cashValue := decimal.NewFromFloat((q0 * pm) + cashPool).Round(2)
			dripAnnualDiv := decimal.NewFromFloat(qDRIP * dpsm * 12.0).Round(2)
			cashAnnualDiv := decimal.NewFromFloat(q0 * dpsm * 12.0).Round(2)

			points = append(points, DRIPSimulationPoint{
				Year:                  yr,
				DRIPPortfolioValue:    dripValue,
				CashPortfolioValue:    cashValue,
				DRIPAccumulatedShares: decimal.NewFromFloat(qDRIP).Round(4),
				CashAccumulatedCash:   decimal.NewFromFloat(cashPool).Round(2),
				DRIPAnnualDividend:    dripAnnualDiv,
				CashAnnualDividend:    cashAnnualDiv,
			})
		}
	}

	terminalDRIP := decimal.Zero
	terminalCash := decimal.Zero
	if len(points) > 0 {
		terminalDRIP = points[len(points)-1].DRIPPortfolioValue
		terminalCash = points[len(points)-1].CashPortfolioValue
	}
	benefit := terminalDRIP.Sub(terminalCash)

	return &DRIPSimulationResult{
		StartingPortfolioValue: startingPortfolioValue,
		StartingAnnualDividend: startingAnnualDividend,
		BaseCurrency:           baseCurrency,
		HorizonYears:           horizon,
		Points:                 points,
		TerminalDRIPValue:      terminalDRIP,
		TerminalCashValue:      terminalCash,
		DRIPBenefitAmount:      benefit,
	}
}

func powFloat(base, exp float64) float64 {
	if base <= 0 {
		return 0
	}
	return math.Pow(base, exp)
}


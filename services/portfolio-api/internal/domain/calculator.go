package domain

import (
	"github.com/shopspring/decimal"
)

var (
	oneHundred = decimal.NewFromInt(100)
	one        = decimal.NewFromInt(1)
)

// CalculateInvestment computes current valuation, today's return, and total return metrics
// for a single instrument holding using exact decimal arithmetic.
func CalculateInvestment(h HoldingWithPrice, baseCurrency string) InvestmentSummary {
	totalValueInst := h.Quantity.Mul(h.LatestPrice)
	fxRate := h.FXRateToBase
	if fxRate.IsZero() {
		fxRate = one
	}

	totalValueBase := totalValueInst.Mul(fxRate)

	// Today's price movement in instrument currency
	priceDelta := h.LatestPrice.Sub(h.PrevPrice)
	todayReturnAmountInst := h.Quantity.Mul(priceDelta)

	// Today's return percentage
	var todayReturnPercent decimal.Decimal
	if h.PrevPrice.IsPositive() {
		todayReturnPercent = priceDelta.DivRound(h.PrevPrice, 6).Mul(oneHundred).Round(2)
	}

	// Total return = Current Value (Base) - Cost Basis (Base) + Realized PnL (Base) + Dividends (Base)
	totalReturnAmountBase := totalValueBase.Sub(h.CostBasisBase).Add(h.RealizedPnLBase).Add(h.DividendsBase)

	var totalReturnPercent decimal.Decimal
	if h.CostBasisBase.IsPositive() {
		totalReturnPercent = totalReturnAmountBase.DivRound(h.CostBasisBase, 6).Mul(oneHundred).Round(1)
	}

	return InvestmentSummary{
		ID:                 h.InstrumentID.String(),
		Ticker:             h.Ticker,
		Name:               h.Name,
		Price:              NewMoney(h.LatestPrice.Round(2), h.InstrumentCurrency),
		Quantity:           h.Quantity,
		TotalValue:         NewMoney(totalValueInst.Round(2), h.InstrumentCurrency),
		TodayReturnAmount:  NewMoney(todayReturnAmountInst.Round(2), h.InstrumentCurrency),
		TodayReturnPercent: todayReturnPercent,
		TotalReturnAmount:  NewMoney(totalReturnAmountBase.Round(2), baseCurrency),
		TotalReturnPercent: totalReturnPercent,
	}
}

// CalculatePortfolioSummary aggregates holdings and cash balances into a consolidated portfolio overview.
func CalculatePortfolioSummary(
	portfolio Portfolio,
	holdings []HoldingWithPrice,
	cashBalances []CashBalance,
	latestValuation *PortfolioValuation,
	cashFXRates map[string]decimal.Decimal,
) PortfolioSummary {
	baseCurrency := portfolio.BaseCurrency

	investments := make([]InvestmentSummary, len(holdings))
	totalHoldingsValueBase := decimal.Zero
	todayReturnAmountBase := decimal.Zero

	for i, h := range holdings {
		inv := CalculateInvestment(h, baseCurrency)
		investments[i] = inv

		fxRate := h.FXRateToBase
		if fxRate.IsZero() {
			fxRate = one
		}

		// Holding value in base currency
		valBase := h.Quantity.Mul(h.LatestPrice).Mul(fxRate)
		totalHoldingsValueBase = totalHoldingsValueBase.Add(valBase)

		// Holding today's return in base currency
		priceDelta := h.LatestPrice.Sub(h.PrevPrice)
		retBase := h.Quantity.Mul(priceDelta).Mul(fxRate)
		todayReturnAmountBase = todayReturnAmountBase.Add(retBase)
	}

	// Calculate total cash in base currency
	totalCashBase := decimal.Zero
	for _, c := range cashBalances {
		fxRate := one
		if c.CurrencyCode != baseCurrency {
			if rate, ok := cashFXRates[c.CurrencyCode]; ok && rate.IsPositive() {
				fxRate = rate
			}
		}
		totalCashBase = totalCashBase.Add(c.Balance.Mul(fxRate))
	}

	totalPortfolioValue := totalHoldingsValueBase.Add(totalCashBase)

	// Portfolio today's return percentage = Return / Previous Valuation * 100
	var todayReturnPercent decimal.Decimal
	prevValuation := totalPortfolioValue.Sub(todayReturnAmountBase)
	if prevValuation.IsPositive() {
		todayReturnPercent = todayReturnAmountBase.DivRound(prevValuation, 6).Mul(oneHundred).Round(2)
	}

	// Annualized Return / TWR from latest valuation snapshot
	var annualizedReturnPercent decimal.Decimal
	if latestValuation != nil && latestValuation.TWRIndex.IsPositive() {
		// TWR index e.g. 1.1420 corresponds to 14.2%
		annualizedReturnPercent = latestValuation.TWRIndex.Sub(one).Mul(oneHundred).Round(1)
	}

	return PortfolioSummary{
		Portfolio:               portfolio,
		TotalValue:              NewMoney(totalPortfolioValue.Round(2), baseCurrency),
		TodayReturnAmount:       NewMoney(todayReturnAmountBase.Round(2), baseCurrency),
		TodayReturnPercent:      todayReturnPercent,
		AnnualizedReturnPercent: annualizedReturnPercent,
		CashBalance:             NewMoney(totalCashBase.Round(2), baseCurrency),
		Investments:             investments,
	}
}

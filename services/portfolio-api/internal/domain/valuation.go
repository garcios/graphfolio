package domain

import (
	"errors"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const (
	// TWRIndexScale defines the standard 12-decimal-place precision for TWR indices.
	TWRIndexScale = 12

	// DailyReturnScale defines the 12-decimal-place precision for daily returns.
	DailyReturnScale = 12

	// ValuationAmountScale defines the 6-decimal-place precision for valuation amounts.
	ValuationAmountScale = 6
)

var (
	// TWRInceptionBaseline is the standard starting baseline for Time-Weighted Return (1.000000000000).
	TWRInceptionBaseline = decimal.NewFromInt(1)
)

// PortfolioValuationSnapshot represents a daily valuation time-series snapshot for a portfolio,
// adhering to GIPS standards and exact fixed-point precision.
type PortfolioValuationSnapshot struct {
	PortfolioID     uuid.UUID
	ValuationDate   time.Time
	MarketValueBase decimal.Decimal
	CashValueBase   decimal.Decimal
	NetFlowBase     decimal.Decimal
	DailyReturn     *decimal.Decimal
	TWRIndex        decimal.Decimal
}

// TotalValue computes the consolidated portfolio value (Market Value + Cash Value).
func (s PortfolioValuationSnapshot) TotalValue() decimal.Decimal {
	return s.MarketValueBase.Add(s.CashValueBase)
}

// ToPortfolioValuation converts the snapshot to the database-compatible domain.PortfolioValuation struct.
func (s PortfolioValuationSnapshot) ToPortfolioValuation() PortfolioValuation {
	dailyReturn := decimal.Zero
	if s.DailyReturn != nil {
		dailyReturn = *s.DailyReturn
	}
	return PortfolioValuation{
		PortfolioID:     s.PortfolioID,
		ValuationDate:   s.ValuationDate,
		MarketValueBase: s.MarketValueBase,
		CashValueBase:   s.CashValueBase,
		NetFlowBase:     s.NetFlowBase,
		DailyReturn:     dailyReturn,
		TWRIndex:        s.TWRIndex,
	}
}

// NewPortfolioValuationSnapshot constructs a PortfolioValuationSnapshot from a PortfolioValuation.
func NewPortfolioValuationSnapshot(v PortfolioValuation) PortfolioValuationSnapshot {
	dailyRet := v.DailyReturn
	return PortfolioValuationSnapshot{
		PortfolioID:     v.PortfolioID,
		ValuationDate:   v.ValuationDate,
		MarketValueBase: v.MarketValueBase,
		CashValueBase:   v.CashValueBase,
		NetFlowBase:     v.NetFlowBase,
		DailyReturn:     &dailyRet,
		TWRIndex:        v.TWRIndex,
	}
}

// CalculateHoldingMarketValue computes the base currency value of an instrument holding:
// Value = Quantity * Price * FX(InstrumentCurrency -> BaseCurrency)
func CalculateHoldingMarketValue(quantity, price, fxRateToBase decimal.Decimal) decimal.Decimal {
	fx := fxRateToBase
	if fx.IsZero() {
		fx = decimal.NewFromInt(1)
	}
	return quantity.Mul(price).Mul(fx)
}

// CalculateCashValue computes the base currency value of a cash balance:
// Value = Balance * FX(Currency -> BaseCurrency)
func CalculateCashValue(balance, fxRateToBase decimal.Decimal) decimal.Decimal {
	fx := fxRateToBase
	if fx.IsZero() {
		fx = decimal.NewFromInt(1)
	}
	return balance.Mul(fx)
}

// CalculateTotalValue computes the total portfolio value:
// V_t = MarketValue_t + CashValue_t
func CalculateTotalValue(marketValueBase, cashValueBase decimal.Decimal) decimal.Decimal {
	return marketValueBase.Add(cashValueBase)
}

// CalculateNetCashFlow computes the net external cash flow:
// F_t = Deposits_t - Withdrawals_t
func CalculateNetCashFlow(depositsBase, withdrawalsBase decimal.Decimal) decimal.Decimal {
	return depositsBase.Sub(withdrawalsBase)
}

// CalculateDailyReturn computes sub-period daily return adjusting for external net flows:
// R_t = ((V_t - F_t) - V_{t-1}) / V_{t-1}
//
// Boundary Cases:
// - If V_{t-1} < 0: returns nil (undefined return under margin constraint/deficit).
// - If V_{t-1} == 0: returns pointer to 0 (initial funding or zero start produces no return).
// - If V_{t-1} > 0 and V_t == 0: full liquidation, R_t = (-F_t - V_{t-1}) / V_{t-1}.
// - If V_{t-1} > 0: R_t = ((V_t - F_t) - V_{t-1}) / V_{t-1} rounded to 12 decimal places.
func CalculateDailyReturn(currentVal, previousVal, netFlow decimal.Decimal) *decimal.Decimal {
	if previousVal.IsNegative() {
		return nil
	}
	if previousVal.IsZero() {
		zero := decimal.Zero
		return &zero
	}

	// Isolated market movement = (CurrentValuation - NetExternalFlow) - PrevValuation
	netGain := currentVal.Sub(netFlow).Sub(previousVal)
	dailyReturn := netGain.DivRound(previousVal, DailyReturnScale)
	return &dailyReturn
}

// ChainTWR calculates the updated TWR index from the previous index and daily return:
// TWR_t = TWR_{t-1} * (1 + R_t)
// If previousTWR is zero or negative (inception), it defaults to TWRInceptionBaseline (1.0).
// If dailyReturn is nil, previousTWR is preserved and rounded to 12 decimal places.
func ChainTWR(previousTWR decimal.Decimal, dailyReturn *decimal.Decimal) decimal.Decimal {
	prev := previousTWR
	if !prev.IsPositive() {
		prev = TWRInceptionBaseline
	}
	if dailyReturn == nil {
		return prev.Round(TWRIndexScale)
	}
	return prev.Mul(decimal.NewFromInt(1).Add(*dailyReturn)).Round(TWRIndexScale)
}

// CalculateAnnualizedReturn computes the GIPS-compliant annualized return percentage.
// Per GIPS Standard 2.A.20, returns for periods less than 365 calendar days MUST NOT be annualized;
// for such periods, isAnnualized returns false and the cumulative period return is returned.
func CalculateAnnualizedReturn(twrIndex decimal.Decimal, days int) (annualized decimal.Decimal, isAnnualized bool, err error) {
	if !twrIndex.IsPositive() {
		return decimal.Zero, false, errors.New("twrIndex must be positive")
	}
	if days <= 0 {
		return decimal.Zero, false, errors.New("days must be positive")
	}

	// GIPS Standard 2.A.20: Periods under 1 year (< 365 days) MUST NOT be annualized.
	if days < 365 {
		periodReturn := twrIndex.Sub(decimal.NewFromInt(1)).Mul(decimal.NewFromInt(100)).Round(2)
		return periodReturn, false, nil
	}

	// Geometric compound annualization: (TWRIndex ^ (365.25 / days) - 1) * 100
	twrFloat, _ := twrIndex.Float64()
	exponent := 365.25 / float64(days)
	annFactor := math.Pow(twrFloat, exponent)

	annDec := decimal.NewFromFloat(annFactor).Sub(decimal.NewFromInt(1)).Mul(decimal.NewFromInt(100)).Round(2)
	return annDec, true, nil
}

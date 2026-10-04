package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type HistoryTimeframe string

const (
	Timeframe1D  HistoryTimeframe = "1D"
	Timeframe1W  HistoryTimeframe = "1W"
	Timeframe1M  HistoryTimeframe = "1M"
	Timeframe1Y  HistoryTimeframe = "1Y"
	TimeframeAll HistoryTimeframe = "ALL"
)

type ValuationPoint struct {
	Date        time.Time
	TotalValue  Money
	MarketValue Money
	CashValue   Money
	TWRIndex    decimal.Decimal
	DailyReturn decimal.Decimal
}

type PortfolioHistory struct {
	Points        []ValuationPoint
	StartValue    Money
	EndValue      Money
	ReturnAmount  Money
	ReturnPercent decimal.Decimal
}

// CalculatePortfolioHistory aggregates a time-ordered slice of valuations into a PortfolioHistory struct.
func CalculatePortfolioHistory(baseCurrency string, valuations []PortfolioValuation) *PortfolioHistory {
	if len(valuations) == 0 {
		zeroMoney := NewMoney(decimal.Zero, baseCurrency)
		return &PortfolioHistory{
			Points:        []ValuationPoint{},
			StartValue:    zeroMoney,
			EndValue:      zeroMoney,
			ReturnAmount:  zeroMoney,
			ReturnPercent: decimal.Zero,
		}
	}

	points := make([]ValuationPoint, len(valuations))
	for i, v := range valuations {
		totalVal := v.MarketValueBase.Add(v.CashValueBase)
		points[i] = ValuationPoint{
			Date:        v.ValuationDate,
			TotalValue:  NewMoney(totalVal.Round(2), baseCurrency),
			MarketValue: NewMoney(v.MarketValueBase.Round(2), baseCurrency),
			CashValue:   NewMoney(v.CashValueBase.Round(2), baseCurrency),
			TWRIndex:    v.TWRIndex,
			DailyReturn: v.DailyReturn,
		}
	}

	startVal := points[0].TotalValue
	endVal := points[len(points)-1].TotalValue
	returnAmt := endVal.Amount.Sub(startVal.Amount)

	var returnPct decimal.Decimal
	if startVal.Amount.IsPositive() {
		returnPct = returnAmt.DivRound(startVal.Amount, 6).Mul(oneHundred).Round(2)
	}

	return &PortfolioHistory{
		Points:        points,
		StartValue:    startVal,
		EndValue:      endVal,
		ReturnAmount:  NewMoney(returnAmt.Round(2), baseCurrency),
		ReturnPercent: returnPct,
	}
}

package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// FXRate represents a single historical foreign exchange mark.
type FXRate struct {
	BaseCurrency  string
	QuoteCurrency string
	RateDate      time.Time
	Rate          decimal.Decimal
	Source        string
}

// FXRateFilter provides filtering criteria for paginated FX queries.
type FXRateFilter struct {
	BaseCurrency  *string
	QuoteCurrency *string
	FromDate      *time.Time
	ToDate        *time.Time
	Limit         int
	Offset        int
}

// CurrencyPairSummary provides aggregate operational telemetry on a currency pair.
type CurrencyPairSummary struct {
	BaseCurrency   string
	QuoteCurrency  string
	LatestRate     decimal.Decimal
	LatestDate     time.Time
	LatestSource   string
	PreviousRate   *decimal.Decimal
	Change1DAmount *decimal.Decimal
	Change1DPct    *decimal.Decimal
	TotalRecords   int
	FirstDate      time.Time
	LastDate       time.Time
}

// CurrencyPairHistory aggregates chronological points and key range statistics.
type CurrencyPairHistory struct {
	BaseCurrency    string
	QuoteCurrency   string
	Points          []FXRatePoint
	StartRate       decimal.Decimal
	EndRate         decimal.Decimal
	PeriodChange    decimal.Decimal
	PeriodChangePct decimal.Decimal
	PeriodHigh      decimal.Decimal
	PeriodLow       decimal.Decimal
}

// FXRatePoint is a single point for charting and trend evaluation.
type FXRatePoint struct {
	Date         time.Time
	Rate         decimal.Decimal
	InvertedRate decimal.Decimal // 1 / Rate (10 decimal precision)
	Source       string
}

// FXRateOverrideInput captures administrative input to manually record or override an FX rate.
type FXRateOverrideInput struct {
	BaseCurrency        string
	QuoteCurrency       string
	RateDate            time.Time
	Rate                decimal.Decimal
	Reason              *string
	RecomputeValuations bool
}

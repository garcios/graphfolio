package domain

import "time"

// BackfillInput defines the parameters for a historical market data backfill job.
type BackfillInput struct {
	FromDate            time.Time
	ToDate              time.Time
	Symbols             []string
	CurrencyPairs       []string
	BackfillAssets      bool
	BackfillFX          bool
	RecomputeValuations bool
}

// BackfillResult summarizes the outcome of the backfill job.
type BackfillResult struct {
	Success       bool
	PricesSynced  int
	FXRatesSynced int
	Message       string
	Warnings      []string
}

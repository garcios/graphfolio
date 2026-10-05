package domain

import "time"

type FeedHealth struct {
	Name     string
	Status   string
	Provider string
	Schedule string
	LastRun  time.Time
	Details  string
}

type IngestionStatus struct {
	Feeds               []FeedHealth
	TrackedInstruments  int
	TrackedCurrencies   int
	LatestPriceDate     *time.Time
	LatestFXDate        *time.Time
	RateLimitRemaining  int
	RateLimitBudget     int
	PendingBackfillJobs int
}

type MarketSyncResult struct {
	Success       bool
	PricesSynced  int
	FXRatesSynced int
	Message       string
}

package marketdata

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// InstrumentRef represents an authoritative reference to an instrument for market data queries.
type InstrumentRef struct {
	ID       uuid.UUID
	Symbol   string
	Exchange string
	Currency string
}

// CurrencyPair represents an ISO 4217 base and quote currency tuple (e.g. Base: USD, Quote: EUR).
type CurrencyPair struct {
	Base  string
	Quote string
}

// PriceRecord represents an ingested closing price mark for an instrument.
type PriceRecord struct {
	InstrumentID uuid.UUID
	Symbol       string
	Exchange     string
	PriceDate    time.Time
	ClosePrice   decimal.Decimal
	Source       string
}

// FXRecord represents an ingested daily exchange rate for a currency pair.
type FXRecord struct {
	BaseCurrency  string
	QuoteCurrency string
	RateDate      time.Time
	Rate          decimal.Decimal
	Source        string
}

// PriceProvider defines the interface for fetching equity, ETF, and fund prices.
type PriceProvider interface {
	// Name returns the provider's human-readable identifier (e.g. "Twelve Data", "Yahoo Finance").
	Name() string
	// FetchClosingPrices retrieves EOD closing marks for a slice of instruments on a given date.
	FetchClosingPrices(ctx context.Context, instruments []InstrumentRef, date time.Time) ([]PriceRecord, error)
	// FetchHistoricalPrices retrieves daily closing marks across an inclusive date range [from, to].
	FetchHistoricalPrices(ctx context.Context, instrument InstrumentRef, from, to time.Time) ([]PriceRecord, error)
}

// FXRateProvider defines the interface for fetching central bank and interbank currency exchange rates.
type FXRateProvider interface {
	// Name returns the provider's human-readable identifier (e.g. "European Central Bank", "Open Exchange Rates").
	Name() string
	// FetchFXRates retrieves daily fixing rates for a slice of currency pairs on a given date.
	FetchFXRates(ctx context.Context, pairs []CurrencyPair, date time.Time) ([]FXRecord, error)
	// FetchHistoricalFXRates retrieves daily fixing rates for a currency pair across an inclusive date range [from, to].
	FetchHistoricalFXRates(ctx context.Context, pair CurrencyPair, from, to time.Time) ([]FXRecord, error)
}

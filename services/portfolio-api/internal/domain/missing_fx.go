package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// MissingFXRateItem represents a cross-currency transaction lacking an EOD FX rate on trade date.
type MissingFXRateItem struct {
	TransactionID    uuid.UUID
	PortfolioID      uuid.UUID
	PortfolioName    string
	PortfolioBaseCur string
	TradeDate        time.Time
	Type             TransactionType
	InstrumentID     *uuid.UUID
	Symbol           *string
	AssetName        *string
	Amount           Money
	TransactionCur   string
	RequiredPair     string // e.g. "USD/AUD"
	RecordedFXRate   *decimal.Decimal
}

// MissingFXRateFilter provides query filtering criteria for missing FX rate diagnostics.
type MissingFXRateFilter struct {
	CurrencyPair *string
	FromDate     *time.Time
	ToDate       *time.Time
	Type         *TransactionType
	Limit        int
	Offset       int
}

// MissingFXRatesReport aggregates paginated missing FX items with summary statistics.
type MissingFXRatesReport struct {
	Items             []MissingFXRateItem
	TotalCount        int
	UniquePairsCount  int
	EarliestTradeDate *time.Time
	LatestTradeDate   *time.Time
}

package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type MissingPriceItem struct {
	TransactionID  uuid.UUID
	PortfolioID    uuid.UUID
	InstrumentID   uuid.UUID
	Symbol         string
	AssetName      string
	ExchangeMIC    string
	TradeDate      time.Time
	Type           TransactionType
	Quantity       decimal.Decimal
	ExecutionPrice Money
	CurrencyCode   string
}

type MissingPriceFilter struct {
	Symbol   *string
	FromDate *time.Time
	ToDate   *time.Time
	Limit    int
	Offset   int
}

type MissingPriceReport struct {
	Items                  []MissingPriceItem
	TotalCount             int
	UniqueInstrumentsCount int
	EarliestTradeDate      *time.Time
	LatestTradeDate        *time.Time
}

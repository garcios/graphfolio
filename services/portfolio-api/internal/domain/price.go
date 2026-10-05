package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type InstrumentPrice struct {
	InstrumentID uuid.UUID
	Symbol       string
	PriceDate    time.Time
	Close        decimal.Decimal
	CurrencyCode string
	Source       string
	CreatedAt    time.Time
}

type PriceFilter struct {
	Symbol   *string
	FromDate *time.Time
	ToDate   *time.Time
	Limit    int
	Offset   int
}

type PriceOverrideInput struct {
	Symbol              string
	PriceDate           time.Time
	Price               decimal.Decimal
	Reason              *string
	RecomputeValuations bool
}

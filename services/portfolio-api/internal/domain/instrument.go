package domain

import "github.com/google/uuid"

type Instrument struct {
	ID           uuid.UUID
	Symbol       string
	ExchangeCode string
	Name         string
	AssetClass   string
	CurrencyCode string
	ISIN         *string
	IsActive     bool
}

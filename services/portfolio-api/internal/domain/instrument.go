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

type CreateInstrumentInput struct {
	Symbol       string
	ExchangeCode string
	Name         string
	AssetClass   string
	CurrencyCode string
	ISIN         *string
}

type UpdateInstrumentInput struct {
	ID       uuid.UUID
	Name     *string
	IsActive *bool
	ISIN     *string
}

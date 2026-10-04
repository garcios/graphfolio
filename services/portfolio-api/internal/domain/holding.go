package domain

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type HoldingWithPrice struct {
	PortfolioID        uuid.UUID
	InstrumentID       uuid.UUID
	Ticker             string
	Name               string
	InstrumentCurrency string
	Quantity           decimal.Decimal
	CostBasis          decimal.Decimal
	CostBasisBase      decimal.Decimal
	RealizedPnLBase    decimal.Decimal
	DividendsBase      decimal.Decimal
	LatestPrice        decimal.Decimal
	PrevPrice          decimal.Decimal
	FXRateToBase       decimal.Decimal
}

type InvestmentSummary struct {
	ID                 string
	Ticker             string
	Name               string
	Price              Money
	Quantity           decimal.Decimal
	TotalValue         Money
	TodayReturnAmount  Money
	TodayReturnPercent decimal.Decimal
	TotalReturnAmount  Money
	TotalReturnPercent decimal.Decimal
}

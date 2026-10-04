package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CostBasisMethod string

const (
	CostBasisMethodAverageCost CostBasisMethod = "AVERAGE_COST"
	CostBasisMethodFIFO        CostBasisMethod = "FIFO"
)

type Portfolio struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	Name            string
	BaseCurrency    string
	CostBasisMethod CostBasisMethod
	CreatedAt       time.Time
}

type CashBalance struct {
	PortfolioID  uuid.UUID
	CurrencyCode string
	Balance      decimal.Decimal
}

type PortfolioValuation struct {
	PortfolioID     uuid.UUID
	ValuationDate   time.Time
	MarketValueBase decimal.Decimal
	CashValueBase   decimal.Decimal
	NetFlowBase     decimal.Decimal
	DailyReturn     decimal.Decimal
	TWRIndex        decimal.Decimal
}

type PortfolioSummary struct {
	Portfolio               Portfolio
	TotalValue              Money
	TodayReturnAmount       Money
	TodayReturnPercent      decimal.Decimal
	AnnualizedReturnPercent decimal.Decimal
	CashBalance             Money
	Investments             []InvestmentSummary
}

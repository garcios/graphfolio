package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TaxLotStatus string

const (
	TaxLotStatusUnspecified TaxLotStatus = "UNSPECIFIED"
	TaxLotStatusOpen        TaxLotStatus = "OPEN"
	TaxLotStatusClosed      TaxLotStatus = "CLOSED"
	TaxLotStatusAll         TaxLotStatus = "ALL"
)

type HoldingTermClassification string

const (
	TermUnspecified HoldingTermClassification = "UNSPECIFIED"
	TermShortTerm   HoldingTermClassification = "SHORT_TERM"
	TermLongTerm    HoldingTermClassification = "LONG_TERM"
)

type TaxLot struct {
	ID                     uuid.UUID
	PortfolioID            uuid.UUID
	InstrumentID           uuid.UUID
	Ticker                 string
	Name                   string
	OpenTransactionID      uuid.UUID
	AcquiredDate           time.Time
	ClosedDate             *time.Time
	OriginalQuantity       decimal.Decimal
	RemainingQuantity      decimal.Decimal
	CostBasis              decimal.Decimal
	CostBasisBase          decimal.Decimal
	UnitCostBase           decimal.Decimal
	CurrentPrice           decimal.Decimal
	MarketValueBase        decimal.Decimal
	UnrealizedPnLBase      decimal.Decimal
	UnrealizedPnLPercent   decimal.Decimal
	HoldingPeriodDays      int
	TermClassification     HoldingTermClassification
	IsCGTDiscountEligible  bool
	IsLossHarvestCandidate bool
	HasWashSaleRisk        bool
}

type LotDisposal struct {
	ID                       uuid.UUID
	TaxLotID                 uuid.UUID
	SellTransactionID        uuid.UUID
	InstrumentID             uuid.UUID
	Ticker                   string
	Name                     string
	AcquiredDate             time.Time
	DisposedDate             time.Time
	Quantity                 decimal.Decimal
	CostBasisReleased        decimal.Decimal
	CostBasisReleasedBase    decimal.Decimal
	ProceedsBase             decimal.Decimal
	RealizedPnLBase          decimal.Decimal
	RealizedPnLPercent       decimal.Decimal
	HoldingPeriodDays        int
	TermClassification       HoldingTermClassification
	IsCGTDiscountEligible    bool
	DiscountedTaxablePnLBase decimal.Decimal
}

// CalculateHoldingDays computes elapsed calendar days between acquisition date and evaluation/disposal date.
func CalculateHoldingDays(acqDate, evalDate time.Time) int {
	d1 := time.Date(acqDate.Year(), acqDate.Month(), acqDate.Day(), 0, 0, 0, 0, time.UTC)
	d2 := time.Date(evalDate.Year(), evalDate.Month(), evalDate.Day(), 0, 0, 0, 0, time.UTC)
	if d2.Before(d1) {
		return 0
	}
	return int(d2.Sub(d1).Hours() / 24)
}

// IsCGTDiscountEligible checks if an asset held from acqDate to evalOrDisposalDate qualifies
// for the 12-month / 365-day CGT discount concession (e.g. ATO ITAA 1997 s 115-25: acquisition day is excluded,
// requiring at least 12 whole months held before disposal).
func IsCGTDiscountEligible(acqDate, evalOrDisposalDate time.Time) bool {
	d1 := time.Date(acqDate.Year(), acqDate.Month(), acqDate.Day(), 0, 0, 0, 0, time.UTC)
	d2 := time.Date(evalOrDisposalDate.Year(), evalOrDisposalDate.Month(), evalOrDisposalDate.Day(), 0, 0, 0, 0, time.UTC)

	days := int(d2.Sub(d1).Hours() / 24)
	if days < 365 {
		return false
	}

	// Under ATO s 115-25: acquisition date is excluded, so disposal must be on or after 1 year + 1 day
	minEligibleDate := d1.AddDate(1, 0, 1)
	return !d2.Before(minEligibleDate)
}

// ClassifyHoldingTerm determines whether a parcel is Short-Term or Long-Term.
func ClassifyHoldingTerm(acqDate, evalOrDisposalDate time.Time) HoldingTermClassification {
	if IsCGTDiscountEligible(acqDate, evalOrDisposalDate) {
		return TermLongTerm
	}
	return TermShortTerm
}

// IsLossHarvestingCandidate flags underwater lots with unrealized loss > $500 and loss percent <= -10%.
func IsLossHarvestingCandidate(unrealizedPnLBase, costBasisBase decimal.Decimal) bool {
	if !unrealizedPnLBase.IsNegative() || costBasisBase.IsZero() {
		return false
	}
	// Absolute loss > $500 (i.e. unrealizedPnLBase < -500)
	threshold := decimal.NewFromInt(-500)
	if unrealizedPnLBase.GreaterThan(threshold) {
		return false
	}
	lossPct := unrealizedPnLBase.Div(costBasisBase).Mul(decimal.NewFromInt(100))
	return lossPct.LessThanOrEqual(decimal.NewFromInt(-10))
}

// CheckWashSaleRisk returns true if any BUY trade occurred within 30 days before or after evalDate.
func CheckWashSaleRisk(buyDates []time.Time, evalDate time.Time) bool {
	dEval := time.Date(evalDate.Year(), evalDate.Month(), evalDate.Day(), 0, 0, 0, 0, time.UTC)
	for _, b := range buyDates {
		dB := time.Date(b.Year(), b.Month(), b.Day(), 0, 0, 0, 0, time.UTC)
		diffDays := int(dB.Sub(dEval).Hours() / 24)
		if diffDays < 0 {
			diffDays = -diffDays
		}
		if diffDays <= 30 {
			return true
		}
	}
	return false
}

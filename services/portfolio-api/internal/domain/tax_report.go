package domain

import (
	"github.com/shopspring/decimal"
)

type TaxableIncomeSummary struct {
	DividendAndInterestIncome decimal.Decimal
	NetCapitalGains            decimal.Decimal
	FrankingCredits            decimal.Decimal
	TotalAssessableIncome      decimal.Decimal
	DividendInterestRatioPct  decimal.Decimal
	CapitalGainsRatioPct      decimal.Decimal
	FrankingCreditsRatioPct   decimal.Decimal
}

type CapitalGainsSummary struct {
	TotalCapitalGains   decimal.Decimal
	LossesApplied       decimal.Decimal
	CGTDiscountAmount   decimal.Decimal
	NetCapitalGain      decimal.Decimal
	CGTDiscountRatePct  decimal.Decimal
	GainDisposalsCount  int
	LossDisposalsCount  int
}

type TaxOffsetsSummary struct {
	FrankingCreditsOffset decimal.Decimal
	ForeignTaxOffsets     decimal.Decimal
	TotalAvailableOffsets decimal.Decimal
}

type TaxReportSummary struct {
	TaxableIncome   TaxableIncomeSummary
	CapitalGains    CapitalGainsSummary
	TaxOffsets      TaxOffsetsSummary
	FiscalYearLabel string
	BaseCurrency    string
}

type TaxReport struct {
	Summary         TaxReportSummary
	Disposals       []LotDisposal
	OpenLots        []TaxLot
	FromDate        string
	ToDate          string
	FiscalYearLabel string
	BaseCurrency    string
}

type RealizedGainsSummary struct {
	TotalProceedsBase         decimal.Decimal
	TotalCostBasisRelieved     decimal.Decimal
	GrossRealizedGainsBase     decimal.Decimal
	GrossRealizedLossesBase    decimal.Decimal
	NetRealizedPnLBase         decimal.Decimal
	ShortTermGainsBase         decimal.Decimal
	LongTermGainsBase          decimal.Decimal
	DiscountedTaxableGainBase  decimal.Decimal
	TotalDisposalCount         int
}

type RealizedGainsReport struct {
	Summary      RealizedGainsSummary
	Disposals    []LotDisposal
	FromDate     string
	ToDate       string
	BaseCurrency string
}

// CalculateCapitalGainsWaterfall reconciles gross gains, gross losses, and computes statutory
// net capital gains applying losses against short-term gains first before long-term gains (ATO ITAA 1997 s 102-5).
func CalculateCapitalGainsWaterfall(disposals []LotDisposal) CapitalGainsSummary {
	shortTermGains := decimal.Zero
	longTermGains := decimal.Zero
	totalLosses := decimal.Zero
	gainCount := 0
	lossCount := 0

	for _, d := range disposals {
		if d.RealizedPnLBase.IsPositive() {
			gainCount++
			if d.IsCGTDiscountEligible || d.TermClassification == TermLongTerm {
				longTermGains = longTermGains.Add(d.RealizedPnLBase)
			} else {
				shortTermGains = shortTermGains.Add(d.RealizedPnLBase)
			}
		} else if d.RealizedPnLBase.IsNegative() {
			lossCount++
			totalLosses = totalLosses.Add(d.RealizedPnLBase.Abs())
		}
	}

	totalCapitalGains := shortTermGains.Add(longTermGains)
	lossesApplied := decimal.Min(totalCapitalGains, totalLosses)

	// Optimal loss offsetting: apply losses first against short-term gains (taxed at 100%)
	lossesToST := decimal.Min(shortTermGains, lossesApplied)
	remainingST := shortTermGains.Sub(lossesToST)
	lossesRemaining := lossesApplied.Sub(lossesToST)

	lossesToLT := decimal.Min(longTermGains, lossesRemaining)
	remainingLT := longTermGains.Sub(lossesToLT)

	// Concessional 50% CGT discount on remaining long-term gains
	cgtDiscountRate := decimal.NewFromFloat(0.50)
	cgtDiscountAmount := remainingLT.Mul(cgtDiscountRate).Round(2)

	// Net capital gain
	netCapitalGain := remainingST.Add(remainingLT.Sub(cgtDiscountAmount))

	return CapitalGainsSummary{
		TotalCapitalGains:  totalCapitalGains.Round(2),
		LossesApplied:      lossesApplied.Round(2),
		CGTDiscountAmount:  cgtDiscountAmount,
		NetCapitalGain:     netCapitalGain.Round(2),
		CGTDiscountRatePct: decimal.NewFromInt(50),
		GainDisposalsCount: gainCount,
		LossDisposalsCount: lossCount,
	}
}

// CalculateTaxableIncomeSummary calculates total assessable income and proportional visual ratio percentages.
func CalculateTaxableIncomeSummary(dividendAndInterestIncome, netCapitalGains, frankingCredits decimal.Decimal) TaxableIncomeSummary {
	totalAssessable := dividendAndInterestIncome.Add(netCapitalGains).Add(frankingCredits).Round(2)

	divIntRatio := decimal.Zero
	cgtRatio := decimal.Zero
	frankingRatio := decimal.Zero

	if totalAssessable.IsPositive() {
		divIntRatio = dividendAndInterestIncome.Div(totalAssessable).Mul(decimal.NewFromInt(100)).Round(2)
		cgtRatio = netCapitalGains.Div(totalAssessable).Mul(decimal.NewFromInt(100)).Round(2)
		frankingRatio = frankingCredits.Div(totalAssessable).Mul(decimal.NewFromInt(100)).Round(2)
	}

	return TaxableIncomeSummary{
		DividendAndInterestIncome: dividendAndInterestIncome.Round(2),
		NetCapitalGains:           netCapitalGains.Round(2),
		FrankingCredits:           frankingCredits.Round(2),
		TotalAssessableIncome:     totalAssessable,
		DividendInterestRatioPct:  divIntRatio,
		CapitalGainsRatioPct:      cgtRatio,
		FrankingCreditsRatioPct:   frankingRatio,
	}
}

// CalculateRealizedGainsSummary calculates summary metrics for realized disposals.
func CalculateRealizedGainsSummary(disposals []LotDisposal) RealizedGainsSummary {
	cgt := CalculateCapitalGainsWaterfall(disposals)

	totalProceeds := decimal.Zero
	totalCostBasisRelieved := decimal.Zero
	shortTermGains := decimal.Zero
	longTermGains := decimal.Zero
	totalLosses := decimal.Zero

	for _, d := range disposals {
		totalProceeds = totalProceeds.Add(d.ProceedsBase)
		totalCostBasisRelieved = totalCostBasisRelieved.Add(d.CostBasisReleasedBase)

		if d.RealizedPnLBase.IsPositive() {
			if d.IsCGTDiscountEligible || d.TermClassification == TermLongTerm {
				longTermGains = longTermGains.Add(d.RealizedPnLBase)
			} else {
				shortTermGains = shortTermGains.Add(d.RealizedPnLBase)
			}
		} else if d.RealizedPnLBase.IsNegative() {
			totalLosses = totalLosses.Add(d.RealizedPnLBase.Abs())
		}
	}

	netRealizedPnL := totalProceeds.Sub(totalCostBasisRelieved)

	return RealizedGainsSummary{
		TotalProceedsBase:        totalProceeds.Round(2),
		TotalCostBasisRelieved:    totalCostBasisRelieved.Round(2),
		GrossRealizedGainsBase:    cgt.TotalCapitalGains,
		GrossRealizedLossesBase:   totalLosses.Round(2),
		NetRealizedPnLBase:        netRealizedPnL.Round(2),
		ShortTermGainsBase:        shortTermGains.Round(2),
		LongTermGainsBase:         longTermGains.Round(2),
		DiscountedTaxableGainBase: cgt.NetCapitalGain,
		TotalDisposalCount:        len(disposals),
	}
}

package domain

import (
	"errors"
	"math"
	"time"

	"github.com/shopspring/decimal"
)

// ExternalCashFlow represents a single external deposit or withdrawal event.
type ExternalCashFlow struct {
	TransactionID string
	Date          time.Time
	Type          string // "DEPOSIT" or "WITHDRAWAL"
	AmountBase    decimal.Decimal // Positive for deposit, negative for withdrawal
	LocalAmount   decimal.Decimal
	LocalCurrency string
}

// CashFlowWeight captures the time-weighted breakdown of an individual cash flow.
type CashFlowWeight struct {
	TransactionID      string
	Date               time.Time
	Type               string
	AmountBase         decimal.Decimal
	Weight             decimal.Decimal
	WeightedAmountBase decimal.Decimal
	LocalAmount        decimal.Decimal
	LocalCurrency      string
}

// ModifiedDietzResult encapsulates all calculated metrics for a Modified Dietz evaluation.
type ModifiedDietzResult struct {
	StartDate               time.Time
	EndDate                 time.Time
	ElapsedDays             int32
	StartValuation          decimal.Decimal
	EndValuation            decimal.Decimal
	NetExternalFlows        decimal.Decimal
	NetInvestmentGain       decimal.Decimal
	WeightedCapitalEmployed decimal.Decimal
	CumulativeReturnPercent decimal.Decimal
	AnnualizedReturnPercent *decimal.Decimal
	IsAnnualized            bool
	IsComputable            bool
	DiagnosticCode          string
	CashFlowWeights         []CashFlowWeight
}

// CalculateModifiedDietz computes the GIPS-compliant Modified Dietz return over a single period.
func CalculateModifiedDietz(
	vStart decimal.Decimal,
	vEnd decimal.Decimal,
	startDate time.Time,
	endDate time.Time,
	flows []ExternalCashFlow,
) (*ModifiedDietzResult, error) {
	if endDate.Before(startDate) {
		return nil, errors.New("endDate cannot be before startDate")
	}

	cdHours := endDate.Sub(startDate).Hours()
	cdDays := int32(math.Round(cdHours / 24.0))
	if cdDays <= 0 {
		return nil, errors.New("period length must be at least 1 day")
	}

	cdDec := decimal.NewFromInt(int64(cdDays))
	totalF := decimal.Zero
	sumWeightedFlows := decimal.Zero
	weights := make([]CashFlowWeight, 0, len(flows))

	for _, flow := range flows {
		dHours := flow.Date.Sub(startDate).Hours()
		dOffset := int32(math.Round(dHours / 24.0))
		if dOffset < 0 {
			dOffset = 0
		}
		if dOffset > cdDays {
			dOffset = cdDays
		}

		weight := cdDec.Sub(decimal.NewFromInt(int64(dOffset))).DivRound(cdDec, 8)
		weightedAmount := flow.AmountBase.Mul(weight).Round(2)

		totalF = totalF.Add(flow.AmountBase)
		sumWeightedFlows = sumWeightedFlows.Add(weightedAmount)

		weights = append(weights, CashFlowWeight{
			TransactionID:      flow.TransactionID,
			Date:               flow.Date,
			Type:               flow.Type,
			AmountBase:         flow.AmountBase,
			Weight:             weight,
			WeightedAmountBase: weightedAmount,
			LocalAmount:        flow.LocalAmount,
			LocalCurrency:      flow.LocalCurrency,
		})
	}

	weightedCapital := vStart.Add(sumWeightedFlows).Round(2)
	netGain := vEnd.Sub(vStart).Sub(totalF).Round(2)

	// Boundary Condition: Zero Capital Employed
	if weightedCapital.IsZero() {
		return &ModifiedDietzResult{
			StartDate:               startDate,
			EndDate:                 endDate,
			ElapsedDays:             cdDays,
			StartValuation:          vStart,
			EndValuation:            vEnd,
			NetExternalFlows:        totalF,
			NetInvestmentGain:       netGain,
			WeightedCapitalEmployed: weightedCapital,
			CumulativeReturnPercent: decimal.Zero,
			AnnualizedReturnPercent: nil,
			IsAnnualized:            false,
			IsComputable:            false,
			DiagnosticCode:          "ZERO_CAPITAL_EMPLOYED",
			CashFlowWeights:         weights,
		}, nil
	}

	// Boundary Condition: Negative Capital Employed (large net withdrawal)
	if weightedCapital.IsNegative() {
		return &ModifiedDietzResult{
			StartDate:               startDate,
			EndDate:                 endDate,
			ElapsedDays:             cdDays,
			StartValuation:          vStart,
			EndValuation:            vEnd,
			NetExternalFlows:        totalF,
			NetInvestmentGain:       netGain,
			WeightedCapitalEmployed: weightedCapital,
			CumulativeReturnPercent: decimal.Zero,
			AnnualizedReturnPercent: nil,
			IsAnnualized:            false,
			IsComputable:            false,
			DiagnosticCode:          "NEGATIVE_AVERAGE_CAPITAL",
			CashFlowWeights:         weights,
		}, nil
	}

	cumReturn := netGain.DivRound(weightedCapital, 8)
	cumReturnPercent := cumReturn.Mul(decimal.NewFromInt(100)).Round(2)

	// GIPS Standard 2.A.20: Sub-1-year prohibition (< 365 calendar days)
	if cdDays < 365 {
		return &ModifiedDietzResult{
			StartDate:               startDate,
			EndDate:                 endDate,
			ElapsedDays:             cdDays,
			StartValuation:          vStart,
			EndValuation:            vEnd,
			NetExternalFlows:        totalF,
			NetInvestmentGain:       netGain,
			WeightedCapitalEmployed: weightedCapital,
			CumulativeReturnPercent: cumReturnPercent,
			AnnualizedReturnPercent: nil,
			IsAnnualized:            false,
			IsComputable:            true,
			DiagnosticCode:          "OK",
			CashFlowWeights:         weights,
		}, nil
	}

	// Geometric compound annualization: ((1 + R)^(365.25 / CD) - 1) * 100
	factor := decimal.NewFromInt(1).Add(cumReturn)
	if factor.IsNegative() || factor.IsZero() {
		negHundred := decimal.NewFromInt(-100)
		return &ModifiedDietzResult{
			StartDate:               startDate,
			EndDate:                 endDate,
			ElapsedDays:             cdDays,
			StartValuation:          vStart,
			EndValuation:            vEnd,
			NetExternalFlows:        totalF,
			NetInvestmentGain:       netGain,
			WeightedCapitalEmployed: weightedCapital,
			CumulativeReturnPercent: cumReturnPercent,
			AnnualizedReturnPercent: &negHundred,
			IsAnnualized:            true,
			IsComputable:            true,
			DiagnosticCode:          "OK",
			CashFlowWeights:         weights,
		}, nil
	}

	fVal, _ := factor.Float64()
	exponent := 365.25 / float64(cdDays)
	annVal := math.Pow(fVal, exponent)
	annPercent := decimal.NewFromFloat(annVal).Sub(decimal.NewFromInt(1)).Mul(decimal.NewFromInt(100)).Round(2)

	return &ModifiedDietzResult{
		StartDate:               startDate,
		EndDate:                 endDate,
		ElapsedDays:             cdDays,
		StartValuation:          vStart,
		EndValuation:            vEnd,
		NetExternalFlows:        totalF,
		NetInvestmentGain:       netGain,
		WeightedCapitalEmployed: weightedCapital,
		CumulativeReturnPercent: cumReturnPercent,
		AnnualizedReturnPercent: &annPercent,
		IsAnnualized:            true,
		IsComputable:            true,
		DiagnosticCode:          "OK",
		CashFlowWeights:         weights,
	}, nil
}

// ChainModifiedDietzSubPeriods geometrically chains consecutive sub-period results across multi-month or multi-year windows.
func ChainModifiedDietzSubPeriods(
	subPeriods []*ModifiedDietzResult,
	startDate time.Time,
	endDate time.Time,
) (*ModifiedDietzResult, error) {
	if len(subPeriods) == 0 {
		return nil, errors.New("cannot chain empty sub-period list")
	}

	var activePeriods []*ModifiedDietzResult
	for _, p := range subPeriods {
		if p.StartValuation.IsZero() && p.EndValuation.IsZero() && p.NetExternalFlows.IsZero() {
			// Inactive sub-period (no capital, no flows, no returns) - skip
			continue
		}
		activePeriods = append(activePeriods, p)
	}

	if len(activePeriods) == 0 {
		return &ModifiedDietzResult{
			StartDate:               startDate,
			EndDate:                 endDate,
			ElapsedDays:             int32(math.Round(endDate.Sub(startDate).Hours() / 24.0)),
			StartValuation:          decimal.Zero,
			EndValuation:            decimal.Zero,
			NetExternalFlows:        decimal.Zero,
			NetInvestmentGain:       decimal.Zero,
			WeightedCapitalEmployed: decimal.Zero,
			CumulativeReturnPercent: decimal.Zero,
			AnnualizedReturnPercent: nil,
			IsAnnualized:            false,
			IsComputable:            false,
			DiagnosticCode:          "ZERO_CAPITAL_EMPLOYED",
			CashFlowWeights:         nil,
		}, nil
	}

	totalDays := int32(0)
	totalNetFlows := decimal.Zero
	totalNetGain := decimal.Zero
	allWeights := make([]CashFlowWeight, 0)
	chainedFactor := decimal.NewFromInt(1)
	isComputable := true
	diagnosticCode := "OK"

	for _, p := range activePeriods {
		totalDays += p.ElapsedDays
		totalNetFlows = totalNetFlows.Add(p.NetExternalFlows)
		totalNetGain = totalNetGain.Add(p.NetInvestmentGain)
		allWeights = append(allWeights, p.CashFlowWeights...)

		if !p.IsComputable {
			isComputable = false
			diagnosticCode = p.DiagnosticCode
		}

		if isComputable {
			periodReturn := p.CumulativeReturnPercent.DivRound(decimal.NewFromInt(100), 8)
			factor := decimal.NewFromInt(1).Add(periodReturn)
			chainedFactor = chainedFactor.Mul(factor)
		}
	}

	startVal := activePeriods[0].StartValuation
	endVal := activePeriods[len(activePeriods)-1].EndValuation
	startPeriodDate := activePeriods[0].StartDate
	endPeriodDate := activePeriods[len(activePeriods)-1].EndDate

	if !isComputable {
		return &ModifiedDietzResult{
			StartDate:               startPeriodDate,
			EndDate:                 endPeriodDate,
			ElapsedDays:             totalDays,
			StartValuation:          startVal,
			EndValuation:            endVal,
			NetExternalFlows:        totalNetFlows,
			NetInvestmentGain:       totalNetGain,
			WeightedCapitalEmployed: decimal.Zero,
			CumulativeReturnPercent: decimal.Zero,
			AnnualizedReturnPercent: nil,
			IsAnnualized:            false,
			IsComputable:            false,
			DiagnosticCode:          diagnosticCode,
			CashFlowWeights:         allWeights,
		}, nil
	}

	cumReturn := chainedFactor.Sub(decimal.NewFromInt(1))
	cumReturnPercent := cumReturn.Mul(decimal.NewFromInt(100)).Round(2)

	// Approximate consolidated weighted capital = NetGain / CumReturn (if CumReturn != 0)
	var consolidatedWeightedCapital decimal.Decimal
	if !cumReturn.IsZero() {
		consolidatedWeightedCapital = totalNetGain.DivRound(cumReturn, 2)
	} else {
		consolidatedWeightedCapital = startVal
	}

	if totalDays < 365 {
		return &ModifiedDietzResult{
			StartDate:               startPeriodDate,
			EndDate:                 endPeriodDate,
			ElapsedDays:             totalDays,
			StartValuation:          startVal,
			EndValuation:            endVal,
			NetExternalFlows:        totalNetFlows,
			NetInvestmentGain:       totalNetGain,
			WeightedCapitalEmployed: consolidatedWeightedCapital,
			CumulativeReturnPercent: cumReturnPercent,
			AnnualizedReturnPercent: nil,
			IsAnnualized:            false,
			IsComputable:            true,
			DiagnosticCode:          "OK",
			CashFlowWeights:         allWeights,
		}, nil
	}

	if chainedFactor.IsNegative() || chainedFactor.IsZero() {
		negHundred := decimal.NewFromInt(-100)
		return &ModifiedDietzResult{
			StartDate:               startPeriodDate,
			EndDate:                 endPeriodDate,
			ElapsedDays:             totalDays,
			StartValuation:          startVal,
			EndValuation:            endVal,
			NetExternalFlows:        totalNetFlows,
			NetInvestmentGain:       totalNetGain,
			WeightedCapitalEmployed: consolidatedWeightedCapital,
			CumulativeReturnPercent: cumReturnPercent,
			AnnualizedReturnPercent: &negHundred,
			IsAnnualized:            true,
			IsComputable:            true,
			DiagnosticCode:          "OK",
			CashFlowWeights:         allWeights,
		}, nil
	}

	fVal, _ := chainedFactor.Float64()
	exponent := 365.25 / float64(totalDays)
	annVal := math.Pow(fVal, exponent)
	annPercent := decimal.NewFromFloat(annVal).Sub(decimal.NewFromInt(1)).Mul(decimal.NewFromInt(100)).Round(2)

	return &ModifiedDietzResult{
		StartDate:               startPeriodDate,
		EndDate:                 endPeriodDate,
		ElapsedDays:             totalDays,
		StartValuation:          startVal,
		EndValuation:            endVal,
		NetExternalFlows:        totalNetFlows,
		NetInvestmentGain:       totalNetGain,
		WeightedCapitalEmployed: consolidatedWeightedCapital,
		CumulativeReturnPercent: cumReturnPercent,
		AnnualizedReturnPercent: &annPercent,
		IsAnnualized:            true,
		IsComputable:            true,
		DiagnosticCode:          "OK",
		CashFlowWeights:         allWeights,
	}, nil
}

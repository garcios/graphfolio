package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestCalculateCapitalGainsWaterfall_Basic(t *testing.T) {
	// TEST-TAX-03: Gross Gains $6,064.99 (Long Term), Gross Losses $350.17
	// Losses applied = $350.17, remaining long-term = $5,714.82, 50% discount = $2,857.41, Net CGT = $2,857.41
	disposals := []LotDisposal{
		{
			ID:                    uuid.New(),
			RealizedPnLBase:       decimal.NewFromFloat(6064.99),
			IsCGTDiscountEligible: true,
			TermClassification:    TermLongTerm,
			AcquiredDate:          time.Date(2023, 1, 10, 0, 0, 0, 0, time.UTC),
			DisposedDate:          time.Date(2024, 7, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:                    uuid.New(),
			RealizedPnLBase:       decimal.NewFromFloat(-350.17),
			IsCGTDiscountEligible: false,
			TermClassification:    TermShortTerm,
			AcquiredDate:          time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
			DisposedDate:          time.Date(2024, 8, 20, 0, 0, 0, 0, time.UTC),
		},
	}

	summary := CalculateCapitalGainsWaterfall(disposals)

	assert.True(t, decimal.NewFromFloat(6064.99).Equal(summary.TotalCapitalGains))
	assert.True(t, decimal.NewFromFloat(350.17).Equal(summary.LossesApplied))
	assert.True(t, decimal.NewFromFloat(2857.41).Equal(summary.CGTDiscountAmount))
	assert.True(t, decimal.NewFromFloat(2857.41).Equal(summary.NetCapitalGain))
	assert.Equal(t, 1, summary.GainDisposalsCount)
	assert.Equal(t, 1, summary.LossDisposalsCount)
}

func TestCalculateCapitalGainsWaterfall_LossSequencing(t *testing.T) {
	// TEST-TAX-04: Loss allocation sequencing with mixed short-term and long-term gains
	// Short-term gains: $1,000.00
	// Long-term gains: $2,000.00
	// Losses: $1,500.00
	// Under tax law (ATO ITAA 1997 s 102-5), losses offset Short-Term gains first!
	// Losses to ST: $1,000 (exhausting ST gains to $0)
	// Remaining losses to LT: $500 ($1,500 - $1,000 = $500)
	// Remaining LT gains: $2,000 - $500 = $1,500
	// 50% discount on remaining LT: $750.00
	// Net Capital Gain: $0 (ST) + ($1,500 - $750) = $750.00
	disposals := []LotDisposal{
		{
			ID:                    uuid.New(),
			RealizedPnLBase:       decimal.NewFromFloat(1000.00),
			IsCGTDiscountEligible: false,
			TermClassification:    TermShortTerm,
		},
		{
			ID:                    uuid.New(),
			RealizedPnLBase:       decimal.NewFromFloat(2000.00),
			IsCGTDiscountEligible: true,
			TermClassification:    TermLongTerm,
		},
		{
			ID:                    uuid.New(),
			RealizedPnLBase:       decimal.NewFromFloat(-1500.00),
			IsCGTDiscountEligible: false,
			TermClassification:    TermShortTerm,
		},
	}

	summary := CalculateCapitalGainsWaterfall(disposals)

	assert.True(t, decimal.NewFromFloat(3000.00).Equal(summary.TotalCapitalGains))
	assert.True(t, decimal.NewFromFloat(1500.00).Equal(summary.LossesApplied))
	assert.True(t, decimal.NewFromFloat(750.00).Equal(summary.CGTDiscountAmount))
	assert.True(t, decimal.NewFromFloat(750.00).Equal(summary.NetCapitalGain))
}

func TestCalculateTaxableIncomeSummary(t *testing.T) {
	// TEST-TAX-05: Assessable income exact cent summation
	// Div/Int ($2,966.80) + CGT ($2,857.41) + Franking ($300.33) = Total Assessable Income ($6,124.54)
	divInt := decimal.NewFromFloat(2966.80)
	cgt := decimal.NewFromFloat(2857.41)
	franking := decimal.NewFromFloat(300.33)

	res := CalculateTaxableIncomeSummary(divInt, cgt, franking)

	assert.True(t, decimal.NewFromFloat(6124.54).Equal(res.TotalAssessableIncome))

	// TEST-TAX-06: Visual progress bar ratio percentages
	// Div/Int 48.44%, CGT 46.66%, Franking 4.90%
	assert.True(t, decimal.NewFromFloat(48.44).Equal(res.DividendInterestRatioPct))
	assert.True(t, decimal.NewFromFloat(46.66).Equal(res.CapitalGainsRatioPct))
	assert.True(t, decimal.NewFromFloat(4.90).Equal(res.FrankingCreditsRatioPct))
}

package domain

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestCalculateHoldingDays(t *testing.T) {
	acq := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	eval1 := time.Date(2025, 3, 14, 0, 0, 0, 0, time.UTC)
	eval2 := time.Date(2025, 3, 16, 0, 0, 0, 0, time.UTC)

	// 2024 is a leap year (366 days in total year).
	// March 15 2024 to March 14 2025 is 364 days.
	assert.Equal(t, 364, CalculateHoldingDays(acq, eval1))
	// March 15 2024 to March 16 2025 is 366 days.
	assert.Equal(t, 366, CalculateHoldingDays(acq, eval2))

	// Before acq returns 0
	evalBefore := time.Date(2024, 3, 10, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, 0, CalculateHoldingDays(acq, evalBefore))
}

func TestIsCGTDiscountEligible(t *testing.T) {
	// TEST-TAX-01: Asset acquired 2024-03-15 evaluated on 2025-03-14
	acq := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	evalDay364 := time.Date(2025, 3, 14, 0, 0, 0, 0, time.UTC)
	assert.False(t, IsCGTDiscountEligible(acq, evalDay364))
	assert.Equal(t, TermShortTerm, ClassifyHoldingTerm(acq, evalDay364))

	// Exactly 1 year minus 1 day / exact anniversary (acquisition day excluded):
	evalAnniversary := time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC)
	assert.False(t, IsCGTDiscountEligible(acq, evalAnniversary))

	// TEST-TAX-02: Asset acquired 2024-03-15 evaluated on 2025-03-16
	evalDay366 := time.Date(2025, 3, 16, 0, 0, 0, 0, time.UTC)
	assert.True(t, IsCGTDiscountEligible(acq, evalDay366))
	assert.Equal(t, TermLongTerm, ClassifyHoldingTerm(acq, evalDay366))
}

func TestIsLossHarvestingCandidate(t *testing.T) {
	// TEST-TAX-08: Open lot with loss > $500 and <= -10% flagged
	costBasis := decimal.NewFromFloat(10000.00)
	loss600 := decimal.NewFromFloat(-600.00) // -6%, so loss > $500 but pct > -10% -> false
	assert.False(t, IsLossHarvestingCandidate(loss600, costBasis))

	loss1200 := decimal.NewFromFloat(-1200.00) // -12% and -$1200 -> true
	assert.True(t, IsLossHarvestingCandidate(loss1200, costBasis))

	loss400 := decimal.NewFromFloat(-400.00) // -40% on $1000 cost basis -> pct is -40%, but loss < $500 -> false
	costBasisSmall := decimal.NewFromFloat(1000.00)
	assert.False(t, IsLossHarvestingCandidate(loss400, costBasisSmall))

	gain := decimal.NewFromFloat(500.00)
	assert.False(t, IsLossHarvestingCandidate(gain, costBasis))
}

func TestCheckWashSaleRisk(t *testing.T) {
	// TEST-TAX-09: Buy trade executed within 30 days flags has_wash_sale_risk = true
	evalDate := time.Date(2025, 4, 15, 0, 0, 0, 0, time.UTC)

	// Buy within 20 days prior
	buyPrior := []time.Time{time.Date(2025, 3, 26, 0, 0, 0, 0, time.UTC)}
	assert.True(t, CheckWashSaleRisk(buyPrior, evalDate))

	// Buy within 10 days after
	buyAfter := []time.Time{time.Date(2025, 4, 25, 0, 0, 0, 0, time.UTC)}
	assert.True(t, CheckWashSaleRisk(buyAfter, evalDate))

	// Buy 45 days prior -> no risk
	buyOld := []time.Time{time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)}
	assert.False(t, CheckWashSaleRisk(buyOld, evalDate))
}

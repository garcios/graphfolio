package domain

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestCalculateModifiedDietz_ScenarioA_OneYearWithMidYearDeposit(t *testing.T) {
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC) // 364 or 365 days
	// Let's set endDate exactly 365 days after startDate:
	endDate = startDate.AddDate(0, 0, 365) // 365 days

	vStart := decimal.NewFromInt(100000)
	vEnd := decimal.NewFromInt(180000)

	depositDate := startDate.AddDate(0, 0, 181) // day 181
	flows := []ExternalCashFlow{
		{
			TransactionID: "tx-1",
			Date:          depositDate,
			Type:          "DEPOSIT",
			AmountBase:    decimal.NewFromInt(50000),
			LocalAmount:   decimal.NewFromInt(50000),
			LocalCurrency: "USD",
		},
	}

	result, err := CalculateModifiedDietz(vStart, vEnd, startDate, endDate, flows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ElapsedDays != 365 {
		t.Errorf("expected 365 elapsed days, got %d", result.ElapsedDays)
	}

	if !result.IsComputable {
		t.Errorf("expected result to be computable")
	}

	if !result.IsAnnualized {
		t.Errorf("expected result to be annualized for 365 days")
	}

	// Net Gain = 180,000 - 100,000 - 50,000 = 30,000
	expectedNetGain := decimal.NewFromInt(30000)
	if !result.NetInvestmentGain.Equal(expectedNetGain) {
		t.Errorf("expected net gain %s, got %s", expectedNetGain, result.NetInvestmentGain)
	}

	// Time weight: (365 - 181) / 365 = 184 / 365 = 0.50410959
	// Weighted Amount = 50,000 * 0.50410959 = 25,205.48
	// Weighted Capital = 100,000 + 25,205.48 = 125,205.48
	expectedWeightedCap := decimal.NewFromFloat(125205.48)
	if !result.WeightedCapitalEmployed.Equal(expectedWeightedCap) {
		t.Errorf("expected weighted capital %s, got %s", expectedWeightedCap, result.WeightedCapitalEmployed)
	}

	// Return = 30,000 / 125,205.48 = 23.96%
	expectedCumReturn := decimal.NewFromFloat(23.96)
	if !result.CumulativeReturnPercent.Equal(expectedCumReturn) {
		t.Errorf("expected cum return %s, got %s", expectedCumReturn, result.CumulativeReturnPercent)
	}

	// Annualized: ((1 + 0.2396061)^(365.25/365) - 1) * 100 = 23.98%
	if result.AnnualizedReturnPercent == nil {
		t.Fatalf("expected non-nil annualized return")
	}
	expectedAnnReturn := decimal.NewFromFloat(23.98)
	if !result.AnnualizedReturnPercent.Equal(expectedAnnReturn) {
		t.Errorf("expected ann return %s, got %s", expectedAnnReturn, *result.AnnualizedReturnPercent)
	}

	if len(result.CashFlowWeights) != 1 {
		t.Fatalf("expected 1 weight item, got %d", len(result.CashFlowWeights))
	}
	if result.CashFlowWeights[0].Type != "DEPOSIT" {
		t.Errorf("expected type DEPOSIT, got %s", result.CashFlowWeights[0].Type)
	}
}

func TestCalculateModifiedDietz_ScenarioB_SubOneYearGIPSProhibition(t *testing.T) {
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 0, 181) // 181 days (< 365)

	vStart := decimal.NewFromInt(50000)
	vEnd := decimal.NewFromInt(78000)

	depositDate := startDate.AddDate(0, 0, 73) // day 73
	flows := []ExternalCashFlow{
		{
			TransactionID: "tx-2",
			Date:          depositDate,
			Type:          "DEPOSIT",
			AmountBase:    decimal.NewFromInt(20000),
			LocalAmount:   decimal.NewFromInt(20000),
			LocalCurrency: "AUD",
		},
	}

	result, err := CalculateModifiedDietz(vStart, vEnd, startDate, endDate, flows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ElapsedDays != 181 {
		t.Errorf("expected 181 days, got %d", result.ElapsedDays)
	}

	// Must NOT be annualized per GIPS 2.A.20
	if result.IsAnnualized {
		t.Errorf("expected isAnnualized=false for sub-1-year period")
	}
	if result.AnnualizedReturnPercent != nil {
		t.Errorf("expected nil annualizedReturnPercent, got %v", *result.AnnualizedReturnPercent)
	}

	// Cum return = 12.92%
	expectedCumReturn := decimal.NewFromFloat(12.92)
	if !result.CumulativeReturnPercent.Equal(expectedCumReturn) {
		t.Errorf("expected cum return %s, got %s", expectedCumReturn, result.CumulativeReturnPercent)
	}
}

func TestCalculateModifiedDietz_BoundaryCases(t *testing.T) {
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 0, 365)

	t.Run("zero external flows reduces to simple holding period return", func(t *testing.T) {
		vStart := decimal.NewFromInt(100000)
		vEnd := decimal.NewFromInt(110000)

		res, err := CalculateModifiedDietz(vStart, vEnd, startDate, endDate, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !res.CumulativeReturnPercent.Equal(decimal.NewFromFloat(10.00)) {
			t.Errorf("expected 10.00%%, got %s", res.CumulativeReturnPercent)
		}
	})

	t.Run("day zero deposit has weight 1.0 and final day deposit has weight 0.0", func(t *testing.T) {
		vStart := decimal.NewFromInt(100000)
		vEnd := decimal.NewFromInt(160000)

		flows := []ExternalCashFlow{
			{
				TransactionID: "flow-start",
				Date:          startDate, // day 0
				Type:          "DEPOSIT",
				AmountBase:    decimal.NewFromInt(10000),
			},
			{
				TransactionID: "flow-end",
				Date:          endDate, // day 365
				Type:          "DEPOSIT",
				AmountBase:    decimal.NewFromInt(20000),
			},
		}

		res, err := CalculateModifiedDietz(vStart, vEnd, startDate, endDate, flows)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(res.CashFlowWeights) != 2 {
			t.Fatalf("expected 2 weights, got %d", len(res.CashFlowWeights))
		}

		// Flow on day 0: weight should be 1.0
		if !res.CashFlowWeights[0].Weight.Equal(decimal.NewFromInt(1)) {
			t.Errorf("expected weight 1.0 on day 0, got %s", res.CashFlowWeights[0].Weight)
		}
		// Flow on day 365: weight should be 0.0
		if !res.CashFlowWeights[1].Weight.Equal(decimal.Zero) {
			t.Errorf("expected weight 0.0 on day 365, got %s", res.CashFlowWeights[1].Weight)
		}

		// Weighted capital = 100,000 + 10,000*1 + 20,000*0 = 110,000
		if !res.WeightedCapitalEmployed.Equal(decimal.NewFromInt(110000)) {
			t.Errorf("expected 110,000 weighted capital, got %s", res.WeightedCapitalEmployed)
		}
	})

	t.Run("zero weighted capital employed", func(t *testing.T) {
		vStart := decimal.Zero
		vEnd := decimal.Zero

		res, err := CalculateModifiedDietz(vStart, vEnd, startDate, endDate, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.IsComputable {
			t.Errorf("expected isComputable=false for zero weighted capital")
		}
		if res.DiagnosticCode != "ZERO_CAPITAL_EMPLOYED" {
			t.Errorf("expected ZERO_CAPITAL_EMPLOYED, got %s", res.DiagnosticCode)
		}
	})

	t.Run("negative weighted capital employed", func(t *testing.T) {
		vStart := decimal.NewFromInt(10000)
		vEnd := decimal.NewFromInt(5000)
		// Large withdrawal early in period leaving negative average capital
		flows := []ExternalCashFlow{
			{
				TransactionID: "large-withdrawal",
				Date:          startDate.AddDate(0, 0, 1),
				Type:          "WITHDRAWAL",
				AmountBase:    decimal.NewFromInt(-20000), // withdrawal
			},
		}

		res, err := CalculateModifiedDietz(vStart, vEnd, startDate, endDate, flows)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.IsComputable {
			t.Errorf("expected isComputable=false for negative average capital")
		}
		if res.DiagnosticCode != "NEGATIVE_AVERAGE_CAPITAL" {
			t.Errorf("expected NEGATIVE_AVERAGE_CAPITAL, got %s", res.DiagnosticCode)
		}
	})
}

func TestChainModifiedDietzSubPeriods(t *testing.T) {
	d1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	d2 := d1.AddDate(0, 0, 365)
	d3 := d2.AddDate(0, 0, 365) // 730 days total

	sub1, err := CalculateModifiedDietz(decimal.NewFromInt(100000), decimal.NewFromInt(110000), d1, d2, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sub2, err := CalculateModifiedDietz(decimal.NewFromInt(110000), decimal.NewFromInt(121000), d2, d3, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	chained, err := ChainModifiedDietzSubPeriods([]*ModifiedDietzResult{sub1, sub2}, d1, d3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if chained.ElapsedDays != 730 {
		t.Errorf("expected 730 days, got %d", chained.ElapsedDays)
	}
	if !chained.IsAnnualized {
		t.Errorf("expected isAnnualized=true for 730 days")
	}

	// Sub1: +10%, Sub2: +10% -> Chained: (1.10 * 1.10) - 1 = +21.00%
	if !chained.CumulativeReturnPercent.Equal(decimal.NewFromFloat(21.00)) {
		t.Errorf("expected 21.00%% cumulative return, got %s", chained.CumulativeReturnPercent)
	}

	// Annualized: ((1.21)^(365.25/730) - 1)*100 ≈ 10.01%
	if chained.AnnualizedReturnPercent == nil {
		t.Fatalf("expected non-nil annualized return")
	}
	expectedAnn := decimal.NewFromFloat(10.01)
	if !chained.AnnualizedReturnPercent.Equal(expectedAnn) {
		t.Errorf("expected %s ann return, got %s", expectedAnn, *chained.AnnualizedReturnPercent)
	}
}

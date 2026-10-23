package domain_test

import (
	"testing"
	"time"

	"portfolio-api/internal/domain"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestAnnualizeDPS(t *testing.T) {
	dps := decimal.RequireFromString("0.50")

	tests := []struct {
		freq     domain.DividendFrequency
		expected string
	}{
		{domain.DividendFreqMonthly, "6"},
		{domain.DividendFreqQuarterly, "2"},
		{domain.DividendFreqSemiAnnual, "1"},
		{domain.DividendFreqAnnual, "0.5"},
		{domain.DividendFreqIrregular, "2"}, // defaults to quarterly 4x
	}

	for _, tt := range tests {
		res := domain.AnnualizeDPS(dps, tt.freq)
		if !res.Equal(decimal.RequireFromString(tt.expected)) {
			t.Errorf("AnnualizeDPS(%s, %s): expected %s, got %s", dps, tt.freq, tt.expected, res)
		}
	}
}

func TestCalculateForwardYieldAnalytics(t *testing.T) {
	inst1 := uuid.New()
	inst2 := uuid.New()
	instClosed := uuid.New()

	now, _ := time.Parse("2006-01-02", "2026-10-15")

	holdings := []domain.HoldingWithPrice{
		{
			InstrumentID:       inst1,
			Ticker:             "AAPL",
			Name:               "Apple Inc.",
			InstrumentCurrency: "USD",
			Quantity:           decimal.RequireFromString("100"),
			CostBasis:          decimal.RequireFromString("12000"), // avg buy: 120 USD
			CostBasisBase:      decimal.RequireFromString("18000"), // 12000 * 1.5 AUD
			LatestPrice:        decimal.RequireFromString("150"),   // market value: 15000 USD
		},
		{
			InstrumentID:       inst2,
			Ticker:             "BHP.AX",
			Name:               "BHP Group",
			InstrumentCurrency: "AUD",
			Quantity:           decimal.RequireFromString("200"),
			CostBasis:          decimal.Zero,                       // Free spin-off / zero cost basis
			CostBasisBase:      decimal.Zero,
			LatestPrice:        decimal.RequireFromString("40"),    // market value: 8000 AUD
		},
		{
			InstrumentID:       instClosed,
			Ticker:             "MSFT",
			Name:               "Microsoft",
			InstrumentCurrency: "USD",
			Quantity:           decimal.Zero,                       // Closed position
			CostBasis:          decimal.Zero,
			CostBasisBase:      decimal.Zero,
			LatestPrice:        decimal.RequireFromString("300"),
		},
	}

	payDate1, _ := time.Parse("2006-01-02", "2026-11-15")
	exDate1, _ := time.Parse("2006-01-02", "2026-10-20")

	payDate2, _ := time.Parse("2006-01-02", "2026-12-01")
	exDate2, _ := time.Parse("2006-01-02", "2026-11-05")

	latestSchedules := map[uuid.UUID]domain.DividendSchedule{
		inst1: {
			ID:             uuid.New(),
			InstrumentID:   inst1,
			AmountPerShare: decimal.RequireFromString("0.75"), // 0.75 * 4 = 3.00 annual DPS
			CurrencyCode:   "USD",
			Frequency:      domain.DividendFreqQuarterly,
			ExDate:         exDate1,
			PaymentDate:    payDate1,
		},
		inst2: {
			ID:             uuid.New(),
			InstrumentID:   inst2,
			AmountPerShare: decimal.RequireFromString("1.50"), // 1.50 * 2 = 3.00 annual DPS
			CurrencyCode:   "AUD",
			Frequency:      domain.DividendFreqSemiAnnual,
			ExDate:         exDate2,
			PaymentDate:    payDate2,
		},
	}

	forwardSchedules := []domain.DividendSchedule{
		latestSchedules[inst1],
		latestSchedules[inst2],
	}

	// Pending schedule: exDate in the past, payDate in the future
	pendingSchedID := uuid.New()
	pendingPastEx, _ := time.Parse("2006-01-02", "2026-10-05")
	pendingFuturePay, _ := time.Parse("2006-01-02", "2026-10-28")
	pendingSchedules := []domain.DividendSchedule{
		{
			ID:                pendingSchedID,
			InstrumentID:      inst1,
			ExDate:            pendingPastEx,
			PaymentDate:       pendingFuturePay,
			AmountPerShare:    decimal.RequireFromString("0.75"),
			CurrencyCode:      "USD",
			WithholdingTaxPct: decimal.RequireFromString("15.00"), // 15% US treaty tax
		},
	}

	entitledMap := map[uuid.UUID]decimal.Decimal{
		pendingSchedID: decimal.RequireFromString("100"), // 100 shares entitled
	}

	fxRates := map[string]decimal.Decimal{
		"USD": decimal.RequireFromString("1.50"), // 1 USD = 1.50 AUD
		"AUD": decimal.NewFromInt(1),
	}

	analytics := domain.CalculateForwardYieldAnalytics(
		holdings,
		latestSchedules,
		forwardSchedules,
		nil,
		pendingSchedules,
		entitledMap,
		fxRates,
		"AUD",
		now,
	)

	if analytics == nil {
		t.Fatalf("expected non-nil analytics")
	}

	// AAPL annual cash = 100 * (0.75 * 4) * 1.50 = 450 AUD
	// BHP annual cash  = 200 * (1.50 * 2) * 1.00 = 600 AUD
	// Total forward annual = 1050 AUD
	expectedForwardAnnual := decimal.RequireFromString("1050")
	if !analytics.AnnualForwardIncome.Amount.Equal(expectedForwardAnnual) {
		t.Errorf("expected annual forward income %s, got %s", expectedForwardAnnual, analytics.AnnualForwardIncome.Amount)
	}

	// Total market value base: AAPL: 100 * 150 * 1.5 = 22500 AUD; BHP: 200 * 40 * 1.0 = 8000 AUD; Total = 30500 AUD
	// Portfolio Yield = 1050 / 30500 * 100 = ~3.442623%
	if !analytics.PortfolioCurrentYield.IsPositive() {
		t.Errorf("expected positive portfolio current yield, got %s", analytics.PortfolioCurrentYield)
	}

	// Check YOC zero-division safety on BHP:
	var bhpHolding *domain.HoldingYieldItem
	var aaplHolding *domain.HoldingYieldItem
	for i := range analytics.HoldingsYield {
		if analytics.HoldingsYield[i].Symbol == "BHP.AX" {
			bhpHolding = &analytics.HoldingsYield[i]
		}
		if analytics.HoldingsYield[i].Symbol == "AAPL" {
			aaplHolding = &analytics.HoldingsYield[i]
		}
	}

	if bhpHolding == nil {
		t.Fatalf("expected BHP.AX holding in yield items")
	}
	if bhpHolding.YieldOnCostPct != nil {
		t.Errorf("expected nil YOC for zero-cost-basis BHP, got %v", *bhpHolding.YieldOnCostPct)
	}

	if aaplHolding == nil {
		t.Fatalf("expected AAPL holding in yield items")
	}
	if aaplHolding.YieldOnCostPct == nil {
		t.Fatalf("expected non-nil YOC for AAPL")
	}
	// AAPL annual DPS = 3.00, avg buy price = 120. YOC = 3.00 / 120 * 100 = 2.5%
	expectedYOC := decimal.RequireFromString("2.5")
	if !aaplHolding.YieldOnCostPct.Equal(expectedYOC) {
		t.Errorf("expected AAPL YOC %s, got %s", expectedYOC, *aaplHolding.YieldOnCostPct)
	}

	// Check Pending Cash Inflow:
	// 100 shares * 0.75 * (1 - 0.15) * 1.50 = 100 * 0.6375 * 1.50 = 95.625 AUD
	expectedPending := decimal.RequireFromString("95.625")
	if !analytics.PendingCashInflow.Amount.Equal(expectedPending) {
		t.Errorf("expected pending cash inflow %s, got %s", expectedPending, analytics.PendingCashInflow.Amount)
	}

	// Check 12 monthly projections exist
	if len(analytics.MonthlyProjections) != 12 {
		t.Errorf("expected 12 monthly projections, got %d", len(analytics.MonthlyProjections))
	}
}

func TestSimulateDRIPCompounding(t *testing.T) {
	startVal := decimal.RequireFromString("100000")
	startDiv := decimal.RequireFromString("4000") // 4% starting yield

	input := domain.DRIPSimulationInput{
		HorizonYears:           10,
		DividendGrowthRatePct:  decimal.RequireFromString("5.0"),
		CapitalAppreciationPct: decimal.RequireFromString("7.0"),
	}

	res := domain.SimulateDRIPCompounding(startVal, startDiv, input, "USD")
	if res == nil {
		t.Fatalf("expected non-nil DRIP result")
	}

	if len(res.Points) != 10 {
		t.Errorf("expected 10 yearly points for 10-year horizon, got %d", len(res.Points))
	}

	// DRIP terminal value must be strictly higher than cash payout strategy
	if res.TerminalDRIPValue.LessThanOrEqual(res.TerminalCashValue) {
		t.Errorf("expected terminal DRIP value (%s) to exceed cash value (%s)", res.TerminalDRIPValue, res.TerminalCashValue)
	}

	if !res.DRIPBenefitAmount.IsPositive() {
		t.Errorf("expected positive DRIP benefit amount, got %s", res.DRIPBenefitAmount)
	}

	diff := res.TerminalDRIPValue.Sub(res.TerminalCashValue)
	if !res.DRIPBenefitAmount.Equal(diff) {
		t.Errorf("expected DRIP benefit amount to equal difference (%s), got %s", diff, res.DRIPBenefitAmount)
	}
}

func TestCalculateForwardYieldAnalytics_UsesHoldingFXRateToBase(t *testing.T) {
	now, _ := time.Parse("2006-01-02", "2026-10-23")
	instID := uuid.New()

	holdings := []domain.HoldingWithPrice{
		{
			InstrumentID:       instID,
			Ticker:             "SAP.DE",
			Name:               "SAP SE",
			InstrumentCurrency: "EUR",
			Quantity:           decimal.RequireFromString("100"),
			LatestPrice:        decimal.RequireFromString("50.00"), // 5000 EUR
			FXRateToBase:       decimal.RequireFromString("1.10"),  // 1 EUR = 1.10 USD -> 5500 USD
		},
	}

	sched := domain.DividendSchedule{
		ID:             uuid.New(),
		InstrumentID:   instID,
		AmountPerShare: decimal.RequireFromString("2.00"),
		CurrencyCode:   "EUR",
		Frequency:      domain.DividendFreqAnnual,
		ExDate:         now.AddDate(0, 1, 0),
		PaymentDate:    now.AddDate(0, 1, 15),
	}

	latestSchedules := map[uuid.UUID]domain.DividendSchedule{
		instID: sched,
	}

	// fxRates map does NOT contain EUR
	emptyFXRates := map[string]decimal.Decimal{}

	analytics := domain.CalculateForwardYieldAnalytics(
		holdings,
		latestSchedules,
		[]domain.DividendSchedule{sched},
		nil,
		nil,
		nil,
		emptyFXRates,
		"USD",
		now,
	)

	if analytics == nil {
		t.Fatalf("expected non-nil analytics")
	}

	if len(analytics.HoldingsYield) != 1 {
		t.Fatalf("expected 1 holding yield item, got %d", len(analytics.HoldingsYield))
	}

	item := analytics.HoldingsYield[0]
	// Market value should be 100 * 50 * 1.10 = 5500 USD
	expectedMarketVal := decimal.RequireFromString("5500.00")
	if !item.CurrentMarketValue.Amount.Equal(expectedMarketVal) {
		t.Errorf("expected CurrentMarketValue %s USD, got %s", expectedMarketVal, item.CurrentMarketValue.Amount)
	}

	// Projected annual cash: 100 * 2.00 * 1.10 = 220 USD
	expectedDivBase := decimal.RequireFromString("220.00")
	if !item.ProjectedAnnualCash.Amount.Equal(expectedDivBase) {
		t.Errorf("expected ProjectedAnnualCash %s USD, got %s", expectedDivBase, item.ProjectedAnnualCash.Amount)
	}
}

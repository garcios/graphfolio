package domain_test

import (
	"testing"
	"time"

	"portfolio-api/internal/domain"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestCalculateHoldingMarketValue(t *testing.T) {
	t.Run("calculates holding value with 1:1 FX rate", func(t *testing.T) {
		qty := decimal.RequireFromString("150.5")
		price := decimal.RequireFromString("200.25")
		fx := decimal.NewFromInt(1)

		val := domain.CalculateHoldingMarketValue(qty, price, fx)
		expected := decimal.RequireFromString("30137.625")

		if !val.Equal(expected) {
			t.Errorf("got holding market value %s, want %s", val, expected)
		}
	})

	t.Run("calculates holding value with non-USD FX conversion", func(t *testing.T) {
		qty := decimal.RequireFromString("100")
		price := decimal.RequireFromString("50")  // 5000 EUR
		fx := decimal.RequireFromString("1.0850") // EUR -> USD

		val := domain.CalculateHoldingMarketValue(qty, price, fx)
		expected := decimal.RequireFromString("5425")

		if !val.Equal(expected) {
			t.Errorf("got holding market value %s, want %s", val, expected)
		}
	})

	t.Run("defaults zero FX rate to 1", func(t *testing.T) {
		qty := decimal.RequireFromString("10")
		price := decimal.RequireFromString("25.5")
		fx := decimal.Zero

		val := domain.CalculateHoldingMarketValue(qty, price, fx)
		expected := decimal.RequireFromString("255")

		if !val.Equal(expected) {
			t.Errorf("got holding market value %s, want %s", val, expected)
		}
	})
}

func TestCalculateCashValue(t *testing.T) {
	t.Run("calculates cash value with direct FX rate", func(t *testing.T) {
		bal := decimal.RequireFromString("10000")
		fx := decimal.RequireFromString("1.25")

		val := domain.CalculateCashValue(bal, fx)
		expected := decimal.RequireFromString("12500")

		if !val.Equal(expected) {
			t.Errorf("got cash value %s, want %s", val, expected)
		}
	})

	t.Run("defaults zero FX rate to 1", func(t *testing.T) {
		bal := decimal.RequireFromString("5000")
		fx := decimal.Zero

		val := domain.CalculateCashValue(bal, fx)
		expected := decimal.RequireFromString("5000")

		if !val.Equal(expected) {
			t.Errorf("got cash value %s, want %s", val, expected)
		}
	})
}

func TestCalculateTotalValue(t *testing.T) {
	marketVal := decimal.RequireFromString("75250.50")
	cashVal := decimal.RequireFromString("12500.25")

	total := domain.CalculateTotalValue(marketVal, cashVal)
	expected := decimal.RequireFromString("87750.75")

	if !total.Equal(expected) {
		t.Errorf("got total value %s, want %s", total, expected)
	}
}

func TestCalculateNetCashFlow(t *testing.T) {
	t.Run("deposits exceed withdrawals", func(t *testing.T) {
		deposits := decimal.RequireFromString("15000")
		withdrawals := decimal.RequireFromString("5000")

		netFlow := domain.CalculateNetCashFlow(deposits, withdrawals)
		expected := decimal.RequireFromString("10000")

		if !netFlow.Equal(expected) {
			t.Errorf("got net flow %s, want %s", netFlow, expected)
		}
	})

	t.Run("withdrawals exceed deposits", func(t *testing.T) {
		deposits := decimal.RequireFromString("2000")
		withdrawals := decimal.RequireFromString("6000")

		netFlow := domain.CalculateNetCashFlow(deposits, withdrawals)
		expected := decimal.RequireFromString("-4000")

		if !netFlow.Equal(expected) {
			t.Errorf("got net flow %s, want %s", netFlow, expected)
		}
	})
}

func TestCalculateDailyReturn(t *testing.T) {
	t.Run("normal trading day with 5 percent gain and zero flow", func(t *testing.T) {
		prevVal := decimal.RequireFromString("10000")
		currentVal := decimal.RequireFromString("10500")
		netFlow := decimal.Zero

		ret := domain.CalculateDailyReturn(currentVal, prevVal, netFlow)
		if ret == nil {
			t.Fatalf("expected non-nil daily return")
		}

		expected := decimal.RequireFromString("0.050000000000")
		if !ret.Equal(expected) {
			t.Errorf("got daily return %s, want %s", ret, expected)
		}
	})

	t.Run("normal trading day with 3 percent loss and zero flow", func(t *testing.T) {
		prevVal := decimal.RequireFromString("10000")
		currentVal := decimal.RequireFromString("9700")
		netFlow := decimal.Zero

		ret := domain.CalculateDailyReturn(currentVal, prevVal, netFlow)
		if ret == nil {
			t.Fatalf("expected non-nil daily return")
		}

		expected := decimal.RequireFromString("-0.030000000000")
		if !ret.Equal(expected) {
			t.Errorf("got daily return %s, want %s", ret, expected)
		}
	})

	t.Run("deposit of 5000 stripped out so return is isolated", func(t *testing.T) {
		// Started at 10,000, market gained 5% (to 10,500), investor added 5,000 -> current = 15,500
		prevVal := decimal.RequireFromString("10000")
		currentVal := decimal.RequireFromString("15500")
		netFlow := decimal.RequireFromString("5000")

		ret := domain.CalculateDailyReturn(currentVal, prevVal, netFlow)
		if ret == nil {
			t.Fatalf("expected non-nil daily return")
		}

		expected := decimal.RequireFromString("0.050000000000")
		if !ret.Equal(expected) {
			t.Errorf("got daily return %s, want %s", ret, expected)
		}
	})

	t.Run("withdrawal of 2000 stripped out so return is isolated", func(t *testing.T) {
		// Started at 10,000, market gained 2% (to 10,200), investor withdrew 2,000 -> current = 8,200
		prevVal := decimal.RequireFromString("10000")
		currentVal := decimal.RequireFromString("8200")
		netFlow := decimal.RequireFromString("-2000")

		ret := domain.CalculateDailyReturn(currentVal, prevVal, netFlow)
		if ret == nil {
			t.Fatalf("expected non-nil daily return")
		}

		expected := decimal.RequireFromString("0.020000000000")
		if !ret.Equal(expected) {
			t.Errorf("got daily return %s, want %s", ret, expected)
		}
	})

	t.Run("zero-start: initial funding deposit produces 0 return", func(t *testing.T) {
		prevVal := decimal.Zero
		currentVal := decimal.RequireFromString("10000")
		netFlow := decimal.RequireFromString("10000")

		ret := domain.CalculateDailyReturn(currentVal, prevVal, netFlow)
		if ret == nil {
			t.Fatalf("expected non-nil return for zero-start deposit")
		}
		if !ret.IsZero() {
			t.Errorf("got daily return %s, want 0", ret)
		}
	})

	t.Run("zero-start: empty portfolio with 0 prior, current, and flow", func(t *testing.T) {
		prevVal := decimal.Zero
		currentVal := decimal.Zero
		netFlow := decimal.Zero

		ret := domain.CalculateDailyReturn(currentVal, prevVal, netFlow)
		if ret == nil {
			t.Fatalf("expected non-nil return for empty portfolio")
		}
		if !ret.IsZero() {
			t.Errorf("got daily return %s, want 0", ret)
		}
	})

	t.Run("negative previous equity returns nil (margin constraint)", func(t *testing.T) {
		prevVal := decimal.RequireFromString("-500")
		currentVal := decimal.RequireFromString("100")
		netFlow := decimal.Zero

		ret := domain.CalculateDailyReturn(currentVal, prevVal, netFlow)
		if ret != nil {
			t.Errorf("expected nil daily return for negative previous equity, got %s", ret)
		}
	})

	t.Run("full liquidation without market movement produces 0 return", func(t *testing.T) {
		// Started at 10,000, withdrew all 10,000 (netFlow = -10,000), currentVal = 0
		prevVal := decimal.RequireFromString("10000")
		currentVal := decimal.Zero
		netFlow := decimal.RequireFromString("-10000")

		ret := domain.CalculateDailyReturn(currentVal, prevVal, netFlow)
		if ret == nil {
			t.Fatalf("expected non-nil daily return")
		}
		if !ret.IsZero() {
			t.Errorf("got daily return %s, want 0", ret)
		}
	})

	t.Run("full liquidation with 10 percent loss before liquidation", func(t *testing.T) {
		// Started at 10,000, market fell to 9,000, withdrew 9,000 (netFlow = -9,000), currentVal = 0
		prevVal := decimal.RequireFromString("10000")
		currentVal := decimal.Zero
		netFlow := decimal.RequireFromString("-9000")

		ret := domain.CalculateDailyReturn(currentVal, prevVal, netFlow)
		if ret == nil {
			t.Fatalf("expected non-nil daily return")
		}

		expected := decimal.RequireFromString("-0.100000000000")
		if !ret.Equal(expected) {
			t.Errorf("got daily return %s, want %s", ret, expected)
		}
	})
}

func TestChainTWR(t *testing.T) {
	t.Run("inception baseline initialized when previous TWR is zero", func(t *testing.T) {
		ret := decimal.Zero
		twr := domain.ChainTWR(decimal.Zero, &ret)
		expected := decimal.RequireFromString("1.000000000000")

		if !twr.Equal(expected) {
			t.Errorf("got chained TWR %s, want %s", twr, expected)
		}
	})

	t.Run("inception baseline initialized when previous TWR is negative", func(t *testing.T) {
		dailyRet := decimal.RequireFromString("0.05")
		twr := domain.ChainTWR(decimal.RequireFromString("-1.0"), &dailyRet)
		expected := decimal.RequireFromString("1.050000000000")

		if !twr.Equal(expected) {
			t.Errorf("got chained TWR %s, want %s", twr, expected)
		}
	})

	t.Run("nil daily return preserves previous TWR to 12 decimals", func(t *testing.T) {
		prevTWR := decimal.RequireFromString("1.125000000000")
		twr := domain.ChainTWR(prevTWR, nil)

		if !twr.Equal(prevTWR) {
			t.Errorf("got chained TWR %s, want %s", twr, prevTWR)
		}
	})

	t.Run("multi-day cumulative compounding sequence", func(t *testing.T) {
		// Day 0: Inception baseline = 1.000000000000
		twr0 := domain.ChainTWR(decimal.Zero, nil)

		// Day 1: +5% gain -> 1.050000000000
		r1 := decimal.RequireFromString("0.050000000000")
		twr1 := domain.ChainTWR(twr0, &r1)
		expected1 := decimal.RequireFromString("1.050000000000")
		if !twr1.Equal(expected1) {
			t.Errorf("day 1: got %s, want %s", twr1, expected1)
		}

		// Day 2: -2% loss -> 1.05 * 0.98 = 1.029000000000
		r2 := decimal.RequireFromString("-0.020000000000")
		twr2 := domain.ChainTWR(twr1, &r2)
		expected2 := decimal.RequireFromString("1.029000000000")
		if !twr2.Equal(expected2) {
			t.Errorf("day 2: got %s, want %s", twr2, expected2)
		}

		// Day 3: Flat (0%) -> 1.029000000000
		r3 := decimal.Zero
		twr3 := domain.ChainTWR(twr2, &r3)
		if !twr3.Equal(expected2) {
			t.Errorf("day 3: got %s, want %s", twr3, expected2)
		}

		// Day 4: +10% gain -> 1.029 * 1.10 = 1.131900000000
		r4 := decimal.RequireFromString("0.100000000000")
		twr4 := domain.ChainTWR(twr3, &r4)
		expected4 := decimal.RequireFromString("1.131900000000")
		if !twr4.Equal(expected4) {
			t.Errorf("day 4: got %s, want %s", twr4, expected4)
		}
	})
}

func TestCalculateAnnualizedReturn(t *testing.T) {
	t.Run("periods under 365 days return non-annualized period return per GIPS 2.A.20", func(t *testing.T) {
		twrIndex := decimal.RequireFromString("1.05") // 5% period return
		days := 90

		ret, isAnnualized, err := domain.CalculateAnnualizedReturn(twrIndex, days)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if isAnnualized {
			t.Errorf("expected isAnnualized = false for period < 365 days")
		}

		expected := decimal.RequireFromString("5.00")
		if !ret.Equal(expected) {
			t.Errorf("got period return %s, want %s", ret, expected)
		}
	})

	t.Run("exactly 1 year (365.25 days) at 10 percent return", func(t *testing.T) {
		twrIndex := decimal.RequireFromString("1.10")
		days := 365

		ret, isAnnualized, err := domain.CalculateAnnualizedReturn(twrIndex, days)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !isAnnualized {
			t.Errorf("expected isAnnualized = true for days >= 365")
		}

		// With 365 days: (1.10 ^ (365.25 / 365) - 1) * 100 = 10.01%
		expected := decimal.RequireFromString("10.01")
		if !ret.Equal(expected) {
			t.Errorf("got annualized return %s, want %s", ret, expected)
		}
	})

	t.Run("2 years (730.5 days) at 21 percent cumulative return", func(t *testing.T) {
		// 1.10 ^ 2 = 1.21 cumulative return -> 10.00% annualized
		twrIndex := decimal.RequireFromString("1.21")
		days := 731

		ret, isAnnualized, err := domain.CalculateAnnualizedReturn(twrIndex, days)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !isAnnualized {
			t.Errorf("expected isAnnualized = true")
		}

		expected := decimal.RequireFromString("9.99")
		if !ret.Equal(expected) {
			t.Errorf("got annualized return %s, want %s", ret, expected)
		}
	})

	t.Run("non-positive twrIndex returns error", func(t *testing.T) {
		_, _, err := domain.CalculateAnnualizedReturn(decimal.Zero, 400)
		if err == nil {
			t.Errorf("expected error for non-positive TWR index")
		}
	})

	t.Run("non-positive days returns error", func(t *testing.T) {
		_, _, err := domain.CalculateAnnualizedReturn(decimal.RequireFromString("1.05"), 0)
		if err == nil {
			t.Errorf("expected error for non-positive days")
		}
	})
}

func TestPortfolioValuationSnapshot(t *testing.T) {
	portfolioID := uuid.New()
	asOfDate := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	dailyRet := decimal.RequireFromString("0.025000000000")

	snap := domain.PortfolioValuationSnapshot{
		PortfolioID:     portfolioID,
		ValuationDate:   asOfDate,
		MarketValueBase: decimal.RequireFromString("80000.50"),
		CashValueBase:   decimal.RequireFromString("20000.25"),
		NetFlowBase:     decimal.RequireFromString("1000.00"),
		DailyReturn:     &dailyRet,
		TWRIndex:        decimal.RequireFromString("1.025000000000"),
	}

	t.Run("TotalValue calculates sum of market and cash value", func(t *testing.T) {
		expected := decimal.RequireFromString("100000.75")
		if !snap.TotalValue().Equal(expected) {
			t.Errorf("got total value %s, want %s", snap.TotalValue(), expected)
		}
	})

	t.Run("ToPortfolioValuation converts accurately", func(t *testing.T) {
		v := snap.ToPortfolioValuation()

		if v.PortfolioID != portfolioID {
			t.Errorf("got portfolioID %s, want %s", v.PortfolioID, portfolioID)
		}
		if !v.ValuationDate.Equal(asOfDate) {
			t.Errorf("got valuationDate %v, want %v", v.ValuationDate, asOfDate)
		}
		if !v.MarketValueBase.Equal(snap.MarketValueBase) {
			t.Errorf("got marketValueBase %s, want %s", v.MarketValueBase, snap.MarketValueBase)
		}
		if !v.CashValueBase.Equal(snap.CashValueBase) {
			t.Errorf("got cashValueBase %s, want %s", v.CashValueBase, snap.CashValueBase)
		}
		if !v.NetFlowBase.Equal(snap.NetFlowBase) {
			t.Errorf("got netFlowBase %s, want %s", v.NetFlowBase, snap.NetFlowBase)
		}
		if !v.DailyReturn.Equal(dailyRet) {
			t.Errorf("got dailyReturn %s, want %s", v.DailyReturn, dailyRet)
		}
		if !v.TWRIndex.Equal(snap.TWRIndex) {
			t.Errorf("got twrIndex %s, want %s", v.TWRIndex, snap.TWRIndex)
		}
	})

	t.Run("NewPortfolioValuationSnapshot converts from PortfolioValuation", func(t *testing.T) {
		v := snap.ToPortfolioValuation()
		recon := domain.NewPortfolioValuationSnapshot(v)

		if recon.PortfolioID != snap.PortfolioID {
			t.Errorf("got portfolioID %s, want %s", recon.PortfolioID, snap.PortfolioID)
		}
		if recon.DailyReturn == nil || !recon.DailyReturn.Equal(dailyRet) {
			t.Errorf("got dailyReturn %v, want %s", recon.DailyReturn, dailyRet)
		}
	})
}

package service_test

import (
	"testing"
	"time"

	"portfolio-api/internal/domain"
	"portfolio-api/internal/service"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestProcessLedgerAverageCostVsFIFO(t *testing.T) {
	portfolioID := uuid.New()
	instrumentID := uuid.New()
	t1 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)
	t3 := time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC)

	qty10 := decimal.NewFromInt(10)
	price100 := decimal.NewFromInt(100)
	price200 := decimal.NewFromInt(200)
	price250 := decimal.NewFromInt(250)

	amt1000 := decimal.NewFromInt(1000)
	amt2000 := decimal.NewFromInt(2000)
	amt2500 := decimal.NewFromInt(2500)

	txs := []domain.Transaction{
		{
			ID:           uuid.New(),
			PortfolioID:  portfolioID,
			InstrumentID: &instrumentID,
			Type:         domain.TxTypeBuy,
			TradeDate:    t1,
			Quantity:     &qty10,
			Price:        &price100,
			Amount:       amt1000,
			CurrencyCode: "USD",
		},
		{
			ID:           uuid.New(),
			PortfolioID:  portfolioID,
			InstrumentID: &instrumentID,
			Type:         domain.TxTypeBuy,
			TradeDate:    t2,
			Quantity:     &qty10,
			Price:        &price200,
			Amount:       amt2000,
			CurrencyCode: "USD",
		},
		{
			ID:           uuid.New(),
			PortfolioID:  portfolioID,
			InstrumentID: &instrumentID,
			Type:         domain.TxTypeSell,
			TradeDate:    t3,
			Quantity:     &qty10,
			Price:        &price250,
			Amount:       amt2500,
			CurrencyCode: "USD",
		},
	}

	// 1. Test AVERAGE_COST
	pAvg := domain.Portfolio{
		ID:              portfolioID,
		BaseCurrency:    "USD",
		CostBasisMethod: domain.CostBasisMethodAverageCost,
	}
	lotsAvg, dispAvg, holdingsAvg, _, err := service.ProcessLedger(pAvg, txs, nil)
	if err != nil {
		t.Fatalf("ProcessLedger AVERAGE_COST failed: %v", err)
	}

	totalReleasedCostAvg := decimal.Zero
	totalRealizedPnLAvg := decimal.Zero
	for _, d := range dispAvg {
		totalReleasedCostAvg = totalReleasedCostAvg.Add(d.CostBasisReleasedBase)
		totalRealizedPnLAvg = totalRealizedPnLAvg.Add(d.RealizedPnLBase)
	}

	// In AVERAGE_COST: 20 units total cost 3,000 -> avg cost 150/unit.
	// Selling 10 units releases 1,500 cost, realized PnL = 2,500 - 1,500 = 1,000.
	if !totalReleasedCostAvg.Equal(decimal.NewFromInt(1500)) {
		t.Errorf("AVERAGE_COST: got released cost %s, want 1500", totalReleasedCostAvg)
	}
	if !totalRealizedPnLAvg.Equal(decimal.NewFromInt(1000)) {
		t.Errorf("AVERAGE_COST: got realized PnL %s, want 1000", totalRealizedPnLAvg)
	}
	if len(holdingsAvg) != 1 || !holdingsAvg[0].CostBasisBase.Equal(decimal.NewFromInt(1500)) {
		t.Errorf("AVERAGE_COST: got remaining holding cost basis %s, want 1500", holdingsAvg[0].CostBasisBase)
	}

	// 2. Test FIFO
	pFIFO := domain.Portfolio{
		ID:              portfolioID,
		BaseCurrency:    "USD",
		CostBasisMethod: domain.CostBasisMethodFIFO,
	}
	lotsFIFO, dispFIFO, holdingsFIFO, _, err := service.ProcessLedger(pFIFO, txs, nil)
	if err != nil {
		t.Fatalf("ProcessLedger FIFO failed: %v", err)
	}

	totalReleasedCostFIFO := decimal.Zero
	totalRealizedPnLFIFO := decimal.Zero
	for _, d := range dispFIFO {
		totalReleasedCostFIFO = totalReleasedCostFIFO.Add(d.CostBasisReleasedBase)
		totalRealizedPnLFIFO = totalRealizedPnLFIFO.Add(d.RealizedPnLBase)
	}

	// In FIFO: First 10 units bought @ 100 (cost 1,000) are sold.
	// Selling 10 units releases 1,000 cost, realized PnL = 2,500 - 1,000 = 1,500.
	// Remaining 10 units @ 200 (cost 2,000).
	if !totalReleasedCostFIFO.Equal(decimal.NewFromInt(1000)) {
		t.Errorf("FIFO: got released cost %s, want 1000", totalReleasedCostFIFO)
	}
	if !totalRealizedPnLFIFO.Equal(decimal.NewFromInt(1500)) {
		t.Errorf("FIFO: got realized PnL %s, want 1500", totalRealizedPnLFIFO)
	}
	if len(holdingsFIFO) != 1 || !holdingsFIFO[0].CostBasisBase.Equal(decimal.NewFromInt(2000)) {
		t.Errorf("FIFO: got remaining holding cost basis %s, want 2000", holdingsFIFO[0].CostBasisBase)
	}

	_ = lotsAvg
	_ = lotsFIFO
}

func TestProcessLedgerStockSplit(t *testing.T) {
	portfolioID := uuid.New()
	instrumentID := uuid.New()
	t1 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)
	t3 := time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC)

	qty10 := decimal.NewFromInt(10)
	price100 := decimal.NewFromInt(100)
	amt1000 := decimal.NewFromInt(1000)

	// Buy 10 units @ 100 on t1
	buyTx := domain.Transaction{
		ID:           uuid.New(),
		PortfolioID:  portfolioID,
		InstrumentID: &instrumentID,
		Type:         domain.TxTypeBuy,
		TradeDate:    t1,
		Quantity:     &qty10,
		Price:        &price100,
		Amount:       amt1000,
		CurrencyCode: "USD",
	}

	t.Run("2:1 forward split doubles shares while preserving cost basis", func(t *testing.T) {
		splitRatio2 := decimal.NewFromInt(2)
		splitTx := domain.Transaction{
			ID:           uuid.New(),
			PortfolioID:  portfolioID,
			InstrumentID: &instrumentID,
			Type:         domain.TxTypeSplit,
			TradeDate:    t2,
			Quantity:     &splitRatio2,
			Amount:       decimal.Zero,
			CurrencyCode: "USD",
		}

		p := domain.Portfolio{
			ID:              portfolioID,
			BaseCurrency:    "USD",
			CostBasisMethod: domain.CostBasisMethodFIFO,
		}

		lots, disposals, holdings, cash, err := service.ProcessLedger(p, []domain.Transaction{buyTx, splitTx}, nil)
		if err != nil {
			t.Fatalf("ProcessLedger failed: %v", err)
		}

		if len(holdings) != 1 {
			t.Fatalf("expected 1 holding, got %d", len(holdings))
		}
		// Quantity doubled: 10 * 2 = 20
		if !holdings[0].Quantity.Equal(decimal.NewFromInt(20)) {
			t.Errorf("expected holding quantity 20, got %s", holdings[0].Quantity)
		}
		// Cost basis unchanged: 1000
		if !holdings[0].CostBasis.Equal(amt1000) {
			t.Errorf("expected cost basis 1000, got %s", holdings[0].CostBasis)
		}
		if len(lots) != 1 {
			t.Fatalf("expected 1 lot, got %d", len(lots))
		}
		if !lots[0].RemainingQuantity.Equal(decimal.NewFromInt(20)) {
			t.Errorf("expected lot remaining quantity 20, got %s", lots[0].RemainingQuantity)
		}
		if len(disposals) != 0 {
			t.Errorf("expected 0 disposals, got %d", len(disposals))
		}
		if len(cash) != 1 || !cash[0].Balance.Equal(decimal.NewFromInt(-1000)) {
			t.Errorf("expected 1 cash balance with -1000, got %v", cash)
		}
	})

	t.Run("post-split sale relieves split-adjusted cost basis", func(t *testing.T) {
		splitRatio2 := decimal.NewFromInt(2)
		splitTx := domain.Transaction{
			ID:           uuid.New(),
			PortfolioID:  portfolioID,
			InstrumentID: &instrumentID,
			Type:         domain.TxTypeSplit,
			TradeDate:    t2,
			Quantity:     &splitRatio2,
			Amount:       decimal.Zero,
			CurrencyCode: "USD",
		}

		// Sell 10 shares @ $80 on t3 (out of 20 split shares with total cost basis $1000 -> $50/sh cost)
		sellQty := decimal.NewFromInt(10)
		sellPrice := decimal.NewFromInt(80)
		sellAmt := decimal.NewFromInt(800)
		sellTx := domain.Transaction{
			ID:           uuid.New(),
			PortfolioID:  portfolioID,
			InstrumentID: &instrumentID,
			Type:         domain.TxTypeSell,
			TradeDate:    t3,
			Quantity:     &sellQty,
			Price:        &sellPrice,
			Amount:       sellAmt,
			CurrencyCode: "USD",
		}

		p := domain.Portfolio{
			ID:              portfolioID,
			BaseCurrency:    "USD",
			CostBasisMethod: domain.CostBasisMethodFIFO,
		}

		_, disposals, holdings, _, err := service.ProcessLedger(p, []domain.Transaction{buyTx, splitTx, sellTx}, nil)
		if err != nil {
			t.Fatalf("ProcessLedger failed: %v", err)
		}

		if len(holdings) != 1 {
			t.Fatalf("expected 1 holding, got %d", len(holdings))
		}
		// Remaining quantity: 20 - 10 = 10
		if !holdings[0].Quantity.Equal(decimal.NewFromInt(10)) {
			t.Errorf("expected remaining holding quantity 10, got %s", holdings[0].Quantity)
		}
		// Remaining cost basis: 1000 - 500 = 500
		if !holdings[0].CostBasis.Equal(decimal.NewFromInt(500)) {
			t.Errorf("expected remaining cost basis 500, got %s", holdings[0].CostBasis)
		}

		if len(disposals) != 1 {
			t.Fatalf("expected 1 disposal, got %d", len(disposals))
		}
		// Relieved cost basis = 500
		if !disposals[0].CostBasisReleasedBase.Equal(decimal.NewFromInt(500)) {
			t.Errorf("expected released cost 500, got %s", disposals[0].CostBasisReleasedBase)
		}
		// Realized PnL = 800 - 500 = 300
		if !disposals[0].RealizedPnLBase.Equal(decimal.NewFromInt(300)) {
			t.Errorf("expected realized PnL 300, got %s", disposals[0].RealizedPnLBase)
		}
	})

	t.Run("split ratio with price as ratio_from", func(t *testing.T) {
		// 3:2 split: quantity=3, price=2 => splitRatio = 1.5
		ratioTo := decimal.NewFromInt(3)
		ratioFrom := decimal.NewFromInt(2)
		splitTx := domain.Transaction{
			ID:           uuid.New(),
			PortfolioID:  portfolioID,
			InstrumentID: &instrumentID,
			Type:         domain.TxTypeSplit,
			TradeDate:    t2,
			Quantity:     &ratioTo,
			Price:        &ratioFrom,
			Amount:       decimal.Zero,
			CurrencyCode: "USD",
		}

		p := domain.Portfolio{
			ID:              portfolioID,
			BaseCurrency:    "USD",
			CostBasisMethod: domain.CostBasisMethodAverageCost,
		}

		_, _, holdings, _, err := service.ProcessLedger(p, []domain.Transaction{buyTx, splitTx}, nil)
		if err != nil {
			t.Fatalf("ProcessLedger failed: %v", err)
		}

		// 10 * 1.5 = 15
		if !holdings[0].Quantity.Equal(decimal.RequireFromString("15")) {
			t.Errorf("expected quantity 15, got %s", holdings[0].Quantity)
		}
		// Cost basis unchanged: 1000
		if !holdings[0].CostBasis.Equal(amt1000) {
			t.Errorf("expected cost basis 1000, got %s", holdings[0].CostBasis)
		}
	})
}

func TestProcessLedgerMultiCurrencyBrokerageFee(t *testing.T) {
	portfolioID := uuid.New()
	instrumentID := uuid.New()
	t1 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)

	qty10 := decimal.NewFromInt(10)
	price100 := decimal.NewFromInt(100)
	amt1000 := decimal.NewFromInt(1000)
	fee15 := decimal.NewFromInt(15)
	audFeeCurrency := "AUD"
	fxRateUSDToAUD := decimal.RequireFromString("1.5") // 1 USD = 1.5 AUD

	p := domain.Portfolio{
		ID:              portfolioID,
		BaseCurrency:    "AUD",
		CostBasisMethod: domain.CostBasisMethodAverageCost,
	}

	buyTx := domain.Transaction{
		ID:              uuid.New(),
		PortfolioID:     portfolioID,
		InstrumentID:    &instrumentID,
		Type:            domain.TxTypeBuy,
		TradeDate:       t1,
		Quantity:        &qty10,
		Price:           &price100,
		Amount:          amt1000,
		CurrencyCode:    "USD",
		Fee:             fee15,
		FeeCurrencyCode: &audFeeCurrency,
		FXRateToBase:    &fxRateUSDToAUD,
	}

	lots, _, holdings, cash, err := service.ProcessLedger(p, []domain.Transaction{buyTx}, nil)
	if err != nil {
		t.Fatalf("ProcessLedger failed: %v", err)
	}

	if len(lots) != 1 {
		t.Fatalf("expected 1 lot, got %d", len(lots))
	}

	// CostBasisBase: 1000 USD * 1.5 = 1500 AUD + 15 AUD fee = 1515 AUD
	expectedCostBasisBase := decimal.RequireFromString("1515")
	if !lots[0].CostBasisBase.Equal(expectedCostBasisBase) {
		t.Errorf("expected CostBasisBase 1515, got %s", lots[0].CostBasisBase)
	}

	// CostBasis in USD: 1000 + 15/1.5 = 1010 USD
	expectedCostBasisUSD := decimal.RequireFromString("1010")
	if !lots[0].CostBasis.Equal(expectedCostBasisUSD) {
		t.Errorf("expected CostBasis 1010 USD, got %s", lots[0].CostBasis)
	}

	if len(holdings) != 1 {
		t.Fatalf("expected 1 holding, got %d", len(holdings))
	}
	if !holdings[0].Quantity.Equal(qty10) {
		t.Errorf("expected holding quantity 10, got %s", holdings[0].Quantity)
	}

	// Check cash balances: USD should be -1000, AUD should be -15
	cashMap := make(map[string]decimal.Decimal)
	for _, c := range cash {
		cashMap[c.CurrencyCode] = c.Balance
	}

	if !cashMap["USD"].Equal(decimal.NewFromInt(-1000)) {
		t.Errorf("expected USD cash balance -1000, got %s", cashMap["USD"])
	}
	if !cashMap["AUD"].Equal(decimal.NewFromInt(-15)) {
		t.Errorf("expected AUD cash balance -15, got %s", cashMap["AUD"])
	}

	// Now test Sell with AUD fee
	sellPrice150 := decimal.NewFromInt(150)
	sellAmt1500 := decimal.NewFromInt(1500)
	sellTx := domain.Transaction{
		ID:              uuid.New(),
		PortfolioID:     portfolioID,
		InstrumentID:    &instrumentID,
		Type:            domain.TxTypeSell,
		TradeDate:       t2,
		Quantity:        &qty10,
		Price:           &sellPrice150,
		Amount:          sellAmt1500,
		CurrencyCode:    "USD",
		Fee:             fee15,
		FeeCurrencyCode: &audFeeCurrency,
		FXRateToBase:    &fxRateUSDToAUD,
	}

	_, disposals, holdingsAfterSell, cashAfterSell, err := service.ProcessLedger(p, []domain.Transaction{buyTx, sellTx}, nil)
	if err != nil {
		t.Fatalf("ProcessLedger sell failed: %v", err)
	}

	if len(holdingsAfterSell) != 0 {
		t.Fatalf("expected 0 holdings after full sale, got %d", len(holdingsAfterSell))
	}

	cashMapAfter := make(map[string]decimal.Decimal)
	for _, c := range cashAfterSell {
		cashMapAfter[c.CurrencyCode] = c.Balance
	}

	// USD cash: -1000 + 1500 = +500
	if !cashMapAfter["USD"].Equal(decimal.NewFromInt(500)) {
		t.Errorf("expected USD cash 500, got %s", cashMapAfter["USD"])
	}
	// AUD cash: -15 (buy fee) - 15 (sell fee) = -30
	if !cashMapAfter["AUD"].Equal(decimal.NewFromInt(-30)) {
		t.Errorf("expected AUD cash -30, got %s", cashMapAfter["AUD"])
	}

	// Realized PnL:
	// Net proceeds base: 1500 USD * 1.5 - 15 AUD = 2250 - 15 = 2235 AUD
	// Cost basis released base: 1515 AUD
	// Realized PnL base: 2235 - 1515 = 720 AUD
	if len(disposals) != 1 {
		t.Fatalf("expected 1 disposal, got %d", len(disposals))
	}
	expectedPnL := decimal.RequireFromString("720")
	if !disposals[0].RealizedPnLBase.Equal(expectedPnL) {
		t.Errorf("expected RealizedPnLBase 720, got %s", disposals[0].RealizedPnLBase)
	}
}



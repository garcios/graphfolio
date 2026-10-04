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

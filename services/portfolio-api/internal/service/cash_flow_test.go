package service

import (
	"context"
	"testing"
	"time"

	"portfolio-api/internal/domain"
	"portfolio-api/internal/repository/mocks"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/mock/gomock"
)

func TestGetCashFlowReport(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	portfolioID := uuid.New()
	userUUID := uuid.New()
	userID := userUUID.String()

	portfolio := &domain.Portfolio{
		ID:           portfolioID,
		UserID:       userUUID,
		Name:         "Main Portfolio",
		BaseCurrency: "AUD",
	}

	fixedNow := time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)
	svc := &portfolioService{
		repo:    mockRepo,
		nowFunc: func() time.Time { return fixedNow },
	}

	symBHP := "BHP.AX"
	symAAPL := "AAPL"
	instNameBHP := "BHP Group Ltd"
	instNameAAPL := "Apple Inc"

	// Historical transaction ledger
	// 1. Prior to 2026-10-01 (Pre-period starting balance)
	// Deposit AUD 10,000 on 2026-09-01
	tx1 := domain.TransactionWithInstrument{
		Transaction: domain.Transaction{
			ID:           uuid.New(),
			PortfolioID:  portfolioID,
			Type:         domain.TxTypeDeposit,
			TradeDate:    time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
			Amount:       decimal.NewFromInt(10000),
			CurrencyCode: "AUD",
		},
	}
	// Buy BHP AUD 4,000 + AUD 10 fee on 2026-09-10
	bhpQty := decimal.NewFromInt(100)
	bhpPrice := decimal.NewFromInt(40)
	tx2 := domain.TransactionWithInstrument{
		Transaction: domain.Transaction{
			ID:           uuid.New(),
			PortfolioID:  portfolioID,
			Type:         domain.TxTypeBuy,
			TradeDate:    time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC),
			Quantity:     &bhpQty,
			Price:        &bhpPrice,
			Amount:       decimal.NewFromInt(4000),
			CurrencyCode: "AUD",
			Fee:          decimal.NewFromInt(10),
		},
		Symbol:         &symBHP,
		InstrumentName: &instNameBHP,
	}
	// Starting cash on 2026-10-01 should be: 10,000 - 4,010 = 5,990 AUD.

	// 2. In-period transactions (October 2026)
	// Dividend from BHP: AUD 200 (no withholding tax) on 2026-10-02
	tx3 := domain.TransactionWithInstrument{
		Transaction: domain.Transaction{
			ID:           uuid.New(),
			PortfolioID:  portfolioID,
			Type:         domain.TxTypeDividend,
			TradeDate:    time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC),
			Amount:       decimal.NewFromInt(200),
			CurrencyCode: "AUD",
		},
		Symbol:         &symBHP,
		InstrumentName: &instNameBHP,
	}

	// Interest on cash: AUD 15 on 2026-10-05
	tx4 := domain.TransactionWithInstrument{
		Transaction: domain.Transaction{
			ID:           uuid.New(),
			PortfolioID:  portfolioID,
			Type:         domain.TxTypeInterest,
			TradeDate:    time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC),
			Amount:       decimal.NewFromInt(15),
			CurrencyCode: "AUD",
		},
	}

	// Stock purchase (AAPL): USD 1,000 @ FX 1.50 -> AUD 1,500 on 2026-10-08
	fxRateUSD := decimal.RequireFromString("1.50")
	tx5 := domain.TransactionWithInstrument{
		Transaction: domain.Transaction{
			ID:           uuid.New(),
			PortfolioID:  portfolioID,
			Type:         domain.TxTypeBuy,
			TradeDate:    time.Date(2026, 10, 8, 0, 0, 0, 0, time.UTC),
			Amount:       decimal.NewFromInt(1000),
			CurrencyCode: "USD",
			FXRateToBase: &fxRateUSD,
		},
		Symbol:         &symAAPL,
		InstrumentName: &instNameAAPL,
	}

	// Stock sale of BHP: AUD 1,200 proceeds - AUD 10 fee = net AUD 1,190 on 2026-10-10
	tx6 := domain.TransactionWithInstrument{
		Transaction: domain.Transaction{
			ID:           uuid.New(),
			PortfolioID:  portfolioID,
			Type:         domain.TxTypeSell,
			TradeDate:    time.Date(2026, 10, 10, 0, 0, 0, 0, time.UTC),
			Amount:       decimal.NewFromInt(1200),
			CurrencyCode: "AUD",
			Fee:          decimal.NewFromInt(10),
		},
		Symbol:         &symBHP,
		InstrumentName: &instNameBHP,
	}

	// Withdrawal: AUD 500 on 2026-10-12
	tx7 := domain.TransactionWithInstrument{
		Transaction: domain.Transaction{
			ID:           uuid.New(),
			PortfolioID:  portfolioID,
			Type:         domain.TxTypeWithdrawal,
			TradeDate:    time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC),
			Amount:       decimal.NewFromInt(500),
			CurrencyCode: "AUD",
		},
	}

	allTxs := []domain.TransactionWithInstrument{tx1, tx2, tx3, tx4, tx5, tx6, tx7}

	t.Run("MTD timeframe correctly computes starting balance, inflows, outflows, and ending balance", func(t *testing.T) {
		mockRepo.EXPECT().FindPortfolioByUser(gomock.Any(), userID).Return(portfolio, nil)
		mockRepo.EXPECT().GetCashFlowTransactions(gomock.Any(), portfolioID, gomock.Any()).Return(allTxs, nil)

		filter := domain.CashFlowFilter{
			UserID:    userID,
			Timeframe: domain.CashFlowTimeframeMTD,
		}

		report, err := svc.GetCashFlowReport(context.Background(), filter)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Starting cash should be 10,000 - 4,010 = 5,990
		expectedStarting := decimal.RequireFromString("5990.00")
		if !report.Summary.StartingCashBalance.Amount.Equal(expectedStarting) {
			t.Errorf("expected starting cash %s, got %s", expectedStarting, report.Summary.StartingCashBalance.Amount)
		}

		// Inflows:
		// Dividend (200) + Interest (15) + Sale Proceeds (1190) = 1405
		expectedInflows := decimal.RequireFromString("1405.00")
		if !report.Summary.TotalInflows.Amount.Equal(expectedInflows) {
			t.Errorf("expected total inflows %s, got %s", expectedInflows, report.Summary.TotalInflows.Amount)
		}

		// Outflows:
		// Buy AAPL (1500) + Withdrawal (500) = 2000
		expectedOutflows := decimal.RequireFromString("2000.00")
		if !report.Summary.TotalOutflows.Amount.Equal(expectedOutflows) {
			t.Errorf("expected total outflows %s, got %s", expectedOutflows, report.Summary.TotalOutflows.Amount)
		}

		// Net Cash Flow: 1405 - 2000 = -595
		expectedNet := decimal.RequireFromString("-595.00")
		if !report.Summary.NetCashFlow.Amount.Equal(expectedNet) {
			t.Errorf("expected net cash flow %s, got %s", expectedNet, report.Summary.NetCashFlow.Amount)
		}

		// Ending Cash Balance: 5990 - 595 = 5395
		expectedEnding := decimal.RequireFromString("5395.00")
		if !report.Summary.EndingCashBalance.Amount.Equal(expectedEnding) {
			t.Errorf("expected ending cash %s, got %s", expectedEnding, report.Summary.EndingCashBalance.Amount)
		}

		// Verify Category Breakdowns
		if !report.Breakdown.Dividends.Amount.Equal(decimal.RequireFromString("200.00")) {
			t.Errorf("expected dividends breakdown 200.00, got %s", report.Breakdown.Dividends.Amount)
		}
		if !report.Breakdown.Interest.Amount.Equal(decimal.RequireFromString("15.00")) {
			t.Errorf("expected interest breakdown 15.00, got %s", report.Breakdown.Interest.Amount)
		}
		if !report.Breakdown.SalesProceeds.Amount.Equal(decimal.RequireFromString("1190.00")) {
			t.Errorf("expected sales proceeds breakdown 1190.00, got %s", report.Breakdown.SalesProceeds.Amount)
		}
		if !report.Breakdown.Purchases.Amount.Equal(decimal.RequireFromString("1500.00")) {
			t.Errorf("expected purchases breakdown 1500.00, got %s", report.Breakdown.Purchases.Amount)
		}
		if !report.Breakdown.Withdrawals.Amount.Equal(decimal.RequireFromString("500.00")) {
			t.Errorf("expected withdrawals breakdown 500.00, got %s", report.Breakdown.Withdrawals.Amount)
		}

		// Verify Itemized Ledger length and running balance
		if len(report.Items) != 5 {
			t.Fatalf("expected 5 in-period items, got %d", len(report.Items))
		}

		// Final item running balance MUST equal Ending Cash Balance!
		lastItem := report.Items[len(report.Items)-1]
		if !lastItem.RunningBalance.Amount.Equal(expectedEnding) {
			t.Errorf("expected last item running balance %s, got %s", expectedEnding, lastItem.RunningBalance.Amount)
		}
	})

	t.Run("Negative cash balance (overdraft) does not break mathematical reconciliation", func(t *testing.T) {
		mockRepo.EXPECT().FindPortfolioByUser(gomock.Any(), userID).Return(portfolio, nil)
		// Only buy without deposit
		mockRepo.EXPECT().GetCashFlowTransactions(gomock.Any(), portfolioID, gomock.Any()).Return([]domain.TransactionWithInstrument{tx2}, nil)

		filter := domain.CashFlowFilter{
			UserID:    userID,
			Timeframe: domain.CashFlowTimeframeAll,
		}

		report, err := svc.GetCashFlowReport(context.Background(), filter)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !report.Summary.StartingCashBalance.Amount.IsZero() {
			t.Errorf("expected starting cash 0, got %s", report.Summary.StartingCashBalance.Amount)
		}

		expectedOutflow := decimal.RequireFromString("4010.00")
		if !report.Summary.TotalOutflows.Amount.Equal(expectedOutflow) {
			t.Errorf("expected total outflows %s, got %s", expectedOutflow, report.Summary.TotalOutflows.Amount)
		}

		expectedEnding := decimal.RequireFromString("-4010.00")
		if !report.Summary.EndingCashBalance.Amount.Equal(expectedEnding) {
			t.Errorf("expected negative ending cash %s, got %s", expectedEnding, report.Summary.EndingCashBalance.Amount)
		}

		// Reconcile: Starting (0) + Inflow (0) - Outflow (4010) == Ending (-4010)
		reconciled := report.Summary.StartingCashBalance.Amount.Add(report.Summary.TotalInflows.Amount).Sub(report.Summary.TotalOutflows.Amount)
		if !reconciled.Equal(report.Summary.EndingCashBalance.Amount) {
			t.Errorf("reconciliation failed: %s != %s", reconciled, report.Summary.EndingCashBalance.Amount)
		}
	})
}

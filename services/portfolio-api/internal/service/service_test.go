package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"portfolio-api/internal/domain"
	"portfolio-api/internal/repository"
	mocks "portfolio-api/internal/repository/mocks"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/mock/gomock"
)

func TestPortfolioService_GetPortfolioSummary(t *testing.T) {
	ctx := context.Background()
	userID := "user-123"
	userUUID := uuid.New()
	portfolioID := uuid.New()
	instrumentID := uuid.New()

	mockPortfolio := &domain.Portfolio{
		ID:              portfolioID,
		UserID:          userUUID,
		Name:            "Core Investment",
		BaseCurrency:    "USD",
		CostBasisMethod: domain.CostBasisMethodAverageCost,
		CreatedAt:       time.Now().UTC().Add(-48 * time.Hour),
	}

	mockHoldings := []domain.HoldingWithPrice{
		{
			PortfolioID:        portfolioID,
			InstrumentID:       instrumentID,
			Ticker:             "AAPL",
			Name:               "Apple Inc.",
			InstrumentCurrency: "USD",
			Quantity:           decimal.NewFromInt(10),
			CostBasis:          decimal.NewFromInt(1500),
			CostBasisBase:      decimal.NewFromInt(1500),
			RealizedPnLBase:    decimal.Zero,
			DividendsBase:      decimal.Zero,
			LatestPrice:        decimal.NewFromInt(180),
			PrevPrice:          decimal.NewFromInt(175),
			FXRateToBase:       decimal.NewFromInt(1),
		},
	}

	mockCash := []domain.CashBalance{
		{
			PortfolioID:  portfolioID,
			CurrencyCode: "USD",
			Balance:      decimal.NewFromInt(500),
		},
		{
			PortfolioID:  portfolioID,
			CurrencyCode: "EUR",
			Balance:      decimal.NewFromInt(200),
		},
	}

	mockValuation := &domain.PortfolioValuation{
		PortfolioID:     portfolioID,
		ValuationDate:   time.Now().UTC(),
		MarketValueBase: decimal.NewFromInt(1800),
		CashValueBase:   decimal.NewFromInt(720),
		NetFlowBase:     decimal.Zero,
		DailyReturn:     decimal.RequireFromString("0.02"),
		TWRIndex:        decimal.RequireFromString("1.15"),
	}

	mockFXRates := map[string]decimal.Decimal{
		"USD": decimal.NewFromInt(1),
		"EUR": decimal.RequireFromString("1.10"),
	}

	t.Run("success with complete portfolio data", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := mocks.NewMockRepository(ctrl)
		svc := NewPortfolioService(mockRepo)

		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)
		mockRepo.EXPECT().GetHoldingsWithMarketData(ctx, portfolioID).Return(mockHoldings, nil)
		mockRepo.EXPECT().GetCashBalances(ctx, portfolioID).Return(mockCash, nil)
		mockRepo.EXPECT().GetLatestValuation(ctx, portfolioID).Return(mockValuation, nil)
		mockRepo.EXPECT().GetCashFXRates(ctx, "USD").Return(mockFXRates, nil)

		summary, err := svc.GetPortfolioSummary(ctx, userID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if summary == nil {
			t.Fatalf("expected non-nil summary")
		}

		if summary.Portfolio.ID != portfolioID {
			t.Errorf("expected portfolio ID %s, got %s", portfolioID, summary.Portfolio.ID)
		}

		// Expected Cash = 500 USD + (200 * 1.10) USD = 720 USD
		expectedCash := decimal.NewFromInt(720)
		if !summary.CashBalance.Amount.Equal(expectedCash) {
			t.Errorf("expected cash balance %s, got %s", expectedCash, summary.CashBalance.Amount)
		}

		// 10 * 180 = 1800 equity + 720 cash = 2520 Total Value
		expectedTotal := decimal.NewFromInt(2520)
		if !summary.TotalValue.Amount.Equal(expectedTotal) {
			t.Errorf("expected total value %s, got %s", expectedTotal, summary.TotalValue.Amount)
		}

		if len(summary.Investments) != 1 {
			t.Fatalf("expected 1 investment, got %d", len(summary.Investments))
		}
		inv := summary.Investments[0]
		if inv.Ticker != "AAPL" {
			t.Errorf("expected ticker AAPL, got %s", inv.Ticker)
		}
	})

	t.Run("success with nil valuation", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := mocks.NewMockRepository(ctrl)
		svc := NewPortfolioService(mockRepo)

		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)
		mockRepo.EXPECT().GetHoldingsWithMarketData(ctx, portfolioID).Return(mockHoldings, nil)
		mockRepo.EXPECT().GetCashBalances(ctx, portfolioID).Return(mockCash, nil)
		mockRepo.EXPECT().GetLatestValuation(ctx, portfolioID).Return(nil, nil)
		mockRepo.EXPECT().GetCashFXRates(ctx, "USD").Return(mockFXRates, nil)

		summary, err := svc.GetPortfolioSummary(ctx, userID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if summary == nil {
			t.Fatalf("expected non-nil summary")
		}

		// When valuation is nil, TodayReturnAmount is calculated from individual investments:
		// 10 * (180 - 175) = 50
		expectedTodayReturn := decimal.NewFromInt(50)
		if !summary.TodayReturnAmount.Amount.Equal(expectedTodayReturn) {
			t.Errorf("expected today return %s, got %s", expectedTodayReturn, summary.TodayReturnAmount.Amount)
		}
	})

	t.Run("error finding portfolio by user", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := mocks.NewMockRepository(ctrl)
		svc := NewPortfolioService(mockRepo)

		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(nil, repository.ErrPortfolioNotFound)

		summary, err := svc.GetPortfolioSummary(ctx, userID)
		if !errors.Is(err, repository.ErrPortfolioNotFound) {
			t.Fatalf("expected ErrPortfolioNotFound, got: %v", err)
		}
		if summary != nil {
			t.Fatalf("expected nil summary on error")
		}
	})

	t.Run("error getting holdings with market data", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := mocks.NewMockRepository(ctrl)
		svc := NewPortfolioService(mockRepo)

		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)
		mockRepo.EXPECT().GetHoldingsWithMarketData(ctx, portfolioID).Return(nil, errors.New("db connection failure"))

		summary, err := svc.GetPortfolioSummary(ctx, userID)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if summary != nil {
			t.Fatalf("expected nil summary on error")
		}
	})

	t.Run("error getting cash balances", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := mocks.NewMockRepository(ctrl)
		svc := NewPortfolioService(mockRepo)

		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)
		mockRepo.EXPECT().GetHoldingsWithMarketData(ctx, portfolioID).Return(mockHoldings, nil)
		mockRepo.EXPECT().GetCashBalances(ctx, portfolioID).Return(nil, errors.New("timeout reading cash"))

		summary, err := svc.GetPortfolioSummary(ctx, userID)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if summary != nil {
			t.Fatalf("expected nil summary on error")
		}
	})

	t.Run("error getting latest valuation", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := mocks.NewMockRepository(ctrl)
		svc := NewPortfolioService(mockRepo)

		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)
		mockRepo.EXPECT().GetHoldingsWithMarketData(ctx, portfolioID).Return(mockHoldings, nil)
		mockRepo.EXPECT().GetCashBalances(ctx, portfolioID).Return(mockCash, nil)
		mockRepo.EXPECT().GetLatestValuation(ctx, portfolioID).Return(nil, errors.New("valuation query failed"))

		summary, err := svc.GetPortfolioSummary(ctx, userID)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if summary != nil {
			t.Fatalf("expected nil summary on error")
		}
	})

	t.Run("error getting cash fx rates", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := mocks.NewMockRepository(ctrl)
		svc := NewPortfolioService(mockRepo)

		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)
		mockRepo.EXPECT().GetHoldingsWithMarketData(ctx, portfolioID).Return(mockHoldings, nil)
		mockRepo.EXPECT().GetCashBalances(ctx, portfolioID).Return(mockCash, nil)
		mockRepo.EXPECT().GetLatestValuation(ctx, portfolioID).Return(mockValuation, nil)
		mockRepo.EXPECT().GetCashFXRates(ctx, "USD").Return(nil, errors.New("fx service unavailable"))

		summary, err := svc.GetPortfolioSummary(ctx, userID)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if summary != nil {
			t.Fatalf("expected nil summary on error")
		}
	})
}

func TestPortfolioService_RebuildProjections(t *testing.T) {
	ctx := context.Background()
	userID := "user-456"
	portfolioID := uuid.New()
	instrumentID := uuid.New()

	mockPortfolio := &domain.Portfolio{
		ID:              portfolioID,
		UserID:          uuid.New(),
		Name:            "Retirement Fund",
		BaseCurrency:    "USD",
		CostBasisMethod: domain.CostBasisMethodFIFO,
		CreatedAt:       time.Now().UTC().Add(-100 * time.Hour),
	}

	qty := decimal.NewFromInt(50)
	price := decimal.NewFromInt(100)
	fxRate := decimal.NewFromInt(1)

	mockTxs := []domain.Transaction{
		{
			ID:           uuid.New(),
			PortfolioID:  portfolioID,
			Type:         domain.TxTypeDeposit,
			TradeDate:    time.Now().UTC().Add(-24 * time.Hour),
			Amount:       decimal.NewFromInt(10000),
			CurrencyCode: "USD",
			Fee:          decimal.Zero,
		},
		{
			ID:           uuid.New(),
			PortfolioID:  portfolioID,
			InstrumentID: &instrumentID,
			Type:         domain.TxTypeBuy,
			TradeDate:    time.Now().UTC().Add(-12 * time.Hour),
			Quantity:     &qty,
			Price:        &price,
			Amount:       decimal.NewFromInt(5000),
			CurrencyCode: "USD",
			Fee:          decimal.NewFromInt(10),
			FXRateToBase: &fxRate,
		},
	}

	t.Run("success rebuilding projections", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := mocks.NewMockRepository(ctrl)
		svc := NewPortfolioService(mockRepo)

		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)
		mockRepo.EXPECT().GetTransactions(ctx, portfolioID).Return(mockTxs, nil)
		mockRepo.EXPECT().GetCorporateActions(ctx, gomock.Len(1)).Return([]domain.CorporateAction{}, nil)
		mockRepo.EXPECT().SaveProjectionsTx(
			ctx,
			portfolioID,
			gomock.Len(1), // 1 buy tax lot
			gomock.Len(0), // 0 disposals
			gomock.Len(1), // 1 holding
			gomock.Len(1), // 1 cash balance (USD)
		).Return(nil)

		err := svc.RebuildProjections(ctx, userID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("error finding portfolio by user", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := mocks.NewMockRepository(ctrl)
		svc := NewPortfolioService(mockRepo)

		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(nil, repository.ErrPortfolioNotFound)

		err := svc.RebuildProjections(ctx, userID)
		if !errors.Is(err, repository.ErrPortfolioNotFound) {
			t.Fatalf("expected ErrPortfolioNotFound, got: %v", err)
		}
	})

	t.Run("error getting transactions", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := mocks.NewMockRepository(ctrl)
		svc := NewPortfolioService(mockRepo)

		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)
		mockRepo.EXPECT().GetTransactions(ctx, portfolioID).Return(nil, errors.New("read error"))

		err := svc.RebuildProjections(ctx, userID)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("error getting corporate actions", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := mocks.NewMockRepository(ctrl)
		svc := NewPortfolioService(mockRepo)

		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)
		mockRepo.EXPECT().GetTransactions(ctx, portfolioID).Return(mockTxs, nil)
		mockRepo.EXPECT().GetCorporateActions(ctx, gomock.Len(1)).Return(nil, errors.New("corporate actions service error"))

		err := svc.RebuildProjections(ctx, userID)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("error saving projections transaction", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := mocks.NewMockRepository(ctrl)
		svc := NewPortfolioService(mockRepo)

		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)
		mockRepo.EXPECT().GetTransactions(ctx, portfolioID).Return(mockTxs, nil)
		mockRepo.EXPECT().GetCorporateActions(ctx, gomock.Len(1)).Return([]domain.CorporateAction{}, nil)
		mockRepo.EXPECT().SaveProjectionsTx(
			ctx,
			portfolioID,
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
		).Return(errors.New("tx write conflict"))

		err := svc.RebuildProjections(ctx, userID)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})
}

func TestPortfolioService_UpdatePortfolioBaseCurrency(t *testing.T) {
	ctx := context.Background()
	userID := "user-123"
	portfolioID := uuid.New()

	t.Run("successful base currency update with fx rate", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := mocks.NewMockRepository(ctrl)
		svc := NewPortfolioService(mockRepo)

		pUSD := &domain.Portfolio{
			ID:           portfolioID,
			UserID:       uuid.New(),
			BaseCurrency: "USD",
		}
		pAUD := &domain.Portfolio{
			ID:           portfolioID,
			UserID:       uuid.New(),
			BaseCurrency: "AUD",
		}

		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(pUSD, nil)
		mockRepo.EXPECT().GetFXRate(ctx, "USD", "AUD").Return(decimal.NewFromFloat(1.523), nil)
		mockRepo.EXPECT().UpdatePortfolioBaseCurrency(ctx, portfolioID, "AUD", decimal.NewFromFloat(1.523)).Return(nil)

		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(pAUD, nil)
		mockRepo.EXPECT().GetHoldingsWithMarketData(ctx, portfolioID).Return([]domain.HoldingWithPrice{}, nil)
		mockRepo.EXPECT().GetCashBalances(ctx, portfolioID).Return([]domain.CashBalance{}, nil)
		mockRepo.EXPECT().GetLatestValuation(ctx, portfolioID).Return(nil, nil)
		mockRepo.EXPECT().GetCashFXRates(ctx, "AUD").Return(map[string]decimal.Decimal{}, nil)

		summary, err := svc.UpdatePortfolioBaseCurrency(ctx, userID, "AUD")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if summary == nil {
			t.Fatalf("expected summary, got nil")
		}
		if summary.TotalValue.CurrencyCode != "AUD" {
			t.Errorf("expected currency AUD, got %s", summary.TotalValue.CurrencyCode)
		}
	})

	t.Run("returns error on invalid currency code", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := mocks.NewMockRepository(ctrl)
		svc := NewPortfolioService(mockRepo)

		_, err := svc.UpdatePortfolioBaseCurrency(ctx, userID, "INVALID")
		if err == nil {
			t.Fatalf("expected error for invalid currency code")
		}
	})

	t.Run("noop if currency is already set", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := mocks.NewMockRepository(ctrl)
		svc := NewPortfolioService(mockRepo)

		pUSD := &domain.Portfolio{
			ID:           portfolioID,
			UserID:       uuid.New(),
			BaseCurrency: "USD",
		}

		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(pUSD, nil)
		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(pUSD, nil)
		mockRepo.EXPECT().GetHoldingsWithMarketData(ctx, portfolioID).Return([]domain.HoldingWithPrice{}, nil)
		mockRepo.EXPECT().GetCashBalances(ctx, portfolioID).Return([]domain.CashBalance{}, nil)
		mockRepo.EXPECT().GetLatestValuation(ctx, portfolioID).Return(nil, nil)
		mockRepo.EXPECT().GetCashFXRates(ctx, "USD").Return(map[string]decimal.Decimal{}, nil)

		summary, err := svc.UpdatePortfolioBaseCurrency(ctx, userID, "USD")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if summary.TotalValue.CurrencyCode != "USD" {
			t.Errorf("expected currency USD, got %s", summary.TotalValue.CurrencyCode)
		}
	})
}

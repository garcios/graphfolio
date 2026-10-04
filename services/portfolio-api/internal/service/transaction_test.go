package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"portfolio-api/internal/domain"
	"portfolio-api/internal/repository"
	repoMocks "portfolio-api/internal/repository/mocks"
	"portfolio-api/internal/service"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/mock/gomock"
)

func TestPortfolioService_AddTransaction(t *testing.T) {
	ctx := context.Background()
	userID := "user-123"
	portfolioID := uuid.New()
	instrumentID := uuid.New()

	mockPortfolio := &domain.Portfolio{
		ID:              portfolioID,
		UserID:          uuid.New(),
		Name:            "Core Portfolio",
		BaseCurrency:    "USD",
		CostBasisMethod: domain.CostBasisMethodFIFO,
		CreatedAt:       time.Now().UTC().Add(-200 * time.Hour),
	}

	mockInstrument := &domain.Instrument{
		ID:           instrumentID,
		Symbol:       "AAPL",
		ExchangeCode: "XNAS",
		Name:         "Apple Inc.",
		AssetClass:   "EQUITY",
		CurrencyCode: "USD",
		IsActive:     true,
	}

	t.Run("successful BUY transaction", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := repoMocks.NewMockRepository(ctrl)
		svc := service.NewPortfolioService(mockRepo)

		qty := decimal.NewFromInt(10)
		price := decimal.RequireFromString("150.00")
		symbol := "AAPL"
		fee := decimal.RequireFromString("5.00")

		input := domain.AddTransactionInput{
			UserID:    userID,
			Type:      domain.TxTypeBuy,
			Symbol:    &symbol,
			TradeDate: time.Now().UTC(),
			Quantity:  &qty,
			Price:     &price,
			Fee:       &fee,
		}

		// 1. Initial lookup in AddTransaction
		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)
		mockRepo.EXPECT().FindInstrumentBySymbol(ctx, "AAPL").Return(mockInstrument, nil)
		mockRepo.EXPECT().InsertTransaction(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, tx domain.Transaction) (*domain.Transaction, error) {
			if !tx.Amount.Equal(decimal.RequireFromString("1500.00")) {
				t.Fatalf("expected amount 1500, got %s", tx.Amount)
			}
			if !tx.Fee.Equal(decimal.RequireFromString("5.00")) {
				t.Fatalf("expected fee 5.00, got %s", tx.Fee)
			}
			tx.ID = uuid.New()
			return &tx, nil
		})

		// 2. RebuildProjections calls
		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)
		mockRepo.EXPECT().GetTransactions(ctx, portfolioID).Return([]domain.Transaction{
			{
				ID:           uuid.New(),
				PortfolioID:  portfolioID,
				InstrumentID: &instrumentID,
				Type:         domain.TxTypeBuy,
				TradeDate:    time.Now().UTC(),
				Quantity:     &qty,
				Price:        &price,
				Amount:       decimal.RequireFromString("1500.00"),
				CurrencyCode: "USD",
				Fee:          fee,
			},
		}, nil)
		mockRepo.EXPECT().GetCorporateActions(ctx, gomock.Len(1)).Return([]domain.CorporateAction{}, nil)
		mockRepo.EXPECT().SaveProjectionsTx(ctx, portfolioID, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

		// 3. GetPortfolioSummary calls
		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)
		mockRepo.EXPECT().GetHoldingsWithMarketData(ctx, portfolioID).Return([]domain.HoldingWithPrice{
			{
				PortfolioID:        portfolioID,
				InstrumentID:       instrumentID,
				Ticker:             "AAPL",
				Name:               "Apple Inc.",
				InstrumentCurrency: "USD",
				Quantity:           qty,
				CostBasis:          decimal.RequireFromString("1505.00"),
				CostBasisBase:      decimal.RequireFromString("1505.00"),
				LatestPrice:        decimal.RequireFromString("160.00"),
				PrevPrice:          decimal.RequireFromString("155.00"),
				FXRateToBase:       decimal.NewFromInt(1),
			},
		}, nil)
		mockRepo.EXPECT().GetCashBalances(ctx, portfolioID).Return([]domain.CashBalance{
			{PortfolioID: portfolioID, CurrencyCode: "USD", Balance: decimal.RequireFromString("8495.00")},
		}, nil)
		mockRepo.EXPECT().GetLatestValuation(ctx, portfolioID).Return(&domain.PortfolioValuation{
			PortfolioID:     portfolioID,
			ValuationDate:   time.Now().UTC(),
			MarketValueBase: decimal.RequireFromString("1600.00"),
			CashValueBase:   decimal.RequireFromString("8495.00"),
			TWRIndex:        decimal.NewFromInt(1),
		}, nil)
		mockRepo.EXPECT().GetCashFXRates(ctx, "USD").Return(map[string]decimal.Decimal{"USD": decimal.NewFromInt(1)}, nil)

		tx, summary, err := svc.AddTransaction(ctx, input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tx == nil {
			t.Fatalf("expected created tx, got nil")
		}
		if summary == nil {
			t.Fatalf("expected summary, got nil")
		}
		if len(summary.Investments) != 1 {
			t.Fatalf("expected 1 investment, got %d", len(summary.Investments))
		}
	})

	t.Run("successful DEPOSIT transaction", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := repoMocks.NewMockRepository(ctrl)
		svc := service.NewPortfolioService(mockRepo)

		depositAmount := decimal.RequireFromString("10000.00")
		input := domain.AddTransactionInput{
			UserID: userID,
			Type:   domain.TxTypeDeposit,
			Amount: &depositAmount,
		}

		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)
		mockRepo.EXPECT().InsertTransaction(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, tx domain.Transaction) (*domain.Transaction, error) {
			if tx.Type != domain.TxTypeDeposit {
				t.Fatalf("expected deposit, got %v", tx.Type)
			}
			if !tx.Amount.Equal(depositAmount) {
				t.Fatalf("expected amount %s, got %s", depositAmount, tx.Amount)
			}
			tx.ID = uuid.New()
			return &tx, nil
		})

		// RebuildProjections calls
		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)
		mockRepo.EXPECT().GetTransactions(ctx, portfolioID).Return([]domain.Transaction{}, nil)
		mockRepo.EXPECT().GetCorporateActions(ctx, gomock.Len(0)).Return([]domain.CorporateAction{}, nil)
		mockRepo.EXPECT().SaveProjectionsTx(ctx, portfolioID, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

		// GetPortfolioSummary calls
		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)
		mockRepo.EXPECT().GetHoldingsWithMarketData(ctx, portfolioID).Return([]domain.HoldingWithPrice{}, nil)
		mockRepo.EXPECT().GetCashBalances(ctx, portfolioID).Return([]domain.CashBalance{
			{PortfolioID: portfolioID, CurrencyCode: "USD", Balance: depositAmount},
		}, nil)
		mockRepo.EXPECT().GetLatestValuation(ctx, portfolioID).Return(nil, nil)
		mockRepo.EXPECT().GetCashFXRates(ctx, "USD").Return(map[string]decimal.Decimal{"USD": decimal.NewFromInt(1)}, nil)

		tx, summary, err := svc.AddTransaction(ctx, input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tx == nil || summary == nil {
			t.Fatalf("expected tx and summary, got nil")
		}
		if !summary.CashBalance.Amount.Equal(depositAmount) {
			t.Fatalf("expected cash balance %s, got %s", depositAmount, summary.CashBalance.Amount)
		}
	})

	t.Run("validation error on BUY without symbol", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := repoMocks.NewMockRepository(ctrl)
		svc := service.NewPortfolioService(mockRepo)

		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)

		input := domain.AddTransactionInput{
			UserID: userID,
			Type:   domain.TxTypeBuy,
		}

		_, _, err := svc.AddTransaction(ctx, input)
		if err == nil {
			t.Fatalf("expected error for missing symbol, got nil")
		}
	})

	t.Run("validation error on BUY with non-positive quantity", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := repoMocks.NewMockRepository(ctrl)
		svc := service.NewPortfolioService(mockRepo)

		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)

		sym := "AAPL"
		zeroQty := decimal.Zero
		input := domain.AddTransactionInput{
			UserID:   userID,
			Type:     domain.TxTypeBuy,
			Symbol:   &sym,
			Quantity: &zeroQty,
		}

		_, _, err := svc.AddTransaction(ctx, input)
		if err == nil {
			t.Fatalf("expected error for non-positive quantity, got nil")
		}
	})

	t.Run("validation error on negative fee", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := repoMocks.NewMockRepository(ctrl)
		svc := service.NewPortfolioService(mockRepo)

		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)

		negativeFee := decimal.RequireFromString("-10.00")
		input := domain.AddTransactionInput{
			UserID: userID,
			Type:   domain.TxTypeDeposit,
			Fee:    &negativeFee,
		}

		_, _, err := svc.AddTransaction(ctx, input)
		if err == nil {
			t.Fatalf("expected error for negative fee, got nil")
		}
	})

	t.Run("error when portfolio not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := repoMocks.NewMockRepository(ctrl)
		svc := service.NewPortfolioService(mockRepo)

		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(nil, repository.ErrPortfolioNotFound)

		input := domain.AddTransactionInput{
			UserID: userID,
			Type:   domain.TxTypeDeposit,
		}

		_, _, err := svc.AddTransaction(ctx, input)
		if !errors.Is(err, repository.ErrPortfolioNotFound) {
			t.Fatalf("expected ErrPortfolioNotFound, got: %v", err)
		}
	})
}

func TestPortfolioService_ListInstruments(t *testing.T) {
	ctx := context.Background()

	t.Run("returns list of instruments from repo", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := repoMocks.NewMockRepository(ctrl)
		svc := service.NewPortfolioService(mockRepo)

		expected := []domain.Instrument{
			{ID: uuid.New(), Symbol: "AAPL", Name: "Apple Inc.", CurrencyCode: "USD", AssetClass: "EQUITY"},
			{ID: uuid.New(), Symbol: "MSFT", Name: "Microsoft", CurrencyCode: "USD", AssetClass: "EQUITY"},
		}

		mockRepo.EXPECT().ListActiveInstruments(ctx).Return(expected, nil)

		result, err := svc.ListInstruments(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 2 {
			t.Fatalf("expected 2 instruments, got %d", len(result))
		}
		if result[0].Symbol != "AAPL" || result[1].Symbol != "MSFT" {
			t.Fatalf("symbols mismatch: %v", result)
		}
	})
}

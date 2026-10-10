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
	serviceMocks "portfolio-api/internal/service/mocks"

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

func TestPortfolioService_ListTransactions(t *testing.T) {
	ctx := context.Background()
	userID := "user-123"
	portfolioID := uuid.New()

	mockPortfolio := &domain.Portfolio{
		ID:              portfolioID,
		UserID:          uuid.New(),
		Name:            "Core Portfolio",
		BaseCurrency:    "USD",
		CostBasisMethod: domain.CostBasisMethodFIFO,
		CreatedAt:       time.Now().UTC().Add(-200 * time.Hour),
	}

	t.Run("success with filters and pagination", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := repoMocks.NewMockRepository(ctrl)
		svc := service.NewPortfolioService(mockRepo)

		txType := domain.TxTypeBuy
		sym := "aapl"
		filter := domain.TransactionFilter{
			Type:     &txType,
			Symbol:   &sym,
			Page:     2,
			PageSize: 10,
		}

		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)
		mockRepo.EXPECT().ListTransactions(ctx, portfolioID, gomock.Any()).DoAndReturn(
			func(_ context.Context, pid uuid.UUID, f domain.TransactionFilter) ([]domain.TransactionWithInstrument, int, error) {
				if pid != portfolioID {
					t.Fatalf("expected portfolioID %s, got %s", portfolioID, pid)
				}
				if f.Page != 2 || f.PageSize != 10 {
					t.Fatalf("expected page 2, pageSize 10; got page %d, pageSize %d", f.Page, f.PageSize)
				}
				if f.Type == nil || *f.Type != domain.TxTypeBuy {
					t.Fatalf("expected type BUY, got %v", f.Type)
				}
				if f.Symbol == nil || *f.Symbol != "AAPL" {
					t.Fatalf("expected uppercased symbol AAPL, got %v", f.Symbol)
				}
				qty := decimal.NewFromInt(10)
				price := decimal.RequireFromString("150.00")
				symbolStr := "AAPL"
				nameStr := "Apple Inc."
				return []domain.TransactionWithInstrument{
					{
						Transaction: domain.Transaction{
							ID:           uuid.New(),
							PortfolioID:  portfolioID,
							Type:         domain.TxTypeBuy,
							TradeDate:    time.Now().UTC(),
							Quantity:     &qty,
							Price:        &price,
							Amount:       decimal.RequireFromString("1500.00"),
							CurrencyCode: "USD",
							Fee:          decimal.RequireFromString("5.00"),
						},
						Symbol:         &symbolStr,
						InstrumentName: &nameStr,
					},
				}, 25, nil
			},
		)

		items, total, err := svc.ListTransactions(ctx, userID, filter)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 25 {
			t.Fatalf("expected total 25, got %d", total)
		}
		if len(items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(items))
		}
		if *items[0].Symbol != "AAPL" {
			t.Fatalf("expected symbol AAPL, got %s", *items[0].Symbol)
		}
	})

	t.Run("defaults pagination bounds", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := repoMocks.NewMockRepository(ctrl)
		svc := service.NewPortfolioService(mockRepo)

		filter := domain.TransactionFilter{
			Page:     -1,
			PageSize: 500,
		}

		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)
		mockRepo.EXPECT().ListTransactions(ctx, portfolioID, gomock.Any()).DoAndReturn(
			func(_ context.Context, _ uuid.UUID, f domain.TransactionFilter) ([]domain.TransactionWithInstrument, int, error) {
				if f.Page != 1 {
					t.Errorf("expected page 1, got %d", f.Page)
				}
				if f.PageSize != 100 {
					t.Errorf("expected capped pageSize 100, got %d", f.PageSize)
				}
				return []domain.TransactionWithInstrument{}, 0, nil
			},
		)

		_, _, err := svc.ListTransactions(ctx, userID, filter)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("portfolio not found returns error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := repoMocks.NewMockRepository(ctrl)
		svc := service.NewPortfolioService(mockRepo)

		mockRepo.EXPECT().FindPortfolioByUser(ctx, "unknown-user").Return(nil, repository.ErrPortfolioNotFound)

		_, _, err := svc.ListTransactions(ctx, "unknown-user", domain.TransactionFilter{})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !errors.Is(err, repository.ErrPortfolioNotFound) {
			t.Fatalf("expected ErrPortfolioNotFound, got %v", err)
		}
	})
}

func TestPortfolioService_DeleteTransaction(t *testing.T) {
	ctx := context.Background()
	userID := "user-123"
	portfolioID := uuid.New()
	txID := uuid.New()

	mockPortfolio := &domain.Portfolio{
		ID:              portfolioID,
		UserID:          uuid.New(),
		Name:            "Core Portfolio",
		BaseCurrency:    "USD",
		CostBasisMethod: domain.CostBasisMethodFIFO,
		CreatedAt:       time.Now().UTC().Add(-200 * time.Hour),
	}

	t.Run("successful deletion replays projections and returns summary", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := repoMocks.NewMockRepository(ctrl)
		svc := service.NewPortfolioService(mockRepo)

		// 1. Initial delete
		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)
		mockRepo.EXPECT().DeleteTransaction(ctx, portfolioID, txID).Return(nil)

		// 2. RebuildProjections
		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)
		mockRepo.EXPECT().GetTransactions(ctx, portfolioID).Return([]domain.Transaction{}, nil)
		mockRepo.EXPECT().GetCorporateActions(ctx, gomock.Len(0)).Return([]domain.CorporateAction{}, nil)
		mockRepo.EXPECT().SaveProjectionsTx(ctx, portfolioID, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

		// 3. GetPortfolioSummary
		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)
		mockRepo.EXPECT().GetHoldingsWithMarketData(ctx, portfolioID).Return([]domain.HoldingWithPrice{}, nil)
		mockRepo.EXPECT().GetCashBalances(ctx, portfolioID).Return([]domain.CashBalance{}, nil)
		mockRepo.EXPECT().GetLatestValuation(ctx, portfolioID).Return(nil, nil)
		mockRepo.EXPECT().GetCashFXRates(ctx, "USD").Return(map[string]decimal.Decimal{"USD": decimal.NewFromInt(1)}, nil)

		summary, err := svc.DeleteTransaction(ctx, userID, txID.String())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if summary == nil {
			t.Fatalf("expected non-nil summary")
		}
	})

	t.Run("non-existent transaction returns not found error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := repoMocks.NewMockRepository(ctrl)
		svc := service.NewPortfolioService(mockRepo)

		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)
		mockRepo.EXPECT().DeleteTransaction(ctx, portfolioID, txID).Return(repository.ErrTransactionNotFound)

		summary, err := svc.DeleteTransaction(ctx, userID, txID.String())
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if summary != nil {
			t.Fatalf("expected nil summary on error")
		}
		if !errors.Is(err, repository.ErrTransactionNotFound) {
			t.Fatalf("expected ErrTransactionNotFound, got %v", err)
		}
	})

	t.Run("invalid transaction UUID returns not found error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := repoMocks.NewMockRepository(ctrl)
		svc := service.NewPortfolioService(mockRepo)

		summary, err := svc.DeleteTransaction(ctx, userID, "not-a-valid-uuid")
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if summary != nil {
			t.Fatalf("expected nil summary on error")
		}
		if !errors.Is(err, repository.ErrTransactionNotFound) {
			t.Fatalf("expected ErrTransactionNotFound, got %v", err)
		}
	})

	t.Run("portfolio not found returns error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := repoMocks.NewMockRepository(ctrl)
		svc := service.NewPortfolioService(mockRepo)

		mockRepo.EXPECT().FindPortfolioByUser(ctx, "unknown-user").Return(nil, repository.ErrPortfolioNotFound)

		summary, err := svc.DeleteTransaction(ctx, "unknown-user", txID.String())
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if summary != nil {
			t.Fatalf("expected nil summary on error")
		}
		if !errors.Is(err, repository.ErrPortfolioNotFound) {
			t.Fatalf("expected ErrPortfolioNotFound, got %v", err)
		}
	})
}

func TestPortfolioService_ValuationHooks(t *testing.T) {
	ctx := context.Background()
	userID := "user-123"
	portfolioID := uuid.New()
	instrumentID := uuid.New()

	fixedNow := time.Date(2026, 6, 15, 14, 30, 0, 0, time.UTC)
	todayUTC := fixedNow.Truncate(24 * time.Hour)

	mockPortfolio := &domain.Portfolio{
		ID:              portfolioID,
		UserID:          uuid.New(),
		Name:            "Core Portfolio",
		BaseCurrency:    "USD",
		CostBasisMethod: domain.CostBasisMethodFIFO,
		CreatedAt:       fixedNow.Add(-30 * 24 * time.Hour),
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

	qty := decimal.NewFromInt(10)
	price := decimal.RequireFromString("150.00")
	symbol := "AAPL"

	setupSummaryExpectations := func(mockRepo *repoMocks.MockRepository) {
		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)
		mockRepo.EXPECT().GetHoldingsWithMarketData(ctx, portfolioID).Return([]domain.HoldingWithPrice{}, nil)
		mockRepo.EXPECT().GetCashBalances(ctx, portfolioID).Return([]domain.CashBalance{}, nil)
		mockRepo.EXPECT().GetLatestValuation(ctx, portfolioID).Return(nil, nil)
		mockRepo.EXPECT().GetCashFXRates(ctx, "USD").Return(map[string]decimal.Decimal{}, nil)
	}

	setupProjectionsExpectations := func(mockRepo *repoMocks.MockRepository) {
		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)
		mockRepo.EXPECT().GetTransactions(ctx, portfolioID).Return([]domain.Transaction{}, nil)
		mockRepo.EXPECT().GetCorporateActions(ctx, gomock.Any()).Return([]domain.CorporateAction{}, nil)
		mockRepo.EXPECT().SaveProjectionsTx(ctx, portfolioID, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	}

	t.Run("past-dated transaction triggers BackfillPortfolioValuations", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := repoMocks.NewMockRepository(ctrl)
		mockValuation := serviceMocks.NewMockValuationService(ctrl)

		svc := service.NewPortfolioService(mockRepo,
			service.WithValuationService(mockValuation),
			service.WithNowFunc(func() time.Time { return fixedNow }),
		)

		pastDate := fixedNow.Add(-5 * 24 * time.Hour)
		input := domain.AddTransactionInput{
			UserID:    userID,
			Type:      domain.TxTypeBuy,
			Symbol:    &symbol,
			TradeDate: pastDate,
			Quantity:  &qty,
			Price:     &price,
		}

		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)
		mockRepo.EXPECT().FindInstrumentBySymbol(ctx, "AAPL").Return(mockInstrument, nil)
		mockRepo.EXPECT().InsertTransaction(ctx, gomock.Any()).Return(&domain.Transaction{ID: uuid.New(), PortfolioID: portfolioID}, nil)

		setupProjectionsExpectations(mockRepo)

		// Hook expectation: backfill from truncated past date
		mockValuation.EXPECT().
			BackfillPortfolioValuations(ctx, portfolioID, pastDate.Truncate(24*time.Hour)).
			Return(nil)

		setupSummaryExpectations(mockRepo)

		_, summary, err := svc.AddTransaction(ctx, input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if summary == nil {
			t.Fatalf("expected non-nil summary")
		}
	})

	t.Run("current-day transaction triggers SnapshotValuation", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := repoMocks.NewMockRepository(ctrl)
		mockValuation := serviceMocks.NewMockValuationService(ctrl)

		svc := service.NewPortfolioService(mockRepo,
			service.WithValuationService(mockValuation),
			service.WithNowFunc(func() time.Time { return fixedNow }),
		)

		input := domain.AddTransactionInput{
			UserID:    userID,
			Type:      domain.TxTypeBuy,
			Symbol:    &symbol,
			TradeDate: fixedNow,
			Quantity:  &qty,
			Price:     &price,
		}

		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)
		mockRepo.EXPECT().FindInstrumentBySymbol(ctx, "AAPL").Return(mockInstrument, nil)
		mockRepo.EXPECT().InsertTransaction(ctx, gomock.Any()).Return(&domain.Transaction{ID: uuid.New(), PortfolioID: portfolioID}, nil)

		setupProjectionsExpectations(mockRepo)

		// Hook expectation: snapshot for todayUTC
		mockValuation.EXPECT().
			SnapshotValuation(ctx, portfolioID, todayUTC).
			Return(&domain.PortfolioValuationSnapshot{}, nil)

		setupSummaryExpectations(mockRepo)

		_, summary, err := svc.AddTransaction(ctx, input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if summary == nil {
			t.Fatalf("expected non-nil summary")
		}
	})

	t.Run("DeleteTransaction triggers BackfillPortfolioValuations from CreatedAt", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := repoMocks.NewMockRepository(ctrl)
		mockValuation := serviceMocks.NewMockValuationService(ctrl)

		svc := service.NewPortfolioService(mockRepo,
			service.WithValuationService(mockValuation),
			service.WithNowFunc(func() time.Time { return fixedNow }),
		)

		txID := uuid.New()
		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)
		mockRepo.EXPECT().DeleteTransaction(ctx, portfolioID, txID).Return(nil)

		setupProjectionsExpectations(mockRepo)

		// Hook expectation: backfill from portfolio.CreatedAt
		mockValuation.EXPECT().
			BackfillPortfolioValuations(ctx, portfolioID, mockPortfolio.CreatedAt).
			Return(nil)

		setupSummaryExpectations(mockRepo)

		summary, err := svc.DeleteTransaction(ctx, userID, txID.String())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if summary == nil {
			t.Fatalf("expected non-nil summary")
		}
	})
}

func TestPortfolioService_RebuildValuations(t *testing.T) {
	ctx := context.Background()
	userID := "user-123"
	portfolioID := uuid.New()

	createdAt := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	mockPortfolio := &domain.Portfolio{
		ID:        portfolioID,
		UserID:    uuid.New(),
		CreatedAt: createdAt,
	}

	t.Run("success with explicit fromDate", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := repoMocks.NewMockRepository(ctrl)
		mockValuation := serviceMocks.NewMockValuationService(ctrl)
		svc := service.NewPortfolioService(mockRepo, service.WithValuationService(mockValuation))

		fromDate := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)
		mockValuation.EXPECT().BackfillPortfolioValuations(ctx, portfolioID, fromDate).Return(nil)

		if err := svc.RebuildValuations(ctx, userID, &fromDate); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("success with nil fromDate defaults to portfolio CreatedAt", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := repoMocks.NewMockRepository(ctrl)
		mockValuation := serviceMocks.NewMockValuationService(ctrl)
		svc := service.NewPortfolioService(mockRepo, service.WithValuationService(mockValuation))

		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)
		mockValuation.EXPECT().BackfillPortfolioValuations(ctx, portfolioID, createdAt).Return(nil)

		if err := svc.RebuildValuations(ctx, userID, nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("portfolio not found returns error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := repoMocks.NewMockRepository(ctrl)
		mockValuation := serviceMocks.NewMockValuationService(ctrl)
		svc := service.NewPortfolioService(mockRepo, service.WithValuationService(mockValuation))

		mockRepo.EXPECT().FindPortfolioByUser(ctx, "unknown-user").Return(nil, repository.ErrPortfolioNotFound)

		err := svc.RebuildValuations(ctx, "unknown-user", nil)
		if !errors.Is(err, repository.ErrPortfolioNotFound) {
			t.Fatalf("expected ErrPortfolioNotFound, got %v", err)
		}
	})
}

func TestPortfolioService_CheckTransactionDuplicates(t *testing.T) {
	ctx := context.Background()
	userID := "user-123"
	portfolioID := uuid.New()

	mockPortfolio := &domain.Portfolio{
		ID:           portfolioID,
		UserID:       uuid.New(),
		Name:         "Core Portfolio",
		BaseCurrency: "AUD",
	}

	t.Run("empty external refs returns empty slice", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := repoMocks.NewMockRepository(ctrl)
		svc := service.NewPortfolioService(mockRepo)

		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)

		refs, err := svc.CheckTransactionDuplicates(ctx, userID, []string{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(refs) != 0 {
			t.Fatalf("expected 0 refs, got %d", len(refs))
		}
	})

	t.Run("queries repository and returns existing refs", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := repoMocks.NewMockRepository(ctrl)
		svc := service.NewPortfolioService(mockRepo)

		queryRefs := []string{"cs_123", "nt_456"}
		existing := []string{"cs_123"}

		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)
		mockRepo.EXPECT().FindExistingExternalRefs(ctx, portfolioID, queryRefs).Return(existing, nil)

		refs, err := svc.CheckTransactionDuplicates(ctx, userID, queryRefs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(refs) != 1 || refs[0] != "cs_123" {
			t.Fatalf("expected ['cs_123'], got %v", refs)
		}
	})

	t.Run("portfolio not found returns error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := repoMocks.NewMockRepository(ctrl)
		svc := service.NewPortfolioService(mockRepo)

		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(nil, repository.ErrPortfolioNotFound)

		_, err := svc.CheckTransactionDuplicates(ctx, userID, []string{"ref1"})
		if !errors.Is(err, repository.ErrPortfolioNotFound) {
			t.Fatalf("expected ErrPortfolioNotFound, got %v", err)
		}
	})
}

func TestPortfolioService_BatchImportTransactions(t *testing.T) {
	ctx := context.Background()
	userID := "user-123"
	portfolioID := uuid.New()
	instrumentID := uuid.New()

	mockPortfolio := &domain.Portfolio{
		ID:              portfolioID,
		UserID:          uuid.New(),
		Name:            "Core Portfolio",
		BaseCurrency:    "AUD",
		CostBasisMethod: domain.CostBasisMethodFIFO,
		CreatedAt:       time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	mockInst := &domain.Instrument{
		ID:           instrumentID,
		Symbol:       "BHP",
		ExchangeCode: "XASX",
		Name:         "BHP Group Limited",
		AssetClass:   "EQUITY",
		CurrencyCode: "AUD",
		IsActive:     true,
	}

	tradeDate := time.Date(2025, 5, 10, 0, 0, 0, 0, time.UTC)

	t.Run("empty transactions list returns zero counts", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := repoMocks.NewMockRepository(ctrl)
		svc := service.NewPortfolioService(mockRepo)

		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)
		// For GetPortfolioSummary:
		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)
		mockRepo.EXPECT().GetHoldingsWithMarketData(ctx, portfolioID).Return(nil, nil)
		mockRepo.EXPECT().GetCashBalances(ctx, portfolioID).Return(nil, nil)
		mockRepo.EXPECT().GetLatestValuation(ctx, portfolioID).Return(nil, nil)
		mockRepo.EXPECT().GetCashFXRates(ctx, "AUD").Return(map[string]decimal.Decimal{}, nil)

		res, err := svc.BatchImportTransactions(ctx, domain.BatchImportInput{
			UserID:         userID,
			Transactions:   nil,
			SkipDuplicates: true,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.ImportedCount != 0 || res.SkippedCount != 0 {
			t.Fatalf("expected 0 imported, 0 skipped; got %d, %d", res.ImportedCount, res.SkippedCount)
		}
	})

	t.Run("successfully imports new transactions and skips duplicates", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := repoMocks.NewMockRepository(ctrl)
		mockValuation := serviceMocks.NewMockValuationService(ctrl)
		svc := service.NewPortfolioService(mockRepo, service.WithValuationService(mockValuation))

		items := []domain.ImportTransactionItem{
			{
				Type:        domain.TxTypeBuy,
				Symbol:      "BHP",
				TradeDate:   tradeDate,
				Quantity:    decimal.NewFromInt(100),
				Price:       decimal.RequireFromString("45.00"),
				Amount:      decimal.RequireFromString("4519.95"),
				Fee:         decimal.RequireFromString("19.95"),
				ExternalRef: "cs_existing_1",
			},
			{
				Type:        domain.TxTypeBuy,
				Symbol:      "BHP",
				TradeDate:   tradeDate,
				Quantity:    decimal.NewFromInt(50),
				Price:       decimal.RequireFromString("46.00"),
				Amount:      decimal.RequireFromString("2319.95"),
				Fee:         decimal.RequireFromString("19.95"),
				ExternalRef: "cs_new_2",
			},
		}

		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)
		// Check duplicates
		mockRepo.EXPECT().FindExistingExternalRefs(ctx, portfolioID, []string{"cs_existing_1", "cs_new_2"}).
			Return([]string{"cs_existing_1"}, nil)

		// Item 2 instrument resolution
		mockRepo.EXPECT().FindInstrumentBySymbol(ctx, "BHP").Return(mockInst, nil)

		// Batch insert of item 2 only
		mockRepo.EXPECT().BatchInsertTransactions(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, txs []domain.Transaction) (int, error) {
				if len(txs) != 1 {
					t.Fatalf("expected 1 tx to insert, got %d", len(txs))
				}
				if *txs[0].ExternalRef != "cs_new_2" {
					t.Fatalf("expected cs_new_2, got %v", *txs[0].ExternalRef)
				}
				return 1, nil
			})

		// Rebuild projections calls
		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)
		mockRepo.EXPECT().GetTransactions(ctx, portfolioID).Return(nil, nil)
		mockRepo.EXPECT().GetCorporateActions(ctx, gomock.Any()).Return([]domain.CorporateAction{}, nil)
		mockRepo.EXPECT().SaveProjectionsTx(ctx, portfolioID, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

		// Valuation backfill called once with earliest trade date
		mockValuation.EXPECT().BackfillPortfolioValuations(ctx, portfolioID, tradeDate).Return(nil)

		// GetPortfolioSummary calls
		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)
		mockRepo.EXPECT().GetHoldingsWithMarketData(ctx, portfolioID).Return(nil, nil)
		mockRepo.EXPECT().GetCashBalances(ctx, portfolioID).Return(nil, nil)
		mockRepo.EXPECT().GetLatestValuation(ctx, portfolioID).Return(nil, nil)
		mockRepo.EXPECT().GetCashFXRates(ctx, "AUD").Return(map[string]decimal.Decimal{}, nil)

		res, err := svc.BatchImportTransactions(ctx, domain.BatchImportInput{
			UserID:         userID,
			Transactions:   items,
			SkipDuplicates: true,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.ImportedCount != 1 {
			t.Fatalf("expected 1 imported, got %d", res.ImportedCount)
		}
		if res.SkippedCount != 1 {
			t.Fatalf("expected 1 skipped, got %d", res.SkippedCount)
		}
	})

	t.Run("missing instrument auto-provisions successfully", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := repoMocks.NewMockRepository(ctrl)
		mockValuation := serviceMocks.NewMockValuationService(ctrl)
		svc := service.NewPortfolioService(mockRepo, service.WithValuationService(mockValuation))

		items := []domain.ImportTransactionItem{
			{
				Type:        domain.TxTypeBuy,
				Symbol:      "NEWASX",
				TradeDate:   tradeDate,
				Quantity:    decimal.NewFromInt(10),
				Price:       decimal.RequireFromString("10.00"),
				Amount:      decimal.RequireFromString("100.00"),
				ExternalRef: "ref_new",
			},
		}

		newInst := &domain.Instrument{
			ID:           uuid.New(),
			Symbol:       "NEWASX.AX",
			ExchangeCode: "XASX",
			CurrencyCode: "AUD",
		}

		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)
		mockRepo.EXPECT().FindExistingExternalRefs(ctx, portfolioID, []string{"ref_new"}).Return(nil, nil)
		mockRepo.EXPECT().FindInstrumentBySymbol(ctx, "NEWASX").Return(nil, repository.ErrInstrumentNotFound)
		mockRepo.EXPECT().CreateInstrument(ctx, gomock.Any()).Return(newInst, nil)
		mockRepo.EXPECT().BatchInsertTransactions(ctx, gomock.Len(1)).Return(1, nil)

		// Rebuild projections calls
		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)
		mockRepo.EXPECT().GetTransactions(ctx, portfolioID).Return(nil, nil)
		mockRepo.EXPECT().GetCorporateActions(ctx, gomock.Any()).Return([]domain.CorporateAction{}, nil)
		mockRepo.EXPECT().SaveProjectionsTx(ctx, portfolioID, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

		// Valuation backfill
		mockValuation.EXPECT().BackfillPortfolioValuations(ctx, portfolioID, tradeDate).Return(nil)

		// GetPortfolioSummary calls
		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)
		mockRepo.EXPECT().GetHoldingsWithMarketData(ctx, portfolioID).Return(nil, nil)
		mockRepo.EXPECT().GetCashBalances(ctx, portfolioID).Return(nil, nil)
		mockRepo.EXPECT().GetLatestValuation(ctx, portfolioID).Return(nil, nil)
		mockRepo.EXPECT().GetCashFXRates(ctx, "AUD").Return(map[string]decimal.Decimal{}, nil)

		res, err := svc.BatchImportTransactions(ctx, domain.BatchImportInput{
			UserID:         userID,
			Transactions:   items,
			SkipDuplicates: true,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.ImportedCount != 1 {
			t.Fatalf("expected 1 imported, got %d", res.ImportedCount)
		}
	})

	t.Run("missing instrument auto-provision failure returns error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := repoMocks.NewMockRepository(ctrl)
		svc := service.NewPortfolioService(mockRepo)

		items := []domain.ImportTransactionItem{
			{
				Type:        domain.TxTypeBuy,
				Symbol:      "UNKNOWN",
				TradeDate:   tradeDate,
				Quantity:    decimal.NewFromInt(10),
				Price:       decimal.RequireFromString("10.00"),
				Amount:      decimal.RequireFromString("100.00"),
				ExternalRef: "ref_unknown",
			},
		}

		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)
		mockRepo.EXPECT().FindExistingExternalRefs(ctx, portfolioID, []string{"ref_unknown"}).Return(nil, nil)
		mockRepo.EXPECT().FindInstrumentBySymbol(ctx, "UNKNOWN").Return(nil, repository.ErrInstrumentNotFound)
		mockRepo.EXPECT().CreateInstrument(ctx, gomock.Any()).Return(nil, errors.New("db error"))

		_, err := svc.BatchImportTransactions(ctx, domain.BatchImportInput{
			UserID:         userID,
			Transactions:   items,
			SkipDuplicates: true,
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

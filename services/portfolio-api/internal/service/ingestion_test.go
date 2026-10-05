package service_test

import (
	"context"
	"testing"
	"time"

	"portfolio-api/internal/domain"
	"portfolio-api/internal/marketdata"
	"portfolio-api/internal/repository/mocks"
	"portfolio-api/internal/service"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/mock/gomock"
)

func TestIngestionService_IngestDailyMarketData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	priceMock := marketdata.NewMockPriceProvider("MockPrice")
	fxMock := marketdata.NewMockFXRateProvider("MockFX")

	inst1ID := uuid.New()
	inst2ID := uuid.New()
	activeInsts := []domain.Instrument{
		{ID: inst1ID, Symbol: "AAPL", ExchangeCode: "NASDAQ", CurrencyCode: "USD"},
		{ID: inst2ID, Symbol: "MSFT", ExchangeCode: "NASDAQ", CurrencyCode: "USD"},
	}

	tradeDate, _ := time.Parse("2006-01-02", "2026-10-02")
	tradeDate = tradeDate.UTC()

	priceMock.SetPrice("AAPL", tradeDate, decimal.RequireFromString("227.63"))
	priceMock.SetPrice("MSFT", tradeDate, decimal.RequireFromString("432.25"))

	fxMock.SetRate("EUR", "USD", tradeDate, decimal.RequireFromString("1.0850"))

	mockRepo.EXPECT().ListActiveInstruments(gomock.Any()).Return(activeInsts, nil)
	mockRepo.EXPECT().BatchUpsertInstrumentPrices(gomock.Any(), gomock.Len(2)).Return(nil)

	mockRepo.EXPECT().ListActiveCurrencies(gomock.Any()).Return([]string{"USD", "EUR"}, nil)
	mockRepo.EXPECT().BatchUpsertFXRates(gomock.Any(), gomock.Any()).Return(nil)

	svc := service.NewIngestionService(mockRepo, priceMock, fxMock, nil)

	res, err := svc.IngestDailyMarketData(context.Background(), tradeDate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !res.Success {
		t.Errorf("expected success true, got false")
	}
	if res.PricesSynced != 2 {
		t.Errorf("expected 2 prices synced, got %d", res.PricesSynced)
	}
	if res.FXRatesSynced == 0 {
		t.Errorf("expected at least 1 FX rate synced, got 0")
	}
}

func TestIngestionService_BackfillInstrumentPrices(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	priceMock := marketdata.NewMockPriceProvider("MockPrice")

	instID := uuid.New()
	from, _ := time.Parse("2006-01-02", "2026-09-30")
	to, _ := time.Parse("2006-01-02", "2026-10-02")

	priceMock.SetPrice("AAPL", from, decimal.RequireFromString("225.00"))
	priceMock.SetPrice("AAPL", from.AddDate(0, 0, 1), decimal.RequireFromString("226.10"))
	priceMock.SetPrice("AAPL", to, decimal.RequireFromString("227.63"))

	mockRepo.EXPECT().BatchUpsertInstrumentPrices(gomock.Any(), gomock.Len(3)).Return(nil)

	svc := service.NewIngestionService(mockRepo, priceMock, nil, nil)

	count, err := svc.BackfillInstrumentPrices(context.Background(), instID, "AAPL", "NASDAQ", from, to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 backfilled records, got %d", count)
	}

	t.Run("returns error when to is before from", func(t *testing.T) {
		_, err := svc.BackfillInstrumentPrices(context.Background(), instID, "AAPL", "NASDAQ", to, from)
		if err == nil {
			t.Fatalf("expected error when to is before from, got nil")
		}
	})
}

func TestIngestionService_BackfillCurrencyPair(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	fxMock := marketdata.NewMockFXRateProvider("MockFX")

	from, _ := time.Parse("2006-01-02", "2026-10-01")
	to, _ := time.Parse("2006-01-02", "2026-10-02")

	fxMock.SetRate("EUR", "USD", from, decimal.RequireFromString("1.0820"))
	fxMock.SetRate("EUR", "USD", to, decimal.RequireFromString("1.0850"))

	mockRepo.EXPECT().BatchUpsertFXRates(gomock.Any(), gomock.Len(2)).Return(nil)

	svc := service.NewIngestionService(mockRepo, nil, fxMock, nil)

	count, err := svc.BackfillCurrencyPair(context.Background(), "EUR", "USD", from, to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 backfilled FX rates, got %d", count)
	}

	t.Run("returns 0 when base equals quote", func(t *testing.T) {
		count, err := svc.BackfillCurrencyPair(context.Background(), "USD", "USD", from, to)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if count != 0 {
			t.Errorf("expected 0 for identical currencies, got %d", count)
		}
	})
}

func TestPortfolioService_AddTransaction_TriggersIngestionHook(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	priceMock := marketdata.NewMockPriceProvider("MockPrice")

	instID := uuid.New()
	portfolioID := uuid.New()
	userUUID := uuid.New()
	userID := userUUID.String()
	sym := "AAPL"
	qty := decimal.NewFromInt(10)
	price := decimal.RequireFromString("150.00")
	tradeDate, _ := time.Parse("2006-01-02", "2026-09-30")
	tradeDate = tradeDate.UTC()

	mockPortfolio := &domain.Portfolio{
		ID:           portfolioID,
		UserID:       userUUID,
		BaseCurrency: "USD",
	}

	mockInst := &domain.Instrument{
		ID:           instID,
		Symbol:       sym,
		ExchangeCode: "NASDAQ",
		CurrencyCode: "USD",
	}

	// 1. Initial lookup in AddTransaction
	mockRepo.EXPECT().FindPortfolioByUser(gomock.Any(), userID).Return(mockPortfolio, nil)
	mockRepo.EXPECT().FindInstrumentBySymbol(gomock.Any(), sym).Return(mockInst, nil)

	// 2. Insert transaction
	mockRepo.EXPECT().InsertTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, tx domain.Transaction) (*domain.Transaction, error) {
			tx.ID = uuid.New()
			return &tx, nil
		},
	)

	// 3. HasPricesForRange check returns false (unpriced ticker for date range)
	mockRepo.EXPECT().HasPricesForRange(gomock.Any(), instID, gomock.Any(), gomock.Any()).Return(false, nil)

	// 4. Ingestion service backfill batch upsert
	priceMock.SetPrice(sym, tradeDate, decimal.RequireFromString("150.00"))
	mockRepo.EXPECT().BatchUpsertInstrumentPrices(gomock.Any(), gomock.Any()).Return(nil)

	// 5. Rebuild projections executes after backfill
	mockRepo.EXPECT().FindPortfolioByUser(gomock.Any(), userID).Return(mockPortfolio, nil)
	mockRepo.EXPECT().GetTransactions(gomock.Any(), portfolioID).Return([]domain.Transaction{
		{
			ID:           uuid.New(),
			PortfolioID:  portfolioID,
			InstrumentID: &instID,
			Type:         domain.TxTypeBuy,
			TradeDate:    tradeDate,
			Quantity:     &qty,
			Price:        &price,
			Amount:       qty.Mul(price),
			CurrencyCode: "USD",
		},
	}, nil)
	mockRepo.EXPECT().GetCorporateActions(gomock.Any(), gomock.Any()).Return(nil, nil)
	mockRepo.EXPECT().SaveProjectionsTx(gomock.Any(), portfolioID, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	// 6. Summary queries
	mockRepo.EXPECT().FindPortfolioByUser(gomock.Any(), userID).Return(mockPortfolio, nil)
	mockRepo.EXPECT().GetHoldingsWithMarketData(gomock.Any(), portfolioID).Return([]domain.HoldingWithPrice{
		{
			InstrumentID: instID,
			Ticker:       sym,
			Quantity:     qty,
			LatestPrice:  price,
		},
	}, nil)
	mockRepo.EXPECT().GetCashBalances(gomock.Any(), portfolioID).Return(nil, nil)
	mockRepo.EXPECT().GetCashFXRates(gomock.Any(), "USD").Return(map[string]decimal.Decimal{"USD": decimal.NewFromInt(1)}, nil)
	mockRepo.EXPECT().GetLatestValuation(gomock.Any(), portfolioID).Return(&domain.PortfolioValuation{
		MarketValueBase: qty.Mul(price),
	}, nil)

	ingestionSvc := service.NewIngestionService(mockRepo, priceMock, nil, nil)
	svc := service.NewPortfolioService(mockRepo, service.WithIngestionService(ingestionSvc))

	input := domain.AddTransactionInput{
		UserID:    userID,
		Symbol:    &sym,
		Type:      domain.TxTypeBuy,
		Quantity:  &qty,
		Price:     &price,
		TradeDate: tradeDate,
	}

	savedTx, summary, err := svc.AddTransaction(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error adding transaction: %v", err)
	}

	if savedTx == nil {
		t.Fatalf("expected non-nil saved transaction")
	}
	if summary == nil {
		t.Fatalf("expected non-nil summary")
	}
	if !summary.TotalValue.Amount.Equal(qty.Mul(price)) {
		t.Errorf("expected total value %s, got %s", qty.Mul(price), summary.TotalValue.Amount)
	}
}

package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"portfolio-api/internal/domain"
	repoMocks "portfolio-api/internal/repository/mocks"
	"portfolio-api/internal/service"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/mock/gomock"
)

func TestValuationService_SnapshotValuation(t *testing.T) {
	ctx := context.Background()
	portfolioID := uuid.New()
	userID := uuid.New()
	instID := uuid.New()

	basePortfolio := &domain.Portfolio{
		ID:              portfolioID,
		UserID:          userID,
		Name:            "Main Growth",
		BaseCurrency:    "USD",
		CostBasisMethod: domain.CostBasisMethodFIFO,
	}

	t.Run("calculates and saves valuation snapshot for current day with holdings and cash", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockRepo := repoMocks.NewMockRepository(ctrl)

		asOfDate := time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC)
		asOfStr := "2026-10-05"

		tx1 := domain.Transaction{
			ID:           uuid.New(),
			PortfolioID:  portfolioID,
			Type:         domain.TxTypeDeposit,
			TradeDate:    asOfDate.AddDate(0, 0, -2), // 2026-10-03
			Amount:       decimal.RequireFromString("10000.00"),
			CurrencyCode: "USD",
		}
		buyQty := decimal.RequireFromString("10")
		buyPrice := decimal.RequireFromString("150.00")
		tx2 := domain.Transaction{
			ID:           uuid.New(),
			PortfolioID:  portfolioID,
			InstrumentID: &instID,
			Type:         domain.TxTypeBuy,
			TradeDate:    asOfDate.AddDate(0, 0, -2), // 2026-10-03
			Quantity:     &buyQty,
			Price:        &buyPrice,
			Amount:       decimal.RequireFromString("1500.00"),
			CurrencyCode: "USD",
			Fee:          decimal.RequireFromString("5.00"),
		}

		mockRepo.EXPECT().FindPortfolioByUser(ctx, portfolioID.String()).Return(basePortfolio, nil)
		mockRepo.EXPECT().GetTransactions(ctx, portfolioID).Return([]domain.Transaction{tx1, tx2}, nil)
		mockRepo.EXPECT().GetCorporateActions(ctx, gomock.Any()).Return([]domain.CorporateAction{}, nil)

		priceMap := map[uuid.UUID]map[string]decimal.Decimal{
			instID: {
				asOfStr: decimal.RequireFromString("160.00"),
			},
		}
		mockRepo.EXPECT().GetHistoricalPriceMatrix(ctx, gomock.Any(), asOfDate, asOfDate).Return(priceMap, nil)

		fxMap := map[string]map[string]decimal.Decimal{
			"USD": {
				asOfStr: decimal.NewFromInt(1),
			},
		}
		mockRepo.EXPECT().GetHistoricalFXMatrix(ctx, []string{"USD"}, "USD", asOfDate, asOfDate).Return(fxMap, nil)

		prevDate := asOfDate.AddDate(0, 0, -1)
		prevVal := &domain.PortfolioValuation{
			PortfolioID:     portfolioID,
			ValuationDate:   prevDate,
			MarketValueBase: decimal.RequireFromString("1550.00"),
			CashValueBase:   decimal.RequireFromString("8495.00"),
			NetFlowBase:     decimal.Zero,
			DailyReturn:     decimal.RequireFromString("0.0045"),
			TWRIndex:        decimal.RequireFromString("1.004500000000"),
		}
		mockRepo.EXPECT().GetLatestValuationBefore(ctx, portfolioID, asOfDate).Return(prevVal, nil)

		mockRepo.EXPECT().UpsertValuationsBatch(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, batch []domain.PortfolioValuation) error {
			if len(batch) != 1 {
				t.Fatalf("expected 1 valuation in batch, got %d", len(batch))
			}
			v := batch[0]
			// Market value: 10 * 160 = 1600.00
			if !v.MarketValueBase.Equal(decimal.RequireFromString("1600.00")) {
				t.Errorf("got MarketValueBase %s, want 1600.00", v.MarketValueBase)
			}
			// Cash: 10000 - 1500 - 5 = 8495.00
			if !v.CashValueBase.Equal(decimal.RequireFromString("8495.00")) {
				t.Errorf("got CashValueBase %s, want 8495.00", v.CashValueBase)
			}
			if v.DailyReturn.IsZero() {
				t.Errorf("expected non-zero daily return")
			}
			if !v.TWRIndex.IsPositive() {
				t.Errorf("expected positive TWR index")
			}
			return nil
		})

		svc := service.NewValuationService(mockRepo)
		snap, err := svc.SnapshotValuation(ctx, portfolioID, asOfDate)
		if err != nil {
			t.Fatalf("SnapshotValuation failed: %v", err)
		}
		if snap == nil {
			t.Fatalf("expected non-nil snapshot")
		}
		if !snap.MarketValueBase.Equal(decimal.RequireFromString("1600.00")) {
			t.Errorf("got snapshot market value %s, want 1600.00", snap.MarketValueBase)
		}
	})

	t.Run("calculates initial funding deposit snapshot with 0 daily return and 1.0 baseline TWR", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockRepo := repoMocks.NewMockRepository(ctrl)

		asOfDate := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
		asOfStr := "2026-10-01"

		depositTx := domain.Transaction{
			ID:           uuid.New(),
			PortfolioID:  portfolioID,
			Type:         domain.TxTypeDeposit,
			TradeDate:    asOfDate,
			Amount:       decimal.RequireFromString("25000.00"),
			CurrencyCode: "USD",
		}

		mockRepo.EXPECT().FindPortfolioByUser(ctx, portfolioID.String()).Return(basePortfolio, nil)
		mockRepo.EXPECT().GetTransactions(ctx, portfolioID).Return([]domain.Transaction{depositTx}, nil)
		mockRepo.EXPECT().GetCorporateActions(ctx, gomock.Any()).Return([]domain.CorporateAction{}, nil)
		mockRepo.EXPECT().GetHistoricalPriceMatrix(ctx, gomock.Any(), asOfDate, asOfDate).Return(map[uuid.UUID]map[string]decimal.Decimal{}, nil)

		fxMap := map[string]map[string]decimal.Decimal{
			"USD": {
				asOfStr: decimal.NewFromInt(1),
			},
		}
		mockRepo.EXPECT().GetHistoricalFXMatrix(ctx, []string{"USD"}, "USD", asOfDate, asOfDate).Return(fxMap, nil)
		mockRepo.EXPECT().GetLatestValuationBefore(ctx, portfolioID, asOfDate).Return(nil, nil)

		mockRepo.EXPECT().UpsertValuationsBatch(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, batch []domain.PortfolioValuation) error {
			if len(batch) != 1 {
				t.Fatalf("expected 1 valuation, got %d", len(batch))
			}
			v := batch[0]
			if !v.CashValueBase.Equal(decimal.RequireFromString("25000.00")) {
				t.Errorf("got cash value %s, want 25000.00", v.CashValueBase)
			}
			if !v.NetFlowBase.Equal(decimal.RequireFromString("25000.00")) {
				t.Errorf("got net flow %s, want 25000.00", v.NetFlowBase)
			}
			if !v.DailyReturn.IsZero() {
				t.Errorf("got daily return %s, want 0", v.DailyReturn)
			}
			if !v.TWRIndex.Equal(decimal.RequireFromString("1.000000000000")) {
				t.Errorf("got TWR index %s, want 1.000000000000", v.TWRIndex)
			}
			return nil
		})

		svc := service.NewValuationService(mockRepo)
		snap, err := svc.SnapshotValuation(ctx, portfolioID, asOfDate)
		if err != nil {
			t.Fatalf("SnapshotValuation failed: %v", err)
		}
		if snap == nil || !snap.CashValueBase.Equal(decimal.RequireFromString("25000.00")) {
			t.Errorf("unexpected snapshot result: %v", snap)
		}
	})

	t.Run("handles portfolio with zero transactions gracefully", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockRepo := repoMocks.NewMockRepository(ctrl)

		asOfDate := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)

		mockRepo.EXPECT().FindPortfolioByUser(ctx, portfolioID.String()).Return(basePortfolio, nil)
		mockRepo.EXPECT().GetTransactions(ctx, portfolioID).Return([]domain.Transaction{}, nil)
		mockRepo.EXPECT().GetLatestValuationBefore(ctx, portfolioID, asOfDate).Return(nil, nil)
		mockRepo.EXPECT().UpsertValuationsBatch(ctx, gomock.Any()).Return(nil)

		svc := service.NewValuationService(mockRepo)
		snap, err := svc.SnapshotValuation(ctx, portfolioID, asOfDate)
		if err != nil {
			t.Fatalf("SnapshotValuation failed: %v", err)
		}
		if !snap.TotalValue().IsZero() {
			t.Errorf("expected zero total value for empty portfolio, got %s", snap.TotalValue())
		}
	})
}

func TestValuationService_BackfillPortfolioValuations(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repoMocks.NewMockRepository(ctrl)
	portfolioID := uuid.New()
	userID := uuid.New()
	instID := uuid.New()

	basePortfolio := &domain.Portfolio{
		ID:              portfolioID,
		UserID:          userID,
		Name:            "Tech Portfolio",
		BaseCurrency:    "USD",
		CostBasisMethod: domain.CostBasisMethodFIFO,
	}

	d1 := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC) // Day 1: Deposit $10,000, Buy 20 AAPL @ $150 ($3000)
	d3 := time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC) // Day 3: AAPL dividend $50, price $157.50
	d4 := time.Date(2026, 10, 4, 0, 0, 0, 0, time.UTC) // Day 4: Sell 10 AAPL @ $170
	d5 := time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC) // Day 5: Withdraw $2,000

	qty20 := decimal.RequireFromString("20")
	price150 := decimal.RequireFromString("150.00")

	divAmount := decimal.RequireFromString("50.00")

	qty10 := decimal.RequireFromString("10")
	price170 := decimal.RequireFromString("170.00")

	txs := []domain.Transaction{
		{
			ID:           uuid.New(),
			PortfolioID:  portfolioID,
			Type:         domain.TxTypeDeposit,
			TradeDate:    d1,
			Amount:       decimal.RequireFromString("10000.00"),
			CurrencyCode: "USD",
		},
		{
			ID:           uuid.New(),
			PortfolioID:  portfolioID,
			InstrumentID: &instID,
			Type:         domain.TxTypeBuy,
			TradeDate:    d1,
			Quantity:     &qty20,
			Price:        &price150,
			Amount:       decimal.RequireFromString("3000.00"),
			CurrencyCode: "USD",
		},
		{
			ID:           uuid.New(),
			PortfolioID:  portfolioID,
			InstrumentID: &instID,
			Type:         domain.TxTypeDividend,
			TradeDate:    d3,
			Amount:       divAmount,
			CurrencyCode: "USD",
		},
		{
			ID:           uuid.New(),
			PortfolioID:  portfolioID,
			InstrumentID: &instID,
			Type:         domain.TxTypeSell,
			TradeDate:    d4,
			Quantity:     &qty10,
			Price:        &price170,
			Amount:       decimal.RequireFromString("1700.00"),
			CurrencyCode: "USD",
		},
		{
			ID:           uuid.New(),
			PortfolioID:  portfolioID,
			Type:         domain.TxTypeWithdrawal,
			TradeDate:    d5,
			Amount:       decimal.RequireFromString("2000.00"),
			CurrencyCode: "USD",
		},
	}

	priceMatrix := map[uuid.UUID]map[string]decimal.Decimal{
		instID: {
			"2026-10-01": decimal.RequireFromString("150.00"),
			"2026-10-02": decimal.RequireFromString("160.00"),
			"2026-10-03": decimal.RequireFromString("157.50"),
			"2026-10-04": decimal.RequireFromString("170.00"),
			"2026-10-05": decimal.RequireFromString("170.00"),
		},
	}

	fxMatrix := map[string]map[string]decimal.Decimal{
		"USD": {
			"2026-10-01": decimal.NewFromInt(1),
			"2026-10-02": decimal.NewFromInt(1),
			"2026-10-03": decimal.NewFromInt(1),
			"2026-10-04": decimal.NewFromInt(1),
			"2026-10-05": decimal.NewFromInt(1),
		},
	}

	mockRepo.EXPECT().FindPortfolioByUser(ctx, portfolioID.String()).Return(basePortfolio, nil)
	mockRepo.EXPECT().GetTransactions(ctx, portfolioID).Return(txs, nil)
	mockRepo.EXPECT().DeleteValuationsFromDate(ctx, portfolioID, d1).Return(nil)
	mockRepo.EXPECT().GetLatestValuationBefore(ctx, portfolioID, d1).Return(nil, nil)
	mockRepo.EXPECT().GetCorporateActions(ctx, []uuid.UUID{instID}).Return([]domain.CorporateAction{}, nil)
	mockRepo.EXPECT().GetHistoricalPriceMatrix(ctx, []uuid.UUID{instID}, d1, d5).Return(priceMatrix, nil)
	mockRepo.EXPECT().GetHistoricalFXMatrix(ctx, []string{"USD"}, "USD", d1, d5).Return(fxMatrix, nil)

	mockRepo.EXPECT().UpsertValuationsBatch(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, batch []domain.PortfolioValuation) error {
		if len(batch) != 5 {
			t.Fatalf("expected 5 daily valuations in replay batch, got %d", len(batch))
		}

		// Day 1 (2026-10-01):
		// Cash = 10,000 - 3,000 = 7,000. Market = 20 * 150 = 3,000. Total = 10,000. Net flow = 10,000.
		// Return = 0%. TWR = 1.000000000000
		v1 := batch[0]
		if !v1.MarketValueBase.Equal(decimal.RequireFromString("3000.00")) {
			t.Errorf("day 1: got market value %s, want 3000.00", v1.MarketValueBase)
		}
		if !v1.CashValueBase.Equal(decimal.RequireFromString("7000.00")) {
			t.Errorf("day 1: got cash value %s, want 7000.00", v1.CashValueBase)
		}
		if !v1.DailyReturn.IsZero() {
			t.Errorf("day 1: got daily return %s, want 0", v1.DailyReturn)
		}
		if !v1.TWRIndex.Equal(decimal.RequireFromString("1.000000000000")) {
			t.Errorf("day 1: got TWR %s, want 1.000000000000", v1.TWRIndex)
		}

		// Day 2 (2026-10-02):
		// AAPL price rises to 160 -> Market = 20 * 160 = 3200. Cash = 7000. Total = 10200.
		// Net flow = 0. Daily return = (10200 - 10000) / 10000 = +0.02 (2.0%). TWR = 1.020000000000
		v2 := batch[1]
		if !v2.MarketValueBase.Equal(decimal.RequireFromString("3200.00")) {
			t.Errorf("day 2: got market value %s, want 3200.00", v2.MarketValueBase)
		}
		expectedRet2 := decimal.RequireFromString("0.020000000000")
		if !v2.DailyReturn.Equal(expectedRet2) {
			t.Errorf("day 2: got daily return %s, want %s", v2.DailyReturn, expectedRet2)
		}
		expectedTWR2 := decimal.RequireFromString("1.020000000000")
		if !v2.TWRIndex.Equal(expectedTWR2) {
			t.Errorf("day 2: got TWR %s, want %s", v2.TWRIndex, expectedTWR2)
		}

		// Day 3 (2026-10-03):
		// Dividend $50 added to cash -> Cash = 7050. Price drops to 157.50 -> Market = 20 * 157.50 = 3150.
		// Total = 3150 + 7050 = 10200. Net flow = 0 (dividend is internal return).
		// Daily return = (10200 - 10200) / 10200 = 0%. TWR = 1.020000000000
		v3 := batch[2]
		if !v3.CashValueBase.Equal(decimal.RequireFromString("7050.00")) {
			t.Errorf("day 3: got cash %s, want 7050.00", v3.CashValueBase)
		}
		if !v3.MarketValueBase.Equal(decimal.RequireFromString("3150.00")) {
			t.Errorf("day 3: got market value %s, want 3150.00", v3.MarketValueBase)
		}
		if !v3.DailyReturn.IsZero() {
			t.Errorf("day 3: got daily return %s, want 0", v3.DailyReturn)
		}
		if !v3.TWRIndex.Equal(expectedTWR2) {
			t.Errorf("day 3: got TWR %s, want %s", v3.TWRIndex, expectedTWR2)
		}

		// Day 4 (2026-10-04):
		// Sold 10 AAPL @ 170 -> Cash = 7050 + 1700 = 8750. Holdings = 10 AAPL @ 170 = 1700.
		// Total = 8750 + 1700 = 10450. Net flow = 0. Prior total = 10200.
		// Net gain = 10450 - 10200 = 250. Daily return = 250 / 10200 = 0.024509803922
		v4 := batch[3]
		if !v4.CashValueBase.Equal(decimal.RequireFromString("8750.00")) {
			t.Errorf("day 4: got cash %s, want 8750.00", v4.CashValueBase)
		}
		if !v4.MarketValueBase.Equal(decimal.RequireFromString("1700.00")) {
			t.Errorf("day 4: got market value %s, want 1700.00", v4.MarketValueBase)
		}

		// Day 5 (2026-10-05):
		// Withdrew $2,000 -> Cash = 8750 - 2000 = 6750. Market = 1700. Total = 8450.
		// Net flow = -$2000. Prior total = 10450.
		// (8450 - (-2000)) - 10450 = 10450 - 10450 = 0. Daily return = 0!
		// TWR must remain unchanged from day 4!
		v5 := batch[4]
		if !v5.NetFlowBase.Equal(decimal.RequireFromString("-2000.00")) {
			t.Errorf("day 5: got net flow %s, want -2000.00", v5.NetFlowBase)
		}
		if !v5.CashValueBase.Equal(decimal.RequireFromString("6750.00")) {
			t.Errorf("day 5: got cash %s, want 6750.00", v5.CashValueBase)
		}
		if !v5.DailyReturn.IsZero() {
			t.Errorf("day 5: got daily return %s, want 0", v5.DailyReturn)
		}
		if !v5.TWRIndex.Equal(v4.TWRIndex) {
			t.Errorf("day 5: TWR %s should equal day 4 TWR %s", v5.TWRIndex, v4.TWRIndex)
		}

		return nil
	})

	svc := service.NewValuationService(mockRepo, service.WithValuationNowFunc(func() time.Time { return d5 }))
	err := svc.BackfillPortfolioValuations(ctx, portfolioID, d1)
	if err != nil {
		t.Fatalf("BackfillPortfolioValuations failed: %v", err)
	}
}

func TestValuationService_RunDailyValuationJob(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repoMocks.NewMockRepository(ctrl)
	pID1 := uuid.New()
	pID2 := uuid.New()
	asOfDate := time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC)

	p1 := &domain.Portfolio{ID: pID1, BaseCurrency: "USD"}
	p2 := &domain.Portfolio{ID: pID2, BaseCurrency: "USD"}

	t.Run("successfully runs snapshot across multiple portfolios", func(t *testing.T) {
		mockRepo.EXPECT().ListActivePortfolios(ctx).Return([]uuid.UUID{pID1, pID2}, nil)

		// Portfolio 1 expectations (empty)
		mockRepo.EXPECT().FindPortfolioByUser(ctx, pID1.String()).Return(p1, nil)
		mockRepo.EXPECT().GetTransactions(ctx, pID1).Return([]domain.Transaction{}, nil)
		mockRepo.EXPECT().GetLatestValuationBefore(ctx, pID1, asOfDate).Return(nil, nil)
		mockRepo.EXPECT().UpsertValuationsBatch(ctx, gomock.Any()).Return(nil)

		// Portfolio 2 expectations (empty)
		mockRepo.EXPECT().FindPortfolioByUser(ctx, pID2.String()).Return(p2, nil)
		mockRepo.EXPECT().GetTransactions(ctx, pID2).Return([]domain.Transaction{}, nil)
		mockRepo.EXPECT().GetLatestValuationBefore(ctx, pID2, asOfDate).Return(nil, nil)
		mockRepo.EXPECT().UpsertValuationsBatch(ctx, gomock.Any()).Return(nil)

		svc := service.NewValuationService(mockRepo)
		err := svc.RunDailyValuationJob(ctx, asOfDate)
		if err != nil {
			t.Fatalf("RunDailyValuationJob failed: %v", err)
		}
	})

	t.Run("accumulates errors if one portfolio fails and reports back", func(t *testing.T) {
		mockRepo.EXPECT().ListActivePortfolios(ctx).Return([]uuid.UUID{pID1, pID2}, nil)

		// Portfolio 1 fails
		mockRepo.EXPECT().FindPortfolioByUser(ctx, pID1.String()).Return(nil, errors.New("db disconnect"))

		// Portfolio 2 proceeds
		mockRepo.EXPECT().FindPortfolioByUser(ctx, pID2.String()).Return(p2, nil)
		mockRepo.EXPECT().GetTransactions(ctx, pID2).Return([]domain.Transaction{}, nil)
		mockRepo.EXPECT().GetLatestValuationBefore(ctx, pID2, asOfDate).Return(nil, nil)
		mockRepo.EXPECT().UpsertValuationsBatch(ctx, gomock.Any()).Return(nil)

		svc := service.NewValuationService(mockRepo)
		err := svc.RunDailyValuationJob(ctx, asOfDate)
		if err == nil {
			t.Fatalf("expected error from RunDailyValuationJob, got nil")
		}
	})
}

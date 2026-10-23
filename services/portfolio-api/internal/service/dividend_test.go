package service_test

import (
	"context"
	"testing"
	"time"

	"portfolio-api/internal/domain"
	"portfolio-api/internal/repository/mocks"
	"portfolio-api/internal/service"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/mock/gomock"
)

func TestPortfolioService_GetDividendCalendar(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)

	portfolioID := uuid.New()
	userUUID := uuid.New()
	userID := userUUID.String()

	mockPortfolio := &domain.Portfolio{
		ID:           portfolioID,
		UserID:       userUUID,
		BaseCurrency: "USD",
	}

	instID := uuid.New()
	now := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	svc := service.NewPortfolioService(mockRepo, service.WithNowFunc(func() time.Time { return now }))

	t.Run("invalid month returns error", func(t *testing.T) {
		mockRepo.EXPECT().FindPortfolioByUser(gomock.Any(), userID).Return(mockPortfolio, nil)
		_, err := svc.GetDividendCalendar(context.Background(), userID, 2026, 13)
		if err == nil {
			t.Fatalf("expected error for month 13, got nil")
		}
	})

	t.Run("empty holdings returns empty calendar", func(t *testing.T) {
		mockRepo.EXPECT().FindPortfolioByUser(gomock.Any(), userID).Return(mockPortfolio, nil)
		mockRepo.EXPECT().GetHoldingsWithMarketData(gomock.Any(), portfolioID).Return(nil, nil)

		events, err := svc.GetDividendCalendar(context.Background(), userID, 2026, 11)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(events) != 0 {
			t.Errorf("expected 0 events, got %d", len(events))
		}
	})

	t.Run("success with schedules and entitlement", func(t *testing.T) {
		mockRepo.EXPECT().FindPortfolioByUser(gomock.Any(), userID).Return(mockPortfolio, nil)
		holdings := []domain.HoldingWithPrice{
			{
				InstrumentID:       instID,
				Ticker:             "AAPL",
				Name:               "Apple Inc.",
				InstrumentCurrency: "USD",
				Quantity:           decimal.RequireFromString("50"),
				CostBasis:          decimal.RequireFromString("6000"),
				LatestPrice:        decimal.RequireFromString("150"),
			},
		}
		mockRepo.EXPECT().GetHoldingsWithMarketData(gomock.Any(), portfolioID).Return(holdings, nil)

		exDate := time.Date(2026, 11, 10, 0, 0, 0, 0, time.UTC)
		payDate := time.Date(2026, 11, 25, 0, 0, 0, 0, time.UTC)
		schedules := []domain.DividendSchedule{
			{
				ID:             uuid.New(),
				InstrumentID:   instID,
				AmountPerShare: decimal.RequireFromString("0.50"),
				CurrencyCode:   "USD",
				Frequency:      domain.DividendFreqQuarterly,
				Status:         domain.DividendStatusAnnounced,
				ExDate:         exDate,
				PaymentDate:    payDate,
			},
		}
		mockRepo.EXPECT().GetDividendCalendarSchedules(gomock.Any(), []uuid.UUID{instID}, gomock.Any(), gomock.Any()).Return(schedules, nil)
		mockRepo.EXPECT().GetCashFXRates(gomock.Any(), "USD").Return(map[string]decimal.Decimal{"USD": decimal.NewFromInt(1)}, nil)
		mockRepo.EXPECT().GetEntitledSharesForExDate(gomock.Any(), portfolioID, instID, exDate).Return(decimal.RequireFromString("50"), nil)

		events, err := svc.GetDividendCalendar(context.Background(), userID, 2026, 11)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(events) != 1 {
			t.Fatalf("expected 1 calendar event, got %d", len(events))
		}
		ev := events[0]
		if ev.Symbol != "AAPL" {
			t.Errorf("expected symbol AAPL, got %s", ev.Symbol)
		}
		if !ev.IsUserEntitled {
			t.Errorf("expected is_user_entitled = true")
		}
		// 50 shares * 0.50 DPS = 25 USD
		expectedGross := decimal.RequireFromString("25")
		if !ev.GrossAmountBase.Amount.Equal(expectedGross) {
			t.Errorf("expected gross amount %s, got %s", expectedGross, ev.GrossAmountBase.Amount)
		}
	})
}

func TestPortfolioService_GetDividendYieldAnalytics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)

	portfolioID := uuid.New()
	userUUID := uuid.New()
	userID := userUUID.String()

	mockPortfolio := &domain.Portfolio{
		ID:           portfolioID,
		UserID:       userUUID,
		BaseCurrency: "USD",
	}

	instID := uuid.New()
	now := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	svc := service.NewPortfolioService(mockRepo, service.WithNowFunc(func() time.Time { return now }))

	mockRepo.EXPECT().FindPortfolioByUser(gomock.Any(), userID).Return(mockPortfolio, nil)
	holdings := []domain.HoldingWithPrice{
		{
			InstrumentID:       instID,
			Ticker:             "AAPL",
			Name:               "Apple Inc.",
			InstrumentCurrency: "USD",
			Quantity:           decimal.RequireFromString("100"),
			CostBasis:          decimal.RequireFromString("10000"), // avg buy: 100
			CostBasisBase:      decimal.RequireFromString("10000"),
			LatestPrice:        decimal.RequireFromString("200"),   // market value: 20000
		},
	}
	mockRepo.EXPECT().GetHoldingsWithMarketData(gomock.Any(), portfolioID).Return(holdings, nil)

	exDate := time.Date(2026, 11, 15, 0, 0, 0, 0, time.UTC)
	payDate := time.Date(2026, 11, 30, 0, 0, 0, 0, time.UTC)
	schedules := []domain.DividendSchedule{
		{
			ID:             uuid.New(),
			InstrumentID:   instID,
			AmountPerShare: decimal.RequireFromString("0.50"), // 0.50 * 4 = 2.00 annual DPS
			CurrencyCode:   "USD",
			Frequency:      domain.DividendFreqQuarterly,
			ExDate:         exDate,
			PaymentDate:    payDate,
		},
	}
	mockRepo.EXPECT().GetUpcomingDividendSchedules(gomock.Any(), []uuid.UUID{instID}, gomock.Any(), gomock.Any()).Return(schedules, nil)
	mockRepo.EXPECT().ListDividendSchedulesByInstrument(gomock.Any(), instID).Return(schedules, nil)
	mockRepo.EXPECT().GetCashFXRates(gomock.Any(), "USD").Return(map[string]decimal.Decimal{"USD": decimal.NewFromInt(1)}, nil)

	analytics, err := svc.GetDividendYieldAnalytics(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if analytics == nil {
		t.Fatalf("expected non-nil analytics")
	}

	// 100 shares * 2.00 annual DPS = 200 USD
	expectedAnnual := decimal.RequireFromString("200")
	if !analytics.AnnualForwardIncome.Amount.Equal(expectedAnnual) {
		t.Errorf("expected annual forward income %s, got %s", expectedAnnual, analytics.AnnualForwardIncome.Amount)
	}

	// Portfolio Current Yield: 200 / 20000 * 100 = 1.0%
	expectedYield := decimal.RequireFromString("1")
	if !analytics.PortfolioCurrentYield.Equal(expectedYield) {
		t.Errorf("expected current yield %s, got %s", expectedYield, analytics.PortfolioCurrentYield)
	}

	// YOC: 200 / 10000 * 100 = 2.0%
	expectedYOC := decimal.RequireFromString("2")
	if analytics.PortfolioYieldOnCost == nil || !analytics.PortfolioYieldOnCost.Equal(expectedYOC) {
		t.Errorf("expected portfolio YOC %s, got %v", expectedYOC, analytics.PortfolioYieldOnCost)
	}
}

func TestPortfolioService_SimulateDRIP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)

	portfolioID := uuid.New()
	userUUID := uuid.New()
	userID := userUUID.String()

	mockPortfolio := &domain.Portfolio{
		ID:           portfolioID,
		UserID:       userUUID,
		BaseCurrency: "USD",
	}

	now := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	svc := service.NewPortfolioService(mockRepo, service.WithNowFunc(func() time.Time { return now }))

	mockRepo.EXPECT().FindPortfolioByUser(gomock.Any(), userID).Return(mockPortfolio, nil).Times(2)
	mockRepo.EXPECT().GetHoldingsWithMarketData(gomock.Any(), portfolioID).Return(nil, nil)
	mockVal := &domain.PortfolioValuation{
		MarketValueBase: decimal.RequireFromString("100000"),
	}
	mockRepo.EXPECT().GetLatestValuation(gomock.Any(), portfolioID).Return(mockVal, nil)

	input := domain.DRIPSimulationInput{
		HorizonYears:           5,
		DividendGrowthRatePct:  decimal.RequireFromString("5.0"),
		CapitalAppreciationPct: decimal.RequireFromString("7.0"),
	}

	res, err := svc.SimulateDRIP(context.Background(), userID, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res == nil {
		t.Fatalf("expected non-nil result")
	}
	if len(res.Points) != 5 {
		t.Errorf("expected 5 yearly points, got %d", len(res.Points))
	}
}

func TestPortfolioService_SimulateDRIP_UsesHoldingsMarketValue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)

	portfolioID := uuid.New()
	userUUID := uuid.New()
	userID := userUUID.String()

	mockPortfolio := &domain.Portfolio{
		ID:           portfolioID,
		UserID:       userUUID,
		BaseCurrency: "USD",
	}

	inst1 := uuid.New()
	inst2 := uuid.New()
	holdings := []domain.HoldingWithPrice{
		{
			InstrumentID:       inst1,
			Ticker:             "AAPL",
			Name:               "Apple Inc.",
			InstrumentCurrency: "USD",
			Quantity:           decimal.RequireFromString("10"),
			LatestPrice:        decimal.RequireFromString("150.00"),
			FXRateToBase:       decimal.NewFromInt(1),
		},
		{
			InstrumentID:       inst2,
			Ticker:             "BHP.AX",
			Name:               "BHP Group",
			InstrumentCurrency: "AUD",
			Quantity:           decimal.RequireFromString("100"),
			LatestPrice:        decimal.RequireFromString("30.00"),
			FXRateToBase:       decimal.RequireFromString("0.65"), // 3000 AUD * 0.65 = 1950 USD
		},
	}

	now := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	svc := service.NewPortfolioService(mockRepo, service.WithNowFunc(func() time.Time { return now }))

	mockRepo.EXPECT().FindPortfolioByUser(gomock.Any(), userID).Return(mockPortfolio, nil).Times(2)
	mockRepo.EXPECT().GetHoldingsWithMarketData(gomock.Any(), portfolioID).Return(holdings, nil)
	mockRepo.EXPECT().GetUpcomingDividendSchedules(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)
	mockRepo.EXPECT().ListDividendSchedulesByInstrument(gomock.Any(), inst1).Return(nil, nil).AnyTimes()
	mockRepo.EXPECT().ListDividendSchedulesByInstrument(gomock.Any(), inst2).Return(nil, nil).AnyTimes()
	mockRepo.EXPECT().GetCashFXRates(gomock.Any(), "USD").Return(nil, nil)

	input := domain.DRIPSimulationInput{
		HorizonYears:           5,
		DividendGrowthRatePct:  decimal.RequireFromString("5.0"),
		CapitalAppreciationPct: decimal.RequireFromString("7.0"),
	}

	res, err := svc.SimulateDRIP(context.Background(), userID, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res == nil {
		t.Fatalf("expected non-nil result")
	}

	// Expected invested assets value: (10 * 150) + (100 * 30 * 0.65) = 1500 + 1950 = 3450 USD
	expectedStartVal := decimal.RequireFromString("3450.00")
	if !res.StartingPortfolioValue.Equal(expectedStartVal) {
		t.Errorf("expected StartingPortfolioValue to be %s (live invested assets), got %s", expectedStartVal, res.StartingPortfolioValue)
	}
}

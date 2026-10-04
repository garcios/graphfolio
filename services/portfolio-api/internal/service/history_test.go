package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"portfolio-api/internal/domain"
	mocks "portfolio-api/internal/repository/mocks"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/mock/gomock"
)

func TestPortfolioService_GetPortfolioHistory(t *testing.T) {
	ctx := context.Background()
	fixedNow := time.Date(2026, 5, 15, 14, 30, 0, 0, time.UTC)
	today := time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)
	userID := "user-456"
	portfolioID := uuid.New()

	mockPortfolio := &domain.Portfolio{
		ID:              portfolioID,
		UserID:          uuid.New(),
		Name:            "Tech Growth",
		BaseCurrency:    "USD",
		CostBasisMethod: domain.CostBasisMethodAverageCost,
		CreatedAt:       fixedNow.AddDate(-2, 0, 0),
	}

	timeframeTestCases := []struct {
		name             string
		timeframe        domain.HistoryTimeframe
		expectedFromDate time.Time
	}{
		{
			name:             "1D timeframe filters by 1 day ago",
			timeframe:        domain.Timeframe1D,
			expectedFromDate: today.AddDate(0, 0, -1),
		},
		{
			name:             "1W timeframe filters by 7 days ago",
			timeframe:        domain.Timeframe1W,
			expectedFromDate: today.AddDate(0, 0, -7),
		},
		{
			name:             "1M timeframe filters by 1 month ago",
			timeframe:        domain.Timeframe1M,
			expectedFromDate: today.AddDate(0, -1, 0),
		},
		{
			name:             "1Y timeframe filters by 1 year ago",
			timeframe:        domain.Timeframe1Y,
			expectedFromDate: today.AddDate(-1, 0, 0),
		},
		{
			name:             "ALL timeframe uses zero time inception",
			timeframe:        domain.TimeframeAll,
			expectedFromDate: time.Time{},
		},
	}

	for _, tc := range timeframeTestCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockRepo := mocks.NewMockRepository(ctrl)

			svc := &portfolioService{
				repo:    mockRepo,
				nowFunc: func() time.Time { return fixedNow },
			}

			mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)
			mockRepo.EXPECT().GetPortfolioValuations(ctx, portfolioID, tc.expectedFromDate).Return([]domain.PortfolioValuation{
				{
					PortfolioID:     portfolioID,
					ValuationDate:   tc.expectedFromDate,
					MarketValueBase: decimal.RequireFromString("100000.00"),
					CashValueBase:   decimal.RequireFromString("5000.00"),
					DailyReturn:     decimal.Zero,
					TWRIndex:        decimal.RequireFromString("1.0000"),
				},
				{
					PortfolioID:     portfolioID,
					ValuationDate:   today,
					MarketValueBase: decimal.RequireFromString("115000.00"),
					CashValueBase:   decimal.RequireFromString("5000.00"),
					DailyReturn:     decimal.RequireFromString("0.0200"),
					TWRIndex:        decimal.RequireFromString("1.1428"),
				},
			}, nil)

			res, err := svc.GetPortfolioHistory(ctx, userID, tc.timeframe)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if res == nil {
				t.Fatalf("expected non-nil response")
			}
			if len(res.Points) != 2 {
				t.Fatalf("expected 2 valuation points, got %d", len(res.Points))
			}

			// Start: 105,000.00, End: 120,000.00 -> Return: 15,000.00
			expectedStart := decimal.RequireFromString("105000.00")
			expectedEnd := decimal.RequireFromString("120000.00")
			expectedReturn := decimal.RequireFromString("15000.00")
			// 15,000 / 105,000 = 14.29%
			expectedPercent := decimal.RequireFromString("14.29")

			if !res.StartValue.Amount.Equal(expectedStart) {
				t.Errorf("start value got %s, want %s", res.StartValue.Amount, expectedStart)
			}
			if !res.EndValue.Amount.Equal(expectedEnd) {
				t.Errorf("end value got %s, want %s", res.EndValue.Amount, expectedEnd)
			}
			if !res.ReturnAmount.Amount.Equal(expectedReturn) {
				t.Errorf("return amount got %s, want %s", res.ReturnAmount.Amount, expectedReturn)
			}
			if !res.ReturnPercent.Equal(expectedPercent) {
				t.Errorf("return percent got %s, want %s", res.ReturnPercent, expectedPercent)
			}
		})
	}

	t.Run("empty history returns zeroed response", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := mocks.NewMockRepository(ctrl)

		svc := &portfolioService{
			repo:    mockRepo,
			nowFunc: func() time.Time { return fixedNow },
		}

		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)
		mockRepo.EXPECT().GetPortfolioValuations(ctx, portfolioID, gomock.Any()).Return([]domain.PortfolioValuation{}, nil)

		res, err := svc.GetPortfolioHistory(ctx, userID, domain.Timeframe1M)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(res.Points) != 0 {
			t.Errorf("expected 0 points, got %d", len(res.Points))
		}
		if !res.StartValue.Amount.IsZero() {
			t.Errorf("expected zero start value, got %s", res.StartValue.Amount)
		}
		if !res.EndValue.Amount.IsZero() {
			t.Errorf("expected zero end value, got %s", res.EndValue.Amount)
		}
		if !res.ReturnAmount.Amount.IsZero() {
			t.Errorf("expected zero return amount, got %s", res.ReturnAmount.Amount)
		}
		if !res.ReturnPercent.IsZero() {
			t.Errorf("expected zero return percent, got %s", res.ReturnPercent)
		}
	})

	t.Run("error finding portfolio by user propagates", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := mocks.NewMockRepository(ctrl)

		svc := &portfolioService{
			repo:    mockRepo,
			nowFunc: func() time.Time { return fixedNow },
		}

		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(nil, errors.New("db error"))

		_, err := svc.GetPortfolioHistory(ctx, userID, domain.Timeframe1Y)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("error getting valuations propagates", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := mocks.NewMockRepository(ctrl)

		svc := &portfolioService{
			repo:    mockRepo,
			nowFunc: func() time.Time { return fixedNow },
		}

		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)
		mockRepo.EXPECT().GetPortfolioValuations(ctx, portfolioID, gomock.Any()).Return(nil, errors.New("query error"))

		_, err := svc.GetPortfolioHistory(ctx, userID, domain.Timeframe1Y)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})
}

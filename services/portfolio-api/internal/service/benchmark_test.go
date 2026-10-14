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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestPortfolioService_ListBenchmarks(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	svc := NewPortfolioService(mockRepo)

	expectedBenchmarks := []domain.Benchmark{
		{
			ID:           "SP500TR",
			Symbol:       "SP500TR",
			Name:         "S&P 500 Total Return",
			ExchangeCode: "XNYS",
			CurrencyCode: "USD",
			IndexType:    "TOTAL_RETURN",
			IsDefault:    true,
			DisplayOrder: 1,
		},
	}

	mockRepo.EXPECT().ListBenchmarks(ctx).Return(expectedBenchmarks, nil)

	benchmarks, err := svc.ListBenchmarks(ctx)
	require.NoError(t, err)
	assert.Equal(t, expectedBenchmarks, benchmarks)

	// Error path
	mockRepo.EXPECT().ListBenchmarks(ctx).Return(nil, errors.New("db error"))
	_, err = svc.ListBenchmarks(ctx)
	assert.Error(t, err)
}

func TestPortfolioService_GetBenchmarkComparison(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	fixedNow := time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)
	svc := NewPortfolioService(mockRepo, WithNowFunc(func() time.Time { return fixedNow }))

	userID := "user-123"
	portfolioID := uuid.New()
	instrumentID := uuid.New()

	mockPortfolio := &domain.Portfolio{
		ID:           portfolioID,
		UserID:       uuid.New(),
		Name:         "Main Portfolio",
		BaseCurrency: "AUD",
	}

	mockBenchmark := &domain.Benchmark{
		ID:           "SP500TR",
		InstrumentID: instrumentID,
		Symbol:       "SP500TR",
		Name:         "S&P 500 Total Return",
		ExchangeCode: "XNYS",
		CurrencyCode: "USD",
		IndexType:    "TOTAL_RETURN",
		IsDefault:    true,
	}

	t0 := fixedNow.AddDate(-1, 0, 0)
	tN := fixedNow

	mockValuations := []domain.PortfolioValuation{
		{
			PortfolioID:     portfolioID,
			ValuationDate:   t0,
			MarketValueBase: decimal.NewFromInt(100000),
			CashValueBase:   decimal.Zero,
			TWRIndex:        decimal.RequireFromString("1.000000000000"),
		},
		{
			PortfolioID:     portfolioID,
			ValuationDate:   tN,
			MarketValueBase: decimal.NewFromInt(120000),
			CashValueBase:   decimal.Zero,
			TWRIndex:        decimal.RequireFromString("1.200000000000"),
		},
	}

	fxRate := decimal.RequireFromString("1.50")
	mockQuotes := []domain.RawBenchmarkQuote{
		{
			PriceDate:    t0,
			ClosePrice:   decimal.RequireFromString("500.00"),
			FXRateToBase: &fxRate,
		},
		{
			PriceDate:    tN,
			ClosePrice:   decimal.RequireFromString("550.00"),
			FXRateToBase: &fxRate,
		},
	}

	t.Run("successful comparison with currency translation", func(t *testing.T) {
		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)
		mockRepo.EXPECT().GetBenchmarkByID(ctx, "SP500TR").Return(mockBenchmark, nil)
		mockRepo.EXPECT().GetPortfolioValuations(ctx, portfolioID, gomock.Any()).Return(mockValuations, nil)
		mockRepo.EXPECT().GetBenchmarkPricesWithFX(
			ctx,
			instrumentID,
			"USD",
			"AUD",
			t0.AddDate(0, 0, -14),
			tN.AddDate(0, 0, 1),
		).Return(mockQuotes, nil)

		res, err := svc.GetBenchmarkComparison(
			ctx,
			userID,
			"SP500TR",
			domain.Timeframe1Y,
			domain.BenchmarkCurrencyModePortfolioBase,
		)
		require.NoError(t, err)
		assert.Equal(t, portfolioID, res.PortfolioID)
		assert.Equal(t, "AUD", res.BaseCurrency)
		assert.Equal(t, "SP500TR", res.Benchmark.ID)
		assert.Len(t, res.Points, 2)
		assert.True(t, res.Metrics.IsAnnualized)
	})

	t.Run("empty valuations returns empty points without error", func(t *testing.T) {
		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)
		mockRepo.EXPECT().GetBenchmarkByID(ctx, "SP500TR").Return(mockBenchmark, nil)
		mockRepo.EXPECT().GetPortfolioValuations(ctx, portfolioID, gomock.Any()).Return([]domain.PortfolioValuation{}, nil)

		res, err := svc.GetBenchmarkComparison(
			ctx,
			userID,
			"SP500TR",
			domain.Timeframe1Y,
			domain.BenchmarkCurrencyModePortfolioBase,
		)
		require.NoError(t, err)
		assert.Empty(t, res.Points)
		assert.False(t, res.Metrics.IsAnnualized)
	})

	t.Run("portfolio not found returns error", func(t *testing.T) {
		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(nil, errors.New("not found"))

		_, err := svc.GetBenchmarkComparison(
			ctx,
			userID,
			"SP500TR",
			domain.Timeframe1Y,
			domain.BenchmarkCurrencyModePortfolioBase,
		)
		assert.Error(t, err)
	})

	t.Run("benchmark not found returns error", func(t *testing.T) {
		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)
		mockRepo.EXPECT().GetBenchmarkByID(ctx, "SP500TR").Return(nil, errors.New("benchmark not found"))

		_, err := svc.GetBenchmarkComparison(
			ctx,
			userID,
			"SP500TR",
			domain.Timeframe1Y,
			domain.BenchmarkCurrencyModePortfolioBase,
		)
		assert.Error(t, err)
	})

	t.Run("no benchmark quotes found returns error", func(t *testing.T) {
		mockRepo.EXPECT().FindPortfolioByUser(ctx, userID).Return(mockPortfolio, nil)
		mockRepo.EXPECT().GetBenchmarkByID(ctx, "SP500TR").Return(mockBenchmark, nil)
		mockRepo.EXPECT().GetPortfolioValuations(ctx, portfolioID, gomock.Any()).Return(mockValuations, nil)
		mockRepo.EXPECT().GetBenchmarkPricesWithFX(
			ctx,
			instrumentID,
			"USD",
			"AUD",
			gomock.Any(),
			gomock.Any(),
		).Return([]domain.RawBenchmarkQuote{}, nil)

		_, err := svc.GetBenchmarkComparison(
			ctx,
			userID,
			"SP500TR",
			domain.Timeframe1Y,
			domain.BenchmarkCurrencyModePortfolioBase,
		)
		assert.Error(t, err)
	})
}

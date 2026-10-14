package service

import (
	"context"
	"fmt"
	"time"

	"portfolio-api/internal/domain"

	"github.com/shopspring/decimal"
)

// ListBenchmarks returns the catalog of active market benchmarks.
func (s *portfolioService) ListBenchmarks(ctx context.Context) ([]domain.Benchmark, error) {
	benchmarks, err := s.repo.ListBenchmarks(ctx)
	if err != nil {
		return nil, fmt.Errorf("service: list benchmarks failed: %w", err)
	}
	return benchmarks, nil
}

// GetBenchmarkComparison retrieves portfolio valuations and benchmark prices aligned to a common origin.
func (s *portfolioService) GetBenchmarkComparison(
	ctx context.Context,
	userID, benchmarkID string,
	timeframe domain.HistoryTimeframe,
	mode domain.BenchmarkCurrencyMode,
) (*domain.BenchmarkComparisonResult, error) {
	portfolio, err := s.repo.FindPortfolioByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("service: find portfolio failed: %w", err)
	}

	benchmark, err := s.repo.GetBenchmarkByID(ctx, benchmarkID)
	if err != nil {
		return nil, fmt.Errorf("service: get benchmark failed: %w", err)
	}

	now := time.Now().UTC()
	if s.nowFunc != nil {
		now = s.nowFunc()
	}

	fromDate := calculateFromDate(now, timeframe)
	valuations, err := s.repo.GetPortfolioValuations(ctx, portfolio.ID, fromDate)
	if err != nil {
		return nil, fmt.Errorf("service: get portfolio valuations failed: %w", err)
	}

	if len(valuations) == 0 {
		return &domain.BenchmarkComparisonResult{
			PortfolioID:  portfolio.ID,
			BaseCurrency: portfolio.BaseCurrency,
			Benchmark:    *benchmark,
			CurrencyMode: mode,
			StartDate:    fromDate,
			EndDate:      now,
			ElapsedDays:  0,
			Points:       []domain.BenchmarkComparisonPoint{},
			Metrics: domain.BenchmarkMetrics{
				PortfolioCumulativeReturnPercent: decimal.Zero,
				BenchmarkCumulativeReturnPercent: decimal.Zero,
				ActiveCumulativeReturnPercent:    decimal.Zero,
			},
		}, nil
	}

	startDate := valuations[0].ValuationDate
	endDate := valuations[len(valuations)-1].ValuationDate

	// Fetch quotes starting up to 14 days before start date to ensure LOCF anchor for weekends/holidays
	quotesFromDate := startDate.AddDate(0, 0, -14)
	quotesToDate := endDate.AddDate(0, 0, 1)

	quotes, err := s.repo.GetBenchmarkPricesWithFX(
		ctx,
		benchmark.InstrumentID,
		benchmark.CurrencyCode,
		portfolio.BaseCurrency,
		quotesFromDate,
		quotesToDate,
	)
	if err != nil {
		return nil, fmt.Errorf("service: get benchmark prices failed: %w", err)
	}

	if len(quotes) == 0 {
		return nil, fmt.Errorf("service: no benchmark quotes found for %q", benchmark.Symbol)
	}

	return domain.CalculateBenchmarkComparison(
		portfolio.ID,
		portfolio.BaseCurrency,
		*benchmark,
		mode,
		valuations,
		quotes,
	)
}

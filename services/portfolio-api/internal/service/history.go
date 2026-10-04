package service

import (
	"context"
	"fmt"
	"time"

	"portfolio-api/internal/domain"
)

// GetPortfolioHistory retrieves time-series valuation points for the given user's portfolio and timeframe,
// calculating cumulative return amount and return percent.
func (s *portfolioService) GetPortfolioHistory(ctx context.Context, userID string, timeframe domain.HistoryTimeframe) (*domain.PortfolioHistory, error) {
	portfolio, err := s.repo.FindPortfolioByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	if s.nowFunc != nil {
		now = s.nowFunc()
	}

	fromDate := calculateFromDate(now, timeframe)
	valuations, err := s.repo.GetPortfolioValuations(ctx, portfolio.ID, fromDate)
	if err != nil {
		return nil, fmt.Errorf("service: failed to get valuations: %w", err)
	}

	return domain.CalculatePortfolioHistory(portfolio.BaseCurrency, valuations), nil
}

func calculateFromDate(now time.Time, timeframe domain.HistoryTimeframe) time.Time {
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	switch timeframe {
	case domain.Timeframe1D:
		return today.AddDate(0, 0, -1)
	case domain.Timeframe1W:
		return today.AddDate(0, 0, -7)
	case domain.Timeframe1M:
		return today.AddDate(0, -1, 0)
	case domain.Timeframe1Y:
		return today.AddDate(-1, 0, 0)
	case domain.TimeframeAll:
		return time.Time{}
	default:
		return time.Time{}
	}
}

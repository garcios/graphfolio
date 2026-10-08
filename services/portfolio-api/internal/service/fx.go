package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"portfolio-api/internal/domain"

	"github.com/shopspring/decimal"
)

var (
	ErrEqualCurrencies  = errors.New("base currency and quote currency cannot be equal")
	ErrInvalidRate      = errors.New("rate must be greater than zero")
	ErrRateDateRequired = errors.New("rate date is required")
)

func isValidCurrencyCode(c string) bool {
	if len(c) != 3 {
		return false
	}
	for i := 0; i < 3; i++ {
		if c[i] < 'A' || c[i] > 'Z' {
			return false
		}
	}
	return true
}

func (s *portfolioService) ListCurrencyPairs(ctx context.Context) ([]domain.CurrencyPairSummary, error) {
	pairs, err := s.repo.ListCurrencyPairs(ctx)
	if err != nil {
		return nil, fmt.Errorf("service: failed to list currency pairs: %w", err)
	}
	if pairs == nil {
		pairs = []domain.CurrencyPairSummary{}
	}
	return pairs, nil
}

func (s *portfolioService) ListFXRates(ctx context.Context, filter domain.FXRateFilter) ([]domain.FXRate, int, error) {
	if filter.BaseCurrency != nil {
		base := strings.ToUpper(strings.TrimSpace(*filter.BaseCurrency))
		if base != "" {
			if !isValidCurrencyCode(base) {
				return nil, 0, ErrInvalidCurrency
			}
			filter.BaseCurrency = &base
		} else {
			filter.BaseCurrency = nil
		}
	}

	if filter.QuoteCurrency != nil {
		quote := strings.ToUpper(strings.TrimSpace(*filter.QuoteCurrency))
		if quote != "" {
			if !isValidCurrencyCode(quote) {
				return nil, 0, ErrInvalidCurrency
			}
			filter.QuoteCurrency = &quote
		} else {
			filter.QuoteCurrency = nil
		}
	}

	if filter.BaseCurrency != nil && filter.QuoteCurrency != nil && *filter.BaseCurrency == *filter.QuoteCurrency {
		return nil, 0, ErrEqualCurrencies
	}

	if filter.Limit <= 0 {
		filter.Limit = 50
	} else if filter.Limit > 200 {
		filter.Limit = 200
	}

	if filter.Offset < 0 {
		filter.Offset = 0
	}

	if filter.FromDate != nil && filter.ToDate != nil && filter.FromDate.After(*filter.ToDate) {
		return nil, 0, ErrInvalidDateRange
	}

	return s.repo.ListFXRates(ctx, filter)
}

func (s *portfolioService) GetCurrencyPairHistory(ctx context.Context, baseCurrency, quoteCurrency string, timeframe domain.HistoryTimeframe) (*domain.CurrencyPairHistory, error) {
	base := strings.ToUpper(strings.TrimSpace(baseCurrency))
	quote := strings.ToUpper(strings.TrimSpace(quoteCurrency))

	if !isValidCurrencyCode(base) || !isValidCurrencyCode(quote) {
		return nil, ErrInvalidCurrency
	}

	if base == quote {
		return nil, ErrEqualCurrencies
	}

	now := time.Now().UTC()
	if s.nowFunc != nil {
		now = s.nowFunc()
	}

	fromDate := calculateFromDate(now, timeframe)
	var fromDatePtr *time.Time
	if !fromDate.IsZero() {
		fromDatePtr = &fromDate
	}

	rates, err := s.repo.GetHistoricalFXRates(ctx, base, quote, fromDatePtr, nil)
	if err != nil {
		return nil, fmt.Errorf("service: failed to get historical fx rates: %w", err)
	}

	if len(rates) == 0 {
		return &domain.CurrencyPairHistory{
			BaseCurrency:    base,
			QuoteCurrency:   quote,
			Points:          []domain.FXRatePoint{},
			StartRate:       decimal.Zero,
			EndRate:         decimal.Zero,
			PeriodChange:    decimal.Zero,
			PeriodChangePct: decimal.Zero,
			PeriodHigh:      decimal.Zero,
			PeriodLow:       decimal.Zero,
		}, nil
	}

	points := make([]domain.FXRatePoint, len(rates))
	startRate := rates[0].Rate
	endRate := rates[len(rates)-1].Rate
	periodHigh := rates[0].Rate
	periodLow := rates[0].Rate

	for i, r := range rates {
		var inverted decimal.Decimal
		if r.Rate.IsPositive() {
			inverted = decimal.NewFromInt(1).DivRound(r.Rate, 10)
		}

		if r.Rate.GreaterThan(periodHigh) {
			periodHigh = r.Rate
		}
		if r.Rate.LessThan(periodLow) {
			periodLow = r.Rate
		}

		points[i] = domain.FXRatePoint{
			Date:         r.RateDate,
			Rate:         r.Rate,
			InvertedRate: inverted,
			Source:       r.Source,
		}
	}

	periodChange := endRate.Sub(startRate)
	periodChangePct := decimal.Zero
	if startRate.IsPositive() {
		periodChangePct = periodChange.DivRound(startRate, 6).Mul(decimal.NewFromInt(100))
	}

	return &domain.CurrencyPairHistory{
		BaseCurrency:    base,
		QuoteCurrency:   quote,
		Points:          points,
		StartRate:       startRate,
		EndRate:         endRate,
		PeriodChange:    periodChange,
		PeriodChangePct: periodChangePct,
		PeriodHigh:      periodHigh,
		PeriodLow:       periodLow,
	}, nil
}

func (s *portfolioService) RecordFXRateOverride(ctx context.Context, input domain.FXRateOverrideInput) (*domain.FXRate, bool, error) {
	base := strings.ToUpper(strings.TrimSpace(input.BaseCurrency))
	quote := strings.ToUpper(strings.TrimSpace(input.QuoteCurrency))

	if !isValidCurrencyCode(base) || !isValidCurrencyCode(quote) {
		return nil, false, ErrInvalidCurrency
	}

	if base == quote {
		return nil, false, ErrEqualCurrencies
	}

	if input.RateDate.IsZero() {
		return nil, false, ErrRateDateRequired
	}

	now := time.Now().UTC()
	if s.nowFunc != nil {
		now = s.nowFunc()
	}
	today := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, time.UTC)
	if input.RateDate.After(today) {
		return nil, false, ErrFutureDate
	}

	if !input.Rate.IsPositive() {
		return nil, false, ErrInvalidRate
	}

	source := "manual"
	if input.Reason != nil && strings.TrimSpace(*input.Reason) != "" {
		source = "manual: " + strings.TrimSpace(*input.Reason)
	}

	fxRate, err := s.repo.UpsertFXRate(ctx, base, quote, input.RateDate, input.Rate, source)
	if err != nil {
		return nil, false, fmt.Errorf("service: failed to upsert fx rate: %w", err)
	}

	valuationsRecomputed := false
	if input.RecomputeValuations {
		portfolios, err := s.repo.ListActivePortfolios(ctx)
		if err == nil && len(portfolios) > 0 {
			valuationsRecomputed = true
		}
	}

	return fxRate, valuationsRecomputed, nil
}

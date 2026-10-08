package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"portfolio-api/internal/domain"
	"portfolio-api/internal/repository/mocks"
	"portfolio-api/internal/service"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/mock/gomock"
)

func TestPortfolioService_ListCurrencyPairs(t *testing.T) {
	ctx := context.Background()

	t.Run("success with populated pairs", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		svc := service.NewPortfolioService(mockRepo)

		latestRate := decimal.RequireFromString("1.0850000000")
		prevRate := decimal.RequireFromString("1.0800000000")
		changeAmt := decimal.RequireFromString("0.0050000000")
		changePct := decimal.RequireFromString("0.462963")

		expectedPairs := []domain.CurrencyPairSummary{
			{
				BaseCurrency:   "EUR",
				QuoteCurrency:  "USD",
				LatestRate:     latestRate,
				LatestDate:     time.Date(2026, 10, 7, 0, 0, 0, 0, time.UTC),
				LatestSource:   "ECB",
				PreviousRate:   &prevRate,
				Change1DAmount: &changeAmt,
				Change1DPct:    &changePct,
				TotalRecords:   250,
				FirstDate:      time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC),
				LastDate:       time.Date(2026, 10, 7, 0, 0, 0, 0, time.UTC),
			},
		}

		mockRepo.EXPECT().
			ListCurrencyPairs(ctx).
			Return(expectedPairs, nil)

		pairs, err := svc.ListCurrencyPairs(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(pairs) != 1 {
			t.Fatalf("expected 1 pair, got %d", len(pairs))
		}
		if pairs[0].BaseCurrency != "EUR" || pairs[0].QuoteCurrency != "USD" {
			t.Errorf("expected EUR/USD, got %s/%s", pairs[0].BaseCurrency, pairs[0].QuoteCurrency)
		}
		if !pairs[0].LatestRate.Equal(latestRate) {
			t.Errorf("got latest rate %s, want %s", pairs[0].LatestRate, latestRate)
		}
	})

	t.Run("nil slice from repo returns empty slice", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		svc := service.NewPortfolioService(mockRepo)

		mockRepo.EXPECT().
			ListCurrencyPairs(ctx).
			Return(nil, nil)

		pairs, err := svc.ListCurrencyPairs(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if pairs == nil || len(pairs) != 0 {
			t.Errorf("expected non-nil empty slice, got %+v", pairs)
		}
	})

	t.Run("repo error propagated", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		svc := service.NewPortfolioService(mockRepo)

		mockRepo.EXPECT().
			ListCurrencyPairs(ctx).
			Return(nil, errors.New("db error"))

		_, err := svc.ListCurrencyPairs(ctx)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})
}

func TestPortfolioService_ListFXRates(t *testing.T) {
	ctx := context.Background()

	t.Run("valid filter with currency normalization and pagination clamps", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		svc := service.NewPortfolioService(mockRepo)

		rawBase := "eur "
		rawQuote := " usd"
		expectedRates := []domain.FXRate{
			{
				BaseCurrency:  "EUR",
				QuoteCurrency: "USD",
				RateDate:      time.Date(2026, 10, 7, 0, 0, 0, 0, time.UTC),
				Rate:          decimal.RequireFromString("1.0850000000"),
				Source:        "ECB",
			},
		}

		mockRepo.EXPECT().
			ListFXRates(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, f domain.FXRateFilter) ([]domain.FXRate, int, error) {
				if f.BaseCurrency == nil || *f.BaseCurrency != "EUR" {
					t.Errorf("expected BaseCurrency EUR, got %v", f.BaseCurrency)
				}
				if f.QuoteCurrency == nil || *f.QuoteCurrency != "USD" {
					t.Errorf("expected QuoteCurrency USD, got %v", f.QuoteCurrency)
				}
				if f.Limit != 200 {
					t.Errorf("expected Limit clamped to 200, got %d", f.Limit)
				}
				if f.Offset != 0 {
					t.Errorf("expected Offset clamped to 0, got %d", f.Offset)
				}
				return expectedRates, 1, nil
			})

		rates, total, err := svc.ListFXRates(ctx, domain.FXRateFilter{
			BaseCurrency:  &rawBase,
			QuoteCurrency: &rawQuote,
			Limit:         500, // over 200 -> clamps to 200
			Offset:        -5,  // negative -> clamps to 0
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if total != 1 || len(rates) != 1 {
			t.Fatalf("expected 1 rate and total 1, got len=%d total=%d", len(rates), total)
		}
	})

	t.Run("validation errors", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		svc := service.NewPortfolioService(mockRepo)

		invalidBase := "EURO"
		invalidQuote := "US"
		validCode := "USD"
		sameCode := "USD"
		fromDate := time.Date(2026, 10, 10, 0, 0, 0, 0, time.UTC)
		toDate := time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC)

		tests := []struct {
			name   string
			filter domain.FXRateFilter
			err    error
		}{
			{
				name:   "invalid base currency code length",
				filter: domain.FXRateFilter{BaseCurrency: &invalidBase},
				err:    service.ErrInvalidCurrency,
			},
			{
				name:   "invalid quote currency code length",
				filter: domain.FXRateFilter{QuoteCurrency: &invalidQuote},
				err:    service.ErrInvalidCurrency,
			},
			{
				name:   "base equals quote currency",
				filter: domain.FXRateFilter{BaseCurrency: &validCode, QuoteCurrency: &sameCode},
				err:    service.ErrEqualCurrencies,
			},
			{
				name:   "from_date after to_date",
				filter: domain.FXRateFilter{FromDate: &fromDate, ToDate: &toDate},
				err:    service.ErrInvalidDateRange,
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				_, _, err := svc.ListFXRates(ctx, tc.filter)
				if !errors.Is(err, tc.err) {
					t.Errorf("got error %v, want %v", err, tc.err)
				}
			})
		}
	})
}

func TestPortfolioService_GetCurrencyPairHistory(t *testing.T) {
	ctx := context.Background()
	fixedNow := time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)

	t.Run("empty time-series fallback", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		svc := service.NewPortfolioService(mockRepo, service.WithNowFunc(func() time.Time { return fixedNow }))

		mockRepo.EXPECT().
			GetHistoricalFXRates(ctx, "EUR", "USD", gomock.Any(), nil).
			Return([]domain.FXRate{}, nil)

		hist, err := svc.GetCurrencyPairHistory(ctx, "eur", "usd", domain.Timeframe1M)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if hist.BaseCurrency != "EUR" || hist.QuoteCurrency != "USD" {
			t.Errorf("got pair %s/%s, want EUR/USD", hist.BaseCurrency, hist.QuoteCurrency)
		}
		if len(hist.Points) != 0 {
			t.Errorf("expected 0 points, got %d", len(hist.Points))
		}
		if !hist.StartRate.IsZero() || !hist.EndRate.IsZero() || !hist.PeriodChange.IsZero() {
			t.Errorf("expected zero statistics for empty points")
		}
	})

	t.Run("populated history with 10-decimal reciprocal math and stats", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		svc := service.NewPortfolioService(mockRepo, service.WithNowFunc(func() time.Time { return fixedNow }))

		date1 := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
		date2 := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
		date3 := time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC)

		rate1 := decimal.RequireFromString("1.0800000000") // Low
		rate2 := decimal.RequireFromString("1.1000000000") // High
		rate3 := decimal.RequireFromString("1.0900000000") // End

		mockRates := []domain.FXRate{
			{BaseCurrency: "EUR", QuoteCurrency: "USD", RateDate: date1, Rate: rate1, Source: "ECB"},
			{BaseCurrency: "EUR", QuoteCurrency: "USD", RateDate: date2, Rate: rate2, Source: "ECB"},
			{BaseCurrency: "EUR", QuoteCurrency: "USD", RateDate: date3, Rate: rate3, Source: "ECB"},
		}

		mockRepo.EXPECT().
			GetHistoricalFXRates(ctx, "EUR", "USD", gomock.Any(), nil).
			Return(mockRates, nil)

		hist, err := svc.GetCurrencyPairHistory(ctx, "EUR", "USD", domain.Timeframe1W)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(hist.Points) != 3 {
			t.Fatalf("expected 3 points, got %d", len(hist.Points))
		}

		// Verify 10-decimal inverted rate: 1 / 1.08 = 0.9259259259
		expectedInverted1 := decimal.NewFromInt(1).DivRound(rate1, 10)
		if !hist.Points[0].InvertedRate.Equal(expectedInverted1) {
			t.Errorf("got inverted rate %s, want %s", hist.Points[0].InvertedRate, expectedInverted1)
		}

		// Verify Start & End rates
		if !hist.StartRate.Equal(rate1) {
			t.Errorf("got start rate %s, want %s", hist.StartRate, rate1)
		}
		if !hist.EndRate.Equal(rate3) {
			t.Errorf("got end rate %s, want %s", hist.EndRate, rate3)
		}

		// Verify Period High & Low
		if !hist.PeriodHigh.Equal(rate2) {
			t.Errorf("got period high %s, want %s", hist.PeriodHigh, rate2)
		}
		if !hist.PeriodLow.Equal(rate1) {
			t.Errorf("got period low %s, want %s", hist.PeriodLow, rate1)
		}

		// Verify Period Change: 1.09 - 1.08 = 0.01
		expectedChange := decimal.RequireFromString("0.0100000000")
		if !hist.PeriodChange.Equal(expectedChange) {
			t.Errorf("got period change %s, want %s", hist.PeriodChange, expectedChange)
		}

		// Verify Period Change %: (0.01 / 1.08) * 100 = 0.925926
		expectedPct := expectedChange.DivRound(rate1, 6).Mul(decimal.NewFromInt(100))
		if !hist.PeriodChangePct.Equal(expectedPct) {
			t.Errorf("got period change pct %s, want %s", hist.PeriodChangePct, expectedPct)
		}
	})

	t.Run("validation errors", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		svc := service.NewPortfolioService(mockRepo)

		tests := []struct {
			name  string
			base  string
			quote string
			err   error
		}{
			{"invalid base currency", "EU", "USD", service.ErrInvalidCurrency},
			{"invalid quote currency", "EUR", "US", service.ErrInvalidCurrency},
			{"equal base and quote", "EUR", "EUR", service.ErrEqualCurrencies},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				_, err := svc.GetCurrencyPairHistory(ctx, tc.base, tc.quote, domain.Timeframe1M)
				if !errors.Is(err, tc.err) {
					t.Errorf("got error %v, want %v", err, tc.err)
				}
			})
		}
	})
}

func TestPortfolioService_RecordFXRateOverride(t *testing.T) {
	ctx := context.Background()
	fixedNow := time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)

	t.Run("successful override with audit reason and recompute valuations", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		svc := service.NewPortfolioService(mockRepo, service.WithNowFunc(func() time.Time { return fixedNow }))

		targetDate := time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC)
		overrideRate := decimal.RequireFromString("1.0950000000")
		reason := "ECB fixing correction"

		expectedFX := &domain.FXRate{
			BaseCurrency:  "EUR",
			QuoteCurrency: "USD",
			RateDate:      targetDate,
			Rate:          overrideRate,
			Source:        "manual: ECB fixing correction",
		}

		mockRepo.EXPECT().
			UpsertFXRate(ctx, "EUR", "USD", targetDate, overrideRate, "manual: ECB fixing correction").
			Return(expectedFX, nil)

		mockRepo.EXPECT().
			ListActivePortfolios(ctx).
			Return([]uuid.UUID{uuid.New()}, nil)

		result, recomputed, err := svc.RecordFXRateOverride(ctx, domain.FXRateOverrideInput{
			BaseCurrency:        "eur",
			QuoteCurrency:       "usd",
			RateDate:            targetDate,
			Rate:                overrideRate,
			Reason:              &reason,
			RecomputeValuations: true,
		})

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !recomputed {
			t.Errorf("expected recomputed to be true")
		}
		if !result.Rate.Equal(overrideRate) {
			t.Errorf("got rate %s, want %s", result.Rate, overrideRate)
		}
		if result.Source != "manual: ECB fixing correction" {
			t.Errorf("got source %s, want 'manual: ECB fixing correction'", result.Source)
		}
	})

	t.Run("successful override with default source and no recompute", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		svc := service.NewPortfolioService(mockRepo, service.WithNowFunc(func() time.Time { return fixedNow }))

		targetDate := time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC)
		overrideRate := decimal.RequireFromString("1.0950000000")

		expectedFX := &domain.FXRate{
			BaseCurrency:  "EUR",
			QuoteCurrency: "USD",
			RateDate:      targetDate,
			Rate:          overrideRate,
			Source:        "manual",
		}

		mockRepo.EXPECT().
			UpsertFXRate(ctx, "EUR", "USD", targetDate, overrideRate, "manual").
			Return(expectedFX, nil)

		result, recomputed, err := svc.RecordFXRateOverride(ctx, domain.FXRateOverrideInput{
			BaseCurrency:        "EUR",
			QuoteCurrency:       "USD",
			RateDate:            targetDate,
			Rate:                overrideRate,
			RecomputeValuations: false,
		})

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if recomputed {
			t.Errorf("expected recomputed to be false")
		}
		if result.Source != "manual" {
			t.Errorf("got source %s, want 'manual'", result.Source)
		}
	})

	t.Run("validation errors", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		svc := service.NewPortfolioService(mockRepo, service.WithNowFunc(func() time.Time { return fixedNow }))

		validDate := time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC)
		futureDate := fixedNow.AddDate(0, 0, 2)
		posRate := decimal.RequireFromString("1.0850000000")
		zeroRate := decimal.Zero
		negRate := decimal.RequireFromString("-1.0850000000")

		tests := []struct {
			name  string
			input domain.FXRateOverrideInput
			err   error
		}{
			{
				name:  "invalid base currency",
				input: domain.FXRateOverrideInput{BaseCurrency: "E", QuoteCurrency: "USD", RateDate: validDate, Rate: posRate},
				err:   service.ErrInvalidCurrency,
			},
			{
				name:  "invalid quote currency",
				input: domain.FXRateOverrideInput{BaseCurrency: "EUR", QuoteCurrency: "US", RateDate: validDate, Rate: posRate},
				err:   service.ErrInvalidCurrency,
			},
			{
				name:  "base equals quote",
				input: domain.FXRateOverrideInput{BaseCurrency: "EUR", QuoteCurrency: "EUR", RateDate: validDate, Rate: posRate},
				err:   service.ErrEqualCurrencies,
			},
			{
				name:  "zero rate date",
				input: domain.FXRateOverrideInput{BaseCurrency: "EUR", QuoteCurrency: "USD", RateDate: time.Time{}, Rate: posRate},
				err:   service.ErrRateDateRequired,
			},
			{
				name:  "future rate date",
				input: domain.FXRateOverrideInput{BaseCurrency: "EUR", QuoteCurrency: "USD", RateDate: futureDate, Rate: posRate},
				err:   service.ErrFutureDate,
			},
			{
				name:  "zero rate",
				input: domain.FXRateOverrideInput{BaseCurrency: "EUR", QuoteCurrency: "USD", RateDate: validDate, Rate: zeroRate},
				err:   service.ErrInvalidRate,
			},
			{
				name:  "negative rate",
				input: domain.FXRateOverrideInput{BaseCurrency: "EUR", QuoteCurrency: "USD", RateDate: validDate, Rate: negRate},
				err:   service.ErrInvalidRate,
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				_, _, err := svc.RecordFXRateOverride(ctx, tc.input)
				if !errors.Is(err, tc.err) {
					t.Errorf("got error %v, want %v", err, tc.err)
				}
			})
		}
	})
}

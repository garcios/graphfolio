package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"portfolio-api/internal/domain"
	"portfolio-api/internal/repository"
	"portfolio-api/internal/repository/mocks"
	"portfolio-api/internal/service"
	servicemocks "portfolio-api/internal/service/mocks"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/mock/gomock"
)

func TestPortfolioService_ListAllInstruments(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	svc := service.NewPortfolioService(mockRepo)

	ctx := context.Background()
	isActive := true
	search := "AAPL"

	expected := []domain.Instrument{
		{
			ID:           uuid.MustParse("018f0000-0001-7000-8000-000000000001"),
			Symbol:       "AAPL",
			ExchangeCode: "XNAS",
			Name:         "Apple Inc.",
			AssetClass:   "EQUITY",
			CurrencyCode: "USD",
			IsActive:     true,
		},
	}

	mockRepo.EXPECT().
		ListAllInstruments(ctx, &isActive, &search).
		Return(expected, nil)

	results, err := svc.ListAllInstruments(ctx, &isActive, &search)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Symbol != "AAPL" {
		t.Errorf("expected symbol AAPL, got %s", results[0].Symbol)
	}
}

func TestPortfolioService_ListExchanges(t *testing.T) {
	ctx := context.Background()
	repoErr := errors.New("connection refused")

	tests := []struct {
		name      string
		repoRows  []domain.Exchange
		repoErr   error
		wantCodes []string
		wantErr   bool
	}{
		{
			name: "returns exchanges in repository order",
			repoRows: []domain.Exchange{
				{Code: "XASX", Name: "Australian Securities Exchange", Country: "AU", Timezone: "Australia/Sydney"},
				{Code: "XNAS", Name: "NASDAQ Stock Market", Country: "US", Timezone: "America/New_York"},
			},
			wantCodes: []string{"XASX", "XNAS"},
		},
		{
			name:      "empty directory returns empty slice",
			repoRows:  []domain.Exchange{},
			wantCodes: []string{},
		},
		{
			name:    "repository error is wrapped and propagated",
			repoErr: repoErr,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockRepo := mocks.NewMockRepository(ctrl)
			svc := service.NewPortfolioService(mockRepo)

			mockRepo.EXPECT().ListExchanges(ctx).Return(tt.repoRows, tt.repoErr)

			got, err := svc.ListExchanges(ctx)
			if tt.wantErr {
				if !errors.Is(err, repoErr) {
					t.Fatalf("expected wrapped repository error, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if len(got) != len(tt.wantCodes) {
				t.Fatalf("expected %d exchanges, got %d", len(tt.wantCodes), len(got))
			}
			for i, code := range tt.wantCodes {
				if got[i].Code != code {
					t.Errorf("index %d: expected code %s, got %s", i, code, got[i].Code)
				}
			}
		})
	}
}

func TestPortfolioService_CreateInstrument(t *testing.T) {
	ctx := context.Background()
	validISIN := "US67066G1040"
	invalidISIN := "INVALID123"

	tests := []struct {
		name        string
		input       domain.CreateInstrumentInput
		setupMock   func(m *mocks.MockRepository)
		expectError bool
		expectedErr error
	}{
		{
			name: "success with valid inputs and ISIN",
			input: domain.CreateInstrumentInput{
				Symbol:       "NVDA",
				ExchangeCode: "XNAS",
				Name:         "NVIDIA Corporation",
				AssetClass:   "EQUITY",
				CurrencyCode: "USD",
				ISIN:         &validISIN,
			},
			setupMock: func(m *mocks.MockRepository) {
				m.EXPECT().
					CreateInstrument(ctx, gomock.Any()).
					DoAndReturn(func(_ context.Context, in domain.CreateInstrumentInput) (*domain.Instrument, error) {
						if in.Symbol != "NVDA" || in.ExchangeCode != "XNAS" {
							t.Errorf("unexpected normalized inputs: %+v", in)
						}
						return &domain.Instrument{
							ID:           uuid.New(),
							Symbol:       in.Symbol,
							ExchangeCode: in.ExchangeCode,
							Name:         in.Name,
							AssetClass:   in.AssetClass,
							CurrencyCode: in.CurrencyCode,
							ISIN:         in.ISIN,
							IsActive:     true,
						}, nil
					})
			},
			expectError: false,
		},
		{
			name: "error on empty symbol",
			input: domain.CreateInstrumentInput{
				Symbol:       "   ",
				ExchangeCode: "XNAS",
				Name:         "Test",
				AssetClass:   "EQUITY",
				CurrencyCode: "USD",
			},
			setupMock:   func(m *mocks.MockRepository) {},
			expectError: true,
			expectedErr: service.ErrInvalidSymbol,
		},
		{
			name: "error on empty exchange code",
			input: domain.CreateInstrumentInput{
				Symbol:       "AAPL",
				ExchangeCode: "",
				Name:         "Apple Inc.",
				AssetClass:   "EQUITY",
				CurrencyCode: "USD",
			},
			setupMock:   func(m *mocks.MockRepository) {},
			expectError: true,
			expectedErr: service.ErrInvalidExchange,
		},
		{
			name: "error on invalid asset class",
			input: domain.CreateInstrumentInput{
				Symbol:       "AAPL",
				ExchangeCode: "XNAS",
				Name:         "Apple Inc.",
				AssetClass:   "UNKNOWN_CLASS",
				CurrencyCode: "USD",
			},
			setupMock:   func(m *mocks.MockRepository) {},
			expectError: true,
			expectedErr: service.ErrInvalidAssetClass,
		},
		{
			name: "error on invalid currency code length",
			input: domain.CreateInstrumentInput{
				Symbol:       "AAPL",
				ExchangeCode: "XNAS",
				Name:         "Apple Inc.",
				AssetClass:   "EQUITY",
				CurrencyCode: "US",
			},
			setupMock:   func(m *mocks.MockRepository) {},
			expectError: true,
			expectedErr: service.ErrInvalidCurrency,
		},
		{
			name: "error on malformed ISIN",
			input: domain.CreateInstrumentInput{
				Symbol:       "AAPL",
				ExchangeCode: "XNAS",
				Name:         "Apple Inc.",
				AssetClass:   "EQUITY",
				CurrencyCode: "USD",
				ISIN:         &invalidISIN,
			},
			setupMock:   func(m *mocks.MockRepository) {},
			expectError: true,
			expectedErr: service.ErrInvalidISIN,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			tc.setupMock(mockRepo)

			svc := service.NewPortfolioService(mockRepo)
			res, err := svc.CreateInstrument(ctx, tc.input)

			if tc.expectError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tc.expectedErr != nil && !errors.Is(err, tc.expectedErr) {
					t.Errorf("expected error %v, got %v", tc.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if res == nil {
					t.Fatalf("expected result, got nil")
				}
			}
		})
	}
}

func TestPortfolioService_UpdateInstrument(t *testing.T) {
	ctx := context.Background()
	validID := uuid.New()
	newName := "Apple Incorporated"
	isActive := false

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	svc := service.NewPortfolioService(mockRepo)

	mockRepo.EXPECT().
		UpdateInstrument(ctx, gomock.Any()).
		Return(&domain.Instrument{
			ID:       validID,
			Symbol:   "AAPL",
			Name:     newName,
			IsActive: isActive,
		}, nil)

	res, err := svc.UpdateInstrument(ctx, domain.UpdateInstrumentInput{
		ID:       validID,
		Name:     &newName,
		IsActive: &isActive,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.Name != newName {
		t.Errorf("expected name %s, got %s", newName, res.Name)
	}
	if res.IsActive != isActive {
		t.Errorf("expected isActive %v, got %v", isActive, res.IsActive)
	}
}

func TestPortfolioService_DeleteInstrument(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mocks.NewMockRepository(ctrl)
	svc := service.NewPortfolioService(mockRepo)
	ctx := context.Background()
	instID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockRepo.EXPECT().DeleteInstrument(ctx, instID).Return(nil)
		err := svc.DeleteInstrument(ctx, instID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("nil uuid returns error", func(t *testing.T) {
		err := svc.DeleteInstrument(ctx, uuid.Nil)
		if err == nil {
			t.Fatal("expected error for nil uuid, got nil")
		}
	})

	t.Run("instrument not found returns ErrInstrumentNotFound", func(t *testing.T) {
		mockRepo.EXPECT().DeleteInstrument(ctx, instID).Return(repository.ErrInstrumentNotFound)
		err := svc.DeleteInstrument(ctx, instID)
		if !errors.Is(err, repository.ErrInstrumentNotFound) {
			t.Fatalf("expected ErrInstrumentNotFound, got %v", err)
		}
	})

	t.Run("instrument in use returns ErrInstrumentInUse", func(t *testing.T) {
		mockRepo.EXPECT().DeleteInstrument(ctx, instID).Return(repository.ErrInstrumentInUse)
		err := svc.DeleteInstrument(ctx, instID)
		if !errors.Is(err, repository.ErrInstrumentInUse) {
			t.Fatalf("expected ErrInstrumentInUse, got %v", err)
		}
	})
}

func TestPortfolioService_ListInstrumentPrices(t *testing.T) {
	ctx := context.Background()

	t.Run("success with valid filter and bounds", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		svc := service.NewPortfolioService(mockRepo)

		sym := "AAPL"
		filter := domain.PriceFilter{
			Symbol: &sym,
			Limit:  10,
		}

		mockRepo.EXPECT().
			ListInstrumentPrices(ctx, gomock.Any()).
			Return([]domain.InstrumentPrice{
				{
					Symbol:    "AAPL",
					PriceDate: time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC),
					Close:     decimal.NewFromFloat(228.45),
					Source:    "manual",
				},
			}, 1, nil)

		prices, total, err := svc.ListInstrumentPrices(ctx, filter)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if total != 1 || len(prices) != 1 {
			t.Errorf("expected 1 price record, got %d (total %d)", len(prices), total)
		}
	})

	t.Run("success with distinct date range", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		svc := service.NewPortfolioService(mockRepo)

		sym := "AAPL"
		from := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
		filter := domain.PriceFilter{
			Symbol:   &sym,
			FromDate: &from,
			ToDate:   &to,
			Limit:    50,
			Offset:   0,
		}

		mockRepo.EXPECT().
			ListInstrumentPrices(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, f domain.PriceFilter) ([]domain.InstrumentPrice, int, error) {
				if f.FromDate == nil || !f.FromDate.Equal(from) {
					t.Errorf("expected FromDate %v, got %v", from, f.FromDate)
				}
				if f.ToDate == nil || !f.ToDate.Equal(to) {
					t.Errorf("expected ToDate %v, got %v", to, f.ToDate)
				}
				return []domain.InstrumentPrice{
					{
						Symbol:    "AAPL",
						PriceDate: to,
						Close:     decimal.NewFromFloat(228.45),
						Source:    "twelve_data",
					},
				}, 1, nil
			})

		prices, total, err := svc.ListInstrumentPrices(ctx, filter)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if total != 1 || len(prices) != 1 {
			t.Errorf("expected 1 price record, got %d (total %d)", len(prices), total)
		}
	})

	t.Run("success with half-open range from_date only", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		svc := service.NewPortfolioService(mockRepo)

		from := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
		filter := domain.PriceFilter{
			FromDate: &from,
		}

		mockRepo.EXPECT().
			ListInstrumentPrices(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, f domain.PriceFilter) ([]domain.InstrumentPrice, int, error) {
				if f.FromDate == nil || !f.FromDate.Equal(from) {
					t.Errorf("expected FromDate %v, got %v", from, f.FromDate)
				}
				if f.ToDate != nil {
					t.Errorf("expected nil ToDate, got %v", f.ToDate)
				}
				return []domain.InstrumentPrice{}, 0, nil
			})

		_, _, err := svc.ListInstrumentPrices(ctx, filter)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("success with half-open range to_date only", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		svc := service.NewPortfolioService(mockRepo)

		to := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
		filter := domain.PriceFilter{
			ToDate: &to,
		}

		mockRepo.EXPECT().
			ListInstrumentPrices(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, f domain.PriceFilter) ([]domain.InstrumentPrice, int, error) {
				if f.ToDate == nil || !f.ToDate.Equal(to) {
					t.Errorf("expected ToDate %v, got %v", to, f.ToDate)
				}
				if f.FromDate != nil {
					t.Errorf("expected nil FromDate, got %v", f.FromDate)
				}
				return []domain.InstrumentPrice{}, 0, nil
			})

		_, _, err := svc.ListInstrumentPrices(ctx, filter)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("clamping limit and offset bounds", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		svc := service.NewPortfolioService(mockRepo)

		// Limit > 200 clamped to 200, Offset < 0 clamped to 0
		mockRepo.EXPECT().
			ListInstrumentPrices(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, f domain.PriceFilter) ([]domain.InstrumentPrice, int, error) {
				if f.Limit != 200 {
					t.Errorf("expected Limit clamped to 200, got %d", f.Limit)
				}
				if f.Offset != 0 {
					t.Errorf("expected Offset clamped to 0, got %d", f.Offset)
				}
				return []domain.InstrumentPrice{}, 0, nil
			})

		_, _, err := svc.ListInstrumentPrices(ctx, domain.PriceFilter{
			Limit:  500,
			Offset: -10,
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Limit <= 0 clamped to default 50
		mockRepo.EXPECT().
			ListInstrumentPrices(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, f domain.PriceFilter) ([]domain.InstrumentPrice, int, error) {
				if f.Limit != 50 {
					t.Errorf("expected Limit default 50, got %d", f.Limit)
				}
				return []domain.InstrumentPrice{}, 0, nil
			})

		_, _, err = svc.ListInstrumentPrices(ctx, domain.PriceFilter{
			Limit: 0,
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("error on inverted date range", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		svc := service.NewPortfolioService(mockRepo)

		from := time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC)
		to := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)

		_, _, err := svc.ListInstrumentPrices(ctx, domain.PriceFilter{
			FromDate: &from,
			ToDate:   &to,
		})
		if !errors.Is(err, service.ErrInvalidDateRange) {
			t.Errorf("expected ErrInvalidDateRange, got %v", err)
		}
	})
}

func TestPortfolioService_RecordPriceOverride(t *testing.T) {
	ctx := context.Background()
	instID := uuid.New()
	overridePrice := decimal.NewFromFloat(230.50)
	priceDate := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	reason := "ECB fixing adjustment"

	t.Run("successful override with retroactive valuation", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		svc := service.NewPortfolioService(mockRepo)

		mockRepo.EXPECT().
			FindInstrumentBySymbol(ctx, "AAPL").
			Return(&domain.Instrument{
				ID:           instID,
				Symbol:       "AAPL",
				CurrencyCode: "USD",
			}, nil)

		mockRepo.EXPECT().
			UpsertInstrumentPrice(ctx, instID, priceDate, overridePrice, "manual: "+reason).
			Return(&domain.InstrumentPrice{
				InstrumentID: instID,
				Symbol:       "AAPL",
				PriceDate:    priceDate,
				Close:        overridePrice,
				CurrencyCode: "USD",
				Source:       "manual: " + reason,
			}, nil)

		mockRepo.EXPECT().
			FindPortfoliosHoldingInstrument(ctx, instID).
			Return([]uuid.UUID{uuid.New()}, nil)

		p, recomputed, err := svc.RecordPriceOverride(ctx, domain.PriceOverrideInput{
			Symbol:              "AAPL",
			PriceDate:           priceDate,
			Price:               overridePrice,
			Reason:              &reason,
			RecomputeValuations: true,
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !recomputed {
			t.Errorf("expected recomputed to be true")
		}
		if !p.Close.Equal(overridePrice) {
			t.Errorf("expected close price %s, got %s", overridePrice, p.Close)
		}
	})

	t.Run("error on non-positive price", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		svc := service.NewPortfolioService(mockRepo)

		_, _, err := svc.RecordPriceOverride(ctx, domain.PriceOverrideInput{
			Symbol:    "AAPL",
			PriceDate: priceDate,
			Price:     decimal.Zero,
		})
		if !errors.Is(err, service.ErrInvalidPrice) {
			t.Errorf("expected ErrInvalidPrice, got %v", err)
		}
	})

	t.Run("error on future price date", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		svc := service.NewPortfolioService(mockRepo)

		futureDate := time.Now().UTC().Add(48 * time.Hour)
		_, _, err := svc.RecordPriceOverride(ctx, domain.PriceOverrideInput{
			Symbol:    "AAPL",
			PriceDate: futureDate,
			Price:     overridePrice,
		})
		if !errors.Is(err, service.ErrFutureDate) {
			t.Errorf("expected ErrFutureDate, got %v", err)
		}
	})

	t.Run("error when instrument not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		svc := service.NewPortfolioService(mockRepo)

		mockRepo.EXPECT().
			FindInstrumentBySymbol(ctx, "UNKNOWN").
			Return(nil, repository.ErrInstrumentNotFound)

		_, _, err := svc.RecordPriceOverride(ctx, domain.PriceOverrideInput{
			Symbol:    "UNKNOWN",
			PriceDate: priceDate,
			Price:     overridePrice,
		})
		if !errors.Is(err, repository.ErrInstrumentNotFound) {
			t.Errorf("expected ErrInstrumentNotFound, got %v", err)
		}
	})
}

func TestPortfolioService_TriggerMarketSync(t *testing.T) {
	ctx := context.Background()

	t.Run("syncs all active instruments when no symbols specified", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		svc := service.NewPortfolioService(mockRepo)

		mockRepo.EXPECT().
			ListActiveInstruments(ctx).
			Return([]domain.Instrument{
				{Symbol: "AAPL"},
				{Symbol: "MSFT"},
				{Symbol: "NVDA"},
			}, nil)

		res, err := svc.TriggerMarketSync(ctx, nil, true)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !res.Success {
			t.Errorf("expected success to be true")
		}
		if res.PricesSynced != 3 {
			t.Errorf("expected 3 prices synced, got %d", res.PricesSynced)
		}
		if res.FXRatesSynced != 7 {
			t.Errorf("expected 7 fx rates synced, got %d", res.FXRatesSynced)
		}
	})
}

func TestPortfolioService_TriggerBackfill_Validation(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	now := time.Date(2025, 10, 1, 12, 0, 0, 0, time.UTC)
	svc := service.NewPortfolioService(mockRepo, service.WithNowFunc(func() time.Time { return now }))

	validFrom := time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)
	validTo := time.Date(2025, 9, 30, 0, 0, 0, 0, time.UTC)

	t.Run("rejects toDate before fromDate", func(t *testing.T) {
		_, err := svc.TriggerBackfill(ctx, domain.BackfillInput{
			FromDate:       validTo,
			ToDate:         validFrom,
			BackfillAssets: true,
		})
		if !errors.Is(err, service.ErrInvalidDateRange) {
			t.Errorf("expected ErrInvalidDateRange, got %v", err)
		}
	})

	t.Run("rejects fromDate in future", func(t *testing.T) {
		future := now.Add(24 * time.Hour)
		_, err := svc.TriggerBackfill(ctx, domain.BackfillInput{
			FromDate:       future,
			ToDate:         future.Add(24 * time.Hour),
			BackfillAssets: true,
		})
		if !errors.Is(err, service.ErrFutureDate) {
			t.Errorf("expected ErrFutureDate, got %v", err)
		}
	})

	t.Run("rejects toDate in future", func(t *testing.T) {
		future := now.Add(48 * time.Hour)
		_, err := svc.TriggerBackfill(ctx, domain.BackfillInput{
			FromDate:       validFrom,
			ToDate:         future,
			BackfillAssets: true,
		})
		if !errors.Is(err, service.ErrFutureDate) {
			t.Errorf("expected ErrFutureDate, got %v", err)
		}
	})

	t.Run("rejects zero dates", func(t *testing.T) {
		_, err := svc.TriggerBackfill(ctx, domain.BackfillInput{
			BackfillAssets: true,
		})
		if !errors.Is(err, service.ErrDateRequired) {
			t.Errorf("expected ErrDateRequired, got %v", err)
		}
	})

	t.Run("rejects empty targets", func(t *testing.T) {
		_, err := svc.TriggerBackfill(ctx, domain.BackfillInput{
			FromDate:       validFrom,
			ToDate:         validTo,
			BackfillAssets: false,
			BackfillFX:     false,
		})
		if !errors.Is(err, service.ErrNoBackfillTarget) {
			t.Errorf("expected ErrNoBackfillTarget, got %v", err)
		}
	})

	t.Run("rejects range over 5 years", func(t *testing.T) {
		tooOld := now.Add(-6 * 365 * 24 * time.Hour)
		_, err := svc.TriggerBackfill(ctx, domain.BackfillInput{
			FromDate:       tooOld,
			ToDate:         now,
			BackfillAssets: true,
		})
		if !errors.Is(err, service.ErrBackfillRangeTooLarge) {
			t.Errorf("expected ErrBackfillRangeTooLarge, got %v", err)
		}
	})
}

func TestPortfolioService_TriggerBackfill_SingleSymbol(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockIngestion := servicemocks.NewMockIngestionService(ctrl)
	now := time.Date(2025, 10, 1, 12, 0, 0, 0, time.UTC)
	svc := service.NewPortfolioService(mockRepo,
		service.WithIngestionService(mockIngestion),
		service.WithNowFunc(func() time.Time { return now }),
	)

	fromDate := time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)
	toDate := time.Date(2025, 9, 30, 0, 0, 0, 0, time.UTC)
	instID := uuid.MustParse("018f0000-0001-7000-8000-000000000001")

	mockRepo.EXPECT().
		FindInstrumentBySymbol(ctx, "AAPL").
		Return(&domain.Instrument{
			ID:           instID,
			Symbol:       "AAPL",
			ExchangeCode: "XNAS",
		}, nil)

	mockIngestion.EXPECT().
		BackfillInstrumentPrices(ctx, instID, "AAPL", "XNAS", fromDate, toDate).
		Return(22, nil)

	res, err := svc.TriggerBackfill(ctx, domain.BackfillInput{
		FromDate:       fromDate,
		ToDate:         toDate,
		Symbols:        []string{"AAPL"},
		BackfillAssets: true,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !res.Success {
		t.Errorf("expected success to be true")
	}
	if res.PricesSynced != 22 {
		t.Errorf("expected 22 prices synced, got %d", res.PricesSynced)
	}
	if res.FXRatesSynced != 0 {
		t.Errorf("expected 0 fx rates synced, got %d", res.FXRatesSynced)
	}
	if len(res.Warnings) != 0 {
		t.Errorf("expected 0 warnings, got %d", len(res.Warnings))
	}
}

func TestPortfolioService_TriggerBackfill_AllActiveAssets(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockIngestion := servicemocks.NewMockIngestionService(ctrl)
	now := time.Date(2025, 10, 1, 12, 0, 0, 0, time.UTC)
	svc := service.NewPortfolioService(mockRepo,
		service.WithIngestionService(mockIngestion),
		service.WithNowFunc(func() time.Time { return now }),
	)

	fromDate := time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)
	toDate := time.Date(2025, 9, 30, 0, 0, 0, 0, time.UTC)
	id1 := uuid.New()
	id2 := uuid.New()

	mockRepo.EXPECT().
		ListActiveInstruments(ctx).
		Return([]domain.Instrument{
			{ID: id1, Symbol: "AAPL", ExchangeCode: "XNAS"},
			{ID: id2, Symbol: "MSFT", ExchangeCode: "XNAS"},
		}, nil)

	mockIngestion.EXPECT().
		BackfillInstrumentPrices(ctx, id1, "AAPL", "XNAS", fromDate, toDate).
		Return(20, nil)

	mockIngestion.EXPECT().
		BackfillInstrumentPrices(ctx, id2, "MSFT", "XNAS", fromDate, toDate).
		Return(20, nil)

	res, err := svc.TriggerBackfill(ctx, domain.BackfillInput{
		FromDate:       fromDate,
		ToDate:         toDate,
		BackfillAssets: true,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !res.Success {
		t.Errorf("expected success to be true")
	}
	if res.PricesSynced != 40 {
		t.Errorf("expected 40 prices synced, got %d", res.PricesSynced)
	}
}

func TestPortfolioService_TriggerBackfill_FXPairs(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockIngestion := servicemocks.NewMockIngestionService(ctrl)
	now := time.Date(2025, 10, 1, 12, 0, 0, 0, time.UTC)
	svc := service.NewPortfolioService(mockRepo,
		service.WithIngestionService(mockIngestion),
		service.WithNowFunc(func() time.Time { return now }),
	)

	fromDate := time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)
	toDate := time.Date(2025, 9, 30, 0, 0, 0, 0, time.UTC)

	t.Run("specific currency pairs", func(t *testing.T) {
		mockIngestion.EXPECT().
			BackfillCurrencyPair(ctx, "EUR", "USD", fromDate, toDate).
			Return(21, nil)

		res, err := svc.TriggerBackfill(ctx, domain.BackfillInput{
			FromDate:      fromDate,
			ToDate:        toDate,
			CurrencyPairs: []string{"EUR/USD"},
			BackfillFX:    true,
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if res.FXRatesSynced != 21 {
			t.Errorf("expected 21 fx rates synced, got %d", res.FXRatesSynced)
		}
	})

	t.Run("all active currencies", func(t *testing.T) {
		mockRepo.EXPECT().
			ListActiveCurrencies(ctx).
			Return([]string{"EUR", "USD"}, nil)

		// Anchor EUR and USD: EUR/USD and USD/EUR
		mockIngestion.EXPECT().
			BackfillCurrencyPair(ctx, "USD", "EUR", fromDate, toDate).
			Return(21, nil)
		mockIngestion.EXPECT().
			BackfillCurrencyPair(ctx, "EUR", "USD", fromDate, toDate).
			Return(21, nil)

		res, err := svc.TriggerBackfill(ctx, domain.BackfillInput{
			FromDate:   fromDate,
			ToDate:     toDate,
			BackfillFX: true,
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if res.FXRatesSynced != 42 {
			t.Errorf("expected 42 fx rates synced, got %d", res.FXRatesSynced)
		}
	})
}

func TestPortfolioService_TriggerBackfill_PartialFailure(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockIngestion := servicemocks.NewMockIngestionService(ctrl)
	now := time.Date(2025, 10, 1, 12, 0, 0, 0, time.UTC)
	svc := service.NewPortfolioService(mockRepo,
		service.WithIngestionService(mockIngestion),
		service.WithNowFunc(func() time.Time { return now }),
	)

	fromDate := time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)
	toDate := time.Date(2025, 9, 30, 0, 0, 0, 0, time.UTC)
	id1 := uuid.New()
	id2 := uuid.New()

	mockRepo.EXPECT().
		FindInstrumentBySymbol(ctx, "AAPL").
		Return(&domain.Instrument{ID: id1, Symbol: "AAPL", ExchangeCode: "XNAS"}, nil)

	mockRepo.EXPECT().
		FindInstrumentBySymbol(ctx, "MSFT").
		Return(&domain.Instrument{ID: id2, Symbol: "MSFT", ExchangeCode: "XNAS"}, nil)

	mockIngestion.EXPECT().
		BackfillInstrumentPrices(ctx, id1, "AAPL", "XNAS", fromDate, toDate).
		Return(20, nil)

	mockIngestion.EXPECT().
		BackfillInstrumentPrices(ctx, id2, "MSFT", "XNAS", fromDate, toDate).
		Return(0, errors.New("provider rate limit exceeded"))

	res, err := svc.TriggerBackfill(ctx, domain.BackfillInput{
		FromDate:       fromDate,
		ToDate:         toDate,
		Symbols:        []string{"AAPL", "MSFT"},
		BackfillAssets: true,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !res.Success {
		t.Errorf("expected success to be true")
	}
	if res.PricesSynced != 20 {
		t.Errorf("expected 20 prices synced, got %d", res.PricesSynced)
	}
	if len(res.Warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(res.Warnings))
	}
}

func TestPortfolioService_TriggerBackfill_RecomputeValuations(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockIngestion := servicemocks.NewMockIngestionService(ctrl)
	mockValuations := servicemocks.NewMockValuationService(ctrl)
	now := time.Date(2025, 10, 1, 12, 0, 0, 0, time.UTC)
	svc := service.NewPortfolioService(mockRepo,
		service.WithIngestionService(mockIngestion),
		service.WithValuationService(mockValuations),
		service.WithNowFunc(func() time.Time { return now }),
	)

	fromDate := time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)
	toDate := time.Date(2025, 9, 30, 0, 0, 0, 0, time.UTC)
	instID := uuid.New()
	portfolioID := uuid.New()

	mockRepo.EXPECT().
		FindInstrumentBySymbol(ctx, "AAPL").
		Return(&domain.Instrument{ID: instID, Symbol: "AAPL", ExchangeCode: "XNAS"}, nil)

	mockIngestion.EXPECT().
		BackfillInstrumentPrices(ctx, instID, "AAPL", "XNAS", fromDate, toDate).
		Return(20, nil)

	mockRepo.EXPECT().
		FindPortfoliosHoldingInstrument(ctx, instID).
		Return([]uuid.UUID{portfolioID}, nil)

	mockValuations.EXPECT().
		BackfillPortfolioValuations(ctx, portfolioID, fromDate).
		Return(nil)

	res, err := svc.TriggerBackfill(ctx, domain.BackfillInput{
		FromDate:            fromDate,
		ToDate:              toDate,
		Symbols:             []string{"AAPL"},
		BackfillAssets:      true,
		RecomputeValuations: true,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !res.Success {
		t.Errorf("expected success to be true")
	}
	if res.PricesSynced != 20 {
		t.Errorf("expected 20 prices synced, got %d", res.PricesSynced)
	}
	if len(res.Warnings) != 0 {
		t.Errorf("expected 0 warnings, got %d", len(res.Warnings))
	}
}

func TestPortfolioService_ListMissingPriceTransactions(t *testing.T) {
	ctx := context.Background()

	t.Run("success returns report with items and bounds", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		svc := service.NewPortfolioService(mockRepo)

		fromDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		toDate := time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC)
		sym := "NVDA"

		expectedReport := &domain.MissingPriceReport{
			Items: []domain.MissingPriceItem{
				{
					TransactionID:  uuid.New(),
					PortfolioID:    uuid.New(),
					InstrumentID:   uuid.New(),
					Symbol:         "NVDA",
					AssetName:      "Nvidia Corp",
					ExchangeMIC:    "XNAS",
					TradeDate:      time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
					Type:           domain.TxTypeBuy,
					Quantity:       decimal.NewFromInt(10),
					ExecutionPrice: domain.NewMoney(decimal.NewFromInt(880), "USD"),
					CurrencyCode:   "USD",
				},
			},
			TotalCount:             1,
			UniqueInstrumentsCount: 1,
			EarliestTradeDate:      &fromDate,
			LatestTradeDate:        &toDate,
		}

		mockRepo.EXPECT().
			ListMissingPriceTransactions(ctx, domain.MissingPriceFilter{
				Symbol:   &sym,
				FromDate: &fromDate,
				ToDate:   &toDate,
				Limit:    50,
				Offset:   0,
			}).
			Return(expectedReport, nil)

		res, err := svc.ListMissingPriceTransactions(ctx, domain.MissingPriceFilter{
			Symbol:   &sym,
			FromDate: &fromDate,
			ToDate:   &toDate,
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if res.TotalCount != 1 {
			t.Errorf("expected total count 1, got %d", res.TotalCount)
		}
		if len(res.Items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(res.Items))
		}
		if res.Items[0].Symbol != "NVDA" {
			t.Errorf("expected NVDA, got %s", res.Items[0].Symbol)
		}
	})

	t.Run("normalizes ALL symbol and clamps limits", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		svc := service.NewPortfolioService(mockRepo)

		sym := "ALL"
		mockRepo.EXPECT().
			ListMissingPriceTransactions(ctx, domain.MissingPriceFilter{
				Symbol: nil,
				Limit:  200,
				Offset: 0,
			}).
			Return(&domain.MissingPriceReport{Items: []domain.MissingPriceItem{}}, nil)

		_, err := svc.ListMissingPriceTransactions(ctx, domain.MissingPriceFilter{
			Symbol: &sym,
			Limit:  500, // should clamp to 200
			Offset: -5,  // should clamp to 0
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("invalid date range returns error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		svc := service.NewPortfolioService(mockRepo)

		fromDate := time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC)
		toDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

		_, err := svc.ListMissingPriceTransactions(ctx, domain.MissingPriceFilter{
			FromDate: &fromDate,
			ToDate:   &toDate,
		})
		if !errors.Is(err, service.ErrInvalidDateRange) {
			t.Errorf("expected ErrInvalidDateRange, got %v", err)
		}
	})
}

func TestPortfolioService_ListMissingFXRateTransactions(t *testing.T) {
	ctx := context.Background()

	t.Run("success returns report with items and bounds", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		svc := service.NewPortfolioService(mockRepo)

		fromDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		toDate := time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC)
		pair := "USD/AUD"
		recRate := decimal.NewFromFloat(1.5280)
		sym := "AAPL"
		asset := "Apple Inc"

		expectedReport := &domain.MissingFXRatesReport{
			Items: []domain.MissingFXRateItem{
				{
					TransactionID:    uuid.New(),
					PortfolioID:      uuid.New(),
					PortfolioName:    "Growth Portfolio",
					PortfolioBaseCur: "AUD",
					TradeDate:        time.Date(2024, 2, 14, 0, 0, 0, 0, time.UTC),
					Type:             domain.TxTypeBuy,
					Symbol:           &sym,
					AssetName:        &asset,
					Amount:           domain.NewMoney(decimal.NewFromInt(3600), "USD"),
					TransactionCur:   "USD",
					RequiredPair:     "USD/AUD",
					RecordedFXRate:   &recRate,
				},
			},
			TotalCount:        1,
			UniquePairsCount:  1,
			EarliestTradeDate: &fromDate,
			LatestTradeDate:   &toDate,
		}

		mockRepo.EXPECT().
			ListMissingFXRateTransactions(ctx, domain.MissingFXRateFilter{
				CurrencyPair: &pair,
				FromDate:     &fromDate,
				ToDate:       &toDate,
				Limit:        50,
				Offset:       0,
			}).
			Return(expectedReport, nil)

		res, err := svc.ListMissingFXRateTransactions(ctx, domain.MissingFXRateFilter{
			CurrencyPair: &pair,
			FromDate:     &fromDate,
			ToDate:       &toDate,
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if res.TotalCount != 1 {
			t.Errorf("expected total count 1, got %d", res.TotalCount)
		}
		if len(res.Items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(res.Items))
		}
		if res.Items[0].RequiredPair != "USD/AUD" {
			t.Errorf("expected USD/AUD, got %s", res.Items[0].RequiredPair)
		}
	})

	t.Run("normalizes ALL currency pair and clamps limits", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		svc := service.NewPortfolioService(mockRepo)

		pair := "ALL"
		mockRepo.EXPECT().
			ListMissingFXRateTransactions(ctx, domain.MissingFXRateFilter{
				CurrencyPair: nil,
				Limit:        200,
				Offset:       0,
			}).
			Return(&domain.MissingFXRatesReport{Items: []domain.MissingFXRateItem{}}, nil)

		_, err := svc.ListMissingFXRateTransactions(ctx, domain.MissingFXRateFilter{
			CurrencyPair: &pair,
			Limit:        500, // should clamp to 200
			Offset:       -5,  // should clamp to 0
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("invalid date range returns error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		svc := service.NewPortfolioService(mockRepo)

		fromDate := time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC)
		toDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

		_, err := svc.ListMissingFXRateTransactions(ctx, domain.MissingFXRateFilter{
			FromDate: &fromDate,
			ToDate:   &toDate,
		})
		if !errors.Is(err, service.ErrInvalidDateRange) {
			t.Errorf("expected ErrInvalidDateRange, got %v", err)
		}
	})
}


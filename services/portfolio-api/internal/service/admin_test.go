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

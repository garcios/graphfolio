package service

import (
	"context"
	"testing"
	"time"

	"portfolio-api/internal/domain"
	"portfolio-api/internal/repository/mocks"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/mock/gomock"
)

func TestGetModifiedDietzReturn_SinglePeriod(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	portfolioID := uuid.New()
	userUUID := uuid.New()
	userID := userUUID.String()

	portfolio := &domain.Portfolio{
		ID:           portfolioID,
		UserID:       userUUID,
		Name:         "Main Portfolio",
		BaseCurrency: "USD",
	}

	svc := &portfolioService{
		repo:    mockRepo,
		nowFunc: func() time.Time { return time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC) },
	}

	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 0, 365) // 365 days

	mockRepo.EXPECT().
		FindPortfolioByUser(gomock.Any(), userID).
		Return(portfolio, nil)

	mockRepo.EXPECT().
		GetEarliestValuation(gomock.Any(), portfolioID).
		Return(&domain.PortfolioValuation{ValuationDate: startDate.AddDate(-1, 0, 0)}, nil)

	mockRepo.EXPECT().
		GetExternalCashFlows(gomock.Any(), portfolioID, startDate, endDate, "USD").
		Return([]domain.ExternalCashFlow{
			{
				TransactionID: "flow-1",
				Date:          startDate.AddDate(0, 0, 181),
				Type:          "DEPOSIT",
				AmountBase:    decimal.NewFromInt(50000),
				LocalAmount:   decimal.NewFromInt(50000),
				LocalCurrency: "USD",
			},
		}, nil)

	mockRepo.EXPECT().
		GetValuationAtOrBefore(gomock.Any(), portfolioID, startDate).
		Return(&domain.PortfolioValuation{
			PortfolioID:     portfolioID,
			ValuationDate:   startDate,
			MarketValueBase: decimal.NewFromInt(100000),
			CashValueBase:   decimal.Zero,
			TWRIndex:        decimal.NewFromInt(1),
		}, nil)

	mockRepo.EXPECT().
		GetValuationAtOrBefore(gomock.Any(), portfolioID, endDate).
		Return(&domain.PortfolioValuation{
			PortfolioID:     portfolioID,
			ValuationDate:   endDate,
			MarketValueBase: decimal.NewFromInt(180000),
			CashValueBase:   decimal.Zero,
			TWRIndex:        decimal.NewFromInt(1),
		}, nil)

	result, baseCurr, err := svc.GetModifiedDietzReturn(context.Background(), userID, startDate, endDate, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if baseCurr != "USD" {
		t.Errorf("expected base currency USD, got %s", baseCurr)
	}

	if result.ElapsedDays != 365 {
		t.Errorf("expected 365 days, got %d", result.ElapsedDays)
	}
	if !result.IsAnnualized {
		t.Errorf("expected isAnnualized=true")
	}
	if !result.CumulativeReturnPercent.Equal(decimal.NewFromFloat(23.96)) {
		t.Errorf("expected 23.96%% cum return, got %s", result.CumulativeReturnPercent)
	}
	if result.AnnualizedReturnPercent == nil || !result.AnnualizedReturnPercent.Equal(decimal.NewFromFloat(23.98)) {
		t.Errorf("expected 23.98%% ann return, got %v", result.AnnualizedReturnPercent)
	}
}

func TestGetModifiedDietzReturn_SubOneYear(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	portfolioID := uuid.New()
	userUUID := uuid.New()
	userID := userUUID.String()

	portfolio := &domain.Portfolio{
		ID:           portfolioID,
		UserID:       userUUID,
		Name:         "Main Portfolio",
		BaseCurrency: "AUD",
	}

	svc := &portfolioService{
		repo:    mockRepo,
		nowFunc: func() time.Time { return time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC) },
	}

	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 0, 181) // 181 days

	mockRepo.EXPECT().
		FindPortfolioByUser(gomock.Any(), userID).
		Return(portfolio, nil)

	mockRepo.EXPECT().
		GetEarliestValuation(gomock.Any(), portfolioID).
		Return(&domain.PortfolioValuation{ValuationDate: startDate.AddDate(-1, 0, 0)}, nil)

	mockRepo.EXPECT().
		GetExternalCashFlows(gomock.Any(), portfolioID, startDate, endDate, "AUD").
		Return([]domain.ExternalCashFlow{
			{
				TransactionID: "flow-aud",
				Date:          startDate.AddDate(0, 0, 73),
				Type:          "DEPOSIT",
				AmountBase:    decimal.NewFromInt(20000),
				LocalAmount:   decimal.NewFromInt(20000),
				LocalCurrency: "AUD",
			},
		}, nil)

	mockRepo.EXPECT().
		GetValuationAtOrBefore(gomock.Any(), portfolioID, startDate).
		Return(&domain.PortfolioValuation{
			PortfolioID:     portfolioID,
			ValuationDate:   startDate,
			MarketValueBase: decimal.NewFromInt(50000),
			CashValueBase:   decimal.Zero,
		}, nil)

	mockRepo.EXPECT().
		GetValuationAtOrBefore(gomock.Any(), portfolioID, endDate).
		Return(&domain.PortfolioValuation{
			PortfolioID:     portfolioID,
			ValuationDate:   endDate,
			MarketValueBase: decimal.NewFromInt(78000),
			CashValueBase:   decimal.Zero,
		}, nil)

	result, baseCurr, err := svc.GetModifiedDietzReturn(context.Background(), userID, startDate, endDate, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if baseCurr != "AUD" {
		t.Errorf("expected base currency AUD, got %s", baseCurr)
	}
	if result.IsAnnualized {
		t.Errorf("expected isAnnualized=false for sub-1-year window")
	}
	if result.AnnualizedReturnPercent != nil {
		t.Errorf("expected nil annualizedReturnPercent")
	}
	if !result.CumulativeReturnPercent.Equal(decimal.NewFromFloat(12.92)) {
		t.Errorf("expected 12.92%%, got %s", result.CumulativeReturnPercent)
	}
}

func TestGetModifiedDietzReturn_InvalidDate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	portfolioID := uuid.New()
	userUUID := uuid.New()
	userID := userUUID.String()

	portfolio := &domain.Portfolio{
		ID:           portfolioID,
		UserID:       userUUID,
		BaseCurrency: "USD",
	}

	svc := &portfolioService{
		repo: mockRepo,
	}

	startDate := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC) // Before start date

	mockRepo.EXPECT().
		FindPortfolioByUser(gomock.Any(), userID).
		Return(portfolio, nil)

	_, _, err := svc.GetModifiedDietzReturn(context.Background(), userID, startDate, endDate, false)
	if err == nil {
		t.Fatalf("expected error for endDate before startDate")
	}
}

func TestGetModifiedDietzReturn_PreInceptionClampingAndChaining(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	portfolioID := uuid.New()
	userUUID := uuid.New()
	userID := userUUID.String()

	portfolio := &domain.Portfolio{
		ID:           portfolioID,
		UserID:       userUUID,
		Name:         "Main Portfolio",
		BaseCurrency: "USD",
	}

	svc := &portfolioService{
		repo: mockRepo,
	}

	// User requests from 2024-01-01, but portfolio inception was 2024-07-01
	requestedStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	inceptionDate := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC) // 1 year from inception

	mockRepo.EXPECT().
		FindPortfolioByUser(gomock.Any(), userID).
		Return(portfolio, nil)

	mockRepo.EXPECT().
		GetEarliestValuation(gomock.Any(), portfolioID).
		Return(&domain.PortfolioValuation{ValuationDate: inceptionDate}, nil)

	// Since requestedStart is clamped to inceptionDate (2024-07-01), flows and intervals start at 2024-07-01
	mockRepo.EXPECT().
		GetExternalCashFlows(gomock.Any(), portfolioID, inceptionDate, endDate, "USD").
		Return([]domain.ExternalCashFlow{
			{
				TransactionID: "flow-init",
				Date:          inceptionDate,
				Type:          "DEPOSIT",
				AmountBase:    decimal.NewFromInt(10000),
				LocalAmount:   decimal.NewFromInt(10000),
				LocalCurrency: "USD",
			},
		}, nil)

	// Inception start valuation is 0
	// For each monthly interval end / start:
	mockRepo.EXPECT().
		GetValuationAtOrBefore(gomock.Any(), portfolioID, gomock.Any()).
		Return(&domain.PortfolioValuation{
			PortfolioID:     portfolioID,
			ValuationDate:   inceptionDate,
			MarketValueBase: decimal.NewFromInt(11000),
			CashValueBase:   decimal.Zero,
		}, nil).AnyTimes()

	result, baseCurr, err := svc.GetModifiedDietzReturn(context.Background(), userID, requestedStart, endDate, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if baseCurr != "USD" {
		t.Errorf("expected USD, got %s", baseCurr)
	}
	if !result.IsComputable {
		t.Errorf("expected computable result, got diagnostic: %s", result.DiagnosticCode)
	}
	if !result.IsAnnualized {
		t.Errorf("expected isAnnualized=true for 1 year active period, elapsed: %d", result.ElapsedDays)
	}
	if result.WeightedCapitalEmployed.IsZero() {
		t.Errorf("expected non-zero weighted capital employed")
	}
}

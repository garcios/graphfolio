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
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestPortfolioService_ListTaxLots(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)

	portfolioID := uuid.New()
	userUUID := uuid.New()
	userID := userUUID.String()

	mockPortfolio := &domain.Portfolio{
		ID:           portfolioID,
		UserID:       userUUID,
		BaseCurrency: "AUD",
	}

	instID := uuid.New()
	now := time.Date(2025, 4, 15, 0, 0, 0, 0, time.UTC)
	svc := service.NewPortfolioService(mockRepo, service.WithNowFunc(func() time.Time { return now }))

	lots := []domain.TaxLot{
		{
			ID:                uuid.New(),
			PortfolioID:       portfolioID,
			InstrumentID:      instID,
			Ticker:            "BHP",
			Name:              "BHP Group Ltd",
			AcquiredDate:      time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
			OriginalQuantity:  decimal.NewFromInt(100),
			RemainingQuantity: decimal.NewFromInt(100),
			CostBasisBase:     decimal.NewFromInt(4000),
		},
	}

	mockRepo.EXPECT().FindPortfolioByUser(gomock.Any(), userID).Return(mockPortfolio, nil)
	mockRepo.EXPECT().GetTaxLots(gomock.Any(), portfolioID, gomock.Any(), domain.TaxLotStatusOpen).Return(lots, nil)
	// Return a BUY date 10 days ago -> triggers wash sale risk
	buyDates := map[uuid.UUID][]time.Time{
		instID: {time.Date(2025, 4, 5, 0, 0, 0, 0, time.UTC)},
	}
	mockRepo.EXPECT().GetBuyTransactionDates(gomock.Any(), portfolioID).Return(buyDates, nil)

	res, err := svc.ListTaxLots(context.Background(), userID, nil, domain.TaxLotStatusOpen)
	assert.NoError(t, err)
	assert.Len(t, res, 1)
	assert.True(t, res[0].HasWashSaleRisk)
}

func TestPortfolioService_GetTaxReport(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)

	portfolioID := uuid.New()
	userUUID := uuid.New()
	userID := userUUID.String()

	mockPortfolio := &domain.Portfolio{
		ID:           portfolioID,
		UserID:       userUUID,
		BaseCurrency: "AUD",
	}

	now := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)
	svc := service.NewPortfolioService(mockRepo, service.WithNowFunc(func() time.Time { return now }))

	fy := int32(2026)
	// FY26 is 2025-07-01 to 2026-06-30
	mockRepo.EXPECT().FindPortfolioByUser(gomock.Any(), userID).Return(mockPortfolio, nil)

	disposals := []domain.LotDisposal{
		{
			ID:                    uuid.New(),
			RealizedPnLBase:       decimal.NewFromFloat(6064.99),
			IsCGTDiscountEligible: true,
			TermClassification:    domain.TermLongTerm,
			DisposedDate:          time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:                    uuid.New(),
			RealizedPnLBase:       decimal.NewFromFloat(-350.17),
			IsCGTDiscountEligible: false,
			TermClassification:    domain.TermShortTerm,
			DisposedDate:          time.Date(2025, 11, 15, 0, 0, 0, 0, time.UTC),
		},
	}
	mockRepo.EXPECT().GetRealizedDisposals(gomock.Any(), portfolioID, gomock.Any(), gomock.Any()).Return(disposals, nil)

	divIncome := decimal.NewFromFloat(2000.00)
	intIncome := decimal.NewFromFloat(966.80)
	frankingCredits := decimal.NewFromFloat(300.33)
	foreignTaxOffsets := decimal.NewFromFloat(324.18)
	mockRepo.EXPECT().GetTaxReportIncomeAndOffsets(gomock.Any(), portfolioID, gomock.Any(), gomock.Any()).
		Return(divIncome, intIncome, frankingCredits, foreignTaxOffsets, nil)

	mockRepo.EXPECT().GetTaxLots(gomock.Any(), portfolioID, nil, domain.TaxLotStatusOpen).Return([]domain.TaxLot{}, nil)
	mockRepo.EXPECT().GetBuyTransactionDates(gomock.Any(), portfolioID).Return(nil, nil)

	rep, err := svc.GetTaxReport(context.Background(), userID, &fy, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, "FY2025-2026", rep.FiscalYearLabel)
	assert.Equal(t, "AUD", rep.BaseCurrency)

	// Capital Gains:
	// Total: 6064.99, Losses: 350.17, CGT Discount: 2857.41, Net: 2857.41
	assert.True(t, decimal.NewFromFloat(6064.99).Equal(rep.Summary.CapitalGains.TotalCapitalGains))
	assert.True(t, decimal.NewFromFloat(350.17).Equal(rep.Summary.CapitalGains.LossesApplied))
	assert.True(t, decimal.NewFromFloat(2857.41).Equal(rep.Summary.CapitalGains.CGTDiscountAmount))
	assert.True(t, decimal.NewFromFloat(2857.41).Equal(rep.Summary.CapitalGains.NetCapitalGain))

	// Taxable Income:
	// Div + Int = 2966.80, Net CGT = 2857.41, Franking = 300.33, Total = 6124.54
	assert.True(t, decimal.NewFromFloat(2966.80).Equal(rep.Summary.TaxableIncome.DividendAndInterestIncome))
	assert.True(t, decimal.NewFromFloat(2857.41).Equal(rep.Summary.TaxableIncome.NetCapitalGains))
	assert.True(t, decimal.NewFromFloat(300.33).Equal(rep.Summary.TaxableIncome.FrankingCredits))
	assert.True(t, decimal.NewFromFloat(6124.54).Equal(rep.Summary.TaxableIncome.TotalAssessableIncome))
	assert.True(t, decimal.NewFromFloat(48.44).Equal(rep.Summary.TaxableIncome.DividendInterestRatioPct))
	assert.True(t, decimal.NewFromFloat(46.66).Equal(rep.Summary.TaxableIncome.CapitalGainsRatioPct))
	assert.True(t, decimal.NewFromFloat(4.90).Equal(rep.Summary.TaxableIncome.FrankingCreditsRatioPct))

	// Offsets:
	assert.True(t, decimal.NewFromFloat(300.33).Equal(rep.Summary.TaxOffsets.FrankingCreditsOffset))
	assert.True(t, decimal.NewFromFloat(324.18).Equal(rep.Summary.TaxOffsets.ForeignTaxOffsets))
	assert.True(t, decimal.NewFromFloat(624.51).Equal(rep.Summary.TaxOffsets.TotalAvailableOffsets))
}

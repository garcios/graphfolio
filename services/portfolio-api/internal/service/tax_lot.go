package service

import (
	"context"
	"fmt"
	"time"

	"portfolio-api/internal/domain"

	"github.com/google/uuid"
)

func (s *portfolioService) ListTaxLots(ctx context.Context, userID string, instrumentID *string, status domain.TaxLotStatus) ([]domain.TaxLot, error) {
	portfolio, err := s.repo.FindPortfolioByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("service: find portfolio: %w", err)
	}

	var instUUID *uuid.UUID
	if instrumentID != nil && *instrumentID != "" {
		parsed, err := uuid.Parse(*instrumentID)
		if err != nil {
			return nil, fmt.Errorf("service: invalid instrument ID: %w", err)
		}
		instUUID = &parsed
	}

	lots, err := s.repo.GetTaxLots(ctx, portfolio.ID, instUUID, status)
	if err != nil {
		return nil, fmt.Errorf("service: get tax lots: %w", err)
	}

	// Fetch buy dates to check wash sale risk
	buyDatesByInst, err := s.repo.GetBuyTransactionDates(ctx, portfolio.ID)
	if err != nil {
		// Non-fatal, default wash sale risk to false
		buyDatesByInst = make(map[uuid.UUID][]time.Time)
	}

	now := s.nowFunc()
	for i := range lots {
		buyDates := buyDatesByInst[lots[i].InstrumentID]
		lots[i].HasWashSaleRisk = domain.CheckWashSaleRisk(buyDates, now)
	}

	return lots, nil
}

func (s *portfolioService) GetRealizedGainsReport(ctx context.Context, userID string, fromDate, toDate *string, fiscalYear *int32) (*domain.RealizedGainsReport, error) {
	portfolio, err := s.repo.FindPortfolioByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("service: find portfolio: %w", err)
	}

	start, end, _, err := s.resolveTaxDateRange(fiscalYear, fromDate, toDate)
	if err != nil {
		return nil, fmt.Errorf("service: resolve date range: %w", err)
	}

	disposals, err := s.repo.GetRealizedDisposals(ctx, portfolio.ID, start, end)
	if err != nil {
		return nil, fmt.Errorf("service: get realized disposals: %w", err)
	}

	summary := domain.CalculateRealizedGainsSummary(disposals)

	return &domain.RealizedGainsReport{
		Summary:      summary,
		Disposals:    disposals,
		FromDate:     start.Format("2006-01-02"),
		ToDate:       end.Format("2006-01-02"),
		BaseCurrency: portfolio.BaseCurrency,
	}, nil
}

func (s *portfolioService) GetTaxReport(ctx context.Context, userID string, fiscalYear *int32, fromDate, toDate *string) (*domain.TaxReport, error) {
	portfolio, err := s.repo.FindPortfolioByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("service: find portfolio: %w", err)
	}

	start, end, fyLabel, err := s.resolveTaxDateRange(fiscalYear, fromDate, toDate)
	if err != nil {
		return nil, fmt.Errorf("service: resolve date range: %w", err)
	}

	// 1. Realized disposals and capital gains waterfall
	disposals, err := s.repo.GetRealizedDisposals(ctx, portfolio.ID, start, end)
	if err != nil {
		return nil, fmt.Errorf("service: get realized disposals: %w", err)
	}
	cgtSummary := domain.CalculateCapitalGainsWaterfall(disposals)

	// 2. Dividend & Interest income, franking credits, withholding taxes
	divIncome, intIncome, frankingCredits, foreignTaxOffsets, err := s.repo.GetTaxReportIncomeAndOffsets(ctx, portfolio.ID, start, end)
	if err != nil {
		return nil, fmt.Errorf("service: get income and offsets: %w", err)
	}

	combinedIncome := divIncome.Add(intIncome)
	taxableIncomeSummary := domain.CalculateTaxableIncomeSummary(combinedIncome, cgtSummary.NetCapitalGain, frankingCredits)

	taxOffsetsSummary := domain.TaxOffsetsSummary{
		FrankingCreditsOffset: frankingCredits.Round(2),
		ForeignTaxOffsets:     foreignTaxOffsets.Round(2),
		TotalAvailableOffsets: frankingCredits.Add(foreignTaxOffsets).Round(2),
	}

	// 3. Open lots with TLH and wash-sale analysis
	openLots, err := s.repo.GetTaxLots(ctx, portfolio.ID, nil, domain.TaxLotStatusOpen)
	if err != nil {
		return nil, fmt.Errorf("service: get open tax lots: %w", err)
	}

	buyDatesByInst, _ := s.repo.GetBuyTransactionDates(ctx, portfolio.ID)
	now := s.nowFunc()
	for i := range openLots {
		if buyDates, ok := buyDatesByInst[openLots[i].InstrumentID]; ok {
			openLots[i].HasWashSaleRisk = domain.CheckWashSaleRisk(buyDates, now)
		}
	}

	summary := domain.TaxReportSummary{
		TaxableIncome:   taxableIncomeSummary,
		CapitalGains:    cgtSummary,
		TaxOffsets:      taxOffsetsSummary,
		FiscalYearLabel: fyLabel,
		BaseCurrency:    portfolio.BaseCurrency,
	}

	return &domain.TaxReport{
		Summary:         summary,
		Disposals:       disposals,
		OpenLots:        openLots,
		FromDate:        start.Format("2006-01-02"),
		ToDate:          end.Format("2006-01-02"),
		FiscalYearLabel: fyLabel,
		BaseCurrency:    portfolio.BaseCurrency,
	}, nil
}

func (s *portfolioService) resolveTaxDateRange(fiscalYear *int32, fromDate, toDate *string) (start, end time.Time, label string, err error) {
	if fromDate != nil && *fromDate != "" && toDate != nil && *toDate != "" {
		start, err = time.Parse("2006-01-02", *fromDate)
		if err != nil {
			return time.Time{}, time.Time{}, "", fmt.Errorf("invalid from_date: %w", err)
		}
		end, err = time.Parse("2006-01-02", *toDate)
		if err != nil {
			return time.Time{}, time.Time{}, "", fmt.Errorf("invalid to_date: %w", err)
		}
		label = fmt.Sprintf("%s to %s", *fromDate, *toDate)
		return start, end, label, nil
	}

	var fy int32
	if fiscalYear != nil && *fiscalYear > 0 {
		fy = *fiscalYear
	} else {
		now := s.nowFunc()
		if now.Month() >= 7 {
			fy = int32(now.Year() + 1)
		} else {
			fy = int32(now.Year())
		}
	}

	label = fmt.Sprintf("FY%d-%d", fy-1, fy)
	start = time.Date(int(fy)-1, 7, 1, 0, 0, 0, 0, time.UTC)
	end = time.Date(int(fy), 6, 30, 23, 59, 59, 999999999, time.UTC)
	return start, end, label, nil
}

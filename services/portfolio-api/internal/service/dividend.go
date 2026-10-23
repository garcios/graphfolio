package service

import (
	"context"
	"fmt"
	"time"

	"portfolio-api/internal/domain"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func (s *portfolioService) GetDividendCalendar(ctx context.Context, userID string, year, month int) ([]domain.DividendCalendarEvent, error) {
	portfolio, err := s.repo.FindPortfolioByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("service: find portfolio failed: %w", err)
	}

	if month < 1 || month > 12 {
		return nil, fmt.Errorf("service: invalid month %d (must be 1..12)", month)
	}
	if year < 1970 || year > 2100 {
		return nil, fmt.Errorf("service: invalid year %d", year)
	}

	// Month boundary: from 1st of month to end of month, with 7 days padding
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	lastDay := firstDay.AddDate(0, 1, -1)
	fromDate := firstDay.AddDate(0, 0, -7)
	toDate := lastDay.AddDate(0, 0, 7)

	holdings, err := s.repo.GetHoldingsWithMarketData(ctx, portfolio.ID)
	if err != nil {
		return nil, fmt.Errorf("service: get holdings failed: %w", err)
	}

	if len(holdings) == 0 {
		return []domain.DividendCalendarEvent{}, nil
	}

	var instIDs []uuid.UUID
	holdingsMap := make(map[uuid.UUID]domain.HoldingWithPrice)
	for _, h := range holdings {
		instIDs = append(instIDs, h.InstrumentID)
		holdingsMap[h.InstrumentID] = h
	}

	schedules, err := s.repo.GetDividendCalendarSchedules(ctx, instIDs, fromDate, toDate)
	if err != nil {
		return nil, fmt.Errorf("service: get calendar schedules failed: %w", err)
	}

	fxRates, err := s.repo.GetCashFXRates(ctx, portfolio.BaseCurrency)
	if err != nil || fxRates == nil {
		fxRates = map[string]decimal.Decimal{portfolio.BaseCurrency: decimal.NewFromInt(1)}
	}

	hundred := decimal.NewFromInt(100)
	var events []domain.DividendCalendarEvent

	for _, sched := range schedules {
		h, exists := holdingsMap[sched.InstrumentID]
		if !exists {
			continue
		}

		entitledShares, err := s.repo.GetEntitledSharesForExDate(ctx, portfolio.ID, sched.InstrumentID, sched.ExDate)
		if err != nil {
			entitledShares = decimal.Zero
		}

		isEntitled := entitledShares.IsPositive()
		sharesForCalculation := entitledShares

		nowUTC := s.nowFunc().UTC()
		if !isEntitled && h.Quantity.IsPositive() && sched.PaymentDate.After(nowUTC) {
			// If not yet ex-date, investor currently holds shares and is expected to be entitled if held through ex-date
			if nowUTC.Before(sched.ExDate) {
				sharesForCalculation = h.Quantity
				isEntitled = true
			}
		}

		fx := decimal.NewFromInt(1)
		if h.FXRateToBase.IsPositive() && (sched.CurrencyCode == "" || sched.CurrencyCode == h.InstrumentCurrency) {
			fx = h.FXRateToBase
		} else if r, ok := fxRates[sched.CurrencyCode]; ok && r.IsPositive() {
			fx = r
		}

		grossBase := sharesForCalculation.Mul(sched.AmountPerShare).Mul(fx)
		whtRate := sched.WithholdingTaxPct.DivRound(hundred, 4)
		netDPS := sched.AmountPerShare.Mul(decimal.NewFromInt(1).Sub(whtRate))
		netBase := sharesForCalculation.Mul(netDPS).Mul(fx)

		events = append(events, domain.DividendCalendarEvent{
			ID:              sched.ID,
			InstrumentID:    sched.InstrumentID,
			Symbol:          h.Ticker,
			Name:            h.Name,
			ExDate:          sched.ExDate,
			RecordDate:      sched.RecordDate,
			PaymentDate:     sched.PaymentDate,
			SharesHeld:      sharesForCalculation,
			DPSDeclared:     domain.NewMoney(sched.AmountPerShare, sched.CurrencyCode),
			GrossAmountBase: domain.NewMoney(grossBase, portfolio.BaseCurrency),
			NetAmountBase:   domain.NewMoney(netBase, portfolio.BaseCurrency),
			Frequency:       sched.Frequency,
			Status:          sched.Status,
			IsUserEntitled:  isEntitled,
			IsEstimated:     sched.IsEstimated,
			FrankingPct:     sched.FrankingPercentage,
		})
	}

	return events, nil
}

func (s *portfolioService) GetDividendYieldAnalytics(ctx context.Context, userID string) (*domain.DividendYieldAnalytics, error) {
	portfolio, err := s.repo.FindPortfolioByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("service: find portfolio failed: %w", err)
	}

	holdings, err := s.repo.GetHoldingsWithMarketData(ctx, portfolio.ID)
	if err != nil {
		return nil, fmt.Errorf("service: get holdings failed: %w", err)
	}

	now := s.nowFunc().UTC()
	var activeInstIDs []uuid.UUID
	for _, h := range holdings {
		if h.Quantity.IsPositive() {
			activeInstIDs = append(activeInstIDs, h.InstrumentID)
		}
	}

	if len(activeInstIDs) == 0 {
		return domain.CalculateForwardYieldAnalytics(
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			map[string]decimal.Decimal{portfolio.BaseCurrency: decimal.NewFromInt(1)},
			portfolio.BaseCurrency,
			now,
		), nil
	}

	// Fetch forward 12-month schedules
	forwardEndDate := now.AddDate(1, 1, 0)
	forwardSchedules, err := s.repo.GetUpcomingDividendSchedules(ctx, activeInstIDs, now.AddDate(0, 0, -30), forwardEndDate)
	if err != nil {
		return nil, fmt.Errorf("service: get upcoming schedules failed: %w", err)
	}

	// Group latest schedule per instrument
	latestSchedules := make(map[uuid.UUID]domain.DividendSchedule)
	for _, sched := range forwardSchedules {
		if existing, ok := latestSchedules[sched.InstrumentID]; !ok || sched.PaymentDate.After(existing.PaymentDate) {
			latestSchedules[sched.InstrumentID] = sched
		}
	}
	// For instruments missing forward schedules, fallback to historical lookup
	for _, id := range activeInstIDs {
		if _, ok := latestSchedules[id]; !ok {
			hist, _ := s.repo.ListDividendSchedulesByInstrument(ctx, id)
			if len(hist) > 0 {
				latestSchedules[id] = hist[len(hist)-1]
			}
		}
	}

	// Fetch pending schedules where ex_date <= now and payment_date >= now
	var pendingSchedules []domain.DividendSchedule
	entitledMap := make(map[uuid.UUID]decimal.Decimal)
	for _, sched := range forwardSchedules {
		if !sched.ExDate.After(now) && sched.PaymentDate.After(now) {
			shares, err := s.repo.GetEntitledSharesForExDate(ctx, portfolio.ID, sched.InstrumentID, sched.ExDate)
			if err == nil && shares.IsPositive() {
				pendingSchedules = append(pendingSchedules, sched)
				entitledMap[sched.ID] = shares
			}
		}
	}

	// Historical schedules for 3Y CAGR
	historicalSchedules := make(map[uuid.UUID][]domain.DividendSchedule)
	for _, id := range activeInstIDs {
		hist, err := s.repo.ListDividendSchedulesByInstrument(ctx, id)
		if err == nil && len(hist) > 0 {
			historicalSchedules[id] = hist
		}
	}

	fxRates, err := s.repo.GetCashFXRates(ctx, portfolio.BaseCurrency)
	if err != nil || fxRates == nil {
		fxRates = map[string]decimal.Decimal{portfolio.BaseCurrency: decimal.NewFromInt(1)}
	}

	analytics := domain.CalculateForwardYieldAnalytics(
		holdings,
		latestSchedules,
		forwardSchedules,
		historicalSchedules,
		pendingSchedules,
		entitledMap,
		fxRates,
		portfolio.BaseCurrency,
		now,
	)

	return analytics, nil
}

func (s *portfolioService) SimulateDRIP(ctx context.Context, userID string, input domain.DRIPSimulationInput) (*domain.DRIPSimulationResult, error) {
	portfolio, err := s.repo.FindPortfolioByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("service: find portfolio failed: %w", err)
	}

	analytics, err := s.GetDividendYieldAnalytics(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("service: get yield analytics for DRIP failed: %w", err)
	}

	// Calculate starting portfolio value from current active holdings (Invested Assets)
	startVal := decimal.Zero
	for _, h := range analytics.HoldingsYield {
		startVal = startVal.Add(h.CurrentMarketValue.Amount)
	}

	// Fallback to latest valuation snapshot only if active holdings market value is zero
	if !startVal.IsPositive() {
		if val, err := s.repo.GetLatestValuation(ctx, portfolio.ID); err == nil && val != nil && val.MarketValueBase.IsPositive() {
			startVal = val.MarketValueBase
		}
	}

	startVal = startVal.Round(2)
	startDiv := analytics.AnnualForwardIncome.Amount

	result := domain.SimulateDRIPCompounding(
		startVal,
		startDiv,
		input,
		portfolio.BaseCurrency,
	)

	return result, nil
}

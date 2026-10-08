package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"portfolio-api/internal/domain"
	"portfolio-api/internal/repository"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ValuationService defines operations for daily portfolio valuation snapshots,
// time-weighted return (TWR) tracking, and historical valuation backfill.
//
//go:generate go run go.uber.org/mock/mockgen -destination=mocks/mock_valuation_service.go -package=mocks portfolio-api/internal/service ValuationService
type ValuationService interface {
	SnapshotValuation(ctx context.Context, portfolioID uuid.UUID, asOfDate time.Time) (*domain.PortfolioValuationSnapshot, error)
	BackfillPortfolioValuations(ctx context.Context, portfolioID uuid.UUID, fromDate time.Time) error
	RunDailyValuationJob(ctx context.Context, asOfDate time.Time) error
}

type ValuationServiceOption func(*valuationService)

// WithValuationNowFunc configures a custom current-time generator for testing.
func WithValuationNowFunc(fn func() time.Time) ValuationServiceOption {
	return func(s *valuationService) {
		s.nowFunc = fn
	}
}

type valuationService struct {
	repo    repository.Repository
	nowFunc func() time.Time
}

// NewValuationService constructs a new valuation and backfill service instance.
func NewValuationService(repo repository.Repository, opts ...ValuationServiceOption) ValuationService {
	s := &valuationService{
		repo:    repo,
		nowFunc: func() time.Time { return time.Now().UTC() },
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// SnapshotValuation computes and stores the daily valuation snapshot for a single date.
func (s *valuationService) SnapshotValuation(ctx context.Context, portfolioID uuid.UUID, asOfDate time.Time) (*domain.PortfolioValuationSnapshot, error) {
	asOf := time.Date(asOfDate.Year(), asOfDate.Month(), asOfDate.Day(), 0, 0, 0, 0, time.UTC)

	portfolio, err := s.repo.FindPortfolioByUser(ctx, portfolioID.String())
	if err != nil {
		return nil, fmt.Errorf("valuation service: find portfolio %s: %w", portfolioID, err)
	}

	txs, err := s.repo.GetTransactions(ctx, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("valuation service: get transactions for portfolio %s: %w", portfolioID, err)
	}

	// Filter transactions on or before asOfDate
	var txsUpToDate []domain.Transaction
	for _, tx := range txs {
		txDate := time.Date(tx.TradeDate.Year(), tx.TradeDate.Month(), tx.TradeDate.Day(), 0, 0, 0, 0, time.UTC)
		if !txDate.After(asOf) {
			txsUpToDate = append(txsUpToDate, tx)
		}
	}

	// If no transactions have occurred up to asOfDate
	if len(txsUpToDate) == 0 {
		prevVal, err := s.repo.GetLatestValuationBefore(ctx, portfolioID, asOf)
		if err != nil {
			return nil, fmt.Errorf("valuation service: get latest valuation before %s: %w", asOf.Format("2006-01-02"), err)
		}
		twrIndex := domain.TWRInceptionBaseline
		if prevVal != nil && prevVal.TWRIndex.IsPositive() {
			twrIndex = prevVal.TWRIndex
		}
		zeroRet := decimal.Zero
		snap := &domain.PortfolioValuationSnapshot{
			PortfolioID:     portfolioID,
			ValuationDate:   asOf,
			MarketValueBase: decimal.Zero,
			CashValueBase:   decimal.Zero,
			NetFlowBase:     decimal.Zero,
			DailyReturn:     &zeroRet,
			TWRIndex:        twrIndex,
		}
		if err := s.repo.UpsertValuationsBatch(ctx, []domain.PortfolioValuation{snap.ToPortfolioValuation()}); err != nil {
			return nil, fmt.Errorf("valuation service: upsert zero valuation snapshot: %w", err)
		}
		return snap, nil
	}

	seenInsts := make(map[uuid.UUID]bool)
	var instIDs []uuid.UUID
	seenCurrs := make(map[string]bool)
	var currencies []string
	instCurrencies := make(map[uuid.UUID]string)

	for _, tx := range txsUpToDate {
		if tx.InstrumentID != nil {
			if !seenInsts[*tx.InstrumentID] {
				seenInsts[*tx.InstrumentID] = true
				instIDs = append(instIDs, *tx.InstrumentID)
			}
			if tx.CurrencyCode != "" {
				instCurrencies[*tx.InstrumentID] = tx.CurrencyCode
			}
		}
		if tx.CurrencyCode != "" && !seenCurrs[tx.CurrencyCode] {
			seenCurrs[tx.CurrencyCode] = true
			currencies = append(currencies, tx.CurrencyCode)
		}
	}

	cas, err := s.repo.GetCorporateActions(ctx, instIDs)
	if err != nil {
		return nil, fmt.Errorf("valuation service: get corporate actions: %w", err)
	}

	_, _, holdings, cash, err := ProcessLedger(*portfolio, txsUpToDate, cas)
	if err != nil {
		return nil, fmt.Errorf("valuation service: process ledger: %w", err)
	}

	priceMatrix, err := s.repo.GetHistoricalPriceMatrix(ctx, instIDs, asOf, asOf)
	if err != nil {
		return nil, fmt.Errorf("valuation service: get price matrix for %s: %w", asOf.Format("2006-01-02"), err)
	}

	fxMatrix, err := s.repo.GetHistoricalFXMatrix(ctx, currencies, portfolio.BaseCurrency, asOf, asOf)
	if err != nil {
		return nil, fmt.Errorf("valuation service: get fx matrix for %s: %w", asOf.Format("2006-01-02"), err)
	}

	asOfStr := asOf.Format("2006-01-02")

	// Calculate Market Value in Base Currency
	marketValBase := decimal.Zero
	for _, h := range holdings {
		if h.Quantity.IsPositive() {
			price := decimal.Zero
			if prices, ok := priceMatrix[h.InstrumentID]; ok {
				price = prices[asOfStr]
			}
			curr := instCurrencies[h.InstrumentID]
			fx := decimal.NewFromInt(1)
			if curr != "" && curr != portfolio.BaseCurrency {
				if rates, ok := fxMatrix[curr]; ok && rates[asOfStr].IsPositive() {
					fx = rates[asOfStr]
				}
			}
			holdingVal := domain.CalculateHoldingMarketValue(h.Quantity, price, fx)
			marketValBase = marketValBase.Add(holdingVal)
		}
	}

	// Calculate Cash Value in Base Currency
	cashValBase := decimal.Zero
	for _, c := range cash {
		fx := decimal.NewFromInt(1)
		if c.CurrencyCode != portfolio.BaseCurrency {
			if rates, ok := fxMatrix[c.CurrencyCode]; ok && rates[asOfStr].IsPositive() {
				fx = rates[asOfStr]
			}
		}
		cVal := domain.CalculateCashValue(c.Balance, fx)
		cashValBase = cashValBase.Add(cVal)
	}

	totalVal := domain.CalculateTotalValue(marketValBase, cashValBase)

	// Calculate Net External Flow on asOfDate
	netFlowBase := decimal.Zero
	for _, tx := range txsUpToDate {
		txDate := time.Date(tx.TradeDate.Year(), tx.TradeDate.Month(), tx.TradeDate.Day(), 0, 0, 0, 0, time.UTC)
		if txDate.Equal(asOf) {
			fx := decimal.NewFromInt(1)
			if tx.CurrencyCode != portfolio.BaseCurrency {
				if rates, ok := fxMatrix[tx.CurrencyCode]; ok && rates[asOfStr].IsPositive() {
					fx = rates[asOfStr]
				}
			}
			if tx.Type == domain.TxTypeDeposit {
				netFlowBase = netFlowBase.Add(tx.Amount.Mul(fx))
			} else if tx.Type == domain.TxTypeWithdrawal {
				netFlowBase = netFlowBase.Sub(tx.Amount.Mul(fx))
			}
		}
	}

	prevVal, err := s.repo.GetLatestValuationBefore(ctx, portfolioID, asOf)
	if err != nil {
		return nil, fmt.Errorf("valuation service: get latest valuation before %s: %w", asOfStr, err)
	}

	prevTotal := decimal.Zero
	prevTWR := domain.TWRInceptionBaseline
	if prevVal != nil {
		prevTotal = prevVal.MarketValueBase.Add(prevVal.CashValueBase)
		if prevVal.TWRIndex.IsPositive() {
			prevTWR = prevVal.TWRIndex
		}
	}

	dailyRet := domain.CalculateDailyReturn(totalVal, prevTotal, netFlowBase)
	twrIndex := domain.ChainTWR(prevTWR, dailyRet)

	snap := &domain.PortfolioValuationSnapshot{
		PortfolioID:     portfolioID,
		ValuationDate:   asOf,
		MarketValueBase: marketValBase.Round(domain.ValuationAmountScale),
		CashValueBase:   cashValBase.Round(domain.ValuationAmountScale),
		NetFlowBase:     netFlowBase.Round(domain.ValuationAmountScale),
		DailyReturn:     dailyRet,
		TWRIndex:        twrIndex,
	}

	if err := s.repo.UpsertValuationsBatch(ctx, []domain.PortfolioValuation{snap.ToPortfolioValuation()}); err != nil {
		return nil, fmt.Errorf("valuation service: upsert snapshot for %s: %w", asOfStr, err)
	}

	return snap, nil
}

// BackfillPortfolioValuations replays history and recomputes all daily valuations from fromDate onwards.
func (s *valuationService) BackfillPortfolioValuations(ctx context.Context, portfolioID uuid.UUID, fromDate time.Time) error {
	from := time.Date(fromDate.Year(), fromDate.Month(), fromDate.Day(), 0, 0, 0, 0, time.UTC)
	now := s.nowFunc()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	portfolio, err := s.repo.FindPortfolioByUser(ctx, portfolioID.String())
	if err != nil {
		return fmt.Errorf("valuation service: find portfolio %s: %w", portfolioID, err)
	}

	txs, err := s.repo.GetTransactions(ctx, portfolioID)
	if err != nil {
		return fmt.Errorf("valuation service: get transactions for portfolio %s: %w", portfolioID, err)
	}

	if len(txs) == 0 {
		return s.repo.DeleteValuationsFromDate(ctx, portfolioID, from)
	}

	// 1. Identify earliest transaction date
	earliestTxDate := time.Date(txs[0].TradeDate.Year(), txs[0].TradeDate.Month(), txs[0].TradeDate.Day(), 0, 0, 0, 0, time.UTC)
	for _, tx := range txs {
		d := time.Date(tx.TradeDate.Year(), tx.TradeDate.Month(), tx.TradeDate.Day(), 0, 0, 0, 0, time.UTC)
		if d.Before(earliestTxDate) {
			earliestTxDate = d
		}
	}

	earliestDate := from
	if earliestTxDate.After(earliestDate) {
		earliestDate = earliestTxDate
	}
	if earliestDate.After(today) {
		earliestDate = today
	}

	// 2. Delete existing records from earliestDate onwards
	if err := s.repo.DeleteValuationsFromDate(ctx, portfolioID, earliestDate); err != nil {
		return fmt.Errorf("valuation service: delete valuations from %s: %w", earliestDate.Format("2006-01-02"), err)
	}

	// 3. Fetch latest valuation snapshot before earliestDate (for TWR chaining)
	prevVal, err := s.repo.GetLatestValuationBefore(ctx, portfolioID, earliestDate)
	if err != nil {
		return fmt.Errorf("valuation service: get latest valuation before %s: %w", earliestDate.Format("2006-01-02"), err)
	}

	runningTWR := domain.TWRInceptionBaseline
	runningTotalVal := decimal.Zero
	if prevVal != nil {
		if prevVal.TWRIndex.IsPositive() {
			runningTWR = prevVal.TWRIndex
		}
		runningTotalVal = prevVal.MarketValueBase.Add(prevVal.CashValueBase)
	}

	// 4. Collect all instrument IDs and currencies
	seenInsts := make(map[uuid.UUID]bool)
	var instIDs []uuid.UUID
	seenCurrs := make(map[string]bool)
	var currencies []string
	instCurrencies := make(map[uuid.UUID]string)

	for _, tx := range txs {
		if tx.InstrumentID != nil {
			if !seenInsts[*tx.InstrumentID] {
				seenInsts[*tx.InstrumentID] = true
				instIDs = append(instIDs, *tx.InstrumentID)
			}
			if tx.CurrencyCode != "" {
				instCurrencies[*tx.InstrumentID] = tx.CurrencyCode
			}
		}
		if tx.CurrencyCode != "" && !seenCurrs[tx.CurrencyCode] {
			seenCurrs[tx.CurrencyCode] = true
			currencies = append(currencies, tx.CurrencyCode)
		}
	}

	cas, err := s.repo.GetCorporateActions(ctx, instIDs)
	if err != nil {
		return fmt.Errorf("valuation service: get corporate actions: %w", err)
	}

	// 5. Pre-fetch historical market prices and FX matrix across [earliestDate, today]
	priceMatrix, err := s.repo.GetHistoricalPriceMatrix(ctx, instIDs, earliestDate, today)
	if err != nil {
		return fmt.Errorf("valuation service: get historical price matrix: %w", err)
	}

	fxMatrix, err := s.repo.GetHistoricalFXMatrix(ctx, currencies, portfolio.BaseCurrency, earliestDate, today)
	if err != nil {
		return fmt.Errorf("valuation service: get historical fx matrix: %w", err)
	}

	// 6. Step day-by-day: earliestDate -> today
	var buffer []domain.PortfolioValuation
	for currDate := earliestDate; !currDate.After(today); currDate = currDate.AddDate(0, 0, 1) {
		dateStr := currDate.Format("2006-01-02")

		var txsUpToDate []domain.Transaction
		netFlowBase := decimal.Zero

		for _, tx := range txs {
			txDate := time.Date(tx.TradeDate.Year(), tx.TradeDate.Month(), tx.TradeDate.Day(), 0, 0, 0, 0, time.UTC)
			if !txDate.After(currDate) {
				txsUpToDate = append(txsUpToDate, tx)
			}
			if txDate.Equal(currDate) {
				fx := decimal.NewFromInt(1)
				if tx.CurrencyCode != portfolio.BaseCurrency {
					if rates, ok := fxMatrix[tx.CurrencyCode]; ok && rates[dateStr].IsPositive() {
						fx = rates[dateStr]
					}
				}
				if tx.Type == domain.TxTypeDeposit {
					netFlowBase = netFlowBase.Add(tx.Amount.Mul(fx))
				} else if tx.Type == domain.TxTypeWithdrawal {
					netFlowBase = netFlowBase.Sub(tx.Amount.Mul(fx))
				}
			}
		}

		_, _, holdings, cash, err := ProcessLedger(*portfolio, txsUpToDate, cas)
		if err != nil {
			return fmt.Errorf("valuation service: process ledger on %s: %w", dateStr, err)
		}

		marketValBase := decimal.Zero
		for _, h := range holdings {
			if h.Quantity.IsPositive() {
				price := decimal.Zero
				if prices, ok := priceMatrix[h.InstrumentID]; ok {
					price = prices[dateStr]
				}
				curr := instCurrencies[h.InstrumentID]
				fx := decimal.NewFromInt(1)
				if curr != "" && curr != portfolio.BaseCurrency {
					if rates, ok := fxMatrix[curr]; ok && rates[dateStr].IsPositive() {
						fx = rates[dateStr]
					}
				}
				holdingVal := domain.CalculateHoldingMarketValue(h.Quantity, price, fx)
				marketValBase = marketValBase.Add(holdingVal)
			}
		}

		cashValBase := decimal.Zero
		for _, c := range cash {
			fx := decimal.NewFromInt(1)
			if c.CurrencyCode != portfolio.BaseCurrency {
				if rates, ok := fxMatrix[c.CurrencyCode]; ok && rates[dateStr].IsPositive() {
					fx = rates[dateStr]
				}
			}
			cVal := domain.CalculateCashValue(c.Balance, fx)
			cashValBase = cashValBase.Add(cVal)
		}

		totalVal := domain.CalculateTotalValue(marketValBase, cashValBase)
		dailyRet := domain.CalculateDailyReturn(totalVal, runningTotalVal, netFlowBase)
		runningTWR = domain.ChainTWR(runningTWR, dailyRet)
		runningTotalVal = totalVal

		dailyReturnVal := decimal.Zero
		if dailyRet != nil {
			dailyReturnVal = *dailyRet
		}

		buffer = append(buffer, domain.PortfolioValuation{
			PortfolioID:     portfolioID,
			ValuationDate:   currDate,
			MarketValueBase: marketValBase.Round(domain.ValuationAmountScale),
			CashValueBase:   cashValBase.Round(domain.ValuationAmountScale),
			NetFlowBase:     netFlowBase.Round(domain.ValuationAmountScale),
			DailyReturn:     dailyReturnVal,
			TWRIndex:        runningTWR,
		})
	}

	// 7. Batch upsert buffer
	if len(buffer) > 0 {
		if err := s.repo.UpsertValuationsBatch(ctx, buffer); err != nil {
			return fmt.Errorf("valuation service: batch upsert valuations: %w", err)
		}
	}

	return nil
}

// RunDailyValuationJob runs the scheduled valuation snapshot across all active portfolios.
func (s *valuationService) RunDailyValuationJob(ctx context.Context, asOfDate time.Time) error {
	portfolioIDs, err := s.repo.ListActivePortfolios(ctx)
	if err != nil {
		return fmt.Errorf("valuation service: list active portfolios: %w", err)
	}

	var errs []string
	for _, pID := range portfolioIDs {
		if _, err := s.SnapshotValuation(ctx, pID, asOfDate); err != nil {
			errs = append(errs, fmt.Sprintf("portfolio %s: %v", pID, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("valuation service: errors during daily valuation job: %s", strings.Join(errs, "; "))
	}
	return nil
}

package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"portfolio-api/internal/domain"
	"portfolio-api/internal/repository"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func (s *portfolioService) AddTransaction(ctx context.Context, input domain.AddTransactionInput) (*domain.Transaction, *domain.PortfolioSummary, error) {
	userID := input.UserID
	if userID == "" {
		userID = "1"
	}

	portfolio, err := s.repo.FindPortfolioByUser(ctx, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("service: find portfolio: %w", err)
	}

	tradeDate := input.TradeDate
	if tradeDate.IsZero() {
		tradeDate = time.Now().UTC()
	}

	fee := decimal.Zero
	if input.Fee != nil {
		if input.Fee.IsNegative() {
			return nil, nil, fmt.Errorf("service: fee cannot be negative")
		}
		fee = *input.Fee
	}

	var instID *uuid.UUID
	var currencyCode string
	var fxRateToBase *decimal.Decimal

	switch input.Type {
	case domain.TxTypeBuy, domain.TxTypeSell:
		if input.Symbol == nil || *input.Symbol == "" {
			return nil, nil, fmt.Errorf("service: symbol is required for %s", input.Type)
		}
		if input.Quantity == nil || !input.Quantity.IsPositive() {
			return nil, nil, fmt.Errorf("service: positive quantity is required for %s", input.Type)
		}
		if input.Price == nil || input.Price.IsNegative() {
			return nil, nil, fmt.Errorf("service: price cannot be negative for %s", input.Type)
		}

		inst, err := s.repo.FindInstrumentBySymbol(ctx, *input.Symbol)
		if err != nil {
			return nil, nil, fmt.Errorf("service: find instrument %s: %w", *input.Symbol, err)
		}
		instID = &inst.ID
		currencyCode = inst.CurrencyCode

		amount := input.Quantity.Mul(*input.Price)
		if input.Amount != nil && input.Amount.IsPositive() {
			amount = *input.Amount
		}

		if inst.CurrencyCode != portfolio.BaseCurrency {
			rate, err := s.repo.GetFXRate(ctx, inst.CurrencyCode, portfolio.BaseCurrency)
			if err == nil && rate.IsPositive() {
				fxRateToBase = &rate
			}
		}

		feeCurrencyCode := currencyCode
		if input.FeeCurrencyCode != nil && *input.FeeCurrencyCode != "" {
			feeCurrencyCode = *input.FeeCurrencyCode
		}

		tx := domain.Transaction{
			PortfolioID:     portfolio.ID,
			InstrumentID:    instID,
			Type:            input.Type,
			TradeDate:       tradeDate,
			Quantity:        input.Quantity,
			Price:           input.Price,
			Amount:          amount,
			CurrencyCode:    currencyCode,
			Fee:             fee,
			FeeCurrencyCode: &feeCurrencyCode,
			FXRateToBase:    fxRateToBase,
			Notes:           input.Notes,
		}

		savedTx, err := s.repo.InsertTransaction(ctx, tx)
		if err != nil {
			return nil, nil, fmt.Errorf("service: insert transaction: %w", err)
		}

		if s.ingestion != nil && instID != nil {
			todayUTC := s.nowFunc().UTC().Truncate(24 * time.Hour)
			hasPrices, err := s.repo.HasPricesForRange(ctx, *instID, tradeDate, todayUTC)
			if err == nil && !hasPrices {
				_, _ = s.ingestion.BackfillInstrumentPrices(ctx, *instID, *input.Symbol, inst.ExchangeCode, tradeDate, todayUTC)
				if inst.CurrencyCode != portfolio.BaseCurrency {
					_, _ = s.ingestion.BackfillCurrencyPair(ctx, inst.CurrencyCode, portfolio.BaseCurrency, tradeDate, todayUTC)
				}
				if feeCurrencyCode != portfolio.BaseCurrency && feeCurrencyCode != inst.CurrencyCode {
					_, _ = s.ingestion.BackfillCurrencyPair(ctx, feeCurrencyCode, portfolio.BaseCurrency, tradeDate, todayUTC)
				}
			}
		}

		summary, err := s.postTransactionUpdate(ctx, userID, portfolio.ID, tradeDate)
		if err != nil {
			return nil, nil, err
		}

		return savedTx, summary, nil

	case domain.TxTypeDeposit, domain.TxTypeWithdrawal:
		if input.Amount == nil || !input.Amount.IsPositive() {
			return nil, nil, fmt.Errorf("service: positive amount is required for %s", input.Type)
		}

		currencyCode = portfolio.BaseCurrency
		if input.CurrencyCode != nil && *input.CurrencyCode != "" {
			currencyCode = *input.CurrencyCode
		}

		if currencyCode != portfolio.BaseCurrency {
			rate, err := s.repo.GetFXRate(ctx, currencyCode, portfolio.BaseCurrency)
			if err == nil && rate.IsPositive() {
				fxRateToBase = &rate
			}
		}

		feeCurrencyCode := currencyCode
		if input.FeeCurrencyCode != nil && *input.FeeCurrencyCode != "" {
			feeCurrencyCode = *input.FeeCurrencyCode
		}

		tx := domain.Transaction{
			PortfolioID:     portfolio.ID,
			Type:            input.Type,
			TradeDate:       tradeDate,
			Amount:          *input.Amount,
			CurrencyCode:    currencyCode,
			Fee:             fee,
			FeeCurrencyCode: &feeCurrencyCode,
			FXRateToBase:    fxRateToBase,
			Notes:           input.Notes,
		}

		savedTx, err := s.repo.InsertTransaction(ctx, tx)
		if err != nil {
			return nil, nil, fmt.Errorf("service: insert transaction: %w", err)
		}

		summary, err := s.postTransactionUpdate(ctx, userID, portfolio.ID, tradeDate)
		if err != nil {
			return nil, nil, err
		}

		return savedTx, summary, nil

	case domain.TxTypeDividend:
		if input.Symbol == nil || *input.Symbol == "" {
			return nil, nil, fmt.Errorf("service: symbol is required for DIVIDEND")
		}
		if input.Amount == nil || !input.Amount.IsPositive() {
			return nil, nil, fmt.Errorf("service: positive amount is required for DIVIDEND")
		}

		inst, err := s.repo.FindInstrumentBySymbol(ctx, *input.Symbol)
		if err != nil {
			return nil, nil, fmt.Errorf("service: find instrument %s: %w", *input.Symbol, err)
		}
		instID = &inst.ID
		currencyCode = inst.CurrencyCode
		if input.CurrencyCode != nil && *input.CurrencyCode != "" {
			currencyCode = *input.CurrencyCode
		}

		if currencyCode != portfolio.BaseCurrency {
			rate, err := s.repo.GetFXRate(ctx, currencyCode, portfolio.BaseCurrency)
			if err == nil && rate.IsPositive() {
				fxRateToBase = &rate
			}
		}

		feeCurrencyCode := currencyCode
		if input.FeeCurrencyCode != nil && *input.FeeCurrencyCode != "" {
			feeCurrencyCode = *input.FeeCurrencyCode
		}

		tx := domain.Transaction{
			PortfolioID:     portfolio.ID,
			InstrumentID:    instID,
			Type:            input.Type,
			TradeDate:       tradeDate,
			Amount:          *input.Amount,
			CurrencyCode:    currencyCode,
			Fee:             fee,
			FeeCurrencyCode: &feeCurrencyCode,
			FXRateToBase:    fxRateToBase,
			Notes:           input.Notes,
		}

		savedTx, err := s.repo.InsertTransaction(ctx, tx)
		if err != nil {
			return nil, nil, fmt.Errorf("service: insert transaction: %w", err)
		}

		summary, err := s.postTransactionUpdate(ctx, userID, portfolio.ID, tradeDate)
		if err != nil {
			return nil, nil, err
		}

		return savedTx, summary, nil

	case domain.TxTypeSplit:
		if input.Symbol == nil || *input.Symbol == "" {
			return nil, nil, fmt.Errorf("service: symbol is required for SPLIT")
		}
		if input.Quantity == nil || !input.Quantity.IsPositive() {
			return nil, nil, fmt.Errorf("service: positive split ratio/quantity is required for SPLIT")
		}

		inst, err := s.repo.FindInstrumentBySymbol(ctx, *input.Symbol)
		if err != nil {
			return nil, nil, fmt.Errorf("service: find instrument %s: %w", *input.Symbol, err)
		}
		instID = &inst.ID
		currencyCode = inst.CurrencyCode

		price := input.Price
		if price != nil && price.IsNegative() {
			return nil, nil, fmt.Errorf("service: price ratio cannot be negative for SPLIT")
		}

		tx := domain.Transaction{
			PortfolioID:  portfolio.ID,
			InstrumentID: instID,
			Type:         input.Type,
			TradeDate:    tradeDate,
			Quantity:     input.Quantity,
			Price:        price,
			Amount:       decimal.Zero,
			CurrencyCode: currencyCode,
			Fee:          fee,
			Notes:        input.Notes,
		}

		savedTx, err := s.repo.InsertTransaction(ctx, tx)
		if err != nil {
			return nil, nil, fmt.Errorf("service: insert transaction: %w", err)
		}

		summary, err := s.postTransactionUpdate(ctx, userID, portfolio.ID, tradeDate)
		if err != nil {
			return nil, nil, err
		}

		return savedTx, summary, nil

	default:
		return nil, nil, fmt.Errorf("service: unsupported transaction type: %s", input.Type)
	}
}

func (s *portfolioService) postTransactionUpdate(ctx context.Context, userID string, portfolioID uuid.UUID, tradeDate time.Time) (*domain.PortfolioSummary, error) {
	if err := s.RebuildProjections(ctx, userID); err != nil {
		return nil, fmt.Errorf("service: rebuild projections: %w", err)
	}

	if s.valuations != nil {
		todayUTC := s.nowFunc().UTC().Truncate(24 * time.Hour)
		tradeDateTruncated := tradeDate.UTC().Truncate(24 * time.Hour)
		if tradeDateTruncated.Before(todayUTC) {
			_ = s.valuations.BackfillPortfolioValuations(ctx, portfolioID, tradeDateTruncated)
		} else {
			_, _ = s.valuations.SnapshotValuation(ctx, portfolioID, todayUTC)
		}
	}

	summary, err := s.GetPortfolioSummary(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("service: get portfolio summary: %w", err)
	}

	return summary, nil
}

func (s *portfolioService) ListInstruments(ctx context.Context) ([]domain.Instrument, error) {
	return s.repo.ListActiveInstruments(ctx)
}

func (s *portfolioService) ListTransactions(ctx context.Context, userID string, filter domain.TransactionFilter) ([]domain.TransactionWithInstrument, int, error) {
	if userID == "" {
		userID = "1"
	}

	portfolio, err := s.repo.FindPortfolioByUser(ctx, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("service: find portfolio: %w", err)
	}

	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	} else if filter.PageSize > 100 {
		filter.PageSize = 100
	}

	if filter.Symbol != nil && *filter.Symbol != "" {
		upper := strings.ToUpper(strings.TrimSpace(*filter.Symbol))
		filter.Symbol = &upper
	}

	items, totalCount, err := s.repo.ListTransactions(ctx, portfolio.ID, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("service: list transactions: %w", err)
	}

	return items, totalCount, nil
}

func (s *portfolioService) DeleteTransaction(ctx context.Context, userID string, transactionID string) (*domain.PortfolioSummary, error) {
	if userID == "" {
		userID = "1"
	}

	txUUID, err := uuid.Parse(transactionID)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid transaction id %q", repository.ErrTransactionNotFound, transactionID)
	}

	portfolio, err := s.repo.FindPortfolioByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("service: find portfolio: %w", err)
	}

	if err := s.repo.DeleteTransaction(ctx, portfolio.ID, txUUID); err != nil {
		return nil, err
	}

	if err := s.RebuildProjections(ctx, userID); err != nil {
		return nil, fmt.Errorf("service: rebuild projections: %w", err)
	}

	if s.valuations != nil {
		_ = s.valuations.BackfillPortfolioValuations(ctx, portfolio.ID, portfolio.CreatedAt)
	}

	summary, err := s.GetPortfolioSummary(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("service: get portfolio summary: %w", err)
	}

	return summary, nil
}

func (s *portfolioService) CheckTransactionDuplicates(ctx context.Context, userID string, externalRefs []string) ([]string, error) {
	if userID == "" {
		userID = "1"
	}

	portfolio, err := s.repo.FindPortfolioByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("service: find portfolio: %w", err)
	}

	if len(externalRefs) == 0 {
		return []string{}, nil
	}

	existingRefs, err := s.repo.FindExistingExternalRefs(ctx, portfolio.ID, externalRefs)
	if err != nil {
		return nil, fmt.Errorf("service: check duplicates: %w", err)
	}
	if existingRefs == nil {
		existingRefs = []string{}
	}
	return existingRefs, nil
}

func (s *portfolioService) BatchImportTransactions(ctx context.Context, input domain.BatchImportInput) (*domain.BatchImportResult, error) {
	userID := input.UserID
	if userID == "" {
		userID = "1"
	}

	portfolio, err := s.repo.FindPortfolioByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("service: find portfolio: %w", err)
	}

	if len(input.Transactions) == 0 {
		summary, err := s.GetPortfolioSummary(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("service: get summary: %w", err)
		}
		return &domain.BatchImportResult{ImportedCount: 0, SkippedCount: 0, Portfolio: summary}, nil
	}

	// 1. Gather all external refs to check duplicates
	var refsToCheck []string
	for _, item := range input.Transactions {
		if item.ExternalRef != "" {
			refsToCheck = append(refsToCheck, item.ExternalRef)
		}
	}

	existingSet := make(map[string]bool)
	if len(refsToCheck) > 0 {
		existingRefs, err := s.repo.FindExistingExternalRefs(ctx, portfolio.ID, refsToCheck)
		if err != nil {
			return nil, fmt.Errorf("service: check duplicates: %w", err)
		}
		for _, ref := range existingRefs {
			existingSet[ref] = true
		}
	}

	var txsToInsert []domain.Transaction
	var earliestTradeDate time.Time
	skippedCount := 0

	// 2. Resolve instruments & build transactions
	for _, item := range input.Transactions {
		if input.SkipDuplicates && item.ExternalRef != "" && existingSet[item.ExternalRef] {
			skippedCount++
			continue
		}

		var instID *uuid.UUID
		currencyCode := portfolio.BaseCurrency
		var fxRate *decimal.Decimal

		if item.Symbol != "" {
			inst, err := s.repo.FindInstrumentBySymbol(ctx, item.Symbol)
			if err != nil {
				if errors.Is(err, repository.ErrInstrumentNotFound) {
					currCode := item.CurrencyCode
					if currCode == "" {
						currCode = portfolio.BaseCurrency
					}
					currCode = strings.ToUpper(strings.TrimSpace(currCode))

					sym := strings.ToUpper(strings.TrimSpace(item.Symbol))
					exchangeCode := "XNAS"
					if currCode == "AUD" {
						exchangeCode = "XASX"
						if !strings.HasSuffix(sym, ".AX") {
							sym = sym + ".AX"
						}
					}

					newInst, createErr := s.repo.CreateInstrument(ctx, domain.CreateInstrumentInput{
						Symbol:       sym,
						ExchangeCode: exchangeCode,
						Name:         strings.ToUpper(strings.TrimSpace(item.Symbol)),
						AssetClass:   "EQUITY",
						CurrencyCode: currCode,
					})
					if createErr != nil {
						if errors.Is(createErr, repository.ErrInstrumentConflict) {
							inst, err = s.repo.FindInstrumentBySymbol(ctx, item.Symbol)
							if err != nil {
								return nil, fmt.Errorf("service: instrument %s resolution failed: %w", item.Symbol, err)
							}
						} else {
							return nil, fmt.Errorf("service: auto-provision instrument %s failed: %w", item.Symbol, createErr)
						}
					} else {
						inst = newInst
					}
				} else {
					return nil, fmt.Errorf("service: instrument %s lookup failed: %w", item.Symbol, err)
				}
			}
			instID = &inst.ID
			currencyCode = inst.CurrencyCode

			if inst.CurrencyCode != portfolio.BaseCurrency {
				rate, err := s.repo.GetFXRate(ctx, inst.CurrencyCode, portfolio.BaseCurrency)
				if err == nil && rate.IsPositive() {
					fxRate = &rate
				}
			}
		} else if item.CurrencyCode != "" {
			currencyCode = strings.ToUpper(strings.TrimSpace(item.CurrencyCode))
			if currencyCode != portfolio.BaseCurrency {
				rate, err := s.repo.GetFXRate(ctx, currencyCode, portfolio.BaseCurrency)
				if err == nil && rate.IsPositive() {
					fxRate = &rate
				}
			}
		}

		qty := item.Quantity
		price := item.Price
		var qtyPtr *decimal.Decimal
		var pricePtr *decimal.Decimal
		if !qty.IsZero() {
			qtyPtr = &qty
		}
		if !price.IsZero() {
			pricePtr = &price
		}

		var extRefPtr *string
		if item.ExternalRef != "" {
			extRefCopy := item.ExternalRef
			extRefPtr = &extRefCopy
		}

		var notesPtr *string
		if item.Notes != "" {
			notesCopy := item.Notes
			notesPtr = &notesCopy
		}

		feeCurrencyCode := currencyCode
		if item.FeeCurrencyCode != nil && *item.FeeCurrencyCode != "" {
			feeCurrencyCode = *item.FeeCurrencyCode
		}

		tx := domain.Transaction{
			PortfolioID:     portfolio.ID,
			InstrumentID:    instID,
			Type:            item.Type,
			TradeDate:       item.TradeDate,
			SettleDate:      item.SettleDate,
			Quantity:        qtyPtr,
			Price:           pricePtr,
			Amount:          item.Amount,
			CurrencyCode:    currencyCode,
			Fee:             item.Fee,
			FeeCurrencyCode: &feeCurrencyCode,
			FXRateToBase:    fxRate,
			ExternalRef:     extRefPtr,
			Notes:           notesPtr,
		}

		txsToInsert = append(txsToInsert, tx)

		if earliestTradeDate.IsZero() || item.TradeDate.Before(earliestTradeDate) {
			earliestTradeDate = item.TradeDate
		}
	}

	// 3. Bulk insert to database
	insertedCount, err := s.repo.BatchInsertTransactions(ctx, txsToInsert)
	if err != nil {
		return nil, fmt.Errorf("service: batch insert transactions: %w", err)
	}

	// 4. Rebuild ledger projections ONCE for the entire batch
	if insertedCount > 0 {
		if err := s.RebuildProjections(ctx, userID); err != nil {
			return nil, fmt.Errorf("service: rebuild projections: %w", err)
		}

		// 5. Retroactively backfill daily portfolio valuations from earliest trade date
		if s.valuations != nil && !earliestTradeDate.IsZero() {
			todayUTC := s.nowFunc().UTC().Truncate(24 * time.Hour)
			earliestTruncated := earliestTradeDate.UTC().Truncate(24 * time.Hour)
			if earliestTruncated.Before(todayUTC) {
				_ = s.valuations.BackfillPortfolioValuations(ctx, portfolio.ID, earliestTruncated)
			} else {
				_, _ = s.valuations.SnapshotValuation(ctx, portfolio.ID, todayUTC)
			}
		}
	}

	summary, err := s.GetPortfolioSummary(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("service: get updated summary: %w", err)
	}

	return &domain.BatchImportResult{
		ImportedCount: insertedCount,
		SkippedCount:  skippedCount + (len(txsToInsert) - insertedCount),
		Portfolio:     summary,
	}, nil
}

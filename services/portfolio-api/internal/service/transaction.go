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

		tx := domain.Transaction{
			PortfolioID:  portfolio.ID,
			InstrumentID: instID,
			Type:         input.Type,
			TradeDate:    tradeDate,
			Quantity:     input.Quantity,
			Price:        input.Price,
			Amount:       amount,
			CurrencyCode: currencyCode,
			Fee:          fee,
			FXRateToBase: fxRateToBase,
			Notes:        input.Notes,
		}

		savedTx, err := s.repo.InsertTransaction(ctx, tx)
		if err != nil {
			return nil, nil, fmt.Errorf("service: insert transaction: %w", err)
		}

		if err := s.RebuildProjections(ctx, userID); err != nil {
			return nil, nil, fmt.Errorf("service: rebuild projections: %w", err)
		}

		summary, err := s.GetPortfolioSummary(ctx, userID)
		if err != nil {
			return nil, nil, fmt.Errorf("service: get portfolio summary: %w", err)
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

		tx := domain.Transaction{
			PortfolioID:  portfolio.ID,
			Type:         input.Type,
			TradeDate:    tradeDate,
			Amount:       *input.Amount,
			CurrencyCode: currencyCode,
			Fee:          fee,
			FXRateToBase: fxRateToBase,
			Notes:        input.Notes,
		}

		savedTx, err := s.repo.InsertTransaction(ctx, tx)
		if err != nil {
			return nil, nil, fmt.Errorf("service: insert transaction: %w", err)
		}

		if err := s.RebuildProjections(ctx, userID); err != nil {
			return nil, nil, fmt.Errorf("service: rebuild projections: %w", err)
		}

		summary, err := s.GetPortfolioSummary(ctx, userID)
		if err != nil {
			return nil, nil, fmt.Errorf("service: get portfolio summary: %w", err)
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

		tx := domain.Transaction{
			PortfolioID:  portfolio.ID,
			InstrumentID: instID,
			Type:         input.Type,
			TradeDate:    tradeDate,
			Amount:       *input.Amount,
			CurrencyCode: currencyCode,
			Fee:          fee,
			FXRateToBase: fxRateToBase,
			Notes:        input.Notes,
		}

		savedTx, err := s.repo.InsertTransaction(ctx, tx)
		if err != nil {
			return nil, nil, fmt.Errorf("service: insert transaction: %w", err)
		}

		if err := s.RebuildProjections(ctx, userID); err != nil {
			return nil, nil, fmt.Errorf("service: rebuild projections: %w", err)
		}

		summary, err := s.GetPortfolioSummary(ctx, userID)
		if err != nil {
			return nil, nil, fmt.Errorf("service: get portfolio summary: %w", err)
		}

		return savedTx, summary, nil

	default:
		return nil, nil, fmt.Errorf("service: unsupported transaction type: %s", input.Type)
	}
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

	summary, err := s.GetPortfolioSummary(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("service: get portfolio summary: %w", err)
	}

	return summary, nil
}

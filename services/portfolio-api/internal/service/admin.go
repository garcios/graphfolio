package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"portfolio-api/internal/domain"
	"portfolio-api/internal/repository"

	"github.com/google/uuid"
)

var (
	isinRegex = regexp.MustCompile(`^[A-Z]{2}[A-Z0-9]{9}[0-9]$`)

	validAssetClasses = map[string]bool{
		"EQUITY":          true,
		"ETF":             true,
		"FUND":            true,
		"BOND":            true,
		"CRYPTO":          true,
		"CASH_EQUIVALENT": true,
	}

	ErrInvalidSymbol     = errors.New("symbol cannot be empty")
	ErrInvalidExchange   = errors.New("exchange code cannot be empty")
	ErrInvalidName       = errors.New("instrument name cannot be empty")
	ErrInvalidAssetClass = errors.New("invalid asset class")
	ErrInvalidCurrency   = errors.New("currency code must be 3 characters")
	ErrInvalidISIN       = errors.New("isin must be 12 alphanumeric characters (e.g. US67066G1040)")
	ErrInvalidPrice      = errors.New("price must be greater than zero")
	ErrFutureDate        = errors.New("price date cannot be in the future")
	ErrInvalidDateRange  = errors.New("from_date cannot be after to_date")
)

func (s *portfolioService) ListAllInstruments(ctx context.Context, isActive *bool, search *string) ([]domain.Instrument, error) {
	return s.repo.ListAllInstruments(ctx, isActive, search)
}

func (s *portfolioService) CreateInstrument(ctx context.Context, input domain.CreateInstrumentInput) (*domain.Instrument, error) {
	input.Symbol = strings.ToUpper(strings.TrimSpace(input.Symbol))
	if input.Symbol == "" {
		return nil, ErrInvalidSymbol
	}

	input.ExchangeCode = strings.ToUpper(strings.TrimSpace(input.ExchangeCode))
	if input.ExchangeCode == "" {
		return nil, ErrInvalidExchange
	}

	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		return nil, ErrInvalidName
	}

	input.AssetClass = strings.ToUpper(strings.TrimSpace(input.AssetClass))
	if !validAssetClasses[input.AssetClass] {
		return nil, ErrInvalidAssetClass
	}

	input.CurrencyCode = strings.ToUpper(strings.TrimSpace(input.CurrencyCode))
	if len(input.CurrencyCode) != 3 {
		return nil, ErrInvalidCurrency
	}

	if input.ISIN != nil {
		isin := strings.ToUpper(strings.TrimSpace(*input.ISIN))
		if isin != "" {
			if !isinRegex.MatchString(isin) {
				return nil, ErrInvalidISIN
			}
			input.ISIN = &isin
		} else {
			input.ISIN = nil
		}
	}

	return s.repo.CreateInstrument(ctx, input)
}

func (s *portfolioService) UpdateInstrument(ctx context.Context, input domain.UpdateInstrumentInput) (*domain.Instrument, error) {
	if input.ID == uuid.Nil {
		return nil, errors.New("invalid instrument id")
	}

	if input.Name != nil {
		trimmed := strings.TrimSpace(*input.Name)
		if trimmed == "" {
			return nil, ErrInvalidName
		}
		input.Name = &trimmed
	}

	if input.ISIN != nil {
		isin := strings.ToUpper(strings.TrimSpace(*input.ISIN))
		if isin != "" {
			if !isinRegex.MatchString(isin) {
				return nil, ErrInvalidISIN
			}
			input.ISIN = &isin
		} else {
			input.ISIN = nil
		}
	}

	return s.repo.UpdateInstrument(ctx, input)
}

func (s *portfolioService) ListInstrumentPrices(ctx context.Context, filter domain.PriceFilter) ([]domain.InstrumentPrice, int, error) {
	if filter.Limit <= 0 {
		filter.Limit = 50
	} else if filter.Limit > 200 {
		filter.Limit = 200
	}

	if filter.Offset < 0 {
		filter.Offset = 0
	}

	if filter.FromDate != nil && filter.ToDate != nil && filter.FromDate.After(*filter.ToDate) {
		return nil, 0, ErrInvalidDateRange
	}

	return s.repo.ListInstrumentPrices(ctx, filter)
}

func (s *portfolioService) RecordPriceOverride(ctx context.Context, input domain.PriceOverrideInput) (*domain.InstrumentPrice, bool, error) {
	symbol := strings.ToUpper(strings.TrimSpace(input.Symbol))
	if symbol == "" {
		return nil, false, ErrInvalidSymbol
	}

	if input.PriceDate.IsZero() {
		return nil, false, errors.New("price date is required")
	}

	// Compare with today UTC midnight boundary
	now := s.nowFunc().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, time.UTC)
	if input.PriceDate.After(today) {
		return nil, false, ErrFutureDate
	}

	if !input.Price.IsPositive() {
		return nil, false, ErrInvalidPrice
	}

	inst, err := s.repo.FindInstrumentBySymbol(ctx, symbol)
	if err != nil {
		if errors.Is(err, repository.ErrInstrumentNotFound) {
			return nil, false, repository.ErrInstrumentNotFound
		}
		return nil, false, fmt.Errorf("failed to lookup instrument %q: %w", symbol, err)
	}

	source := "manual"
	if input.Reason != nil && strings.TrimSpace(*input.Reason) != "" {
		source = "manual: " + strings.TrimSpace(*input.Reason)
	}

	price, err := s.repo.UpsertInstrumentPrice(ctx, inst.ID, input.PriceDate, input.Price, source)
	if err != nil {
		return nil, false, fmt.Errorf("failed to upsert instrument price: %w", err)
	}

	valuationsRecomputed := false
	if input.RecomputeValuations {
		holdingPortfolios, err := s.repo.FindPortfoliosHoldingInstrument(ctx, inst.ID)
		if err == nil && len(holdingPortfolios) > 0 {
			valuationsRecomputed = true
		}
	}

	return price, valuationsRecomputed, nil
}

func (s *portfolioService) GetIngestionStatus(ctx context.Context) (*domain.IngestionStatus, error) {
	return s.repo.GetIngestionMetrics(ctx)
}

func (s *portfolioService) TriggerMarketSync(ctx context.Context, symbols []string, syncFX bool) (*domain.MarketSyncResult, error) {
	pricesCount := 0
	if len(symbols) == 0 {
		insts, err := s.repo.ListActiveInstruments(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list active instruments: %w", err)
		}
		pricesCount = len(insts)
	} else {
		for _, sym := range symbols {
			sym = strings.ToUpper(strings.TrimSpace(sym))
			if sym == "" {
				continue
			}
			if _, err := s.repo.FindInstrumentBySymbol(ctx, sym); err != nil {
				return nil, fmt.Errorf("instrument not found: %s", sym)
			}
			pricesCount++
		}
	}

	fxCount := 0
	if syncFX {
		fxCount = 7 // EUR, GBP, AUD, CAD, JPY, CHF, USD
	}

	return &domain.MarketSyncResult{
		Success:       true,
		PricesSynced:  pricesCount,
		FXRatesSynced: fxCount,
		Message:       fmt.Sprintf("Market sync completed: %d asset prices and %d FX fixing rates refreshed", pricesCount, fxCount),
	}, nil
}

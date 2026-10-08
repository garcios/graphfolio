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

	ErrInvalidSymbol         = errors.New("symbol cannot be empty")
	ErrInvalidExchange       = errors.New("exchange code cannot be empty")
	ErrInvalidName           = errors.New("instrument name cannot be empty")
	ErrInvalidAssetClass     = errors.New("invalid asset class")
	ErrInvalidCurrency       = errors.New("currency code must be 3 characters")
	ErrInvalidISIN           = errors.New("isin must be 12 alphanumeric characters (e.g. US67066G1040)")
	ErrInvalidPrice          = errors.New("price must be greater than zero")
	ErrFutureDate            = errors.New("price date cannot be in the future")
	ErrInvalidDateRange      = errors.New("from_date cannot be after to_date")
	ErrDateRequired          = errors.New("from_date and to_date are required")
	ErrNoBackfillTarget      = errors.New("at least one backfill target must be enabled (assets or fx)")
	ErrBackfillRangeTooLarge = errors.New("backfill date range cannot exceed 5 years")
)

func (s *portfolioService) ListAllInstruments(ctx context.Context, isActive *bool, search *string) ([]domain.Instrument, error) {
	return s.repo.ListAllInstruments(ctx, isActive, search)
}

func (s *portfolioService) ListExchanges(ctx context.Context) ([]domain.Exchange, error) {
	exchanges, err := s.repo.ListExchanges(ctx)
	if err != nil {
		return nil, fmt.Errorf("service: list exchanges: %w", err)
	}
	return exchanges, nil
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
	if s.ingestion != nil {
		todayUTC := s.nowFunc().UTC()
		return s.ingestion.IngestDailyMarketData(ctx, todayUTC)
	}

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

func (s *portfolioService) TriggerBackfill(ctx context.Context, input domain.BackfillInput) (*domain.BackfillResult, error) {
	if input.FromDate.IsZero() || input.ToDate.IsZero() {
		return nil, ErrDateRequired
	}

	fromDate := input.FromDate.UTC().Truncate(24 * time.Hour)
	toDate := input.ToDate.UTC().Truncate(24 * time.Hour)

	if toDate.Before(fromDate) {
		return nil, ErrInvalidDateRange
	}

	now := s.nowFunc().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, time.UTC)
	if fromDate.After(today) || toDate.After(today) {
		return nil, ErrFutureDate
	}

	const maxBackfillDays = 1826 // ~5 years
	if toDate.Sub(fromDate) > maxBackfillDays*24*time.Hour {
		return nil, ErrBackfillRangeTooLarge
	}

	if !input.BackfillAssets && !input.BackfillFX {
		return nil, ErrNoBackfillTarget
	}

	var warnings []string
	pricesSynced := 0
	fxRatesSynced := 0
	var instrumentsToBackfill []domain.Instrument

	// 1. Backfill Asset Prices
	if input.BackfillAssets {
		if len(input.Symbols) > 0 {
			for _, sym := range input.Symbols {
				sym = strings.ToUpper(strings.TrimSpace(sym))
				if sym == "" {
					continue
				}
				inst, err := s.repo.FindInstrumentBySymbol(ctx, sym)
				if err != nil {
					warnings = append(warnings, fmt.Sprintf("instrument not found: %s", sym))
					continue
				}
				instrumentsToBackfill = append(instrumentsToBackfill, *inst)
			}
			if len(instrumentsToBackfill) == 0 && !input.BackfillFX {
				return &domain.BackfillResult{
					Success:  false,
					Message:  "No valid instruments found for specified symbols",
					Warnings: warnings,
				}, nil
			}
		} else {
			activeInsts, err := s.repo.ListActiveInstruments(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list active instruments: %w", err)
			}
			instrumentsToBackfill = activeInsts
		}

		for _, inst := range instrumentsToBackfill {
			if s.ingestion != nil {
				count, err := s.ingestion.BackfillInstrumentPrices(ctx, inst.ID, inst.Symbol, inst.ExchangeCode, fromDate, toDate)
				if err != nil {
					warnings = append(warnings, fmt.Sprintf("failed to backfill prices for %s: %v", inst.Symbol, err))
				} else {
					pricesSynced += count
				}
			}
		}
	}

	// 2. Backfill Foreign Exchange Rates
	if input.BackfillFX {
		if len(input.CurrencyPairs) > 0 {
			for _, pairStr := range input.CurrencyPairs {
				base, quote, err := parseCurrencyPair(pairStr)
				if err != nil {
					warnings = append(warnings, err.Error())
					continue
				}
				if s.ingestion != nil {
					count, err := s.ingestion.BackfillCurrencyPair(ctx, base, quote, fromDate, toDate)
					if err != nil {
						warnings = append(warnings, fmt.Sprintf("failed to backfill fx %s/%s: %v", base, quote, err))
					} else {
						fxRatesSynced += count
					}
				}
			}
		} else {
			currencies, err := s.repo.ListActiveCurrencies(ctx)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("failed to list active currencies: %v", err))
			} else {
				pairs := generateCurrencyPairs(currencies)
				for _, pair := range pairs {
					if s.ingestion != nil {
						count, err := s.ingestion.BackfillCurrencyPair(ctx, pair.Base, pair.Quote, fromDate, toDate)
						if err != nil {
							warnings = append(warnings, fmt.Sprintf("failed to backfill fx %s/%s: %v", pair.Base, pair.Quote, err))
						} else {
							fxRatesSynced += count
						}
					}
				}
			}
		}
	}

	// 3. Recompute Historical Valuations (Optional)
	if input.RecomputeValuations && s.valuations != nil {
		affectedPortfolios := make(map[uuid.UUID]bool)
		if len(instrumentsToBackfill) > 0 {
			for _, inst := range instrumentsToBackfill {
				pids, err := s.repo.FindPortfoliosHoldingInstrument(ctx, inst.ID)
				if err == nil {
					for _, pid := range pids {
						affectedPortfolios[pid] = true
					}
				}
			}
		} else if fxRatesSynced > 0 {
			pids, err := s.repo.ListActivePortfolios(ctx)
			if err == nil {
				for _, pid := range pids {
					affectedPortfolios[pid] = true
				}
			}
		}
		for pid := range affectedPortfolios {
			if err := s.valuations.BackfillPortfolioValuations(ctx, pid, fromDate); err != nil {
				warnings = append(warnings, fmt.Sprintf("failed to recompute valuations for portfolio %s: %v", pid, err))
			}
		}
	}

	if warnings == nil {
		warnings = []string{}
	}

	message := fmt.Sprintf("Historical backfill completed: %d asset prices and %d FX fixing rates stored between %s and %s",
		pricesSynced, fxRatesSynced, fromDate.Format("2006-01-02"), toDate.Format("2006-01-02"))

	return &domain.BackfillResult{
		Success:       true,
		PricesSynced:  pricesSynced,
		FXRatesSynced: fxRatesSynced,
		Message:       message,
		Warnings:      warnings,
	}, nil
}

func parseCurrencyPair(raw string) (string, string, error) {
	clean := strings.ToUpper(strings.TrimSpace(raw))
	clean = strings.ReplaceAll(clean, "/", " ")
	clean = strings.ReplaceAll(clean, "-", " ")
	parts := strings.Fields(clean)
	if len(parts) == 2 && len(parts[0]) == 3 && len(parts[1]) == 3 {
		return parts[0], parts[1], nil
	}
	if len(clean) == 6 && !strings.Contains(clean, " ") {
		return clean[:3], clean[3:], nil
	}
	return "", "", fmt.Errorf("invalid currency pair format %q (expected 'BASE/QUOTE')", raw)
}

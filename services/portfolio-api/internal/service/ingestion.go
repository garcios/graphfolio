package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"portfolio-api/internal/domain"
	"portfolio-api/internal/marketdata"
	"portfolio-api/internal/repository"

	"github.com/google/uuid"
)

// IngestionService coordinates fetching external market closing prices and FX fixing rates,
// validating the data, and persisting records into the repository.
//
//go:generate go run go.uber.org/mock/mockgen -destination=mocks/mock_ingestion_service.go -package=mocks portfolio-api/internal/service IngestionService
type IngestionService interface {
	// IngestDailyMarketData coordinates daily price and FX updates for a specific trade date.
	IngestDailyMarketData(ctx context.Context, asOfDate time.Time) (*domain.MarketSyncResult, error)

	// BackfillInstrumentPrices pulls historical daily closes for an instrument across a date range.
	BackfillInstrumentPrices(ctx context.Context, instrumentID uuid.UUID, symbol, exchange string, fromDate, toDate time.Time) (int, error)

	// BackfillCurrencyPair pulls historical FX rates for a currency pair across a date range.
	BackfillCurrencyPair(ctx context.Context, base, quote string, fromDate, toDate time.Time) (int, error)
}

type ingestionService struct {
	repo          repository.Repository
	priceProvider marketdata.PriceProvider
	fxProvider    marketdata.FXRateProvider
	limiter       *marketdata.RateLimiter
}

// NewIngestionService constructs an IngestionService instance.
func NewIngestionService(
	repo repository.Repository,
	priceProvider marketdata.PriceProvider,
	fxProvider marketdata.FXRateProvider,
	limiter *marketdata.RateLimiter,
) IngestionService {
	return &ingestionService{
		repo:          repo,
		priceProvider: priceProvider,
		fxProvider:    fxProvider,
		limiter:       limiter,
	}
}

// IngestDailyMarketData coordinates daily price and FX updates for a specific trade date.
func (s *ingestionService) IngestDailyMarketData(ctx context.Context, asOfDate time.Time) (*domain.MarketSyncResult, error) {
	if asOfDate.IsZero() {
		asOfDate = time.Now().UTC()
	}
	asOfDate = asOfDate.UTC().Truncate(24 * time.Hour)

	pricesSynced := 0
	fxSynced := 0

	// 1. Ingest closing marks for active instruments
	if s.priceProvider != nil {
		activeInsts, err := s.repo.ListActiveInstruments(ctx)
		if err != nil {
			log.Printf("ingestion: failed to list active instruments: %v", err)
		} else if len(activeInsts) > 0 {
			refs := make([]marketdata.InstrumentRef, 0, len(activeInsts))
			for _, inst := range activeInsts {
				refs = append(refs, marketdata.InstrumentRef{
					ID:       inst.ID,
					Symbol:   inst.Symbol,
					Exchange: inst.ExchangeCode,
					Currency: inst.CurrencyCode,
				})
			}

			priceRecords, err := s.priceProvider.FetchClosingPrices(ctx, refs, asOfDate)
			if err != nil {
				log.Printf("ingestion: price provider failed for %s: %v", asOfDate.Format("2006-01-02"), err)
			} else {
				var validPrices []marketdata.PriceRecord
				for _, rec := range priceRecords {
					if err := marketdata.ValidatePriceRecord(rec); err == nil {
						validPrices = append(validPrices, rec)
					}
				}

				if len(validPrices) > 0 {
					if err := s.repo.BatchUpsertInstrumentPrices(ctx, validPrices); err != nil {
						log.Printf("ingestion: batch upsert instrument prices failed: %v", err)
					} else {
						pricesSynced = len(validPrices)
					}
				}
			}
		}
	}

	// 2. Ingest FX rates for active currencies
	if s.fxProvider != nil {
		activeCurrs, err := s.repo.ListActiveCurrencies(ctx)
		if err != nil {
			log.Printf("ingestion: failed to list active currencies: %v", err)
		} else if len(activeCurrs) > 0 {
			pairs := generateCurrencyPairs(activeCurrs)
			fxRecords, err := s.fxProvider.FetchFXRates(ctx, pairs, asOfDate)
			if err != nil {
				log.Printf("ingestion: fx provider failed for %s: %v", asOfDate.Format("2006-01-02"), err)
			} else {
				var validRates []marketdata.FXRecord
				for _, r := range fxRecords {
					if err := marketdata.ValidateFXRecord(r); err == nil {
						validRates = append(validRates, r)
					}
				}

				if len(validRates) > 0 {
					if err := s.repo.BatchUpsertFXRates(ctx, validRates); err != nil {
						log.Printf("ingestion: batch upsert fx rates failed: %v", err)
					} else {
						fxSynced = len(validRates)
					}
				}
			}
		}
	}

	return &domain.MarketSyncResult{
		Success:       true,
		PricesSynced:  pricesSynced,
		FXRatesSynced: fxSynced,
		Message: fmt.Sprintf("Market sync completed: %d asset prices and %d FX fixing rates refreshed for %s",
			pricesSynced, fxSynced, asOfDate.Format("2006-01-02")),
	}, nil
}

// BackfillInstrumentPrices pulls historical daily closes for an instrument across a date range.
func (s *ingestionService) BackfillInstrumentPrices(
	ctx context.Context,
	instrumentID uuid.UUID,
	symbol, exchange string,
	fromDate, toDate time.Time,
) (int, error) {
	if s.priceProvider == nil {
		return 0, nil
	}

	fromDate = fromDate.UTC().Truncate(24 * time.Hour)
	toDate = toDate.UTC().Truncate(24 * time.Hour)
	if toDate.Before(fromDate) {
		return 0, fmt.Errorf("to date %s cannot be before from date %s", toDate.Format("2006-01-02"), fromDate.Format("2006-01-02"))
	}

	ref := marketdata.InstrumentRef{
		ID:       instrumentID,
		Symbol:   symbol,
		Exchange: exchange,
	}

	records, err := s.priceProvider.FetchHistoricalPrices(ctx, ref, fromDate, toDate)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch historical prices for %s: %w", symbol, err)
	}

	var validRecords []marketdata.PriceRecord
	for _, r := range records {
		if r.InstrumentID == uuid.Nil {
			r.InstrumentID = instrumentID
		}
		if err := marketdata.ValidatePriceRecord(r); err == nil {
			validRecords = append(validRecords, r)
		}
	}

	if len(validRecords) == 0 {
		return 0, nil
	}

	if err := s.repo.BatchUpsertInstrumentPrices(ctx, validRecords); err != nil {
		return 0, fmt.Errorf("failed to persist historical prices for %s: %w", symbol, err)
	}

	return len(validRecords), nil
}

// BackfillCurrencyPair pulls historical FX rates for a currency pair across a date range.
func (s *ingestionService) BackfillCurrencyPair(
	ctx context.Context,
	base, quote string,
	fromDate, toDate time.Time,
) (int, error) {
	if s.fxProvider == nil {
		return 0, nil
	}

	base = strings.ToUpper(strings.TrimSpace(base))
	quote = strings.ToUpper(strings.TrimSpace(quote))
	if base == "" || quote == "" || base == quote {
		return 0, nil
	}

	fromDate = fromDate.UTC().Truncate(24 * time.Hour)
	toDate = toDate.UTC().Truncate(24 * time.Hour)
	if toDate.Before(fromDate) {
		return 0, fmt.Errorf("to date cannot be before from date")
	}

	pair := marketdata.CurrencyPair{Base: base, Quote: quote}
	records, err := s.fxProvider.FetchHistoricalFXRates(ctx, pair, fromDate, toDate)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch historical fx rates for %s/%s: %w", base, quote, err)
	}

	var validRecords []marketdata.FXRecord
	for _, r := range records {
		if err := marketdata.ValidateFXRecord(r); err == nil {
			validRecords = append(validRecords, r)
		}
	}

	if len(validRecords) == 0 {
		return 0, nil
	}

	if err := s.repo.BatchUpsertFXRates(ctx, validRecords); err != nil {
		return 0, fmt.Errorf("failed to persist historical fx rates for %s/%s: %w", base, quote, err)
	}

	return len(validRecords), nil
}

// generateCurrencyPairs constructs standard pairs against anchor currencies (USD and EUR).
func generateCurrencyPairs(currencies []string) []marketdata.CurrencyPair {
	anchors := []string{"USD", "EUR"}
	pairMap := make(map[string]marketdata.CurrencyPair)

	for _, c := range currencies {
		c = strings.ToUpper(strings.TrimSpace(c))
		if len(c) != 3 {
			continue
		}

		for _, anchor := range anchors {
			if c != anchor {
				// Anchor -> Currency
				key1 := fmt.Sprintf("%s/%s", anchor, c)
				pairMap[key1] = marketdata.CurrencyPair{Base: anchor, Quote: c}

				// Currency -> Anchor
				key2 := fmt.Sprintf("%s/%s", c, anchor)
				pairMap[key2] = marketdata.CurrencyPair{Base: c, Quote: anchor}
			}
		}
	}

	pairs := make([]marketdata.CurrencyPair, 0, len(pairMap))
	for _, p := range pairMap {
		pairs = append(pairs, p)
	}
	return pairs
}

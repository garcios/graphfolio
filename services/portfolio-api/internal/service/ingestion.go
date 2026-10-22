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

	// IngestDividendSchedules syncs upcoming forward schedules and cadence projections for active holdings or specific symbols.
	IngestDividendSchedules(ctx context.Context, options domain.DividendSyncOptions) (*domain.DividendSyncResult, error)

	// BackfillInstrumentDividends pulls historical and forward dividend schedules for a specific instrument.
	BackfillInstrumentDividends(ctx context.Context, instrumentID uuid.UUID, symbol, exchange, currency string) (int, error)
}

type IngestionOption func(*ingestionService)

func WithDividendProvider(p marketdata.DividendProvider) IngestionOption {
	return func(s *ingestionService) {
		s.dividendProvider = p
	}
}

func WithYahooDividendProvider(p marketdata.DividendProvider) IngestionOption {
	return func(s *ingestionService) {
		s.yahooDividend = p
	}
}

func WithTwelveDividendProvider(p marketdata.DividendProvider) IngestionOption {
	return func(s *ingestionService) {
		s.twelveDividend = p
	}
}

type ingestionService struct {
	repo             repository.Repository
	priceProvider    marketdata.PriceProvider
	fxProvider       marketdata.FXRateProvider
	limiter          *marketdata.RateLimiter
	dividendProvider marketdata.DividendProvider
	yahooDividend    marketdata.DividendProvider
	twelveDividend   marketdata.DividendProvider
}

// NewIngestionService constructs an IngestionService instance.
func NewIngestionService(
	repo repository.Repository,
	priceProvider marketdata.PriceProvider,
	fxProvider marketdata.FXRateProvider,
	limiter *marketdata.RateLimiter,
	opts ...IngestionOption,
) IngestionService {
	s := &ingestionService{
		repo:          repo,
		priceProvider: priceProvider,
		fxProvider:    fxProvider,
		limiter:       limiter,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// IngestDailyMarketData coordinates daily price and FX updates for a specific trade date.
func (s *ingestionService) IngestDailyMarketData(ctx context.Context, asOfDate time.Time) (*domain.MarketSyncResult, error) {
	if asOfDate.IsZero() {
		asOfDate = time.Now().UTC()
	}
	asOfDate = asOfDate.UTC().Truncate(24 * time.Hour)

	pricesSynced := 0
	fxSynced := 0

	// 1. Ingest closing marks for active instruments and benchmark instruments
	if s.priceProvider != nil {
		activeInsts, err := s.repo.ListActiveInstruments(ctx)
		if err != nil {
			log.Printf("ingestion: failed to list active instruments: %v", err)
		}
		benchInsts, err := s.repo.ListBenchmarkInstruments(ctx)
		if err != nil {
			log.Printf("ingestion: failed to list benchmark instruments: %v", err)
		}

		instMap := make(map[uuid.UUID]marketdata.InstrumentRef)
		for _, inst := range activeInsts {
			instMap[inst.ID] = marketdata.InstrumentRef{
				ID:       inst.ID,
				Symbol:   inst.Symbol,
				Exchange: inst.ExchangeCode,
				Currency: inst.CurrencyCode,
			}
		}
		for _, inst := range benchInsts {
			instMap[inst.ID] = marketdata.InstrumentRef{
				ID:            inst.ID,
				Symbol:        inst.Symbol,
				Exchange:      inst.ExchangeCode,
				Currency:      inst.CurrencyCode,
				IsTotalReturn: true,
			}
		}

		if len(instMap) > 0 {
			refs := make([]marketdata.InstrumentRef, 0, len(instMap))
			for _, ref := range instMap {
				refs = append(refs, ref)
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

	isTotalReturn := false
	if strings.EqualFold(symbol, "SP500TR") || strings.EqualFold(symbol, "VAS") || strings.EqualFold(symbol, "QQQ") {
		isTotalReturn = true
	}

	ref := marketdata.InstrumentRef{
		ID:            instrumentID,
		Symbol:        symbol,
		Exchange:      exchange,
		IsTotalReturn: isTotalReturn,
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

// IngestDividendSchedules syncs upcoming forward schedules and cadence projections for active holdings or specific symbols.
func (s *ingestionService) IngestDividendSchedules(ctx context.Context, options domain.DividendSyncOptions) (*domain.DividendSyncResult, error) {
	start := time.Now()

	// 1. Resolve Dividend Provider based on ForceProvider or default cascade
	provider := s.dividendProvider
	if strings.EqualFold(options.ForceProvider, "YAHOO") && s.yahooDividend != nil {
		provider = s.yahooDividend
	} else if (strings.EqualFold(options.ForceProvider, "TWELVE_DATA") || strings.EqualFold(options.ForceProvider, "TWELVEDATA")) && s.twelveDividend != nil {
		provider = s.twelveDividend
	}

	if provider == nil {
		return &domain.DividendSyncResult{
			Success: true,
			Message: "Dividend sync completed (no provider configured)",
		}, nil
	}

	// 2. Resolve target instruments
	var targets []marketdata.InstrumentRef
	if len(options.Symbols) > 0 {
		for _, sym := range options.Symbols {
			sym = strings.TrimSpace(sym)
			if sym == "" {
				continue
			}
			inst, err := s.repo.FindInstrumentBySymbol(ctx, sym)
			if err != nil {
				log.Printf("ingestion: instrument not found for symbol %s: %v", sym, err)
				continue
			}
			if options.ExchangeCode != nil && *options.ExchangeCode != "" && !strings.EqualFold(inst.ExchangeCode, *options.ExchangeCode) {
				continue
			}
			targets = append(targets, marketdata.InstrumentRef{
				ID:       inst.ID,
				Symbol:   inst.Symbol,
				Exchange: inst.ExchangeCode,
				Currency: inst.CurrencyCode,
			})
		}
	} else {
		activeInsts, err := s.repo.ListActiveInstruments(ctx)
		if err != nil {
			return nil, fmt.Errorf("ingestion: failed to list active instruments: %w", err)
		}
		for _, inst := range activeInsts {
			if options.ExchangeCode != nil && *options.ExchangeCode != "" && !strings.EqualFold(inst.ExchangeCode, *options.ExchangeCode) {
				continue
			}
			targets = append(targets, marketdata.InstrumentRef{
				ID:       inst.ID,
				Symbol:   inst.Symbol,
				Exchange: inst.ExchangeCode,
				Currency: inst.CurrencyCode,
			})
		}
	}

	if len(targets) == 0 {
		return &domain.DividendSyncResult{
			Success:    true,
			Message:    "No matching active instruments for dividend sync",
			DurationMs: time.Since(start).Milliseconds(),
		}, nil
	}

	totalSynced := 0
	primaryCount := 0
	fallbackCount := 0
	placeholdersUpgraded := 0
	var failedSymbols []string

	now := time.Now().UTC()

	for _, inst := range targets {
		upcoming, upErr := provider.GetUpcomingDividends(ctx, inst)
		if upErr != nil {
			log.Printf("ingestion: get upcoming dividends for %s failed: %v", inst.Symbol, upErr)
		}

		history, histErr := provider.GetHistoricalDividends(ctx, inst, now.AddDate(-5, 0, 0), now)
		if histErr != nil {
			log.Printf("ingestion: get historical dividends for %s failed: %v", inst.Symbol, histErr)
		}

		if upErr != nil && histErr != nil && len(upcoming) == 0 && len(history) == 0 {
			failedSymbols = append(failedSymbols, inst.Symbol)
			continue
		}

		allEvents := marketdata.SynthesizeForwardEstimates(inst, upcoming, history, now)
		if len(allEvents) == 0 {
			continue
		}

		var schedules []domain.DividendSchedule
		for _, ev := range allEvents {
			schedules = append(schedules, domain.DividendSchedule{
				ID:                 uuid.New(),
				InstrumentID:       inst.ID,
				ExDate:             ev.ExDate,
				RecordDate:         ev.RecordDate,
				PaymentDate:        ev.PaymentDate,
				AmountPerShare:     ev.AmountPerShare,
				CurrencyCode:       ev.CurrencyCode,
				Frequency:          domain.DividendFrequency(ev.Frequency),
				Status:             domain.DividendStatus(ev.Status),
				IsEstimated:        ev.IsEstimated,
				IsOverride:         false,
				FrankingPercentage: ev.FrankingPct,
				WithholdingTaxPct:  ev.WithholdingTaxPct,
				Source:             ev.Source,
			})

			if !ev.IsEstimated {
				if ev.Source == "yahoo" {
					primaryCount++
				} else if ev.Source == "twelvedata" {
					fallbackCount++
				}
			}
		}

		synced, upgraded, err := s.repo.UpsertDividendSchedulesBatch(ctx, schedules)
		if err != nil {
			log.Printf("ingestion: upsert dividend schedules for %s failed: %v", inst.Symbol, err)
			failedSymbols = append(failedSymbols, inst.Symbol)
			continue
		}

		totalSynced += synced
		placeholdersUpgraded += upgraded
	}

	return &domain.DividendSyncResult{
		Success:              len(failedSymbols) == 0,
		SchedulesSynced:      totalSynced,
		PrimarySourceCount:   primaryCount,
		FallbackSourceCount:  fallbackCount,
		PlaceholdersUpgraded: placeholdersUpgraded,
		Message: fmt.Sprintf("Dividend sync completed: %d schedules synced (%d via Yahoo, %d via Twelve Data fallback, %d estimates upgraded)",
			totalSynced, primaryCount, fallbackCount, placeholdersUpgraded),
		DurationMs:    time.Since(start).Milliseconds(),
		FailedSymbols: failedSymbols,
	}, nil
}

// BackfillInstrumentDividends pulls historical and forward dividend schedules for a specific instrument.
func (s *ingestionService) BackfillInstrumentDividends(ctx context.Context, instrumentID uuid.UUID, symbol, exchange, currency string) (int, error) {
	opts := domain.DividendSyncOptions{
		Symbols: []string{symbol},
	}
	res, err := s.IngestDividendSchedules(ctx, opts)
	if err != nil {
		return 0, err
	}
	return res.SchedulesSynced, nil
}


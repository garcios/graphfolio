package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"pkg/database"
	"portfolio-api/internal/marketdata"
	"portfolio-api/internal/repository"
	"portfolio-api/internal/service"
)

func main() {
	var (
		asOfStr            = flag.String("as-of", "", "Date for daily sync (YYYY-MM-DD, defaults to today)")
		fromStr            = flag.String("from", "", "Start date for historical range backfill (YYYY-MM-DD)")
		toStr              = flag.String("to", "", "End date for historical range backfill (YYYY-MM-DD, defaults to today)")
		symbolFlag         = flag.String("symbol", "", "Specific instrument symbol to backfill (e.g. AAPL)")
		pairFlag           = flag.String("pair", "", "Specific currency pair to backfill (e.g. EUR/USD)")
		rebuildProjections = flag.Bool("rebuild-projections", true, "Recompute portfolio projections and valuations after ingestion")
	)
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	log.Println("[market-ingest] Initializing Market Data Ingestion Pipeline...")

	// 1. Connect to database
	dbCfg, err := database.ConfigFromEnv("PORTFOLIO_DB_URL")
	if err != nil {
		log.Fatalf("[market-ingest] database configuration error: %v", err)
	}

	pool, err := database.NewPool(ctx, dbCfg)
	if err != nil {
		log.Fatalf("[market-ingest] failed to connect to database: %v", err)
	}
	defer pool.Close()

	repo := repository.NewPostgresRepository(pool)

	// 2. Initialize market data providers
	limiter := marketdata.NewRateLimiter(5.0, 5)
	twelveKey := os.Getenv("TWELVE_DATA_API_KEY")

	twelveProvider := marketdata.NewTwelveDataProvider(marketdata.TwelveDataConfig{
		APIKey:  twelveKey,
		Limiter: limiter,
	})
	yahooProvider := marketdata.NewYahooFinanceProvider(marketdata.YahooFinanceConfig{
		Limiter: limiter,
	})
	priceProvider := marketdata.NewResilientPriceProvider(twelveProvider, yahooProvider)

	ecbProvider := marketdata.NewECBProvider(marketdata.ECBProviderConfig{
		Limiter: limiter,
	})

	ingestionSvc := service.NewIngestionService(repo, priceProvider, ecbProvider, limiter)
	portfolioSvc := service.NewPortfolioService(repo, service.WithIngestionService(ingestionSvc))

	// 3. Determine Execution Mode
	todayUTC := time.Now().UTC().Truncate(24 * time.Hour)

	if *fromStr != "" {
		// ====================================================================
		// Mode A: Historical Backfill
		// ====================================================================
		fromDate, err := time.Parse("2006-01-02", *fromStr)
		if err != nil {
			log.Fatalf("[market-ingest] invalid -from date format (must be YYYY-MM-DD): %v", err)
		}
		fromDate = fromDate.UTC().Truncate(24 * time.Hour)

		toDate := todayUTC
		if *toStr != "" {
			t, err := time.Parse("2006-01-02", *toStr)
			if err != nil {
				log.Fatalf("[market-ingest] invalid -to date format (must be YYYY-MM-DD): %v", err)
			}
			toDate = t.UTC().Truncate(24 * time.Hour)
		}

		if toDate.Before(fromDate) {
			log.Fatalf("[market-ingest] -to date cannot be before -from date")
		}

		log.Printf("[market-ingest] Executing historical backfill from %s to %s...",
			fromDate.Format("2006-01-02"), toDate.Format("2006-01-02"))

		if *symbolFlag != "" {
			// Single symbol backfill
			sym := strings.ToUpper(strings.TrimSpace(*symbolFlag))
			inst, err := repo.FindInstrumentBySymbol(ctx, sym)
			if err != nil {
				log.Fatalf("[market-ingest] instrument %s not found: %v", sym, err)
			}

			count, err := ingestionSvc.BackfillInstrumentPrices(ctx, inst.ID, inst.Symbol, inst.ExchangeCode, fromDate, toDate)
			if err != nil {
				log.Fatalf("[market-ingest] backfill failed for %s: %v", sym, err)
			}
			log.Printf("[market-ingest] Backfill complete: %d closing prices persisted for %s", count, sym)

		} else if *pairFlag != "" {
			// Single currency pair backfill
			parts := strings.Split(strings.ReplaceAll(*pairFlag, "/", " "), " ")
			if len(parts) < 2 {
				log.Fatalf("[market-ingest] invalid -pair format (e.g. EUR/USD or EUR USD)")
			}
			base := strings.ToUpper(strings.TrimSpace(parts[0]))
			quote := strings.ToUpper(strings.TrimSpace(parts[1]))

			count, err := ingestionSvc.BackfillCurrencyPair(ctx, base, quote, fromDate, toDate)
			if err != nil {
				log.Fatalf("[market-ingest] backfill failed for %s/%s: %v", base, quote, err)
			}
			log.Printf("[market-ingest] Backfill complete: %d FX fixing rates persisted for %s/%s", count, base, quote)

		} else {
			// Backfill all active instruments and standard FX pairs
			insts, err := repo.ListActiveInstruments(ctx)
			if err != nil {
				log.Fatalf("[market-ingest] failed to list active instruments: %v", err)
			}

			log.Printf("[market-ingest] Backfilling %d active instruments...", len(insts))
			totalPrices := 0
			for _, inst := range insts {
				count, err := ingestionSvc.BackfillInstrumentPrices(ctx, inst.ID, inst.Symbol, inst.ExchangeCode, fromDate, toDate)
				if err != nil {
					log.Printf("[market-ingest] backfill failed for %s: %v", inst.Symbol, err)
					continue
				}
				totalPrices += count
			}
			log.Printf("[market-ingest] Ingested %d historical prices across %d instruments", totalPrices, len(insts))

			// Backfill active FX pairs
			currs, err := repo.ListActiveCurrencies(ctx)
			if err != nil {
				log.Printf("[market-ingest] failed to list active currencies: %v", err)
			} else {
				totalFX := 0
				for _, c := range currs {
					if c != "USD" {
						c1, _ := ingestionSvc.BackfillCurrencyPair(ctx, "USD", c, fromDate, toDate)
						c2, _ := ingestionSvc.BackfillCurrencyPair(ctx, c, "USD", fromDate, toDate)
						totalFX += c1 + c2
					}
				}
				log.Printf("[market-ingest] Ingested %d historical FX rates", totalFX)
			}
		}

	} else {
		// ====================================================================
		// Mode B: Daily Ingestion Sync
		// ====================================================================
		asOfDate := todayUTC
		if *asOfStr != "" {
			t, err := time.Parse("2006-01-02", *asOfStr)
			if err != nil {
				log.Fatalf("[market-ingest] invalid -as-of date format (must be YYYY-MM-DD): %v", err)
			}
			asOfDate = t.UTC().Truncate(24 * time.Hour)
		}

		log.Printf("[market-ingest] Running daily sync for trade date %s...", asOfDate.Format("2006-01-02"))
		result, err := ingestionSvc.IngestDailyMarketData(ctx, asOfDate)
		if err != nil {
			log.Fatalf("[market-ingest] daily sync failed: %v", err)
		}

		fmt.Println("------------------------------------------------------------")
		fmt.Printf("✓ %s\n", result.Message)
		fmt.Printf("  • Asset Prices Synced:   %d\n", result.PricesSynced)
		fmt.Printf("  • FX Rates Synced:       %d\n", result.FXRatesSynced)
		fmt.Println("------------------------------------------------------------")
	}

	// 4. Optionally Rebuild Projections
	if *rebuildProjections {
		log.Println("[market-ingest] Recomputing valuations for demo portfolios...")
		// Demo user 1
		if err := portfolioSvc.RebuildProjections(ctx, "1"); err != nil {
			log.Printf("[market-ingest] notice: rebuild projections for user '1': %v", err)
		} else {
			log.Println("[market-ingest] ✓ Portfolio projections and valuations recomputed successfully")
		}
	}

	log.Println("[market-ingest] Pipeline execution finished.")
}

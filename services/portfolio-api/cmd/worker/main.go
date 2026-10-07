package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pkg/database"
	"portfolio-api/internal/repository"
	"portfolio-api/internal/service"

	"github.com/google/uuid"
)

func main() {
	var (
		asOfStr       = flag.String("as-of", "", "Date for daily valuation snapshot (YYYY-MM-DD, defaults to today UTC)")
		fromStr       = flag.String("from", "", "Start date for historical valuation backfill (YYYY-MM-DD)")
		portfolioFlag = flag.String("portfolio", "", "Specific portfolio ID (UUID) to run valuation for")
		daemonFlag    = flag.Bool("daemon", false, "Run worker as a continuous background daemon")
		cronHour      = flag.Int("cron-hour", 22, "UTC hour to run daily EOD valuation job in daemon mode (default 22:00 UTC)")
	)
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	log.Println("[valuation-worker] Initializing Portfolio Valuation Worker...")

	// 1. Connect to database
	dbCfg, err := database.ConfigFromEnv("PORTFOLIO_DB_URL")
	if err != nil {
		log.Fatalf("[valuation-worker] database configuration error: %v", err)
	}

	pool, err := database.NewPool(ctx, dbCfg)
	if err != nil {
		log.Fatalf("[valuation-worker] failed to connect to database: %v", err)
	}
	defer pool.Close()

	repo := repository.NewPostgresRepository(pool)
	valuationSvc := service.NewValuationService(repo)

	// 2. Daemon Mode vs One-Shot Mode
	if *daemonFlag {
		runDaemon(ctx, repo, valuationSvc, *cronHour)
		return
	}

	// One-Shot Mode Execution
	startTime := time.Now()
	todayUTC := time.Now().UTC().Truncate(24 * time.Hour)
	asOfDate := todayUTC
	if *asOfStr != "" {
		parsed, err := time.Parse("2006-01-02", *asOfStr)
		if err != nil {
			log.Fatalf("[valuation-worker] invalid -as-of date format (must be YYYY-MM-DD): %v", err)
		}
		asOfDate = parsed.UTC().Truncate(24 * time.Hour)
	}

	var specificPortfolio *uuid.UUID
	if *portfolioFlag != "" {
		pID, err := uuid.Parse(*portfolioFlag)
		if err != nil {
			log.Fatalf("[valuation-worker] invalid -portfolio UUID %q: %v", *portfolioFlag, err)
		}
		specificPortfolio = &pID
	}

	if *fromStr != "" {
		// Backfill Mode
		fromDate, err := time.Parse("2006-01-02", *fromStr)
		if err != nil {
			log.Fatalf("[valuation-worker] invalid -from date format (must be YYYY-MM-DD): %v", err)
		}
		fromDate = fromDate.UTC().Truncate(24 * time.Hour)

		if specificPortfolio != nil {
			log.Printf("[valuation-worker] Backfilling valuations for portfolio %s from %s...",
				specificPortfolio, fromDate.Format("2006-01-02"))
			if err := valuationSvc.BackfillPortfolioValuations(ctx, *specificPortfolio, fromDate); err != nil {
				log.Fatalf("[valuation-worker] backfill failed: %v", err)
			}
			log.Printf("[valuation-worker] Completed backfill for portfolio %s in %v",
				specificPortfolio, time.Since(startTime))
		} else {
			portfolios, err := repo.ListActivePortfolios(ctx)
			if err != nil {
				log.Fatalf("[valuation-worker] failed to list active portfolios: %v", err)
			}
			log.Printf("[valuation-worker] Backfilling valuations for %d portfolios from %s...",
				len(portfolios), fromDate.Format("2006-01-02"))

			successCount := 0
			failCount := 0
			for _, pID := range portfolios {
				if err := valuationSvc.BackfillPortfolioValuations(ctx, pID, fromDate); err != nil {
					log.Printf("[valuation-worker] ERROR backfilling portfolio %s: %v", pID, err)
					failCount++
				} else {
					successCount++
				}
			}
			log.Printf("[valuation-worker] Backfill finished in %v. Total: %d, Success: %d, Failed: %d",
				time.Since(startTime), len(portfolios), successCount, failCount)
			if failCount > 0 {
				os.Exit(1)
			}
		}
		return
	}

	// Daily Valuation Snapshot Mode
	if specificPortfolio != nil {
		log.Printf("[valuation-worker] Generating snapshot for portfolio %s as of %s...",
			specificPortfolio, asOfDate.Format("2006-01-02"))
		snap, err := valuationSvc.SnapshotValuation(ctx, *specificPortfolio, asOfDate)
		if err != nil {
			log.Fatalf("[valuation-worker] snapshot failed: %v", err)
		}
		log.Printf("[valuation-worker] Completed snapshot for %s: TotalValue=%s, TWR=%s in %v",
			specificPortfolio, snap.TotalValue().StringFixed(2), snap.TWRIndex.StringFixed(4), time.Since(startTime))
	} else {
		portfolios, err := repo.ListActivePortfolios(ctx)
		if err != nil {
			log.Fatalf("[valuation-worker] failed to list active portfolios: %v", err)
		}
		log.Printf("[valuation-worker] Running daily valuation job for %d active portfolios as of %s...",
			len(portfolios), asOfDate.Format("2006-01-02"))

		if err := valuationSvc.RunDailyValuationJob(ctx, asOfDate); err != nil {
			log.Fatalf("[valuation-worker] daily valuation job completed with errors in %v: %v",
				time.Since(startTime), err)
		}
		log.Printf("[valuation-worker] Daily valuation job completed successfully for %d portfolios in %v",
			len(portfolios), time.Since(startTime))
	}
}

func runDaemon(ctx context.Context, repo repository.Repository, svc service.ValuationService, scheduledHour int) {
	log.Printf("[valuation-worker] Starting in daemon mode (scheduled daily EOD at %02d:00 UTC)...", scheduledHour)

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	var lastRunDate string

	for {
		select {
		case <-ctx.Done():
			log.Println("[valuation-worker] Daemon received shutdown signal, exiting...")
			return
		case now := <-ticker.C:
			nowUTC := now.UTC()
			todayStr := nowUTC.Format("2006-01-02")

			// Check if weekday (Mon-Fri) and hour matched and not yet run today
			weekday := nowUTC.Weekday()
			isWeekday := weekday >= time.Monday && weekday <= time.Friday
			if isWeekday && nowUTC.Hour() == scheduledHour && lastRunDate != todayStr {
				log.Printf("[valuation-worker] Scheduled trigger at %s (%s). Running daily valuation job...",
					nowUTC.Format(time.RFC3339), weekday)
				startTime := time.Now()
				asOfDate := nowUTC.Truncate(24 * time.Hour)

				portfolios, err := repo.ListActivePortfolios(ctx)
				if err != nil {
					log.Printf("[valuation-worker] daemon error listing portfolios: %v", err)
					continue
				}

				if err := svc.RunDailyValuationJob(ctx, asOfDate); err != nil {
					log.Printf("[valuation-worker] daemon daily job completed with errors in %v: %v",
						time.Since(startTime), err)
				} else {
					log.Printf("[valuation-worker] daemon daily job succeeded for %d portfolios in %v",
						len(portfolios), time.Since(startTime))
				}
				lastRunDate = todayStr
			}
		}
	}
}

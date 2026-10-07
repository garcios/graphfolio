package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pkg/database"
	"portfolio-api/internal"
	"portfolio-api/internal/marketdata"
	"portfolio-api/internal/repository"
	"portfolio-api/internal/service"
	"portfolio-api/migrations"

	pb "graphfolio/proto/portfolio/v1"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Database Connection & Optional Migrations
	var pool *pgxpool.Pool
	var svc service.PortfolioService

	dbCfg, err := database.ConfigFromEnv("PORTFOLIO_DB_URL")
	if err != nil {
		log.Printf("portfolio-api: %v (running with fallback mock)", err)
	} else {
		// Run migrations on start if requested
		if os.Getenv("MIGRATE_ON_START") == "true" {
			log.Println("portfolio-api: running database migrations...")
			if err := database.RunMigrations(migrations.FS, dbCfg.DSN); err != nil {
				log.Fatalf("portfolio-api: migration failed: %v", err)
			}
			log.Println("portfolio-api: database migrations applied successfully")
		}

		poolCtx, poolCancel := context.WithTimeout(ctx, 10*time.Second)
		pool, err = database.NewPool(poolCtx, dbCfg)
		poolCancel()
		if err != nil {
			log.Printf("portfolio-api: failed to connect to database: %v (running with fallback mock)", err)
		} else {
			defer pool.Close()
			log.Printf("portfolio-api: database pool connected")
			repo := repository.NewPostgresRepository(pool)

			// Market data providers & ingestion pipeline
			rateLimiter := marketdata.NewRateLimiter(5.0, 5)
			twelveDataKey := os.Getenv("TWELVE_DATA_API_KEY")
			twelveProvider := marketdata.NewTwelveDataProvider(marketdata.TwelveDataConfig{
				APIKey:  twelveDataKey,
				Limiter: rateLimiter,
			})
			yahooProvider := marketdata.NewYahooFinanceProvider(marketdata.YahooFinanceConfig{
				Limiter: rateLimiter,
			})
			priceProvider := marketdata.NewResilientPriceProvider(twelveProvider, yahooProvider)
			ecbProvider := marketdata.NewECBProvider(marketdata.ECBProviderConfig{
				Limiter: rateLimiter,
			})
			ingestionSvc := service.NewIngestionService(repo, priceProvider, ecbProvider, rateLimiter)
			valuationSvc := service.NewValuationService(repo)

			svc = service.NewPortfolioService(repo,
				service.WithIngestionService(ingestionSvc),
				service.WithValuationService(valuationSvc),
			)
		}
	}

	// 2. gRPC Listener
	port := os.Getenv("PORT")
	if port == "" {
		port = "50051"
	}
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("portfolio-api: failed to listen on :%s: %v", port, err)
	}

	s := grpc.NewServer()
	pb.RegisterPortfolioServiceServer(s, internal.NewPortfolioServer(svc))
	reflection.Register(s)

	// 3. Graceful Shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		log.Println("portfolio-api: shutting down gRPC server...")
		s.GracefulStop()
		cancel()
	}()

	log.Printf("portfolio-api listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil && err != grpc.ErrServerStopped {
		log.Fatalf("portfolio-api: server failed: %v", err)
	}
}

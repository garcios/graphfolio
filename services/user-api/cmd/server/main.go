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
	"user-api/migrations"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Database Connection & Optional Migrations
	var pool *pgxpool.Pool
	dbCfg, err := database.ConfigFromEnv("USER_DB_URL")
	if err != nil {
		log.Printf("user-api: %v (running without database)", err)
	} else {
		// Run migrations on start if requested
		if os.Getenv("MIGRATE_ON_START") == "true" {
			log.Println("user-api: running database migrations...")
			if err := database.RunMigrations(migrations.FS, dbCfg.DSN); err != nil {
				log.Fatalf("user-api: migration failed: %v", err)
			}
			log.Println("user-api: database migrations applied successfully")
		}

		poolCtx, poolCancel := context.WithTimeout(ctx, 10*time.Second)
		pool, err = database.NewPool(poolCtx, dbCfg)
		poolCancel()
		if err != nil {
			log.Printf("user-api: failed to connect to database: %v", err)
		} else {
			defer pool.Close()
			log.Printf("user-api: database pool connected")
		}
	}

	// 2. gRPC Listener
	port := os.Getenv("PORT")
	if port == "" {
		port = "50052"
	}
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("user-api: failed to listen on :%s: %v", port, err)
	}

	s := grpc.NewServer()
	reflection.Register(s)

	// 3. Graceful Shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		log.Println("user-api: shutting down gRPC server...")
		s.GracefulStop()
		cancel()
	}()

	log.Printf("user-api listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil && err != grpc.ErrServerStopped {
		log.Fatalf("user-api: server failed: %v", err)
	}
}

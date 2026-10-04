package database

import (
	"context"
	"fmt"

	pgxdecimal "github.com/jackc/pgx-shopspring-decimal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool initializes and validates a pgxpool.Pool configured for the GraphFolio domain,
// including registering the shopspring decimal extension on every new connection.
func NewPool(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	cleanDSN, err := SanitizeDSN(cfg.DSN)
	if err != nil {
		return nil, err
	}

	poolCfg, err := pgxpool.ParseConfig(cleanDSN)
	if err != nil {
		return nil, fmt.Errorf("database: invalid pgx pool configuration: %w", err)
	}

	if cfg.MaxConns > 0 {
		poolCfg.MaxConns = cfg.MaxConns
	}
	if cfg.MinConns >= 0 {
		poolCfg.MinConns = cfg.MinConns
	}
	if cfg.MaxConnLifetime > 0 {
		poolCfg.MaxConnLifetime = cfg.MaxConnLifetime
	}
	if cfg.MaxConnIdleTime > 0 {
		poolCfg.MaxConnIdleTime = cfg.MaxConnIdleTime
	}
	if cfg.HealthCheckPeriod > 0 {
		poolCfg.HealthCheckPeriod = cfg.HealthCheckPeriod
	}

	// Register shopspring-decimal mapping on every new connection
	existingAfterConnect := poolCfg.AfterConnect
	poolCfg.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		pgxdecimal.Register(conn.TypeMap())
		if existingAfterConnect != nil {
			return existingAfterConnect(ctx, conn)
		}
		return nil
	}

	connectCtx := ctx
	if cfg.ConnectTimeout > 0 {
		var cancel context.CancelFunc
		connectCtx, cancel = context.WithTimeout(ctx, cfg.ConnectTimeout)
		defer cancel()
	}

	pool, err := pgxpool.NewWithConfig(connectCtx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("database: failed to create pool: %w", err)
	}

	if err := pool.Ping(connectCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database: ping failed: %w", err)
	}

	return pool, nil
}

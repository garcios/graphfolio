package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"portfolio-api/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

var (
	ErrPortfolioNotFound   = errors.New("portfolio not found")
	ErrInstrumentNotFound  = errors.New("instrument not found")
	ErrTransactionNotFound = errors.New("transaction not found")
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) FindPortfolioByUser(ctx context.Context, userID string) (*domain.Portfolio, error) {
	row := r.pool.QueryRow(ctx, findPortfolioByUserSQL, userID)

	var p domain.Portfolio
	var method string
	err := row.Scan(
		&p.ID,
		&p.UserID,
		&p.Name,
		&p.BaseCurrency,
		&method,
		&p.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPortfolioNotFound
		}
		return nil, fmt.Errorf("repository: failed to find portfolio by user %q: %w", userID, err)
	}

	p.CostBasisMethod = domain.CostBasisMethod(method)
	return &p, nil
}

func (r *PostgresRepository) GetHoldingsWithMarketData(ctx context.Context, portfolioID uuid.UUID) ([]domain.HoldingWithPrice, error) {
	rows, err := r.pool.Query(ctx, getHoldingsWithMarketDataSQL, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("repository: query holdings failed: %w", err)
	}
	defer rows.Close()

	var holdings []domain.HoldingWithPrice
	for rows.Next() {
		var h domain.HoldingWithPrice
		err := rows.Scan(
			&h.PortfolioID,
			&h.InstrumentID,
			&h.Ticker,
			&h.Name,
			&h.InstrumentCurrency,
			&h.Quantity,
			&h.CostBasis,
			&h.CostBasisBase,
			&h.RealizedPnLBase,
			&h.DividendsBase,
			&h.LatestPrice,
			&h.PrevPrice,
			&h.FXRateToBase,
		)
		if err != nil {
			return nil, fmt.Errorf("repository: scan holding failed: %w", err)
		}
		holdings = append(holdings, h)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: holdings rows error: %w", err)
	}

	return holdings, nil
}

func (r *PostgresRepository) GetCashBalances(ctx context.Context, portfolioID uuid.UUID) ([]domain.CashBalance, error) {
	rows, err := r.pool.Query(ctx, getCashBalancesSQL, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("repository: query cash balances failed: %w", err)
	}
	defer rows.Close()

	var balances []domain.CashBalance
	for rows.Next() {
		var b domain.CashBalance
		if err := rows.Scan(&b.PortfolioID, &b.CurrencyCode, &b.Balance); err != nil {
			return nil, fmt.Errorf("repository: scan cash balance failed: %w", err)
		}
		balances = append(balances, b)
	}

	return balances, rows.Err()
}

func (r *PostgresRepository) GetLatestValuation(ctx context.Context, portfolioID uuid.UUID) (*domain.PortfolioValuation, error) {
	row := r.pool.QueryRow(ctx, getLatestValuationSQL, portfolioID)

	var v domain.PortfolioValuation
	err := row.Scan(
		&v.PortfolioID,
		&v.ValuationDate,
		&v.MarketValueBase,
		&v.CashValueBase,
		&v.NetFlowBase,
		&v.DailyReturn,
		&v.TWRIndex,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // non-fatal: portfolio may not have valuation history yet
		}
		return nil, fmt.Errorf("repository: query valuation failed: %w", err)
	}

	return &v, nil
}

func (r *PostgresRepository) GetPortfolioValuations(ctx context.Context, portfolioID uuid.UUID, fromDate time.Time) ([]domain.PortfolioValuation, error) {
	queryDate := fromDate
	if queryDate.IsZero() {
		queryDate = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	rows, err := r.pool.Query(ctx, getPortfolioValuationsSQL, portfolioID, queryDate)
	if err != nil {
		return nil, fmt.Errorf("repository: query portfolio valuations failed: %w", err)
	}
	defer rows.Close()

	var valuations []domain.PortfolioValuation
	for rows.Next() {
		var v domain.PortfolioValuation
		if err := rows.Scan(
			&v.PortfolioID,
			&v.ValuationDate,
			&v.MarketValueBase,
			&v.CashValueBase,
			&v.NetFlowBase,
			&v.DailyReturn,
			&v.TWRIndex,
		); err != nil {
			return nil, fmt.Errorf("repository: scan portfolio valuation failed: %w", err)
		}
		valuations = append(valuations, v)
	}

	return valuations, rows.Err()
}

func (r *PostgresRepository) GetCashFXRates(ctx context.Context, baseCurrency string) (map[string]decimal.Decimal, error) {
	rows, err := r.pool.Query(ctx, getCashFXRatesSQL, baseCurrency)
	if err != nil {
		return nil, fmt.Errorf("repository: query fx rates failed: %w", err)
	}
	defer rows.Close()

	rates := make(map[string]decimal.Decimal)
	for rows.Next() {
		var ccy string
		var rate decimal.Decimal
		if err := rows.Scan(&ccy, &rate); err != nil {
			return nil, fmt.Errorf("repository: scan fx rate failed: %w", err)
		}
		rates[ccy] = rate
	}

	return rates, rows.Err()
}

func (r *PostgresRepository) GetTransactions(ctx context.Context, portfolioID uuid.UUID) ([]domain.Transaction, error) {
	rows, err := r.pool.Query(ctx, getTransactionsSQL, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("repository: query transactions failed: %w", err)
	}
	defer rows.Close()

	var txs []domain.Transaction
	for rows.Next() {
		var tx domain.Transaction
		var txType string
		err := rows.Scan(
			&tx.ID,
			&tx.PortfolioID,
			&tx.InstrumentID,
			&txType,
			&tx.TradeDate,
			&tx.Quantity,
			&tx.Price,
			&tx.Amount,
			&tx.CurrencyCode,
			&tx.Fee,
			&tx.WithholdingTax,
			&tx.FXRateToBase,
			&tx.ExternalRef,
			&tx.Notes,
		)
		if err != nil {
			return nil, fmt.Errorf("repository: scan transaction failed: %w", err)
		}
		tx.Type = domain.TransactionType(txType)
		txs = append(txs, tx)
	}

	return txs, rows.Err()
}

func (r *PostgresRepository) GetCorporateActions(ctx context.Context, instrumentIDs []uuid.UUID) ([]domain.CorporateAction, error) {
	if len(instrumentIDs) == 0 {
		return nil, nil
	}

	rows, err := r.pool.Query(ctx, getCorporateActionsSQL, instrumentIDs)
	if err != nil {
		return nil, fmt.Errorf("repository: query corporate actions failed: %w", err)
	}
	defer rows.Close()

	var cas []domain.CorporateAction
	for rows.Next() {
		var ca domain.CorporateAction
		var caType string
		err := rows.Scan(
			&ca.ID,
			&ca.InstrumentID,
			&caType,
			&ca.ExDate,
			&ca.RatioFrom,
			&ca.RatioTo,
			&ca.NewInstrumentID,
			&ca.CashAmount,
			&ca.CurrencyCode,
			&ca.CostBasisPct,
		)
		if err != nil {
			return nil, fmt.Errorf("repository: scan corporate action failed: %w", err)
		}
		ca.Type = domain.CorporateActionType(caType)
		cas = append(cas, ca)
	}

	return cas, rows.Err()
}

func (r *PostgresRepository) SaveProjectionsTx(
	ctx context.Context,
	portfolioID uuid.UUID,
	lots []domain.TaxLot,
	disposals []domain.LotDisposal,
	holdings []domain.Holding,
	cash []domain.CashBalance,
) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("repository: failed to begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Lock portfolio to serialize concurrent rebuilds
	var lockedID uuid.UUID
	err = tx.QueryRow(ctx, "SELECT id FROM portfolio.portfolios WHERE id = $1 FOR UPDATE", portfolioID).Scan(&lockedID)
	if err != nil {
		return fmt.Errorf("repository: failed to lock portfolio %s: %w", portfolioID, err)
	}

	// Clean out existing projections
	if _, err := tx.Exec(ctx, "DELETE FROM portfolio.lot_disposals WHERE tax_lot_id IN (SELECT id FROM portfolio.tax_lots WHERE portfolio_id = $1)", portfolioID); err != nil {
		return fmt.Errorf("repository: failed to clear disposals: %w", err)
	}
	if _, err := tx.Exec(ctx, "DELETE FROM portfolio.tax_lots WHERE portfolio_id = $1", portfolioID); err != nil {
		return fmt.Errorf("repository: failed to clear tax lots: %w", err)
	}
	if _, err := tx.Exec(ctx, "DELETE FROM portfolio.holdings WHERE portfolio_id = $1", portfolioID); err != nil {
		return fmt.Errorf("repository: failed to clear holdings: %w", err)
	}
	if _, err := tx.Exec(ctx, "DELETE FROM portfolio.cash_balances WHERE portfolio_id = $1", portfolioID); err != nil {
		return fmt.Errorf("repository: failed to clear cash balances: %w", err)
	}

	// Insert Tax Lots
	for _, lot := range lots {
		_, err := tx.Exec(ctx, `
			INSERT INTO portfolio.tax_lots (
				id, portfolio_id, instrument_id, open_transaction_id, acquired_date,
				original_quantity, remaining_quantity, cost_basis, cost_basis_base, closed_date
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`, lot.ID, lot.PortfolioID, lot.InstrumentID, lot.OpenTransactionID, lot.AcquiredDate,
			lot.OriginalQuantity, lot.RemainingQuantity, lot.CostBasis, lot.CostBasisBase, lot.ClosedDate)
		if err != nil {
			return fmt.Errorf("repository: failed to insert tax lot %s: %w", lot.ID, err)
		}
	}

	// Insert Lot Disposals
	for _, d := range disposals {
		_, err := tx.Exec(ctx, `
			INSERT INTO portfolio.lot_disposals (
				id, tax_lot_id, sell_transaction_id, quantity, cost_basis_released,
				cost_basis_released_base, proceeds_base, realized_pnl_base
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, d.ID, d.TaxLotID, d.SellTransactionID, d.Quantity, d.CostBasisReleased,
			d.CostBasisReleasedBase, d.ProceedsBase, d.RealizedPnLBase)
		if err != nil {
			return fmt.Errorf("repository: failed to insert disposal %s: %w", d.ID, err)
		}
	}

	// Insert Holdings
	for _, h := range holdings {
		_, err := tx.Exec(ctx, `
			INSERT INTO portfolio.holdings (
				portfolio_id, instrument_id, quantity, cost_basis, cost_basis_base,
				realized_pnl_base, dividends_base
			) VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, h.PortfolioID, h.InstrumentID, h.Quantity, h.CostBasis, h.CostBasisBase,
			h.RealizedPnLBase, h.DividendsBase)
		if err != nil {
			return fmt.Errorf("repository: failed to insert holding %s: %w", h.InstrumentID, err)
		}
	}

	// Insert Cash Balances
	for _, c := range cash {
		_, err := tx.Exec(ctx, `
			INSERT INTO portfolio.cash_balances (portfolio_id, currency_code, balance)
			VALUES ($1, $2, $3)
		`, c.PortfolioID, c.CurrencyCode, c.Balance)
		if err != nil {
			return fmt.Errorf("repository: failed to insert cash balance %s: %w", c.CurrencyCode, err)
		}
	}

	return tx.Commit(ctx)
}

func (r *PostgresRepository) InsertTransaction(ctx context.Context, tx domain.Transaction) (*domain.Transaction, error) {
	row := r.pool.QueryRow(ctx, insertTransactionSQL,
		tx.PortfolioID,
		tx.InstrumentID,
		string(tx.Type),
		tx.TradeDate,
		tx.Quantity,
		tx.Price,
		tx.Amount,
		tx.CurrencyCode,
		tx.Fee,
		tx.WithholdingTax,
		tx.FXRateToBase,
		tx.ExternalRef,
		tx.Notes,
	)

	var id uuid.UUID
	if err := row.Scan(&id); err != nil {
		return nil, fmt.Errorf("repository: insert transaction failed: %w", err)
	}

	tx.ID = id
	return &tx, nil
}

func (r *PostgresRepository) ListTransactions(ctx context.Context, portfolioID uuid.UUID, filter domain.TransactionFilter) ([]domain.TransactionWithInstrument, int, error) {
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	} else if pageSize > 100 {
		pageSize = 100
	}
	limit := pageSize
	offset := (page - 1) * pageSize

	var typeFilter *string
	if filter.Type != nil && *filter.Type != "" {
		s := string(*filter.Type)
		typeFilter = &s
	}

	rows, err := r.pool.Query(ctx, listTransactionsSQL, portfolioID, typeFilter, filter.Symbol, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("repository: list transactions failed: %w", err)
	}
	defer rows.Close()

	var items []domain.TransactionWithInstrument
	totalCount := 0
	for rows.Next() {
		var item domain.TransactionWithInstrument
		var txType string
		var count int
		err := rows.Scan(
			&item.ID,
			&item.PortfolioID,
			&item.InstrumentID,
			&txType,
			&item.TradeDate,
			&item.Quantity,
			&item.Price,
			&item.Amount,
			&item.CurrencyCode,
			&item.Fee,
			&item.WithholdingTax,
			&item.FXRateToBase,
			&item.ExternalRef,
			&item.Notes,
			&item.CreatedAt,
			&item.Symbol,
			&item.InstrumentName,
			&count,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("repository: scan transaction failed: %w", err)
		}
		item.Type = domain.TransactionType(txType)
		totalCount = count
		items = append(items, item)
	}

	if items == nil {
		items = []domain.TransactionWithInstrument{}
	}

	return items, totalCount, rows.Err()
}

func (r *PostgresRepository) DeleteTransaction(ctx context.Context, portfolioID uuid.UUID, transactionID uuid.UUID) error {
	cmdTag, err := r.pool.Exec(ctx, deleteTransactionSQL, transactionID, portfolioID)
	if err != nil {
		return fmt.Errorf("repository: delete transaction failed: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrTransactionNotFound
	}
	return nil
}

func (r *PostgresRepository) FindInstrumentBySymbol(ctx context.Context, symbol string) (*domain.Instrument, error) {
	row := r.pool.QueryRow(ctx, findInstrumentBySymbolSQL, symbol)

	var inst domain.Instrument
	err := row.Scan(
		&inst.ID,
		&inst.Symbol,
		&inst.ExchangeCode,
		&inst.Name,
		&inst.AssetClass,
		&inst.CurrencyCode,
		&inst.ISIN,
		&inst.IsActive,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInstrumentNotFound
		}
		return nil, fmt.Errorf("repository: find instrument %q failed: %w", symbol, err)
	}

	return &inst, nil
}

func (r *PostgresRepository) ListActiveInstruments(ctx context.Context) ([]domain.Instrument, error) {
	rows, err := r.pool.Query(ctx, listActiveInstrumentsSQL)
	if err != nil {
		return nil, fmt.Errorf("repository: list instruments failed: %w", err)
	}
	defer rows.Close()

	var insts []domain.Instrument
	for rows.Next() {
		var inst domain.Instrument
		err := rows.Scan(
			&inst.ID,
			&inst.Symbol,
			&inst.ExchangeCode,
			&inst.Name,
			&inst.AssetClass,
			&inst.CurrencyCode,
			&inst.ISIN,
			&inst.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("repository: scan instrument failed: %w", err)
		}
		insts = append(insts, inst)
	}

	return insts, rows.Err()
}

func (r *PostgresRepository) GetFXRate(ctx context.Context, fromCurrency, toCurrency string) (decimal.Decimal, error) {
	if fromCurrency == toCurrency {
		return decimal.NewFromInt(1), nil
	}

	// Try direct rate
	var directRate decimal.Decimal
	err := r.pool.QueryRow(ctx, getDirectFXRateSQL, fromCurrency, toCurrency).Scan(&directRate)
	if err == nil && directRate.IsPositive() {
		return directRate, nil
	}

	// Try inverse rate
	var inverseRate decimal.Decimal
	err = r.pool.QueryRow(ctx, getInverseFXRateSQL, fromCurrency, toCurrency).Scan(&inverseRate)
	if err == nil && inverseRate.IsPositive() {
		return decimal.NewFromInt(1).Div(inverseRate), nil
	}

	return decimal.NewFromInt(1), nil
}

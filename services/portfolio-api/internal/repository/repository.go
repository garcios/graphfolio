package repository

import (
	"context"
	"time"

	"portfolio-api/internal/domain"
	"portfolio-api/internal/marketdata"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Repository defines the persistence operations for portfolios, market data, and ledger projections.
//
//go:generate go run go.uber.org/mock/mockgen -destination=mocks/mock_repository.go -package=mocks portfolio-api/internal/repository Repository
type Repository interface {
	FindPortfolioByUser(ctx context.Context, userID string) (*domain.Portfolio, error)
	ListActivePortfolios(ctx context.Context) ([]uuid.UUID, error)
	UpdatePortfolioBaseCurrency(ctx context.Context, portfolioID uuid.UUID, baseCurrency string, fxRate decimal.Decimal) error
	GetHoldingsWithMarketData(ctx context.Context, portfolioID uuid.UUID) ([]domain.HoldingWithPrice, error)
	GetCashBalances(ctx context.Context, portfolioID uuid.UUID) ([]domain.CashBalance, error)
	GetLatestValuation(ctx context.Context, portfolioID uuid.UUID) (*domain.PortfolioValuation, error)
	GetPortfolioValuations(ctx context.Context, portfolioID uuid.UUID, fromDate time.Time) ([]domain.PortfolioValuation, error)
	GetCashFXRates(ctx context.Context, baseCurrency string) (map[string]decimal.Decimal, error)

	GetTransactions(ctx context.Context, portfolioID uuid.UUID) ([]domain.Transaction, error)
	GetCorporateActions(ctx context.Context, instrumentIDs []uuid.UUID) ([]domain.CorporateAction, error)
	SaveProjectionsTx(ctx context.Context, portfolioID uuid.UUID, lots []domain.TaxLot, disposals []domain.LotDisposal, holdings []domain.Holding, cash []domain.CashBalance) error

	InsertTransaction(ctx context.Context, tx domain.Transaction) (*domain.Transaction, error)
	ListTransactions(ctx context.Context, portfolioID uuid.UUID, filter domain.TransactionFilter) ([]domain.TransactionWithInstrument, int, error)
	DeleteTransaction(ctx context.Context, portfolioID uuid.UUID, transactionID uuid.UUID) error
	FindInstrumentBySymbol(ctx context.Context, symbol string) (*domain.Instrument, error)
	ListActiveInstruments(ctx context.Context) ([]domain.Instrument, error)
	GetFXRate(ctx context.Context, fromCurrency, toCurrency string) (decimal.Decimal, error)

	// Admin: Asset Directory Management
	ListAllInstruments(ctx context.Context, isActive *bool, search *string) ([]domain.Instrument, error)
	CreateInstrument(ctx context.Context, input domain.CreateInstrumentInput) (*domain.Instrument, error)
	UpdateInstrument(ctx context.Context, input domain.UpdateInstrumentInput) (*domain.Instrument, error)

	// Admin: Price Management & Overrides
	ListInstrumentPrices(ctx context.Context, filter domain.PriceFilter) ([]domain.InstrumentPrice, int, error)
	UpsertInstrumentPrice(ctx context.Context, instrumentID uuid.UUID, priceDate time.Time, closePrice decimal.Decimal, source string) (*domain.InstrumentPrice, error)
	FindPortfoliosHoldingInstrument(ctx context.Context, instrumentID uuid.UUID) ([]uuid.UUID, error)

	// Admin: Ingestion Diagnostics
	GetIngestionMetrics(ctx context.Context) (*domain.IngestionStatus, error)

	// Market Data & FX Ingestion Pipeline
	BatchUpsertInstrumentPrices(ctx context.Context, records []marketdata.PriceRecord) error
	BatchUpsertFXRates(ctx context.Context, records []marketdata.FXRecord) error
	ListActiveCurrencies(ctx context.Context) ([]string, error)
	HasPricesForRange(ctx context.Context, instrumentID uuid.UUID, fromDate, toDate time.Time) (bool, error)

	// Portfolio Valuation Engine & Historical Backfill
	UpsertValuationsBatch(ctx context.Context, valuations []domain.PortfolioValuation) error
	GetLatestValuationBefore(ctx context.Context, portfolioID uuid.UUID, beforeDate time.Time) (*domain.PortfolioValuation, error)
	GetHistoricalPriceMatrix(ctx context.Context, instrumentIDs []uuid.UUID, fromDate, toDate time.Time) (map[uuid.UUID]map[string]decimal.Decimal, error)
	GetHistoricalFXMatrix(ctx context.Context, currencies []string, baseCurrency string, fromDate, toDate time.Time) (map[string]map[string]decimal.Decimal, error)
	DeleteValuationsFromDate(ctx context.Context, portfolioID uuid.UUID, fromDate time.Time) error

	// Admin: FX Rates & Currency Pair Inspection
	ListCurrencyPairs(ctx context.Context) ([]domain.CurrencyPairSummary, error)
	ListFXRates(ctx context.Context, filter domain.FXRateFilter) ([]domain.FXRate, int, error)
	GetHistoricalFXRates(ctx context.Context, baseCurrency, quoteCurrency string, fromDate, toDate *time.Time) ([]domain.FXRate, error)
	UpsertFXRate(ctx context.Context, baseCurrency, quoteCurrency string, rateDate time.Time, rate decimal.Decimal, source string) (*domain.FXRate, error)
}

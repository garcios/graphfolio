package repository

import (
	"context"
	"time"

	"portfolio-api/internal/domain"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Repository defines the persistence operations for portfolios, market data, and ledger projections.
//
//go:generate go run go.uber.org/mock/mockgen -destination=mocks/mock_repository.go -package=mocks portfolio-api/internal/repository Repository
type Repository interface {
	FindPortfolioByUser(ctx context.Context, userID string) (*domain.Portfolio, error)
	GetHoldingsWithMarketData(ctx context.Context, portfolioID uuid.UUID) ([]domain.HoldingWithPrice, error)
	GetCashBalances(ctx context.Context, portfolioID uuid.UUID) ([]domain.CashBalance, error)
	GetLatestValuation(ctx context.Context, portfolioID uuid.UUID) (*domain.PortfolioValuation, error)
	GetPortfolioValuations(ctx context.Context, portfolioID uuid.UUID, fromDate time.Time) ([]domain.PortfolioValuation, error)
	GetCashFXRates(ctx context.Context, baseCurrency string) (map[string]decimal.Decimal, error)

	GetTransactions(ctx context.Context, portfolioID uuid.UUID) ([]domain.Transaction, error)
	GetCorporateActions(ctx context.Context, instrumentIDs []uuid.UUID) ([]domain.CorporateAction, error)
	SaveProjectionsTx(ctx context.Context, portfolioID uuid.UUID, lots []domain.TaxLot, disposals []domain.LotDisposal, holdings []domain.Holding, cash []domain.CashBalance) error

	InsertTransaction(ctx context.Context, tx domain.Transaction) (*domain.Transaction, error)
	FindInstrumentBySymbol(ctx context.Context, symbol string) (*domain.Instrument, error)
	ListActiveInstruments(ctx context.Context) ([]domain.Instrument, error)
	GetFXRate(ctx context.Context, fromCurrency, toCurrency string) (decimal.Decimal, error)
}

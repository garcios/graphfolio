package service

import (
	"context"
	"fmt"
	"time"

	"portfolio-api/internal/domain"
	"portfolio-api/internal/repository"
)

//go:generate go run go.uber.org/mock/mockgen -destination=mocks/mock_service.go -package=mocks portfolio-api/internal/service PortfolioService
type PortfolioService interface {
	GetPortfolioSummary(ctx context.Context, userID string) (*domain.PortfolioSummary, error)
	RebuildProjections(ctx context.Context, userID string) error
	AddTransaction(ctx context.Context, input domain.AddTransactionInput) (*domain.Transaction, *domain.PortfolioSummary, error)
	ListInstruments(ctx context.Context) ([]domain.Instrument, error)
	GetPortfolioHistory(ctx context.Context, userID string, timeframe domain.HistoryTimeframe) (*domain.PortfolioHistory, error)
	ListTransactions(ctx context.Context, userID string, filter domain.TransactionFilter) ([]domain.TransactionWithInstrument, int, error)
	DeleteTransaction(ctx context.Context, userID string, transactionID string) (*domain.PortfolioSummary, error)

	// Admin: Asset Directory Management
	ListAllInstruments(ctx context.Context, isActive *bool, search *string) ([]domain.Instrument, error)
	CreateInstrument(ctx context.Context, input domain.CreateInstrumentInput) (*domain.Instrument, error)
	UpdateInstrument(ctx context.Context, input domain.UpdateInstrumentInput) (*domain.Instrument, error)

	// Admin: Price Management & Overrides
	ListInstrumentPrices(ctx context.Context, filter domain.PriceFilter) ([]domain.InstrumentPrice, int, error)
	RecordPriceOverride(ctx context.Context, input domain.PriceOverrideInput) (*domain.InstrumentPrice, bool, error)

	// Admin: Ingestion Pipeline & Diagnostics
	GetIngestionStatus(ctx context.Context) (*domain.IngestionStatus, error)
	TriggerMarketSync(ctx context.Context, symbols []string, syncFX bool) (*domain.MarketSyncResult, error)
}

type portfolioService struct {
	repo    repository.Repository
	nowFunc func() time.Time
}

func NewPortfolioService(repo repository.Repository) PortfolioService {
	return &portfolioService{
		repo:    repo,
		nowFunc: func() time.Time { return time.Now().UTC() },
	}
}

func (s *portfolioService) GetPortfolioSummary(ctx context.Context, userID string) (*domain.PortfolioSummary, error) {
	portfolio, err := s.repo.FindPortfolioByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	holdings, err := s.repo.GetHoldingsWithMarketData(ctx, portfolio.ID)
	if err != nil {
		return nil, fmt.Errorf("service: failed to get holdings: %w", err)
	}

	cash, err := s.repo.GetCashBalances(ctx, portfolio.ID)
	if err != nil {
		return nil, fmt.Errorf("service: failed to get cash balances: %w", err)
	}

	valuation, err := s.repo.GetLatestValuation(ctx, portfolio.ID)
	if err != nil {
		return nil, fmt.Errorf("service: failed to get valuation: %w", err)
	}

	fxRates, err := s.repo.GetCashFXRates(ctx, portfolio.BaseCurrency)
	if err != nil {
		return nil, fmt.Errorf("service: failed to get fx rates: %w", err)
	}

	summary := domain.CalculatePortfolioSummary(*portfolio, holdings, cash, valuation, fxRates)
	return &summary, nil
}

package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"portfolio-api/internal/domain"
	"portfolio-api/internal/repository"

	"github.com/shopspring/decimal"
)

//go:generate go run go.uber.org/mock/mockgen -destination=mocks/mock_service.go -package=mocks portfolio-api/internal/service PortfolioService
type PortfolioService interface {
	GetPortfolioSummary(ctx context.Context, userID string) (*domain.PortfolioSummary, error)
	UpdatePortfolioBaseCurrency(ctx context.Context, userID string, baseCurrency string) (*domain.PortfolioSummary, error)
	RebuildProjections(ctx context.Context, userID string) error
	AddTransaction(ctx context.Context, input domain.AddTransactionInput) (*domain.Transaction, *domain.PortfolioSummary, error)
	ListInstruments(ctx context.Context) ([]domain.Instrument, error)
	GetPortfolioHistory(ctx context.Context, userID string, timeframe domain.HistoryTimeframe) (*domain.PortfolioHistory, error)
	ListTransactions(ctx context.Context, userID string, filter domain.TransactionFilter) ([]domain.TransactionWithInstrument, int, error)
	DeleteTransaction(ctx context.Context, userID string, transactionID string) (*domain.PortfolioSummary, error)
	RebuildValuations(ctx context.Context, userID string, fromDate *time.Time) error

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

	// Admin: FX Rates & Currency Pair Inspection
	ListCurrencyPairs(ctx context.Context) ([]domain.CurrencyPairSummary, error)
	ListFXRates(ctx context.Context, filter domain.FXRateFilter) ([]domain.FXRate, int, error)
	GetCurrencyPairHistory(ctx context.Context, baseCurrency, quoteCurrency string, timeframe domain.HistoryTimeframe) (*domain.CurrencyPairHistory, error)
	RecordFXRateOverride(ctx context.Context, input domain.FXRateOverrideInput) (*domain.FXRate, bool, error)
}

type ServiceOption func(*portfolioService)

func WithIngestionService(ingestion IngestionService) ServiceOption {
	return func(s *portfolioService) {
		s.ingestion = ingestion
	}
}

func WithValuationService(valuations ValuationService) ServiceOption {
	return func(s *portfolioService) {
		s.valuations = valuations
	}
}

func WithNowFunc(fn func() time.Time) ServiceOption {
	return func(s *portfolioService) {
		s.nowFunc = fn
	}
}

type portfolioService struct {
	repo       repository.Repository
	nowFunc    func() time.Time
	ingestion  IngestionService
	valuations ValuationService
}

func NewPortfolioService(repo repository.Repository, opts ...ServiceOption) PortfolioService {
	s := &portfolioService{
		repo:    repo,
		nowFunc: func() time.Time { return time.Now().UTC() },
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
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

func (s *portfolioService) UpdatePortfolioBaseCurrency(ctx context.Context, userID string, baseCurrency string) (*domain.PortfolioSummary, error) {
	cleanCurrency := strings.TrimSpace(strings.ToUpper(baseCurrency))
	if len(cleanCurrency) != 3 {
		return nil, fmt.Errorf("service: invalid currency code %q (must be 3 letters)", baseCurrency)
	}

	portfolio, err := s.repo.FindPortfolioByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	if portfolio.BaseCurrency == cleanCurrency {
		return s.GetPortfolioSummary(ctx, userID)
	}

	// Calculate conversion rate from current base currency to new base currency
	rate, err := s.repo.GetFXRate(ctx, portfolio.BaseCurrency, cleanCurrency)
	if err != nil || !rate.IsPositive() {
		rate = decimal.NewFromInt(1)
	}

	if err := s.repo.UpdatePortfolioBaseCurrency(ctx, portfolio.ID, cleanCurrency, rate); err != nil {
		return nil, fmt.Errorf("service: failed to update portfolio base currency: %w", err)
	}

	return s.GetPortfolioSummary(ctx, userID)
}

func (s *portfolioService) RebuildValuations(ctx context.Context, userID string, fromDate *time.Time) error {
	portfolio, err := s.repo.FindPortfolioByUser(ctx, userID)
	if err != nil {
		return err
	}
	from := portfolio.CreatedAt
	if fromDate != nil {
		from = *fromDate
	}
	if s.valuations != nil {
		return s.valuations.BackfillPortfolioValuations(ctx, portfolio.ID, from)
	}
	return nil
}

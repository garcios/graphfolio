package service

import (
	"context"
	"fmt"

	"portfolio-api/internal/domain"
	"portfolio-api/internal/repository"
)

//go:generate go run go.uber.org/mock/mockgen -destination=mocks/mock_service.go -package=mocks portfolio-api/internal/service PortfolioService
type PortfolioService interface {
	GetPortfolioSummary(ctx context.Context, userID string) (*domain.PortfolioSummary, error)
	RebuildProjections(ctx context.Context, userID string) error
	AddTransaction(ctx context.Context, input domain.AddTransactionInput) (*domain.Transaction, *domain.PortfolioSummary, error)
	ListInstruments(ctx context.Context) ([]domain.Instrument, error)
}

type portfolioService struct {
	repo repository.Repository
}

func NewPortfolioService(repo repository.Repository) PortfolioService {
	return &portfolioService{repo: repo}
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

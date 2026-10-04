package internal

import (
	"context"
	"errors"

	"portfolio-api/internal/domain"
	"portfolio-api/internal/repository"
	"portfolio-api/internal/service"

	"pkg/decimalpb"

	commonpb "graphfolio/proto/common/v1"
	pb "graphfolio/proto/portfolio/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PortfolioServer struct {
	pb.UnimplementedPortfolioServiceServer
	svc service.PortfolioService
}

func NewPortfolioServer(svc service.PortfolioService) *PortfolioServer {
	return &PortfolioServer{svc: svc}
}

func (s *PortfolioServer) GetPortfolio(ctx context.Context, req *pb.GetPortfolioRequest) (*pb.GetPortfolioResponse, error) {
	if s.svc != nil {
		userID := req.GetUserId()
		if userID == "" {
			userID = "1"
		}

		summary, err := s.svc.GetPortfolioSummary(ctx, userID)
		if err != nil {
			if errors.Is(err, repository.ErrPortfolioNotFound) {
				return nil, status.Errorf(codes.NotFound, "portfolio not found for user: %s", userID)
			}
			return nil, status.Errorf(codes.Internal, "failed to get portfolio: %v", err)
		}

		return &pb.GetPortfolioResponse{
			Portfolio: mapSummaryToProto(summary),
		}, nil
	}

	// Fallback to static mock if service is not initialized
	return &pb.GetPortfolioResponse{
		Portfolio: fallbackMock(),
	}, nil
}

func mapSummaryToProto(summary *domain.PortfolioSummary) *pb.Portfolio {
	investments := make([]*pb.Investment, len(summary.Investments))
	for i, inv := range summary.Investments {
		investments[i] = &pb.Investment{
			Id:                 inv.ID,
			Ticker:             inv.Ticker,
			Name:               inv.Name,
			Price:              decimalpb.MoneyToProto(inv.Price.Amount, inv.Price.CurrencyCode),
			Quantity:           decimalpb.ToProto(inv.Quantity),
			TotalValue:         decimalpb.MoneyToProto(inv.TotalValue.Amount, inv.TotalValue.CurrencyCode),
			TodayReturnAmount:  decimalpb.MoneyToProto(inv.TodayReturnAmount.Amount, inv.TodayReturnAmount.CurrencyCode),
			TodayReturnPercent: decimalpb.ToProto(inv.TodayReturnPercent),
			TotalReturnAmount:  decimalpb.MoneyToProto(inv.TotalReturnAmount.Amount, inv.TotalReturnAmount.CurrencyCode),
			TotalReturnPercent: decimalpb.ToProto(inv.TotalReturnPercent),
		}
	}

	return &pb.Portfolio{
		TotalValue:              decimalpb.MoneyToProto(summary.TotalValue.Amount, summary.TotalValue.CurrencyCode),
		TodayReturnAmount:       decimalpb.MoneyToProto(summary.TodayReturnAmount.Amount, summary.TodayReturnAmount.CurrencyCode),
		TodayReturnPercent:      decimalpb.ToProto(summary.TodayReturnPercent),
		AnnualizedReturnPercent: decimalpb.ToProto(summary.AnnualizedReturnPercent),
		CashBalance:             decimalpb.MoneyToProto(summary.CashBalance.Amount, summary.CashBalance.CurrencyCode),
		Investments:             investments,
	}
}

func money(amount string, ccy string) *commonpb.Money {
	m, _ := decimalpb.MoneyFromString(amount, ccy)
	return m
}

func dec(val string) *commonpb.Decimal {
	d, _ := decimalpb.FromString(val)
	return d
}

func fallbackMock() *pb.Portfolio {
	return &pb.Portfolio{
		TotalValue:              money("124532.89", "USD"),
		TodayReturnAmount:       money("1234.50", "USD"),
		TodayReturnPercent:      dec("1.00"),
		AnnualizedReturnPercent: dec("14.2"),
		CashBalance:             money("8450.00", "USD"),
		Investments: []*pb.Investment{
			{
				Id:                 "1",
				Ticker:             "AAPL",
				Name:               "Apple Inc.",
				Price:              money("185.92", "USD"),
				Quantity:           dec("142.5"),
				TotalValue:         money("26493.60", "USD"),
				TodayReturnAmount:  money("555.75", "USD"),
				TodayReturnPercent: dec("2.14"),
				TotalReturnAmount:  money("4230.10", "USD"),
				TotalReturnPercent: dec("18.9"),
			},
		},
	}
}

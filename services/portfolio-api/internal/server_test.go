package internal

import (
	"context"
	"errors"
	"testing"
	"time"

	"portfolio-api/internal/domain"
	"portfolio-api/internal/repository"
	mocks "portfolio-api/internal/service/mocks"

	pb "graphfolio/proto/portfolio/v1"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestPortfolioServer_GetPortfolio(t *testing.T) {
	ctx := context.Background()

	mockSummary := &domain.PortfolioSummary{
		Portfolio: domain.Portfolio{
			ID:              uuid.New(),
			UserID:          uuid.New(),
			Name:            "Main Tech",
			BaseCurrency:    "USD",
			CostBasisMethod: domain.CostBasisMethodAverageCost,
			CreatedAt:       time.Now().UTC(),
		},
		TotalValue: domain.Money{
			Amount:       decimal.RequireFromString("150000.50"),
			CurrencyCode: "USD",
		},
		TodayReturnAmount: domain.Money{
			Amount:       decimal.RequireFromString("1250.25"),
			CurrencyCode: "USD",
		},
		TodayReturnPercent:      decimal.RequireFromString("0.84"),
		AnnualizedReturnPercent: decimal.RequireFromString("12.50"),
		CashBalance: domain.Money{
			Amount:       decimal.RequireFromString("25000.00"),
			CurrencyCode: "USD",
		},
		Investments: []domain.InvestmentSummary{
			{
				ID:     "aapl-id",
				Ticker: "AAPL",
				Name:   "Apple Inc.",
				Price: domain.Money{
					Amount:       decimal.RequireFromString("180.50"),
					CurrencyCode: "USD",
				},
				Quantity: decimal.NewFromInt(100),
				TotalValue: domain.Money{
					Amount:       decimal.RequireFromString("18050.00"),
					CurrencyCode: "USD",
				},
				TodayReturnAmount: domain.Money{
					Amount:       decimal.RequireFromString("250.00"),
					CurrencyCode: "USD",
				},
				TodayReturnPercent: decimal.RequireFromString("1.40"),
				TotalReturnAmount: domain.Money{
					Amount:       decimal.RequireFromString("3050.00"),
					CurrencyCode: "USD",
				},
				TotalReturnPercent: decimal.RequireFromString("20.33"),
			},
		},
	}

	t.Run("success returns portfolio proto", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		mockSvc.EXPECT().GetPortfolioSummary(ctx, "user-abc").Return(mockSummary, nil)

		res, err := server.GetPortfolio(ctx, &pb.GetPortfolioRequest{UserId: "user-abc"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res == nil || res.Portfolio == nil {
			t.Fatalf("expected non-nil portfolio response")
		}

		p := res.Portfolio
		if p.TotalValue.CurrencyCode != "USD" {
			t.Errorf("expected currency USD, got %s", p.TotalValue.CurrencyCode)
		}
		if len(p.Investments) != 1 {
			t.Fatalf("expected 1 investment, got %d", len(p.Investments))
		}
		if p.Investments[0].Ticker != "AAPL" {
			t.Errorf("expected ticker AAPL, got %s", p.Investments[0].Ticker)
		}
	})

	t.Run("defaults user id to 1 when empty", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		// Expect default userID "1"
		mockSvc.EXPECT().GetPortfolioSummary(ctx, "1").Return(mockSummary, nil)

		res, err := server.GetPortfolio(ctx, &pb.GetPortfolioRequest{UserId: ""})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res == nil || res.Portfolio == nil {
			t.Fatalf("expected non-nil response")
		}
	})

	t.Run("portfolio not found returns NotFound gRPC code", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		mockSvc.EXPECT().GetPortfolioSummary(ctx, "unknown-user").Return(nil, repository.ErrPortfolioNotFound)

		res, err := server.GetPortfolio(ctx, &pb.GetPortfolioRequest{UserId: "unknown-user"})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if res != nil {
			t.Fatalf("expected nil response on error")
		}

		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("expected gRPC status error")
		}
		if st.Code() != codes.NotFound {
			t.Errorf("expected NotFound code, got: %v", st.Code())
		}
	})

	t.Run("service internal error returns Internal gRPC code", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		mockSvc.EXPECT().GetPortfolioSummary(ctx, "user-err").Return(nil, errors.New("database connection reset"))

		res, err := server.GetPortfolio(ctx, &pb.GetPortfolioRequest{UserId: "user-err"})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if res != nil {
			t.Fatalf("expected nil response on error")
		}

		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("expected gRPC status error")
		}
		if st.Code() != codes.Internal {
			t.Errorf("expected Internal code, got: %v", st.Code())
		}
	})

	t.Run("fallback mock returned when service is nil", func(t *testing.T) {
		server := NewPortfolioServer(nil)

		res, err := server.GetPortfolio(ctx, &pb.GetPortfolioRequest{UserId: "user-any"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res == nil || res.Portfolio == nil {
			t.Fatalf("expected non-nil response from fallback")
		}
		if len(res.Portfolio.Investments) == 0 {
			t.Errorf("expected fallback mock investments")
		}
	})
}

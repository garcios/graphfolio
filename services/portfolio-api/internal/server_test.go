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

func TestPortfolioServer_AddTransaction(t *testing.T) {
	ctx := context.Background()

	mockSummary := &domain.PortfolioSummary{
		TotalValue: domain.Money{
			Amount:       decimal.RequireFromString("1000.00"),
			CurrencyCode: "USD",
		},
		CashBalance: domain.Money{
			Amount:       decimal.RequireFromString("500.00"),
			CurrencyCode: "USD",
		},
	}

	t.Run("success adding BUY transaction", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		txID := uuid.New()
		mockSvc.EXPECT().AddTransaction(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, input domain.AddTransactionInput) (*domain.Transaction, *domain.PortfolioSummary, error) {
			if input.Type != domain.TxTypeBuy {
				t.Errorf("expected BUY type, got %v", input.Type)
			}
			if input.Symbol == nil || *input.Symbol != "AAPL" {
				t.Errorf("expected symbol AAPL, got %v", input.Symbol)
			}
			return &domain.Transaction{ID: txID}, mockSummary, nil
		})

		req := &pb.AddTransactionRequest{
			UserId:    "user-1",
			Type:      pb.TransactionType_TRANSACTION_TYPE_BUY,
			Symbol:    "AAPL",
			TradeDate: "2025-01-05",
			Quantity:  dec("10"),
			Price:     money("150.00", "USD"),
			Fee:       money("5.00", "USD"),
		}

		res, err := server.AddTransaction(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res == nil {
			t.Fatalf("expected response, got nil")
		}
		if res.TransactionId != txID.String() {
			t.Errorf("expected tx ID %s, got %s", txID.String(), res.TransactionId)
		}
		if res.Portfolio == nil {
			t.Fatalf("expected non-nil portfolio")
		}
	})

	t.Run("unavailable when service is nil", func(t *testing.T) {
		server := NewPortfolioServer(nil)
		_, err := server.AddTransaction(ctx, &pb.AddTransactionRequest{
			Type: pb.TransactionType_TRANSACTION_TYPE_BUY,
		})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		st, _ := status.FromError(err)
		if st.Code() != codes.Unavailable {
			t.Errorf("expected Unavailable code, got: %v", st.Code())
		}
	})

	t.Run("invalid argument on unspecified type", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		_, err := server.AddTransaction(ctx, &pb.AddTransactionRequest{
			Type: pb.TransactionType_TRANSACTION_TYPE_UNSPECIFIED,
		})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		st, _ := status.FromError(err)
		if st.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument code, got: %v", st.Code())
		}
	})

	t.Run("invalid argument on bad date", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		_, err := server.AddTransaction(ctx, &pb.AddTransactionRequest{
			Type:      pb.TransactionType_TRANSACTION_TYPE_BUY,
			TradeDate: "not-a-date",
		})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		st, _ := status.FromError(err)
		if st.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument code, got: %v", st.Code())
		}
	})

	t.Run("not found when portfolio not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		mockSvc.EXPECT().AddTransaction(ctx, gomock.Any()).Return(nil, nil, repository.ErrPortfolioNotFound)

		_, err := server.AddTransaction(ctx, &pb.AddTransactionRequest{
			Type: pb.TransactionType_TRANSACTION_TYPE_BUY,
		})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		st, _ := status.FromError(err)
		if st.Code() != codes.NotFound {
			t.Errorf("expected NotFound code, got: %v", st.Code())
		}
	})
}

func TestPortfolioServer_ListInstruments(t *testing.T) {
	ctx := context.Background()

	t.Run("success listing instruments", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		instID := uuid.New()
		mockSvc.EXPECT().ListInstruments(ctx).Return([]domain.Instrument{
			{
				ID:           instID,
				Symbol:       "AAPL",
				Name:         "Apple Inc.",
				CurrencyCode: "USD",
				AssetClass:   "EQUITY",
			},
		}, nil)

		res, err := server.ListInstruments(ctx, &pb.ListInstrumentsRequest{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res == nil || len(res.Instruments) != 1 {
			t.Fatalf("expected 1 instrument, got %v", res)
		}
		if res.Instruments[0].Symbol != "AAPL" {
			t.Errorf("expected AAPL, got %s", res.Instruments[0].Symbol)
		}
	})

	t.Run("unavailable when service is nil", func(t *testing.T) {
		server := NewPortfolioServer(nil)
		_, err := server.ListInstruments(ctx, &pb.ListInstrumentsRequest{})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		st, _ := status.FromError(err)
		if st.Code() != codes.Unavailable {
			t.Errorf("expected Unavailable code, got: %v", st.Code())
		}
	})
}

func TestPortfolioServer_GetPortfolioHistory(t *testing.T) {
	ctx := context.Background()

	mockHistory := &domain.PortfolioHistory{
		Points: []domain.ValuationPoint{
			{
				Date:        time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				TotalValue:  domain.NewMoney(decimal.RequireFromString("100000.00"), "USD"),
				MarketValue: domain.NewMoney(decimal.RequireFromString("95000.00"), "USD"),
				CashValue:   domain.NewMoney(decimal.RequireFromString("5000.00"), "USD"),
				TWRIndex:    decimal.RequireFromString("1.0000"),
				DailyReturn: decimal.Zero,
			},
			{
				Date:        time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
				TotalValue:  domain.NewMoney(decimal.RequireFromString("105000.00"), "USD"),
				MarketValue: domain.NewMoney(decimal.RequireFromString("100000.00"), "USD"),
				CashValue:   domain.NewMoney(decimal.RequireFromString("5000.00"), "USD"),
				TWRIndex:    decimal.RequireFromString("1.0500"),
				DailyReturn: decimal.RequireFromString("0.0500"),
			},
		},
		StartValue:    domain.NewMoney(decimal.RequireFromString("100000.00"), "USD"),
		EndValue:      domain.NewMoney(decimal.RequireFromString("105000.00"), "USD"),
		ReturnAmount:  domain.NewMoney(decimal.RequireFromString("5000.00"), "USD"),
		ReturnPercent: decimal.RequireFromString("5.00"),
	}

	t.Run("success returns portfolio history proto", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		mockSvc.EXPECT().GetPortfolioHistory(ctx, "user-123", domain.Timeframe1M).Return(mockHistory, nil)

		res, err := server.GetPortfolioHistory(ctx, &pb.GetPortfolioHistoryRequest{
			UserId:    "user-123",
			Timeframe: pb.HistoryTimeframe_HISTORY_TIMEFRAME_1M,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res == nil {
			t.Fatalf("expected non-nil response")
		}
		if len(res.Points) != 2 {
			t.Fatalf("expected 2 points, got %d", len(res.Points))
		}
		if res.Points[0].Date != "2026-01-01" {
			t.Errorf("expected date 2026-01-01, got %s", res.Points[0].Date)
		}
		if res.Points[0].TotalValue.CurrencyCode != "USD" {
			t.Errorf("expected currency USD, got %s", res.Points[0].TotalValue.CurrencyCode)
		}
		if res.StartValue.CurrencyCode != "USD" {
			t.Errorf("expected USD start value currency, got %s", res.StartValue.CurrencyCode)
		}
	})

	t.Run("defaults user id to 1 when empty", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		mockSvc.EXPECT().GetPortfolioHistory(ctx, "1", domain.Timeframe1Y).Return(mockHistory, nil)

		res, err := server.GetPortfolioHistory(ctx, &pb.GetPortfolioHistoryRequest{
			Timeframe: pb.HistoryTimeframe_HISTORY_TIMEFRAME_1Y,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res == nil {
			t.Fatalf("expected non-nil response")
		}
	})

	t.Run("unavailable when service is nil", func(t *testing.T) {
		server := NewPortfolioServer(nil)
		_, err := server.GetPortfolioHistory(ctx, &pb.GetPortfolioHistoryRequest{
			Timeframe: pb.HistoryTimeframe_HISTORY_TIMEFRAME_1M,
		})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		st, _ := status.FromError(err)
		if st.Code() != codes.Unavailable {
			t.Errorf("expected Unavailable code, got: %v", st.Code())
		}
	})

	t.Run("invalid argument on invalid timeframe", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		_, err := server.GetPortfolioHistory(ctx, &pb.GetPortfolioHistoryRequest{
			UserId:    "user-123",
			Timeframe: pb.HistoryTimeframe(999),
		})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		st, _ := status.FromError(err)
		if st.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument code, got: %v", st.Code())
		}
	})

	t.Run("portfolio not found returns NotFound gRPC code", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		mockSvc.EXPECT().GetPortfolioHistory(ctx, "unknown-user", domain.TimeframeAll).Return(nil, repository.ErrPortfolioNotFound)

		_, err := server.GetPortfolioHistory(ctx, &pb.GetPortfolioHistoryRequest{
			UserId:    "unknown-user",
			Timeframe: pb.HistoryTimeframe_HISTORY_TIMEFRAME_ALL,
		})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		st, _ := status.FromError(err)
		if st.Code() != codes.NotFound {
			t.Errorf("expected NotFound code, got: %v", st.Code())
		}
	})

	t.Run("internal error on service error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		mockSvc.EXPECT().GetPortfolioHistory(ctx, "user-123", domain.Timeframe1W).Return(nil, errors.New("db failure"))

		_, err := server.GetPortfolioHistory(ctx, &pb.GetPortfolioHistoryRequest{
			UserId:    "user-123",
			Timeframe: pb.HistoryTimeframe_HISTORY_TIMEFRAME_1W,
		})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		st, _ := status.FromError(err)
		if st.Code() != codes.Internal {
			t.Errorf("expected Internal code, got: %v", st.Code())
		}
	})
}

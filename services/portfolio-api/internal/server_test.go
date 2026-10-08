package internal

import (
	"context"
	"errors"
	"testing"
	"time"

	"portfolio-api/internal/domain"
	"portfolio-api/internal/repository"
	"portfolio-api/internal/service"
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

func TestPortfolioServer_UpdatePortfolioBaseCurrency(t *testing.T) {
	ctx := context.Background()

	mockSummary := &domain.PortfolioSummary{
		Portfolio: domain.Portfolio{
			ID:           uuid.New(),
			BaseCurrency: "AUD",
		},
		TotalValue: domain.NewMoney(decimal.NewFromFloat(150000), "AUD"),
	}

	t.Run("successful base currency update", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		mockSvc.EXPECT().UpdatePortfolioBaseCurrency(ctx, "user-1", "AUD").Return(mockSummary, nil)

		res, err := server.UpdatePortfolioBaseCurrency(ctx, &pb.UpdatePortfolioBaseCurrencyRequest{
			UserId:       "user-1",
			BaseCurrency: "AUD",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res == nil || res.Portfolio == nil {
			t.Fatalf("expected non-nil response")
		}
		if res.Portfolio.TotalValue.CurrencyCode != "AUD" {
			t.Errorf("expected currency AUD, got %s", res.Portfolio.TotalValue.CurrencyCode)
		}
	})

	t.Run("returns InvalidArgument when base_currency is empty", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		res, err := server.UpdatePortfolioBaseCurrency(ctx, &pb.UpdatePortfolioBaseCurrencyRequest{
			UserId:       "user-1",
			BaseCurrency: "",
		})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if res != nil {
			t.Fatalf("expected nil response on error")
		}
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument code, got: %v", err)
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

func TestPortfolioServer_ListTransactions(t *testing.T) {
	ctx := context.Background()

	qty := decimal.NewFromInt(10)
	price := decimal.RequireFromString("150.00")
	sym := "AAPL"
	instName := "Apple Inc."
	notes := "Initial buy"
	createdAt := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)

	mockItems := []domain.TransactionWithInstrument{
		{
			Transaction: domain.Transaction{
				ID:           uuid.New(),
				PortfolioID:  uuid.New(),
				Type:         domain.TxTypeBuy,
				TradeDate:    time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
				Quantity:     &qty,
				Price:        &price,
				Amount:       decimal.RequireFromString("1500.00"),
				CurrencyCode: "USD",
				Fee:          decimal.RequireFromString("5.00"),
				Notes:        &notes,
				CreatedAt:    createdAt,
			},
			Symbol:         &sym,
			InstrumentName: &instName,
		},
		{
			Transaction: domain.Transaction{
				ID:           uuid.New(),
				PortfolioID:  uuid.New(),
				Type:         domain.TxTypeDeposit,
				TradeDate:    time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC),
				Quantity:     nil, // nullable
				Price:        nil, // nullable
				Amount:       decimal.RequireFromString("5000.00"),
				CurrencyCode: "USD",
				Fee:          decimal.Zero,
				Notes:        nil,
				CreatedAt:    time.Time{},
			},
			Symbol:         nil, // empty for cash
			InstrumentName: nil,
		},
	}

	t.Run("success returns transaction items proto with nullable mapping", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		mockSvc.EXPECT().ListTransactions(ctx, "user-123", gomock.Any()).DoAndReturn(
			func(_ context.Context, uid string, f domain.TransactionFilter) ([]domain.TransactionWithInstrument, int, error) {
				if uid != "user-123" {
					t.Errorf("expected user-123, got %s", uid)
				}
				if f.Page != 1 || f.PageSize != 20 {
					t.Errorf("expected page 1, pageSize 20; got %d, %d", f.Page, f.PageSize)
				}
				return mockItems, 2, nil
			},
		)

		res, err := server.ListTransactions(ctx, &pb.ListTransactionsRequest{
			UserId: "user-123",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res == nil {
			t.Fatalf("expected non-nil response")
		}
		if res.TotalCount != 2 || res.Page != 1 || res.PageSize != 20 {
			t.Errorf("unexpected pagination metadata: total=%d, page=%d, pageSize=%d", res.TotalCount, res.Page, res.PageSize)
		}
		if len(res.Transactions) != 2 {
			t.Fatalf("expected 2 transactions, got %d", len(res.Transactions))
		}

		// Verify BUY item with non-nil fields
		buyItem := res.Transactions[0]
		if buyItem.Type != pb.TransactionType_TRANSACTION_TYPE_BUY {
			t.Errorf("expected BUY, got %v", buyItem.Type)
		}
		if buyItem.Symbol != "AAPL" {
			t.Errorf("expected symbol AAPL, got %s", buyItem.Symbol)
		}
		if buyItem.InstrumentName != "Apple Inc." {
			t.Errorf("expected name Apple Inc., got %s", buyItem.InstrumentName)
		}
		if buyItem.Quantity == nil || buyItem.Quantity.Value != "10" {
			t.Errorf("expected quantity 10, got %v", buyItem.Quantity)
		}
		if buyItem.Price == nil || buyItem.Price.Amount == nil || buyItem.Price.Amount.Value != "150" {
			t.Errorf("expected price 150, got %v", buyItem.Price)
		}
		if buyItem.Notes != "Initial buy" {
			t.Errorf("expected notes, got %s", buyItem.Notes)
		}
		if buyItem.CreatedAt == "" {
			t.Errorf("expected non-empty created_at")
		}

		// Verify DEPOSIT item with nullable fields (quantity, price, symbol, notes)
		depItem := res.Transactions[1]
		if depItem.Type != pb.TransactionType_TRANSACTION_TYPE_DEPOSIT {
			t.Errorf("expected DEPOSIT, got %v", depItem.Type)
		}
		if depItem.Symbol != "" {
			t.Errorf("expected empty symbol for deposit, got %s", depItem.Symbol)
		}
		if depItem.Quantity != nil {
			t.Errorf("expected nil quantity for deposit, got %v", depItem.Quantity)
		}
		if depItem.Price != nil {
			t.Errorf("expected nil price for deposit, got %v", depItem.Price)
		}
		if depItem.Notes != "" {
			t.Errorf("expected empty notes for deposit, got %s", depItem.Notes)
		}
	})

	t.Run("defaults user id to 1 when empty", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		mockSvc.EXPECT().ListTransactions(ctx, "1", gomock.Any()).Return(mockItems, 2, nil)

		res, err := server.ListTransactions(ctx, &pb.ListTransactionsRequest{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res == nil || len(res.Transactions) != 2 {
			t.Fatalf("expected 2 transactions, got %v", res)
		}
	})

	t.Run("invalid argument on bad pagination", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		badRequests := []*pb.ListTransactionsRequest{
			{Page: -1},
			{PageSize: -5},
			{PageSize: 101},
		}

		for _, req := range badRequests {
			_, err := server.ListTransactions(ctx, req)
			if err == nil {
				t.Fatalf("expected error for bad pagination %v, got nil", req)
			}
			st, ok := status.FromError(err)
			if !ok || st.Code() != codes.InvalidArgument {
				t.Errorf("expected InvalidArgument code for %v, got: %v", req, err)
			}
		}
	})

	t.Run("unavailable when service is nil", func(t *testing.T) {
		server := NewPortfolioServer(nil)
		_, err := server.ListTransactions(ctx, &pb.ListTransactionsRequest{})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		st, _ := status.FromError(err)
		if st.Code() != codes.Unavailable {
			t.Errorf("expected Unavailable code, got: %v", st.Code())
		}
	})

	t.Run("portfolio not found returns NotFound gRPC code", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		mockSvc.EXPECT().ListTransactions(ctx, "unknown-user", gomock.Any()).Return(nil, 0, repository.ErrPortfolioNotFound)

		_, err := server.ListTransactions(ctx, &pb.ListTransactionsRequest{UserId: "unknown-user"})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		st, _ := status.FromError(err)
		if st.Code() != codes.NotFound {
			t.Errorf("expected NotFound code, got: %v", st.Code())
		}
	})

	t.Run("internal error on service failure", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		mockSvc.EXPECT().ListTransactions(ctx, "user-err", gomock.Any()).Return(nil, 0, errors.New("db crash"))

		_, err := server.ListTransactions(ctx, &pb.ListTransactionsRequest{UserId: "user-err"})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		st, _ := status.FromError(err)
		if st.Code() != codes.Internal {
			t.Errorf("expected Internal code, got: %v", st.Code())
		}
	})
}

func TestPortfolioServer_DeleteTransaction(t *testing.T) {
	ctx := context.Background()

	mockSummary := &domain.PortfolioSummary{
		Portfolio: domain.Portfolio{
			ID:           uuid.New(),
			UserID:       uuid.New(),
			Name:         "Main Tech",
			BaseCurrency: "USD",
		},
		TotalValue: domain.Money{
			Amount:       decimal.RequireFromString("150000.00"),
			CurrencyCode: "USD",
		},
		CashBalance: domain.Money{
			Amount:       decimal.RequireFromString("25000.00"),
			CurrencyCode: "USD",
		},
	}

	t.Run("success deletes transaction and returns updated portfolio", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		txID := uuid.New().String()
		mockSvc.EXPECT().DeleteTransaction(ctx, "user-123", txID).Return(mockSummary, nil)

		res, err := server.DeleteTransaction(ctx, &pb.DeleteTransactionRequest{
			UserId:        "user-123",
			TransactionId: txID,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res == nil {
			t.Fatalf("expected non-nil response")
		}
		if !res.Success {
			t.Errorf("expected Success=true")
		}
		if res.Portfolio == nil {
			t.Fatalf("expected non-nil portfolio in response")
		}
	})

	t.Run("defaults user id to 1 when empty", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		txID := uuid.New().String()
		mockSvc.EXPECT().DeleteTransaction(ctx, "1", txID).Return(mockSummary, nil)

		res, err := server.DeleteTransaction(ctx, &pb.DeleteTransactionRequest{
			TransactionId: txID,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res == nil || !res.Success {
			t.Fatalf("expected success response")
		}
	})

	t.Run("invalid argument when transaction id is empty", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		_, err := server.DeleteTransaction(ctx, &pb.DeleteTransactionRequest{
			UserId:        "user-123",
			TransactionId: "",
		})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		st, _ := status.FromError(err)
		if st.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument code, got: %v", st.Code())
		}
	})

	t.Run("not found for invalid or non-existent transaction id", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		mockSvc.EXPECT().DeleteTransaction(ctx, "user-123", "non-existent-id").Return(nil, repository.ErrTransactionNotFound)

		_, err := server.DeleteTransaction(ctx, &pb.DeleteTransactionRequest{
			UserId:        "user-123",
			TransactionId: "non-existent-id",
		})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		st, _ := status.FromError(err)
		if st.Code() != codes.NotFound {
			t.Errorf("expected NotFound code, got: %v", st.Code())
		}
	})

	t.Run("not found when portfolio not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		mockSvc.EXPECT().DeleteTransaction(ctx, "unknown-user", "tx-123").Return(nil, repository.ErrPortfolioNotFound)

		_, err := server.DeleteTransaction(ctx, &pb.DeleteTransactionRequest{
			UserId:        "unknown-user",
			TransactionId: "tx-123",
		})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		st, _ := status.FromError(err)
		if st.Code() != codes.NotFound {
			t.Errorf("expected NotFound code, got: %v", st.Code())
		}
	})

	t.Run("unavailable when service is nil", func(t *testing.T) {
		server := NewPortfolioServer(nil)
		_, err := server.DeleteTransaction(ctx, &pb.DeleteTransactionRequest{
			TransactionId: "tx-123",
		})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		st, _ := status.FromError(err)
		if st.Code() != codes.Unavailable {
			t.Errorf("expected Unavailable code, got: %v", st.Code())
		}
	})

	t.Run("internal error on service failure", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		mockSvc.EXPECT().DeleteTransaction(ctx, "user-123", "tx-123").Return(nil, errors.New("db error"))

		_, err := server.DeleteTransaction(ctx, &pb.DeleteTransactionRequest{
			UserId:        "user-123",
			TransactionId: "tx-123",
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

func TestPortfolioServer_ListAllInstruments(t *testing.T) {
	ctx := context.Background()

	t.Run("success returns proto instruments", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		isin := "US0378331005"
		mockSvc.EXPECT().
			ListAllInstruments(ctx, gomock.Any(), gomock.Any()).
			Return([]domain.Instrument{
				{
					ID:           uuid.MustParse("018f0000-0001-7000-8000-000000000001"),
					Symbol:       "AAPL",
					ExchangeCode: "XNAS",
					Name:         "Apple Inc.",
					AssetClass:   "EQUITY",
					CurrencyCode: "USD",
					ISIN:         &isin,
					IsActive:     true,
				},
			}, nil)

		isActive := true
		search := "AAPL"
		res, err := server.ListAllInstruments(ctx, &pb.ListAllInstrumentsRequest{
			IsActive: &isActive,
			Search:   &search,
		})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if len(res.Instruments) != 1 {
			t.Fatalf("expected 1 instrument, got: %d", len(res.Instruments))
		}
		if res.Instruments[0].Symbol != "AAPL" || res.Instruments[0].Isin != isin {
			t.Errorf("unexpected instrument: %+v", res.Instruments[0])
		}
	})
}

func TestPortfolioServer_ListExchanges(t *testing.T) {
	ctx := context.Background()

	t.Run("success maps all fields", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		mockSvc.EXPECT().ListExchanges(ctx).Return([]domain.Exchange{
			{Code: "XNAS", Name: "NASDAQ Stock Market", Country: "US", Timezone: "America/New_York"},
		}, nil)

		res, err := server.ListExchanges(ctx, &pb.ListExchangesRequest{})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if len(res.Exchanges) != 1 {
			t.Fatalf("expected 1 exchange, got: %d", len(res.Exchanges))
		}
		ex := res.Exchanges[0]
		if ex.Code != "XNAS" || ex.Name != "NASDAQ Stock Market" || ex.Country != "US" || ex.Timezone != "America/New_York" {
			t.Errorf("unexpected exchange mapping: %+v", ex)
		}
	})

	t.Run("empty list returns empty exchanges", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		mockSvc.EXPECT().ListExchanges(ctx).Return([]domain.Exchange{}, nil)

		res, err := server.ListExchanges(ctx, &pb.ListExchangesRequest{})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if len(res.Exchanges) != 0 {
			t.Errorf("expected 0 exchanges, got: %d", len(res.Exchanges))
		}
	})

	t.Run("service error returns Internal", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		mockSvc.EXPECT().ListExchanges(ctx).Return(nil, errors.New("db down"))

		_, err := server.ListExchanges(ctx, &pb.ListExchangesRequest{})
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.Internal {
			t.Errorf("expected Internal code, got: %v", err)
		}
	})
}

func TestPortfolioServer_CreateInstrument(t *testing.T) {
	ctx := context.Background()

	t.Run("success returns created instrument", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		mockSvc.EXPECT().
			CreateInstrument(ctx, gomock.Any()).
			Return(&domain.Instrument{
				ID:           uuid.New(),
				Symbol:       "NVDA",
				ExchangeCode: "XNAS",
				Name:         "NVIDIA Corp.",
				AssetClass:   "EQUITY",
				CurrencyCode: "USD",
				IsActive:     true,
			}, nil)

		res, err := server.CreateInstrument(ctx, &pb.CreateInstrumentRequest{
			Symbol:       "NVDA",
			ExchangeCode: "XNAS",
			Name:         "NVIDIA Corp.",
			AssetClass:   "EQUITY",
			CurrencyCode: "USD",
		})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if res.Instrument.Symbol != "NVDA" {
			t.Errorf("expected symbol NVDA, got: %s", res.Instrument.Symbol)
		}
	})

	t.Run("conflict returns AlreadyExists", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		mockSvc.EXPECT().
			CreateInstrument(ctx, gomock.Any()).
			Return(nil, repository.ErrInstrumentConflict)

		_, err := server.CreateInstrument(ctx, &pb.CreateInstrumentRequest{
			Symbol:       "AAPL",
			ExchangeCode: "XNAS",
			Name:         "Apple Inc.",
			AssetClass:   "EQUITY",
			CurrencyCode: "USD",
		})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		st, _ := status.FromError(err)
		if st.Code() != codes.AlreadyExists {
			t.Errorf("expected AlreadyExists, got: %v", st.Code())
		}
	})

	t.Run("validation error returns InvalidArgument", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		mockSvc.EXPECT().
			CreateInstrument(ctx, gomock.Any()).
			Return(nil, service.ErrInvalidSymbol)

		_, err := server.CreateInstrument(ctx, &pb.CreateInstrumentRequest{})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		st, _ := status.FromError(err)
		if st.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got: %v", st.Code())
		}
	})
}

func TestPortfolioServer_UpdateInstrument(t *testing.T) {
	ctx := context.Background()
	validID := uuid.New()

	t.Run("success returns updated instrument", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		isActive := false
		mockSvc.EXPECT().
			UpdateInstrument(ctx, gomock.Any()).
			Return(&domain.Instrument{
				ID:           validID,
				Symbol:       "AAPL",
				ExchangeCode: "XNAS",
				Name:         "Apple Inc.",
				AssetClass:   "EQUITY",
				CurrencyCode: "USD",
				IsActive:     false,
			}, nil)

		idStr := validID.String()
		res, err := server.UpdateInstrument(ctx, &pb.UpdateInstrumentRequest{
			Id:       idStr,
			IsActive: &isActive,
		})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if res.Instrument.IsActive != false {
			t.Errorf("expected IsActive false, got: %v", res.Instrument.IsActive)
		}
	})

	t.Run("invalid UUID returns InvalidArgument", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		_, err := server.UpdateInstrument(ctx, &pb.UpdateInstrumentRequest{
			Id: "not-a-uuid",
		})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		st, _ := status.FromError(err)
		if st.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got: %v", st.Code())
		}
	})
}

func TestPortfolioServer_DeleteInstrument(t *testing.T) {
	ctx := context.Background()
	validID := uuid.New()

	t.Run("nil service returns Unavailable", func(t *testing.T) {
		server := NewPortfolioServer(nil)
		_, err := server.DeleteInstrument(ctx, &pb.DeleteInstrumentRequest{Id: validID.String()})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		st, _ := status.FromError(err)
		if st.Code() != codes.Unavailable {
			t.Errorf("expected Unavailable, got %v", st.Code())
		}
	})

	t.Run("empty ID returns InvalidArgument", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		_, err := server.DeleteInstrument(ctx, &pb.DeleteInstrumentRequest{Id: ""})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		st, _ := status.FromError(err)
		if st.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got %v", st.Code())
		}
	})

	t.Run("invalid UUID returns InvalidArgument", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		_, err := server.DeleteInstrument(ctx, &pb.DeleteInstrumentRequest{Id: "invalid-uuid"})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		st, _ := status.FromError(err)
		if st.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got %v", st.Code())
		}
	})

	t.Run("success returns response", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		mockSvc.EXPECT().
			DeleteInstrument(ctx, validID).
			Return(nil)

		res, err := server.DeleteInstrument(ctx, &pb.DeleteInstrumentRequest{Id: validID.String()})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !res.Success {
			t.Errorf("expected Success true, got false")
		}
		if res.Id != validID.String() {
			t.Errorf("expected Id %s, got %s", validID.String(), res.Id)
		}
	})

	t.Run("not found returns NotFound", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		mockSvc.EXPECT().
			DeleteInstrument(ctx, validID).
			Return(repository.ErrInstrumentNotFound)

		_, err := server.DeleteInstrument(ctx, &pb.DeleteInstrumentRequest{Id: validID.String()})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		st, _ := status.FromError(err)
		if st.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", st.Code())
		}
	})

	t.Run("in use returns FailedPrecondition", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		mockSvc.EXPECT().
			DeleteInstrument(ctx, validID).
			Return(repository.ErrInstrumentInUse)

		_, err := server.DeleteInstrument(ctx, &pb.DeleteInstrumentRequest{Id: validID.String()})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		st, _ := status.FromError(err)
		if st.Code() != codes.FailedPrecondition {
			t.Errorf("expected FailedPrecondition, got %v", st.Code())
		}
	})

	t.Run("internal error returns Internal", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		mockSvc.EXPECT().
			DeleteInstrument(ctx, validID).
			Return(errors.New("db error"))

		_, err := server.DeleteInstrument(ctx, &pb.DeleteInstrumentRequest{Id: validID.String()})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		st, _ := status.FromError(err)
		if st.Code() != codes.Internal {
			t.Errorf("expected Internal, got %v", st.Code())
		}
	})
}

func TestPortfolioServer_ListInstrumentPrices(t *testing.T) {
	ctx := context.Background()

	t.Run("success returns prices", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		mockSvc.EXPECT().
			ListInstrumentPrices(ctx, gomock.Any()).
			Return([]domain.InstrumentPrice{
				{
					InstrumentID: uuid.New(),
					Symbol:       "AAPL",
					PriceDate:    time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC),
					Close:        decimal.NewFromFloat(228.45),
					CurrencyCode: "USD",
					Source:       "manual",
				},
			}, 1, nil)

		sym := "AAPL"
		res, err := server.ListInstrumentPrices(ctx, &pb.ListInstrumentPricesRequest{
			Symbol: &sym,
			Limit:  10,
		})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if res.TotalCount != 1 || len(res.Prices) != 1 {
			t.Errorf("expected 1 price, got: %d", len(res.Prices))
		}
	})

	t.Run("invalid date format returns InvalidArgument", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		badDate := "02-10-2026"
		_, err := server.ListInstrumentPrices(ctx, &pb.ListInstrumentPricesRequest{
			FromDate: &badDate,
		})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		st, _ := status.FromError(err)
		if st.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got: %v", st.Code())
		}
	})
}

func TestPortfolioServer_RecordPriceOverride(t *testing.T) {
	ctx := context.Background()
	overridePrice := decimal.NewFromFloat(228.45)
	priceDate := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)

	t.Run("success returns overridden price item", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		mockSvc.EXPECT().
			RecordPriceOverride(ctx, gomock.Any()).
			Return(&domain.InstrumentPrice{
				InstrumentID: uuid.New(),
				Symbol:       "AAPL",
				PriceDate:    priceDate,
				Close:        overridePrice,
				CurrencyCode: "USD",
				Source:       "manual",
			}, true, nil)

		reason := "ECB fixing adjustment"
		res, err := server.RecordPriceOverride(ctx, &pb.RecordPriceOverrideRequest{
			Symbol:              "AAPL",
			PriceDate:           "2026-10-02",
			Price:               dec("228.45"),
			Reason:              &reason,
			RecomputeValuations: true,
		})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if !res.ValuationsRecomputed {
			t.Errorf("expected ValuationsRecomputed true")
		}
		if res.Price.Symbol != "AAPL" {
			t.Errorf("expected symbol AAPL, got: %s", res.Price.Symbol)
		}
	})

	t.Run("instrument not found returns NotFound", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		mockSvc.EXPECT().
			RecordPriceOverride(ctx, gomock.Any()).
			Return(nil, false, repository.ErrInstrumentNotFound)

		_, err := server.RecordPriceOverride(ctx, &pb.RecordPriceOverrideRequest{
			Symbol:    "UNKNOWN",
			PriceDate: "2026-10-02",
			Price:     dec("100.00"),
		})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		st, _ := status.FromError(err)
		if st.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got: %v", st.Code())
		}
	})
}

func TestPortfolioServer_GetIngestionStatus(t *testing.T) {
	ctx := context.Background()

	t.Run("success returns feed diagnostics", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		now := time.Now().UTC()
		mockSvc.EXPECT().
			GetIngestionStatus(ctx).
			Return(&domain.IngestionStatus{
				Feeds: []domain.FeedHealth{
					{
						Name:     "EOD Equity Feeds",
						Status:   "ACTIVE",
						Provider: "Twelve Data",
						Schedule: "Daily at 21:00 UTC",
						LastRun:  now,
						Details:  "Healthy",
					},
				},
				TrackedInstruments:  5,
				TrackedCurrencies:   4,
				RateLimitRemaining:  800,
				RateLimitBudget:     800,
				PendingBackfillJobs: 0,
			}, nil)

		res, err := server.GetIngestionStatus(ctx, &pb.GetIngestionStatusRequest{})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if res.TrackedInstruments != 5 || len(res.Feeds) != 1 {
			t.Errorf("unexpected ingestion status response: %+v", res)
		}
	})
}

func TestPortfolioServer_TriggerMarketSync(t *testing.T) {
	ctx := context.Background()

	t.Run("success returns sync outcome", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		mockSvc.EXPECT().
			TriggerMarketSync(ctx, []string{"AAPL"}, true).
			Return(&domain.MarketSyncResult{
				Success:       true,
				PricesSynced:  1,
				FXRatesSynced: 7,
				Message:       "Sync successful",
			}, nil)

		res, err := server.TriggerMarketSync(ctx, &pb.TriggerMarketSyncRequest{
			Symbols: []string{"AAPL"},
			SyncFx:  true,
		})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if !res.Success || res.PricesSynced != 1 || res.FxRatesSynced != 7 {
			t.Errorf("unexpected sync response: %+v", res)
		}
	})
}

func TestPortfolioServer_RebuildValuations(t *testing.T) {
	ctx := context.Background()

	t.Run("success with explicit from_date", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		expectedDate := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
		mockSvc.EXPECT().RebuildValuations(ctx, "user-123", &expectedDate).Return(nil)

		res, err := server.RebuildValuations(ctx, &pb.RebuildValuationsRequest{
			UserId:   "user-123",
			FromDate: "2025-06-01",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res == nil || !res.Success {
			t.Fatalf("expected successful response, got: %+v", res)
		}
	})

	t.Run("success with empty from_date", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		mockSvc.EXPECT().RebuildValuations(ctx, "user-123", nil).Return(nil)

		res, err := server.RebuildValuations(ctx, &pb.RebuildValuationsRequest{
			UserId: "user-123",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res == nil || !res.Success {
			t.Fatalf("expected successful response, got: %+v", res)
		}
	})

	t.Run("defaults user_id to 1 if empty", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		mockSvc.EXPECT().RebuildValuations(ctx, "1", nil).Return(nil)

		res, err := server.RebuildValuations(ctx, &pb.RebuildValuationsRequest{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !res.Success {
			t.Fatalf("expected success")
		}
	})

	t.Run("invalid from_date format returns InvalidArgument", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		res, err := server.RebuildValuations(ctx, &pb.RebuildValuationsRequest{
			UserId:   "user-123",
			FromDate: "06-01-2025",
		})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if res != nil {
			t.Fatalf("expected nil response on error")
		}
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got %v", err)
		}
	})

	t.Run("portfolio not found returns NotFound", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		mockSvc.EXPECT().RebuildValuations(ctx, "missing-user", nil).Return(repository.ErrPortfolioNotFound)

		res, err := server.RebuildValuations(ctx, &pb.RebuildValuationsRequest{
			UserId: "missing-user",
		})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if res != nil {
			t.Fatalf("expected nil response on error")
		}
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", err)
		}
	})

	t.Run("service unavailable when svc is nil", func(t *testing.T) {
		server := NewPortfolioServer(nil)

		res, err := server.RebuildValuations(ctx, &pb.RebuildValuationsRequest{})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if res != nil {
			t.Fatalf("expected nil response")
		}
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.Unavailable {
			t.Errorf("expected Unavailable, got %v", err)
		}
	})

	t.Run("internal error on service failure", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		mockSvc.EXPECT().RebuildValuations(ctx, "user-123", nil).Return(errors.New("db connection failure"))

		res, err := server.RebuildValuations(ctx, &pb.RebuildValuationsRequest{
			UserId: "user-123",
		})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if res != nil {
			t.Fatalf("expected nil response")
		}
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.Internal {
			t.Errorf("expected Internal, got %v", err)
		}
	})
}

func TestPortfolioServer_ListCurrencyPairs(t *testing.T) {
	ctx := context.Background()

	t.Run("success returns currency pairs", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		prevRate := decimal.RequireFromString("1.0800000000")
		changeAmt := decimal.RequireFromString("0.0050000000")
		changePct := decimal.RequireFromString("0.462963")

		mockPairs := []domain.CurrencyPairSummary{
			{
				BaseCurrency:   "EUR",
				QuoteCurrency:  "USD",
				LatestRate:     decimal.RequireFromString("1.0850000000"),
				LatestDate:     time.Date(2026, 10, 7, 0, 0, 0, 0, time.UTC),
				LatestSource:   "ECB",
				PreviousRate:   &prevRate,
				Change1DAmount: &changeAmt,
				Change1DPct:    &changePct,
				TotalRecords:   100,
				FirstDate:      time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC),
				LastDate:       time.Date(2026, 10, 7, 0, 0, 0, 0, time.UTC),
			},
		}

		mockSvc.EXPECT().ListCurrencyPairs(ctx).Return(mockPairs, nil)

		res, err := server.ListCurrencyPairs(ctx, &pb.ListCurrencyPairsRequest{})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if len(res.Pairs) != 1 {
			t.Fatalf("expected 1 pair, got: %d", len(res.Pairs))
		}
		if res.Pairs[0].BaseCurrency != "EUR" || res.Pairs[0].QuoteCurrency != "USD" {
			t.Errorf("expected EUR/USD, got: %s/%s", res.Pairs[0].BaseCurrency, res.Pairs[0].QuoteCurrency)
		}
		if res.Pairs[0].TotalRecords != 100 {
			t.Errorf("expected 100 total records, got: %d", res.Pairs[0].TotalRecords)
		}
		if res.Pairs[0].PreviousRate == nil || res.Pairs[0].Change_1DAmount == nil || res.Pairs[0].Change_1DPct == nil {
			t.Errorf("expected non-nil optional rate fields")
		}
	})

	t.Run("service unavailable when svc is nil", func(t *testing.T) {
		server := NewPortfolioServer(nil)
		_, err := server.ListCurrencyPairs(ctx, &pb.ListCurrencyPairsRequest{})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.Unavailable {
			t.Errorf("expected Unavailable, got %v", err)
		}
	})

	t.Run("internal error on service failure", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		mockSvc.EXPECT().ListCurrencyPairs(ctx).Return(nil, errors.New("query failed"))

		_, err := server.ListCurrencyPairs(ctx, &pb.ListCurrencyPairsRequest{})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.Internal {
			t.Errorf("expected Internal, got %v", err)
		}
	})
}

func TestPortfolioServer_ListFXRates(t *testing.T) {
	ctx := context.Background()

	t.Run("success with filters and pagination", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		mockRates := []domain.FXRate{
			{
				BaseCurrency:  "EUR",
				QuoteCurrency: "USD",
				RateDate:      time.Date(2026, 10, 7, 0, 0, 0, 0, time.UTC),
				Rate:          decimal.RequireFromString("1.0850000000"),
				Source:        "ECB",
			},
		}

		mockSvc.EXPECT().
			ListFXRates(ctx, gomock.Any()).
			Return(mockRates, 42, nil)

		base := "EUR"
		quote := "USD"
		fromDate := "2026-10-01"
		toDate := "2026-10-07"
		res, err := server.ListFXRates(ctx, &pb.ListFXRatesRequest{
			BaseCurrency:  &base,
			QuoteCurrency: &quote,
			FromDate:      &fromDate,
			ToDate:        &toDate,
			Limit:         25,
			Offset:        0,
		})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if res.TotalCount != 42 {
			t.Errorf("expected total count 42, got %d", res.TotalCount)
		}
		if len(res.Rates) != 1 {
			t.Fatalf("expected 1 rate, got %d", len(res.Rates))
		}
		if res.Rates[0].BaseCurrency != "EUR" || res.Rates[0].QuoteCurrency != "USD" {
			t.Errorf("expected EUR/USD, got: %s/%s", res.Rates[0].BaseCurrency, res.Rates[0].QuoteCurrency)
		}
		if res.Rates[0].InvertedRate == nil {
			t.Errorf("expected inverted rate populated")
		}
	})

	t.Run("invalid from_date format returns InvalidArgument", func(t *testing.T) {
		server := NewPortfolioServer(mocks.NewMockPortfolioService(gomock.NewController(t)))
		invalidDate := "01-10-2026"
		_, err := server.ListFXRates(ctx, &pb.ListFXRatesRequest{FromDate: &invalidDate})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got %v", err)
		}
	})

	t.Run("invalid to_date format returns InvalidArgument", func(t *testing.T) {
		server := NewPortfolioServer(mocks.NewMockPortfolioService(gomock.NewController(t)))
		invalidDate := "07-10-2026"
		_, err := server.ListFXRates(ctx, &pb.ListFXRatesRequest{ToDate: &invalidDate})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got %v", err)
		}
	})

	t.Run("service validation error returns InvalidArgument", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		mockSvc.EXPECT().
			ListFXRates(ctx, gomock.Any()).
			Return(nil, 0, service.ErrInvalidCurrency)

		invalidBase := "EUROPE"
		_, err := server.ListFXRates(ctx, &pb.ListFXRatesRequest{BaseCurrency: &invalidBase})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got %v", err)
		}
	})

	t.Run("service unavailable when svc is nil", func(t *testing.T) {
		server := NewPortfolioServer(nil)
		_, err := server.ListFXRates(ctx, &pb.ListFXRatesRequest{})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.Unavailable {
			t.Errorf("expected Unavailable, got %v", err)
		}
	})
}

func TestPortfolioServer_GetCurrencyPairHistory(t *testing.T) {
	ctx := context.Background()

	t.Run("success returns history points and statistics", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		mockHist := &domain.CurrencyPairHistory{
			BaseCurrency:  "EUR",
			QuoteCurrency: "USD",
			Points: []domain.FXRatePoint{
				{
					Date:         time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
					Rate:         decimal.RequireFromString("1.0800000000"),
					InvertedRate: decimal.RequireFromString("0.9259259259"),
					Source:       "ECB",
				},
				{
					Date:         time.Date(2026, 10, 7, 0, 0, 0, 0, time.UTC),
					Rate:         decimal.RequireFromString("1.0900000000"),
					InvertedRate: decimal.RequireFromString("0.9174311927"),
					Source:       "ECB",
				},
			},
			StartRate:       decimal.RequireFromString("1.0800000000"),
			EndRate:         decimal.RequireFromString("1.0900000000"),
			PeriodChange:    decimal.RequireFromString("0.0100000000"),
			PeriodChangePct: decimal.RequireFromString("0.925926"),
			PeriodHigh:      decimal.RequireFromString("1.0900000000"),
			PeriodLow:       decimal.RequireFromString("1.0800000000"),
		}

		mockSvc.EXPECT().
			GetCurrencyPairHistory(ctx, "EUR", "USD", domain.Timeframe1M).
			Return(mockHist, nil)

		res, err := server.GetCurrencyPairHistory(ctx, &pb.GetCurrencyPairHistoryRequest{
			BaseCurrency:  "EUR",
			QuoteCurrency: "USD",
			Timeframe:     pb.HistoryTimeframe_HISTORY_TIMEFRAME_1M,
		})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if res.BaseCurrency != "EUR" || res.QuoteCurrency != "USD" {
			t.Errorf("expected EUR/USD, got %s/%s", res.BaseCurrency, res.QuoteCurrency)
		}
		if len(res.Points) != 2 {
			t.Fatalf("expected 2 points, got %d", len(res.Points))
		}
		if res.StartRate == nil || res.EndRate == nil || res.PeriodChange == nil || res.PeriodChangePct == nil {
			t.Errorf("expected summary statistics populated")
		}
	})

	t.Run("missing base currency returns InvalidArgument", func(t *testing.T) {
		server := NewPortfolioServer(mocks.NewMockPortfolioService(gomock.NewController(t)))
		_, err := server.GetCurrencyPairHistory(ctx, &pb.GetCurrencyPairHistoryRequest{
			BaseCurrency:  "",
			QuoteCurrency: "USD",
		})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got %v", err)
		}
	})

	t.Run("missing quote currency returns InvalidArgument", func(t *testing.T) {
		server := NewPortfolioServer(mocks.NewMockPortfolioService(gomock.NewController(t)))
		_, err := server.GetCurrencyPairHistory(ctx, &pb.GetCurrencyPairHistoryRequest{
			BaseCurrency:  "EUR",
			QuoteCurrency: "",
		})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got %v", err)
		}
	})

	t.Run("invalid timeframe returns InvalidArgument", func(t *testing.T) {
		server := NewPortfolioServer(mocks.NewMockPortfolioService(gomock.NewController(t)))
		_, err := server.GetCurrencyPairHistory(ctx, &pb.GetCurrencyPairHistoryRequest{
			BaseCurrency:  "EUR",
			QuoteCurrency: "USD",
			Timeframe:     pb.HistoryTimeframe(999),
		})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got %v", err)
		}
	})

	t.Run("service unavailable when svc is nil", func(t *testing.T) {
		server := NewPortfolioServer(nil)
		_, err := server.GetCurrencyPairHistory(ctx, &pb.GetCurrencyPairHistoryRequest{
			BaseCurrency:  "EUR",
			QuoteCurrency: "USD",
		})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.Unavailable {
			t.Errorf("expected Unavailable, got %v", err)
		}
	})
}

func TestPortfolioServer_RecordFXRateOverride(t *testing.T) {
	ctx := context.Background()
	rateDate := time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC)
	overrideRate := decimal.RequireFromString("1.0950000000")

	t.Run("success returns overridden fx rate item", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		mockSvc.EXPECT().
			RecordFXRateOverride(ctx, gomock.Any()).
			Return(&domain.FXRate{
				BaseCurrency:  "EUR",
				QuoteCurrency: "USD",
				RateDate:      rateDate,
				Rate:          overrideRate,
				Source:        "manual: fix holiday gap",
			}, true, nil)

		reason := "fix holiday gap"
		res, err := server.RecordFXRateOverride(ctx, &pb.RecordFXRateOverrideRequest{
			BaseCurrency:        "EUR",
			QuoteCurrency:       "USD",
			RateDate:            "2026-10-05",
			Rate:                dec("1.0950000000"),
			Reason:              &reason,
			RecomputeValuations: true,
		})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if !res.ValuationsRecomputed {
			t.Errorf("expected ValuationsRecomputed true")
		}
		if res.Rate.BaseCurrency != "EUR" || res.Rate.QuoteCurrency != "USD" {
			t.Errorf("expected EUR/USD, got: %s/%s", res.Rate.BaseCurrency, res.Rate.QuoteCurrency)
		}
		if res.Rate.Source != "manual: fix holiday gap" {
			t.Errorf("expected source manual: fix holiday gap, got: %s", res.Rate.Source)
		}
	})

	t.Run("missing required fields returns InvalidArgument", func(t *testing.T) {
		server := NewPortfolioServer(mocks.NewMockPortfolioService(gomock.NewController(t)))

		// missing base
		_, err := server.RecordFXRateOverride(ctx, &pb.RecordFXRateOverrideRequest{QuoteCurrency: "USD", RateDate: "2026-10-05", Rate: dec("1.0")})
		if err == nil {
			t.Errorf("expected error for missing base")
		}
		// missing quote
		_, err = server.RecordFXRateOverride(ctx, &pb.RecordFXRateOverrideRequest{BaseCurrency: "EUR", RateDate: "2026-10-05", Rate: dec("1.0")})
		if err == nil {
			t.Errorf("expected error for missing quote")
		}
		// missing rate_date
		_, err = server.RecordFXRateOverride(ctx, &pb.RecordFXRateOverrideRequest{BaseCurrency: "EUR", QuoteCurrency: "USD", Rate: dec("1.0")})
		if err == nil {
			t.Errorf("expected error for missing rate_date")
		}
		// missing rate
		_, err = server.RecordFXRateOverride(ctx, &pb.RecordFXRateOverrideRequest{BaseCurrency: "EUR", QuoteCurrency: "USD", RateDate: "2026-10-05"})
		if err == nil {
			t.Errorf("expected error for missing rate")
		}
	})

	t.Run("invalid rate_date format returns InvalidArgument", func(t *testing.T) {
		server := NewPortfolioServer(mocks.NewMockPortfolioService(gomock.NewController(t)))
		_, err := server.RecordFXRateOverride(ctx, &pb.RecordFXRateOverrideRequest{
			BaseCurrency:  "EUR",
			QuoteCurrency: "USD",
			RateDate:      "05-10-2026",
			Rate:          dec("1.0"),
		})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got %v", err)
		}
	})

	t.Run("service validation error returns InvalidArgument", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		mockSvc.EXPECT().
			RecordFXRateOverride(ctx, gomock.Any()).
			Return(nil, false, service.ErrInvalidRate)

		_, err := server.RecordFXRateOverride(ctx, &pb.RecordFXRateOverrideRequest{
			BaseCurrency:  "EUR",
			QuoteCurrency: "USD",
			RateDate:      "2026-10-05",
			Rate:          dec("0.0"),
		})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got %v", err)
		}
	})

	t.Run("service unavailable when svc is nil", func(t *testing.T) {
		server := NewPortfolioServer(nil)
		_, err := server.RecordFXRateOverride(ctx, &pb.RecordFXRateOverrideRequest{
			BaseCurrency:  "EUR",
			QuoteCurrency: "USD",
			RateDate:      "2026-10-05",
			Rate:          dec("1.0"),
		})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.Unavailable {
			t.Errorf("expected Unavailable, got %v", err)
		}
	})
}

func TestPortfolioServer_TriggerBackfill(t *testing.T) {
	ctx := context.Background()

	t.Run("successful backfill execution", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		mockSvc.EXPECT().
			TriggerBackfill(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, in domain.BackfillInput) (*domain.BackfillResult, error) {
				if in.FromDate.Format("2006-01-02") != "2025-01-01" {
					t.Errorf("expected from_date 2025-01-01, got %s", in.FromDate.Format("2006-01-02"))
				}
				if in.ToDate.Format("2006-01-02") != "2025-01-31" {
					t.Errorf("expected to_date 2025-01-31, got %s", in.ToDate.Format("2006-01-02"))
				}
				if len(in.Symbols) != 1 || in.Symbols[0] != "AAPL" {
					t.Errorf("expected symbols [AAPL], got %v", in.Symbols)
				}
				if !in.BackfillAssets || in.BackfillFX || !in.RecomputeValuations {
					t.Errorf("unexpected scope flags: %+v", in)
				}
				return &domain.BackfillResult{
					Success:       true,
					PricesSynced:  22,
					FXRatesSynced: 0,
					Message:       "Historical backfill completed: 22 asset prices stored",
					Warnings:      []string{},
				}, nil
			})

		resp, err := server.TriggerBackfill(ctx, &pb.TriggerBackfillRequest{
			FromDate:            "2025-01-01",
			ToDate:              "2025-01-31",
			Symbols:             []string{"AAPL"},
			BackfillAssets:      true,
			BackfillFx:          false,
			RecomputeValuations: true,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !resp.GetSuccess() {
			t.Errorf("expected success to be true")
		}
		if resp.GetPricesSynced() != 22 {
			t.Errorf("expected 22 prices synced, got %d", resp.GetPricesSynced())
		}
		if resp.GetFxRatesSynced() != 0 {
			t.Errorf("expected 0 fx rates synced, got %d", resp.GetFxRatesSynced())
		}
	})

	t.Run("defaults to_date to now when empty", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		mockSvc.EXPECT().
			TriggerBackfill(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, in domain.BackfillInput) (*domain.BackfillResult, error) {
				if in.ToDate.IsZero() {
					t.Errorf("expected to_date to be set, got zero time")
				}
				return &domain.BackfillResult{
					Success:      true,
					PricesSynced: 10,
				}, nil
			})

		resp, err := server.TriggerBackfill(ctx, &pb.TriggerBackfillRequest{
			FromDate:       "2025-01-01",
			ToDate:         "",
			BackfillAssets: true,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !resp.GetSuccess() {
			t.Errorf("expected success to be true")
		}
	})

	t.Run("missing from_date returns InvalidArgument", func(t *testing.T) {
		server := NewPortfolioServer(mocks.NewMockPortfolioService(gomock.NewController(t)))
		_, err := server.TriggerBackfill(ctx, &pb.TriggerBackfillRequest{
			FromDate: "",
		})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got %v", err)
		}
	})

	t.Run("invalid from_date format returns InvalidArgument", func(t *testing.T) {
		server := NewPortfolioServer(mocks.NewMockPortfolioService(gomock.NewController(t)))
		_, err := server.TriggerBackfill(ctx, &pb.TriggerBackfillRequest{
			FromDate: "01/01/2025",
		})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got %v", err)
		}
	})

	t.Run("invalid to_date format returns InvalidArgument", func(t *testing.T) {
		server := NewPortfolioServer(mocks.NewMockPortfolioService(gomock.NewController(t)))
		_, err := server.TriggerBackfill(ctx, &pb.TriggerBackfillRequest{
			FromDate: "2025-01-01",
			ToDate:   "invalid",
		})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got %v", err)
		}
	})

	t.Run("service validation error returns InvalidArgument", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		mockSvc.EXPECT().
			TriggerBackfill(ctx, gomock.Any()).
			Return(nil, service.ErrInvalidDateRange)

		_, err := server.TriggerBackfill(ctx, &pb.TriggerBackfillRequest{
			FromDate:       "2025-02-01",
			ToDate:         "2025-01-01",
			BackfillAssets: true,
		})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got %v", err)
		}
	})

	t.Run("service internal error returns Internal", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSvc := mocks.NewMockPortfolioService(ctrl)
		server := NewPortfolioServer(mockSvc)

		mockSvc.EXPECT().
			TriggerBackfill(ctx, gomock.Any()).
			Return(nil, errors.New("db connection failure"))

		_, err := server.TriggerBackfill(ctx, &pb.TriggerBackfillRequest{
			FromDate:       "2025-01-01",
			ToDate:         "2025-01-31",
			BackfillAssets: true,
		})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.Internal {
			t.Errorf("expected Internal, got %v", err)
		}
	})

	t.Run("service unavailable when svc is nil", func(t *testing.T) {
		server := NewPortfolioServer(nil)
		_, err := server.TriggerBackfill(ctx, &pb.TriggerBackfillRequest{
			FromDate: "2025-01-01",
		})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.Unavailable {
			t.Errorf("expected Unavailable, got %v", err)
		}
	})
}

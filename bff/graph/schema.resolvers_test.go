package graph

import (
	"context"
	"testing"

	"bff/graph/model"

	commonpb "graphfolio/proto/common/v1"
	pb "graphfolio/proto/portfolio/v1"

	"github.com/shopspring/decimal"
	"google.golang.org/grpc"
)

type fakePortfolioClient struct {
	pb.PortfolioServiceClient
	getPortfolioFn        func(ctx context.Context, in *pb.GetPortfolioRequest) (*pb.GetPortfolioResponse, error)
	addTransactionFn      func(ctx context.Context, in *pb.AddTransactionRequest) (*pb.AddTransactionResponse, error)
	listInstrumentsFn     func(ctx context.Context, in *pb.ListInstrumentsRequest) (*pb.ListInstrumentsResponse, error)
	getPortfolioHistoryFn func(ctx context.Context, in *pb.GetPortfolioHistoryRequest) (*pb.GetPortfolioHistoryResponse, error)
}

func (f *fakePortfolioClient) GetPortfolio(ctx context.Context, in *pb.GetPortfolioRequest, opts ...grpc.CallOption) (*pb.GetPortfolioResponse, error) {
	if f.getPortfolioFn != nil {
		return f.getPortfolioFn(ctx, in)
	}
	return nil, nil
}

func (f *fakePortfolioClient) AddTransaction(ctx context.Context, in *pb.AddTransactionRequest, opts ...grpc.CallOption) (*pb.AddTransactionResponse, error) {
	if f.addTransactionFn != nil {
		return f.addTransactionFn(ctx, in)
	}
	return nil, nil
}

func (f *fakePortfolioClient) ListInstruments(ctx context.Context, in *pb.ListInstrumentsRequest, opts ...grpc.CallOption) (*pb.ListInstrumentsResponse, error) {
	if f.listInstrumentsFn != nil {
		return f.listInstrumentsFn(ctx, in)
	}
	return nil, nil
}

func (f *fakePortfolioClient) GetPortfolioHistory(ctx context.Context, in *pb.GetPortfolioHistoryRequest, opts ...grpc.CallOption) (*pb.GetPortfolioHistoryResponse, error) {
	if f.getPortfolioHistoryFn != nil {
		return f.getPortfolioHistoryFn(ctx, in)
	}
	return nil, nil
}

func TestQueryResolver_Portfolio(t *testing.T) {
	ctx := context.Background()

	fakeClient := &fakePortfolioClient{
		getPortfolioFn: func(ctx context.Context, in *pb.GetPortfolioRequest) (*pb.GetPortfolioResponse, error) {
			if in.GetUserId() != "1" {
				t.Errorf("expected user_id 1, got %s", in.GetUserId())
			}
			return &pb.GetPortfolioResponse{
				Portfolio: &pb.Portfolio{
					TotalValue: &commonpb.Money{
						Amount:       &commonpb.Decimal{Value: "150000.00"},
						CurrencyCode: "USD",
					},
					CashBalance: &commonpb.Money{
						Amount:       &commonpb.Decimal{Value: "5000.00"},
						CurrencyCode: "USD",
					},
				},
			}, nil
		},
	}

	resolver := &Resolver{PortfolioClient: fakeClient}
	qResolver := resolver.Query()

	result, err := qResolver.Portfolio(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatalf("expected result, got nil")
	}
	if result.TotalValue.Amount.String() != "150000" && result.TotalValue.Amount.String() != "150000.00" {
		t.Errorf("expected total value 150000.00, got %s", result.TotalValue.Amount.String())
	}
}

func TestQueryResolver_Instruments(t *testing.T) {
	ctx := context.Background()

	fakeClient := &fakePortfolioClient{
		listInstrumentsFn: func(ctx context.Context, in *pb.ListInstrumentsRequest) (*pb.ListInstrumentsResponse, error) {
			return &pb.ListInstrumentsResponse{
				Instruments: []*pb.Instrument{
					{
						Id:           "inst-1",
						Symbol:       "AAPL",
						Name:         "Apple Inc.",
						CurrencyCode: "USD",
						AssetClass:   "EQUITY",
					},
				},
			}, nil
		},
	}

	resolver := &Resolver{PortfolioClient: fakeClient}
	qResolver := resolver.Query()

	insts, err := qResolver.Instruments(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(insts) != 1 {
		t.Fatalf("expected 1 instrument, got %d", len(insts))
	}
	if insts[0].Symbol != "AAPL" {
		t.Errorf("expected AAPL, got %s", insts[0].Symbol)
	}
}

func TestMutationResolver_AddTransaction(t *testing.T) {
	ctx := context.Background()

	fakeClient := &fakePortfolioClient{
		addTransactionFn: func(ctx context.Context, in *pb.AddTransactionRequest) (*pb.AddTransactionResponse, error) {
			if in.GetType() != pb.TransactionType_TRANSACTION_TYPE_BUY {
				t.Errorf("expected BUY transaction, got %v", in.GetType())
			}
			if in.GetSymbol() != "AAPL" {
				t.Errorf("expected AAPL, got %s", in.GetSymbol())
			}
			return &pb.AddTransactionResponse{
				TransactionId: "tx-new-123",
				Portfolio: &pb.Portfolio{
					TotalValue: &commonpb.Money{
						Amount:       &commonpb.Decimal{Value: "160000.00"},
						CurrencyCode: "USD",
					},
					CashBalance: &commonpb.Money{
						Amount:       &commonpb.Decimal{Value: "3500.00"},
						CurrencyCode: "USD",
					},
				},
			}, nil
		},
	}

	resolver := &Resolver{PortfolioClient: fakeClient}
	mResolver := resolver.Mutation()

	sym := "AAPL"
	qty := model.Decimal(decimal.NewFromInt(10))
	price := model.Decimal(decimal.RequireFromString("150.00"))
	fee := model.Decimal(decimal.RequireFromString("5.00"))

	payload, err := mResolver.AddTransaction(ctx, model.AddTransactionInput{
		Type:      model.TransactionTypeBuy,
		Symbol:    &sym,
		TradeDate: "2025-01-05",
		Quantity:  &qty,
		Price:     &price,
		Fee:       &fee,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload == nil {
		t.Fatalf("expected payload, got nil")
	}
	if payload.TransactionID != "tx-new-123" {
		t.Errorf("expected tx ID tx-new-123, got %s", payload.TransactionID)
	}
	if payload.Portfolio == nil {
		t.Fatalf("expected portfolio in payload, got nil")
	}
}

func TestQueryResolver_PortfolioHistory(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully maps valuation time series and returns", func(t *testing.T) {
		fakeClient := &fakePortfolioClient{
			getPortfolioHistoryFn: func(ctx context.Context, in *pb.GetPortfolioHistoryRequest) (*pb.GetPortfolioHistoryResponse, error) {
				if in.GetUserId() != "1" {
					t.Errorf("expected user_id 1, got %s", in.GetUserId())
				}
				if in.GetTimeframe() != pb.HistoryTimeframe_HISTORY_TIMEFRAME_1M {
					t.Errorf("expected timeframe 1M, got %v", in.GetTimeframe())
				}
				return &pb.GetPortfolioHistoryResponse{
					Points: []*pb.ValuationPoint{
						{
							Date: "2026-04-01",
							TotalValue: &commonpb.Money{
								Amount:       &commonpb.Decimal{Value: "100000.00"},
								CurrencyCode: "USD",
							},
							MarketValue: &commonpb.Money{
								Amount:       &commonpb.Decimal{Value: "95000.00"},
								CurrencyCode: "USD",
							},
							CashValue: &commonpb.Money{
								Amount:       &commonpb.Decimal{Value: "5000.00"},
								CurrencyCode: "USD",
							},
							TwrIndex:    &commonpb.Decimal{Value: "1.0000"},
							DailyReturn: &commonpb.Decimal{Value: "0.0000"},
						},
						{
							Date: "2026-05-01",
							TotalValue: &commonpb.Money{
								Amount:       &commonpb.Decimal{Value: "115000.00"},
								CurrencyCode: "USD",
							},
							MarketValue: &commonpb.Money{
								Amount:       &commonpb.Decimal{Value: "110000.00"},
								CurrencyCode: "USD",
							},
							CashValue: &commonpb.Money{
								Amount:       &commonpb.Decimal{Value: "5000.00"},
								CurrencyCode: "USD",
							},
							TwrIndex:    &commonpb.Decimal{Value: "1.1500"},
							DailyReturn: &commonpb.Decimal{Value: "0.0200"},
						},
					},
					StartValue: &commonpb.Money{
						Amount:       &commonpb.Decimal{Value: "100000.00"},
						CurrencyCode: "USD",
					},
					EndValue: &commonpb.Money{
						Amount:       &commonpb.Decimal{Value: "115000.00"},
						CurrencyCode: "USD",
					},
					ReturnAmount: &commonpb.Money{
						Amount:       &commonpb.Decimal{Value: "15000.00"},
						CurrencyCode: "USD",
					},
					ReturnPercent: &commonpb.Decimal{Value: "15.00"},
				}, nil
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		qResolver := resolver.Query()

		res, err := qResolver.PortfolioHistory(ctx, model.HistoryTimeframeTimeframe1m)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res == nil {
			t.Fatalf("expected result, got nil")
		}
		if len(res.Points) != 2 {
			t.Fatalf("expected 2 points, got %d", len(res.Points))
		}
		if res.Points[0].Date != "2026-04-01" {
			t.Errorf("expected date 2026-04-01, got %s", res.Points[0].Date)
		}
		if res.Points[0].TotalValue.CurrencyCode != "USD" {
			t.Errorf("expected USD currency, got %s", res.Points[0].TotalValue.CurrencyCode)
		}
		if res.Points[1].DailyReturn == nil || res.Points[1].DailyReturn.String() != "0.02" {
			t.Errorf("expected daily return 0.02, got %v", res.Points[1].DailyReturn)
		}
		if res.StartValue.Amount.String() != "100000" && res.StartValue.Amount.String() != "100000.00" {
			t.Errorf("expected start value 100000.00, got %s", res.StartValue.Amount.String())
		}
		if res.EndValue.Amount.String() != "115000" && res.EndValue.Amount.String() != "115000.00" {
			t.Errorf("expected end value 115000.00, got %s", res.EndValue.Amount.String())
		}
		if res.ReturnAmount.Amount.String() != "15000" && res.ReturnAmount.Amount.String() != "15000.00" {
			t.Errorf("expected return amount 15000.00, got %s", res.ReturnAmount.Amount.String())
		}
		if res.ReturnPercent.String() != "15" && res.ReturnPercent.String() != "15.00" {
			t.Errorf("expected return percent 15.00, got %s", res.ReturnPercent.String())
		}
	})

	t.Run("propagates client error", func(t *testing.T) {
		fakeClient := &fakePortfolioClient{
			getPortfolioHistoryFn: func(ctx context.Context, in *pb.GetPortfolioHistoryRequest) (*pb.GetPortfolioHistoryResponse, error) {
				return nil, context.DeadlineExceeded
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		qResolver := resolver.Query()

		_, err := qResolver.PortfolioHistory(ctx, model.HistoryTimeframeTimeframe1y)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if err != context.DeadlineExceeded {
			t.Errorf("expected DeadlineExceeded, got %v", err)
		}
	})
}

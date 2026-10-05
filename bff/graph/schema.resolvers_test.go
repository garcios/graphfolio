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
	listTransactionsFn    func(ctx context.Context, in *pb.ListTransactionsRequest) (*pb.ListTransactionsResponse, error)
	deleteTransactionFn   func(ctx context.Context, in *pb.DeleteTransactionRequest) (*pb.DeleteTransactionResponse, error)
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

func (f *fakePortfolioClient) ListTransactions(ctx context.Context, in *pb.ListTransactionsRequest, opts ...grpc.CallOption) (*pb.ListTransactionsResponse, error) {
	if f.listTransactionsFn != nil {
		return f.listTransactionsFn(ctx, in)
	}
	return nil, nil
}

func (f *fakePortfolioClient) DeleteTransaction(ctx context.Context, in *pb.DeleteTransactionRequest, opts ...grpc.CallOption) (*pb.DeleteTransactionResponse, error) {
	if f.deleteTransactionFn != nil {
		return f.deleteTransactionFn(ctx, in)
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

func TestQueryResolver_Transactions(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully maps transactions connection and pagination", func(t *testing.T) {
		typeArg := model.TransactionTypeBuy
		symbolArg := "AAPL"
		pageArg := 2
		pageSizeArg := 10

		fakeClient := &fakePortfolioClient{
			listTransactionsFn: func(ctx context.Context, in *pb.ListTransactionsRequest) (*pb.ListTransactionsResponse, error) {
				if in.GetUserId() != "1" {
					t.Errorf("expected user_id 1, got %s", in.GetUserId())
				}
				if in.GetType() != pb.TransactionType_TRANSACTION_TYPE_BUY {
					t.Errorf("expected type BUY, got %v", in.GetType())
				}
				if in.GetSymbol() != "AAPL" {
					t.Errorf("expected symbol AAPL, got %s", in.GetSymbol())
				}
				if in.GetPage() != 2 {
					t.Errorf("expected page 2, got %d", in.GetPage())
				}
				if in.GetPageSize() != 10 {
					t.Errorf("expected pageSize 10, got %d", in.GetPageSize())
				}

				return &pb.ListTransactionsResponse{
					Transactions: []*pb.TransactionItem{
						{
							Id:             "tx-1",
							Type:           pb.TransactionType_TRANSACTION_TYPE_BUY,
							Symbol:         "AAPL",
							InstrumentName: "Apple Inc.",
							TradeDate:      "2026-02-01",
							Quantity:       &commonpb.Decimal{Value: "10.0"},
							Price: &commonpb.Money{
								Amount:       &commonpb.Decimal{Value: "150.00"},
								CurrencyCode: "USD",
							},
							Amount: &commonpb.Money{
								Amount:       &commonpb.Decimal{Value: "1500.00"},
								CurrencyCode: "USD",
							},
							Fee: &commonpb.Money{
								Amount:       &commonpb.Decimal{Value: "5.00"},
								CurrencyCode: "USD",
							},
							Notes:     "Purchased shares",
							CreatedAt: "2026-02-01T12:00:00Z",
						},
					},
					TotalCount: 42,
					Page:       2,
					PageSize:   10,
				}, nil
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		qResolver := resolver.Query()

		res, err := qResolver.Transactions(ctx, &typeArg, &symbolArg, &pageArg, &pageSizeArg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res == nil {
			t.Fatalf("expected non-nil response")
		}
		if res.TotalCount != 42 {
			t.Errorf("expected total count 42, got %d", res.TotalCount)
		}
		if res.Page != 2 {
			t.Errorf("expected page 2, got %d", res.Page)
		}
		if res.PageSize != 10 {
			t.Errorf("expected page size 10, got %d", res.PageSize)
		}
		if len(res.Items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(res.Items))
		}

		item := res.Items[0]
		if item.ID != "tx-1" {
			t.Errorf("expected id tx-1, got %s", item.ID)
		}
		if item.Type != model.TransactionTypeBuy {
			t.Errorf("expected type BUY, got %v", item.Type)
		}
		if item.Symbol == nil || *item.Symbol != "AAPL" {
			t.Errorf("expected symbol AAPL, got %v", item.Symbol)
		}
		if item.InstrumentName == nil || *item.InstrumentName != "Apple Inc." {
			t.Errorf("expected name Apple Inc., got %v", item.InstrumentName)
		}
		if item.Quantity == nil || item.Quantity.String() != "10" {
			t.Errorf("expected quantity 10, got %v", item.Quantity)
		}
		if item.Price == nil || item.Price.Amount.String() != "15000" && item.Price.Amount.String() != "150" && item.Price.Amount.String() != "150.00" {
			t.Errorf("expected price 150.00, got %v", item.Price)
		}
		if item.Amount.Amount.String() != "1500" && item.Amount.Amount.String() != "1500.00" {
			t.Errorf("expected amount 1500.00, got %s", item.Amount.Amount.String())
		}
		if item.Fee.Amount.String() != "5" && item.Fee.Amount.String() != "5.00" {
			t.Errorf("expected fee 5.00, got %s", item.Fee.Amount.String())
		}
		if item.Notes == nil || *item.Notes != "Purchased shares" {
			t.Errorf("expected notes, got %v", item.Notes)
		}
		if item.CreatedAt != "2026-02-01T12:00:00Z" {
			t.Errorf("expected created at, got %s", item.CreatedAt)
		}
	})

	t.Run("propagates client error", func(t *testing.T) {
		fakeClient := &fakePortfolioClient{
			listTransactionsFn: func(ctx context.Context, in *pb.ListTransactionsRequest) (*pb.ListTransactionsResponse, error) {
				return nil, context.DeadlineExceeded
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		qResolver := resolver.Query()

		_, err := qResolver.Transactions(ctx, nil, nil, nil, nil)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if err != context.DeadlineExceeded {
			t.Errorf("expected DeadlineExceeded, got %v", err)
		}
	})
}

func TestMutationResolver_DeleteTransaction(t *testing.T) {
	ctx := context.Background()

	t.Run("success deletes transaction and returns updated portfolio", func(t *testing.T) {
		fakeClient := &fakePortfolioClient{
			deleteTransactionFn: func(ctx context.Context, in *pb.DeleteTransactionRequest) (*pb.DeleteTransactionResponse, error) {
				if in.GetUserId() != "1" {
					t.Errorf("expected user_id 1, got %s", in.GetUserId())
				}
				if in.GetTransactionId() != "tx-to-delete-456" {
					t.Errorf("expected tx ID tx-to-delete-456, got %s", in.GetTransactionId())
				}
				return &pb.DeleteTransactionResponse{
					Success: true,
					Portfolio: &pb.Portfolio{
						TotalValue: &commonpb.Money{
							Amount:       &commonpb.Decimal{Value: "145000.00"},
							CurrencyCode: "USD",
						},
						CashBalance: &commonpb.Money{
							Amount:       &commonpb.Decimal{Value: "10000.00"},
							CurrencyCode: "USD",
						},
					},
				}, nil
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		mResolver := resolver.Mutation()

		payload, err := mResolver.DeleteTransaction(ctx, "tx-to-delete-456")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if payload == nil {
			t.Fatalf("expected non-nil payload")
		}
		if !payload.Success {
			t.Errorf("expected Success=true")
		}
		if payload.Portfolio == nil {
			t.Fatalf("expected portfolio in payload")
		}
		if payload.Portfolio.TotalValue.Amount.String() != "145000" && payload.Portfolio.TotalValue.Amount.String() != "145000.00" {
			t.Errorf("expected total value 145000.00, got %s", payload.Portfolio.TotalValue.Amount.String())
		}
	})

	t.Run("propagates client error", func(t *testing.T) {
		fakeClient := &fakePortfolioClient{
			deleteTransactionFn: func(ctx context.Context, in *pb.DeleteTransactionRequest) (*pb.DeleteTransactionResponse, error) {
				return nil, context.DeadlineExceeded
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		mResolver := resolver.Mutation()

		_, err := mResolver.DeleteTransaction(ctx, "tx-fail")
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if err != context.DeadlineExceeded {
			t.Errorf("expected DeadlineExceeded, got %v", err)
		}
	})
}

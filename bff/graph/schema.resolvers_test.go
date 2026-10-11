package graph

import (
	"context"
	"errors"
	"testing"

	"bff/graph/model"

	commonpb "graphfolio/proto/common/v1"
	pb "graphfolio/proto/portfolio/v1"
	userpb "graphfolio/proto/user/v1"

	"github.com/shopspring/decimal"
	"google.golang.org/grpc"
)

type fakeUserClient struct {
	userpb.UserServiceClient
	getUserPreferencesFn      func(ctx context.Context, in *userpb.GetUserPreferencesRequest) (*userpb.GetUserPreferencesResponse, error)
	updateUserPreferencesFn   func(ctx context.Context, in *userpb.UpdateUserPreferencesRequest) (*userpb.UpdateUserPreferencesResponse, error)
	listSupportedCurrenciesFn func(ctx context.Context, in *userpb.ListSupportedCurrenciesRequest) (*userpb.ListSupportedCurrenciesResponse, error)
}

func (f *fakeUserClient) GetUserPreferences(ctx context.Context, in *userpb.GetUserPreferencesRequest, opts ...grpc.CallOption) (*userpb.GetUserPreferencesResponse, error) {
	if f.getUserPreferencesFn != nil {
		return f.getUserPreferencesFn(ctx, in)
	}
	return nil, nil
}

func (f *fakeUserClient) UpdateUserPreferences(ctx context.Context, in *userpb.UpdateUserPreferencesRequest, opts ...grpc.CallOption) (*userpb.UpdateUserPreferencesResponse, error) {
	if f.updateUserPreferencesFn != nil {
		return f.updateUserPreferencesFn(ctx, in)
	}
	return nil, nil
}

func (f *fakeUserClient) ListSupportedCurrencies(ctx context.Context, in *userpb.ListSupportedCurrenciesRequest, opts ...grpc.CallOption) (*userpb.ListSupportedCurrenciesResponse, error) {
	if f.listSupportedCurrenciesFn != nil {
		return f.listSupportedCurrenciesFn(ctx, in)
	}
	return nil, nil
}

type fakePortfolioClient struct {
	pb.PortfolioServiceClient
	getPortfolioFn                func(ctx context.Context, in *pb.GetPortfolioRequest) (*pb.GetPortfolioResponse, error)
	updatePortfolioBaseCurrencyFn func(ctx context.Context, in *pb.UpdatePortfolioBaseCurrencyRequest) (*pb.UpdatePortfolioBaseCurrencyResponse, error)
	addTransactionFn              func(ctx context.Context, in *pb.AddTransactionRequest) (*pb.AddTransactionResponse, error)
	listInstrumentsFn             func(ctx context.Context, in *pb.ListInstrumentsRequest) (*pb.ListInstrumentsResponse, error)
	getPortfolioHistoryFn         func(ctx context.Context, in *pb.GetPortfolioHistoryRequest) (*pb.GetPortfolioHistoryResponse, error)
	listTransactionsFn            func(ctx context.Context, in *pb.ListTransactionsRequest) (*pb.ListTransactionsResponse, error)
	deleteTransactionFn           func(ctx context.Context, in *pb.DeleteTransactionRequest) (*pb.DeleteTransactionResponse, error)
	listAllInstrumentsFn          func(ctx context.Context, in *pb.ListAllInstrumentsRequest) (*pb.ListAllInstrumentsResponse, error)
	createInstrumentFn            func(ctx context.Context, in *pb.CreateInstrumentRequest) (*pb.CreateInstrumentResponse, error)
	updateInstrumentFn            func(ctx context.Context, in *pb.UpdateInstrumentRequest) (*pb.UpdateInstrumentResponse, error)
	deleteInstrumentFn            func(ctx context.Context, in *pb.DeleteInstrumentRequest) (*pb.DeleteInstrumentResponse, error)
	listExchangesFn               func(ctx context.Context, in *pb.ListExchangesRequest) (*pb.ListExchangesResponse, error)
	listInstrumentPricesFn        func(ctx context.Context, in *pb.ListInstrumentPricesRequest) (*pb.ListInstrumentPricesResponse, error)
	recordPriceOverrideFn         func(ctx context.Context, in *pb.RecordPriceOverrideRequest) (*pb.RecordPriceOverrideResponse, error)
	getIngestionStatusFn          func(ctx context.Context, in *pb.GetIngestionStatusRequest) (*pb.GetIngestionStatusResponse, error)
	triggerMarketSyncFn           func(ctx context.Context, in *pb.TriggerMarketSyncRequest) (*pb.TriggerMarketSyncResponse, error)
	triggerBackfillFn             func(ctx context.Context, in *pb.TriggerBackfillRequest) (*pb.TriggerBackfillResponse, error)
	listCurrencyPairsFn           func(ctx context.Context, in *pb.ListCurrencyPairsRequest) (*pb.ListCurrencyPairsResponse, error)
	listFXRatesFn                 func(ctx context.Context, in *pb.ListFXRatesRequest) (*pb.ListFXRatesResponse, error)
	getCurrencyPairHistoryFn      func(ctx context.Context, in *pb.GetCurrencyPairHistoryRequest) (*pb.GetCurrencyPairHistoryResponse, error)
	recordFXRateOverrideFn        func(ctx context.Context, in *pb.RecordFXRateOverrideRequest) (*pb.RecordFXRateOverrideResponse, error)
	checkTransactionDuplicatesFn  func(ctx context.Context, in *pb.CheckTransactionDuplicatesRequest) (*pb.CheckTransactionDuplicatesResponse, error)
	batchImportTransactionsFn     func(ctx context.Context, in *pb.BatchImportTransactionsRequest) (*pb.BatchImportTransactionsResponse, error)
	getCashFlowReportFn           func(ctx context.Context, in *pb.GetCashFlowReportRequest) (*pb.GetCashFlowReportResponse, error)
}

func (f *fakePortfolioClient) GetPortfolio(ctx context.Context, in *pb.GetPortfolioRequest, opts ...grpc.CallOption) (*pb.GetPortfolioResponse, error) {
	if f.getPortfolioFn != nil {
		return f.getPortfolioFn(ctx, in)
	}
	return nil, nil
}

func (f *fakePortfolioClient) UpdatePortfolioBaseCurrency(ctx context.Context, in *pb.UpdatePortfolioBaseCurrencyRequest, opts ...grpc.CallOption) (*pb.UpdatePortfolioBaseCurrencyResponse, error) {
	if f.updatePortfolioBaseCurrencyFn != nil {
		return f.updatePortfolioBaseCurrencyFn(ctx, in)
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

func (f *fakePortfolioClient) ListAllInstruments(ctx context.Context, in *pb.ListAllInstrumentsRequest, opts ...grpc.CallOption) (*pb.ListAllInstrumentsResponse, error) {
	if f.listAllInstrumentsFn != nil {
		return f.listAllInstrumentsFn(ctx, in)
	}
	return nil, nil
}

func (f *fakePortfolioClient) CreateInstrument(ctx context.Context, in *pb.CreateInstrumentRequest, opts ...grpc.CallOption) (*pb.CreateInstrumentResponse, error) {
	if f.createInstrumentFn != nil {
		return f.createInstrumentFn(ctx, in)
	}
	return nil, nil
}

func (f *fakePortfolioClient) UpdateInstrument(ctx context.Context, in *pb.UpdateInstrumentRequest, opts ...grpc.CallOption) (*pb.UpdateInstrumentResponse, error) {
	if f.updateInstrumentFn != nil {
		return f.updateInstrumentFn(ctx, in)
	}
	return nil, nil
}

func (f *fakePortfolioClient) DeleteInstrument(ctx context.Context, in *pb.DeleteInstrumentRequest, opts ...grpc.CallOption) (*pb.DeleteInstrumentResponse, error) {
	if f.deleteInstrumentFn != nil {
		return f.deleteInstrumentFn(ctx, in)
	}
	return nil, nil
}

func (f *fakePortfolioClient) ListExchanges(ctx context.Context, in *pb.ListExchangesRequest, opts ...grpc.CallOption) (*pb.ListExchangesResponse, error) {
	if f.listExchangesFn != nil {
		return f.listExchangesFn(ctx, in)
	}
	return nil, nil
}

func (f *fakePortfolioClient) ListInstrumentPrices(ctx context.Context, in *pb.ListInstrumentPricesRequest, opts ...grpc.CallOption) (*pb.ListInstrumentPricesResponse, error) {
	if f.listInstrumentPricesFn != nil {
		return f.listInstrumentPricesFn(ctx, in)
	}
	return nil, nil
}

func (f *fakePortfolioClient) RecordPriceOverride(ctx context.Context, in *pb.RecordPriceOverrideRequest, opts ...grpc.CallOption) (*pb.RecordPriceOverrideResponse, error) {
	if f.recordPriceOverrideFn != nil {
		return f.recordPriceOverrideFn(ctx, in)
	}
	return nil, nil
}

func (f *fakePortfolioClient) GetIngestionStatus(ctx context.Context, in *pb.GetIngestionStatusRequest, opts ...grpc.CallOption) (*pb.GetIngestionStatusResponse, error) {
	if f.getIngestionStatusFn != nil {
		return f.getIngestionStatusFn(ctx, in)
	}
	return nil, nil
}

func (f *fakePortfolioClient) TriggerMarketSync(ctx context.Context, in *pb.TriggerMarketSyncRequest, opts ...grpc.CallOption) (*pb.TriggerMarketSyncResponse, error) {
	if f.triggerMarketSyncFn != nil {
		return f.triggerMarketSyncFn(ctx, in)
	}
	return nil, nil
}

func (f *fakePortfolioClient) TriggerBackfill(ctx context.Context, in *pb.TriggerBackfillRequest, opts ...grpc.CallOption) (*pb.TriggerBackfillResponse, error) {
	if f.triggerBackfillFn != nil {
		return f.triggerBackfillFn(ctx, in)
	}
	return nil, nil
}

func (f *fakePortfolioClient) ListCurrencyPairs(ctx context.Context, in *pb.ListCurrencyPairsRequest, opts ...grpc.CallOption) (*pb.ListCurrencyPairsResponse, error) {
	if f.listCurrencyPairsFn != nil {
		return f.listCurrencyPairsFn(ctx, in)
	}
	return nil, nil
}

func (f *fakePortfolioClient) ListFXRates(ctx context.Context, in *pb.ListFXRatesRequest, opts ...grpc.CallOption) (*pb.ListFXRatesResponse, error) {
	if f.listFXRatesFn != nil {
		return f.listFXRatesFn(ctx, in)
	}
	return nil, nil
}

func (f *fakePortfolioClient) GetCurrencyPairHistory(ctx context.Context, in *pb.GetCurrencyPairHistoryRequest, opts ...grpc.CallOption) (*pb.GetCurrencyPairHistoryResponse, error) {
	if f.getCurrencyPairHistoryFn != nil {
		return f.getCurrencyPairHistoryFn(ctx, in)
	}
	return nil, nil
}

func (f *fakePortfolioClient) RecordFXRateOverride(ctx context.Context, in *pb.RecordFXRateOverrideRequest, opts ...grpc.CallOption) (*pb.RecordFXRateOverrideResponse, error) {
	if f.recordFXRateOverrideFn != nil {
		return f.recordFXRateOverrideFn(ctx, in)
	}
	return nil, nil
}

func (f *fakePortfolioClient) CheckTransactionDuplicates(ctx context.Context, in *pb.CheckTransactionDuplicatesRequest, opts ...grpc.CallOption) (*pb.CheckTransactionDuplicatesResponse, error) {
	if f.checkTransactionDuplicatesFn != nil {
		return f.checkTransactionDuplicatesFn(ctx, in)
	}
	return nil, nil
}

func (f *fakePortfolioClient) BatchImportTransactions(ctx context.Context, in *pb.BatchImportTransactionsRequest, opts ...grpc.CallOption) (*pb.BatchImportTransactionsResponse, error) {
	if f.batchImportTransactionsFn != nil {
		return f.batchImportTransactionsFn(ctx, in)
	}
	return nil, nil
}

func (f *fakePortfolioClient) GetCashFlowReport(ctx context.Context, in *pb.GetCashFlowReportRequest, opts ...grpc.CallOption) (*pb.GetCashFlowReportResponse, error) {
	if f.getCashFlowReportFn != nil {
		return f.getCashFlowReportFn(ctx, in)
	}
	return nil, nil
}

func TestQueryResolver_Portfolio(t *testing.T) {
	ctx := context.Background()

	fakeClient := &fakePortfolioClient{
		getPortfolioFn: func(ctx context.Context, in *pb.GetPortfolioRequest) (*pb.GetPortfolioResponse, error) {
			if in.GetUserId() != "1" && in.GetUserId() != defaultUserID {
				t.Errorf("expected user_id 1 or defaultUserID, got %s", in.GetUserId())
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

func TestQueryResolver_Portfolio_AverageBuyPrice(t *testing.T) {
	ctx := context.Background()

	fakeClient := &fakePortfolioClient{
		getPortfolioFn: func(ctx context.Context, in *pb.GetPortfolioRequest) (*pb.GetPortfolioResponse, error) {
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
					Investments: []*pb.Investment{
						{
							Id:     "inv-1",
							Ticker: "AAPL",
							Name:   "Apple Inc.",
							Price: &commonpb.Money{
								Amount:       &commonpb.Decimal{Value: "185.92"},
								CurrencyCode: "USD",
							},
							AverageBuyPrice: &commonpb.Money{
								Amount:       &commonpb.Decimal{Value: "156.24"},
								CurrencyCode: "USD",
							},
							Quantity: &commonpb.Decimal{Value: "100"},
						},
						{
							Id:     "inv-2",
							Ticker: "GIFT",
							Name:   "Free Gift",
							Price: &commonpb.Money{
								Amount:       &commonpb.Decimal{Value: "10.00"},
								CurrencyCode: "USD",
							},
							AverageBuyPrice: nil,
							Quantity:        &commonpb.Decimal{Value: "10"},
						},
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
	if len(result.Investments) != 2 {
		t.Fatalf("expected 2 investments, got %d", len(result.Investments))
	}

	inv1 := result.Investments[0]
	if inv1.AverageBuyPrice == nil {
		t.Fatalf("expected non-nil averageBuyPrice for inv1")
	}
	if inv1.AverageBuyPrice.Amount.String() != "156.24" {
		t.Errorf("expected 156.24, got %s", inv1.AverageBuyPrice.Amount.String())
	}
	if inv1.AverageBuyPrice.CurrencyCode != "USD" {
		t.Errorf("expected currency USD, got %s", inv1.AverageBuyPrice.CurrencyCode)
	}

	inv2 := result.Investments[1]
	if inv2.AverageBuyPrice != nil {
		t.Errorf("expected nil averageBuyPrice for inv2, got %v", inv2.AverageBuyPrice)
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

func TestMutationResolver_AddTransaction_FeeCurrencyCode(t *testing.T) {
	ctx := context.Background()

	var receivedFee *commonpb.Money
	fakeClient := &fakePortfolioClient{
		addTransactionFn: func(ctx context.Context, in *pb.AddTransactionRequest) (*pb.AddTransactionResponse, error) {
			receivedFee = in.GetFee()
			return &pb.AddTransactionResponse{
				TransactionId: "tx-fee-123",
				Portfolio: &pb.Portfolio{
					TotalValue: &commonpb.Money{
						Amount:       &commonpb.Decimal{Value: "1000.00"},
						CurrencyCode: "USD",
					},
					CashBalance: &commonpb.Money{
						Amount:       &commonpb.Decimal{Value: "500.00"},
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
	currencyUSD := "USD"
	feeCurrencyAUD := "AUD"

	payload, err := mResolver.AddTransaction(ctx, model.AddTransactionInput{
		Type:            model.TransactionTypeBuy,
		Symbol:          &sym,
		TradeDate:       "2025-01-05",
		Quantity:        &qty,
		Price:           &price,
		CurrencyCode:    &currencyUSD,
		Fee:             &fee,
		FeeCurrencyCode: &feeCurrencyAUD,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload == nil {
		t.Fatalf("expected payload, got nil")
	}
	if receivedFee == nil {
		t.Fatalf("expected receivedFee, got nil")
	}
	if receivedFee.GetCurrencyCode() != "AUD" {
		t.Errorf("expected fee currency AUD, got %s", receivedFee.GetCurrencyCode())
	}
	if receivedFee.GetAmount().GetValue() != "5.00" && receivedFee.GetAmount().GetValue() != "5" {
		t.Errorf("expected fee amount 5, got %s", receivedFee.GetAmount().GetValue())
	}
}

func TestMutationResolver_AddTransaction_Split(t *testing.T) {
	ctx := context.Background()

	fakeClient := &fakePortfolioClient{
		addTransactionFn: func(ctx context.Context, in *pb.AddTransactionRequest) (*pb.AddTransactionResponse, error) {
			if in.GetType() != pb.TransactionType_TRANSACTION_TYPE_SPLIT {
				t.Errorf("expected SPLIT transaction, got %v", in.GetType())
			}
			if in.GetSymbol() != "AAPL" {
				t.Errorf("expected AAPL, got %s", in.GetSymbol())
			}
			if in.GetQuantity().GetValue() != "2" {
				t.Errorf("expected quantity 2, got %s", in.GetQuantity().GetValue())
			}
			return &pb.AddTransactionResponse{
				TransactionId: "tx-split-123",
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
	qty := model.Decimal(decimal.NewFromInt(2))

	payload, err := mResolver.AddTransaction(ctx, model.AddTransactionInput{
		Type:      model.TransactionTypeSplit,
		Symbol:    &sym,
		TradeDate: "2025-06-01",
		Quantity:  &qty,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload == nil {
		t.Fatalf("expected payload, got nil")
	}
	if payload.TransactionID != "tx-split-123" {
		t.Errorf("expected tx ID tx-split-123, got %s", payload.TransactionID)
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

func TestQueryResolver_AllInstruments(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully maps all instruments", func(t *testing.T) {
		isin := "US0378331005"
		fakeClient := &fakePortfolioClient{
			listAllInstrumentsFn: func(ctx context.Context, in *pb.ListAllInstrumentsRequest) (*pb.ListAllInstrumentsResponse, error) {
				return &pb.ListAllInstrumentsResponse{
					Instruments: []*pb.Instrument{
						{
							Id:           "inst-1",
							Symbol:       "AAPL",
							Name:         "Apple Inc.",
							CurrencyCode: "USD",
							AssetClass:   "EQUITY",
							ExchangeCode: "XNAS",
							Isin:         isin,
							IsActive:     true,
						},
					},
				}, nil
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		qResolver := resolver.Query()

		isActive := true
		search := "AAPL"
		insts, err := qResolver.AllInstruments(ctx, &isActive, &search)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(insts) != 1 {
			t.Fatalf("expected 1 instrument, got %d", len(insts))
		}
		if insts[0].Symbol != "AAPL" || insts[0].ExchangeCode != "XNAS" || !insts[0].IsActive {
			t.Errorf("unexpected instrument: %+v", insts[0])
		}
	})
}

func TestMutationResolver_CreateInstrument(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully creates instrument", func(t *testing.T) {
		fakeClient := &fakePortfolioClient{
			createInstrumentFn: func(ctx context.Context, in *pb.CreateInstrumentRequest) (*pb.CreateInstrumentResponse, error) {
				return &pb.CreateInstrumentResponse{
					Instrument: &pb.Instrument{
						Id:           "inst-nvda",
						Symbol:       in.Symbol,
						ExchangeCode: in.ExchangeCode,
						Name:         in.Name,
						AssetClass:   in.AssetClass,
						CurrencyCode: in.CurrencyCode,
						IsActive:     true,
					},
				}, nil
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		mResolver := resolver.Mutation()

		inst, err := mResolver.CreateInstrument(ctx, model.CreateInstrumentInput{
			Symbol:       "NVDA",
			ExchangeCode: "XNAS",
			Name:         "NVIDIA Corp.",
			AssetClass:   "EQUITY",
			CurrencyCode: "USD",
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if inst.Symbol != "NVDA" || inst.ExchangeCode != "XNAS" {
			t.Errorf("unexpected instrument: %+v", inst)
		}
	})
}

func TestMutationResolver_UpdateInstrument(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully updates instrument status", func(t *testing.T) {
		fakeClient := &fakePortfolioClient{
			updateInstrumentFn: func(ctx context.Context, in *pb.UpdateInstrumentRequest) (*pb.UpdateInstrumentResponse, error) {
				return &pb.UpdateInstrumentResponse{
					Instrument: &pb.Instrument{
						Id:           in.Id,
						Symbol:       "AAPL",
						ExchangeCode: "XNAS",
						Name:         "Apple Inc.",
						AssetClass:   "EQUITY",
						CurrencyCode: "USD",
						IsActive:     *in.IsActive,
					},
				}, nil
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		mResolver := resolver.Mutation()

		isActive := false
		inst, err := mResolver.UpdateInstrument(ctx, model.UpdateInstrumentInput{
			ID:       "inst-1",
			IsActive: &isActive,
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if inst.IsActive != false {
			t.Errorf("expected IsActive false, got %v", inst.IsActive)
		}
	})
}

func TestMutationResolver_DeleteInstrument(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully deletes instrument", func(t *testing.T) {
		fakeClient := &fakePortfolioClient{
			deleteInstrumentFn: func(ctx context.Context, in *pb.DeleteInstrumentRequest) (*pb.DeleteInstrumentResponse, error) {
				if in.Id != "inst-1" {
					t.Errorf("expected id inst-1, got %s", in.Id)
				}
				return &pb.DeleteInstrumentResponse{
					Success: true,
					Id:      in.Id,
				}, nil
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		mResolver := resolver.Mutation()

		payload, err := mResolver.DeleteInstrument(ctx, "inst-1")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !payload.Success {
			t.Errorf("expected Success true, got false")
		}
		if payload.ID != "inst-1" {
			t.Errorf("expected ID inst-1, got %s", payload.ID)
		}
	})

	t.Run("propagates gRPC error", func(t *testing.T) {
		fakeClient := &fakePortfolioClient{
			deleteInstrumentFn: func(ctx context.Context, in *pb.DeleteInstrumentRequest) (*pb.DeleteInstrumentResponse, error) {
				return nil, errors.New("cannot delete instrument: in use")
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		mResolver := resolver.Mutation()

		_, err := mResolver.DeleteInstrument(ctx, "inst-1")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "cannot delete instrument: in use" {
			t.Errorf("expected error message to match, got %v", err)
		}
	})
}

func TestQueryResolver_InstrumentPrices(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully queries instrument prices", func(t *testing.T) {
		fakeClient := &fakePortfolioClient{
			listInstrumentPricesFn: func(ctx context.Context, in *pb.ListInstrumentPricesRequest) (*pb.ListInstrumentPricesResponse, error) {
				return &pb.ListInstrumentPricesResponse{
					Prices: []*pb.InstrumentPriceItem{
						{
							InstrumentId: "inst-1",
							Symbol:       "AAPL",
							PriceDate:    "2026-10-02",
							Price: &commonpb.Money{
								Amount:       &commonpb.Decimal{Value: "228.45"},
								CurrencyCode: "USD",
							},
							Source:    "manual",
							UpdatedAt: "2026-10-02T21:05:00Z",
						},
					},
					TotalCount: 1,
				}, nil
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		qResolver := resolver.Query()

		sym := "AAPL"
		res, err := qResolver.InstrumentPrices(ctx, &sym, nil, nil, nil, nil)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if res.TotalCount != 1 || len(res.Items) != 1 {
			t.Fatalf("expected 1 price, got %d", len(res.Items))
		}
		if res.Items[0].Price.Amount.String() != "228.45" {
			t.Errorf("expected price 228.45, got %s", res.Items[0].Price.Amount.String())
		}
	})

	t.Run("successfully queries instrument prices with date range and pagination", func(t *testing.T) {
		fakeClient := &fakePortfolioClient{
			listInstrumentPricesFn: func(ctx context.Context, in *pb.ListInstrumentPricesRequest) (*pb.ListInstrumentPricesResponse, error) {
				if in.Symbol == nil || *in.Symbol != "AAPL" {
					t.Errorf("expected symbol AAPL, got %v", in.Symbol)
				}
				if in.FromDate == nil || *in.FromDate != "2026-09-01" {
					t.Errorf("expected fromDate 2026-09-01, got %v", in.FromDate)
				}
				if in.ToDate == nil || *in.ToDate != "2026-10-01" {
					t.Errorf("expected toDate 2026-10-01, got %v", in.ToDate)
				}
				if in.Limit != 25 {
					t.Errorf("expected limit 25, got %d", in.Limit)
				}
				if in.Offset != 50 {
					t.Errorf("expected offset 50, got %d", in.Offset)
				}
				return &pb.ListInstrumentPricesResponse{
					Prices: []*pb.InstrumentPriceItem{
						{
							InstrumentId: "inst-1",
							Symbol:       "AAPL",
							PriceDate:    "2026-10-01",
							Price: &commonpb.Money{
								Amount:       &commonpb.Decimal{Value: "230.25"},
								CurrencyCode: "USD",
							},
							Source:    "twelve_data",
							UpdatedAt: "2026-10-01T21:00:00Z",
						},
					},
					TotalCount: 88,
				}, nil
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		qResolver := resolver.Query()

		sym := "AAPL"
		from := "2026-09-01"
		to := "2026-10-01"
		limit := 25
		offset := 50
		res, err := qResolver.InstrumentPrices(ctx, &sym, &from, &to, &limit, &offset)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if res.TotalCount != 88 || len(res.Items) != 1 {
			t.Fatalf("expected total count 88 and 1 price, got total %d, items %d", res.TotalCount, len(res.Items))
		}
		if res.Items[0].Price.Amount.String() != "230.25" {
			t.Errorf("expected price 230.25, got %s", res.Items[0].Price.Amount.String())
		}
	})
}

func TestMutationResolver_RecordPriceOverride(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully records price override", func(t *testing.T) {
		fakeClient := &fakePortfolioClient{
			recordPriceOverrideFn: func(ctx context.Context, in *pb.RecordPriceOverrideRequest) (*pb.RecordPriceOverrideResponse, error) {
				return &pb.RecordPriceOverrideResponse{
					Price: &pb.InstrumentPriceItem{
						InstrumentId: "inst-1",
						Symbol:       in.Symbol,
						PriceDate:    in.PriceDate,
						Price: &commonpb.Money{
							Amount:       in.Price,
							CurrencyCode: "USD",
						},
						Source: "manual: audit note",
					},
					ValuationsRecomputed: true,
				}, nil
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		mResolver := resolver.Mutation()

		reason := "audit note"
		recompute := true
		payload, err := mResolver.RecordPriceOverride(ctx, model.RecordPriceOverrideInput{
			Symbol:              "AAPL",
			PriceDate:           "2026-10-02",
			Price:               model.Decimal(decimal.RequireFromString("228.45")),
			Reason:              &reason,
			RecomputeValuations: &recompute,
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !payload.ValuationsRecomputed {
			t.Errorf("expected ValuationsRecomputed true")
		}
		if payload.Price.Symbol != "AAPL" || payload.Price.Price.Amount.String() != "228.45" {
			t.Errorf("unexpected price payload: %+v", payload.Price)
		}
	})
}

func TestQueryResolver_IngestionStatus(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully queries ingestion status", func(t *testing.T) {
		fakeClient := &fakePortfolioClient{
			getIngestionStatusFn: func(ctx context.Context, in *pb.GetIngestionStatusRequest) (*pb.GetIngestionStatusResponse, error) {
				return &pb.GetIngestionStatusResponse{
					Feeds: []*pb.FeedHealthStatus{
						{
							Name:     "EOD Equity Feeds",
							Status:   "ACTIVE",
							Provider: "Twelve Data",
							Schedule: "Daily at 21:00 UTC",
							LastRun:  "2026-10-02 21:05 UTC",
							Details:  "Healthy",
						},
					},
					TrackedInstruments:  5,
					TrackedCurrencies:   4,
					RateLimitRemaining:  800,
					RateLimitBudget:     800,
					PendingBackfillJobs: 0,
				}, nil
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		qResolver := resolver.Query()

		status, err := qResolver.IngestionStatus(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if status.TrackedInstruments != 5 || len(status.Feeds) != 1 {
			t.Errorf("unexpected ingestion status: %+v", status)
		}
	})
}

func TestMutationResolver_TriggerMarketSync(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully triggers market sync", func(t *testing.T) {
		fakeClient := &fakePortfolioClient{
			triggerMarketSyncFn: func(ctx context.Context, in *pb.TriggerMarketSyncRequest) (*pb.TriggerMarketSyncResponse, error) {
				return &pb.TriggerMarketSyncResponse{
					Success:       true,
					PricesSynced:  2,
					FxRatesSynced: 7,
					Message:       "Sync completed",
				}, nil
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		mResolver := resolver.Mutation()

		syncFx := true
		payload, err := mResolver.TriggerMarketSync(ctx, []string{"AAPL", "MSFT"}, &syncFx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !payload.Success || payload.PricesSynced != 2 || payload.FxRatesSynced != 7 {
			t.Errorf("unexpected sync payload: %+v", payload)
		}
	})
}

func TestMutationResolver_TriggerBackfill(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully triggers backfill with all options", func(t *testing.T) {
		backfillAssets := true
		backfillFx := true
		recompute := true

		fakeClient := &fakePortfolioClient{
			triggerBackfillFn: func(ctx context.Context, in *pb.TriggerBackfillRequest) (*pb.TriggerBackfillResponse, error) {
				if in.FromDate != "2025-01-01" || in.ToDate != "2025-01-31" {
					t.Errorf("unexpected date range: %s to %s", in.FromDate, in.ToDate)
				}
				if len(in.Symbols) != 1 || in.Symbols[0] != "AAPL" {
					t.Errorf("unexpected symbols: %v", in.Symbols)
				}
				if len(in.CurrencyPairs) != 1 || in.CurrencyPairs[0] != "EUR/USD" {
					t.Errorf("unexpected currency pairs: %v", in.CurrencyPairs)
				}
				if !in.BackfillAssets || !in.BackfillFx || !in.RecomputeValuations {
					t.Errorf("unexpected flags: assets=%v, fx=%v, recompute=%v", in.BackfillAssets, in.BackfillFx, in.RecomputeValuations)
				}

				return &pb.TriggerBackfillResponse{
					Success:       true,
					PricesSynced:  42,
					FxRatesSynced: 14,
					Message:       "Backfill completed successfully",
					Warnings:      []string{"Minor warning"},
				}, nil
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		mResolver := resolver.Mutation()

		payload, err := mResolver.TriggerBackfill(ctx, model.TriggerBackfillInput{
			FromDate:            "2025-01-01",
			ToDate:              "2025-01-31",
			Symbols:             []string{"AAPL"},
			CurrencyPairs:       []string{"EUR/USD"},
			BackfillAssets:      &backfillAssets,
			BackfillFx:          &backfillFx,
			RecomputeValuations: &recompute,
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !payload.Success {
			t.Errorf("expected success to be true")
		}
		if payload.PricesSynced != 42 {
			t.Errorf("expected 42 prices synced, got %d", payload.PricesSynced)
		}
		if payload.FxRatesSynced != 14 {
			t.Errorf("expected 14 fx rates synced, got %d", payload.FxRatesSynced)
		}
		if payload.Message != "Backfill completed successfully" {
			t.Errorf("unexpected message: %s", payload.Message)
		}
		if len(payload.Warnings) != 1 || payload.Warnings[0] != "Minor warning" {
			t.Errorf("unexpected warnings: %v", payload.Warnings)
		}
	})

	t.Run("defaults boolean flags when nil", func(t *testing.T) {
		fakeClient := &fakePortfolioClient{
			triggerBackfillFn: func(ctx context.Context, in *pb.TriggerBackfillRequest) (*pb.TriggerBackfillResponse, error) {
				if !in.BackfillAssets {
					t.Errorf("expected backfillAssets to default to true")
				}
				if !in.BackfillFx {
					t.Errorf("expected backfillFx to default to true")
				}
				if in.RecomputeValuations {
					t.Errorf("expected recomputeValuations to default to false")
				}
				return &pb.TriggerBackfillResponse{
					Success:      true,
					PricesSynced: 10,
				}, nil
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		mResolver := resolver.Mutation()

		payload, err := mResolver.TriggerBackfill(ctx, model.TriggerBackfillInput{
			FromDate: "2025-01-01",
			ToDate:   "2025-01-10",
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !payload.Success {
			t.Errorf("expected success to be true")
		}
	})

	t.Run("returns error when gRPC call fails", func(t *testing.T) {
		fakeClient := &fakePortfolioClient{
			triggerBackfillFn: func(ctx context.Context, in *pb.TriggerBackfillRequest) (*pb.TriggerBackfillResponse, error) {
				return nil, errors.New("gRPC connection failure")
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		mResolver := resolver.Mutation()

		_, err := mResolver.TriggerBackfill(ctx, model.TriggerBackfillInput{
			FromDate: "2025-01-01",
			ToDate:   "2025-01-10",
		})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})
}

func TestQueryResolver_UserPreferences(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully retrieves user preferences", func(t *testing.T) {
		fakeUser := &fakeUserClient{
			getUserPreferencesFn: func(ctx context.Context, in *userpb.GetUserPreferencesRequest) (*userpb.GetUserPreferencesResponse, error) {
				return &userpb.GetUserPreferencesResponse{
					Preferences: &userpb.UserPreferences{
						UserId:          "018f0000-0000-7000-8000-000000000001",
						Email:           "oscar@example.com",
						DisplayName:     "Oscar Garcia",
						DisplayCurrency: "USD",
						Theme:           "DARK",
						CreatedAt:       "2026-01-01T00:00:00Z",
						UpdatedAt:       "2026-01-01T00:00:00Z",
					},
				}, nil
			},
		}

		resolver := &Resolver{UserClient: fakeUser}
		qResolver := resolver.Query()

		prefs, err := qResolver.UserPreferences(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if prefs.UserID != "018f0000-0000-7000-8000-000000000001" || prefs.DisplayName != "Oscar Garcia" || prefs.DisplayCurrency != "USD" {
			t.Errorf("unexpected preferences: %+v", prefs)
		}
	})

	t.Run("handles error from user service", func(t *testing.T) {
		fakeUser := &fakeUserClient{
			getUserPreferencesFn: func(ctx context.Context, in *userpb.GetUserPreferencesRequest) (*userpb.GetUserPreferencesResponse, error) {
				return nil, errors.New("user service unavailable")
			},
		}

		resolver := &Resolver{UserClient: fakeUser}
		qResolver := resolver.Query()

		_, err := qResolver.UserPreferences(ctx)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})
}

func TestQueryResolver_SupportedCurrencies(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully retrieves supported currencies", func(t *testing.T) {
		fakeUser := &fakeUserClient{
			listSupportedCurrenciesFn: func(ctx context.Context, in *userpb.ListSupportedCurrenciesRequest) (*userpb.ListSupportedCurrenciesResponse, error) {
				return &userpb.ListSupportedCurrenciesResponse{
					Currencies: []*userpb.CurrencyInfo{
						{Code: "USD", Name: "US Dollar", Symbol: "$"},
						{Code: "EUR", Name: "Euro", Symbol: "€"},
					},
				}, nil
			},
		}

		resolver := &Resolver{UserClient: fakeUser}
		qResolver := resolver.Query()

		currencies, err := qResolver.SupportedCurrencies(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(currencies) != 2 || currencies[0].Code != "USD" || currencies[1].Code != "EUR" {
			t.Errorf("unexpected currencies: %+v", currencies)
		}
	})

	t.Run("handles error from user service", func(t *testing.T) {
		fakeUser := &fakeUserClient{
			listSupportedCurrenciesFn: func(ctx context.Context, in *userpb.ListSupportedCurrenciesRequest) (*userpb.ListSupportedCurrenciesResponse, error) {
				return nil, errors.New("currency fetch error")
			},
		}

		resolver := &Resolver{UserClient: fakeUser}
		qResolver := resolver.Query()

		_, err := qResolver.SupportedCurrencies(ctx)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})
}

func TestMutationResolver_UpdateUserPreferences(t *testing.T) {
	ctx := context.Background()
	newName := "Oscar Garcia Updated"
	newCurr := "EUR"
	newTheme := "LIGHT"

	t.Run("successfully updates preferences and attaches portfolio", func(t *testing.T) {
		fakeUser := &fakeUserClient{
			updateUserPreferencesFn: func(ctx context.Context, in *userpb.UpdateUserPreferencesRequest) (*userpb.UpdateUserPreferencesResponse, error) {
				return &userpb.UpdateUserPreferencesResponse{
					Preferences: &userpb.UserPreferences{
						UserId:          in.UserId,
						Email:           "oscar@example.com",
						DisplayName:     *in.DisplayName,
						DisplayCurrency: *in.DisplayCurrency,
						Theme:           *in.Theme,
						CreatedAt:       "2026-01-01T00:00:00Z",
						UpdatedAt:       "2026-01-02T00:00:00Z",
					},
				}, nil
			},
		}
		fakePort := &fakePortfolioClient{
			updatePortfolioBaseCurrencyFn: func(ctx context.Context, in *pb.UpdatePortfolioBaseCurrencyRequest) (*pb.UpdatePortfolioBaseCurrencyResponse, error) {
				return &pb.UpdatePortfolioBaseCurrencyResponse{
					Portfolio: &pb.Portfolio{
						TotalValue: &commonpb.Money{
							Amount:       &commonpb.Decimal{Value: "100000.00"},
							CurrencyCode: in.BaseCurrency,
						},
					},
				}, nil
			},
			getPortfolioFn: func(ctx context.Context, in *pb.GetPortfolioRequest) (*pb.GetPortfolioResponse, error) {
				return &pb.GetPortfolioResponse{
					Portfolio: &pb.Portfolio{
						TotalValue: &commonpb.Money{
							Amount:       &commonpb.Decimal{Value: "100000.00"},
							CurrencyCode: "EUR",
						},
					},
				}, nil
			},
		}

		resolver := &Resolver{
			UserClient:      fakeUser,
			PortfolioClient: fakePort,
		}
		mResolver := resolver.Mutation()

		payload, err := mResolver.UpdateUserPreferences(ctx, model.UpdateUserPreferencesInput{
			DisplayName:     &newName,
			DisplayCurrency: &newCurr,
			Theme:           &newTheme,
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if payload.Preferences.DisplayName != newName || payload.Preferences.DisplayCurrency != newCurr || payload.Preferences.Theme != newTheme {
			t.Errorf("unexpected preferences: %+v", payload.Preferences)
		}
		if payload.Portfolio == nil || payload.Portfolio.TotalValue.CurrencyCode != "EUR" {
			t.Errorf("unexpected portfolio: %+v", payload.Portfolio)
		}
	})

	t.Run("handles error from user service", func(t *testing.T) {
		fakeUser := &fakeUserClient{
			updateUserPreferencesFn: func(ctx context.Context, in *userpb.UpdateUserPreferencesRequest) (*userpb.UpdateUserPreferencesResponse, error) {
				return nil, errors.New("invalid currency code")
			},
		}

		resolver := &Resolver{UserClient: fakeUser}
		mResolver := resolver.Mutation()

		_, err := mResolver.UpdateUserPreferences(ctx, model.UpdateUserPreferencesInput{
			DisplayCurrency: &newCurr,
		})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})
}

func TestMutationResolver_RecordFXRateOverride(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully records fx rate override", func(t *testing.T) {
		fakeClient := &fakePortfolioClient{
			recordFXRateOverrideFn: func(ctx context.Context, in *pb.RecordFXRateOverrideRequest) (*pb.RecordFXRateOverrideResponse, error) {
				if in.BaseCurrency != "EUR" || in.QuoteCurrency != "USD" {
					t.Errorf("unexpected currency pair: %s/%s", in.BaseCurrency, in.QuoteCurrency)
				}
				if in.RateDate != "2026-10-05" {
					t.Errorf("unexpected rate date: %s", in.RateDate)
				}
				if !in.RecomputeValuations {
					t.Errorf("expected recompute valuations to be true")
				}
				return &pb.RecordFXRateOverrideResponse{
					Rate: &pb.FXRateItem{
						BaseCurrency:  in.BaseCurrency,
						QuoteCurrency: in.QuoteCurrency,
						RateDate:      in.RateDate,
						Rate:          in.Rate,
						InvertedRate:  &commonpb.Decimal{Value: "0.92165899"},
						Source:        "manual: audit note",
					},
					ValuationsRecomputed: true,
				}, nil
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		mResolver := resolver.Mutation()

		reason := "audit note"
		recompute := true
		payload, err := mResolver.RecordFXRateOverride(ctx, model.RecordFXRateOverrideInput{
			BaseCurrency:        "EUR",
			QuoteCurrency:       "USD",
			RateDate:            "2026-10-05",
			Rate:                model.Decimal(decimal.RequireFromString("1.0850")),
			Reason:              &reason,
			RecomputeValuations: &recompute,
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !payload.ValuationsRecomputed {
			t.Errorf("expected ValuationsRecomputed true")
		}
		if payload.Rate.Pair != "EUR/USD" || payload.Rate.Rate.String() != "1.085" {
			t.Errorf("unexpected fx rate payload: %+v", payload.Rate)
		}
		if payload.Rate.InvertedRate.String() != "0.92165899" {
			t.Errorf("unexpected inverted rate: %s", payload.Rate.InvertedRate.String())
		}
	})

	t.Run("handles error from portfolio service", func(t *testing.T) {
		fakeClient := &fakePortfolioClient{
			recordFXRateOverrideFn: func(ctx context.Context, in *pb.RecordFXRateOverrideRequest) (*pb.RecordFXRateOverrideResponse, error) {
				return nil, errors.New("rate must be positive")
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		mResolver := resolver.Mutation()

		_, err := mResolver.RecordFXRateOverride(ctx, model.RecordFXRateOverrideInput{
			BaseCurrency:  "EUR",
			QuoteCurrency: "USD",
			RateDate:      "2026-10-05",
			Rate:          model.Decimal(decimal.RequireFromString("-1.0")),
		})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})
}

func TestQueryResolver_CurrencyPairs(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully queries currency pairs", func(t *testing.T) {
		fakeClient := &fakePortfolioClient{
			listCurrencyPairsFn: func(ctx context.Context, in *pb.ListCurrencyPairsRequest) (*pb.ListCurrencyPairsResponse, error) {
				return &pb.ListCurrencyPairsResponse{
					Pairs: []*pb.CurrencyPairItem{
						{
							BaseCurrency:    "EUR",
							QuoteCurrency:   "USD",
							LatestRate:      &commonpb.Decimal{Value: "1.0850"},
							LatestDate:      "2026-10-05",
							LatestSource:    "ECB",
							PreviousRate:    &commonpb.Decimal{Value: "1.0800"},
							Change_1DAmount: &commonpb.Decimal{Value: "0.0050"},
							Change_1DPct:    &commonpb.Decimal{Value: "0.4630"},
							TotalRecords:    250,
							FirstDate:       "2025-10-01",
							LastDate:        "2026-10-05",
						},
					},
				}, nil
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		qResolver := resolver.Query()

		pairs, err := qResolver.CurrencyPairs(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(pairs) != 1 {
			t.Fatalf("expected 1 pair, got %d", len(pairs))
		}
		p := pairs[0]
		if p.Pair != "EUR/USD" || p.BaseCurrency != "EUR" || p.QuoteCurrency != "USD" {
			t.Errorf("unexpected pair identifier: %s", p.Pair)
		}
		if p.LatestRate.String() != "1.085" {
			t.Errorf("unexpected latest rate: %s", p.LatestRate.String())
		}
		if p.Change1dAmount == nil || p.Change1dAmount.String() != "0.005" {
			t.Errorf("unexpected change 1d amount: %v", p.Change1dAmount)
		}
		if p.TotalRecords != 250 {
			t.Errorf("unexpected total records: %d", p.TotalRecords)
		}
	})

	t.Run("handles error from portfolio service", func(t *testing.T) {
		fakeClient := &fakePortfolioClient{
			listCurrencyPairsFn: func(ctx context.Context, in *pb.ListCurrencyPairsRequest) (*pb.ListCurrencyPairsResponse, error) {
				return nil, errors.New("db connection failure")
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		qResolver := resolver.Query()

		_, err := qResolver.CurrencyPairs(ctx)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})
}

func TestQueryResolver_CurrencyPairHistory(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully queries currency pair history", func(t *testing.T) {
		fakeClient := &fakePortfolioClient{
			getCurrencyPairHistoryFn: func(ctx context.Context, in *pb.GetCurrencyPairHistoryRequest) (*pb.GetCurrencyPairHistoryResponse, error) {
				if in.BaseCurrency != "EUR" || in.QuoteCurrency != "USD" {
					t.Errorf("unexpected currencies: %s/%s", in.BaseCurrency, in.QuoteCurrency)
				}
				if in.Timeframe != pb.HistoryTimeframe_HISTORY_TIMEFRAME_1M {
					t.Errorf("unexpected timeframe: %v", in.Timeframe)
				}
				return &pb.GetCurrencyPairHistoryResponse{
					BaseCurrency:  in.BaseCurrency,
					QuoteCurrency: in.QuoteCurrency,
					Points: []*pb.FXHistoryPoint{
						{
							Date:         "2026-09-05",
							Rate:         &commonpb.Decimal{Value: "1.0700"},
							InvertedRate: &commonpb.Decimal{Value: "0.934579"},
							Source:       "ECB",
						},
						{
							Date:         "2026-10-05",
							Rate:         &commonpb.Decimal{Value: "1.0850"},
							InvertedRate: &commonpb.Decimal{Value: "0.921659"},
							Source:       "ECB",
						},
					},
					StartRate:       &commonpb.Decimal{Value: "1.0700"},
					EndRate:         &commonpb.Decimal{Value: "1.0850"},
					PeriodChange:    &commonpb.Decimal{Value: "0.0150"},
					PeriodChangePct: &commonpb.Decimal{Value: "1.4019"},
					PeriodHigh:      &commonpb.Decimal{Value: "1.0850"},
					PeriodLow:       &commonpb.Decimal{Value: "1.0700"},
				}, nil
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		qResolver := resolver.Query()

		hist, err := qResolver.CurrencyPairHistory(ctx, "EUR", "USD", model.HistoryTimeframeTimeframe1m)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if hist.Pair != "EUR/USD" {
			t.Errorf("expected pair EUR/USD, got %s", hist.Pair)
		}
		if len(hist.Points) != 2 {
			t.Fatalf("expected 2 points, got %d", len(hist.Points))
		}
		if hist.StartRate.String() != "1.07" || hist.EndRate.String() != "1.085" {
			t.Errorf("unexpected start/end rates: %s / %s", hist.StartRate.String(), hist.EndRate.String())
		}
		if hist.PeriodChange.String() != "0.015" {
			t.Errorf("unexpected period change: %s", hist.PeriodChange.String())
		}
	})

	t.Run("handles error from portfolio service", func(t *testing.T) {
		fakeClient := &fakePortfolioClient{
			getCurrencyPairHistoryFn: func(ctx context.Context, in *pb.GetCurrencyPairHistoryRequest) (*pb.GetCurrencyPairHistoryResponse, error) {
				return nil, errors.New("currency pair not found")
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		qResolver := resolver.Query()

		_, err := qResolver.CurrencyPairHistory(ctx, "XYZ", "USD", model.HistoryTimeframeTimeframe1m)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})
}

func TestQueryResolver_FxRates(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully queries fx rates with filters", func(t *testing.T) {
		baseCurr := "EUR"
		quoteCurr := "USD"
		fromDate := "2026-10-01"
		toDate := "2026-10-05"
		limit := 10
		offset := 0

		fakeClient := &fakePortfolioClient{
			listFXRatesFn: func(ctx context.Context, in *pb.ListFXRatesRequest) (*pb.ListFXRatesResponse, error) {
				if in.GetBaseCurrency() != baseCurr || in.GetQuoteCurrency() != quoteCurr {
					t.Errorf("unexpected pair filters: %s/%s", in.GetBaseCurrency(), in.GetQuoteCurrency())
				}
				if in.GetFromDate() != fromDate || in.GetToDate() != toDate {
					t.Errorf("unexpected date range: %s to %s", in.GetFromDate(), in.GetToDate())
				}
				if in.Limit != int32(limit) || in.Offset != int32(offset) {
					t.Errorf("unexpected pagination: limit=%d, offset=%d", in.Limit, in.Offset)
				}
				return &pb.ListFXRatesResponse{
					Rates: []*pb.FXRateItem{
						{
							BaseCurrency:  "EUR",
							QuoteCurrency: "USD",
							RateDate:      "2026-10-05",
							Rate:          &commonpb.Decimal{Value: "1.0850"},
							InvertedRate:  &commonpb.Decimal{Value: "0.921659"},
							Source:        "ECB",
						},
					},
					TotalCount: 1,
				}, nil
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		qResolver := resolver.Query()

		res, err := qResolver.FxRates(ctx, &baseCurr, &quoteCurr, &fromDate, &toDate, &limit, &offset)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if res.TotalCount != 1 {
			t.Fatalf("expected totalCount 1, got %d", res.TotalCount)
		}
		if len(res.Items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(res.Items))
		}
		item := res.Items[0]
		if item.Pair != "EUR/USD" || item.Rate.String() != "1.085" {
			t.Errorf("unexpected rate item: %+v", item)
		}
	})

	t.Run("handles error from portfolio service", func(t *testing.T) {
		fakeClient := &fakePortfolioClient{
			listFXRatesFn: func(ctx context.Context, in *pb.ListFXRatesRequest) (*pb.ListFXRatesResponse, error) {
				return nil, errors.New("query failed")
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		qResolver := resolver.Query()

		_, err := qResolver.FxRates(ctx, nil, nil, nil, nil, nil, nil)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})
}

func TestQueryResolver_Exchanges(t *testing.T) {
	ctx := context.Background()

	t.Run("returns exchanges mapped correctly", func(t *testing.T) {
		fakeClient := &fakePortfolioClient{
			listExchangesFn: func(ctx context.Context, in *pb.ListExchangesRequest) (*pb.ListExchangesResponse, error) {
				return &pb.ListExchangesResponse{
					Exchanges: []*pb.Exchange{
						{
							Code:     "XNAS",
							Name:     "NASDAQ Stock Market",
							Country:  "US",
							Timezone: "America/New_York",
						},
						{
							Code:     "XNYS",
							Name:     "New York Stock Exchange",
							Country:  "US",
							Timezone: "America/New_York",
						},
					},
				}, nil
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		qResolver := resolver.Query()

		res, err := qResolver.Exchanges(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(res) != 2 {
			t.Fatalf("expected 2 exchanges, got %d", len(res))
		}
		if res[0].Code != "XNAS" || res[0].Name != "NASDAQ Stock Market" || res[0].Country != "US" || res[0].Timezone != "America/New_York" {
			t.Errorf("unexpected exchange item 0: %+v", res[0])
		}
		if res[1].Code != "XNYS" || res[1].Name != "New York Stock Exchange" {
			t.Errorf("unexpected exchange item 1: %+v", res[1])
		}
	})

	t.Run("handles error from portfolio service", func(t *testing.T) {
		fakeClient := &fakePortfolioClient{
			listExchangesFn: func(ctx context.Context, in *pb.ListExchangesRequest) (*pb.ListExchangesResponse, error) {
				return nil, errors.New("rpc unavailable")
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		qResolver := resolver.Query()

		_, err := qResolver.Exchanges(ctx)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})
}

func TestQueryResolver_CheckTransactionDuplicates(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully returns existing external refs", func(t *testing.T) {
		fakeClient := &fakePortfolioClient{
			checkTransactionDuplicatesFn: func(ctx context.Context, in *pb.CheckTransactionDuplicatesRequest) (*pb.CheckTransactionDuplicatesResponse, error) {
				if in.GetUserId() != "1" {
					t.Errorf("expected user_id 1, got %s", in.GetUserId())
				}
				return &pb.CheckTransactionDuplicatesResponse{
					ExistingExternalRefs: []string{"cs_123"},
				}, nil
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		qResolver := resolver.Query()

		res, err := qResolver.CheckTransactionDuplicates(ctx, []string{"cs_123", "cs_456"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(res) != 1 || res[0] != "cs_123" {
			t.Fatalf("expected ['cs_123'], got %v", res)
		}
	})

	t.Run("handles error from service", func(t *testing.T) {
		fakeClient := &fakePortfolioClient{
			checkTransactionDuplicatesFn: func(ctx context.Context, in *pb.CheckTransactionDuplicatesRequest) (*pb.CheckTransactionDuplicatesResponse, error) {
				return nil, errors.New("service failure")
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		qResolver := resolver.Query()

		_, err := qResolver.CheckTransactionDuplicates(ctx, []string{"cs_123"})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestMutationResolver_ImportTransactions(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully forwards import transactions request", func(t *testing.T) {
		fakeClient := &fakePortfolioClient{
			batchImportTransactionsFn: func(ctx context.Context, in *pb.BatchImportTransactionsRequest) (*pb.BatchImportTransactionsResponse, error) {
				if len(in.GetTransactions()) != 1 {
					t.Errorf("expected 1 tx, got %d", len(in.GetTransactions()))
				}
				tx := in.GetTransactions()[0]
				if tx.GetSymbol() != "BHP" || tx.GetExternalRef() != "cs_123" {
					t.Errorf("unexpected tx: %+v", tx)
				}
				return &pb.BatchImportTransactionsResponse{
					Success:       true,
					ImportedCount: 1,
					SkippedCount:  0,
					Message:       "Imported 1 transactions",
					Portfolio: &pb.Portfolio{
						TotalValue: &commonpb.Money{
							Amount:       &commonpb.Decimal{Value: "1000.00"},
							CurrencyCode: "AUD",
						},
					},
				}, nil
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		mResolver := resolver.Mutation()

		skip := true
		res, err := mResolver.ImportTransactions(ctx, model.BatchImportTransactionsInput{
			Transactions: []*model.ImportTransactionInput{
				{
					ExternalRef: "cs_123",
					Symbol:      "BHP",
					Type:        model.TransactionTypeBuy,
					TradeDate:   "2025-05-10",
					Quantity:    model.Decimal(decimal.NewFromInt(10)),
					Price:       model.Decimal(decimal.RequireFromString("45.00")),
					Amount:      model.Decimal(decimal.RequireFromString("450.00")),
					Fee:         model.Decimal(decimal.Zero),
				},
			},
			SkipDuplicates: &skip,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !res.Success || res.ImportedCount != 1 || res.SkippedCount != 0 {
			t.Fatalf("unexpected response: %+v", res)
		}
	})

	t.Run("handles error from service", func(t *testing.T) {
		fakeClient := &fakePortfolioClient{
			batchImportTransactionsFn: func(ctx context.Context, in *pb.BatchImportTransactionsRequest) (*pb.BatchImportTransactionsResponse, error) {
				return nil, errors.New("batch import failed")
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		mResolver := resolver.Mutation()

		_, err := mResolver.ImportTransactions(ctx, model.BatchImportTransactionsInput{
			Transactions: []*model.ImportTransactionInput{},
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestQueryResolver_CashFlowReport(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully queries cash flow report", func(t *testing.T) {
		sym := "BHP.AX"
		name := "BHP Group"
		fakeClient := &fakePortfolioClient{
			getCashFlowReportFn: func(ctx context.Context, in *pb.GetCashFlowReportRequest) (*pb.GetCashFlowReportResponse, error) {
				if in.GetUserId() != "1" {
					t.Errorf("expected user_id 1, got %s", in.GetUserId())
				}
				if in.GetTimeframe() != pb.CashFlowTimeframe_CASH_FLOW_TIMEFRAME_MTD {
					t.Errorf("expected MTD timeframe, got %v", in.GetTimeframe())
				}
				return &pb.GetCashFlowReportResponse{
					Summary: &pb.CashFlowSummary{
						StartingCashBalance: &commonpb.Money{Amount: &commonpb.Decimal{Value: "5000.00"}, CurrencyCode: "AUD"},
						TotalInflows:        &commonpb.Money{Amount: &commonpb.Decimal{Value: "1500.00"}, CurrencyCode: "AUD"},
						TotalOutflows:       &commonpb.Money{Amount: &commonpb.Decimal{Value: "2000.00"}, CurrencyCode: "AUD"},
						NetCashFlow:         &commonpb.Money{Amount: &commonpb.Decimal{Value: "-500.00"}, CurrencyCode: "AUD"},
						EndingCashBalance:   &commonpb.Money{Amount: &commonpb.Decimal{Value: "4500.00"}, CurrencyCode: "AUD"},
					},
					Breakdown: &pb.CashFlowCategoryBreakdown{
						Deposits:      &commonpb.Money{Amount: &commonpb.Decimal{Value: "1000.00"}, CurrencyCode: "AUD"},
						Dividends:     &commonpb.Money{Amount: &commonpb.Decimal{Value: "500.00"}, CurrencyCode: "AUD"},
						Interest:      &commonpb.Money{Amount: &commonpb.Decimal{Value: "0.00"}, CurrencyCode: "AUD"},
						SalesProceeds: &commonpb.Money{Amount: &commonpb.Decimal{Value: "0.00"}, CurrencyCode: "AUD"},
						Withdrawals:   &commonpb.Money{Amount: &commonpb.Decimal{Value: "0.00"}, CurrencyCode: "AUD"},
						Purchases:     &commonpb.Money{Amount: &commonpb.Decimal{Value: "2000.00"}, CurrencyCode: "AUD"},
						Fees:          &commonpb.Money{Amount: &commonpb.Decimal{Value: "0.00"}, CurrencyCode: "AUD"},
						Taxes:         &commonpb.Money{Amount: &commonpb.Decimal{Value: "0.00"}, CurrencyCode: "AUD"},
					},
					Items: []*pb.CashFlowItem{
						{
							Id:             "item-1",
							EventDate:      "2026-10-05",
							Type:           pb.TransactionType_TRANSACTION_TYPE_DIVIDEND,
							FlowDirection:  "INFLOW",
							Category:       "DIVIDENDS",
							Symbol:         sym,
							InstrumentName: name,
							Description:    "Dividend from BHP.AX",
							NetAmount:      &commonpb.Money{Amount: &commonpb.Decimal{Value: "500.00"}, CurrencyCode: "AUD"},
							RunningBalance: &commonpb.Money{Amount: &commonpb.Decimal{Value: "5500.00"}, CurrencyCode: "AUD"},
							LocalAmount:    &commonpb.Money{Amount: &commonpb.Decimal{Value: "500.00"}, CurrencyCode: "AUD"},
							Fee:            &commonpb.Money{Amount: &commonpb.Decimal{Value: "0.00"}, CurrencyCode: "AUD"},
							WithholdingTax: &commonpb.Money{Amount: &commonpb.Decimal{Value: "0.00"}, CurrencyCode: "AUD"},
						},
					},
					BaseCurrency: "AUD",
					FromDate:     "2026-10-01",
					ToDate:       "2026-10-15",
				}, nil
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		qResolver := resolver.Query()

		res, err := qResolver.CashFlowReport(ctx, model.CashFlowFilterInput{
			Timeframe: model.CashFlowTimeframeMtd,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res == nil {
			t.Fatalf("expected result, got nil")
		}
		if res.BaseCurrency != "AUD" {
			t.Errorf("expected base currency AUD, got %s", res.BaseCurrency)
		}
		if res.Summary.StartingCashBalance.Amount.String() != "5000.00" && res.Summary.StartingCashBalance.Amount.String() != "5000" {
			t.Errorf("expected starting cash 5000.00, got %s", res.Summary.StartingCashBalance.Amount.String())
		}
		if res.Summary.EndingCashBalance.Amount.String() != "4500.00" && res.Summary.EndingCashBalance.Amount.String() != "4500" {
			t.Errorf("expected ending cash 4500.00, got %s", res.Summary.EndingCashBalance.Amount.String())
		}
		if len(res.Items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(res.Items))
		}
		if res.Items[0].Description != "Dividend from BHP.AX" {
			t.Errorf("expected description 'Dividend from BHP.AX', got %s", res.Items[0].Description)
		}
	})

	t.Run("handles error from portfolio service", func(t *testing.T) {
		fakeClient := &fakePortfolioClient{
			getCashFlowReportFn: func(ctx context.Context, in *pb.GetCashFlowReportRequest) (*pb.GetCashFlowReportResponse, error) {
				return nil, errors.New("portfolio service unavailable")
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		qResolver := resolver.Query()

		_, err := qResolver.CashFlowReport(ctx, model.CashFlowFilterInput{
			Timeframe: model.CashFlowTimeframeMtd,
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}


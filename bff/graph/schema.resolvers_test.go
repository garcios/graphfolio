package graph

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"bff/graph/model"
	"bff/internal/auth"

	commonpb "graphfolio/proto/common/v1"
	pb "graphfolio/proto/portfolio/v1"
	userpb "graphfolio/proto/user/v1"
	"pkg/decimalpb"

	"github.com/shopspring/decimal"
	"google.golang.org/grpc"
)

type fakeUserClient struct {
	userpb.UserServiceClient
	getUserPreferencesFn      func(ctx context.Context, in *userpb.GetUserPreferencesRequest) (*userpb.GetUserPreferencesResponse, error)
	updateUserPreferencesFn   func(ctx context.Context, in *userpb.UpdateUserPreferencesRequest) (*userpb.UpdateUserPreferencesResponse, error)
	listSupportedCurrenciesFn func(ctx context.Context, in *userpb.ListSupportedCurrenciesRequest) (*userpb.ListSupportedCurrenciesResponse, error)
	registerFn                func(ctx context.Context, in *userpb.RegisterRequest) (*userpb.RegisterResponse, error)
	loginFn                   func(ctx context.Context, in *userpb.LoginRequest) (*userpb.LoginResponse, error)
	refreshTokenFn            func(ctx context.Context, in *userpb.RefreshTokenRequest) (*userpb.RefreshTokenResponse, error)
	logoutFn                  func(ctx context.Context, in *userpb.LogoutRequest) (*userpb.LogoutResponse, error)
	getMeFn                   func(ctx context.Context, in *userpb.GetMeRequest) (*userpb.GetMeResponse, error)
	changePasswordFn          func(ctx context.Context, in *userpb.ChangePasswordRequest) (*userpb.ChangePasswordResponse, error)
}

func (f *fakeUserClient) ChangePassword(ctx context.Context, in *userpb.ChangePasswordRequest, opts ...grpc.CallOption) (*userpb.ChangePasswordResponse, error) {
	if f.changePasswordFn != nil {
		return f.changePasswordFn(ctx, in)
	}
	return nil, nil
}

func (f *fakeUserClient) Register(ctx context.Context, in *userpb.RegisterRequest, opts ...grpc.CallOption) (*userpb.RegisterResponse, error) {
	if f.registerFn != nil {
		return f.registerFn(ctx, in)
	}
	return nil, nil
}

func (f *fakeUserClient) Login(ctx context.Context, in *userpb.LoginRequest, opts ...grpc.CallOption) (*userpb.LoginResponse, error) {
	if f.loginFn != nil {
		return f.loginFn(ctx, in)
	}
	return nil, nil
}

func (f *fakeUserClient) RefreshToken(ctx context.Context, in *userpb.RefreshTokenRequest, opts ...grpc.CallOption) (*userpb.RefreshTokenResponse, error) {
	if f.refreshTokenFn != nil {
		return f.refreshTokenFn(ctx, in)
	}
	return nil, nil
}

func (f *fakeUserClient) Logout(ctx context.Context, in *userpb.LogoutRequest, opts ...grpc.CallOption) (*userpb.LogoutResponse, error) {
	if f.logoutFn != nil {
		return f.logoutFn(ctx, in)
	}
	return nil, nil
}

func (f *fakeUserClient) GetMe(ctx context.Context, in *userpb.GetMeRequest, opts ...grpc.CallOption) (*userpb.GetMeResponse, error) {
	if f.getMeFn != nil {
		return f.getMeFn(ctx, in)
	}
	return nil, nil
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
	getPortfolioFn                 func(ctx context.Context, in *pb.GetPortfolioRequest) (*pb.GetPortfolioResponse, error)
	updatePortfolioBaseCurrencyFn  func(ctx context.Context, in *pb.UpdatePortfolioBaseCurrencyRequest) (*pb.UpdatePortfolioBaseCurrencyResponse, error)
	addTransactionFn               func(ctx context.Context, in *pb.AddTransactionRequest) (*pb.AddTransactionResponse, error)
	listInstrumentsFn              func(ctx context.Context, in *pb.ListInstrumentsRequest) (*pb.ListInstrumentsResponse, error)
	getPortfolioHistoryFn          func(ctx context.Context, in *pb.GetPortfolioHistoryRequest) (*pb.GetPortfolioHistoryResponse, error)
	listTransactionsFn             func(ctx context.Context, in *pb.ListTransactionsRequest) (*pb.ListTransactionsResponse, error)
	deleteTransactionFn            func(ctx context.Context, in *pb.DeleteTransactionRequest) (*pb.DeleteTransactionResponse, error)
	updateTransactionFn            func(ctx context.Context, in *pb.UpdateTransactionRequest) (*pb.UpdateTransactionResponse, error)
	listAllInstrumentsFn           func(ctx context.Context, in *pb.ListAllInstrumentsRequest) (*pb.ListAllInstrumentsResponse, error)
	createInstrumentFn             func(ctx context.Context, in *pb.CreateInstrumentRequest) (*pb.CreateInstrumentResponse, error)
	updateInstrumentFn             func(ctx context.Context, in *pb.UpdateInstrumentRequest) (*pb.UpdateInstrumentResponse, error)
	deleteInstrumentFn             func(ctx context.Context, in *pb.DeleteInstrumentRequest) (*pb.DeleteInstrumentResponse, error)
	listExchangesFn                func(ctx context.Context, in *pb.ListExchangesRequest) (*pb.ListExchangesResponse, error)
	listInstrumentPricesFn         func(ctx context.Context, in *pb.ListInstrumentPricesRequest) (*pb.ListInstrumentPricesResponse, error)
	recordPriceOverrideFn          func(ctx context.Context, in *pb.RecordPriceOverrideRequest) (*pb.RecordPriceOverrideResponse, error)
	getIngestionStatusFn           func(ctx context.Context, in *pb.GetIngestionStatusRequest) (*pb.GetIngestionStatusResponse, error)
	triggerMarketSyncFn            func(ctx context.Context, in *pb.TriggerMarketSyncRequest) (*pb.TriggerMarketSyncResponse, error)
	triggerDividendSyncFn          func(ctx context.Context, in *pb.TriggerDividendSyncRequest) (*pb.TriggerDividendSyncResponse, error)
	triggerBackfillFn              func(ctx context.Context, in *pb.TriggerBackfillRequest) (*pb.TriggerBackfillResponse, error)
	listCurrencyPairsFn            func(ctx context.Context, in *pb.ListCurrencyPairsRequest) (*pb.ListCurrencyPairsResponse, error)
	listFXRatesFn                  func(ctx context.Context, in *pb.ListFXRatesRequest) (*pb.ListFXRatesResponse, error)
	getCurrencyPairHistoryFn       func(ctx context.Context, in *pb.GetCurrencyPairHistoryRequest) (*pb.GetCurrencyPairHistoryResponse, error)
	recordFXRateOverrideFn         func(ctx context.Context, in *pb.RecordFXRateOverrideRequest) (*pb.RecordFXRateOverrideResponse, error)
	checkTransactionDuplicatesFn   func(ctx context.Context, in *pb.CheckTransactionDuplicatesRequest) (*pb.CheckTransactionDuplicatesResponse, error)
	batchImportTransactionsFn      func(ctx context.Context, in *pb.BatchImportTransactionsRequest) (*pb.BatchImportTransactionsResponse, error)
	getCashFlowReportFn            func(ctx context.Context, in *pb.GetCashFlowReportRequest) (*pb.GetCashFlowReportResponse, error)
	listMissingPriceTransactionsFn  func(ctx context.Context, in *pb.ListMissingPriceTransactionsRequest) (*pb.ListMissingPriceTransactionsResponse, error)
	listMissingFXRateTransactionsFn func(ctx context.Context, in *pb.ListMissingFXRateTransactionsRequest) (*pb.ListMissingFXRateTransactionsResponse, error)
	getModifiedDietzReturnFn        func(ctx context.Context, in *pb.GetModifiedDietzReturnRequest) (*pb.GetModifiedDietzReturnResponse, error)
	listBenchmarksFn                func(ctx context.Context, in *pb.ListBenchmarksRequest) (*pb.ListBenchmarksResponse, error)
	getBenchmarkComparisonFn        func(ctx context.Context, in *pb.GetBenchmarkComparisonRequest) (*pb.GetBenchmarkComparisonResponse, error)
	createPortfolioFn               func(ctx context.Context, in *pb.CreatePortfolioRequest) (*pb.CreatePortfolioResponse, error)
	getDividendCalendarFn           func(ctx context.Context, in *pb.GetDividendCalendarRequest) (*pb.GetDividendCalendarResponse, error)
	getDividendYieldAnalyticsFn     func(ctx context.Context, in *pb.GetDividendYieldAnalyticsRequest) (*pb.GetDividendYieldAnalyticsResponse, error)
	simulateDRIPFn                  func(ctx context.Context, in *pb.SimulateDRIPRequest) (*pb.SimulateDRIPResponse, error)
}

func (f *fakePortfolioClient) CreatePortfolio(ctx context.Context, in *pb.CreatePortfolioRequest, opts ...grpc.CallOption) (*pb.CreatePortfolioResponse, error) {
	if f.createPortfolioFn != nil {
		return f.createPortfolioFn(ctx, in)
	}
	return nil, nil
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

func (f *fakePortfolioClient) UpdateTransaction(ctx context.Context, in *pb.UpdateTransactionRequest, opts ...grpc.CallOption) (*pb.UpdateTransactionResponse, error) {
	if f.updateTransactionFn != nil {
		return f.updateTransactionFn(ctx, in)
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

func (f *fakePortfolioClient) TriggerDividendSync(ctx context.Context, in *pb.TriggerDividendSyncRequest, opts ...grpc.CallOption) (*pb.TriggerDividendSyncResponse, error) {
	if f.triggerDividendSyncFn != nil {
		return f.triggerDividendSyncFn(ctx, in)
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

func (f *fakePortfolioClient) ListMissingPriceTransactions(ctx context.Context, in *pb.ListMissingPriceTransactionsRequest, opts ...grpc.CallOption) (*pb.ListMissingPriceTransactionsResponse, error) {
	if f.listMissingPriceTransactionsFn != nil {
		return f.listMissingPriceTransactionsFn(ctx, in)
	}
	return nil, nil
}

func (f *fakePortfolioClient) ListMissingFXRateTransactions(ctx context.Context, in *pb.ListMissingFXRateTransactionsRequest, opts ...grpc.CallOption) (*pb.ListMissingFXRateTransactionsResponse, error) {
	if f.listMissingFXRateTransactionsFn != nil {
		return f.listMissingFXRateTransactionsFn(ctx, in)
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

	result, err := qResolver.Portfolio(ctx, nil)
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

	result, err := qResolver.Portfolio(ctx, nil)
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

func TestQueryResolver_Portfolio_IncludeClosed(t *testing.T) {
	ctx := context.Background()

	var receivedIncludeClosed bool
	fakeClient := &fakePortfolioClient{
		getPortfolioFn: func(ctx context.Context, in *pb.GetPortfolioRequest) (*pb.GetPortfolioResponse, error) {
			receivedIncludeClosed = in.IncludeClosedPositions
			return &pb.GetPortfolioResponse{
				Portfolio: &pb.Portfolio{
					TotalValue: &commonpb.Money{
						Amount:       &commonpb.Decimal{Value: "0.00"},
						CurrencyCode: "USD",
					},
					Investments: []*pb.Investment{
						{
							Id:                 "inv-closed",
							Ticker:             "TSLA",
							Name:               "Tesla Inc",
							Price:              &commonpb.Money{Amount: &commonpb.Decimal{Value: "210.00"}, CurrencyCode: "USD"},
							Quantity:           &commonpb.Decimal{Value: "0"},
							TotalValue:         &commonpb.Money{Amount: &commonpb.Decimal{Value: "0.00"}, CurrencyCode: "USD"},
							CapitalGainAmount:  &commonpb.Money{Amount: &commonpb.Decimal{Value: "500.00"}, CurrencyCode: "USD"},
							CapitalGainPercent: &commonpb.Decimal{Value: "25.00"},
							TotalReturnAmount:  &commonpb.Money{Amount: &commonpb.Decimal{Value: "500.00"}, CurrencyCode: "USD"},
							TotalReturnPercent: &commonpb.Decimal{Value: "25.0"},
							IsClosed:           true,
						},
					},
				},
			}, nil
		},
	}

	resolver := &Resolver{PortfolioClient: fakeClient}
	qResolver := resolver.Query()

	inc := true
	result, err := qResolver.Portfolio(ctx, &inc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !receivedIncludeClosed {
		t.Errorf("expected gRPC request to have IncludeClosedPositions = true")
	}
	if len(result.Investments) != 1 {
		t.Fatalf("expected 1 investment, got %d", len(result.Investments))
	}
	if !result.Investments[0].IsClosed {
		t.Errorf("expected IsClosed to be true")
	}
	if result.Investments[0].Ticker != "TSLA" {
		t.Errorf("expected ticker TSLA, got %s", result.Investments[0].Ticker)
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

func TestMutationResolver_AddTransaction_Interest(t *testing.T) {
	ctx := context.Background()

	fakeClient := &fakePortfolioClient{
		addTransactionFn: func(ctx context.Context, in *pb.AddTransactionRequest) (*pb.AddTransactionResponse, error) {
			if in.GetType() != pb.TransactionType_TRANSACTION_TYPE_INTEREST {
				t.Errorf("expected TRANSACTION_TYPE_INTEREST, got %v", in.GetType())
			}
			if in.GetSymbol() != "" {
				t.Errorf("expected empty symbol for interest, got %s", in.GetSymbol())
			}
			if in.GetAmount().GetCurrencyCode() != "AUD" {
				t.Errorf("expected currency AUD, got %s", in.GetAmount().GetCurrencyCode())
			}
			amt, err := decimalpb.FromProto(in.GetAmount().GetAmount())
			if err != nil {
				t.Fatalf("failed to parse amount: %v", err)
			}
			if !amt.Equal(decimal.RequireFromString("125.50")) {
				t.Errorf("expected amount 125.50, got %s", amt)
			}
			if in.GetNotes() != "CMA Monthly Interest" {
				t.Errorf("expected notes 'CMA Monthly Interest', got %s", in.GetNotes())
			}
			return &pb.AddTransactionResponse{
				TransactionId: "tx-interest-123",
				Portfolio: &pb.Portfolio{
					TotalValue:  decimalpb.MoneyToProto(decimal.RequireFromString("10125.50"), "AUD"),
					CashBalance: decimalpb.MoneyToProto(decimal.RequireFromString("125.50"), "AUD"),
				},
			}, nil
		},
	}

	resolver := &Resolver{PortfolioClient: fakeClient}
	mResolver := resolver.Mutation()

	amount := model.Decimal(decimal.RequireFromString("125.50"))
	curr := "AUD"
	notes := "CMA Monthly Interest"

	payload, err := mResolver.AddTransaction(ctx, model.AddTransactionInput{
		Type:         model.TransactionTypeInterest,
		TradeDate:    "2026-10-18",
		Amount:       &amount,
		CurrencyCode: &curr,
		Notes:        &notes,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload == nil {
		t.Fatalf("expected payload, got nil")
	}
	if payload.TransactionID != "tx-interest-123" {
		t.Errorf("expected tx ID tx-interest-123, got %s", payload.TransactionID)
	}
	if !decimal.Decimal(payload.Portfolio.CashBalance.Amount).Equal(decimal.RequireFromString("125.50")) {
		t.Errorf("expected cash balance 125.50, got %s", decimal.Decimal(payload.Portfolio.CashBalance.Amount))
	}
}

func TestQueryResolver_PortfolioHistory(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully maps valuation time series and returns", func(t *testing.T) {
		fakeClient := &fakePortfolioClient{
			getPortfolioHistoryFn: func(ctx context.Context, in *pb.GetPortfolioHistoryRequest) (*pb.GetPortfolioHistoryResponse, error) {
				if in.GetUserId() != "1" && in.GetUserId() != defaultUserID {
					t.Errorf("expected user_id 1 or defaultUserID, got %s", in.GetUserId())
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
				if in.GetUserId() != "1" && in.GetUserId() != defaultUserID {
					t.Errorf("expected user_id 1 or defaultUserID, got %s", in.GetUserId())
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

		sortDir := model.SortDirectionAsc
		res, err := qResolver.Transactions(ctx, &typeArg, &symbolArg, &pageArg, &pageSizeArg, &sortDir)
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

		_, err := qResolver.Transactions(ctx, nil, nil, nil, nil, nil)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if err != context.DeadlineExceeded {
			t.Errorf("expected DeadlineExceeded, got %v", err)
		}
	})
}

func TestQueryResolver_ExportTransactions(t *testing.T) {
	ctx := context.Background()

	t.Run("success requests unpaginated ascending transactions", func(t *testing.T) {
		typeArg := model.TransactionTypeBuy
		symbolArg := "AAPL"

		fakeClient := &fakePortfolioClient{
			listTransactionsFn: func(ctx context.Context, in *pb.ListTransactionsRequest) (*pb.ListTransactionsResponse, error) {
				if in.PageSize != -1 {
					t.Errorf("expected PageSize -1 for export, got %d", in.PageSize)
				}
				if in.SortDirection != pb.SortDirection_SORT_DIRECTION_ASC {
					t.Errorf("expected SortDirectionAsc, got %v", in.SortDirection)
				}
				if in.Symbol != "AAPL" {
					t.Errorf("expected symbol AAPL, got %s", in.Symbol)
				}
				return &pb.ListTransactionsResponse{
					Transactions: []*pb.TransactionItem{
						{
							Id:             "tx-1",
							Type:           pb.TransactionType_TRANSACTION_TYPE_BUY,
							Symbol:         "AAPL",
							InstrumentName: "Apple Inc.",
							TradeDate:      "2023-01-15",
							Quantity:       &commonpb.Decimal{Value: "10"},
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
							WithholdingTax: &commonpb.Money{
								Amount:       &commonpb.Decimal{Value: "0.00"},
								CurrencyCode: "USD",
							},
							FxRateToBase: &commonpb.Decimal{Value: "1.500000"},
							ExternalRef:  "BROKER-123",
							Notes:        "Inception purchase",
							CreatedAt:    "2023-01-15T10:00:00Z",
						},
					},
					TotalCount: 1,
				}, nil
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		qResolver := resolver.Query()

		items, err := qResolver.ExportTransactions(ctx, &typeArg, &symbolArg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(items))
		}
		if items[0].ID != "tx-1" {
			t.Errorf("expected tx-1, got %s", items[0].ID)
		}
		if items[0].ExternalRef == nil || *items[0].ExternalRef != "BROKER-123" {
			t.Errorf("expected BROKER-123 external ref, got %v", items[0].ExternalRef)
		}
		if items[0].FxRateToBase == nil || (items[0].FxRateToBase.String() != "1.5" && items[0].FxRateToBase.String() != "1.500000") {
			t.Errorf("expected fxRateToBase 1.5, got %v", items[0].FxRateToBase)
		}
	})
}

func TestMutationResolver_DeleteTransaction(t *testing.T) {
	ctx := context.Background()

	t.Run("success deletes transaction and returns updated portfolio", func(t *testing.T) {
		fakeClient := &fakePortfolioClient{
			deleteTransactionFn: func(ctx context.Context, in *pb.DeleteTransactionRequest) (*pb.DeleteTransactionResponse, error) {
				if in.GetUserId() != "1" && in.GetUserId() != defaultUserID {
					t.Errorf("expected user_id 1 or defaultUserID, got %s", in.GetUserId())
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

func TestMutationResolver_UpdateTransaction(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully updates transaction and returns payload", func(t *testing.T) {
		fakeClient := &fakePortfolioClient{
			updateTransactionFn: func(ctx context.Context, in *pb.UpdateTransactionRequest) (*pb.UpdateTransactionResponse, error) {
				if in.TransactionId != "tx-123" {
					t.Errorf("expected transaction id tx-123, got %s", in.TransactionId)
				}
				if in.Type != pb.TransactionType_TRANSACTION_TYPE_BUY {
					t.Errorf("expected BUY, got %v", in.Type)
				}
				return &pb.UpdateTransactionResponse{
					Success: true,
					Transaction: &pb.TransactionItem{
						Id:        "tx-123",
						Type:      pb.TransactionType_TRANSACTION_TYPE_BUY,
						Symbol:    "AAPL",
						TradeDate: "2025-04-12",
						Amount:    &commonpb.Money{Amount: &commonpb.Decimal{Value: "9000.00"}, CurrencyCode: "USD"},
						Fee:       &commonpb.Money{Amount: &commonpb.Decimal{Value: "5.00"}, CurrencyCode: "USD"},
					},
					Portfolio: &pb.Portfolio{
						TotalValue: &commonpb.Money{Amount: &commonpb.Decimal{Value: "150000.00"}, CurrencyCode: "USD"},
					},
				}, nil
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		mResolver := resolver.Mutation()

		sym := "AAPL"
		qty := model.FromDecimal(decimal.RequireFromString("50"))
		price := model.FromDecimal(decimal.RequireFromString("180.00"))
		amount := model.FromDecimal(decimal.RequireFromString("9000.00"))
		fee := model.FromDecimal(decimal.RequireFromString("5.00"))

		payload, err := mResolver.UpdateTransaction(ctx, "tx-123", model.UpdateTransactionInput{
			Type:      model.TransactionTypeBuy,
			Symbol:    &sym,
			TradeDate: "2025-04-12",
			Quantity:  &qty,
			Price:     &price,
			Amount:    amount,
			Fee:       &fee,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if payload == nil || !payload.Success {
			t.Fatalf("expected success payload")
		}
		if *payload.Transaction.Symbol != "AAPL" {
			t.Errorf("expected symbol AAPL, got %s", *payload.Transaction.Symbol)
		}
	})

	t.Run("propagates error on failure", func(t *testing.T) {
		fakeClient := &fakePortfolioClient{
			updateTransactionFn: func(ctx context.Context, in *pb.UpdateTransactionRequest) (*pb.UpdateTransactionResponse, error) {
				return nil, errors.New("negative share inventory")
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		mResolver := resolver.Mutation()

		_, err := mResolver.UpdateTransaction(ctx, "tx-123", model.UpdateTransactionInput{
			Type:      model.TransactionTypeBuy,
			TradeDate: "2025-04-12",
			Amount:    model.FromDecimal(decimal.RequireFromString("100")),
		})
		if err == nil {
			t.Fatal("expected error, got nil")
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

func TestMutationResolver_TriggerDividendSync(t *testing.T) {
	t.Run("successfully triggers dividend sync for admin user", func(t *testing.T) {
		adminCtx := context.WithValue(context.Background(), auth.UserRoleKey, "ADMIN")

		fakeClient := &fakePortfolioClient{
			triggerDividendSyncFn: func(ctx context.Context, in *pb.TriggerDividendSyncRequest) (*pb.TriggerDividendSyncResponse, error) {
				return &pb.TriggerDividendSyncResponse{
					Success:              true,
					SchedulesSynced:      4,
					PrimarySourceCount:   3,
					FallbackSourceCount:  1,
					PlaceholdersUpgraded: 1,
					Message:              "Synced 4 dividend schedules",
					DurationMs:           250,
				}, nil
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		mResolver := resolver.Mutation()

		syncHistorical := true
		forceProvider := "YAHOO"
		input := &model.TriggerDividendSyncInput{
			Symbols:        []string{"AAPL", "BHP.AX"},
			SyncHistorical: &syncHistorical,
			ForceProvider:  &forceProvider,
		}

		payload, err := mResolver.TriggerDividendSync(adminCtx, input)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !payload.Success || payload.SchedulesSynced != 4 || payload.PrimarySourceCount != 3 {
			t.Errorf("unexpected dividend sync payload: %+v", payload)
		}
	})

	t.Run("returns forbidden error for standard non-admin user", func(t *testing.T) {
		userCtx := context.WithValue(context.Background(), auth.UserRoleKey, "USER")

		resolver := &Resolver{PortfolioClient: &fakePortfolioClient{}}
		mResolver := resolver.Mutation()

		_, err := mResolver.TriggerDividendSync(userCtx, nil)
		if err == nil {
			t.Fatalf("expected error for non-admin, got nil")
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
						UserId:                "018f0000-0000-7000-8000-000000000001",
						Email:                 "oscar@example.com",
						DisplayName:           "Oscar Garcia",
						DisplayCurrency:       "USD",
						Theme:                 "DARK",
						FinancialYearEndMonth: 6,
						FinancialYearEndDay:   30,
						CreatedAt:             "2026-01-01T00:00:00Z",
						UpdatedAt:             "2026-01-01T00:00:00Z",
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
		if prefs.FinancialYearEndMonth != 6 || prefs.FinancialYearEndDay != 30 {
			t.Errorf("unexpected EOFY preferences: %d/%d", prefs.FinancialYearEndMonth, prefs.FinancialYearEndDay)
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

	t.Run("successfully updates EOFY preferences", func(t *testing.T) {
		fyM := 6
		fyD := 30
		fakeUser := &fakeUserClient{
			updateUserPreferencesFn: func(ctx context.Context, in *userpb.UpdateUserPreferencesRequest) (*userpb.UpdateUserPreferencesResponse, error) {
				if in.FinancialYearEndMonth == nil || *in.FinancialYearEndMonth != 6 {
					t.Errorf("expected FinancialYearEndMonth 6, got %v", in.FinancialYearEndMonth)
				}
				if in.FinancialYearEndDay == nil || *in.FinancialYearEndDay != 30 {
					t.Errorf("expected FinancialYearEndDay 30, got %v", in.FinancialYearEndDay)
				}
				return &userpb.UpdateUserPreferencesResponse{
					Preferences: &userpb.UserPreferences{
						UserId:                in.UserId,
						Email:                 "oscar@example.com",
						DisplayName:           "Oscar Garcia",
						DisplayCurrency:       "AUD",
						Theme:                 "DARK",
						FinancialYearEndMonth: *in.FinancialYearEndMonth,
						FinancialYearEndDay:   *in.FinancialYearEndDay,
						CreatedAt:             "2026-01-01T00:00:00Z",
						UpdatedAt:             "2026-01-02T00:00:00Z",
					},
				}, nil
			},
		}
		fakePort := &fakePortfolioClient{
			getPortfolioFn: func(ctx context.Context, in *pb.GetPortfolioRequest) (*pb.GetPortfolioResponse, error) {
				return &pb.GetPortfolioResponse{
					Portfolio: &pb.Portfolio{
						TotalValue: &commonpb.Money{
							Amount:       &commonpb.Decimal{Value: "100000.00"},
							CurrencyCode: "AUD",
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
			FinancialYearEndMonth: &fyM,
			FinancialYearEndDay:   &fyD,
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if payload.Preferences.FinancialYearEndMonth != 6 || payload.Preferences.FinancialYearEndDay != 30 {
			t.Errorf("unexpected EOFY in payload: %d/%d", payload.Preferences.FinancialYearEndMonth, payload.Preferences.FinancialYearEndDay)
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
				if in.GetUserId() != "1" && in.GetUserId() != defaultUserID {
					t.Errorf("expected user_id 1 or defaultUserID, got %s", in.GetUserId())
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
				if in.GetUserId() != "1" && in.GetUserId() != defaultUserID {
					t.Errorf("expected user_id 1 or defaultUserID, got %s", in.GetUserId())
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

func TestQueryResolver_MissingPriceTransactions(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully resolves missing price report", func(t *testing.T) {
		sym := "NVDA"
		fromDate := "2024-01-01"
		toDate := "2024-10-01"
		limit := 50
		offset := 0

		fakeClient := &fakePortfolioClient{
			listMissingPriceTransactionsFn: func(ctx context.Context, in *pb.ListMissingPriceTransactionsRequest) (*pb.ListMissingPriceTransactionsResponse, error) {
				if in.GetSymbol() != "NVDA" {
					t.Errorf("expected symbol NVDA, got %s", in.GetSymbol())
				}
				earliest := "2024-03-15"
				latest := "2024-06-20"
				return &pb.ListMissingPriceTransactionsResponse{
					Items: []*pb.MissingPriceTransactionItem{
						{
							TransactionId:  "tx-1",
							PortfolioId:    "port-1",
							InstrumentId:   "inst-1",
							Symbol:         "NVDA",
							AssetName:      "Nvidia Corp",
							ExchangeMic:    "XNAS",
							TradeDate:      "2024-03-15",
							Type:           pb.TransactionType_TRANSACTION_TYPE_BUY,
							Quantity:       &commonpb.Decimal{Value: "10.00"},
							ExecutionPrice: &commonpb.Money{Amount: &commonpb.Decimal{Value: "880.00"}, CurrencyCode: "USD"},
							CurrencyCode:   "USD",
						},
					},
					TotalCount:             1,
					UniqueInstrumentsCount: 1,
					EarliestTradeDate:      &earliest,
					LatestTradeDate:        &latest,
				}, nil
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		qResolver := resolver.Query()

		res, err := qResolver.MissingPriceTransactions(ctx, &sym, &fromDate, &toDate, &limit, &offset)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res == nil {
			t.Fatalf("expected result, got nil")
		}
		if res.TotalCount != 1 {
			t.Errorf("expected total count 1, got %d", res.TotalCount)
		}
		if res.UniqueInstrumentsCount != 1 {
			t.Errorf("expected unique instruments count 1, got %d", res.UniqueInstrumentsCount)
		}
		if len(res.Items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(res.Items))
		}
		if res.Items[0].Symbol != "NVDA" {
			t.Errorf("expected symbol NVDA, got %s", res.Items[0].Symbol)
		}
		if res.Items[0].Type != model.TransactionTypeBuy {
			t.Errorf("expected type BUY, got %v", res.Items[0].Type)
		}
		if *res.EarliestTradeDate != "2024-03-15" {
			t.Errorf("expected earliest trade date 2024-03-15, got %s", *res.EarliestTradeDate)
		}
	})

	t.Run("handles error from portfolio service", func(t *testing.T) {
		fakeClient := &fakePortfolioClient{
			listMissingPriceTransactionsFn: func(ctx context.Context, in *pb.ListMissingPriceTransactionsRequest) (*pb.ListMissingPriceTransactionsResponse, error) {
				return nil, errors.New("service unavailable")
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		qResolver := resolver.Query()

		_, err := qResolver.MissingPriceTransactions(ctx, nil, nil, nil, nil, nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestQueryResolver_MissingFXRateTransactions(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully resolves missing fx rate report", func(t *testing.T) {
		pair := "USD/AUD"
		fromDate := "2024-01-01"
		toDate := "2024-10-01"
		txType := model.TransactionTypeBuy
		limit := 50
		offset := 0

		fakeClient := &fakePortfolioClient{
			listMissingFXRateTransactionsFn: func(ctx context.Context, in *pb.ListMissingFXRateTransactionsRequest) (*pb.ListMissingFXRateTransactionsResponse, error) {
				if in.GetCurrencyPair() != "USD/AUD" {
					t.Errorf("expected currency pair USD/AUD, got %s", in.GetCurrencyPair())
				}
				earliest := "2024-02-14"
				latest := "2024-08-11"
				instID := "inst-1"
				sym := "AAPL"
				asset := "Apple Inc"
				return &pb.ListMissingFXRateTransactionsResponse{
					Items: []*pb.MissingFXRateTransactionItem{
						{
							TransactionId:    "tx-1",
							PortfolioId:      "port-1",
							PortfolioName:    "Growth Portfolio",
							PortfolioBaseCur: "AUD",
							TradeDate:        "2024-02-14",
							Type:             pb.TransactionType_TRANSACTION_TYPE_BUY,
							InstrumentId:     &instID,
							Symbol:           &sym,
							AssetName:        &asset,
							Amount:           &commonpb.Money{Amount: &commonpb.Decimal{Value: "3600.00"}, CurrencyCode: "USD"},
							TransactionCur:   "USD",
							RequiredPair:     "USD/AUD",
							RecordedFxRate:   &commonpb.Decimal{Value: "1.5280000000"},
						},
					},
					TotalCount:         1,
					UniquePairsCount:   1,
					EarliestTradeDate:  &earliest,
					LatestTradeDate:    &latest,
				}, nil
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		qResolver := resolver.Query()

		res, err := qResolver.MissingFXRateTransactions(ctx, &pair, &fromDate, &toDate, &txType, &limit, &offset)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res == nil {
			t.Fatalf("expected result, got nil")
		}
		if res.TotalCount != 1 {
			t.Errorf("expected total count 1, got %d", res.TotalCount)
		}
		if res.UniquePairsCount != 1 {
			t.Errorf("expected unique pairs count 1, got %d", res.UniquePairsCount)
		}
		if len(res.Items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(res.Items))
		}
		item := res.Items[0]
		if item.RequiredPair != "USD/AUD" {
			t.Errorf("expected pair USD/AUD, got %s", item.RequiredPair)
		}
		if item.Type != model.TransactionTypeBuy {
			t.Errorf("expected type BUY, got %v", item.Type)
		}
		if item.PortfolioBaseCurrency != "AUD" {
			t.Errorf("expected base currency AUD, got %s", item.PortfolioBaseCurrency)
		}
		if *res.EarliestTradeDate != "2024-02-14" {
			t.Errorf("expected earliest trade date 2024-02-14, got %s", *res.EarliestTradeDate)
		}
	})

	t.Run("handles error from portfolio service", func(t *testing.T) {
		fakeClient := &fakePortfolioClient{
			listMissingFXRateTransactionsFn: func(ctx context.Context, in *pb.ListMissingFXRateTransactionsRequest) (*pb.ListMissingFXRateTransactionsResponse, error) {
				return nil, errors.New("service unavailable")
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		qResolver := resolver.Query()

		_, err := qResolver.MissingFXRateTransactions(ctx, nil, nil, nil, nil, nil, nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}


func (f *fakePortfolioClient) GetModifiedDietzReturn(ctx context.Context, in *pb.GetModifiedDietzReturnRequest, opts ...grpc.CallOption) (*pb.GetModifiedDietzReturnResponse, error) {
	if f.getModifiedDietzReturnFn != nil {
		return f.getModifiedDietzReturnFn(ctx, in)
	}
	return nil, nil
}

func TestQueryResolver_ModifiedDietzReturn(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully resolves modified dietz return", func(t *testing.T) {
		fakeClient := &fakePortfolioClient{
			getModifiedDietzReturnFn: func(ctx context.Context, in *pb.GetModifiedDietzReturnRequest) (*pb.GetModifiedDietzReturnResponse, error) {
				if in.GetUserId() != "1" && in.GetUserId() != defaultUserID {
					t.Errorf("expected user_id 1 or defaultUserID, got %s", in.GetUserId())
				}
				if in.GetStartDate() != "2025-01-01" {
					t.Errorf("expected start date 2025-01-01, got %s", in.GetStartDate())
				}
				if in.GetEndDate() != "2025-12-31" {
					t.Errorf("expected end date 2025-12-31, got %s", in.GetEndDate())
				}

				return &pb.GetModifiedDietzReturnResponse{
					PortfolioId:             "1",
					BaseCurrency:            "USD",
					StartDate:               "2025-01-01",
					EndDate:                 "2025-12-31",
					ElapsedDays:             365,
					StartValuation:          decimalpb.MoneyToProto(decimal.NewFromInt(100000), "USD"),
					EndValuation:            decimalpb.MoneyToProto(decimal.NewFromInt(180000), "USD"),
					NetExternalFlows:        decimalpb.MoneyToProto(decimal.NewFromInt(50000), "USD"),
					NetInvestmentGain:       decimalpb.MoneyToProto(decimal.NewFromInt(30000), "USD"),
					WeightedCapitalEmployed: decimalpb.MoneyToProto(decimal.NewFromFloat(125205.48), "USD"),
					CumulativeReturnPercent: decimalpb.ToProto(decimal.NewFromFloat(23.96)),
					AnnualizedReturnPercent: decimalpb.ToProto(decimal.NewFromFloat(23.98)),
					IsAnnualized:            true,
					IsComputable:            true,
					DiagnosticCode:          "OK",
					CashFlowWeights: []*pb.ExternalCashFlowWeight{
						{
							TransactionId:  "tx-1",
							Date:           "2025-07-01",
							Type:           pb.TransactionType_TRANSACTION_TYPE_DEPOSIT,
							Amount:         decimalpb.MoneyToProto(decimal.NewFromInt(50000), "USD"),
							Weight:         decimalpb.ToProto(decimal.NewFromFloat(0.50410959)),
							WeightedAmount: decimalpb.MoneyToProto(decimal.NewFromFloat(25205.48), "USD"),
							LocalAmount:    decimalpb.MoneyToProto(decimal.NewFromInt(50000), "USD"),
						},
					},
				}, nil
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		qResolver := resolver.Query()

		res, err := qResolver.ModifiedDietzReturn(ctx, "2025-01-01", "2025-12-31", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if res.BaseCurrency != "USD" {
			t.Errorf("expected USD, got %s", res.BaseCurrency)
		}
		if res.ElapsedDays != 365 {
			t.Errorf("expected 365, got %d", res.ElapsedDays)
		}
		if !res.IsAnnualized {
			t.Errorf("expected isAnnualized=true")
		}
		if res.AnnualizedReturnPercent == nil || res.AnnualizedReturnPercent.String() != "23.98" {
			t.Errorf("expected annualized return 23.98, got %v", res.AnnualizedReturnPercent)
		}
		if len(res.CashFlowWeights) != 1 {
			t.Fatalf("expected 1 weight, got %d", len(res.CashFlowWeights))
		}
	})

	t.Run("handles error from portfolio service", func(t *testing.T) {
		fakeClient := &fakePortfolioClient{
			getModifiedDietzReturnFn: func(ctx context.Context, in *pb.GetModifiedDietzReturnRequest) (*pb.GetModifiedDietzReturnResponse, error) {
				return nil, errors.New("service unavailable")
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		qResolver := resolver.Query()

		_, err := qResolver.ModifiedDietzReturn(ctx, "2025-01-01", "2025-12-31", nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func (f *fakePortfolioClient) ListBenchmarks(ctx context.Context, in *pb.ListBenchmarksRequest, opts ...grpc.CallOption) (*pb.ListBenchmarksResponse, error) {
	if f.listBenchmarksFn != nil {
		return f.listBenchmarksFn(ctx, in)
	}
	return nil, nil
}

func (f *fakePortfolioClient) GetBenchmarkComparison(ctx context.Context, in *pb.GetBenchmarkComparisonRequest, opts ...grpc.CallOption) (*pb.GetBenchmarkComparisonResponse, error) {
	if f.getBenchmarkComparisonFn != nil {
		return f.getBenchmarkComparisonFn(ctx, in)
	}
	return nil, nil
}

func TestQueryResolver_Benchmarks(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully resolves benchmarks", func(t *testing.T) {
		fakeClient := &fakePortfolioClient{
			listBenchmarksFn: func(ctx context.Context, in *pb.ListBenchmarksRequest) (*pb.ListBenchmarksResponse, error) {
				return &pb.ListBenchmarksResponse{
					Benchmarks: []*pb.Benchmark{
						{
							Id:           "SP500TR",
							Symbol:       "SP500TR",
							Name:         "S&P 500 Total Return",
							ExchangeCode: "XNYS",
							CurrencyCode: "USD",
							IndexType:    "TOTAL_RETURN",
							IsDefault:    true,
						},
					},
				}, nil
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		qResolver := resolver.Query()

		res, err := qResolver.Benchmarks(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(res) != 1 {
			t.Fatalf("expected 1 benchmark, got %d", len(res))
		}
		if res[0].ID != "SP500TR" || res[0].Symbol != "SP500TR" {
			t.Errorf("unexpected benchmark: %+v", res[0])
		}
	})

	t.Run("handles error from portfolio service", func(t *testing.T) {
		fakeClient := &fakePortfolioClient{
			listBenchmarksFn: func(ctx context.Context, in *pb.ListBenchmarksRequest) (*pb.ListBenchmarksResponse, error) {
				return nil, errors.New("service unavailable")
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		qResolver := resolver.Query()

		_, err := qResolver.Benchmarks(ctx)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestQueryResolver_BenchmarkComparison(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully resolves benchmark comparison", func(t *testing.T) {
		portAnn, _ := decimalpb.FromString("15.50")
		benchAnn, _ := decimalpb.FromString("10.20")
		actAnn, _ := decimalpb.FromString("5.30")
		beta, _ := decimalpb.FromString("1.10")
		te, _ := decimalpb.FromString("4.20")
		ir, _ := decimalpb.FromString("1.26")
		cumPort, _ := decimalpb.FromString("20.00")
		cumBench, _ := decimalpb.FromString("15.00")
		actCum, _ := decimalpb.FromString("5.00")

		fakeClient := &fakePortfolioClient{
			getBenchmarkComparisonFn: func(ctx context.Context, in *pb.GetBenchmarkComparisonRequest) (*pb.GetBenchmarkComparisonResponse, error) {
				return &pb.GetBenchmarkComparisonResponse{
					PortfolioId:  "port-1",
					BaseCurrency: "USD",
					Benchmark: &pb.Benchmark{
						Id:           "SP500TR",
						Symbol:       "SP500TR",
						Name:         "S&P 500 Total Return",
						ExchangeCode: "XNYS",
						CurrencyCode: "USD",
						IndexType:    "TOTAL_RETURN",
						IsDefault:    true,
					},
					CurrencyMode: pb.BenchmarkCurrencyMode_BENCHMARK_CURRENCY_MODE_PORTFOLIO_BASE,
					StartDate:    "2025-01-01",
					EndDate:      "2026-01-01",
					ElapsedDays:  365,
					Points: []*pb.BenchmarkComparisonPoint{
						{
							Date:                   "2025-01-01",
							PortfolioReturnPercent: cumPort,
							BenchmarkReturnPercent: cumBench,
							ActiveReturnPercent:    actCum,
							PortfolioValue:         decimalpb.MoneyToProto(decimal.NewFromInt(100000), "USD"),
							BenchmarkPriceAdjusted: decimalpb.MoneyToProto(decimal.RequireFromString("500.00"), "USD"),
						},
					},
					Metrics: &pb.BenchmarkMetrics{
						PortfolioCumulativeReturnPercent: cumPort,
						BenchmarkCumulativeReturnPercent: cumBench,
						ActiveCumulativeReturnPercent:    actCum,
						PortfolioAnnualizedReturnPercent: portAnn,
						BenchmarkAnnualizedReturnPercent: benchAnn,
						ActiveAnnualizedReturnPercent:    actAnn,
						IsAnnualized:                     true,
						Beta:                             beta,
						TrackingErrorPercent:             te,
						InformationRatio:                 ir,
					},
				}, nil
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		qResolver := resolver.Query()

		res, err := qResolver.BenchmarkComparison(ctx, "SP500TR", model.HistoryTimeframeTimeframe1y, nil)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if res.PortfolioID != "port-1" || res.Benchmark.ID != "SP500TR" {
			t.Errorf("unexpected payload: %+v", res)
		}
		if len(res.Points) != 1 {
			t.Fatalf("expected 1 point, got %d", len(res.Points))
		}
		if res.Metrics.Beta == nil || (res.Metrics.Beta.String() != "1.10" && res.Metrics.Beta.String() != "1.1") {
			t.Errorf("expected beta 1.10 or 1.1, got %v", res.Metrics.Beta)
		}
	})

	t.Run("handles error from portfolio service", func(t *testing.T) {
		fakeClient := &fakePortfolioClient{
			getBenchmarkComparisonFn: func(ctx context.Context, in *pb.GetBenchmarkComparisonRequest) (*pb.GetBenchmarkComparisonResponse, error) {
				return nil, errors.New("service unavailable")
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		qResolver := resolver.Query()

		_, err := qResolver.BenchmarkComparison(ctx, "SP500TR", model.HistoryTimeframeTimeframe1y, nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestMutationResolver_Register(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully registers user and creates portfolio", func(t *testing.T) {
		fakeUser := &fakeUserClient{
			registerFn: func(ctx context.Context, in *userpb.RegisterRequest) (*userpb.RegisterResponse, error) {
				if in.Email != "newuser@example.com" {
					t.Errorf("expected email newuser@example.com, got %s", in.Email)
				}
				if in.Password != "Password123!" {
					t.Errorf("expected password Password123!, got %s", in.Password)
				}
				if in.DisplayName != "New Investor" {
					t.Errorf("expected displayName New Investor, got %s", in.DisplayName)
				}
				if in.DisplayCurrency != "USD" {
					t.Errorf("expected displayCurrency USD, got %s", in.DisplayCurrency)
				}
				return &userpb.RegisterResponse{
					User: &userpb.UserPreferences{
						UserId:          "user-reg-1",
						Email:           in.Email,
						DisplayName:     in.DisplayName,
						DisplayCurrency: in.DisplayCurrency,
						Theme:           "DARK",
					},
					AccessToken:  "at-reg-123",
					RefreshToken: "rt-reg-456",
				}, nil
			},
		}

		portfolioCreated := false
		fakePortfolio := &fakePortfolioClient{
			createPortfolioFn: func(ctx context.Context, in *pb.CreatePortfolioRequest) (*pb.CreatePortfolioResponse, error) {
				if in.UserId != "user-reg-1" {
					t.Errorf("expected user_id user-reg-1, got %s", in.UserId)
				}
				if in.BaseCurrency != "USD" {
					t.Errorf("expected baseCurrency USD, got %s", in.BaseCurrency)
				}
				portfolioCreated = true
				return &pb.CreatePortfolioResponse{PortfolioId: "port-reg-1", BaseCurrency: in.BaseCurrency}, nil
			},
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/query", nil)
		testCtx := context.WithValue(ctx, auth.ResponseWriterKey, w)
		testCtx = context.WithValue(testCtx, auth.RequestKey, r)

		resolver := &Resolver{UserClient: fakeUser, PortfolioClient: fakePortfolio}
		mResolver := resolver.Mutation()

		res, err := mResolver.Register(testCtx, model.RegisterInput{
			Email:           "newuser@example.com",
			Password:        "Password123!",
			DisplayName:     "New Investor",
			DisplayCurrency: "USD",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.AccessToken != "at-reg-123" {
			t.Errorf("expected access token at-reg-123, got %s", res.AccessToken)
		}
		if res.User.UserID != "user-reg-1" || res.User.Email != "newuser@example.com" {
			t.Errorf("unexpected user in payload: %+v", res.User)
		}
		if !portfolioCreated {
			t.Errorf("expected portfolio to be created")
		}

		cookies := w.Result().Cookies()
		foundCookie := false
		for _, c := range cookies {
			if c.Name == auth.RefreshCookieName && c.Value == "rt-reg-456" {
				foundCookie = true
			}
		}
		if !foundCookie {
			t.Errorf("expected refresh cookie to be set in response writer")
		}
	})

	t.Run("handles user client error", func(t *testing.T) {
		fakeUser := &fakeUserClient{
			registerFn: func(ctx context.Context, in *userpb.RegisterRequest) (*userpb.RegisterResponse, error) {
				return nil, errors.New("email already registered")
			},
		}

		resolver := &Resolver{UserClient: fakeUser, PortfolioClient: &fakePortfolioClient{}}
		mResolver := resolver.Mutation()

		_, err := mResolver.Register(ctx, model.RegisterInput{
			Email:           "exists@example.com",
			Password:        "Password123!",
			DisplayName:     "Existing",
			DisplayCurrency: "USD",
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestMutationResolver_Login(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully logs in user and sets cookie", func(t *testing.T) {
		fakeUser := &fakeUserClient{
			loginFn: func(ctx context.Context, in *userpb.LoginRequest) (*userpb.LoginResponse, error) {
				if in.Email != "login@example.com" || in.Password != "secret" {
					return nil, errors.New("invalid credentials")
				}
				return &userpb.LoginResponse{
					User: &userpb.UserPreferences{
						UserId:          "user-log-1",
						Email:           in.Email,
						DisplayName:     "Logged In User",
						DisplayCurrency: "USD",
						Theme:           "DARK",
					},
					AccessToken:  "at-log-123",
					RefreshToken: "rt-log-456",
				}, nil
			},
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/query", nil)
		testCtx := context.WithValue(ctx, auth.ResponseWriterKey, w)
		testCtx = context.WithValue(testCtx, auth.RequestKey, r)

		resolver := &Resolver{UserClient: fakeUser}
		mResolver := resolver.Mutation()

		res, err := mResolver.Login(testCtx, model.LoginInput{
			Email:    "login@example.com",
			Password: "secret",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.AccessToken != "at-log-123" {
			t.Errorf("expected access token at-log-123, got %s", res.AccessToken)
		}
		if res.User.UserID != "user-log-1" {
			t.Errorf("expected user-log-1, got %s", res.User.UserID)
		}
	})

	t.Run("handles invalid credentials error", func(t *testing.T) {
		fakeUser := &fakeUserClient{
			loginFn: func(ctx context.Context, in *userpb.LoginRequest) (*userpb.LoginResponse, error) {
				return nil, errors.New("invalid email or password")
			},
		}

		resolver := &Resolver{UserClient: fakeUser}
		mResolver := resolver.Mutation()

		_, err := mResolver.Login(ctx, model.LoginInput{
			Email:    "wrong@example.com",
			Password: "wrong",
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestMutationResolver_RefreshToken(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully refreshes token with cookie", func(t *testing.T) {
		fakeUser := &fakeUserClient{
			refreshTokenFn: func(ctx context.Context, in *userpb.RefreshTokenRequest) (*userpb.RefreshTokenResponse, error) {
				if in.RefreshToken != "valid-rt-cookie" {
					t.Errorf("expected valid-rt-cookie, got %s", in.RefreshToken)
				}
				return &userpb.RefreshTokenResponse{
					User: &userpb.UserPreferences{
						UserId:          "user-ref-1",
						Email:           "ref@example.com",
						DisplayName:     "Refresh User",
						DisplayCurrency: "USD",
						Theme:           "DARK",
					},
					AccessToken:  "at-new-123",
					RefreshToken: "rt-new-456",
				}, nil
			},
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/query", nil)
		r.AddCookie(&http.Cookie{
			Name:  auth.RefreshCookieName,
			Value: "valid-rt-cookie",
		})
		testCtx := context.WithValue(ctx, auth.ResponseWriterKey, w)
		testCtx = context.WithValue(testCtx, auth.RequestKey, r)

		resolver := &Resolver{UserClient: fakeUser}
		mResolver := resolver.Mutation()

		res, err := mResolver.RefreshToken(testCtx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.AccessToken != "at-new-123" {
			t.Errorf("expected at-new-123, got %s", res.AccessToken)
		}
	})

	t.Run("fails when refresh cookie is missing", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/query", nil)
		testCtx := context.WithValue(ctx, auth.ResponseWriterKey, w)
		testCtx = context.WithValue(testCtx, auth.RequestKey, r)

		resolver := &Resolver{UserClient: &fakeUserClient{}}
		mResolver := resolver.Mutation()

		_, err := mResolver.RefreshToken(testCtx)
		if err == nil {
			t.Fatal("expected error for missing cookie, got nil")
		}
	})
}

func TestMutationResolver_Logout(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully logs out and clears cookie", func(t *testing.T) {
		logoutCalled := false
		fakeUser := &fakeUserClient{
			logoutFn: func(ctx context.Context, in *userpb.LogoutRequest) (*userpb.LogoutResponse, error) {
				if in.RefreshToken != "token-to-revoke" {
					t.Errorf("expected token-to-revoke, got %s", in.RefreshToken)
				}
				logoutCalled = true
				return &userpb.LogoutResponse{Success: true}, nil
			},
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/query", nil)
		r.AddCookie(&http.Cookie{
			Name:  auth.RefreshCookieName,
			Value: "token-to-revoke",
		})
		testCtx := context.WithValue(ctx, auth.ResponseWriterKey, w)
		testCtx = context.WithValue(testCtx, auth.RequestKey, r)

		resolver := &Resolver{UserClient: fakeUser}
		mResolver := resolver.Mutation()

		ok, err := mResolver.Logout(testCtx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ok {
			t.Errorf("expected true")
		}
		if !logoutCalled {
			t.Errorf("expected logout to be called")
		}

		cookies := w.Result().Cookies()
		foundCleared := false
		for _, c := range cookies {
			if c.Name == auth.RefreshCookieName && c.MaxAge == -1 {
				foundCleared = true
			}
		}
		if !foundCleared {
			t.Errorf("expected cookie to be cleared with MaxAge -1")
		}
	})
}

func TestQueryResolver_Me(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully returns authenticated user", func(t *testing.T) {
		fakeUser := &fakeUserClient{
			getMeFn: func(ctx context.Context, in *userpb.GetMeRequest) (*userpb.GetMeResponse, error) {
				if in.UserId != "user-authenticated-123" {
					t.Errorf("expected user-authenticated-123, got %s", in.UserId)
				}
				return &userpb.GetMeResponse{
					User: &userpb.UserPreferences{
						UserId:          in.UserId,
						Email:           "authed@example.com",
						DisplayName:     "Authed User",
						DisplayCurrency: "AUD",
						Theme:           "DARK",
					},
				}, nil
			},
		}

		testCtx := context.WithValue(ctx, auth.UserIDKey, "user-authenticated-123")
		resolver := &Resolver{UserClient: fakeUser}
		qResolver := resolver.Query()

		user, err := qResolver.Me(testCtx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if user.UserID != "user-authenticated-123" || user.Email != "authed@example.com" {
			t.Errorf("unexpected user: %+v", user)
		}
	})

	t.Run("handles unauthenticated error", func(t *testing.T) {
		resolver := &Resolver{UserClient: &fakeUserClient{}}
		qResolver := resolver.Query()

		_, err := qResolver.Me(ctx)
		if err == nil {
			t.Fatal("expected error for unauthenticated user, got nil")
		}
	})
}

func TestMutationResolver_ChangePassword(t *testing.T) {
	ctx := context.Background()
	testUserID := "018f0000-0000-7000-8000-000000000001"
	authedCtx := context.WithValue(ctx, auth.UserIDKey, testUserID)

	t.Run("success", func(t *testing.T) {
		fakeUser := &fakeUserClient{
			changePasswordFn: func(ctx context.Context, in *userpb.ChangePasswordRequest) (*userpb.ChangePasswordResponse, error) {
				if in.UserId != testUserID {
					t.Errorf("expected user ID %s, got %s", testUserID, in.UserId)
				}
				if in.CurrentPassword != "OldPass#123" {
					t.Errorf("expected current password OldPass#123, got %s", in.CurrentPassword)
				}
				if in.NewPassword != "NewPass#456" {
					t.Errorf("expected new password NewPass#456, got %s", in.NewPassword)
				}
				if !in.RevokeOtherSessions {
					t.Errorf("expected RevokeOtherSessions true")
				}
				return &userpb.ChangePasswordResponse{
					Success: true,
					Message: "Password updated successfully",
				}, nil
			},
		}

		resolver := &Resolver{UserClient: fakeUser}
		mResolver := resolver.Mutation()

		revokeOthers := true
		res, err := mResolver.ChangePassword(authedCtx, model.ChangePasswordInput{
			CurrentPassword:     "OldPass#123",
			NewPassword:         "NewPass#456",
			RevokeOtherSessions: &revokeOthers,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !res.Success {
			t.Errorf("expected success true")
		}
	})

	t.Run("gRPC error forwarded", func(t *testing.T) {
		fakeUser := &fakeUserClient{
			changePasswordFn: func(ctx context.Context, in *userpb.ChangePasswordRequest) (*userpb.ChangePasswordResponse, error) {
				return nil, errors.New("current password is incorrect")
			},
		}

		resolver := &Resolver{UserClient: fakeUser}
		mResolver := resolver.Mutation()

		_, err := mResolver.ChangePassword(authedCtx, model.ChangePasswordInput{
			CurrentPassword: "WrongPassword#123",
			NewPassword:     "NewPass#456",
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func (f *fakePortfolioClient) GetDividendCalendar(ctx context.Context, in *pb.GetDividendCalendarRequest, opts ...grpc.CallOption) (*pb.GetDividendCalendarResponse, error) {
	if f.getDividendCalendarFn != nil {
		return f.getDividendCalendarFn(ctx, in)
	}
	return nil, nil
}

func (f *fakePortfolioClient) GetDividendYieldAnalytics(ctx context.Context, in *pb.GetDividendYieldAnalyticsRequest, opts ...grpc.CallOption) (*pb.GetDividendYieldAnalyticsResponse, error) {
	if f.getDividendYieldAnalyticsFn != nil {
		return f.getDividendYieldAnalyticsFn(ctx, in)
	}
	return nil, nil
}

func (f *fakePortfolioClient) SimulateDRIP(ctx context.Context, in *pb.SimulateDRIPRequest, opts ...grpc.CallOption) (*pb.SimulateDRIPResponse, error) {
	if f.simulateDRIPFn != nil {
		return f.simulateDRIPFn(ctx, in)
	}
	return nil, nil
}

func TestQueryResolver_DividendCalendar(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully resolves dividend calendar", func(t *testing.T) {
		fakeClient := &fakePortfolioClient{
			getDividendCalendarFn: func(ctx context.Context, in *pb.GetDividendCalendarRequest) (*pb.GetDividendCalendarResponse, error) {
				if in.GetYear() != 2026 || in.GetMonth() != 5 {
					t.Errorf("unexpected year/month: %d/%d", in.GetYear(), in.GetMonth())
				}
				return &pb.GetDividendCalendarResponse{
					Events: []*pb.DividendCalendarEvent{
						{
							Id:              "div-1",
							InstrumentId:    "inst-1",
							Symbol:          "BHP.AX",
							Name:            "BHP Group",
							ExDate:          "2026-05-15",
							PaymentDate:     "2026-05-28",
							SharesHeld:      decimalpb.ToProto(decimal.NewFromInt(100)),
							DpsDeclared:     decimalpb.MoneyToProto(decimal.NewFromFloat(1.50), "AUD"),
							GrossAmountBase: decimalpb.MoneyToProto(decimal.NewFromFloat(150.00), "USD"),
							NetAmountBase:   decimalpb.MoneyToProto(decimal.NewFromFloat(150.00), "USD"),
							Frequency:       pb.DividendFrequency_DIVIDEND_FREQUENCY_SEMI_ANNUAL,
							Status:          pb.DividendEventStatus_DIVIDEND_EVENT_STATUS_ANNOUNCED,
							IsUserEntitled:  true,
							IsEstimated:     false,
							FrankingPct:     decimalpb.ToProto(decimal.NewFromInt(100)),
						},
					},
				}, nil
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		res, err := resolver.Query().DividendCalendar(ctx, 2026, 5)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if len(res) != 1 {
			t.Fatalf("expected 1 event, got %d", len(res))
		}
		if res[0].Symbol != "BHP.AX" {
			t.Errorf("expected BHP.AX, got %s", res[0].Symbol)
		}
		if res[0].Status != model.DividendStatusAnnounced {
			t.Errorf("expected ANNOUNCED status, got %s", res[0].Status)
		}
	})

	t.Run("returns error when portfolioClient fails", func(t *testing.T) {
		fakeClient := &fakePortfolioClient{
			getDividendCalendarFn: func(ctx context.Context, in *pb.GetDividendCalendarRequest) (*pb.GetDividendCalendarResponse, error) {
				return nil, errors.New("rpc failed")
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		_, err := resolver.Query().DividendCalendar(ctx, 2026, 5)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestQueryResolver_DividendYieldAnalytics(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully resolves yield analytics", func(t *testing.T) {
		fakeClient := &fakePortfolioClient{
			getDividendYieldAnalyticsFn: func(ctx context.Context, in *pb.GetDividendYieldAnalyticsRequest) (*pb.GetDividendYieldAnalyticsResponse, error) {
				return &pb.GetDividendYieldAnalyticsResponse{
					AnnualForwardIncome:   decimalpb.MoneyToProto(decimal.NewFromInt(1200), "USD"),
					MonthlyAverageIncome:  decimalpb.MoneyToProto(decimal.NewFromInt(100), "USD"),
					PortfolioCurrentYield: decimalpb.ToProto(decimal.NewFromFloat(0.045)),
					PortfolioYieldOnCost:  decimalpb.ToProto(decimal.NewFromFloat(0.060)),
					PendingCashInflow:     decimalpb.MoneyToProto(decimal.NewFromInt(300), "USD"),
					MonthlyProjections: []*pb.MonthlyDividendProjection{
						{
							Year:            2026,
							Month:           5,
							ProjectedIncome: decimalpb.MoneyToProto(decimal.NewFromInt(100), "USD"),
							DividendCount:   2,
						},
					},
					HoldingsYield: []*pb.HoldingYieldItem{
						{
							InstrumentId:        "inst-1",
							Symbol:              "AAPL",
							Name:                "Apple Inc.",
							Quantity:            decimalpb.ToProto(decimal.NewFromInt(50)),
							CurrentPrice:        decimalpb.MoneyToProto(decimal.NewFromInt(200), "USD"),
							AverageBuyPrice:     decimalpb.MoneyToProto(decimal.NewFromInt(150), "USD"),
							CurrentMarketValue:  decimalpb.MoneyToProto(decimal.NewFromInt(10000), "USD"),
							CostBasis:           decimalpb.MoneyToProto(decimal.NewFromInt(7500), "USD"),
							AnnualDps:           decimalpb.MoneyToProto(decimal.NewFromFloat(1.00), "USD"),
							ProjectedAnnualCash: decimalpb.MoneyToProto(decimal.NewFromInt(50), "USD"),
							CurrentYieldPct:     decimalpb.ToProto(decimal.NewFromFloat(0.005)),
							YieldOnCostPct:      decimalpb.ToProto(decimal.NewFromFloat(0.0066)),
							DividendCagr_3Y:     decimalpb.ToProto(decimal.NewFromFloat(0.05)),
							Frequency:           pb.DividendFrequency_DIVIDEND_FREQUENCY_QUARTERLY,
							NextExDate:          "2026-05-10",
							NextPaymentDate:     "2026-05-24",
						},
					},
				}, nil
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		res, err := resolver.Query().DividendYieldAnalytics(ctx)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if res.AnnualForwardIncome.CurrencyCode != "USD" {
			t.Errorf("expected USD currency, got %s", res.AnnualForwardIncome.CurrencyCode)
		}
		if len(res.HoldingsYield) != 1 {
			t.Fatalf("expected 1 holding, got %d", len(res.HoldingsYield))
		}
		if res.HoldingsYield[0].Symbol != "AAPL" {
			t.Errorf("expected AAPL, got %s", res.HoldingsYield[0].Symbol)
		}
	})
}

func TestQueryResolver_SimulateDRIP(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully resolves DRIP simulation", func(t *testing.T) {
		fakeClient := &fakePortfolioClient{
			simulateDRIPFn: func(ctx context.Context, in *pb.SimulateDRIPRequest) (*pb.SimulateDRIPResponse, error) {
				return &pb.SimulateDRIPResponse{
					StartingPortfolioValue: decimalpb.ToProto(decimal.NewFromInt(100000)),
					StartingAnnualDividend: decimalpb.ToProto(decimal.NewFromInt(4000)),
					BaseCurrency:           "USD",
					HorizonYears:           in.GetHorizonYears(),
					Points: []*pb.DRIPSimulationPoint{
						{
							Year:                  1,
							DripPortfolioValue:    decimalpb.ToProto(decimal.NewFromInt(110000)),
							CashPortfolioValue:    decimalpb.ToProto(decimal.NewFromInt(105000)),
							DripAccumulatedShares: decimalpb.ToProto(decimal.NewFromInt(1040)),
							CashAccumulatedCash:   decimalpb.ToProto(decimal.NewFromInt(4000)),
							DripAnnualDividend:    decimalpb.ToProto(decimal.NewFromInt(4400)),
							CashAnnualDividend:    decimalpb.ToProto(decimal.NewFromInt(4200)),
						},
					},
					TerminalDripValue: decimalpb.ToProto(decimal.NewFromInt(110000)),
					TerminalCashValue: decimalpb.ToProto(decimal.NewFromInt(105000)),
					DripBenefitAmount: decimalpb.ToProto(decimal.NewFromInt(5000)),
				}, nil
			},
		}

		resolver := &Resolver{PortfolioClient: fakeClient}
		res, err := resolver.Query().SimulateDrip(ctx, model.DRIPSimulationInput{
			HorizonYears: 10,
		})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if res.HorizonYears != 10 {
			t.Errorf("expected 10 horizon years, got %d", res.HorizonYears)
		}
		if len(res.Points) != 1 {
			t.Fatalf("expected 1 point, got %d", len(res.Points))
		}
	})
}






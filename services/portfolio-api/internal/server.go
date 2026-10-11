package internal

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"portfolio-api/internal/domain"
	"portfolio-api/internal/repository"
	"portfolio-api/internal/service"

	"pkg/decimalpb"

	commonpb "graphfolio/proto/common/v1"
	pb "graphfolio/proto/portfolio/v1"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
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

func (s *PortfolioServer) UpdatePortfolioBaseCurrency(ctx context.Context, req *pb.UpdatePortfolioBaseCurrencyRequest) (*pb.UpdatePortfolioBaseCurrencyResponse, error) {
	if s.svc == nil {
		return nil, status.Error(codes.Unavailable, "service not initialized")
	}

	userID := req.GetUserId()
	if userID == "" {
		userID = "1"
	}

	baseCurrency := req.GetBaseCurrency()
	if baseCurrency == "" {
		return nil, status.Error(codes.InvalidArgument, "base_currency cannot be empty")
	}

	summary, err := s.svc.UpdatePortfolioBaseCurrency(ctx, userID, baseCurrency)
	if err != nil {
		if errors.Is(err, repository.ErrPortfolioNotFound) {
			return nil, status.Errorf(codes.NotFound, "portfolio not found for user: %s", userID)
		}
		return nil, status.Errorf(codes.Internal, "failed to update portfolio base currency: %v", err)
	}

	return &pb.UpdatePortfolioBaseCurrencyResponse{
		Portfolio: mapSummaryToProto(summary),
	}, nil
}

func (s *PortfolioServer) AddTransaction(ctx context.Context, req *pb.AddTransactionRequest) (*pb.AddTransactionResponse, error) {
	if s.svc == nil {
		return nil, status.Error(codes.Unavailable, "service not initialized")
	}

	var txType domain.TransactionType
	switch req.GetType() {
	case pb.TransactionType_TRANSACTION_TYPE_BUY:
		txType = domain.TxTypeBuy
	case pb.TransactionType_TRANSACTION_TYPE_SELL:
		txType = domain.TxTypeSell
	case pb.TransactionType_TRANSACTION_TYPE_DIVIDEND:
		txType = domain.TxTypeDividend
	case pb.TransactionType_TRANSACTION_TYPE_DEPOSIT:
		txType = domain.TxTypeDeposit
	case pb.TransactionType_TRANSACTION_TYPE_WITHDRAWAL:
		txType = domain.TxTypeWithdrawal
	case pb.TransactionType_TRANSACTION_TYPE_INTEREST:
		txType = domain.TxTypeInterest
	case pb.TransactionType_TRANSACTION_TYPE_FEE:
		txType = domain.TxTypeFee
	case pb.TransactionType_TRANSACTION_TYPE_TAX:
		txType = domain.TxTypeTax
	case pb.TransactionType_TRANSACTION_TYPE_TRANSFER_IN:
		txType = domain.TxTypeTransferIn
	case pb.TransactionType_TRANSACTION_TYPE_TRANSFER_OUT:
		txType = domain.TxTypeTransferOut
	case pb.TransactionType_TRANSACTION_TYPE_FX_CONVERSION:
		txType = domain.TxTypeFXConversion
	case pb.TransactionType_TRANSACTION_TYPE_SPLIT:
		txType = domain.TxTypeSplit
	default:
		return nil, status.Errorf(codes.InvalidArgument, "invalid transaction type: %v", req.GetType())
	}

	var tradeDate time.Time
	if req.GetTradeDate() != "" {
		parsedDate, err := time.Parse("2006-01-02", req.GetTradeDate())
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid trade date format (expected YYYY-MM-DD): %v", err)
		}
		tradeDate = parsedDate
	}

	var qty *decimal.Decimal
	if req.GetQuantity() != nil && req.GetQuantity().GetValue() != "" {
		q, err := decimalpb.FromProto(req.GetQuantity())
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid quantity: %v", err)
		}
		qty = &q
	}

	var price *decimal.Decimal
	if req.GetPrice() != nil && req.GetPrice().GetAmount() != nil {
		p, _, err := decimalpb.MoneyFromProto(req.GetPrice())
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid price: %v", err)
		}
		price = &p
	}

	var amount *decimal.Decimal
	var ccy *string
	if req.GetAmount() != nil && req.GetAmount().GetAmount() != nil {
		a, c, err := decimalpb.MoneyFromProto(req.GetAmount())
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid amount: %v", err)
		}
		amount = &a
		if c != "" {
			ccy = &c
		}
	}

	var fee *decimal.Decimal
	var feeCurrency *string
	if req.GetFee() != nil && req.GetFee().GetAmount() != nil {
		f, c, err := decimalpb.MoneyFromProto(req.GetFee())
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid fee: %v", err)
		}
		fee = &f
		if c != "" {
			feeCurrency = &c
		}
	}

	var symbol *string
	if req.GetSymbol() != "" {
		s := req.GetSymbol()
		symbol = &s
	}

	var notes *string
	if req.GetNotes() != "" {
		n := req.GetNotes()
		notes = &n
	}

	input := domain.AddTransactionInput{
		UserID:          req.GetUserId(),
		Type:            txType,
		Symbol:          symbol,
		TradeDate:       tradeDate,
		Quantity:        qty,
		Price:           price,
		Amount:          amount,
		CurrencyCode:    ccy,
		Fee:             fee,
		FeeCurrencyCode: feeCurrency,
		Notes:           notes,
	}

	tx, summary, err := s.svc.AddTransaction(ctx, input)
	if err != nil {
		if errors.Is(err, repository.ErrPortfolioNotFound) {
			return nil, status.Errorf(codes.NotFound, "portfolio not found: %v", err)
		}
		if errors.Is(err, repository.ErrInstrumentNotFound) {
			return nil, status.Errorf(codes.NotFound, "instrument not found: %v", err)
		}
		return nil, status.Errorf(codes.InvalidArgument, "failed to add transaction: %v", err)
	}

	return &pb.AddTransactionResponse{
		TransactionId: tx.ID.String(),
		Portfolio:     mapSummaryToProto(summary),
	}, nil
}

func (s *PortfolioServer) ListInstruments(ctx context.Context, req *pb.ListInstrumentsRequest) (*pb.ListInstrumentsResponse, error) {
	if s.svc == nil {
		return nil, status.Error(codes.Unavailable, "service not initialized")
	}

	instruments, err := s.svc.ListInstruments(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list instruments: %v", err)
	}

	protoInsts := make([]*pb.Instrument, len(instruments))
	for i, inst := range instruments {
		protoInsts[i] = mapInstrumentToProto(inst)
	}

	return &pb.ListInstrumentsResponse{
		Instruments: protoInsts,
	}, nil
}

func (s *PortfolioServer) GetPortfolioHistory(ctx context.Context, req *pb.GetPortfolioHistoryRequest) (*pb.GetPortfolioHistoryResponse, error) {
	if s.svc == nil {
		return nil, status.Error(codes.Unavailable, "service not initialized")
	}

	userID := req.GetUserId()
	if userID == "" {
		userID = "1"
	}

	var timeframe domain.HistoryTimeframe
	switch req.GetTimeframe() {
	case pb.HistoryTimeframe_HISTORY_TIMEFRAME_1D:
		timeframe = domain.Timeframe1D
	case pb.HistoryTimeframe_HISTORY_TIMEFRAME_1W:
		timeframe = domain.Timeframe1W
	case pb.HistoryTimeframe_HISTORY_TIMEFRAME_1M:
		timeframe = domain.Timeframe1M
	case pb.HistoryTimeframe_HISTORY_TIMEFRAME_1Y:
		timeframe = domain.Timeframe1Y
	case pb.HistoryTimeframe_HISTORY_TIMEFRAME_ALL, pb.HistoryTimeframe_HISTORY_TIMEFRAME_UNSPECIFIED:
		timeframe = domain.TimeframeAll
	default:
		return nil, status.Errorf(codes.InvalidArgument, "invalid timeframe: %v", req.GetTimeframe())
	}

	history, err := s.svc.GetPortfolioHistory(ctx, userID, timeframe)
	if err != nil {
		if errors.Is(err, repository.ErrPortfolioNotFound) {
			return nil, status.Errorf(codes.NotFound, "portfolio not found for user: %s", userID)
		}
		return nil, status.Errorf(codes.Internal, "failed to get portfolio history: %v", err)
	}

	protoPoints := make([]*pb.ValuationPoint, len(history.Points))
	for i, pt := range history.Points {
		protoPoints[i] = &pb.ValuationPoint{
			Date:        pt.Date.Format("2006-01-02"),
			TotalValue:  decimalpb.MoneyToProto(pt.TotalValue.Amount, pt.TotalValue.CurrencyCode),
			MarketValue: decimalpb.MoneyToProto(pt.MarketValue.Amount, pt.MarketValue.CurrencyCode),
			CashValue:   decimalpb.MoneyToProto(pt.CashValue.Amount, pt.CashValue.CurrencyCode),
			TwrIndex:    decimalpb.ToProto(pt.TWRIndex),
			DailyReturn: decimalpb.ToProto(pt.DailyReturn),
		}
	}

	return &pb.GetPortfolioHistoryResponse{
		Points:        protoPoints,
		StartValue:    decimalpb.MoneyToProto(history.StartValue.Amount, history.StartValue.CurrencyCode),
		EndValue:      decimalpb.MoneyToProto(history.EndValue.Amount, history.EndValue.CurrencyCode),
		ReturnAmount:  decimalpb.MoneyToProto(history.ReturnAmount.Amount, history.ReturnAmount.CurrencyCode),
		ReturnPercent: decimalpb.ToProto(history.ReturnPercent),
	}, nil
}

func (s *PortfolioServer) ListTransactions(ctx context.Context, req *pb.ListTransactionsRequest) (*pb.ListTransactionsResponse, error) {
	if s.svc == nil {
		return nil, status.Error(codes.Unavailable, "service not initialized")
	}

	userID := req.GetUserId()
	if userID == "" {
		userID = "1"
	}

	if req.GetPage() < 0 {
		return nil, status.Errorf(codes.InvalidArgument, "page cannot be negative")
	}
	if req.GetPageSize() < 0 || req.GetPageSize() > 100 {
		return nil, status.Errorf(codes.InvalidArgument, "page size must be between 0 and 100")
	}

	page := int(req.GetPage())
	if page == 0 {
		page = 1
	}
	pageSize := int(req.GetPageSize())
	if pageSize == 0 {
		pageSize = 20
	}

	var typeFilter *domain.TransactionType
	if req.GetType() != pb.TransactionType_TRANSACTION_TYPE_UNSPECIFIED {
		var t domain.TransactionType
		switch req.GetType() {
		case pb.TransactionType_TRANSACTION_TYPE_BUY:
			t = domain.TxTypeBuy
		case pb.TransactionType_TRANSACTION_TYPE_SELL:
			t = domain.TxTypeSell
		case pb.TransactionType_TRANSACTION_TYPE_DIVIDEND:
			t = domain.TxTypeDividend
		case pb.TransactionType_TRANSACTION_TYPE_DEPOSIT:
			t = domain.TxTypeDeposit
		case pb.TransactionType_TRANSACTION_TYPE_WITHDRAWAL:
			t = domain.TxTypeWithdrawal
		case pb.TransactionType_TRANSACTION_TYPE_INTEREST:
			t = domain.TxTypeInterest
		case pb.TransactionType_TRANSACTION_TYPE_FEE:
			t = domain.TxTypeFee
		case pb.TransactionType_TRANSACTION_TYPE_TAX:
			t = domain.TxTypeTax
		case pb.TransactionType_TRANSACTION_TYPE_TRANSFER_IN:
			t = domain.TxTypeTransferIn
		case pb.TransactionType_TRANSACTION_TYPE_TRANSFER_OUT:
			t = domain.TxTypeTransferOut
		case pb.TransactionType_TRANSACTION_TYPE_FX_CONVERSION:
			t = domain.TxTypeFXConversion
		case pb.TransactionType_TRANSACTION_TYPE_SPLIT:
			t = domain.TxTypeSplit
		default:
			return nil, status.Errorf(codes.InvalidArgument, "invalid transaction type filter: %v", req.GetType())
		}
		typeFilter = &t
	}

	var symbolFilter *string
	if req.GetSymbol() != "" {
		sym := req.GetSymbol()
		symbolFilter = &sym
	}

	filter := domain.TransactionFilter{
		Type:     typeFilter,
		Symbol:   symbolFilter,
		Page:     page,
		PageSize: pageSize,
	}

	items, totalCount, err := s.svc.ListTransactions(ctx, userID, filter)
	if err != nil {
		if errors.Is(err, repository.ErrPortfolioNotFound) {
			return nil, status.Errorf(codes.NotFound, "portfolio not found for user: %s", userID)
		}
		return nil, status.Errorf(codes.Internal, "failed to list transactions: %v", err)
	}

	protoItems := make([]*pb.TransactionItem, len(items))
	for i, item := range items {
		feeCurr := item.CurrencyCode
		if item.FeeCurrencyCode != nil && *item.FeeCurrencyCode != "" {
			feeCurr = *item.FeeCurrencyCode
		}
		protoItem := &pb.TransactionItem{
			Id:        item.ID.String(),
			Type:      mapDomainTxTypeToProto(item.Type),
			TradeDate: item.TradeDate.Format("2006-01-02"),
			Amount:    decimalpb.MoneyToProto(item.Amount, item.CurrencyCode),
			Fee:       decimalpb.MoneyToProto(item.Fee, feeCurr),
			CreatedAt: "",
		}
		if !item.CreatedAt.IsZero() {
			protoItem.CreatedAt = item.CreatedAt.Format(time.RFC3339)
		}
		if item.Symbol != nil {
			protoItem.Symbol = *item.Symbol
		}
		if item.InstrumentName != nil {
			protoItem.InstrumentName = *item.InstrumentName
		}
		if item.Notes != nil {
			protoItem.Notes = *item.Notes
		}
		if item.Quantity != nil {
			protoItem.Quantity = decimalpb.ToProto(*item.Quantity)
		}
		if item.Price != nil {
			protoItem.Price = decimalpb.MoneyToProto(*item.Price, item.CurrencyCode)
		}
		protoItems[i] = protoItem
	}

	return &pb.ListTransactionsResponse{
		Transactions: protoItems,
		TotalCount:   int32(totalCount),
		Page:         int32(page),
		PageSize:     int32(pageSize),
	}, nil
}

func (s *PortfolioServer) DeleteTransaction(ctx context.Context, req *pb.DeleteTransactionRequest) (*pb.DeleteTransactionResponse, error) {
	if s.svc == nil {
		return nil, status.Error(codes.Unavailable, "service not initialized")
	}

	if req.GetTransactionId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "transaction id is required")
	}

	userID := req.GetUserId()
	if userID == "" {
		userID = "1"
	}

	summary, err := s.svc.DeleteTransaction(ctx, userID, req.GetTransactionId())
	if err != nil {
		if errors.Is(err, repository.ErrTransactionNotFound) || errors.Is(err, repository.ErrPortfolioNotFound) {
			return nil, status.Errorf(codes.NotFound, "transaction not found: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to delete transaction: %v", err)
	}

	return &pb.DeleteTransactionResponse{
		Success:   true,
		Portfolio: mapSummaryToProto(summary),
	}, nil
}

func (s *PortfolioServer) CheckTransactionDuplicates(ctx context.Context, req *pb.CheckTransactionDuplicatesRequest) (*pb.CheckTransactionDuplicatesResponse, error) {
	if s.svc == nil {
		return nil, status.Error(codes.Unavailable, "service not initialized")
	}

	userID := req.GetUserId()
	if userID == "" {
		userID = "1"
	}

	existingRefs, err := s.svc.CheckTransactionDuplicates(ctx, userID, req.GetExternalRefs())
	if err != nil {
		if errors.Is(err, repository.ErrPortfolioNotFound) {
			return nil, status.Errorf(codes.NotFound, "portfolio not found: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to check duplicates: %v", err)
	}

	return &pb.CheckTransactionDuplicatesResponse{
		ExistingExternalRefs: existingRefs,
	}, nil
}

func (s *PortfolioServer) BatchImportTransactions(ctx context.Context, req *pb.BatchImportTransactionsRequest) (*pb.BatchImportTransactionsResponse, error) {
	if s.svc == nil {
		return nil, status.Error(codes.Unavailable, "service not initialized")
	}

	userID := req.GetUserId()
	if userID == "" {
		userID = "1"
	}

	var items []domain.ImportTransactionItem
	for i, item := range req.GetTransactions() {
		var txType domain.TransactionType
		switch item.GetType() {
		case pb.TransactionType_TRANSACTION_TYPE_BUY:
			txType = domain.TxTypeBuy
		case pb.TransactionType_TRANSACTION_TYPE_SELL:
			txType = domain.TxTypeSell
		case pb.TransactionType_TRANSACTION_TYPE_DIVIDEND:
			txType = domain.TxTypeDividend
		case pb.TransactionType_TRANSACTION_TYPE_DEPOSIT:
			txType = domain.TxTypeDeposit
		case pb.TransactionType_TRANSACTION_TYPE_WITHDRAWAL:
			txType = domain.TxTypeWithdrawal
		case pb.TransactionType_TRANSACTION_TYPE_INTEREST:
			txType = domain.TxTypeInterest
		case pb.TransactionType_TRANSACTION_TYPE_FEE:
			txType = domain.TxTypeFee
		case pb.TransactionType_TRANSACTION_TYPE_TAX:
			txType = domain.TxTypeTax
		case pb.TransactionType_TRANSACTION_TYPE_TRANSFER_IN:
			txType = domain.TxTypeTransferIn
		case pb.TransactionType_TRANSACTION_TYPE_TRANSFER_OUT:
			txType = domain.TxTypeTransferOut
		case pb.TransactionType_TRANSACTION_TYPE_FX_CONVERSION:
			txType = domain.TxTypeFXConversion
		case pb.TransactionType_TRANSACTION_TYPE_SPLIT:
			txType = domain.TxTypeSplit
		default:
			return nil, status.Errorf(codes.InvalidArgument, "item %d: invalid transaction type: %v", i, item.GetType())
		}

		tradeDate, err := time.Parse("2006-01-02", item.GetTradeDate())
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "item %d: invalid trade date %q: %v", i, item.GetTradeDate(), err)
		}

		var settleDate *time.Time
		if item.GetSettleDate() != "" {
			sd, err := time.Parse("2006-01-02", item.GetSettleDate())
			if err == nil {
				settleDate = &sd
			}
		}

		var qty decimal.Decimal
		if item.GetQuantity() != nil {
			q, err := decimalpb.FromProto(item.GetQuantity())
			if err == nil {
				qty = q
			}
		}

		var price decimal.Decimal
		if item.GetPrice() != nil && item.GetPrice().GetAmount() != nil {
			p, _, err := decimalpb.MoneyFromProto(item.GetPrice())
			if err == nil {
				price = p
			}
		}

		var amount decimal.Decimal
		if item.GetAmount() != nil && item.GetAmount().GetAmount() != nil {
			a, _, err := decimalpb.MoneyFromProto(item.GetAmount())
			if err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "item %d: invalid amount: %v", i, err)
			}
			amount = a
		}

		var fee decimal.Decimal
		var feeCurrency *string
		if item.GetFee() != nil && item.GetFee().GetAmount() != nil {
			f, c, err := decimalpb.MoneyFromProto(item.GetFee())
			if err == nil {
				fee = f
				if c != "" {
					feeCurrency = &c
				}
			}
		}

		var currencyCode string
		if item.GetAmount() != nil && item.GetAmount().GetCurrencyCode() != "" {
			currencyCode = item.GetAmount().GetCurrencyCode()
		} else if item.GetPrice() != nil && item.GetPrice().GetCurrencyCode() != "" {
			currencyCode = item.GetPrice().GetCurrencyCode()
		}

		items = append(items, domain.ImportTransactionItem{
			Type:            txType,
			Symbol:          item.GetSymbol(),
			TradeDate:       tradeDate,
			SettleDate:      settleDate,
			Quantity:        qty,
			Price:           price,
			Amount:          amount,
			Fee:             fee,
			FeeCurrencyCode: feeCurrency,
			CurrencyCode:    currencyCode,
			ExternalRef:     item.GetExternalRef(),
			Notes:           item.GetNotes(),
		})
	}

	result, err := s.svc.BatchImportTransactions(ctx, domain.BatchImportInput{
		UserID:         userID,
		Transactions:   items,
		SkipDuplicates: req.GetSkipDuplicates(),
	})
	if err != nil {
		if errors.Is(err, repository.ErrPortfolioNotFound) {
			return nil, status.Errorf(codes.NotFound, "portfolio not found: %v", err)
		}
		return nil, status.Errorf(codes.InvalidArgument, "batch import failed: %v", err)
	}

	return &pb.BatchImportTransactionsResponse{
		Success:       true,
		ImportedCount: int32(result.ImportedCount),
		SkippedCount:  int32(result.SkippedCount),
		Portfolio:     mapSummaryToProto(result.Portfolio),
		Message:       fmt.Sprintf("Successfully imported %d transactions (skipped %d)", result.ImportedCount, result.SkippedCount),
	}, nil
}

// ------------------------------------------------------------------
// Admin RPC Handlers
// ------------------------------------------------------------------

func (s *PortfolioServer) ListAllInstruments(ctx context.Context, req *pb.ListAllInstrumentsRequest) (*pb.ListAllInstrumentsResponse, error) {
	if s.svc == nil {
		return nil, status.Error(codes.Unavailable, "service not initialized")
	}

	var isActive *bool
	if req.IsActive != nil {
		isActive = req.IsActive
	}

	var search *string
	if req.Search != nil && *req.Search != "" {
		search = req.Search
	}

	insts, err := s.svc.ListAllInstruments(ctx, isActive, search)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list all instruments: %v", err)
	}

	protoInsts := make([]*pb.Instrument, len(insts))
	for i, inst := range insts {
		protoInsts[i] = mapInstrumentToProto(inst)
	}

	return &pb.ListAllInstrumentsResponse{
		Instruments: protoInsts,
	}, nil
}

func (s *PortfolioServer) ListExchanges(ctx context.Context, _ *pb.ListExchangesRequest) (*pb.ListExchangesResponse, error) {
	if s.svc == nil {
		return nil, status.Error(codes.Unavailable, "service not initialized")
	}

	exchanges, err := s.svc.ListExchanges(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list exchanges: %v", err)
	}

	protoExchanges := make([]*pb.Exchange, len(exchanges))
	for i, ex := range exchanges {
		protoExchanges[i] = &pb.Exchange{
			Code:     ex.Code,
			Name:     ex.Name,
			Country:  ex.Country,
			Timezone: ex.Timezone,
		}
	}

	return &pb.ListExchangesResponse{
		Exchanges: protoExchanges,
	}, nil
}

func (s *PortfolioServer) CreateInstrument(ctx context.Context, req *pb.CreateInstrumentRequest) (*pb.CreateInstrumentResponse, error) {
	if s.svc == nil {
		return nil, status.Error(codes.Unavailable, "service not initialized")
	}

	var isin *string
	if req.Isin != nil && *req.Isin != "" {
		isin = req.Isin
	}

	input := domain.CreateInstrumentInput{
		Symbol:       req.GetSymbol(),
		ExchangeCode: req.GetExchangeCode(),
		Name:         req.GetName(),
		AssetClass:   req.GetAssetClass(),
		CurrencyCode: req.GetCurrencyCode(),
		ISIN:         isin,
	}

	inst, err := s.svc.CreateInstrument(ctx, input)
	if err != nil {
		if errors.Is(err, repository.ErrInstrumentConflict) {
			return nil, status.Errorf(codes.AlreadyExists, "instrument %s on %s already exists", req.GetSymbol(), req.GetExchangeCode())
		}
		if errors.Is(err, service.ErrInvalidSymbol) ||
			errors.Is(err, service.ErrInvalidExchange) ||
			errors.Is(err, service.ErrInvalidName) ||
			errors.Is(err, service.ErrInvalidAssetClass) ||
			errors.Is(err, service.ErrInvalidCurrency) ||
			errors.Is(err, service.ErrInvalidISIN) {
			return nil, status.Errorf(codes.InvalidArgument, "%v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to create instrument: %v", err)
	}

	return &pb.CreateInstrumentResponse{
		Instrument: mapInstrumentToProto(*inst),
	}, nil
}

func (s *PortfolioServer) UpdateInstrument(ctx context.Context, req *pb.UpdateInstrumentRequest) (*pb.UpdateInstrumentResponse, error) {
	if s.svc == nil {
		return nil, status.Error(codes.Unavailable, "service not initialized")
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid instrument id: %v", err)
	}

	inst, err := s.svc.UpdateInstrument(ctx, domain.UpdateInstrumentInput{
		ID:       id,
		Name:     req.Name,
		IsActive: req.IsActive,
		ISIN:     req.Isin,
	})
	if err != nil {
		if errors.Is(err, repository.ErrInstrumentNotFound) {
			return nil, status.Errorf(codes.NotFound, "instrument not found: %s", req.GetId())
		}
		if errors.Is(err, service.ErrInvalidName) || errors.Is(err, service.ErrInvalidISIN) {
			return nil, status.Errorf(codes.InvalidArgument, "%v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to update instrument: %v", err)
	}

	return &pb.UpdateInstrumentResponse{
		Instrument: mapInstrumentToProto(*inst),
	}, nil
}

func (s *PortfolioServer) DeleteInstrument(ctx context.Context, req *pb.DeleteInstrumentRequest) (*pb.DeleteInstrumentResponse, error) {
	if s.svc == nil {
		return nil, status.Error(codes.Unavailable, "service not initialized")
	}

	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "instrument id cannot be empty")
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid instrument id: %v", err)
	}

	err = s.svc.DeleteInstrument(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrInstrumentNotFound) {
			return nil, status.Errorf(codes.NotFound, "instrument %s not found", req.GetId())
		}
		if errors.Is(err, repository.ErrInstrumentInUse) {
			return nil, status.Errorf(codes.FailedPrecondition, "cannot delete instrument: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to delete instrument: %v", err)
	}

	return &pb.DeleteInstrumentResponse{
		Success: true,
		Id:      req.GetId(),
	}, nil
}

func (s *PortfolioServer) ListInstrumentPrices(ctx context.Context, req *pb.ListInstrumentPricesRequest) (*pb.ListInstrumentPricesResponse, error) {
	if s.svc == nil {
		return nil, status.Error(codes.Unavailable, "service not initialized")
	}

	var symbol *string
	if req.Symbol != nil && *req.Symbol != "" {
		symbol = req.Symbol
	}

	var fromDate *time.Time
	if req.FromDate != nil && *req.FromDate != "" {
		d, err := time.Parse("2006-01-02", *req.FromDate)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid from_date format (must be YYYY-MM-DD): %v", err)
		}
		fromDate = &d
	}

	var toDate *time.Time
	if req.ToDate != nil && *req.ToDate != "" {
		d, err := time.Parse("2006-01-02", *req.ToDate)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid to_date format (must be YYYY-MM-DD): %v", err)
		}
		toDate = &d
	}

	prices, total, err := s.svc.ListInstrumentPrices(ctx, domain.PriceFilter{
		Symbol:   symbol,
		FromDate: fromDate,
		ToDate:   toDate,
		Limit:    int(req.GetLimit()),
		Offset:   int(req.GetOffset()),
	})
	if err != nil {
		if errors.Is(err, service.ErrInvalidDateRange) {
			return nil, status.Errorf(codes.InvalidArgument, "%v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to list instrument prices: %v", err)
	}

	protoPrices := make([]*pb.InstrumentPriceItem, len(prices))
	for i, p := range prices {
		protoPrices[i] = mapInstrumentPriceToProto(p)
	}

	return &pb.ListInstrumentPricesResponse{
		Prices:     protoPrices,
		TotalCount: int32(total),
	}, nil
}

func (s *PortfolioServer) RecordPriceOverride(ctx context.Context, req *pb.RecordPriceOverrideRequest) (*pb.RecordPriceOverrideResponse, error) {
	if s.svc == nil {
		return nil, status.Error(codes.Unavailable, "service not initialized")
	}

	if req.GetSymbol() == "" {
		return nil, status.Error(codes.InvalidArgument, "symbol is required")
	}

	if req.GetPriceDate() == "" {
		return nil, status.Error(codes.InvalidArgument, "price_date is required")
	}

	priceDate, err := time.Parse("2006-01-02", req.GetPriceDate())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid price_date format (must be YYYY-MM-DD): %v", err)
	}

	if req.GetPrice() == nil {
		return nil, status.Error(codes.InvalidArgument, "price is required")
	}

	priceDec, err := decimalpb.FromProto(req.GetPrice())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid price format: %v", err)
	}

	var reason *string
	if req.Reason != nil && *req.Reason != "" {
		reason = req.Reason
	}

	price, recomputed, err := s.svc.RecordPriceOverride(ctx, domain.PriceOverrideInput{
		Symbol:              req.GetSymbol(),
		PriceDate:           priceDate,
		Price:               priceDec,
		Reason:              reason,
		RecomputeValuations: req.GetRecomputeValuations(),
	})
	if err != nil {
		if errors.Is(err, repository.ErrInstrumentNotFound) {
			return nil, status.Errorf(codes.NotFound, "instrument not found: %s", req.GetSymbol())
		}
		if errors.Is(err, service.ErrInvalidPrice) || errors.Is(err, service.ErrFutureDate) || errors.Is(err, service.ErrInvalidSymbol) {
			return nil, status.Errorf(codes.InvalidArgument, "%v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to record price override: %v", err)
	}

	return &pb.RecordPriceOverrideResponse{
		Price:                mapInstrumentPriceToProto(*price),
		ValuationsRecomputed: recomputed,
	}, nil
}

func (s *PortfolioServer) GetIngestionStatus(ctx context.Context, req *pb.GetIngestionStatusRequest) (*pb.GetIngestionStatusResponse, error) {
	if s.svc == nil {
		return nil, status.Error(codes.Unavailable, "service not initialized")
	}

	statusMetrics, err := s.svc.GetIngestionStatus(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get ingestion status: %v", err)
	}

	feeds := make([]*pb.FeedHealthStatus, len(statusMetrics.Feeds))
	for i, f := range statusMetrics.Feeds {
		feeds[i] = &pb.FeedHealthStatus{
			Name:     f.Name,
			Status:   f.Status,
			Provider: f.Provider,
			Schedule: f.Schedule,
			LastRun:  f.LastRun.Format(time.RFC3339),
			Details:  f.Details,
		}
	}

	resp := &pb.GetIngestionStatusResponse{
		Feeds:               feeds,
		TrackedInstruments:  int32(statusMetrics.TrackedInstruments),
		TrackedCurrencies:   int32(statusMetrics.TrackedCurrencies),
		RateLimitRemaining:  int32(statusMetrics.RateLimitRemaining),
		RateLimitBudget:     int32(statusMetrics.RateLimitBudget),
		PendingBackfillJobs: int32(statusMetrics.PendingBackfillJobs),
	}
	if statusMetrics.LatestPriceDate != nil {
		resp.LatestPriceDate = statusMetrics.LatestPriceDate.Format("2006-01-02")
	}
	if statusMetrics.LatestFXDate != nil {
		resp.LatestFxDate = statusMetrics.LatestFXDate.Format("2006-01-02")
	}

	return resp, nil
}

func (s *PortfolioServer) TriggerMarketSync(ctx context.Context, req *pb.TriggerMarketSyncRequest) (*pb.TriggerMarketSyncResponse, error) {
	if s.svc == nil {
		return nil, status.Error(codes.Unavailable, "service not initialized")
	}

	res, err := s.svc.TriggerMarketSync(ctx, req.GetSymbols(), req.GetSyncFx())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to trigger market sync: %v", err)
	}

	return &pb.TriggerMarketSyncResponse{
		Success:       res.Success,
		PricesSynced:  int32(res.PricesSynced),
		FxRatesSynced: int32(res.FXRatesSynced),
		Message:       res.Message,
	}, nil
}

func (s *PortfolioServer) RebuildValuations(ctx context.Context, req *pb.RebuildValuationsRequest) (*pb.RebuildValuationsResponse, error) {
	if s.svc == nil {
		return nil, status.Error(codes.Unavailable, "service not initialized")
	}

	userID := req.GetUserId()
	if userID == "" {
		userID = "1"
	}

	var fromDate *time.Time
	if req.GetFromDate() != "" {
		parsed, err := time.Parse("2006-01-02", req.GetFromDate())
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid from_date format: %v (expected YYYY-MM-DD)", err)
		}
		utcDate := parsed.UTC()
		fromDate = &utcDate
	}

	if err := s.svc.RebuildValuations(ctx, userID, fromDate); err != nil {
		if errors.Is(err, repository.ErrPortfolioNotFound) {
			return nil, status.Errorf(codes.NotFound, "portfolio not found for user: %s", userID)
		}
		return nil, status.Errorf(codes.Internal, "failed to rebuild valuations: %v", err)
	}

	return &pb.RebuildValuationsResponse{
		Success: true,
		Message: "Valuations recomputed successfully",
	}, nil
}

func (s *PortfolioServer) ListCurrencyPairs(ctx context.Context, req *pb.ListCurrencyPairsRequest) (*pb.ListCurrencyPairsResponse, error) {
	if s.svc == nil {
		return nil, status.Error(codes.Unavailable, "service not initialized")
	}

	pairs, err := s.svc.ListCurrencyPairs(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list currency pairs: %v", err)
	}

	protoPairs := make([]*pb.CurrencyPairItem, len(pairs))
	for i, p := range pairs {
		item := &pb.CurrencyPairItem{
			BaseCurrency:  p.BaseCurrency,
			QuoteCurrency: p.QuoteCurrency,
			LatestRate:    decimalpb.ToProto(p.LatestRate),
			LatestDate:    p.LatestDate.Format("2006-01-02"),
			LatestSource:  p.LatestSource,
			TotalRecords:  int32(p.TotalRecords),
			FirstDate:     p.FirstDate.Format("2006-01-02"),
			LastDate:      p.LastDate.Format("2006-01-02"),
		}
		if p.PreviousRate != nil {
			item.PreviousRate = decimalpb.ToProto(*p.PreviousRate)
		}
		if p.Change1DAmount != nil {
			item.Change_1DAmount = decimalpb.ToProto(*p.Change1DAmount)
		}
		if p.Change1DPct != nil {
			item.Change_1DPct = decimalpb.ToProto(*p.Change1DPct)
		}
		protoPairs[i] = item
	}

	return &pb.ListCurrencyPairsResponse{
		Pairs: protoPairs,
	}, nil
}

func (s *PortfolioServer) ListFXRates(ctx context.Context, req *pb.ListFXRatesRequest) (*pb.ListFXRatesResponse, error) {
	if s.svc == nil {
		return nil, status.Error(codes.Unavailable, "service not initialized")
	}

	var fromDate, toDate *time.Time
	if req.FromDate != nil && *req.FromDate != "" {
		fd, err := time.Parse("2006-01-02", *req.FromDate)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid from_date format: %v", err)
		}
		fromDate = &fd
	}
	if req.ToDate != nil && *req.ToDate != "" {
		td, err := time.Parse("2006-01-02", *req.ToDate)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid to_date format: %v", err)
		}
		toDate = &td
	}

	rates, total, err := s.svc.ListFXRates(ctx, domain.FXRateFilter{
		BaseCurrency:  req.BaseCurrency,
		QuoteCurrency: req.QuoteCurrency,
		FromDate:      fromDate,
		ToDate:        toDate,
		Limit:         int(req.GetLimit()),
		Offset:        int(req.GetOffset()),
	})
	if err != nil {
		if errors.Is(err, service.ErrInvalidCurrency) || errors.Is(err, service.ErrEqualCurrencies) || errors.Is(err, service.ErrInvalidDateRange) {
			return nil, status.Errorf(codes.InvalidArgument, "%v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to list fx rates: %v", err)
	}

	protoRates := make([]*pb.FXRateItem, len(rates))
	for i, r := range rates {
		protoRates[i] = mapFXRateToProto(r)
	}

	return &pb.ListFXRatesResponse{
		Rates:      protoRates,
		TotalCount: int32(total),
	}, nil
}

func (s *PortfolioServer) GetCurrencyPairHistory(ctx context.Context, req *pb.GetCurrencyPairHistoryRequest) (*pb.GetCurrencyPairHistoryResponse, error) {
	if s.svc == nil {
		return nil, status.Error(codes.Unavailable, "service not initialized")
	}

	if req.GetBaseCurrency() == "" {
		return nil, status.Error(codes.InvalidArgument, "base_currency is required")
	}
	if req.GetQuoteCurrency() == "" {
		return nil, status.Error(codes.InvalidArgument, "quote_currency is required")
	}

	var timeframe domain.HistoryTimeframe
	switch req.GetTimeframe() {
	case pb.HistoryTimeframe_HISTORY_TIMEFRAME_1D:
		timeframe = domain.Timeframe1D
	case pb.HistoryTimeframe_HISTORY_TIMEFRAME_1W:
		timeframe = domain.Timeframe1W
	case pb.HistoryTimeframe_HISTORY_TIMEFRAME_1M:
		timeframe = domain.Timeframe1M
	case pb.HistoryTimeframe_HISTORY_TIMEFRAME_1Y:
		timeframe = domain.Timeframe1Y
	case pb.HistoryTimeframe_HISTORY_TIMEFRAME_ALL, pb.HistoryTimeframe_HISTORY_TIMEFRAME_UNSPECIFIED:
		timeframe = domain.TimeframeAll
	default:
		return nil, status.Errorf(codes.InvalidArgument, "invalid timeframe: %v", req.GetTimeframe())
	}

	hist, err := s.svc.GetCurrencyPairHistory(ctx, req.GetBaseCurrency(), req.GetQuoteCurrency(), timeframe)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCurrency) || errors.Is(err, service.ErrEqualCurrencies) {
			return nil, status.Errorf(codes.InvalidArgument, "%v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to get currency pair history: %v", err)
	}

	protoPoints := make([]*pb.FXHistoryPoint, len(hist.Points))
	for i, pt := range hist.Points {
		protoPoints[i] = &pb.FXHistoryPoint{
			Date:         pt.Date.Format("2006-01-02"),
			Rate:         decimalpb.ToProto(pt.Rate),
			InvertedRate: decimalpb.ToProto(pt.InvertedRate),
			Source:       pt.Source,
		}
	}

	return &pb.GetCurrencyPairHistoryResponse{
		BaseCurrency:    hist.BaseCurrency,
		QuoteCurrency:   hist.QuoteCurrency,
		Points:          protoPoints,
		StartRate:       decimalpb.ToProto(hist.StartRate),
		EndRate:         decimalpb.ToProto(hist.EndRate),
		PeriodChange:    decimalpb.ToProto(hist.PeriodChange),
		PeriodChangePct: decimalpb.ToProto(hist.PeriodChangePct),
		PeriodHigh:      decimalpb.ToProto(hist.PeriodHigh),
		PeriodLow:       decimalpb.ToProto(hist.PeriodLow),
	}, nil
}

func (s *PortfolioServer) RecordFXRateOverride(ctx context.Context, req *pb.RecordFXRateOverrideRequest) (*pb.RecordFXRateOverrideResponse, error) {
	if s.svc == nil {
		return nil, status.Error(codes.Unavailable, "service not initialized")
	}

	if req.GetBaseCurrency() == "" {
		return nil, status.Error(codes.InvalidArgument, "base_currency is required")
	}
	if req.GetQuoteCurrency() == "" {
		return nil, status.Error(codes.InvalidArgument, "quote_currency is required")
	}
	if req.GetRateDate() == "" {
		return nil, status.Error(codes.InvalidArgument, "rate_date is required")
	}

	rateDate, err := time.Parse("2006-01-02", req.GetRateDate())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid rate_date format (must be YYYY-MM-DD): %v", err)
	}

	if req.GetRate() == nil {
		return nil, status.Error(codes.InvalidArgument, "rate is required")
	}

	rateDec, err := decimalpb.FromProto(req.GetRate())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid rate format: %v", err)
	}

	var reason *string
	if req.Reason != nil && strings.TrimSpace(*req.Reason) != "" {
		reason = req.Reason
	}

	fxRate, recomputed, err := s.svc.RecordFXRateOverride(ctx, domain.FXRateOverrideInput{
		BaseCurrency:        req.GetBaseCurrency(),
		QuoteCurrency:       req.GetQuoteCurrency(),
		RateDate:            rateDate,
		Rate:                rateDec,
		Reason:              reason,
		RecomputeValuations: req.GetRecomputeValuations(),
	})
	if err != nil {
		if errors.Is(err, service.ErrInvalidCurrency) ||
			errors.Is(err, service.ErrEqualCurrencies) ||
			errors.Is(err, service.ErrRateDateRequired) ||
			errors.Is(err, service.ErrFutureDate) ||
			errors.Is(err, service.ErrInvalidRate) {
			return nil, status.Errorf(codes.InvalidArgument, "%v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to record fx rate override: %v", err)
	}

	return &pb.RecordFXRateOverrideResponse{
		Rate:                 mapFXRateToProto(*fxRate),
		ValuationsRecomputed: recomputed,
	}, nil
}

func (s *PortfolioServer) TriggerBackfill(ctx context.Context, req *pb.TriggerBackfillRequest) (*pb.TriggerBackfillResponse, error) {
	if s.svc == nil {
		return nil, status.Error(codes.Unavailable, "service not initialized")
	}

	if req.GetFromDate() == "" {
		return nil, status.Error(codes.InvalidArgument, "from_date is required (YYYY-MM-DD)")
	}

	fromDate, err := time.Parse("2006-01-02", req.GetFromDate())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid from_date format (must be YYYY-MM-DD): %v", err)
	}

	toDate := time.Now().UTC()
	if req.GetToDate() != "" {
		toDate, err = time.Parse("2006-01-02", req.GetToDate())
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid to_date format (must be YYYY-MM-DD): %v", err)
		}
	}

	input := domain.BackfillInput{
		FromDate:            fromDate,
		ToDate:              toDate,
		Symbols:             req.GetSymbols(),
		CurrencyPairs:       req.GetCurrencyPairs(),
		BackfillAssets:      req.GetBackfillAssets(),
		BackfillFX:          req.GetBackfillFx(),
		RecomputeValuations: req.GetRecomputeValuations(),
	}

	res, err := s.svc.TriggerBackfill(ctx, input)
	if err != nil {
		if errors.Is(err, service.ErrInvalidDateRange) ||
			errors.Is(err, service.ErrFutureDate) ||
			errors.Is(err, service.ErrDateRequired) ||
			errors.Is(err, service.ErrNoBackfillTarget) ||
			errors.Is(err, service.ErrBackfillRangeTooLarge) {
			return nil, status.Errorf(codes.InvalidArgument, "%v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to trigger backfill: %v", err)
	}

	warnings := res.Warnings
	if warnings == nil {
		warnings = []string{}
	}

	return &pb.TriggerBackfillResponse{
		Success:       res.Success,
		PricesSynced:  int32(res.PricesSynced),
		FxRatesSynced: int32(res.FXRatesSynced),
		Message:       res.Message,
		Warnings:      warnings,
	}, nil
}

func mapFXRateToProto(r domain.FXRate) *pb.FXRateItem {
	var inverted decimal.Decimal
	if r.Rate.IsPositive() {
		inverted = decimal.NewFromInt(1).DivRound(r.Rate, 10)
	}
	return &pb.FXRateItem{
		BaseCurrency:  r.BaseCurrency,
		QuoteCurrency: r.QuoteCurrency,
		RateDate:      r.RateDate.Format("2006-01-02"),
		Rate:          decimalpb.ToProto(r.Rate),
		InvertedRate:  decimalpb.ToProto(inverted),
		Source:        r.Source,
	}
}

func mapInstrumentToProto(inst domain.Instrument) *pb.Instrument {
	p := &pb.Instrument{
		Id:           inst.ID.String(),
		Symbol:       inst.Symbol,
		Name:         inst.Name,
		CurrencyCode: inst.CurrencyCode,
		AssetClass:   inst.AssetClass,
		ExchangeCode: inst.ExchangeCode,
		IsActive:     inst.IsActive,
	}
	if inst.ISIN != nil {
		p.Isin = *inst.ISIN
	}
	return p
}

func mapInstrumentPriceToProto(p domain.InstrumentPrice) *pb.InstrumentPriceItem {
	item := &pb.InstrumentPriceItem{
		InstrumentId: p.InstrumentID.String(),
		Symbol:       p.Symbol,
		PriceDate:    p.PriceDate.Format("2006-01-02"),
		Price:        decimalpb.MoneyToProto(p.Close, p.CurrencyCode),
		Source:       p.Source,
	}
	if !p.CreatedAt.IsZero() {
		item.UpdatedAt = p.CreatedAt.Format(time.RFC3339)
	}
	return item
}

func mapDomainTxTypeToProto(t domain.TransactionType) pb.TransactionType {
	switch t {
	case domain.TxTypeBuy:
		return pb.TransactionType_TRANSACTION_TYPE_BUY
	case domain.TxTypeSell:
		return pb.TransactionType_TRANSACTION_TYPE_SELL
	case domain.TxTypeDividend:
		return pb.TransactionType_TRANSACTION_TYPE_DIVIDEND
	case domain.TxTypeDeposit:
		return pb.TransactionType_TRANSACTION_TYPE_DEPOSIT
	case domain.TxTypeWithdrawal:
		return pb.TransactionType_TRANSACTION_TYPE_WITHDRAWAL
	case domain.TxTypeInterest:
		return pb.TransactionType_TRANSACTION_TYPE_INTEREST
	case domain.TxTypeFee:
		return pb.TransactionType_TRANSACTION_TYPE_FEE
	case domain.TxTypeTax:
		return pb.TransactionType_TRANSACTION_TYPE_TAX
	case domain.TxTypeTransferIn:
		return pb.TransactionType_TRANSACTION_TYPE_TRANSFER_IN
	case domain.TxTypeTransferOut:
		return pb.TransactionType_TRANSACTION_TYPE_TRANSFER_OUT
	case domain.TxTypeFXConversion:
		return pb.TransactionType_TRANSACTION_TYPE_FX_CONVERSION
	case domain.TxTypeSplit:
		return pb.TransactionType_TRANSACTION_TYPE_SPLIT
	default:
		return pb.TransactionType_TRANSACTION_TYPE_UNSPECIFIED
	}
}

func mapSummaryToProto(summary *domain.PortfolioSummary) *pb.Portfolio {
	investments := make([]*pb.Investment, len(summary.Investments))
	for i, inv := range summary.Investments {
		var avgBuyPriceProto *commonpb.Money
		if inv.AverageBuyPrice.Amount.IsPositive() {
			avgBuyPriceProto = decimalpb.MoneyToProto(inv.AverageBuyPrice.Amount, inv.AverageBuyPrice.CurrencyCode)
		}

		investments[i] = &pb.Investment{
			Id:                  inv.ID,
			Ticker:              inv.Ticker,
			Name:                inv.Name,
			Price:               decimalpb.MoneyToProto(inv.Price.Amount, inv.Price.CurrencyCode),
			AverageBuyPrice:     avgBuyPriceProto,
			Quantity:            decimalpb.ToProto(inv.Quantity),
			TotalValue:          decimalpb.MoneyToProto(inv.TotalValue.Amount, inv.TotalValue.CurrencyCode),
			TodayReturnAmount:   decimalpb.MoneyToProto(inv.TodayReturnAmount.Amount, inv.TodayReturnAmount.CurrencyCode),
			TodayReturnPercent:  decimalpb.ToProto(inv.TodayReturnPercent),
			TotalReturnAmount:   decimalpb.MoneyToProto(inv.TotalReturnAmount.Amount, inv.TotalReturnAmount.CurrencyCode),
			TotalReturnPercent:  decimalpb.ToProto(inv.TotalReturnPercent),
			CapitalGainAmount:   decimalpb.MoneyToProto(inv.CapitalGainAmount.Amount, inv.CapitalGainAmount.CurrencyCode),
			CapitalGainPercent:  decimalpb.ToProto(inv.CapitalGainPercent),
			IncomeAmount:        decimalpb.MoneyToProto(inv.IncomeAmount.Amount, inv.IncomeAmount.CurrencyCode),
			IncomeYieldPercent:  decimalpb.ToProto(inv.IncomeYieldPercent),
			CurrencyGainAmount:  decimalpb.MoneyToProto(inv.CurrencyGainAmount.Amount, inv.CurrencyGainAmount.CurrencyCode),
			CurrencyGainPercent: decimalpb.ToProto(inv.CurrencyGainPercent),
			IsInternational:     inv.IsInternational,
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
				Id:                  "1",
				Ticker:              "AAPL",
				Name:                "Apple Inc.",
				Price:               money("185.92", "USD"),
				AverageBuyPrice:     money("156.24", "USD"),
				Quantity:            dec("142.5"),
				TotalValue:          money("26493.60", "USD"),
				TodayReturnAmount:   money("555.75", "USD"),
				TodayReturnPercent:  dec("2.14"),
				TotalReturnAmount:   money("4230.10", "USD"),
				TotalReturnPercent:  dec("18.9"),
				CapitalGainAmount:   money("4230.10", "USD"),
				CapitalGainPercent:  dec("18.9"),
				IncomeAmount:        money("0.00", "USD"),
				IncomeYieldPercent:  dec("0.00"),
				CurrencyGainAmount:  money("0.00", "USD"),
				CurrencyGainPercent: dec("0.00"),
				IsInternational:     false,
			},
		},
	}
}

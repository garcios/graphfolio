package internal

import (
	"context"
	"errors"
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
	if req.GetFee() != nil && req.GetFee().GetAmount() != nil {
		f, _, err := decimalpb.MoneyFromProto(req.GetFee())
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid fee: %v", err)
		}
		fee = &f
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
		UserID:       req.GetUserId(),
		Type:         txType,
		Symbol:       symbol,
		TradeDate:    tradeDate,
		Quantity:     qty,
		Price:        price,
		Amount:       amount,
		CurrencyCode: ccy,
		Fee:          fee,
		Notes:        notes,
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
		protoItem := &pb.TransactionItem{
			Id:        item.ID.String(),
			Type:      mapDomainTxTypeToProto(item.Type),
			TradeDate: item.TradeDate.Format("2006-01-02"),
			Amount:    decimalpb.MoneyToProto(item.Amount, item.CurrencyCode),
			Fee:       decimalpb.MoneyToProto(item.Fee, item.CurrencyCode),
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
	default:
		return pb.TransactionType_TRANSACTION_TYPE_UNSPECIFIED
	}
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

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

		var summary *domain.PortfolioSummary
		var err error
		if req.GetIncludeClosedPositions() {
			summary, err = s.svc.GetPortfolioSummary(ctx, userID, true)
		} else {
			summary, err = s.svc.GetPortfolioSummary(ctx, userID)
		}
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

func (s *PortfolioServer) CreatePortfolio(ctx context.Context, req *pb.CreatePortfolioRequest) (*pb.CreatePortfolioResponse, error) {
	if s.svc == nil {
		return nil, status.Error(codes.Unavailable, "service not initialized")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	name := req.GetName()
	if name == "" {
		name = "Main Portfolio"
	}
	curr := req.GetBaseCurrency()
	if curr == "" {
		curr = "USD"
	}

	p, err := s.svc.CreatePortfolio(ctx, req.GetUserId(), name, curr, req.GetCostBasisMethod())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create portfolio: %v", err)
	}

	return &pb.CreatePortfolioResponse{
		PortfolioId:  p.ID.String(),
		Name:         p.Name,
		BaseCurrency: p.BaseCurrency,
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

	var frankingCredit *decimal.Decimal
	if req.GetFrankingCredit() != nil && req.GetFrankingCredit().GetAmount() != nil {
		fc, _, err := decimalpb.MoneyFromProto(req.GetFrankingCredit())
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid franking credit: %v", err)
		}
		frankingCredit = &fc
	}

	var frankingPct *decimal.Decimal
	if req.GetFrankingPct() != nil {
		fp, err := decimalpb.FromProto(req.GetFrankingPct())
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid franking percentage: %v", err)
		}
		frankingPct = &fp
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
		FrankingCredit:  frankingCredit,
		FrankingPct:     frankingPct,
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
	if (req.GetPageSize() < 0 && req.GetPageSize() != -1) || req.GetPageSize() > 100 {
		return nil, status.Errorf(codes.InvalidArgument, "page size must be between 0 and 100, or -1 for unpaginated export")
	}

	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
	if pageSize == -1 {
		page = 1
		pageSize = -1
	} else {
		if page == 0 {
			page = 1
		}
		if pageSize == 0 {
			pageSize = 20
		}
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

	sortDir := domain.SortDirectionDesc
	if req.GetSortDirection() == pb.SortDirection_SORT_DIRECTION_ASC {
		sortDir = domain.SortDirectionAsc
	}

	filter := domain.TransactionFilter{
		Type:          typeFilter,
		Symbol:        symbolFilter,
		Page:          page,
		PageSize:      pageSize,
		SortDirection: sortDir,
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
		if item.WithholdingTax.IsPositive() {
			protoItem.WithholdingTax = decimalpb.MoneyToProto(item.WithholdingTax, item.CurrencyCode)
		}
		if item.FXRateToBase != nil {
			protoItem.FxRateToBase = decimalpb.ToProto(*item.FXRateToBase)
		}
		if item.ExternalRef != nil {
			protoItem.ExternalRef = *item.ExternalRef
		}
		if item.FrankingCredit.IsPositive() {
			protoItem.FrankingCredit = decimalpb.MoneyToProto(item.FrankingCredit, item.CurrencyCode)
		}
		if item.FrankingPct != nil {
			protoItem.FrankingPct = decimalpb.ToProto(*item.FrankingPct)
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

func (s *PortfolioServer) UpdateTransaction(ctx context.Context, req *pb.UpdateTransactionRequest) (*pb.UpdateTransactionResponse, error) {
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
	if req.GetCurrencyCode() != "" {
		c := req.GetCurrencyCode()
		ccy = &c
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
	if req.GetFeeCurrencyCode() != "" {
		fc := req.GetFeeCurrencyCode()
		feeCurrency = &fc
	}

	var fxRate *decimal.Decimal
	if req.GetFxRateToBase() != nil && req.GetFxRateToBase().GetValue() != "" {
		fx, err := decimalpb.FromProto(req.GetFxRateToBase())
		if err == nil {
			fxRate = &fx
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

	var frankingCredit *decimal.Decimal
	if req.GetFrankingCredit() != nil && req.GetFrankingCredit().GetAmount() != nil {
		fc, _, err := decimalpb.MoneyFromProto(req.GetFrankingCredit())
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid franking credit: %v", err)
		}
		frankingCredit = &fc
	}

	var frankingPct *decimal.Decimal
	if req.GetFrankingPct() != nil {
		fp, err := decimalpb.FromProto(req.GetFrankingPct())
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid franking percentage: %v", err)
		}
		frankingPct = &fp
	}

	input := domain.UpdateTransactionInput{
		UserID:          userID,
		TransactionID:   req.GetTransactionId(),
		Type:            txType,
		TradeDate:       tradeDate,
		Symbol:          symbol,
		Quantity:        qty,
		Price:           price,
		Amount:          amount,
		CurrencyCode:    ccy,
		Fee:             fee,
		FeeCurrencyCode: feeCurrency,
		Notes:           notes,
		FXRateToBase:    fxRate,
		FrankingCredit:  frankingCredit,
		FrankingPct:     frankingPct,
	}

	item, summary, err := s.svc.UpdateTransaction(ctx, input)
	if err != nil {
		if errors.Is(err, repository.ErrTransactionNotFound) || errors.Is(err, repository.ErrPortfolioNotFound) {
			return nil, status.Errorf(codes.NotFound, "transaction not found: %v", err)
		}
		if strings.Contains(err.Error(), "negative share inventory") {
			return nil, status.Errorf(codes.FailedPrecondition, "%v", err)
		}
		if strings.HasPrefix(err.Error(), "service:") {
			return nil, status.Errorf(codes.InvalidArgument, "%v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to update transaction: %v", err)
	}

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
	if item.WithholdingTax.IsPositive() {
		protoItem.WithholdingTax = decimalpb.MoneyToProto(item.WithholdingTax, item.CurrencyCode)
	}
	if item.FXRateToBase != nil {
		protoItem.FxRateToBase = decimalpb.ToProto(*item.FXRateToBase)
	}
	if item.ExternalRef != nil {
		protoItem.ExternalRef = *item.ExternalRef
	}
	if item.FrankingCredit.IsPositive() {
		protoItem.FrankingCredit = decimalpb.MoneyToProto(item.FrankingCredit, item.CurrencyCode)
	}
	if item.FrankingPct != nil {
		protoItem.FrankingPct = decimalpb.ToProto(*item.FrankingPct)
	}

	return &pb.UpdateTransactionResponse{
		Success:     true,
		Transaction: protoItem,
		Portfolio:   mapSummaryToProto(summary),
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

		var frankingCredit decimal.Decimal
		if item.GetFrankingCredit() != nil {
			fc, err := decimalpb.FromProto(item.GetFrankingCredit())
			if err == nil {
				frankingCredit = fc
			}
		}

		var frankingPct *decimal.Decimal
		if item.GetFrankingPct() != nil {
			fp, err := decimalpb.FromProto(item.GetFrankingPct())
			if err == nil {
				frankingPct = &fp
			}
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
			FrankingCredit:  frankingCredit,
			FrankingPct:     frankingPct,
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

func (s *PortfolioServer) TriggerDividendSync(ctx context.Context, req *pb.TriggerDividendSyncRequest) (*pb.TriggerDividendSyncResponse, error) {
	if s.svc == nil {
		return nil, status.Error(codes.Unavailable, "service not initialized")
	}

	var exchangeCode *string
	if req.ExchangeCode != nil && *req.ExchangeCode != "" {
		exchangeCode = req.ExchangeCode
	}

	res, err := s.svc.TriggerDividendSync(ctx, req.GetSymbols(), exchangeCode, req.GetSyncHistorical(), req.GetForceProvider())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to trigger dividend sync: %v", err)
	}

	return &pb.TriggerDividendSyncResponse{
		Success:              res.Success,
		SchedulesSynced:      int32(res.SchedulesSynced),
		PrimarySourceCount:   int32(res.PrimarySourceCount),
		FallbackSourceCount:  int32(res.FallbackSourceCount),
		PlaceholdersUpgraded: int32(res.PlaceholdersUpgraded),
		Message:              res.Message,
		DurationMs:           res.DurationMs,
		FailedSymbols:        res.FailedSymbols,
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

func (s *PortfolioServer) GetCashFlowReport(ctx context.Context, req *pb.GetCashFlowReportRequest) (*pb.GetCashFlowReportResponse, error) {
	if s.svc == nil {
		return nil, status.Error(codes.Unavailable, "service not initialized")
	}

	userID := req.GetUserId()
	if userID == "" {
		userID = "1"
	}

	timeframe := mapProtoCashFlowTimeframeToDomain(req.GetTimeframe())

	var fromDate, toDate time.Time
	if req.GetFromDate() != "" {
		if parsed, err := time.Parse("2006-01-02", req.GetFromDate()); err == nil {
			fromDate = parsed.UTC()
		}
	}
	if req.GetToDate() != "" {
		if parsed, err := time.Parse("2006-01-02", req.GetToDate()); err == nil {
			toDate = parsed.UTC()
		}
	}

	filter := domain.CashFlowFilter{
		UserID:    userID,
		Timeframe: timeframe,
		FromDate:  fromDate,
		ToDate:    toDate,
		Currency:  req.GetCurrency(),
	}

	report, err := s.svc.GetCashFlowReport(ctx, filter)
	if err != nil {
		if errors.Is(err, repository.ErrPortfolioNotFound) {
			return nil, status.Errorf(codes.NotFound, "portfolio not found for user: %s", userID)
		}
		return nil, status.Errorf(codes.Internal, "failed to generate cash flow report: %v", err)
	}

	items := make([]*pb.CashFlowItem, len(report.Items))
	for i, item := range report.Items {
		var sym, name string
		if item.Symbol != nil {
			sym = *item.Symbol
		}
		if item.InstrumentName != nil {
			name = *item.InstrumentName
		}
		items[i] = &pb.CashFlowItem{
			Id:             item.ID.String(),
			EventDate:      item.EventDate.Format("2006-01-02"),
			Type:           mapDomainTxTypeToProto(item.Type),
			FlowDirection:  string(item.FlowDirection),
			Category:       string(item.Category),
			Symbol:         sym,
			InstrumentName: name,
			Description:    item.Description,
			NetAmount:      decimalpb.MoneyToProto(item.NetAmount.Amount, item.NetAmount.CurrencyCode),
			RunningBalance: decimalpb.MoneyToProto(item.RunningBalance.Amount, item.RunningBalance.CurrencyCode),
			LocalAmount:    decimalpb.MoneyToProto(item.LocalAmount.Amount, item.LocalAmount.CurrencyCode),
			Fee:            decimalpb.MoneyToProto(item.Fee.Amount, item.Fee.CurrencyCode),
			WithholdingTax: decimalpb.MoneyToProto(item.WithholdingTax.Amount, item.WithholdingTax.CurrencyCode),
		}
	}

	var fromDateStr string
	if !report.FromDate.IsZero() {
		fromDateStr = report.FromDate.Format("2006-01-02")
	}

	return &pb.GetCashFlowReportResponse{
		Summary: &pb.CashFlowSummary{
			StartingCashBalance:  decimalpb.MoneyToProto(report.Summary.StartingCashBalance.Amount, report.Summary.StartingCashBalance.CurrencyCode),
			TotalInflows:         decimalpb.MoneyToProto(report.Summary.TotalInflows.Amount, report.Summary.TotalInflows.CurrencyCode),
			TotalOutflows:        decimalpb.MoneyToProto(report.Summary.TotalOutflows.Amount, report.Summary.TotalOutflows.CurrencyCode),
			NetCashFlow:          decimalpb.MoneyToProto(report.Summary.NetCashFlow.Amount, report.Summary.NetCashFlow.CurrencyCode),
			EndingCashBalance:    decimalpb.MoneyToProto(report.Summary.EndingCashBalance.Amount, report.Summary.EndingCashBalance.CurrencyCode),
			UnrealizedFxGainLoss: decimalpb.MoneyToProto(report.Summary.UnrealizedFXGainLoss.Amount, report.Summary.UnrealizedFXGainLoss.CurrencyCode),
		},
		Breakdown: &pb.CashFlowCategoryBreakdown{
			Deposits:      decimalpb.MoneyToProto(report.Breakdown.Deposits.Amount, report.Breakdown.Deposits.CurrencyCode),
			Dividends:     decimalpb.MoneyToProto(report.Breakdown.Dividends.Amount, report.Breakdown.Dividends.CurrencyCode),
			Interest:      decimalpb.MoneyToProto(report.Breakdown.Interest.Amount, report.Breakdown.Interest.CurrencyCode),
			SalesProceeds: decimalpb.MoneyToProto(report.Breakdown.SalesProceeds.Amount, report.Breakdown.SalesProceeds.CurrencyCode),
			Withdrawals:   decimalpb.MoneyToProto(report.Breakdown.Withdrawals.Amount, report.Breakdown.Withdrawals.CurrencyCode),
			Purchases:     decimalpb.MoneyToProto(report.Breakdown.Purchases.Amount, report.Breakdown.Purchases.CurrencyCode),
			Fees:          decimalpb.MoneyToProto(report.Breakdown.Fees.Amount, report.Breakdown.Fees.CurrencyCode),
			Taxes:         decimalpb.MoneyToProto(report.Breakdown.Taxes.Amount, report.Breakdown.Taxes.CurrencyCode),
		},
		Items:        items,
		BaseCurrency: report.BaseCurrency,
		FromDate:     fromDateStr,
		ToDate:       report.ToDate.Format("2006-01-02"),
	}, nil
}

func (s *PortfolioServer) GetModifiedDietzReturn(ctx context.Context, req *pb.GetModifiedDietzReturnRequest) (*pb.GetModifiedDietzReturnResponse, error) {
	if s.svc == nil {
		return nil, status.Error(codes.Unavailable, "service not initialized")
	}

	userID := req.GetUserId()
	if userID == "" {
		userID = "1"
	}

	if req.GetStartDate() == "" {
		return nil, status.Error(codes.InvalidArgument, "start_date cannot be empty")
	}
	startDate, err := time.Parse("2006-01-02", req.GetStartDate())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid start_date format (must be YYYY-MM-DD): %v", err)
	}

	if req.GetEndDate() == "" {
		return nil, status.Error(codes.InvalidArgument, "end_date cannot be empty")
	}
	endDate, err := time.Parse("2006-01-02", req.GetEndDate())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid end_date format (must be YYYY-MM-DD): %v", err)
	}

	if endDate.Before(startDate) {
		return nil, status.Errorf(codes.InvalidArgument, "end_date (%s) cannot be before start_date (%s)", req.GetEndDate(), req.GetStartDate())
	}

	result, baseCurrency, err := s.svc.GetModifiedDietzReturn(ctx, userID, startDate, endDate, req.GetUseMonthlyLinking())
	if err != nil {
		if errors.Is(err, repository.ErrPortfolioNotFound) {
			return nil, status.Errorf(codes.NotFound, "portfolio not found for user: %s", userID)
		}
		return nil, status.Errorf(codes.Internal, "failed to compute modified dietz return: %v", err)
	}

	weightsProto := make([]*pb.ExternalCashFlowWeight, len(result.CashFlowWeights))
	for i, w := range result.CashFlowWeights {
		weightsProto[i] = &pb.ExternalCashFlowWeight{
			TransactionId:  w.TransactionID,
			Date:           w.Date.Format("2006-01-02"),
			Type:           mapTxTypeStringToProto(w.Type),
			Amount:         decimalpb.MoneyToProto(w.AmountBase, baseCurrency),
			Weight:         decimalpb.ToProto(w.Weight),
			WeightedAmount: decimalpb.MoneyToProto(w.WeightedAmountBase, baseCurrency),
			LocalAmount:    decimalpb.MoneyToProto(w.LocalAmount, w.LocalCurrency),
		}
	}

	var annReturnProto *commonpb.Decimal
	if result.AnnualizedReturnPercent != nil {
		annReturnProto = decimalpb.ToProto(*result.AnnualizedReturnPercent)
	}

	startDateStr := startDate.Format("2006-01-02")
	if !result.StartDate.IsZero() {
		startDateStr = result.StartDate.Format("2006-01-02")
	}
	endDateStr := endDate.Format("2006-01-02")
	if !result.EndDate.IsZero() {
		endDateStr = result.EndDate.Format("2006-01-02")
	}

	return &pb.GetModifiedDietzReturnResponse{
		PortfolioId:             userID,
		BaseCurrency:            baseCurrency,
		StartDate:               startDateStr,
		EndDate:                 endDateStr,
		ElapsedDays:             result.ElapsedDays,
		StartValuation:          decimalpb.MoneyToProto(result.StartValuation, baseCurrency),
		EndValuation:            decimalpb.MoneyToProto(result.EndValuation, baseCurrency),
		NetExternalFlows:        decimalpb.MoneyToProto(result.NetExternalFlows, baseCurrency),
		NetInvestmentGain:       decimalpb.MoneyToProto(result.NetInvestmentGain, baseCurrency),
		WeightedCapitalEmployed: decimalpb.MoneyToProto(result.WeightedCapitalEmployed, baseCurrency),
		CumulativeReturnPercent: decimalpb.ToProto(result.CumulativeReturnPercent),
		AnnualizedReturnPercent: annReturnProto,
		IsAnnualized:            result.IsAnnualized,
		IsComputable:            result.IsComputable,
		DiagnosticCode:          result.DiagnosticCode,
		CashFlowWeights:         weightsProto,
	}, nil
}

func (s *PortfolioServer) ListBenchmarks(ctx context.Context, req *pb.ListBenchmarksRequest) (*pb.ListBenchmarksResponse, error) {
	if s.svc == nil {
		return nil, status.Error(codes.Unavailable, "service not initialized")
	}

	benchmarks, err := s.svc.ListBenchmarks(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list benchmarks: %v", err)
	}

	benchmarksProto := make([]*pb.Benchmark, len(benchmarks))
	for i, b := range benchmarks {
		benchmarksProto[i] = &pb.Benchmark{
			Id:           b.ID,
			Symbol:       b.Symbol,
			Name:         b.Name,
			ExchangeCode: b.ExchangeCode,
			CurrencyCode: b.CurrencyCode,
			IndexType:    b.IndexType,
			IsDefault:    b.IsDefault,
		}
	}

	return &pb.ListBenchmarksResponse{
		Benchmarks: benchmarksProto,
	}, nil
}

func (s *PortfolioServer) GetBenchmarkComparison(ctx context.Context, req *pb.GetBenchmarkComparisonRequest) (*pb.GetBenchmarkComparisonResponse, error) {
	if s.svc == nil {
		return nil, status.Error(codes.Unavailable, "service not initialized")
	}

	userID := req.GetUserId()
	if userID == "" {
		userID = "1"
	}

	benchmarkID := req.GetBenchmarkId()
	if benchmarkID == "" {
		return nil, status.Error(codes.InvalidArgument, "benchmark_id cannot be empty")
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
	case pb.HistoryTimeframe_HISTORY_TIMEFRAME_ALL:
		timeframe = domain.TimeframeAll
	default:
		timeframe = domain.TimeframeAll
	}

	var currencyMode domain.BenchmarkCurrencyMode
	switch req.GetCurrencyMode() {
	case pb.BenchmarkCurrencyMode_BENCHMARK_CURRENCY_MODE_BENCHMARK_NATIVE:
		currencyMode = domain.BenchmarkCurrencyModeBenchmarkNative
	default:
		currencyMode = domain.BenchmarkCurrencyModePortfolioBase
	}

	result, err := s.svc.GetBenchmarkComparison(ctx, userID, benchmarkID, timeframe, currencyMode)
	if err != nil {
		if errors.Is(err, repository.ErrPortfolioNotFound) || strings.Contains(err.Error(), "portfolio not found") {
			return nil, status.Errorf(codes.NotFound, "portfolio not found for user: %s", userID)
		}
		if strings.Contains(err.Error(), "not found") {
			return nil, status.Errorf(codes.NotFound, "benchmark not found: %s", benchmarkID)
		}
		return nil, status.Errorf(codes.Internal, "failed to get benchmark comparison: %v", err)
	}

	pointsProto := make([]*pb.BenchmarkComparisonPoint, len(result.Points))
	for i, pt := range result.Points {
		pointsProto[i] = &pb.BenchmarkComparisonPoint{
			Date:                   pt.Date.Format("2006-01-02"),
			PortfolioReturnPercent: decimalpb.ToProto(pt.PortfolioReturnPercent),
			BenchmarkReturnPercent: decimalpb.ToProto(pt.BenchmarkReturnPercent),
			ActiveReturnPercent:    decimalpb.ToProto(pt.ActiveReturnPercent),
			PortfolioValue:         decimalpb.MoneyToProto(pt.PortfolioValue.Amount, pt.PortfolioValue.CurrencyCode),
			BenchmarkPriceAdjusted: decimalpb.MoneyToProto(pt.BenchmarkPriceAdjusted.Amount, pt.BenchmarkPriceAdjusted.CurrencyCode),
		}
	}

	var portAnn, benchAnn, activeAnn, beta, te, ir *commonpb.Decimal
	if result.Metrics.PortfolioAnnualizedReturnPercent != nil {
		portAnn = decimalpb.ToProto(*result.Metrics.PortfolioAnnualizedReturnPercent)
	}
	if result.Metrics.BenchmarkAnnualizedReturnPercent != nil {
		benchAnn = decimalpb.ToProto(*result.Metrics.BenchmarkAnnualizedReturnPercent)
	}
	if result.Metrics.ActiveAnnualizedReturnPercent != nil {
		activeAnn = decimalpb.ToProto(*result.Metrics.ActiveAnnualizedReturnPercent)
	}
	if result.Metrics.Beta != nil {
		beta = decimalpb.ToProto(*result.Metrics.Beta)
	}
	if result.Metrics.TrackingErrorPercent != nil {
		te = decimalpb.ToProto(*result.Metrics.TrackingErrorPercent)
	}
	if result.Metrics.InformationRatio != nil {
		ir = decimalpb.ToProto(*result.Metrics.InformationRatio)
	}

	metricsProto := &pb.BenchmarkMetrics{
		PortfolioCumulativeReturnPercent: decimalpb.ToProto(result.Metrics.PortfolioCumulativeReturnPercent),
		BenchmarkCumulativeReturnPercent: decimalpb.ToProto(result.Metrics.BenchmarkCumulativeReturnPercent),
		ActiveCumulativeReturnPercent:    decimalpb.ToProto(result.Metrics.ActiveCumulativeReturnPercent),
		PortfolioAnnualizedReturnPercent: portAnn,
		BenchmarkAnnualizedReturnPercent: benchAnn,
		ActiveAnnualizedReturnPercent:    activeAnn,
		IsAnnualized:                     result.Metrics.IsAnnualized,
		Beta:                             beta,
		TrackingErrorPercent:             te,
		InformationRatio:                 ir,
	}

	benchProto := &pb.Benchmark{
		Id:           result.Benchmark.ID,
		Symbol:       result.Benchmark.Symbol,
		Name:         result.Benchmark.Name,
		ExchangeCode: result.Benchmark.ExchangeCode,
		CurrencyCode: result.Benchmark.CurrencyCode,
		IndexType:    result.Benchmark.IndexType,
		IsDefault:    result.Benchmark.IsDefault,
	}

	var protoCurrencyMode pb.BenchmarkCurrencyMode
	if result.CurrencyMode == domain.BenchmarkCurrencyModeBenchmarkNative {
		protoCurrencyMode = pb.BenchmarkCurrencyMode_BENCHMARK_CURRENCY_MODE_BENCHMARK_NATIVE
	} else {
		protoCurrencyMode = pb.BenchmarkCurrencyMode_BENCHMARK_CURRENCY_MODE_PORTFOLIO_BASE
	}

	startDateStr := result.StartDate.Format("2006-01-02")
	endDateStr := result.EndDate.Format("2006-01-02")

	return &pb.GetBenchmarkComparisonResponse{
		PortfolioId:  result.PortfolioID.String(),
		BaseCurrency: result.BaseCurrency,
		Benchmark:    benchProto,
		CurrencyMode: protoCurrencyMode,
		StartDate:    startDateStr,
		EndDate:      endDateStr,
		ElapsedDays:  int32(result.ElapsedDays),
		Points:       pointsProto,
		Metrics:      metricsProto,
	}, nil
}

func mapTxTypeStringToProto(t string) pb.TransactionType {
	switch strings.ToUpper(t) {
	case "DEPOSIT":
		return pb.TransactionType_TRANSACTION_TYPE_DEPOSIT
	case "WITHDRAWAL":
		return pb.TransactionType_TRANSACTION_TYPE_WITHDRAWAL
	case "TRANSFER_IN":
		return pb.TransactionType_TRANSACTION_TYPE_TRANSFER_IN
	case "TRANSFER_OUT":
		return pb.TransactionType_TRANSACTION_TYPE_TRANSFER_OUT
	default:
		return pb.TransactionType_TRANSACTION_TYPE_UNSPECIFIED
	}
}


func mapProtoCashFlowTimeframeToDomain(tf pb.CashFlowTimeframe) domain.CashFlowTimeframe {
	switch tf {
	case pb.CashFlowTimeframe_CASH_FLOW_TIMEFRAME_MTD:
		return domain.CashFlowTimeframeMTD
	case pb.CashFlowTimeframe_CASH_FLOW_TIMEFRAME_YTD:
		return domain.CashFlowTimeframeYTD
	case pb.CashFlowTimeframe_CASH_FLOW_TIMEFRAME_1M:
		return domain.CashFlowTimeframe1M
	case pb.CashFlowTimeframe_CASH_FLOW_TIMEFRAME_3M:
		return domain.CashFlowTimeframe3M
	case pb.CashFlowTimeframe_CASH_FLOW_TIMEFRAME_6M:
		return domain.CashFlowTimeframe6M
	case pb.CashFlowTimeframe_CASH_FLOW_TIMEFRAME_1Y:
		return domain.CashFlowTimeframe1Y
	case pb.CashFlowTimeframe_CASH_FLOW_TIMEFRAME_ALL:
		return domain.CashFlowTimeframeAll
	case pb.CashFlowTimeframe_CASH_FLOW_TIMEFRAME_CUSTOM:
		return domain.CashFlowTimeframeCustom
	default:
		return domain.CashFlowTimeframeYTD
	}
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

func (s *PortfolioServer) ListMissingPriceTransactions(ctx context.Context, req *pb.ListMissingPriceTransactionsRequest) (*pb.ListMissingPriceTransactionsResponse, error) {
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

	report, err := s.svc.ListMissingPriceTransactions(ctx, domain.MissingPriceFilter{
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
		return nil, status.Errorf(codes.Internal, "failed to list missing price transactions: %v", err)
	}

	protoItems := make([]*pb.MissingPriceTransactionItem, len(report.Items))
	for i, item := range report.Items {
		protoItems[i] = &pb.MissingPriceTransactionItem{
			TransactionId:  item.TransactionID.String(),
			PortfolioId:    item.PortfolioID.String(),
			InstrumentId:   item.InstrumentID.String(),
			Symbol:         item.Symbol,
			AssetName:      item.AssetName,
			ExchangeMic:    item.ExchangeMIC,
			TradeDate:      item.TradeDate.Format("2006-01-02"),
			Type:           mapDomainTxTypeToProto(item.Type),
			Quantity:       decimalpb.ToProto(item.Quantity),
			ExecutionPrice: decimalpb.MoneyToProto(item.ExecutionPrice.Amount, item.ExecutionPrice.CurrencyCode),
			CurrencyCode:   item.CurrencyCode,
		}
	}

	resp := &pb.ListMissingPriceTransactionsResponse{
		Items:                  protoItems,
		TotalCount:             int32(report.TotalCount),
		UniqueInstrumentsCount: int32(report.UniqueInstrumentsCount),
	}

	if report.EarliestTradeDate != nil {
		earliestStr := report.EarliestTradeDate.Format("2006-01-02")
		resp.EarliestTradeDate = &earliestStr
	}
	if report.LatestTradeDate != nil {
		latestStr := report.LatestTradeDate.Format("2006-01-02")
		resp.LatestTradeDate = &latestStr
	}

	return resp, nil
}

func (s *PortfolioServer) ListMissingFXRateTransactions(ctx context.Context, req *pb.ListMissingFXRateTransactionsRequest) (*pb.ListMissingFXRateTransactionsResponse, error) {
	if s.svc == nil {
		return nil, status.Error(codes.Unavailable, "service not initialized")
	}

	var currencyPair *string
	if req.CurrencyPair != nil && *req.CurrencyPair != "" {
		currencyPair = req.CurrencyPair
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

	var typeFilter *domain.TransactionType
	if req.Type != nil && *req.Type != pb.TransactionType_TRANSACTION_TYPE_UNSPECIFIED {
		var t domain.TransactionType
		switch *req.Type {
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
			return nil, status.Errorf(codes.InvalidArgument, "invalid transaction type filter: %v", *req.Type)
		}
		typeFilter = &t
	}

	report, err := s.svc.ListMissingFXRateTransactions(ctx, domain.MissingFXRateFilter{
		CurrencyPair: currencyPair,
		FromDate:     fromDate,
		ToDate:       toDate,
		Type:         typeFilter,
		Limit:        int(req.GetLimit()),
		Offset:       int(req.GetOffset()),
	})
	if err != nil {
		if errors.Is(err, service.ErrInvalidDateRange) {
			return nil, status.Errorf(codes.InvalidArgument, "%v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to list missing fx rate transactions: %v", err)
	}

	protoItems := make([]*pb.MissingFXRateTransactionItem, len(report.Items))
	for i, item := range report.Items {
		var instID *string
		if item.InstrumentID != nil {
			idStr := item.InstrumentID.String()
			instID = &idStr
		}
		var recordedRate *commonpb.Decimal
		if item.RecordedFXRate != nil {
			recordedRate = decimalpb.ToProto(*item.RecordedFXRate)
		}


		protoItems[i] = &pb.MissingFXRateTransactionItem{
			TransactionId:    item.TransactionID.String(),
			PortfolioId:      item.PortfolioID.String(),
			PortfolioName:    item.PortfolioName,
			PortfolioBaseCur: item.PortfolioBaseCur,
			TradeDate:        item.TradeDate.Format("2006-01-02"),
			Type:             mapDomainTxTypeToProto(item.Type),
			InstrumentId:     instID,
			Symbol:           item.Symbol,
			AssetName:        item.AssetName,
			Amount:           decimalpb.MoneyToProto(item.Amount.Amount, item.Amount.CurrencyCode),
			TransactionCur:   item.TransactionCur,
			RequiredPair:     item.RequiredPair,
			RecordedFxRate:   recordedRate,
		}
	}

	resp := &pb.ListMissingFXRateTransactionsResponse{
		Items:            protoItems,
		TotalCount:       int32(report.TotalCount),
		UniquePairsCount: int32(report.UniquePairsCount),
	}

	if report.EarliestTradeDate != nil {
		earliestStr := report.EarliestTradeDate.Format("2006-01-02")
		resp.EarliestTradeDate = &earliestStr
	}
	if report.LatestTradeDate != nil {
		latestStr := report.LatestTradeDate.Format("2006-01-02")
		resp.LatestTradeDate = &latestStr
	}

	return resp, nil
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
			IsClosed:            inv.IsClosed,
		}
	}

	return &pb.Portfolio{
		TotalValue:              decimalpb.MoneyToProto(summary.TotalValue.Amount, summary.TotalValue.CurrencyCode),
		TodayReturnAmount:       decimalpb.MoneyToProto(summary.TodayReturnAmount.Amount, summary.TodayReturnAmount.CurrencyCode),
		TodayReturnPercent:      decimalpb.ToProto(summary.TodayReturnPercent),
		AnnualizedReturnPercent: decimalpb.ToProto(summary.AnnualizedReturnPercent),
		CashBalance:             decimalpb.MoneyToProto(summary.CashBalance.Amount, summary.CashBalance.CurrencyCode),
		Investments:             investments,
		InceptionDate:           summary.InceptionDate,
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
		InceptionDate:           "2020-07-29",
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

func (s *PortfolioServer) GetDividendCalendar(ctx context.Context, req *pb.GetDividendCalendarRequest) (*pb.GetDividendCalendarResponse, error) {
	if s.svc == nil {
		return nil, status.Error(codes.Unavailable, "service not initialized")
	}

	events, err := s.svc.GetDividendCalendar(ctx, req.GetUserId(), int(req.GetYear()), int(req.GetMonth()))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get dividend calendar: %v", err)
	}

	var pbEvents []*pb.DividendCalendarEvent
	for _, ev := range events {
		pbFreq := pb.DividendFrequency_DIVIDEND_FREQUENCY_UNSPECIFIED
		switch ev.Frequency {
		case domain.DividendFreqMonthly:
			pbFreq = pb.DividendFrequency_DIVIDEND_FREQUENCY_MONTHLY
		case domain.DividendFreqQuarterly:
			pbFreq = pb.DividendFrequency_DIVIDEND_FREQUENCY_QUARTERLY
		case domain.DividendFreqSemiAnnual:
			pbFreq = pb.DividendFrequency_DIVIDEND_FREQUENCY_SEMI_ANNUAL
		case domain.DividendFreqAnnual:
			pbFreq = pb.DividendFrequency_DIVIDEND_FREQUENCY_ANNUAL
		case domain.DividendFreqIrregular:
			pbFreq = pb.DividendFrequency_DIVIDEND_FREQUENCY_IRREGULAR
		}

		pbStatus := pb.DividendEventStatus_DIVIDEND_EVENT_STATUS_UNSPECIFIED
		switch ev.Status {
		case domain.DividendStatusAnnounced:
			pbStatus = pb.DividendEventStatus_DIVIDEND_EVENT_STATUS_ANNOUNCED
		case domain.DividendStatusExPassed:
			pbStatus = pb.DividendEventStatus_DIVIDEND_EVENT_STATUS_EX_DATE_PASSED
		case domain.DividendStatusPaid:
			pbStatus = pb.DividendEventStatus_DIVIDEND_EVENT_STATUS_PAID
		case domain.DividendStatusCancelled:
			pbStatus = pb.DividendEventStatus_DIVIDEND_EVENT_STATUS_CANCELLED
		}

		recordDate := ""
		if ev.RecordDate != nil {
			recordDate = ev.RecordDate.Format("2006-01-02")
		}

		pbEvents = append(pbEvents, &pb.DividendCalendarEvent{
			Id:              ev.ID.String(),
			InstrumentId:    ev.InstrumentID.String(),
			Symbol:          ev.Symbol,
			Name:            ev.Name,
			ExDate:          ev.ExDate.Format("2006-01-02"),
			RecordDate:      recordDate,
			PaymentDate:     ev.PaymentDate.Format("2006-01-02"),
			SharesHeld:      decimalpb.ToProto(ev.SharesHeld),
			DpsDeclared:     decimalpb.MoneyToProto(ev.DPSDeclared.Amount, ev.DPSDeclared.CurrencyCode),
			GrossAmountBase: decimalpb.MoneyToProto(ev.GrossAmountBase.Amount, ev.GrossAmountBase.CurrencyCode),
			NetAmountBase:   decimalpb.MoneyToProto(ev.NetAmountBase.Amount, ev.NetAmountBase.CurrencyCode),
			Frequency:       pbFreq,
			Status:          pbStatus,
			IsUserEntitled:  ev.IsUserEntitled,
			IsEstimated:     ev.IsEstimated,
			FrankingPct:     decimalpb.ToProto(ev.FrankingPct),
		})
	}

	return &pb.GetDividendCalendarResponse{
		Events: pbEvents,
	}, nil
}

func (s *PortfolioServer) GetDividendYieldAnalytics(ctx context.Context, req *pb.GetDividendYieldAnalyticsRequest) (*pb.GetDividendYieldAnalyticsResponse, error) {
	if s.svc == nil {
		return nil, status.Error(codes.Unavailable, "service not initialized")
	}

	analytics, err := s.svc.GetDividendYieldAnalytics(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get dividend yield analytics: %v", err)
	}

	var pbProjections []*pb.MonthlyDividendProjection
	for _, mp := range analytics.MonthlyProjections {
		pbProjections = append(pbProjections, &pb.MonthlyDividendProjection{
			Year:            int32(mp.Year),
			Month:           int32(mp.Month),
			ProjectedIncome: decimalpb.MoneyToProto(mp.ProjectedIncome.Amount, mp.ProjectedIncome.CurrencyCode),
			DividendCount:   int32(mp.DividendCount),
		})
	}

	var pbHoldings []*pb.HoldingYieldItem
	for _, hy := range analytics.HoldingsYield {
		var yocProto *commonpb.Decimal
		if hy.YieldOnCostPct != nil {
			yocProto = decimalpb.ToProto(*hy.YieldOnCostPct)
		}
		var cagrProto *commonpb.Decimal
		if hy.DividendCAGR3Y != nil {
			cagrProto = decimalpb.ToProto(*hy.DividendCAGR3Y)
		}

		nextEx := ""
		if hy.NextExDate != nil {
			nextEx = hy.NextExDate.Format("2006-01-02")
		}
		nextPay := ""
		if hy.NextPaymentDate != nil {
			nextPay = hy.NextPaymentDate.Format("2006-01-02")
		}

		pbFreq := pb.DividendFrequency_DIVIDEND_FREQUENCY_UNSPECIFIED
		switch hy.Frequency {
		case domain.DividendFreqMonthly:
			pbFreq = pb.DividendFrequency_DIVIDEND_FREQUENCY_MONTHLY
		case domain.DividendFreqQuarterly:
			pbFreq = pb.DividendFrequency_DIVIDEND_FREQUENCY_QUARTERLY
		case domain.DividendFreqSemiAnnual:
			pbFreq = pb.DividendFrequency_DIVIDEND_FREQUENCY_SEMI_ANNUAL
		case domain.DividendFreqAnnual:
			pbFreq = pb.DividendFrequency_DIVIDEND_FREQUENCY_ANNUAL
		case domain.DividendFreqIrregular:
			pbFreq = pb.DividendFrequency_DIVIDEND_FREQUENCY_IRREGULAR
		}

		pbHoldings = append(pbHoldings, &pb.HoldingYieldItem{
			InstrumentId:        hy.InstrumentID.String(),
			Symbol:              hy.Symbol,
			Name:                hy.Name,
			Quantity:            decimalpb.ToProto(hy.Quantity),
			CurrentPrice:        decimalpb.MoneyToProto(hy.CurrentPrice.Amount, hy.CurrentPrice.CurrencyCode),
			AverageBuyPrice:     decimalpb.MoneyToProto(hy.AverageBuyPrice.Amount, hy.AverageBuyPrice.CurrencyCode),
			CurrentMarketValue:  decimalpb.MoneyToProto(hy.CurrentMarketValue.Amount, hy.CurrentMarketValue.CurrencyCode),
			CostBasis:           decimalpb.MoneyToProto(hy.CostBasis.Amount, hy.CostBasis.CurrencyCode),
			AnnualDps:           decimalpb.MoneyToProto(hy.AnnualDPS.Amount, hy.AnnualDPS.CurrencyCode),
			ProjectedAnnualCash: decimalpb.MoneyToProto(hy.ProjectedAnnualCash.Amount, hy.ProjectedAnnualCash.CurrencyCode),
			CurrentYieldPct:     decimalpb.ToProto(hy.CurrentYieldPct),
			YieldOnCostPct:      yocProto,
			DividendCagr_3Y:     cagrProto,
			Frequency:           pbFreq,
			NextExDate:          nextEx,
			NextPaymentDate:     nextPay,
		})
	}

	var portfolioYOCProto *commonpb.Decimal
	if analytics.PortfolioYieldOnCost != nil {
		portfolioYOCProto = decimalpb.ToProto(*analytics.PortfolioYieldOnCost)
	}

	return &pb.GetDividendYieldAnalyticsResponse{
		AnnualForwardIncome:   decimalpb.MoneyToProto(analytics.AnnualForwardIncome.Amount, analytics.AnnualForwardIncome.CurrencyCode),
		MonthlyAverageIncome:  decimalpb.MoneyToProto(analytics.MonthlyAverageIncome.Amount, analytics.MonthlyAverageIncome.CurrencyCode),
		PortfolioCurrentYield: decimalpb.ToProto(analytics.PortfolioCurrentYield),
		PortfolioYieldOnCost:  portfolioYOCProto,
		PendingCashInflow:     decimalpb.MoneyToProto(analytics.PendingCashInflow.Amount, analytics.PendingCashInflow.CurrencyCode),
		MonthlyProjections:    pbProjections,
		HoldingsYield:         pbHoldings,
	}, nil
}

func (s *PortfolioServer) SimulateDRIP(ctx context.Context, req *pb.SimulateDRIPRequest) (*pb.SimulateDRIPResponse, error) {
	if s.svc == nil {
		return nil, status.Error(codes.Unavailable, "service not initialized")
	}

	var divGrowth decimal.Decimal
	if req.DividendGrowthRatePct != nil {
		d, err := decimalpb.FromProto(req.DividendGrowthRatePct)
		if err == nil {
			divGrowth = d
		}
	}
	var capGrowth decimal.Decimal
	if req.CapitalAppreciationPct != nil {
		c, err := decimalpb.FromProto(req.CapitalAppreciationPct)
		if err == nil {
			capGrowth = c
		}
	}

	input := domain.DRIPSimulationInput{
		HorizonYears:           int(req.GetHorizonYears()),
		DividendGrowthRatePct:  divGrowth,
		CapitalAppreciationPct: capGrowth,
	}

	res, err := s.svc.SimulateDRIP(ctx, req.GetUserId(), input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to simulate DRIP: %v", err)
	}

	var pbPoints []*pb.DRIPSimulationPoint
	for _, pt := range res.Points {
		pbPoints = append(pbPoints, &pb.DRIPSimulationPoint{
			Year:                  int32(pt.Year),
			DripPortfolioValue:    decimalpb.ToProto(pt.DRIPPortfolioValue),
			CashPortfolioValue:    decimalpb.ToProto(pt.CashPortfolioValue),
			DripAccumulatedShares: decimalpb.ToProto(pt.DRIPAccumulatedShares),
			CashAccumulatedCash:   decimalpb.ToProto(pt.CashAccumulatedCash),
			DripAnnualDividend:    decimalpb.ToProto(pt.DRIPAnnualDividend),
			CashAnnualDividend:    decimalpb.ToProto(pt.CashAnnualDividend),
		})
	}

	return &pb.SimulateDRIPResponse{
		StartingPortfolioValue: decimalpb.ToProto(res.StartingPortfolioValue),
		StartingAnnualDividend: decimalpb.ToProto(res.StartingAnnualDividend),
		BaseCurrency:           res.BaseCurrency,
		HorizonYears:           int32(res.HorizonYears),
		Points:                 pbPoints,
		TerminalDripValue:      decimalpb.ToProto(res.TerminalDRIPValue),
		TerminalCashValue:      decimalpb.ToProto(res.TerminalCashValue),
		DripBenefitAmount:      decimalpb.ToProto(res.DRIPBenefitAmount),
	}, nil
}

func (s *PortfolioServer) ListTaxLots(ctx context.Context, req *pb.ListTaxLotsRequest) (*pb.ListTaxLotsResponse, error) {
	if s.svc == nil {
		return nil, status.Error(codes.Unavailable, "service not initialized")
	}

	var statusFilter domain.TaxLotStatus
	switch req.GetStatus() {
	case pb.TaxLotStatus_TAX_LOT_STATUS_OPEN:
		statusFilter = domain.TaxLotStatusOpen
	case pb.TaxLotStatus_TAX_LOT_STATUS_CLOSED:
		statusFilter = domain.TaxLotStatusClosed
	default:
		statusFilter = domain.TaxLotStatusAll
	}

	var instID *string
	if req.InstrumentId != nil && req.GetInstrumentId() != "" {
		idStr := req.GetInstrumentId()
		instID = &idStr
	}

	lots, err := s.svc.ListTaxLots(ctx, req.GetUserId(), instID, statusFilter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list tax lots: %v", err)
	}

	port, _ := s.svc.GetPortfolioSummary(ctx, req.GetUserId())
	baseCurr := "AUD"
	if port != nil {
		baseCurr = port.Portfolio.BaseCurrency
	}

	var pbLots []*pb.TaxLotItem
	for _, l := range lots {
		pbLots = append(pbLots, mapDomainTaxLotToProto(l, baseCurr))
	}

	return &pb.ListTaxLotsResponse{
		TaxLots: pbLots,
	}, nil
}

func (s *PortfolioServer) GetRealizedGainsReport(ctx context.Context, req *pb.GetRealizedGainsReportRequest) (*pb.GetRealizedGainsReportResponse, error) {
	if s.svc == nil {
		return nil, status.Error(codes.Unavailable, "service not initialized")
	}

	var fromDate, toDate *string
	if req.FromDate != nil && req.GetFromDate() != "" {
		fd := req.GetFromDate()
		fromDate = &fd
	}
	if req.ToDate != nil && req.GetToDate() != "" {
		td := req.GetToDate()
		toDate = &td
	}

	rep, err := s.svc.GetRealizedGainsReport(ctx, req.GetUserId(), fromDate, toDate, req.FiscalYear)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get realized gains report: %v", err)
	}

	var pbDisposals []*pb.RealizedDisposalItem
	for _, d := range rep.Disposals {
		pbDisposals = append(pbDisposals, mapDomainDisposalToProto(d, rep.BaseCurrency))
	}

	return &pb.GetRealizedGainsReportResponse{
		Summary: &pb.RealizedGainsSummary{
			TotalProceedsBase:         decimalpb.MoneyToProto(rep.Summary.TotalProceedsBase, rep.BaseCurrency),
			TotalCostBasisRelieved:     decimalpb.MoneyToProto(rep.Summary.TotalCostBasisRelieved, rep.BaseCurrency),
			GrossRealizedGainsBase:     decimalpb.MoneyToProto(rep.Summary.GrossRealizedGainsBase, rep.BaseCurrency),
			GrossRealizedLossesBase:    decimalpb.MoneyToProto(rep.Summary.GrossRealizedLossesBase, rep.BaseCurrency),
			NetRealizedPnlBase:         decimalpb.MoneyToProto(rep.Summary.NetRealizedPnLBase, rep.BaseCurrency),
			ShortTermGainsBase:         decimalpb.MoneyToProto(rep.Summary.ShortTermGainsBase, rep.BaseCurrency),
			LongTermGainsBase:          decimalpb.MoneyToProto(rep.Summary.LongTermGainsBase, rep.BaseCurrency),
			DiscountedTaxableGainBase:  decimalpb.MoneyToProto(rep.Summary.DiscountedTaxableGainBase, rep.BaseCurrency),
			TotalDisposalCount:         int32(rep.Summary.TotalDisposalCount),
		},
		Disposals:    pbDisposals,
		FromDate:     rep.FromDate,
		ToDate:       rep.ToDate,
		BaseCurrency: rep.BaseCurrency,
	}, nil
}

func (s *PortfolioServer) GetTaxReport(ctx context.Context, req *pb.GetTaxReportRequest) (*pb.GetTaxReportResponse, error) {
	if s.svc == nil {
		return nil, status.Error(codes.Unavailable, "service not initialized")
	}

	var fromDate, toDate *string
	if req.FromDate != nil && req.GetFromDate() != "" {
		fd := req.GetFromDate()
		fromDate = &fd
	}
	if req.ToDate != nil && req.GetToDate() != "" {
		td := req.GetToDate()
		toDate = &td
	}

	rep, err := s.svc.GetTaxReport(ctx, req.GetUserId(), req.FiscalYear, fromDate, toDate)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get tax report: %v", err)
	}

	var pbDisposals []*pb.RealizedDisposalItem
	for _, d := range rep.Disposals {
		pbDisposals = append(pbDisposals, mapDomainDisposalToProto(d, rep.BaseCurrency))
	}

	var pbLots []*pb.TaxLotItem
	for _, l := range rep.OpenLots {
		pbLots = append(pbLots, mapDomainTaxLotToProto(l, rep.BaseCurrency))
	}

	return &pb.GetTaxReportResponse{
		Summary: &pb.TaxReportSummary{
			TaxableIncome: &pb.TaxableIncomeSummary{
				DividendAndInterestIncome: decimalpb.MoneyToProto(rep.Summary.TaxableIncome.DividendAndInterestIncome, rep.BaseCurrency),
				NetCapitalGains:            decimalpb.MoneyToProto(rep.Summary.TaxableIncome.NetCapitalGains, rep.BaseCurrency),
				FrankingCredits:            decimalpb.MoneyToProto(rep.Summary.TaxableIncome.FrankingCredits, rep.BaseCurrency),
				TotalAssessableIncome:      decimalpb.MoneyToProto(rep.Summary.TaxableIncome.TotalAssessableIncome, rep.BaseCurrency),
				DividendInterestRatioPct:  decimalpb.ToProto(rep.Summary.TaxableIncome.DividendInterestRatioPct),
				CapitalGainsRatioPct:      decimalpb.ToProto(rep.Summary.TaxableIncome.CapitalGainsRatioPct),
				FrankingCreditsRatioPct:   decimalpb.ToProto(rep.Summary.TaxableIncome.FrankingCreditsRatioPct),
			},
			CapitalGains: &pb.CapitalGainsSummary{
				TotalCapitalGains:  decimalpb.MoneyToProto(rep.Summary.CapitalGains.TotalCapitalGains, rep.BaseCurrency),
				LossesApplied:      decimalpb.MoneyToProto(rep.Summary.CapitalGains.LossesApplied, rep.BaseCurrency),
				CgtDiscountAmount:  decimalpb.MoneyToProto(rep.Summary.CapitalGains.CGTDiscountAmount, rep.BaseCurrency),
				NetCapitalGain:     decimalpb.MoneyToProto(rep.Summary.CapitalGains.NetCapitalGain, rep.BaseCurrency),
				CgtDiscountRatePct: decimalpb.ToProto(rep.Summary.CapitalGains.CGTDiscountRatePct),
				GainDisposalsCount: int32(rep.Summary.CapitalGains.GainDisposalsCount),
				LossDisposalsCount: int32(rep.Summary.CapitalGains.LossDisposalsCount),
			},
			TaxOffsets: &pb.TaxOffsetsSummary{
				FrankingCreditsOffset: decimalpb.MoneyToProto(rep.Summary.TaxOffsets.FrankingCreditsOffset, rep.BaseCurrency),
				ForeignTaxOffsets:     decimalpb.MoneyToProto(rep.Summary.TaxOffsets.ForeignTaxOffsets, rep.BaseCurrency),
				TotalAvailableOffsets: decimalpb.MoneyToProto(rep.Summary.TaxOffsets.TotalAvailableOffsets, rep.BaseCurrency),
			},
			FiscalYearLabel: rep.Summary.FiscalYearLabel,
			BaseCurrency:    rep.Summary.BaseCurrency,
		},
		Disposals:       pbDisposals,
		OpenLots:        pbLots,
		FromDate:        rep.FromDate,
		ToDate:          rep.ToDate,
		FiscalYearLabel: rep.FiscalYearLabel,
		BaseCurrency:    rep.BaseCurrency,
	}, nil
}

func mapDomainTaxLotToProto(tl domain.TaxLot, baseCurrency string) *pb.TaxLotItem {
	var closedDateStr string
	if tl.ClosedDate != nil {
		closedDateStr = tl.ClosedDate.Format("2006-01-02")
	}

	termClassification := pb.HoldingTermClassification_HOLDING_TERM_SHORT_TERM
	if tl.TermClassification == domain.TermLongTerm {
		termClassification = pb.HoldingTermClassification_HOLDING_TERM_LONG_TERM
	}

	return &pb.TaxLotItem{
		Id:                          tl.ID.String(),
		PortfolioId:                 tl.PortfolioID.String(),
		InstrumentId:                tl.InstrumentID.String(),
		Ticker:                      tl.Ticker,
		Name:                        tl.Name,
		OpenTransactionId:           tl.OpenTransactionID.String(),
		AcquiredDate:                tl.AcquiredDate.Format("2006-01-02"),
		ClosedDate:                  closedDateStr,
		OriginalQuantity:            decimalpb.ToProto(tl.OriginalQuantity),
		RemainingQuantity:           decimalpb.ToProto(tl.RemainingQuantity),
		CostBasis:                   decimalpb.MoneyToProto(tl.CostBasis, baseCurrency),
		CostBasisBase:               decimalpb.MoneyToProto(tl.CostBasisBase, baseCurrency),
		UnitCostBase:                decimalpb.MoneyToProto(tl.UnitCostBase, baseCurrency),
		CurrentPrice:                decimalpb.MoneyToProto(tl.CurrentPrice, baseCurrency),
		MarketValueBase:             decimalpb.MoneyToProto(tl.MarketValueBase, baseCurrency),
		UnrealizedPnlBase:           decimalpb.MoneyToProto(tl.UnrealizedPnLBase, baseCurrency),
		UnrealizedPnlPercent:        decimalpb.ToProto(tl.UnrealizedPnLPercent),
		HoldingPeriodDays:           int32(tl.HoldingPeriodDays),
		TermClassification:          termClassification,
		IsCgtDiscountEligible:       tl.IsCGTDiscountEligible,
		IsLossHarvestingCandidate:   tl.IsLossHarvestCandidate,
		HasWashSaleRisk:             tl.HasWashSaleRisk,
	}
}

func mapDomainDisposalToProto(d domain.LotDisposal, baseCurrency string) *pb.RealizedDisposalItem {
	termClassification := pb.HoldingTermClassification_HOLDING_TERM_SHORT_TERM
	if d.TermClassification == domain.TermLongTerm {
		termClassification = pb.HoldingTermClassification_HOLDING_TERM_LONG_TERM
	}

	return &pb.RealizedDisposalItem{
		DisposalId:                 d.ID.String(),
		TaxLotId:                   d.TaxLotID.String(),
		SellTransactionId:          d.SellTransactionID.String(),
		InstrumentId:               d.InstrumentID.String(),
		Ticker:                     d.Ticker,
		Name:                       d.Name,
		AcquiredDate:               d.AcquiredDate.Format("2006-01-02"),
		DisposedDate:               d.DisposedDate.Format("2006-01-02"),
		QuantityDisposed:           decimalpb.ToProto(d.Quantity),
		CostBasisRelievedBase:      decimalpb.MoneyToProto(d.CostBasisReleasedBase, baseCurrency),
		ProceedsBase:               decimalpb.MoneyToProto(d.ProceedsBase, baseCurrency),
		RealizedPnlBase:            decimalpb.MoneyToProto(d.RealizedPnLBase, baseCurrency),
		RealizedPnlPercent:         decimalpb.ToProto(d.RealizedPnLPercent),
		HoldingPeriodDays:          int32(d.HoldingPeriodDays),
		TermClassification:         termClassification,
		IsCgtDiscountEligible:      d.IsCGTDiscountEligible,
		DiscountedTaxablePnlBase:   decimalpb.MoneyToProto(d.DiscountedTaxablePnLBase, baseCurrency),
	}
}



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
		protoInsts[i] = &pb.Instrument{
			Id:           inst.ID.String(),
			Symbol:       inst.Symbol,
			Name:         inst.Name,
			CurrencyCode: inst.CurrencyCode,
			AssetClass:   inst.AssetClass,
		}
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

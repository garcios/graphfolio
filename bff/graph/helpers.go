package graph

import (
	"bff/graph/model"

	commonpb "graphfolio/proto/common/v1"
	pb "graphfolio/proto/portfolio/v1"

	"github.com/shopspring/decimal"
)

func toModelMoney(m *commonpb.Money) *model.Money {
	if m == nil {
		return &model.Money{
			Amount:       model.Decimal(decimal.Zero),
			CurrencyCode: "USD",
		}
	}
	d, err := decimal.NewFromString(m.GetAmount().GetValue())
	if err != nil {
		d = decimal.Zero
	}
	return &model.Money{
		Amount:       model.Decimal(d),
		CurrencyCode: m.GetCurrencyCode(),
	}
}

func toModelDecimal(p *commonpb.Decimal) model.Decimal {
	if p == nil {
		return model.Decimal(decimal.Zero)
	}
	d, err := decimal.NewFromString(p.GetValue())
	if err != nil {
		return model.Decimal(decimal.Zero)
	}
	return model.Decimal(d)
}

func toModelPortfolio(p *pb.Portfolio) *model.Portfolio {
	if p == nil {
		return nil
	}

	investments := make([]*model.Investment, len(p.Investments))
	for i, inv := range p.Investments {
		investments[i] = &model.Investment{
			ID:                 inv.Id,
			Ticker:             inv.Ticker,
			Name:               inv.Name,
			Price:              toModelMoney(inv.Price),
			Quantity:           toModelDecimal(inv.Quantity),
			TotalValue:         toModelMoney(inv.TotalValue),
			TodayReturnAmount:  toModelMoney(inv.TodayReturnAmount),
			TodayReturnPercent: toModelDecimal(inv.TodayReturnPercent),
			TotalReturnAmount:  toModelMoney(inv.TotalReturnAmount),
			TotalReturnPercent: toModelDecimal(inv.TotalReturnPercent),
		}
	}

	return &model.Portfolio{
		TotalValue:              toModelMoney(p.TotalValue),
		TodayReturnAmount:       toModelMoney(p.TodayReturnAmount),
		TodayReturnPercent:      toModelDecimal(p.TodayReturnPercent),
		AnnualizedReturnPercent: toModelDecimal(p.AnnualizedReturnPercent),
		CashBalance:             toModelMoney(p.CashBalance),
		Investments:             investments,
	}
}

func toProtoDecimal(d *model.Decimal) *commonpb.Decimal {
	if d == nil {
		return nil
	}
	return &commonpb.Decimal{
		Value: d.String(),
	}
}

func toProtoMoney(d *model.Decimal, currencyCode string) *commonpb.Money {
	if d == nil {
		return nil
	}
	if currencyCode == "" {
		currencyCode = "USD"
	}
	return &commonpb.Money{
		Amount:       toProtoDecimal(d),
		CurrencyCode: currencyCode,
	}
}

func toProtoTransactionType(t model.TransactionType) pb.TransactionType {
	switch t {
	case model.TransactionTypeBuy:
		return pb.TransactionType_TRANSACTION_TYPE_BUY
	case model.TransactionTypeSell:
		return pb.TransactionType_TRANSACTION_TYPE_SELL
	case model.TransactionTypeDividend:
		return pb.TransactionType_TRANSACTION_TYPE_DIVIDEND
	case model.TransactionTypeDeposit:
		return pb.TransactionType_TRANSACTION_TYPE_DEPOSIT
	case model.TransactionTypeWithdrawal:
		return pb.TransactionType_TRANSACTION_TYPE_WITHDRAWAL
	case model.TransactionTypeInterest:
		return pb.TransactionType_TRANSACTION_TYPE_INTEREST
	case model.TransactionTypeFee:
		return pb.TransactionType_TRANSACTION_TYPE_FEE
	case model.TransactionTypeTax:
		return pb.TransactionType_TRANSACTION_TYPE_TAX
	case model.TransactionTypeTransferIn:
		return pb.TransactionType_TRANSACTION_TYPE_TRANSFER_IN
	case model.TransactionTypeTransferOut:
		return pb.TransactionType_TRANSACTION_TYPE_TRANSFER_OUT
	case model.TransactionTypeFxConversion:
		return pb.TransactionType_TRANSACTION_TYPE_FX_CONVERSION
	default:
		return pb.TransactionType_TRANSACTION_TYPE_UNSPECIFIED
	}
}

func toProtoHistoryTimeframe(t model.HistoryTimeframe) pb.HistoryTimeframe {
	switch t {
	case model.HistoryTimeframeTimeframe1d:
		return pb.HistoryTimeframe_HISTORY_TIMEFRAME_1D
	case model.HistoryTimeframeTimeframe1w:
		return pb.HistoryTimeframe_HISTORY_TIMEFRAME_1W
	case model.HistoryTimeframeTimeframe1m:
		return pb.HistoryTimeframe_HISTORY_TIMEFRAME_1M
	case model.HistoryTimeframeTimeframe1y:
		return pb.HistoryTimeframe_HISTORY_TIMEFRAME_1Y
	case model.HistoryTimeframeTimeframeAll:
		return pb.HistoryTimeframe_HISTORY_TIMEFRAME_ALL
	default:
		return pb.HistoryTimeframe_HISTORY_TIMEFRAME_UNSPECIFIED
	}
}

func toModelPortfolioHistory(resp *pb.GetPortfolioHistoryResponse) *model.PortfolioHistory {
	if resp == nil {
		return nil
	}

	points := make([]*model.ValuationPoint, len(resp.Points))
	for i, pt := range resp.Points {
		var dailyRet *model.Decimal
		if pt.DailyReturn != nil {
			d := toModelDecimal(pt.DailyReturn)
			dailyRet = &d
		}

		points[i] = &model.ValuationPoint{
			Date:        pt.Date,
			TotalValue:  toModelMoney(pt.TotalValue),
			MarketValue: toModelMoney(pt.MarketValue),
			CashValue:   toModelMoney(pt.CashValue),
			TwrIndex:    toModelDecimal(pt.TwrIndex),
			DailyReturn: dailyRet,
		}
	}

	return &model.PortfolioHistory{
		Points:        points,
		StartValue:    toModelMoney(resp.StartValue),
		EndValue:      toModelMoney(resp.EndValue),
		ReturnAmount:  toModelMoney(resp.ReturnAmount),
		ReturnPercent: toModelDecimal(resp.ReturnPercent),
	}
}

package graph

import (
	"bff/graph/model"

	commonpb "graphfolio/proto/common/v1"
	pb "graphfolio/proto/portfolio/v1"
	userpb "graphfolio/proto/user/v1"

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

func toModelTransactionItem(item *pb.TransactionItem) *model.TransactionItem {
	if item == nil {
		return nil
	}
	var qty *model.Decimal
	if item.Quantity != nil {
		d := toModelDecimal(item.Quantity)
		qty = &d
	}
	var price *model.Money
	if item.Price != nil {
		price = toModelMoney(item.Price)
	}
	var symbol *string
	if item.Symbol != "" {
		s := item.Symbol
		symbol = &s
	}
	var instName *string
	if item.InstrumentName != "" {
		n := item.InstrumentName
		instName = &n
	}
	var notes *string
	if item.Notes != "" {
		n := item.Notes
		notes = &n
	}

	return &model.TransactionItem{
		ID:             item.Id,
		Type:           toModelTransactionType(item.Type),
		Symbol:         symbol,
		InstrumentName: instName,
		TradeDate:      item.TradeDate,
		Quantity:       qty,
		Price:          price,
		Amount:         toModelMoney(item.Amount),
		Fee:            toModelMoney(item.Fee),
		Notes:          notes,
		CreatedAt:      item.CreatedAt,
	}
}

func toModelTransactionType(t pb.TransactionType) model.TransactionType {
	switch t {
	case pb.TransactionType_TRANSACTION_TYPE_BUY:
		return model.TransactionTypeBuy
	case pb.TransactionType_TRANSACTION_TYPE_SELL:
		return model.TransactionTypeSell
	case pb.TransactionType_TRANSACTION_TYPE_DIVIDEND:
		return model.TransactionTypeDividend
	case pb.TransactionType_TRANSACTION_TYPE_DEPOSIT:
		return model.TransactionTypeDeposit
	case pb.TransactionType_TRANSACTION_TYPE_WITHDRAWAL:
		return model.TransactionTypeWithdrawal
	case pb.TransactionType_TRANSACTION_TYPE_INTEREST:
		return model.TransactionTypeInterest
	case pb.TransactionType_TRANSACTION_TYPE_FEE:
		return model.TransactionTypeFee
	case pb.TransactionType_TRANSACTION_TYPE_TAX:
		return model.TransactionTypeTax
	case pb.TransactionType_TRANSACTION_TYPE_TRANSFER_IN:
		return model.TransactionTypeTransferIn
	case pb.TransactionType_TRANSACTION_TYPE_TRANSFER_OUT:
		return model.TransactionTypeTransferOut
	case pb.TransactionType_TRANSACTION_TYPE_FX_CONVERSION:
		return model.TransactionTypeFxConversion
	default:
		return model.TransactionTypeBuy
	}
}

func toModelTransactionsConnection(resp *pb.ListTransactionsResponse) *model.TransactionsConnection {
	if resp == nil {
		return &model.TransactionsConnection{
			Items:      []*model.TransactionItem{},
			TotalCount: 0,
			Page:       1,
			PageSize:   20,
		}
	}

	items := make([]*model.TransactionItem, len(resp.Transactions))
	for i, tx := range resp.Transactions {
		items[i] = toModelTransactionItem(tx)
	}

	return &model.TransactionsConnection{
		Items:      items,
		TotalCount: int(resp.TotalCount),
		Page:       int(resp.Page),
		PageSize:   int(resp.PageSize),
	}
}

func toModelInstrument(inst *pb.Instrument) *model.Instrument {
	if inst == nil {
		return nil
	}
	var isin *string
	if inst.Isin != "" {
		isin = &inst.Isin
	}
	return &model.Instrument{
		ID:           inst.Id,
		Symbol:       inst.Symbol,
		Name:         inst.Name,
		CurrencyCode: inst.CurrencyCode,
		AssetClass:   inst.AssetClass,
		ExchangeCode: inst.ExchangeCode,
		Isin:         isin,
		IsActive:     inst.IsActive,
	}
}

func toModelInstrumentPrice(p *pb.InstrumentPriceItem) *model.InstrumentPrice {
	if p == nil {
		return nil
	}
	return &model.InstrumentPrice{
		ID:        p.InstrumentId + "_" + p.PriceDate,
		Symbol:    p.Symbol,
		PriceDate: p.PriceDate,
		Price:     toModelMoney(p.Price),
		Source:    p.Source,
		UpdatedAt: p.UpdatedAt,
	}
}

func toModelInstrumentPricesConnection(resp *pb.ListInstrumentPricesResponse) *model.InstrumentPricesConnection {
	if resp == nil {
		return &model.InstrumentPricesConnection{
			Items:      []*model.InstrumentPrice{},
			TotalCount: 0,
		}
	}

	items := make([]*model.InstrumentPrice, len(resp.Prices))
	for i, p := range resp.Prices {
		items[i] = toModelInstrumentPrice(p)
	}

	return &model.InstrumentPricesConnection{
		Items:      items,
		TotalCount: int(resp.TotalCount),
	}
}

func toModelIngestionStatus(resp *pb.GetIngestionStatusResponse) *model.IngestionStatus {
	if resp == nil {
		return nil
	}
	feeds := make([]*model.FeedHealthStatus, len(resp.Feeds))
	for i, f := range resp.Feeds {
		feeds[i] = &model.FeedHealthStatus{
			Name:     f.Name,
			Status:   f.Status,
			Provider: f.Provider,
			Schedule: f.Schedule,
			LastRun:  f.LastRun,
			Details:  f.Details,
		}
	}

	var latestPriceDate *string
	if resp.LatestPriceDate != "" {
		latestPriceDate = &resp.LatestPriceDate
	}
	var latestFxDate *string
	if resp.LatestFxDate != "" {
		latestFxDate = &resp.LatestFxDate
	}

	return &model.IngestionStatus{
		Feeds:               feeds,
		TrackedInstruments:  int(resp.TrackedInstruments),
		TrackedCurrencies:   int(resp.TrackedCurrencies),
		LatestPriceDate:     latestPriceDate,
		LatestFxDate:        latestFxDate,
		RateLimitRemaining:  int(resp.RateLimitRemaining),
		RateLimitBudget:     int(resp.RateLimitBudget),
		PendingBackfillJobs: int(resp.PendingBackfillJobs),
	}
}

func toModelUserPreferences(p *userpb.UserPreferences) *model.UserPreferences {
	if p == nil {
		return nil
	}
	return &model.UserPreferences{
		UserID:          p.UserId,
		Email:           p.Email,
		DisplayName:     p.DisplayName,
		DisplayCurrency: p.DisplayCurrency,
		Theme:           p.Theme,
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
	}
}

func toModelCurrency(c *userpb.CurrencyInfo) *model.Currency {
	if c == nil {
		return nil
	}
	return &model.Currency{
		Code:   c.Code,
		Name:   c.Name,
		Symbol: c.Symbol,
	}
}

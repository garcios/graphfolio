package graph

import (
	"bff/graph/model"

	commonpb "graphfolio/proto/common/v1"
	pb "graphfolio/proto/portfolio/v1"
	userpb "graphfolio/proto/user/v1"

	"fmt"

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

func toNullableModelMoney(m *commonpb.Money) *model.Money {
	if m == nil || m.Amount == nil {
		return nil
	}
	d, err := decimal.NewFromString(m.GetAmount().GetValue())
	if err != nil || !d.IsPositive() {
		return nil
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
			ID:                  inv.Id,
			Ticker:              inv.Ticker,
			Name:                inv.Name,
			Price:               toModelMoney(inv.Price),
			AverageBuyPrice:     toNullableModelMoney(inv.AverageBuyPrice),
			Quantity:            toModelDecimal(inv.Quantity),
			TotalValue:          toModelMoney(inv.TotalValue),
			TodayReturnAmount:   toModelMoney(inv.TodayReturnAmount),
			TodayReturnPercent:  toModelDecimal(inv.TodayReturnPercent),
			TotalReturnAmount:   toModelMoney(inv.TotalReturnAmount),
			TotalReturnPercent:  toModelDecimal(inv.TotalReturnPercent),
			CapitalGainAmount:   toModelMoney(inv.CapitalGainAmount),
			CapitalGainPercent:  toModelDecimal(inv.CapitalGainPercent),
			IncomeAmount:        toModelMoney(inv.IncomeAmount),
			IncomeYieldPercent:  toModelDecimal(inv.IncomeYieldPercent),
			CurrencyGainAmount:  toModelMoney(inv.CurrencyGainAmount),
			CurrencyGainPercent: toModelDecimal(inv.CurrencyGainPercent),
			IsInternational:     inv.IsInternational,
			IsClosed:            inv.IsClosed,
		}
	}

	var inceptionDate *string
	if p.InceptionDate != "" {
		inceptionDate = &p.InceptionDate
	}

	return &model.Portfolio{
		TotalValue:              toModelMoney(p.TotalValue),
		TodayReturnAmount:       toModelMoney(p.TodayReturnAmount),
		TodayReturnPercent:      toModelDecimal(p.TodayReturnPercent),
		AnnualizedReturnPercent: toModelDecimal(p.AnnualizedReturnPercent),
		CashBalance:             toModelMoney(p.CashBalance),
		Investments:             investments,
		InceptionDate:           inceptionDate,
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
	case model.TransactionTypeSplit:
		return pb.TransactionType_TRANSACTION_TYPE_SPLIT
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

	var withholdingTax *model.Money
	if item.WithholdingTax != nil {
		withholdingTax = toModelMoney(item.WithholdingTax)
	}
	var fxRateToBase *model.Decimal
	if item.FxRateToBase != nil {
		d := toModelDecimal(item.FxRateToBase)
		fxRateToBase = &d
	}
	var externalRef *string
	if item.ExternalRef != "" {
		ref := item.ExternalRef
		externalRef = &ref
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
		WithholdingTax: withholdingTax,
		FxRateToBase:   fxRateToBase,
		ExternalRef:    externalRef,
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
	case pb.TransactionType_TRANSACTION_TYPE_SPLIT:
		return model.TransactionTypeSplit
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

func toModelUserRole(r userpb.UserRole) model.UserRole {
	if r == userpb.UserRole_USER_ROLE_ADMIN {
		return model.UserRoleAdmin
	}
	return model.UserRoleUser
}

func toModelUserPreferences(p *userpb.UserPreferences) *model.UserPreferences {
	if p == nil {
		return nil
	}
	return &model.UserPreferences{
		UserID:             p.UserId,
		Email:              p.Email,
		DisplayName:        p.DisplayName,
		DisplayCurrency:    p.DisplayCurrency,
		Theme:              p.Theme,
		Role:               toModelUserRole(p.Role),
		MustChangePassword: p.MustChangePassword,
		CreatedAt:          p.CreatedAt,
		UpdatedAt:          p.UpdatedAt,
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

func toModelCurrencyPair(p *pb.CurrencyPairItem) *model.CurrencyPair {
	if p == nil {
		return nil
	}
	var previousRate, change1dAmount, change1dPct *model.Decimal
	if p.PreviousRate != nil {
		pr := toModelDecimal(p.PreviousRate)
		previousRate = &pr
	}
	if p.Change_1DAmount != nil {
		ca := toModelDecimal(p.Change_1DAmount)
		change1dAmount = &ca
	}
	if p.Change_1DPct != nil {
		cp := toModelDecimal(p.Change_1DPct)
		change1dPct = &cp
	}

	return &model.CurrencyPair{
		BaseCurrency:   p.BaseCurrency,
		QuoteCurrency:  p.QuoteCurrency,
		Pair:           fmt.Sprintf("%s/%s", p.BaseCurrency, p.QuoteCurrency),
		LatestRate:     toModelDecimal(p.LatestRate),
		LatestDate:     p.LatestDate,
		LatestSource:   p.LatestSource,
		PreviousRate:   previousRate,
		Change1dAmount: change1dAmount,
		Change1dPct:    change1dPct,
		TotalRecords:   int(p.TotalRecords),
		FirstDate:      p.FirstDate,
		LastDate:       p.LastDate,
	}
}

func toModelFXRate(r *pb.FXRateItem) *model.FXRate {
	if r == nil {
		return nil
	}
	return &model.FXRate{
		BaseCurrency:  r.BaseCurrency,
		QuoteCurrency: r.QuoteCurrency,
		Pair:          fmt.Sprintf("%s/%s", r.BaseCurrency, r.QuoteCurrency),
		RateDate:      r.RateDate,
		Rate:          toModelDecimal(r.Rate),
		InvertedRate:  toModelDecimal(r.InvertedRate),
		Source:        r.Source,
	}
}

func toModelFXRatesConnection(resp *pb.ListFXRatesResponse) *model.FXRatesConnection {
	if resp == nil {
		return &model.FXRatesConnection{
			Items:      []*model.FXRate{},
			TotalCount: 0,
		}
	}
	items := make([]*model.FXRate, len(resp.Rates))
	for i, r := range resp.Rates {
		items[i] = toModelFXRate(r)
	}
	return &model.FXRatesConnection{
		Items:      items,
		TotalCount: int(resp.TotalCount),
	}
}

func toModelFXHistoryPoint(pt *pb.FXHistoryPoint) *model.FXHistoryPoint {
	if pt == nil {
		return nil
	}
	return &model.FXHistoryPoint{
		Date:         pt.Date,
		Rate:         toModelDecimal(pt.Rate),
		InvertedRate: toModelDecimal(pt.InvertedRate),
		Source:       pt.Source,
	}
}

func toModelCurrencyPairHistory(resp *pb.GetCurrencyPairHistoryResponse) *model.CurrencyPairHistory {
	if resp == nil {
		return nil
	}
	points := make([]*model.FXHistoryPoint, len(resp.Points))
	for i, pt := range resp.Points {
		points[i] = toModelFXHistoryPoint(pt)
	}
	return &model.CurrencyPairHistory{
		BaseCurrency:    resp.BaseCurrency,
		QuoteCurrency:   resp.QuoteCurrency,
		Pair:            fmt.Sprintf("%s/%s", resp.BaseCurrency, resp.QuoteCurrency),
		Points:          points,
		StartRate:       toModelDecimal(resp.StartRate),
		EndRate:         toModelDecimal(resp.EndRate),
		PeriodChange:    toModelDecimal(resp.PeriodChange),
		PeriodChangePct: toModelDecimal(resp.PeriodChangePct),
		PeriodHigh:      toModelDecimal(resp.PeriodHigh),
		PeriodLow:       toModelDecimal(resp.PeriodLow),
	}
}

func toModelExchange(e *pb.Exchange) *model.Exchange {
	if e == nil {
		return nil
	}
	return &model.Exchange{
		Code:     e.Code,
		Name:     e.Name,
		Country:  e.Country,
		Timezone: e.Timezone,
	}
}

func toProtoCashFlowTimeframe(tf model.CashFlowTimeframe) pb.CashFlowTimeframe {
	switch tf {
	case model.CashFlowTimeframeMtd:
		return pb.CashFlowTimeframe_CASH_FLOW_TIMEFRAME_MTD
	case model.CashFlowTimeframeYtd:
		return pb.CashFlowTimeframe_CASH_FLOW_TIMEFRAME_YTD
	case model.CashFlowTimeframeM1:
		return pb.CashFlowTimeframe_CASH_FLOW_TIMEFRAME_1M
	case model.CashFlowTimeframeM3:
		return pb.CashFlowTimeframe_CASH_FLOW_TIMEFRAME_3M
	case model.CashFlowTimeframeM6:
		return pb.CashFlowTimeframe_CASH_FLOW_TIMEFRAME_6M
	case model.CashFlowTimeframeY1:
		return pb.CashFlowTimeframe_CASH_FLOW_TIMEFRAME_1Y
	case model.CashFlowTimeframeAll:
		return pb.CashFlowTimeframe_CASH_FLOW_TIMEFRAME_ALL
	case model.CashFlowTimeframeCustom:
		return pb.CashFlowTimeframe_CASH_FLOW_TIMEFRAME_CUSTOM
	default:
		return pb.CashFlowTimeframe_CASH_FLOW_TIMEFRAME_YTD
	}
}

func toModelCashFlowReport(resp *pb.GetCashFlowReportResponse) *model.CashFlowReport {
	if resp == nil {
		return nil
	}

	summary := &model.CashFlowSummary{
		StartingCashBalance: toModelMoney(resp.GetSummary().GetStartingCashBalance()),
		TotalInflows:        toModelMoney(resp.GetSummary().GetTotalInflows()),
		TotalOutflows:       toModelMoney(resp.GetSummary().GetTotalOutflows()),
		NetCashFlow:         toModelMoney(resp.GetSummary().GetNetCashFlow()),
		EndingCashBalance:   toModelMoney(resp.GetSummary().GetEndingCashBalance()),
	}

	breakdown := &model.CashFlowCategoryBreakdown{
		Deposits:      toModelMoney(resp.GetBreakdown().GetDeposits()),
		Dividends:     toModelMoney(resp.GetBreakdown().GetDividends()),
		Interest:      toModelMoney(resp.GetBreakdown().GetInterest()),
		SalesProceeds: toModelMoney(resp.GetBreakdown().GetSalesProceeds()),
		Withdrawals:   toModelMoney(resp.GetBreakdown().GetWithdrawals()),
		Purchases:     toModelMoney(resp.GetBreakdown().GetPurchases()),
		Fees:          toModelMoney(resp.GetBreakdown().GetFees()),
		Taxes:         toModelMoney(resp.GetBreakdown().GetTaxes()),
	}

	items := make([]*model.CashFlowItem, len(resp.GetItems()))
	for i, item := range resp.GetItems() {
		flowDir := model.CashFlowDirectionInflow
		if item.GetFlowDirection() == "OUTFLOW" {
			flowDir = model.CashFlowDirectionOutflow
		}

		var symPtr, namePtr *string
		if item.GetSymbol() != "" {
			s := item.GetSymbol()
			symPtr = &s
		}
		if item.GetInstrumentName() != "" {
			n := item.GetInstrumentName()
			namePtr = &n
		}

		items[i] = &model.CashFlowItem{
			ID:             item.GetId(),
			EventDate:      item.GetEventDate(),
			Type:           toModelTransactionType(item.GetType()),
			FlowDirection:  flowDir,
			Category:       model.CashFlowCategory(item.GetCategory()),
			Symbol:         symPtr,
			InstrumentName: namePtr,
			Description:    item.GetDescription(),
			NetAmount:      toModelMoney(item.GetNetAmount()),
			RunningBalance: toModelMoney(item.GetRunningBalance()),
			LocalAmount:    toModelMoney(item.GetLocalAmount()),
			Fee:            toModelMoney(item.GetFee()),
			WithholdingTax: toModelMoney(item.GetWithholdingTax()),
		}
	}

	return &model.CashFlowReport{
		Summary:      summary,
		Breakdown:    breakdown,
		Items:        items,
		BaseCurrency: resp.GetBaseCurrency(),
		FromDate:     resp.GetFromDate(),
		ToDate:       resp.GetToDate(),
	}
}

func toModelMissingPriceReport(resp *pb.ListMissingPriceTransactionsResponse) *model.MissingPriceReport {
	if resp == nil {
		return &model.MissingPriceReport{
			Items:                  []*model.MissingPriceTransactionItem{},
			TotalCount:             0,
			UniqueInstrumentsCount: 0,
		}
	}

	items := make([]*model.MissingPriceTransactionItem, len(resp.Items))
	for i, item := range resp.Items {
		items[i] = &model.MissingPriceTransactionItem{
			TransactionID:  item.GetTransactionId(),
			PortfolioID:    item.GetPortfolioId(),
			InstrumentID:   item.GetInstrumentId(),
			Symbol:         item.GetSymbol(),
			AssetName:      item.GetAssetName(),
			ExchangeMic:    item.GetExchangeMic(),
			TradeDate:      item.GetTradeDate(),
			Type:           toModelTransactionType(item.GetType()),
			Quantity:       toModelDecimal(item.GetQuantity()),
			ExecutionPrice: toModelMoney(item.GetExecutionPrice()),
			CurrencyCode:   item.GetCurrencyCode(),
		}
	}

	return &model.MissingPriceReport{
		Items:                  items,
		TotalCount:             int(resp.GetTotalCount()),
		UniqueInstrumentsCount: int(resp.GetUniqueInstrumentsCount()),
		EarliestTradeDate:      resp.EarliestTradeDate,
		LatestTradeDate:        resp.LatestTradeDate,
	}
}

func toModelMissingFXRatesReport(resp *pb.ListMissingFXRateTransactionsResponse) *model.MissingFXRatesReport {
	if resp == nil {
		return &model.MissingFXRatesReport{
			Items:            []*model.MissingFXRateTransactionItem{},
			TotalCount:       0,
			UniquePairsCount: 0,
		}
	}

	items := make([]*model.MissingFXRateTransactionItem, len(resp.Items))
	for i, item := range resp.Items {
		var recordedRate *model.Decimal
		if item.RecordedFxRate != nil {
			r := toModelDecimal(item.GetRecordedFxRate())
			recordedRate = &r
		}


		items[i] = &model.MissingFXRateTransactionItem{
			TransactionID:         item.GetTransactionId(),
			PortfolioID:           item.GetPortfolioId(),
			PortfolioName:         item.GetPortfolioName(),
			PortfolioBaseCurrency: item.GetPortfolioBaseCur(),
			TradeDate:             item.GetTradeDate(),
			Type:                  toModelTransactionType(item.GetType()),
			InstrumentID:          item.InstrumentId,
			Symbol:                item.Symbol,
			AssetName:             item.AssetName,
			Amount:                toModelMoney(item.GetAmount()),
			TransactionCurrency:   item.GetTransactionCur(),
			RequiredPair:          item.GetRequiredPair(),
			RecordedFxRate:        recordedRate,
		}
	}

	return &model.MissingFXRatesReport{
		Items:             items,
		TotalCount:        int(resp.GetTotalCount()),
		UniquePairsCount:  int(resp.GetUniquePairsCount()),
		EarliestTradeDate: resp.EarliestTradeDate,
		LatestTradeDate:   resp.LatestTradeDate,
	}
}


func toModelModifiedDietzReturn(resp *pb.GetModifiedDietzReturnResponse) *model.ModifiedDietzReturnPayload {
	if resp == nil {
		return nil
	}

	weights := make([]*model.ExternalCashFlowWeightItem, len(resp.GetCashFlowWeights()))
	for i, w := range resp.GetCashFlowWeights() {
		weights[i] = &model.ExternalCashFlowWeightItem{
			TransactionID:  w.GetTransactionId(),
			Date:           w.GetDate(),
			Type:           toModelTransactionType(w.GetType()),
			Amount:         toModelMoney(w.GetAmount()),
			Weight:         toModelDecimal(w.GetWeight()),
			WeightedAmount: toModelMoney(w.GetWeightedAmount()),
			LocalAmount:    toModelMoney(w.GetLocalAmount()),
		}
	}

	var annReturn *model.Decimal
	if resp.AnnualizedReturnPercent != nil {
		d := toModelDecimal(resp.AnnualizedReturnPercent)
		annReturn = &d
	}

	return &model.ModifiedDietzReturnPayload{
		PortfolioID:             resp.GetPortfolioId(),
		BaseCurrency:            resp.GetBaseCurrency(),
		StartDate:               resp.GetStartDate(),
		EndDate:                 resp.GetEndDate(),
		ElapsedDays:             int(resp.GetElapsedDays()),
		StartValuation:          toModelMoney(resp.GetStartValuation()),
		EndValuation:            toModelMoney(resp.GetEndValuation()),
		NetExternalFlows:        toModelMoney(resp.GetNetExternalFlows()),
		NetInvestmentGain:       toModelMoney(resp.GetNetInvestmentGain()),
		WeightedCapitalEmployed: toModelMoney(resp.GetWeightedCapitalEmployed()),
		CumulativeReturnPercent: toModelDecimal(resp.GetCumulativeReturnPercent()),
		AnnualizedReturnPercent: annReturn,
		IsAnnualized:            resp.GetIsAnnualized(),
		IsComputable:            resp.GetIsComputable(),
		DiagnosticCode:          resp.GetDiagnosticCode(),
		CashFlowWeights:         weights,
	}
}

func toModelBenchmark(b *pb.Benchmark) *model.Benchmark {
	if b == nil {
		return nil
	}
	return &model.Benchmark{
		ID:           b.GetId(),
		Symbol:       b.GetSymbol(),
		Name:         b.GetName(),
		ExchangeCode: b.GetExchangeCode(),
		CurrencyCode: b.GetCurrencyCode(),
		IndexType:    b.GetIndexType(),
		IsDefault:    b.GetIsDefault(),
	}
}

func toProtoBenchmarkCurrencyMode(m *model.BenchmarkCurrencyMode) pb.BenchmarkCurrencyMode {
	if m == nil {
		return pb.BenchmarkCurrencyMode_BENCHMARK_CURRENCY_MODE_PORTFOLIO_BASE
	}
	switch *m {
	case model.BenchmarkCurrencyModeBenchmarkNative:
		return pb.BenchmarkCurrencyMode_BENCHMARK_CURRENCY_MODE_BENCHMARK_NATIVE
	case model.BenchmarkCurrencyModePortfolioBase:
		return pb.BenchmarkCurrencyMode_BENCHMARK_CURRENCY_MODE_PORTFOLIO_BASE
	default:
		return pb.BenchmarkCurrencyMode_BENCHMARK_CURRENCY_MODE_PORTFOLIO_BASE
	}
}

func toModelBenchmarkCurrencyMode(m pb.BenchmarkCurrencyMode) model.BenchmarkCurrencyMode {
	switch m {
	case pb.BenchmarkCurrencyMode_BENCHMARK_CURRENCY_MODE_BENCHMARK_NATIVE:
		return model.BenchmarkCurrencyModeBenchmarkNative
	default:
		return model.BenchmarkCurrencyModePortfolioBase
	}
}

func toModelBenchmarkComparison(resp *pb.GetBenchmarkComparisonResponse) *model.BenchmarkComparisonPayload {
	if resp == nil {
		return nil
	}

	points := make([]*model.BenchmarkComparisonPoint, len(resp.GetPoints()))
	for i, pt := range resp.GetPoints() {
		points[i] = &model.BenchmarkComparisonPoint{
			Date:                   pt.GetDate(),
			PortfolioReturnPercent: toModelDecimal(pt.GetPortfolioReturnPercent()),
			BenchmarkReturnPercent: toModelDecimal(pt.GetBenchmarkReturnPercent()),
			ActiveReturnPercent:    toModelDecimal(pt.GetActiveReturnPercent()),
			PortfolioValue:         toModelMoney(pt.GetPortfolioValue()),
			BenchmarkPriceAdjusted: toModelMoney(pt.GetBenchmarkPriceAdjusted()),
		}
	}

	var portAnn, benchAnn, activeAnn, beta, te, ir *model.Decimal
	if resp.GetMetrics().GetPortfolioAnnualizedReturnPercent() != nil {
		d := toModelDecimal(resp.GetMetrics().GetPortfolioAnnualizedReturnPercent())
		portAnn = &d
	}
	if resp.GetMetrics().GetBenchmarkAnnualizedReturnPercent() != nil {
		d := toModelDecimal(resp.GetMetrics().GetBenchmarkAnnualizedReturnPercent())
		benchAnn = &d
	}
	if resp.GetMetrics().GetActiveAnnualizedReturnPercent() != nil {
		d := toModelDecimal(resp.GetMetrics().GetActiveAnnualizedReturnPercent())
		activeAnn = &d
	}
	if resp.GetMetrics().GetBeta() != nil {
		d := toModelDecimal(resp.GetMetrics().GetBeta())
		beta = &d
	}
	if resp.GetMetrics().GetTrackingErrorPercent() != nil {
		d := toModelDecimal(resp.GetMetrics().GetTrackingErrorPercent())
		te = &d
	}
	if resp.GetMetrics().GetInformationRatio() != nil {
		d := toModelDecimal(resp.GetMetrics().GetInformationRatio())
		ir = &d
	}

	metrics := &model.BenchmarkMetrics{
		PortfolioCumulativeReturnPercent: toModelDecimal(resp.GetMetrics().GetPortfolioCumulativeReturnPercent()),
		BenchmarkCumulativeReturnPercent: toModelDecimal(resp.GetMetrics().GetBenchmarkCumulativeReturnPercent()),
		ActiveCumulativeReturnPercent:    toModelDecimal(resp.GetMetrics().GetActiveCumulativeReturnPercent()),
		PortfolioAnnualizedReturnPercent: portAnn,
		BenchmarkAnnualizedReturnPercent: benchAnn,
		ActiveAnnualizedReturnPercent:    activeAnn,
		IsAnnualized:                     resp.GetMetrics().GetIsAnnualized(),
		Beta:                             beta,
		TrackingErrorPercent:             te,
		InformationRatio:                 ir,
	}

	return &model.BenchmarkComparisonPayload{
		PortfolioID:  resp.GetPortfolioId(),
		BaseCurrency: resp.GetBaseCurrency(),
		Benchmark:    toModelBenchmark(resp.GetBenchmark()),
		CurrencyMode: toModelBenchmarkCurrencyMode(resp.GetCurrencyMode()),
		StartDate:    resp.GetStartDate(),
		EndDate:      resp.GetEndDate(),
		ElapsedDays:  int(resp.GetElapsedDays()),
		Points:       points,
		Metrics:      metrics,
	}
}


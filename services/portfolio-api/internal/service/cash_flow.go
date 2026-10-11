package service

import (
	"context"
	"fmt"
	"time"

	"portfolio-api/internal/domain"

	"github.com/shopspring/decimal"
)

func (s *portfolioService) GetCashFlowReport(ctx context.Context, input domain.CashFlowFilter) (*domain.CashFlowReport, error) {
	userID := input.UserID
	if userID == "" {
		userID = "1"
	}

	portfolio, err := s.repo.FindPortfolioByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("service: find portfolio: %w", err)
	}

	baseCurrency := portfolio.BaseCurrency
	if input.Currency != "" {
		baseCurrency = input.Currency
	}

	now := s.nowFunc().UTC()
	fromDate, toDate := resolveCashFlowWindow(input, now)

	// Fetch all transactions settled on or before toDate
	txsWithInst, err := s.repo.GetCashFlowTransactions(ctx, portfolio.ID, toDate)
	if err != nil {
		return nil, fmt.Errorf("service: get cash flow transactions: %w", err)
	}

	startingCash := decimal.Zero
	runningBalance := decimal.Zero

	var (
		totalInflows  decimal.Decimal
		totalOutflows decimal.Decimal

		depositsInflow  decimal.Decimal
		dividendsInflow decimal.Decimal
		interestInflow  decimal.Decimal
		salesInflow     decimal.Decimal

		withdrawalsOutflow decimal.Decimal
		purchasesOutflow   decimal.Decimal
		feesOutflow        decimal.Decimal
		taxesOutflow       decimal.Decimal
	)

	reportItems := make([]domain.CashFlowItem, 0)

	for _, twi := range txsWithInst {
		tx := twi.Transaction
		eventDate := tx.TradeDate
		if tx.SettleDate != nil && !tx.SettleDate.IsZero() {
			eventDate = *tx.SettleDate
		}
		eventDateUTC := eventDate.UTC().Truncate(24 * time.Hour)

		netImpact, flowDir, category, feeBase, taxBase := domain.CalculateTransactionCashImpact(tx, baseCurrency)

		// Before period window -> contribute to starting cash balance
		if !fromDate.IsZero() && eventDateUTC.Before(fromDate) {
			startingCash = startingCash.Add(netImpact)
			continue
		}

		// Within window [fromDate, toDate]
		if runningBalance.IsZero() && len(reportItems) == 0 {
			runningBalance = startingCash
		}
		runningBalance = runningBalance.Add(netImpact)

		// Aggregate inflow vs outflow
		if netImpact.IsPositive() {
			totalInflows = totalInflows.Add(netImpact)
			switch category {
			case domain.CategoryCapitalDeposits:
				depositsInflow = depositsInflow.Add(netImpact)
			case domain.CategoryDividends:
				dividendsInflow = dividendsInflow.Add(netImpact)
			case domain.CategoryInterest:
				interestInflow = interestInflow.Add(netImpact)
			case domain.CategorySaleProceeds:
				salesInflow = salesInflow.Add(netImpact)
			default:
				depositsInflow = depositsInflow.Add(netImpact)
			}
		} else if netImpact.IsNegative() {
			outflowAbs := netImpact.Abs()
			totalOutflows = totalOutflows.Add(outflowAbs)
			switch category {
			case domain.CategoryCapitalWithdrawals:
				withdrawalsOutflow = withdrawalsOutflow.Add(outflowAbs)
			case domain.CategoryPurchases:
				purchasesOutflow = purchasesOutflow.Add(outflowAbs)
			case domain.CategoryFees:
				feesOutflow = feesOutflow.Add(outflowAbs)
			case domain.CategoryTaxes:
				taxesOutflow = taxesOutflow.Add(outflowAbs)
			default:
				feesOutflow = feesOutflow.Add(outflowAbs)
			}
		}

		desc := formatTransactionDescription(twi)

		reportItems = append(reportItems, domain.CashFlowItem{
			ID:             tx.ID,
			EventDate:      eventDateUTC,
			Type:           tx.Type,
			FlowDirection:  flowDir,
			Category:       category,
			Symbol:         twi.Symbol,
			InstrumentName: twi.InstrumentName,
			Description:    desc,
			NetAmount:      domain.NewMoney(netImpact.Round(2), baseCurrency),
			RunningBalance: domain.NewMoney(runningBalance.Round(2), baseCurrency),
			LocalAmount:    domain.NewMoney(tx.Amount.Round(2), tx.CurrencyCode),
			Fee:            domain.NewMoney(feeBase.Round(2), baseCurrency),
			WithholdingTax: domain.NewMoney(taxBase.Round(2), baseCurrency),
		})
	}

	if len(reportItems) == 0 {
		runningBalance = startingCash
	}

	netCashFlow := totalInflows.Sub(totalOutflows)
	endingCash := startingCash.Add(netCashFlow)

	summary := domain.CashFlowSummary{
		StartingCashBalance: domain.NewMoney(startingCash.Round(2), baseCurrency),
		TotalInflows:        domain.NewMoney(totalInflows.Round(2), baseCurrency),
		TotalOutflows:       domain.NewMoney(totalOutflows.Round(2), baseCurrency),
		NetCashFlow:         domain.NewMoney(netCashFlow.Round(2), baseCurrency),
		EndingCashBalance:   domain.NewMoney(endingCash.Round(2), baseCurrency),
	}

	breakdown := domain.CashFlowCategoryBreakdown{
		Deposits:      domain.NewMoney(depositsInflow.Round(2), baseCurrency),
		Dividends:     domain.NewMoney(dividendsInflow.Round(2), baseCurrency),
		Interest:      domain.NewMoney(interestInflow.Round(2), baseCurrency),
		SalesProceeds: domain.NewMoney(salesInflow.Round(2), baseCurrency),
		Withdrawals:   domain.NewMoney(withdrawalsOutflow.Round(2), baseCurrency),
		Purchases:     domain.NewMoney(purchasesOutflow.Round(2), baseCurrency),
		Fees:          domain.NewMoney(feesOutflow.Round(2), baseCurrency),
		Taxes:         domain.NewMoney(taxesOutflow.Round(2), baseCurrency),
	}

	return &domain.CashFlowReport{
		Summary:      summary,
		Breakdown:    breakdown,
		Items:        reportItems,
		BaseCurrency: baseCurrency,
		FromDate:     fromDate,
		ToDate:       toDate,
	}, nil
}

func resolveCashFlowWindow(filter domain.CashFlowFilter, now time.Time) (time.Time, time.Time) {
	toDate := now.Truncate(24 * time.Hour)
	if !filter.ToDate.IsZero() {
		toDate = filter.ToDate.UTC().Truncate(24 * time.Hour)
	}

	var fromDate time.Time

	switch filter.Timeframe {
	case domain.CashFlowTimeframeMTD:
		fromDate = time.Date(toDate.Year(), toDate.Month(), 1, 0, 0, 0, 0, time.UTC)

	case domain.CashFlowTimeframeYTD:
		fromDate = time.Date(toDate.Year(), 1, 1, 0, 0, 0, 0, time.UTC)

	case domain.CashFlowTimeframe1M:
		fromDate = toDate.AddDate(0, -1, 0)

	case domain.CashFlowTimeframe3M:
		fromDate = toDate.AddDate(0, -3, 0)

	case domain.CashFlowTimeframe6M:
		fromDate = toDate.AddDate(0, -6, 0)

	case domain.CashFlowTimeframe1Y:
		fromDate = toDate.AddDate(-1, 0, 0)

	case domain.CashFlowTimeframeAll:
		fromDate = time.Time{}

	case domain.CashFlowTimeframeCustom:
		if !filter.FromDate.IsZero() {
			fromDate = filter.FromDate.UTC().Truncate(24 * time.Hour)
		} else {
			fromDate = toDate.AddDate(0, -1, 0)
		}

	default:
		// Default to YTD
		fromDate = time.Date(toDate.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	}

	return fromDate, toDate
}

func formatTransactionDescription(twi domain.TransactionWithInstrument) string {
	tx := twi.Transaction
	if tx.Notes != nil && *tx.Notes != "" {
		return *tx.Notes
	}

	sym := ""
	if twi.Symbol != nil && *twi.Symbol != "" {
		sym = *twi.Symbol
	}

	switch tx.Type {
	case domain.TxTypeDeposit:
		return "Capital Deposit"
	case domain.TxTypeWithdrawal:
		return "Capital Withdrawal"
	case domain.TxTypeDividend:
		if sym != "" {
			return fmt.Sprintf("Dividend from %s", sym)
		}
		return "Cash Dividend"
	case domain.TxTypeInterest:
		return "Cash Interest"
	case domain.TxTypeBuy:
		if sym != "" && tx.Quantity != nil && tx.Price != nil {
			return fmt.Sprintf("Buy %s %s @ %s", tx.Quantity.String(), sym, tx.Price.String())
		} else if sym != "" {
			return fmt.Sprintf("Purchase of %s", sym)
		}
		return "Security Purchase"
	case domain.TxTypeSell:
		if sym != "" && tx.Quantity != nil && tx.Price != nil {
			return fmt.Sprintf("Sell %s %s @ %s", tx.Quantity.String(), sym, tx.Price.String())
		} else if sym != "" {
			return fmt.Sprintf("Sale of %s", sym)
		}
		return "Security Sale"
	case domain.TxTypeFee:
		return "Fee Payment"
	case domain.TxTypeTax:
		return "Tax Payment"
	case domain.TxTypeTransferIn:
		if sym != "" {
			return fmt.Sprintf("Transfer In %s", sym)
		}
		return "Transfer In"
	case domain.TxTypeTransferOut:
		if sym != "" {
			return fmt.Sprintf("Transfer Out %s", sym)
		}
		return "Transfer Out"
	default:
		return string(tx.Type)
	}
}

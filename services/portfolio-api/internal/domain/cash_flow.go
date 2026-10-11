package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CashFlowTimeframe string

const (
	CashFlowTimeframeMTD    CashFlowTimeframe = "MTD"
	CashFlowTimeframeYTD    CashFlowTimeframe = "YTD"
	CashFlowTimeframe1M     CashFlowTimeframe = "1M"
	CashFlowTimeframe3M     CashFlowTimeframe = "3M"
	CashFlowTimeframe6M     CashFlowTimeframe = "6M"
	CashFlowTimeframe1Y     CashFlowTimeframe = "1Y"
	CashFlowTimeframeAll    CashFlowTimeframe = "ALL"
	CashFlowTimeframeCustom CashFlowTimeframe = "CUSTOM"
)

type CashFlowCategory string

const (
	CategoryCapitalDeposits    CashFlowCategory = "CAPITAL_DEPOSITS"
	CategoryDividends          CashFlowCategory = "DIVIDENDS"
	CategoryInterest           CashFlowCategory = "INTEREST"
	CategorySaleProceeds       CashFlowCategory = "SALE_PROCEEDS"
	CategoryCapitalWithdrawals CashFlowCategory = "CAPITAL_WITHDRAWALS"
	CategoryPurchases          CashFlowCategory = "PURCHASES"
	CategoryFees               CashFlowCategory = "FEES"
	CategoryTaxes              CashFlowCategory = "TAXES"
)

type CashFlowDirection string

const (
	FlowInflow  CashFlowDirection = "INFLOW"
	FlowOutflow CashFlowDirection = "OUTFLOW"
)

type CashFlowFilter struct {
	UserID    string
	Timeframe CashFlowTimeframe
	FromDate  time.Time
	ToDate    time.Time
	Currency  string
}

type CashFlowItem struct {
	ID             uuid.UUID
	EventDate      time.Time
	Type           TransactionType
	FlowDirection  CashFlowDirection
	Category       CashFlowCategory
	Symbol         *string
	InstrumentName *string
	Description    string
	NetAmount      Money // In base currency (+ for inflow, - for outflow)
	RunningBalance Money // Cumulative cash balance in base currency
	LocalAmount    Money // In transaction local currency
	Fee            Money // In base currency
	WithholdingTax Money // In base currency
}

type CashFlowSummary struct {
	StartingCashBalance Money
	TotalInflows        Money
	TotalOutflows       Money
	NetCashFlow         Money
	EndingCashBalance   Money
}

type CashFlowCategoryBreakdown struct {
	Deposits      Money
	Dividends     Money
	Interest      Money
	SalesProceeds Money
	Withdrawals   Money
	Purchases     Money
	Fees          Money
	Taxes         Money
}

type CashFlowReport struct {
	Summary      CashFlowSummary
	Breakdown    CashFlowCategoryBreakdown
	Items        []CashFlowItem
	BaseCurrency string
	FromDate     time.Time
	ToDate       time.Time
}

// NetCashImpact computes the base-currency net cash movement (inflow > 0, outflow < 0, neutral = 0).
func CalculateTransactionCashImpact(
	tx Transaction,
	baseCurrency string,
) (netImpact decimal.Decimal, direction CashFlowDirection, category CashFlowCategory, feeBase decimal.Decimal, taxBase decimal.Decimal) {
	fx := decimal.NewFromInt(1)
	if tx.FXRateToBase != nil && tx.FXRateToBase.IsPositive() {
		fx = *tx.FXRateToBase
	}

	feeFX := fx
	feeCurr := tx.CurrencyCode
	if tx.FeeCurrencyCode != nil && *tx.FeeCurrencyCode != "" {
		feeCurr = *tx.FeeCurrencyCode
		if feeCurr == baseCurrency {
			feeFX = decimal.NewFromInt(1)
		}
	}

	feeBase = tx.Fee.Mul(feeFX)
	taxBase = tx.WithholdingTax.Mul(fx)

	switch tx.Type {
	case TxTypeDeposit, TxTypeTransferIn:
		direction = FlowInflow
		category = CategoryCapitalDeposits
		amountBase := tx.Amount.Mul(fx)
		netImpact = amountBase.Sub(feeBase)

	case TxTypeWithdrawal, TxTypeTransferOut:
		direction = FlowOutflow
		category = CategoryCapitalWithdrawals
		amountBase := tx.Amount.Mul(fx)
		netImpact = amountBase.Add(feeBase).Neg()

	case TxTypeDividend:
		direction = FlowInflow
		category = CategoryDividends
		netIncomeLocal := tx.Amount.Sub(tx.WithholdingTax)
		netImpact = netIncomeLocal.Mul(fx)

	case TxTypeInterest:
		direction = FlowInflow
		category = CategoryInterest
		netIncomeLocal := tx.Amount.Sub(tx.WithholdingTax)
		netImpact = netIncomeLocal.Mul(fx)

	case TxTypeBuy:
		direction = FlowOutflow
		category = CategoryPurchases
		totalCostLocal := tx.Amount
		amountBase := totalCostLocal.Mul(fx)
		netImpact = amountBase.Add(feeBase).Neg()

	case TxTypeSell:
		direction = FlowInflow
		category = CategorySaleProceeds
		amountBase := tx.Amount.Mul(fx)
		netImpact = amountBase.Sub(feeBase)

	case TxTypeFee:
		direction = FlowOutflow
		category = CategoryFees
		netImpact = tx.Amount.Mul(fx).Neg()

	case TxTypeTax:
		direction = FlowOutflow
		category = CategoryTaxes
		netImpact = tx.Amount.Mul(fx).Neg()

	case TxTypeSplit:
		// Corporate stock splits have zero cash impact
		netImpact = decimal.Zero
		direction = FlowInflow
		category = CategoryCapitalDeposits

	default:
		netImpact = decimal.Zero
		direction = FlowInflow
		category = CategoryCapitalDeposits
	}

	return netImpact, direction, category, feeBase, taxBase
}

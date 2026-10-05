package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransactionType string

const (
	TxTypeBuy          TransactionType = "BUY"
	TxTypeSell         TransactionType = "SELL"
	TxTypeDividend     TransactionType = "DIVIDEND"
	TxTypeInterest     TransactionType = "INTEREST"
	TxTypeDeposit      TransactionType = "DEPOSIT"
	TxTypeWithdrawal   TransactionType = "WITHDRAWAL"
	TxTypeFee          TransactionType = "FEE"
	TxTypeTax          TransactionType = "TAX"
	TxTypeTransferIn   TransactionType = "TRANSFER_IN"
	TxTypeTransferOut  TransactionType = "TRANSFER_OUT"
	TxTypeFXConversion TransactionType = "FX_CONVERSION"
)

type Transaction struct {
	ID             uuid.UUID
	PortfolioID    uuid.UUID
	InstrumentID   *uuid.UUID
	Type           TransactionType
	TradeDate      time.Time
	Quantity       *decimal.Decimal
	Price          *decimal.Decimal
	Amount         decimal.Decimal
	CurrencyCode   string
	Fee            decimal.Decimal
	WithholdingTax decimal.Decimal
	FXRateToBase   *decimal.Decimal
	ExternalRef    *string
	Notes          *string
	CreatedAt      time.Time
}

type TransactionWithInstrument struct {
	Transaction
	Symbol         *string
	InstrumentName *string
}

type TransactionFilter struct {
	Type     *TransactionType
	Symbol   *string
	Page     int
	PageSize int
}

type AddTransactionInput struct {
	UserID       string
	Type         TransactionType
	Symbol       *string
	TradeDate    time.Time
	Quantity     *decimal.Decimal
	Price        *decimal.Decimal
	Amount       *decimal.Decimal
	CurrencyCode *string
	Fee          *decimal.Decimal
	Notes        *string
}

type TaxLot struct {
	ID                uuid.UUID
	PortfolioID       uuid.UUID
	InstrumentID      uuid.UUID
	OpenTransactionID uuid.UUID
	AcquiredDate      time.Time
	OriginalQuantity  decimal.Decimal
	RemainingQuantity decimal.Decimal
	CostBasis         decimal.Decimal
	CostBasisBase     decimal.Decimal
	ClosedDate        *time.Time
}

type LotDisposal struct {
	ID                    uuid.UUID
	TaxLotID              uuid.UUID
	SellTransactionID     uuid.UUID
	Quantity              decimal.Decimal
	CostBasisReleased     decimal.Decimal
	CostBasisReleasedBase decimal.Decimal
	ProceedsBase          decimal.Decimal
	RealizedPnLBase       decimal.Decimal
}

type CorporateActionType string

const (
	CATypeSplit           CorporateActionType = "SPLIT"
	CATypeReverseSplit    CorporateActionType = "REVERSE_SPLIT"
	CATypeSpinOff         CorporateActionType = "SPIN_OFF"
	CATypeMerger          CorporateActionType = "MERGER"
	CATypeSymbolChange    CorporateActionType = "SYMBOL_CHANGE"
	CATypeReturnOfCapital CorporateActionType = "RETURN_OF_CAPITAL"
)

type CorporateAction struct {
	ID              uuid.UUID
	InstrumentID    uuid.UUID
	Type            CorporateActionType
	ExDate          time.Time
	RatioFrom       decimal.Decimal
	RatioTo         decimal.Decimal
	NewInstrumentID *uuid.UUID
	CashAmount      *decimal.Decimal
	CurrencyCode    *string
	CostBasisPct    *decimal.Decimal
}

type Holding struct {
	PortfolioID     uuid.UUID
	InstrumentID    uuid.UUID
	Quantity        decimal.Decimal
	CostBasis       decimal.Decimal
	CostBasisBase   decimal.Decimal
	RealizedPnLBase decimal.Decimal
	DividendsBase   decimal.Decimal
}

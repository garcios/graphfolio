export type SupportedBroker = 'commsec' | 'nabtrade' | 'auto';

export type NormalizedTxType =
  | 'BUY'
  | 'SELL'
  | 'DIVIDEND'
  | 'DEPOSIT'
  | 'WITHDRAWAL'
  | 'INTEREST';

export interface NormalizedTransactionRow {
  rowNumber: number;
  externalRef: string;
  symbol: string;
  type: NormalizedTxType;
  tradeDate: string;        // YYYY-MM-DD
  settleDate?: string;      // YYYY-MM-DD
  quantity: string;         // Arbitrary-precision numeric string
  price: string;            // Unit price numeric string
  amount: string;           // Net consideration numeric string
  fee: string;              // Brokerage fee numeric string
  currencyCode: string;     // Default "AUD"
  notes?: string;           // Audit metadata (Broker name, original row)
  frankingCredit?: string;  // Explicit franking credit amount
  frankingPct?: string;     // Explicit franking percentage (0-100)
  isDuplicate?: boolean;    // Flagged by pre-flight check
}

export interface RowValidationError {
  rowNumber: number;
  field: string;
  rawValue: string;
  message: string;
}

export interface ParseResult {
  broker: 'commsec' | 'nabtrade';
  totalRowsRead: number;
  validRows: NormalizedTransactionRow[];
  errors: RowValidationError[];
}

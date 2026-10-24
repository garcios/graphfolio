import type { TransactionItem } from '../../components/TransactionLedger';

const TYPE_PRIORITY: Record<string, number> = {
  DEPOSIT: 1,
  TRANSFER_IN: 2,
  BUY: 3,
  SPLIT: 4,
  DIVIDEND: 5,
  INTEREST: 6,
  SELL: 7,
  TRANSFER_OUT: 8,
  FEE: 9,
  TAX: 10,
  WITHDRAWAL: 11,
};

/**
 * Escapes a single CSV field following RFC 4180 rules and formula injection prevention:
 * 1. Sanitizes dangerous formula prefixes (=, +, -, @, \t, \r) by prepending a single quote.
 * 2. Escapes double quotes by doubling them ("" -> """").
 * 3. Wraps value in double quotes if it contains commas, quotes, or newlines.
 */
export function escapeCSVField(value: string | number | null | undefined): string {
  if (value === null || value === undefined) {
    return '';
  }
  let str = String(value);

  // Formula injection prevention (CWE-1236):
  // If the cell begins with formula trigger characters (=, +, -, @, \t, \r), neutralize by prepending a single quote.
  // Standard signed numbers (e.g. -14500.00, +100) are safe numeric literals and should retain their numeric type.
  const isPureNumber = /^[-+]?\d+(\.\d+)?$/.test(str.trim());
  if (!isPureNumber && /^[=+\-@\t\r]/.test(str)) {
    str = `'${str}`;
  }

  // RFC 4180 escaping: wrap in quotes if it contains commas, quotes, or newlines
  if (str.includes('"') || str.includes(',') || str.includes('\n') || str.includes('\r')) {
    return `"${str.replace(/"/g, '""')}"`;
  }
  return str;
}

/**
 * Formats a list of transactions into an authoritative RFC 4180 CSV string.
 * Strictly guarantees chronological ascending ordering (oldest first).
 * Prepends UTF-8 BOM (\uFEFF) to guarantee clean rendering in Microsoft Excel across platforms.
 */
export function formatTransactionsCSV(transactions: TransactionItem[]): string {
  // Sort ascending by trade date with deterministic lifecycle and timestamp tie-breakers
  const sorted = [...transactions].sort((a, b) => {
    // 1. Primary: Chronological Trade Date (oldest first)
    const dateComp = a.tradeDate.localeCompare(b.tradeDate);
    if (dateComp !== 0) return dateComp;

    // 2. Secondary: Lifecycle Priority (Deposits before Buys before Sells)
    const priorityA = TYPE_PRIORITY[a.type] ?? 99;
    const priorityB = TYPE_PRIORITY[b.type] ?? 99;
    if (priorityA !== priorityB) {
      return priorityA - priorityB;
    }

    // 3. Tertiary: Ingestion Timestamp
    const createdComp = (a.createdAt || '').localeCompare(b.createdAt || '');
    if (createdComp !== 0) return createdComp;

    // 4. Deterministic Tie-Breaker: ID
    return (a.id || '').localeCompare(b.id || '');
  });

  const headers = [
    'Date',
    'Type',
    'Symbol',
    'Instrument Name',
    'Quantity',
    'Price',
    'Currency',
    'Gross Amount',
    'Net Amount',
    'Fee',
    'Fee Currency',
    'Withholding Tax',
    'Franking Credit',
    'Franking %',
    'FX Rate to Base',
    'External Reference',
    'Notes',
  ];

  const rows: string[] = [headers.join(',')];

  for (const tx of sorted) {
    const currency = tx.amount?.currencyCode || tx.price?.currencyCode || '';
    const feeCurrency = tx.fee?.currencyCode || currency;

    // Handle Split transactions (split ratio in quantity, 0 gross/net amounts)
    const isSplit = tx.type === 'SPLIT';
    const grossAmount = isSplit ? '0.00' : tx.amount?.amount || '0.00';

    // Calculate signed net cash flow (inflows positive, outflows negative)
    let netAmount = grossAmount;
    if (!isSplit) {
      const grossNum = parseFloat(grossAmount) || 0;
      if (['BUY', 'WITHDRAWAL', 'FEE', 'TAX'].includes(tx.type)) {
        netAmount = grossNum > 0 ? `-${grossAmount}` : grossAmount;
      }
    }

    const row = [
      escapeCSVField(tx.tradeDate),
      escapeCSVField(tx.type),
      escapeCSVField(tx.symbol || ''),
      escapeCSVField(tx.instrumentName || (tx.symbol ? '' : 'Cash')),
      escapeCSVField(tx.quantity || ''),
      escapeCSVField(isSplit ? '' : tx.price?.amount || ''),
      escapeCSVField(currency),
      escapeCSVField(grossAmount),
      escapeCSVField(netAmount),
      escapeCSVField(tx.fee?.amount || '0.00'),
      escapeCSVField(feeCurrency),
      escapeCSVField(tx.withholdingTax?.amount || '0.00'),
      escapeCSVField(tx.frankingCredit?.amount || '0.00'),
      escapeCSVField(tx.frankingPct || ''),
      escapeCSVField(tx.fxRateToBase || '1.000000'),
      escapeCSVField(tx.externalRef || ''),
      escapeCSVField(tx.notes || ''),
    ];
    rows.push(row.join(','));
  }

  // Prepend UTF-8 BOM
  return `\uFEFF${rows.join('\r\n')}`;
}

/**
 * Triggers a browser file download using a Blob and clean object URL revocation.
 */
export function downloadCSV(content: string, filename: string): void {
  const blob = new Blob([content], { type: 'text/csv;charset=utf-8;' });
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.setAttribute('download', filename);
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  setTimeout(() => {
    URL.revokeObjectURL(url);
  }, 100);
}

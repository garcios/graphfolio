import type { NormalizedTransactionRow, RowValidationError, ParseResult } from '../types';
import { cleanCurrency, parseAustralianDate, generateIdempotentHash, parseCSVLine } from '../utils';

export async function parseCommSecCSV(lines: string[], headerIdx: number): Promise<ParseResult> {
  const headers = parseCSVLine(lines[headerIdx]).map((h) => h.toLowerCase());

  const getCol = (cells: string[], names: string[]): string | undefined => {
    for (const name of names) {
      const idx = headers.indexOf(name);
      if (idx !== -1 && idx < cells.length) {
        return cells[idx]?.trim();
      }
    }
    return undefined;
  };

  const validRows: NormalizedTransactionRow[] = [];
  const errors: RowValidationError[] = [];

  for (let i = headerIdx + 1; i < lines.length; i++) {
    const rawLine = lines[i].trim();
    if (!rawLine) continue;

    const cleanCells = parseCSVLine(rawLine);
    const rowNum = i + 1;

    const rawDate = getCol(cleanCells, ['date', 'transaction date']);
    const rawType = getCol(cleanCells, ['type', 'transaction type']);
    const rawCode = getCol(cleanCells, ['code', 'asx code', 'security']);
    const rawUnits = getCol(cleanCells, ['units', 'quantity']);
    const rawPrice = getCol(cleanCells, ['unit price ($)', 'price', 'unit price']);
    const rawFee = getCol(cleanCells, ['brokerage ($)', 'brokerage (inc gst)', 'brokerage', 'fee']);
    const rawTotal = getCol(cleanCells, ['total value ($)', 'net consideration', 'value']);

    try {
      if (!rawDate) throw new Error('Missing transaction date');
      const tradeDate = parseAustralianDate(rawDate);

      if (!rawType) throw new Error('Missing transaction type');
      let type: 'BUY' | 'SELL' | 'DIVIDEND';
      const upperType = rawType.toUpperCase();
      if (upperType.includes('BUY')) type = 'BUY';
      else if (upperType.includes('SELL')) type = 'SELL';
      else if (upperType.includes('DIV') || upperType === 'DRP') type = 'DIVIDEND';
      else throw new Error(`Unsupported transaction type: "${rawType}"`);

      if (!rawCode) throw new Error('Missing ASX stock code');
      const symbol = rawCode.toUpperCase().replace(/\.AX$/i, '').trim();

      const quantity = cleanCurrency(rawUnits);
      if (type !== 'DIVIDEND' && (parseFloat(quantity) <= 0 || isNaN(parseFloat(quantity)))) {
        throw new Error(`Quantity must be greater than zero, got: "${rawUnits || ''}"`);
      }

      const price = cleanCurrency(rawPrice);
      if (type !== 'DIVIDEND' && (parseFloat(price) < 0 || isNaN(parseFloat(price)))) {
        throw new Error(`Price cannot be negative, got: "${rawPrice || ''}"`);
      }

      const fee = cleanCurrency(rawFee);
      let amount = cleanCurrency(rawTotal);

      // Reconstruct amount if missing
      if (parseFloat(amount) === 0 && type !== 'DIVIDEND') {
        const principal = parseFloat(quantity) * parseFloat(price);
        amount = (type === 'BUY' ? principal + parseFloat(fee) : principal - parseFloat(fee)).toFixed(2);
      }

      const hashSeed = `commsec:${tradeDate}:${symbol}:${type}:${quantity}:${amount}:${fee}`;
      const externalRef = `cs_${await generateIdempotentHash(hashSeed)}`;

      validRows.push({
        rowNumber: rowNum,
        externalRef,
        symbol,
        type,
        tradeDate,
        quantity: type === 'DIVIDEND' && parseFloat(quantity) === 0 ? '0' : quantity,
        price: type === 'DIVIDEND' && parseFloat(price) === 0 ? '0.00' : price,
        amount,
        fee,
        currencyCode: 'AUD',
        notes: `Imported from CommSec CSV (Row ${rowNum})`,
      });
    } catch (err: any) {
      errors.push({
        rowNumber: rowNum,
        field: 'Row',
        rawValue: rawLine,
        message: err?.message || 'Malformed row',
      });
    }
  }

  return {
    broker: 'commsec',
    totalRowsRead: lines.length - (headerIdx + 1),
    validRows,
    errors,
  };
}

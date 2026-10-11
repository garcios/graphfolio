import type { NormalizedTransactionRow, RowValidationError, ParseResult } from '../types';
import { cleanCurrency, parseAustralianDate, generateIdempotentHash, parseCSVLine } from '../utils';

const COMMSEC_TRADE_REGEX =
  /^([BS])\s+([0-9.]+)\s+([A-Za-z0-9.]+)\s+@\s+([0-9.]+)/i;

const COMMSEC_DIVIDEND_REGEX =
  /(?:DIVIDEND|DIV|DISTRIBUTION)\s*(?:[-–:]\s*)?([A-Za-z0-9.]+)?/i;

async function parseCommSecCashAccountCSV(
  lines: string[],
  headerIdx: number,
  headers: string[]
): Promise<ParseResult> {
  const dateIdx = headers.findIndex((h) => h === 'date' || h === 'transaction date');
  const refIdx = headers.findIndex((h) => h === 'reference' || h === 'ref');
  const detailsIdx = headers.findIndex((h) => h === 'details' || h === 'description');
  const debitIdx = headers.findIndex((h) => h === 'debit' || h.startsWith('debit'));
  const creditIdx = headers.findIndex((h) => h === 'credit' || h.startsWith('credit'));

  const validRows: NormalizedTransactionRow[] = [];
  const errors: RowValidationError[] = [];

  for (let i = headerIdx + 1; i < lines.length; i++) {
    const rawLine = lines[i].trim();
    if (!rawLine) continue;

    // Skip legal disclaimers or footer lines
    if (
      rawLine.toLowerCase().includes('commsec is a service provided') ||
      rawLine.startsWith('***')
    ) {
      break;
    }

    const cells = parseCSVLine(rawLine);
    const rowNum = i + 1;

    const rawDate = dateIdx !== -1 && dateIdx < cells.length ? cells[dateIdx]?.trim() : '';
    const rawRef = refIdx !== -1 && refIdx < cells.length ? cells[refIdx]?.trim() : '';
    const rawDetails = detailsIdx !== -1 && detailsIdx < cells.length ? cells[detailsIdx]?.trim() : '';
    const rawDebit = debitIdx !== -1 && debitIdx < cells.length ? cells[debitIdx]?.trim() : '';
    const rawCredit = creditIdx !== -1 && creditIdx < cells.length ? cells[creditIdx]?.trim() : '';

    // Ignore informational rows (e.g. Opening Balance, Closing Balance)
    const lowerDetails = rawDetails.toLowerCase();
    if (
      lowerDetails.startsWith('opening balance') ||
      lowerDetails.startsWith('closing balance') ||
      (!rawDebit && !rawCredit && !rawDetails)
    ) {
      continue;
    }

    if (!rawDate) {
      errors.push({
        rowNumber: rowNum,
        field: 'Date',
        rawValue: rawLine,
        message: 'Missing transaction date',
      });
      continue;
    }

    let tradeDate: string;
    try {
      tradeDate = parseAustralianDate(rawDate);
    } catch (err: any) {
      errors.push({
        rowNumber: rowNum,
        field: 'Date',
        rawValue: rawDate,
        message: err?.message || 'Invalid date format',
      });
      continue;
    }

    const debitVal = cleanCurrency(rawDebit);
    const creditVal = cleanCurrency(rawCredit);
    const hasDebit = parseFloat(debitVal) > 0;
    const hasCredit = parseFloat(creditVal) > 0;

    if (!hasDebit && !hasCredit) {
      // Neither debit nor credit: informational row
      continue;
    }

    try {
      // 1. Check for Equity/ETF Trade row: "B 4 NDQ @ 62.806028" or "S 125 NDQ @ 63.680000"
      const tradeMatch = COMMSEC_TRADE_REGEX.exec(rawDetails);
      if (tradeMatch) {
        const action = tradeMatch[1].toUpperCase() === 'B' ? 'BUY' : 'SELL';
        const quantity = tradeMatch[2];
        const symbol = tradeMatch[3].toUpperCase().replace(/\.AX$/i, '').trim();
        let price = cleanCurrency(tradeMatch[4]);
        if (price.includes('.')) {
          price = price.replace(/0+$/, '');
          if (price.endsWith('.')) price += '00';
          else if (price.split('.')[1].length === 1) price += '0';
        }

        const principal = parseFloat(quantity) * parseFloat(price);
        let fee = '0.00';
        let amount = principal.toFixed(2);

        if (action === 'BUY') {
          if (hasDebit) {
            const debitNum = parseFloat(debitVal);
            const feeNum = Math.max(0, debitNum - principal);
            if (feeNum > 0.005) {
              fee = feeNum.toFixed(2);
              amount = principal.toFixed(2);
            } else {
              amount = debitVal;
              fee = '0.00';
            }
          }
        } else {
          // SELL
          if (hasCredit) {
            const creditNum = parseFloat(creditVal);
            const feeNum = Math.max(0, principal - creditNum);
            if (feeNum > 0.005) {
              fee = feeNum.toFixed(2);
              amount = principal.toFixed(2);
            } else {
              amount = creditVal;
              fee = '0.00';
            }
          }
        }

        let externalRef: string;
        if (rawRef) {
          externalRef = `commsec:${rawRef}`;
        } else {
          const hashSeed = `commsec:${tradeDate}:${symbol}:${action}:${quantity}:${amount}:${fee}`;
          externalRef = `cs_${await generateIdempotentHash(hashSeed)}`;
        }

        validRows.push({
          rowNumber: rowNum,
          externalRef,
          symbol,
          type: action,
          tradeDate,
          quantity,
          price,
          amount,
          fee,
          currencyCode: 'AUD',
          notes: rawRef
            ? `Imported from CommSec Cash Account (Order: ${rawRef})`
            : `Imported from CommSec Cash Account (${action} ${symbol})`,
        });
        continue;
      }

      // 2. Check for Dividends
      if (
        (lowerDetails.includes('dividend') || lowerDetails.includes('distribution')) &&
        hasCredit
      ) {
        const divMatch = COMMSEC_DIVIDEND_REGEX.exec(rawDetails);
        const divSymbol = divMatch && divMatch[1] ? divMatch[1].toUpperCase().replace(/\.AX$/i, '').trim() : '';

        let externalRef: string;
        if (rawRef) {
          externalRef = `commsec:div:${rawRef}`;
        } else {
          const hashSeed = `commsec:div:${tradeDate}:${divSymbol}:${creditVal}:${rowNum}`;
          externalRef = `cs_div_${await generateIdempotentHash(hashSeed)}`;
        }

        validRows.push({
          rowNumber: rowNum,
          externalRef,
          symbol: divSymbol,
          type: 'DIVIDEND',
          tradeDate,
          quantity: '0',
          price: '0.00',
          amount: creditVal,
          fee: '0.00',
          currencyCode: 'AUD',
          notes: rawRef
            ? `Imported from CommSec Cash Account (Dividend: ${rawRef})`
            : 'Imported from CommSec Cash Account (Dividend)',
        });
        continue;
      }

      // 3. Check for Interest
      if (lowerDetails.includes('interest') && hasCredit) {
        let externalRef: string;
        if (rawRef) {
          externalRef = `commsec:interest:${rawRef}`;
        } else {
          const hashSeed = `commsec:interest:${tradeDate}:${creditVal}:${rowNum}`;
          externalRef = `cs_int_${await generateIdempotentHash(hashSeed)}`;
        }

        validRows.push({
          rowNumber: rowNum,
          externalRef,
          symbol: '',
          type: 'INTEREST',
          tradeDate,
          quantity: '0',
          price: '0.00',
          amount: creditVal,
          fee: '0.00',
          currencyCode: 'AUD',
          notes: 'Imported from CommSec Cash Account (Interest)',
        });
        continue;
      }

      // 4. Cash Inflow -> DEPOSIT
      if (hasCredit && !hasDebit) {
        let externalRef: string;
        if (rawRef) {
          externalRef = `commsec:cash:${rawRef}`;
        } else {
          const hashSeed = `commsec:cash:dep:${tradeDate}:${creditVal}:${rowNum}`;
          externalRef = `cs_dep_${await generateIdempotentHash(hashSeed)}`;
        }

        validRows.push({
          rowNumber: rowNum,
          externalRef,
          symbol: '',
          type: 'DEPOSIT',
          tradeDate,
          quantity: '0',
          price: '0.00',
          amount: creditVal,
          fee: '0.00',
          currencyCode: 'AUD',
          notes: rawRef
            ? `Imported from CommSec Cash Account (Deposit: ${rawRef})`
            : 'Imported from CommSec Cash Account (Deposit)',
        });
        continue;
      }

      // 5. Cash Outflow -> WITHDRAWAL
      if (hasDebit && !hasCredit) {
        let externalRef: string;
        if (rawRef) {
          externalRef = `commsec:cash:${rawRef}`;
        } else {
          const hashSeed = `commsec:cash:wth:${tradeDate}:${debitVal}:${rowNum}`;
          externalRef = `cs_wth_${await generateIdempotentHash(hashSeed)}`;
        }

        validRows.push({
          rowNumber: rowNum,
          externalRef,
          symbol: '',
          type: 'WITHDRAWAL',
          tradeDate,
          quantity: '0',
          price: '0.00',
          amount: debitVal,
          fee: '0.00',
          currencyCode: 'AUD',
          notes: rawRef
            ? `Imported from CommSec Cash Account (Withdrawal: ${rawRef})`
            : 'Imported from CommSec Cash Account (Withdrawal)',
        });
        continue;
      }

      throw new Error(`Unrecognized CommSec cash transaction: "${rawDetails}"`);
    } catch (err: any) {
      errors.push({
        rowNumber: rowNum,
        field: 'Row',
        rawValue: rawLine,
        message: err?.message || 'Malformed row',
      });
    }
  }

  // Sort chronologically ascending; if dates match, sort by original row number descending (file is latest-first)
  validRows.sort((a, b) => {
    const cmp = a.tradeDate.localeCompare(b.tradeDate);
    if (cmp !== 0) return cmp;
    return b.rowNumber - a.rowNumber;
  });

  return {
    broker: 'commsec',
    totalRowsRead: lines.length - (headerIdx + 1),
    validRows,
    errors,
  };
}

async function parseCommSecOrdersCSV(
  lines: string[],
  headerIdx: number,
  headers: string[]
): Promise<ParseResult> {
  const getCol = (cells: string[], names: string[]): string | undefined => {
    for (const name of names) {
      const normalizedName = name.toLowerCase().replace(/\s*\(\$\)\s*$/, '').trim();
      const idx = headers.indexOf(normalizedName);
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

export async function parseCommSecCSV(lines: string[], headerIdx: number): Promise<ParseResult> {
  const rawHeaders = parseCSVLine(lines[headerIdx]);
  const headers = rawHeaders.map((h) => h.toLowerCase().replace(/\s*\(\$\)\s*$/, '').trim());

  const isCashStatement =
    (headers.includes('details') || headers.includes('description')) &&
    headers.includes('balance') &&
    (headers.includes('debit') || headers.includes('credit'));

  if (isCashStatement) {
    return parseCommSecCashAccountCSV(lines, headerIdx, headers);
  }

  return parseCommSecOrdersCSV(lines, headerIdx, headers);
}

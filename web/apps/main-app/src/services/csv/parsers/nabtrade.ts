import type { NormalizedTransactionRow, RowValidationError, ParseResult } from '../types';
import { cleanCurrency, parseAustralianDate, generateIdempotentHash, parseCSVLine } from '../utils';

const TRADE_REGEX =
  /^(BUY|SELL)\s+([A-Z0-9]+)\.([A-Z]+)\s+([0-9.]+)\s+([A-Z]{3})\s+([0-9.]+)\s+([0-9]+)\s+([A-Z0-9-]+)(?:\s+([0-9.]+))?/i;
const DIV_REGEX =
  /^DIVIDEND\s+on\s+([A-Z0-9]+)\.([A-Z]+)(?:\s+\(WHT\s+of\s+([A-Z]{3})\s+([+-]?[0-9.]+)\))?(?:\s*-\s*([A-Z]{3})\s+to\s+([A-Z]{3})\s+@\s+([0-9.]+))?/i;

const CASH_TRANSFER_REGEX =
  /^nabtrade:\s*([0-9]+)\s+FUNDS\s+TRANSFER/i;

const DOMESTIC_DIV_REGEX =
  /^FUNDS\s+TRANSFER\s+DIVIDEND\s*-\s*(.+)/i;

const NAME_TO_TICKER: Record<string, string> = {
  'GENESIS ENERGY': 'GNE',
  'WOODSIDE': 'WDS',
  'RESMED': 'RMD',
  'ATLAS ARTERIA': 'ALX',
  'WESFARMERS': 'WES',
  'COMMONWEALTH BAN': 'CBA',
  'NAB INTERIM': 'NAB',
  'NAB FINAL': 'NAB',
  'ANZ ': 'ANZ',
  'SONIC HEALTHCARE': 'SHL',
  'ACCENT GROUP': 'AX1',
  'AGL PAYMENT': 'AGL',
  'DRR PAYMENT': 'DRR',
};

interface TradeGroup {
  action: 'BUY' | 'SELL';
  symbol: string;
  exchange: string;
  quantity: string;
  currency: string;
  price: string;
  confirm: string;
  account: string;
  fxRate?: string;
  debitRow?: { date: string; amount: string; rowNum: number };
  creditRow?: { date: string; amount: string; rowNum: number };
}

interface DivRow {
  rowNumber: number;
  symbol: string;
  tradeDate: string;
  creditAmt: string;
  fxRate?: string;
  whtAmt?: string;
  rawDesc: string;
}

async function parseNabtradeInternationalCSV(
  lines: string[],
  headerIdx: number,
  headers: string[]
): Promise<ParseResult> {
  const dateIdx = headers.indexOf('date');
  const typeIdx = headers.indexOf('type');
  const descIdx = headers.indexOf('description');
  const debitIdx = headers.indexOf('debit');
  const creditIdx = headers.indexOf('credit');

  const tradeGroups = new Map<string, TradeGroup>();
  const divRows: DivRow[] = [];
  const cashRows: NormalizedTransactionRow[] = [];
  const errors: RowValidationError[] = [];

  for (let i = headerIdx + 1; i < lines.length; i++) {
    const rawLine = lines[i].trim();
    if (!rawLine) continue;

    // Terminate on legal footer notes
    if (rawLine.toLowerCase().includes('nabtrade is a service provided') || rawLine.startsWith('***')) {
      break;
    }

    const cells = parseCSVLine(rawLine);
    const rowNum = i + 1;

    const rawDate = dateIdx !== -1 && dateIdx < cells.length ? cells[dateIdx] : '';
    const rawType = typeIdx !== -1 && typeIdx < cells.length ? cells[typeIdx].trim() : '';
    const rawDesc = descIdx !== -1 && descIdx < cells.length ? cells[descIdx] : '';
    const rawDebit = debitIdx !== -1 && debitIdx < cells.length ? cells[debitIdx] : '';
    const rawCredit = creditIdx !== -1 && creditIdx < cells.length ? cells[creditIdx] : '';

    // Ignore informational / non-transactional rows
    if (
      rawType.toLowerCase() === 'interestchange' ||
      rawDesc.toLowerCase().startsWith('please note from')
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

    let date: string;
    try {
      date = parseAustralianDate(rawDate);
    } catch (err: any) {
      errors.push({
        rowNumber: rowNum,
        field: 'Date',
        rawValue: rawDate,
        message: err?.message || 'Invalid date format',
      });
      continue;
    }

    const tradeMatch = TRADE_REGEX.exec(rawDesc);
    if (tradeMatch) {
      const [, action, rawSymbol, exch, qty, curr, price, confirm, acct, fx] = tradeMatch;
      if (!tradeGroups.has(confirm)) {
        tradeGroups.set(confirm, {
          action: action.toUpperCase() as 'BUY' | 'SELL',
          symbol: rawSymbol.toUpperCase(),
          exchange: exch.toUpperCase(),
          quantity: qty,
          currency: curr.toUpperCase(),
          price,
          confirm,
          account: acct,
          fxRate: fx,
        });
      }
      const group = tradeGroups.get(confirm)!;
      const debitVal = cleanCurrency(rawDebit);
      const creditVal = cleanCurrency(rawCredit);

      if (parseFloat(debitVal) > 0) {
        group.debitRow = { date, amount: debitVal, rowNum };
      }
      if (parseFloat(creditVal) > 0) {
        group.creditRow = { date, amount: creditVal, rowNum };
      }
      continue;
    }

    const divMatch = DIV_REGEX.exec(rawDesc);
    if (divMatch) {
      const [, rawSymbol, , , whtAmt, , , fxRate] = divMatch;
      const creditVal = cleanCurrency(rawCredit);
      const debitVal = cleanCurrency(rawDebit);

      if (parseFloat(creditVal) > 0) {
        divRows.push({
          rowNumber: rowNum,
          symbol: rawSymbol.toUpperCase(),
          tradeDate: date,
          creditAmt: creditVal,
          fxRate,
          whtAmt,
          rawDesc,
        });
      } else if (parseFloat(debitVal) > 0) {
        // Cash sweep leg - paired with the dividend credit leg
      }
      continue;
    }

    // Cash transfers (Deposits / Withdrawals): nabtrade: 24461797 FUNDS TRANSFER 083543 200538980 deposit Oscar
    const cashTransferMatch = CASH_TRANSFER_REGEX.exec(rawDesc);
    if (cashTransferMatch) {
      const refNum = cashTransferMatch[1];
      const debitVal = cleanCurrency(rawDebit);
      const creditVal = cleanCurrency(rawCredit);

      if (parseFloat(creditVal) > 0) {
        cashRows.push({
          rowNumber: rowNum,
          externalRef: refNum ? `nabtrade:cash:${refNum}` : `nt_dep_${await generateIdempotentHash(`nabtrade:cash:dep:${date}:${creditVal}:${rowNum}`)}`,
          symbol: '',
          type: 'DEPOSIT',
          tradeDate: date,
          settleDate: undefined,
          quantity: '0',
          price: '0.00',
          amount: creditVal,
          fee: '0.00',
          currencyCode: 'AUD',
          notes: refNum ? `Imported from nabtrade Cash Account (Deposit: ${refNum})` : 'Imported from nabtrade Cash Account (Deposit)',
        });
        continue;
      } else if (parseFloat(debitVal) > 0) {
        cashRows.push({
          rowNumber: rowNum,
          externalRef: refNum ? `nabtrade:cash:${refNum}` : `nt_wth_${await generateIdempotentHash(`nabtrade:cash:wth:${date}:${debitVal}:${rowNum}`)}`,
          symbol: '',
          type: 'WITHDRAWAL',
          tradeDate: date,
          settleDate: undefined,
          quantity: '0',
          price: '0.00',
          amount: debitVal,
          fee: '0.00',
          currencyCode: 'AUD',
          notes: refNum ? `Imported from nabtrade Cash Account (Withdrawal: ${refNum})` : 'Imported from nabtrade Cash Account (Withdrawal)',
        });
        continue;
      }
    }

    // Monthly cash account interest
    if (rawType.toLowerCase() === 'interest' || rawDesc.toUpperCase() === 'INTEREST') {
      const creditVal = cleanCurrency(rawCredit);
      if (parseFloat(creditVal) > 0) {
        const hashSeed = `nabtrade:cash:interest:${date}:${creditVal}:${rowNum}`;
        const hash = await generateIdempotentHash(hashSeed);
        cashRows.push({
          rowNumber: rowNum,
          externalRef: `nt_int_${hash}`,
          symbol: '',
          type: 'INTEREST',
          tradeDate: date,
          settleDate: undefined,
          quantity: '0',
          price: '0.00',
          amount: creditVal,
          fee: '0.00',
          currencyCode: 'AUD',
          notes: 'Imported from nabtrade Cash Account (Interest)',
        });
        continue;
      }
    }

    // Domestic ETF/Equity dividends: FUNDS TRANSFER DIVIDEND - IVV PAYMENT OCT26/00877665
    const domDivMatch = DOMESTIC_DIV_REGEX.exec(rawDesc);
    if (domDivMatch) {
      const rest = domDivMatch[1].trim();
      const upperRest = rest.toUpperCase();
      let symbol = '';
      for (const [name, sym] of Object.entries(NAME_TO_TICKER)) {
        if (upperRest.startsWith(name)) {
          symbol = sym;
          break;
        }
      }
      if (!symbol) {
        const firstWord = rest.split(/[\s\-/]/)[0].toUpperCase();
        if (/^[A-Z0-9]{3,5}$/.test(firstWord)) {
          symbol = firstWord;
        }
      }

      const creditVal = cleanCurrency(rawCredit);
      if (parseFloat(creditVal) > 0) {
        const hashSeed = `nabtrade:cash:div:${date}:${symbol || 'DOM'}:${creditVal}:${rowNum}`;
        const hash = await generateIdempotentHash(hashSeed);
        cashRows.push({
          rowNumber: rowNum,
          externalRef: `nt_div_${hash}`,
          symbol,
          type: 'DIVIDEND',
          tradeDate: date,
          settleDate: undefined,
          quantity: '0',
          price: '0.00',
          amount: creditVal,
          fee: '0.00',
          currencyCode: 'AUD',
          notes: `Imported from nabtrade Cash Account Dividend (${rest})`,
        });
        continue;
      }
    }

    // Remediation payments / special cash credit adjustments
    if (rawDesc.toUpperCase().includes('REMEDIATION')) {
      const creditVal = cleanCurrency(rawCredit);
      if (parseFloat(creditVal) > 0) {
        const hashSeed = `nabtrade:cash:rem:${date}:${creditVal}:${rowNum}`;
        const hash = await generateIdempotentHash(hashSeed);
        cashRows.push({
          rowNumber: rowNum,
          externalRef: `nt_rem_${hash}`,
          symbol: '',
          type: 'INTEREST',
          tradeDate: date,
          settleDate: undefined,
          quantity: '0',
          price: '0.00',
          amount: creditVal,
          fee: '0.00',
          currencyCode: 'AUD',
          notes: `Imported from nabtrade Cash Account (Remediation: ${rawDesc})`,
        });
        continue;
      }
    }

    errors.push({
      rowNumber: rowNum,
      field: 'Description',
      rawValue: rawLine,
      message: `Unrecognized nabtrade cash account transaction description: "${rawDesc}"`,
    });
  }

  const validRows: NormalizedTransactionRow[] = [...cashRows];

  for (const group of tradeGroups.values()) {
    let tradeDate = '';
    let settleDate: string | undefined = undefined;
    let rowNumber = 0;

    if (group.action === 'BUY') {
      tradeDate = group.debitRow ? group.debitRow.date : group.creditRow?.date || '';
      if (group.creditRow && group.debitRow && group.creditRow.date !== group.debitRow.date) {
        settleDate = group.creditRow.date;
      }
      rowNumber = group.debitRow ? group.debitRow.rowNum : group.creditRow?.rowNum || 0;
    } else {
      tradeDate = group.creditRow ? group.creditRow.date : group.debitRow?.date || '';
      if (group.debitRow && group.creditRow && group.debitRow.date !== group.creditRow.date) {
        settleDate = group.debitRow.date;
      }
      rowNumber = group.creditRow ? group.creditRow.rowNum : group.debitRow?.rowNum || 0;
    }

    const amount = (parseFloat(group.quantity) * parseFloat(group.price)).toFixed(2);
    let fee = '0.00';
    if (group.action === 'BUY' && group.debitRow) {
      if (group.fxRate) {
        const principalAud = (parseFloat(group.quantity) * parseFloat(group.price)) / parseFloat(group.fxRate);
        const feeAud = Math.max(0, parseFloat(group.debitRow.amount) - principalAud);
        if (feeAud > 0.05) {
          fee = (feeAud * parseFloat(group.fxRate)).toFixed(2);
        }
      } else {
        const principal = parseFloat(group.quantity) * parseFloat(group.price);
        const feeAud = Math.max(0, parseFloat(group.debitRow.amount) - principal);
        if (feeAud > 0.05) {
          fee = feeAud.toFixed(2);
        }
      }
    } else if (group.action === 'SELL' && group.creditRow) {
      if (!group.fxRate) {
        const principal = parseFloat(group.quantity) * parseFloat(group.price);
        const feeAud = Math.max(0, principal - parseFloat(group.creditRow.amount));
        if (feeAud > 0.05) {
          fee = feeAud.toFixed(2);
        }
      }
    }

    const notes = group.fxRate
      ? `Imported from nabtrade Cash Account (Order: ${group.confirm}, FX: ${group.fxRate})`
      : `Imported from nabtrade Cash Account (Order: ${group.confirm})`;

    validRows.push({
      rowNumber,
      externalRef: `nabtrade:${group.confirm}`,
      symbol: group.symbol,
      type: group.action,
      tradeDate,
      settleDate,
      quantity: group.quantity,
      price: group.price,
      amount,
      fee,
      currencyCode: group.currency,
      notes,
    });
  }

  for (const div of divRows) {
    let amount = div.creditAmt;
    if (div.fxRate && parseFloat(div.fxRate) > 0) {
      amount = (parseFloat(div.creditAmt) / parseFloat(div.fxRate)).toFixed(2);
    }
    const hashSeed = `nabtrade:${div.tradeDate}:${div.symbol}:DIVIDEND:${div.creditAmt}`;
    const hash = await generateIdempotentHash(hashSeed);
    const externalRef = `nt_${hash}`;

    validRows.push({
      rowNumber: div.rowNumber,
      externalRef,
      symbol: div.symbol,
      type: 'DIVIDEND',
      tradeDate: div.tradeDate,
      settleDate: undefined,
      quantity: '0',
      price: '0.00',
      amount,
      fee: '0.00',
      currencyCode: 'USD',
      notes: `Imported from nabtrade International Dividend (${div.whtAmt ? 'WHT: USD ' + div.whtAmt + ' ' : ''}@ ${div.fxRate || '1.0'})`,
    });
  }

  // Chronological sorting
  validRows.sort((a, b) => a.tradeDate.localeCompare(b.tradeDate));

  return {
    broker: 'nabtrade',
    totalRowsRead: lines.length - (headerIdx + 1),
    validRows,
    errors,
  };
}

export async function parseNabtradeCSV(lines: string[], headerIdx: number): Promise<ParseResult> {
  const headers = parseCSVLine(lines[headerIdx]).map((h) => h.toLowerCase());

  const isInternational =
    headers.includes('description') &&
    headers.includes('balance') &&
    (headers.includes('debit') || headers.includes('credit'));

  if (isInternational) {
    return parseNabtradeInternationalCSV(lines, headerIdx, headers);
  }

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

    // Terminate on legal footer notes
    if (rawLine.toLowerCase().includes('nabtrade is a service provided') || rawLine.startsWith('***')) {
      break;
    }

    const cells = parseCSVLine(rawLine);
    const rowNum = i + 1;

    const rawConfirm = getCol(cells, ['confirmation number', 'ref number', 'trade ref']);
    const rawTradeDate = getCol(cells, ['trade date', 'date']);
    const rawSettleDate = getCol(cells, ['settlement date', 'settle date']);
    const rawType = getCol(cells, ['transaction type', 'type', 'action']);
    const rawCode = getCol(cells, ['code', 'security code', 'symbol']);
    const rawQty = getCol(cells, ['quantity', 'units']);
    const rawPrice = getCol(cells, ['price', 'unit price', 'average price']);
    const rawBrokerage = getCol(cells, ['brokerage', 'fees', 'brokerage (inc. gst)']);
    const rawConsideration = getCol(cells, ['total consideration', 'net amount', 'gross amount']);

    try {
      if (!rawTradeDate) throw new Error('Missing trade date');
      const tradeDate = parseAustralianDate(rawTradeDate);

      let settleDate: string | undefined;
      if (rawSettleDate) {
        try {
          settleDate = parseAustralianDate(rawSettleDate);
        } catch {
          // Non-fatal if settle date fails
        }
      }

      if (!rawType) throw new Error('Missing transaction type');
      let type: 'BUY' | 'SELL' | 'DIVIDEND';
      const upperType = rawType.toUpperCase();
      if (upperType.includes('BUY')) type = 'BUY';
      else if (upperType.includes('SELL')) type = 'SELL';
      else if (upperType.includes('DIV') || upperType === 'DRP' || upperType === 'APPLICATION') type = 'DIVIDEND';
      else throw new Error(`Unsupported transaction type: "${rawType}"`);

      if (!rawCode) throw new Error('Missing stock code');
      const symbol = rawCode.toUpperCase().replace(/\.AX$/i, '').trim();

      const quantity = cleanCurrency(rawQty);
      if (type !== 'DIVIDEND' && (parseFloat(quantity) <= 0 || isNaN(parseFloat(quantity)))) {
        throw new Error(`Quantity must be greater than zero, got: "${rawQty || ''}"`);
      }

      const price = cleanCurrency(rawPrice);
      if (type !== 'DIVIDEND' && (parseFloat(price) < 0 || isNaN(parseFloat(price)))) {
        throw new Error(`Price cannot be negative, got: "${rawPrice || ''}"`);
      }

      const fee = cleanCurrency(rawBrokerage);
      let amount = cleanCurrency(rawConsideration);

      // Reconstruct amount if missing
      if (parseFloat(amount) === 0 && type !== 'DIVIDEND') {
        const principal = parseFloat(quantity) * parseFloat(price);
        amount = (type === 'BUY' ? principal + parseFloat(fee) : principal - parseFloat(fee)).toFixed(2);
      }

      let externalRef: string;
      if (rawConfirm) {
        externalRef = `nabtrade:${rawConfirm}`;
      } else {
        const hashSeed = `nabtrade:${tradeDate}:${symbol}:${type}:${quantity}:${amount}`;
        externalRef = `nt_${await generateIdempotentHash(hashSeed)}`;
      }

      validRows.push({
        rowNumber: rowNum,
        externalRef,
        symbol,
        type,
        tradeDate,
        settleDate,
        quantity: type === 'DIVIDEND' && parseFloat(quantity) === 0 ? '0' : quantity,
        price: type === 'DIVIDEND' && parseFloat(price) === 0 ? '0.00' : price,
        amount,
        fee,
        currencyCode: 'AUD',
        notes: `Imported from nabtrade CSV (Conf: ${rawConfirm || 'N/A'})`,
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
    broker: 'nabtrade',
    totalRowsRead: lines.length - (headerIdx + 1),
    validRows,
    errors,
  };
}

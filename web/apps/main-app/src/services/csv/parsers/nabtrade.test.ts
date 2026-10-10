import { describe, it, expect } from 'vitest';
import { parseNabtradeCSV } from './nabtrade';

describe('nabtrade CSV Parser', () => {
  it('parses trades with confirmation numbers, settle dates, and .AX suffix removal', async () => {
    const lines = [
      'Account: 12345678',
      'Confirmation Number,Trade Date,Settlement Date,Transaction Type,Code,Quantity,Price,Brokerage,Total Consideration',
      'C10982347,12/03/2024,14/03/2024,BUY,MQG.AX,"20","182.40","14.95","3,662.95"',
      'C10982348,15/03/2024,19/03/2024,SELL,BHP,100,44.50,14.95,4435.05',
    ];

    const res = await parseNabtradeCSV(lines, 1);
    expect(res.broker).toBe('nabtrade');
    expect(res.validRows.length).toBe(2);
    expect(res.errors.length).toBe(0);

    const row1 = res.validRows[0];
    expect(row1.externalRef).toBe('nabtrade:C10982347');
    expect(row1.type).toBe('BUY');
    expect(row1.symbol).toBe('MQG');
    expect(row1.tradeDate).toBe('2024-03-12');
    expect(row1.settleDate).toBe('2024-03-14');
    expect(row1.quantity).toBe('20');
    expect(row1.price).toBe('182.40');
    expect(row1.fee).toBe('14.95');
    expect(row1.amount).toBe('3662.95');

    const row2 = res.validRows[1];
    expect(row2.externalRef).toBe('nabtrade:C10982348');
    expect(row2.type).toBe('SELL');
    expect(row2.symbol).toBe('BHP');
    expect(row2.tradeDate).toBe('2024-03-15');
    expect(row2.settleDate).toBe('2024-03-19');
    expect(row2.amount).toBe('4435.05');
  });

  it('stops reading cleanly on legal disclaimer footer text', async () => {
    const lines = [
      'Confirmation Number,Trade Date,Settlement Date,Transaction Type,Code,Quantity,Price,Brokerage,Total Consideration',
      'C10982347,12/03/2024,14/03/2024,BUY,WES,50,60.00,10.00,3010.00',
      'nabtrade is a service provided by Wealthhub Securities Limited...',
      'Additional legal terms',
    ];

    const res = await parseNabtradeCSV(lines, 0);
    expect(res.validRows.length).toBe(1);
    expect(res.validRows[0].symbol).toBe('WES');
    expect(res.errors.length).toBe(0);
  });

  it('generates fallback hash if confirmation number is missing', async () => {
    const lines = [
      'Confirmation Number,Trade Date,Settlement Date,Transaction Type,Code,Quantity,Price,Brokerage,Total Consideration',
      ',12/03/2024,,BUY,WES,50,60.00,10.00,3010.00',
    ];

    const res = await parseNabtradeCSV(lines, 0);
    expect(res.validRows.length).toBe(1);
    expect(res.validRows[0].externalRef).toMatch(/^nt_[a-f0-9]{32}$/);
  });

  it('parses NAB International cash account CSV statements pairing dual debit/credit entries', async () => {
    const lines = [
      'Date,Description,Debit,Credit,Balance',
      '2026-10-09,"DIVIDEND on AVGO.US (WHT of USD -3.61) - USD to AUD @ 1.4247",29.12,,0',
      '2026-10-08,"DIVIDEND on AVGO.US (WHT of USD -3.61) - USD to AUD @ 1.4247",,29.12,29.12',
      '2026-09-23,"SELL AMD.NAS 6 USD 615.94 189411314 NT2678442-004 0.7147",5150.74,,0',
      '2026-09-23,"SELL AMD.NAS 6 USD 615.94 189411314 NT2678442-004 0.7147",,5150.74,5150.74',
      '2026-05-01,"BUY META.NAS 1 USD 613.6 181109809 NT2678442-004 0.7113",,872.56,0',
      '2026-04-30,"BUY META.NAS 1 USD 613.6 181109809 NT2678442-004 0.7113",872.56,,-872.56',
    ];

    const res = await parseNabtradeCSV(lines, 0);
    expect(res.broker).toBe('nabtrade');
    expect(res.errors.length).toBe(0);
    expect(res.validRows.length).toBe(3);

    // 1. BUY META
    const buyRow = res.validRows.find((r) => r.type === 'BUY');
    expect(buyRow).toBeDefined();
    expect(buyRow?.symbol).toBe('META');
    expect(buyRow?.externalRef).toBe('nabtrade:181109809');
    expect(buyRow?.tradeDate).toBe('2026-04-30');
    expect(buyRow?.settleDate).toBe('2026-05-01');
    expect(buyRow?.quantity).toBe('1');
    expect(buyRow?.price).toBe('613.6');
    expect(buyRow?.amount).toBe('613.60');
    expect(buyRow?.currencyCode).toBe('USD');

    // 2. SELL AMD
    const sellRow = res.validRows.find((r) => r.type === 'SELL');
    expect(sellRow).toBeDefined();
    expect(sellRow?.symbol).toBe('AMD');
    expect(sellRow?.externalRef).toBe('nabtrade:189411314');
    expect(sellRow?.tradeDate).toBe('2026-09-23');
    expect(sellRow?.quantity).toBe('6');
    expect(sellRow?.price).toBe('615.94');
    expect(sellRow?.amount).toBe('3695.64');

    // 3. DIVIDEND AVGO
    const divRow = res.validRows.find((r) => r.type === 'DIVIDEND');
    expect(divRow).toBeDefined();
    expect(divRow?.symbol).toBe('AVGO');
    expect(divRow?.tradeDate).toBe('2026-10-08');
    expect(divRow?.externalRef).toMatch(/^nt_[a-f0-9]{32}$/);
    expect(divRow?.amount).toBe('20.44');
    expect(divRow?.quantity).toBe('0');
    expect(divRow?.price).toBe('0.00');
  });

  it('correctly handles NYSE assets, fractional prices, and zero withholding tax dividends', async () => {
    const lines = [
      'Date,Description,Debit,Credit,Balance',
      '2024-09-18,"DIVIDEND on AZN.US (WHT of USD 0.00) - USD to AUD @ 1.4769",21.71,,0',
      '2024-09-17,"DIVIDEND on AZN.US (WHT of USD 0.00) - USD to AUD @ 1.4769",,21.71,21.71',
      '2025-03-05,"BUY TSM.NYS 3 USD 179.273333 156939177 NT2678442-004 0.6172",,881.21,0',
      '2025-03-05,"BUY TSM.NYS 3 USD 179.273333 156939177 NT2678442-004 0.6172",881.21,,-881.21',
      '2026-02-09,"SELL UBER.NYS 64 USD 73.99 175986187 NT2678442-004 0.7059",6688.15,,0',
      '2026-02-07,"SELL UBER.NYS 64 USD 73.99 175986187 NT2678442-004 0.7059",,6688.15,6688.15',
    ];

    const res = await parseNabtradeCSV(lines, 0);
    expect(res.errors.length).toBe(0);
    expect(res.validRows.length).toBe(3);

    const tsm = res.validRows.find((r) => r.symbol === 'TSM');
    expect(tsm?.type).toBe('BUY');
    expect(tsm?.externalRef).toBe('nabtrade:156939177');
    expect(tsm?.quantity).toBe('3');
    expect(tsm?.price).toBe('179.273333');

    const uber = res.validRows.find((r) => r.symbol === 'UBER');
    expect(uber?.type).toBe('SELL');
    expect(uber?.externalRef).toBe('nabtrade:175986187');
    expect(uber?.tradeDate).toBe('2026-02-07');
    expect(uber?.settleDate).toBe('2026-02-09');

    const azn = res.validRows.find((r) => r.symbol === 'AZN');
    expect(azn?.type).toBe('DIVIDEND');
    expect(azn?.tradeDate).toBe('2024-09-17');
    expect(azn?.amount).toBe('14.70'); // 21.71 / 1.4769
  });

  it('parses nabtrade domestic ASX cash account transactions without FX rates', async () => {
    const lines = [
      'Date,Description,Debit,Credit,Balance',
      '2026-09-03,"BUY IVV.ASX 12 AUD 71.43 188373640 NT2678442-002",,867.11,0',
      '2026-09-01,"BUY IVV.ASX 12 AUD 71.43 188373640 NT2678442-002",867.11,,-867.11',
      '2025-08-13,"SELL WTC.ASX 30 AUD 115.56 164863911 NT2678442-002",3451.91,,0',
      '2025-08-11,"SELL WTC.ASX 30 AUD 115.56 164863911 NT2678442-002",,3451.91,3451.91',
    ];

    const res = await parseNabtradeCSV(lines, 0);
    expect(res.errors.length).toBe(0);
    expect(res.validRows.length).toBe(2);

    const buy = res.validRows.find((r) => r.type === 'BUY');
    expect(buy).toBeDefined();
    expect(buy?.symbol).toBe('IVV');
    expect(buy?.currencyCode).toBe('AUD');
    expect(buy?.quantity).toBe('12');
    expect(buy?.price).toBe('71.43');
    expect(buy?.amount).toBe('857.16'); // 12 * 71.43
    expect(buy?.fee).toBe('9.95'); // 867.11 - 857.16
    expect(buy?.tradeDate).toBe('2026-09-01');
    expect(buy?.settleDate).toBe('2026-09-03');
    expect(buy?.externalRef).toBe('nabtrade:188373640');
    expect(buy?.notes).toBe('Imported from nabtrade Cash Account (Order: 188373640)');

    const sell = res.validRows.find((r) => r.type === 'SELL');
    expect(sell).toBeDefined();
    expect(sell?.symbol).toBe('WTC');
    expect(sell?.currencyCode).toBe('AUD');
    expect(sell?.quantity).toBe('30');
    expect(sell?.price).toBe('115.56');
    expect(sell?.amount).toBe('3466.80'); // 30 * 115.56
    expect(sell?.fee).toBe('14.89'); // 3466.80 - 3451.91
    expect(sell?.tradeDate).toBe('2025-08-11');
    expect(sell?.settleDate).toBe('2025-08-13');
    expect(sell?.externalRef).toBe('nabtrade:164863911');
  });
});

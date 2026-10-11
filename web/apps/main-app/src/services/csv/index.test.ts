import { describe, it, expect } from 'vitest';
import { parseBrokerCSV } from './index';

describe('parseBrokerCSV end-to-end', () => {
  it('automatically detects and parses a multi-trade nabtrade International CSV statement', async () => {
    const csvContent = `Date,Description,Debit,Credit,Balance
2026-10-09,"DIVIDEND on AVGO.US (WHT of USD -3.61) - USD to AUD @ 1.4247",29.12,,0
2026-10-09,"DIVIDEND on GOOGL.US (WHT of USD -4.09) - USD to AUD @ 1.4248",33.04,,29.12
2026-10-08,"DIVIDEND on GOOGL.US (WHT of USD -4.09) - USD to AUD @ 1.4248",,33.04,62.16
2026-10-08,"DIVIDEND on AVGO.US (WHT of USD -3.61) - USD to AUD @ 1.4247",,29.12,29.12
2026-09-23,"SELL AMD.NAS 6 USD 615.94 189411314 NT2678442-004 0.7147",5150.74,,0
2026-09-23,"SELL AMD.NAS 6 USD 615.94 189411314 NT2678442-004 0.7147",,5150.74,5150.74
2026-05-01,"BUY META.NAS 1 USD 613.6 181109809 NT2678442-004 0.7113",,872.56,0
2026-04-30,"BUY META.NAS 1 USD 613.6 181109809 NT2678442-004 0.7113",872.56,,-872.56`;

    const res = await parseBrokerCSV(csvContent, 'auto');
    expect(res.broker).toBe('nabtrade');
    expect(res.errors.length).toBe(0);
    expect(res.validRows.length).toBe(4); // 2 dividends, 1 sell, 1 buy
    expect(res.totalRowsRead).toBe(8);

    const buy = res.validRows.find((r) => r.type === 'BUY');
    expect(buy?.symbol).toBe('META');
    expect(buy?.tradeDate).toBe('2026-04-30');
    expect(buy?.settleDate).toBe('2026-05-01');

    const sell = res.validRows.find((r) => r.type === 'SELL');
    expect(sell?.symbol).toBe('AMD');
    expect(sell?.tradeDate).toBe('2026-09-23');

    const avgo = res.validRows.find((r) => r.symbol === 'AVGO');
    expect(avgo?.type).toBe('DIVIDEND');
    expect(avgo?.tradeDate).toBe('2026-10-08');

    const googl = res.validRows.find((r) => r.symbol === 'GOOGL');
    expect(googl?.type).toBe('DIVIDEND');
    expect(googl?.tradeDate).toBe('2026-10-08');
  });

  it('automatically detects and parses a CommSec Cash Account CSV statement', async () => {
    const csvContent = `Date,Reference,Details,Debit($),Credit($),Balance($)
24/06/2026,P36741365,Direct Transfer - Payee MR TOM CRUZ,7960.00,,0.00
22/06/2026,C176305381,S 125 NDQ @ 63.680000  ,,7960.00,-7960.00
09/06/2026,R71985282,Direct Transfer 054190 20173861 Drawer MR TOM CRUZ,,251.22,0.00
04/06/2026,C175717445,B 4 NDQ @ 62.806028  ,251.22,,251.22`;

    const res = await parseBrokerCSV(csvContent, 'auto');
    expect(res.broker).toBe('commsec');
    expect(res.errors.length).toBe(0);
    expect(res.validRows.length).toBe(4);
    expect(res.totalRowsRead).toBe(4);

    const types = res.validRows.map((r) => r.type);
    expect(types).toEqual(['BUY', 'DEPOSIT', 'SELL', 'WITHDRAWAL']);
  });
});

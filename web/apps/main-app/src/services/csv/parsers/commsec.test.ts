import { describe, it, expect } from 'vitest';
import { parseCommSecCSV } from './commsec';

describe('CommSec CSV Parser', () => {
  it('parses standard buy and sell trade rows with commas and dollar signs', async () => {
    const lines = [
      'Date,Type,Code,Units,Unit Price ($),Brokerage ($),Total Value ($)',
      '15/03/2024,Buy,BHP,"1,000","$45.20","$19.95","$45,219.95"',
      '16/03/2024,Sell,CBA,200,$115.50,$19.95,"$23,080.05"',
    ];

    const res = await parseCommSecCSV(lines, 0);
    expect(res.broker).toBe('commsec');
    expect(res.validRows.length).toBe(2);
    expect(res.errors.length).toBe(0);

    const row1 = res.validRows[0];
    expect(row1.type).toBe('BUY');
    expect(row1.symbol).toBe('BHP');
    expect(row1.tradeDate).toBe('2024-03-15');
    expect(row1.quantity).toBe('1000');
    expect(row1.price).toBe('45.20');
    expect(row1.fee).toBe('19.95');
    expect(row1.amount).toBe('45219.95');
    expect(row1.externalRef).toMatch(/^cs_[a-f0-9]{32}$/);

    const row2 = res.validRows[1];
    expect(row2.type).toBe('SELL');
    expect(row2.symbol).toBe('CBA');
    expect(row2.tradeDate).toBe('2024-03-16');
    expect(row2.quantity).toBe('200');
    expect(row2.price).toBe('115.50');
    expect(row2.amount).toBe('23080.05');
  });

  it('parses dividend rows where quantity/unit price are omitted', async () => {
    const lines = [
      'Date,Type,Code,Units,Unit Price ($),Brokerage ($),Total Value ($)',
      '20/03/2024,Div,VAS,,,0.00,350.00',
    ];

    const res = await parseCommSecCSV(lines, 0);
    expect(res.validRows.length).toBe(1);
    expect(res.errors.length).toBe(0);

    const div = res.validRows[0];
    expect(div.type).toBe('DIVIDEND');
    expect(div.symbol).toBe('VAS');
    expect(div.amount).toBe('350.00');
    expect(div.fee).toBe('0.00');
  });

  it('reconstructs total value if missing for Buy and Sell', async () => {
    const lines = [
      'Date,Type,Code,Units,Unit Price ($),Brokerage ($),Total Value ($)',
      '10/01/2024,Buy,TLS,100,4.00,10.00,',
      '11/01/2024,Sell,TLS,50,4.00,5.00,',
    ];

    const res = await parseCommSecCSV(lines, 0);
    expect(res.validRows.length).toBe(2);
    // Buy total = 100 * 4.00 + 10.00 = 410.00
    expect(res.validRows[0].amount).toBe('410.00');
    // Sell total = 50 * 4.00 - 5.00 = 195.00
    expect(res.validRows[1].amount).toBe('195.00');
  });

  it('handles accounting negative parentheses and strips .AX suffix', async () => {
    const lines = [
      'Date,Type,Code,Units,Unit Price ($),Brokerage ($),Total Value ($)',
      '15/03/2024,Sell,MQG.AX,10,180.00,(19.95),(1780.05)',
    ];

    const res = await parseCommSecCSV(lines, 0);
    expect(res.validRows.length).toBe(1);
    expect(res.validRows[0].symbol).toBe('MQG');
    expect(res.validRows[0].fee).toBe('19.95');
    expect(res.validRows[0].amount).toBe('1780.05');
  });

  it('records malformed rows without crashing valid rows', async () => {
    const lines = [
      'Date,Type,Code,Units,Unit Price ($),Brokerage ($),Total Value ($)',
      '32/13/2024,Buy,BHP,10,45.00,0,450.00', // Invalid date
      '15/03/2024,Buy,BHP,0,45.00,0,450.00',   // Zero quantity on buy
      '15/03/2024,Buy,BHP,10,45.00,0,450.00',   // Valid
    ];

    const res = await parseCommSecCSV(lines, 0);
    expect(res.validRows.length).toBe(1);
    expect(res.errors.length).toBe(2);
    expect(res.errors[0].rowNumber).toBe(2);
    expect(res.errors[0].message).toMatch(/date/i);
    expect(res.errors[1].rowNumber).toBe(3);
    expect(res.errors[1].message).toMatch(/quantity/i);
  });
});

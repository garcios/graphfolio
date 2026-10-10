import { describe, it, expect } from 'vitest';
import { detectBrokerFormat } from './detector';

describe('BrokerFormatDetector', () => {
  it('detects CommSec headers with standard columns', () => {
    const lines = [
      'Date,Type,Code,Units,Unit Price ($),Brokerage ($),Total Value ($)',
      '15/03/2024,Buy,BHP,100,45.20,19.95,4539.95',
    ];
    const res = detectBrokerFormat(lines);
    expect(res.detected).toBe('commsec');
    expect(res.headerIndex).toBe(0);
  });

  it('detects CommSec headers with alternate headers and leading blank row', () => {
    const lines = [
      '',
      'Transaction Date,Transaction Type,ASX Code,Quantity,Unit Price,Fee,Net Consideration',
      '15/03/2024,Buy,VAS,50,92.10,10.00,4615.00',
    ];
    const res = detectBrokerFormat(lines);
    expect(res.detected).toBe('commsec');
    expect(res.headerIndex).toBe(1);
  });

  it('detects nabtrade headers with leading metadata rows', () => {
    const lines = [
      'Account Number: 12345678',
      'Account Name: John Doe',
      'Generated: 09/10/2026',
      'Confirmation Number,Trade Date,Settlement Date,Transaction Type,Code,Quantity,Price,Brokerage,Total Consideration',
      'C10982347,12/03/2024,14/03/2024,BUY,MQG,20,182.40,14.95,3662.95',
    ];
    const res = detectBrokerFormat(lines);
    expect(res.detected).toBe('nabtrade');
    expect(res.headerIndex).toBe(3);
  });

  it('detects nabtrade international cash transactions CSV format', () => {
    const lines = [
      'Date,Description,Debit,Credit,Balance',
      '2026-10-09,"DIVIDEND on AVGO.US (WHT of USD -3.61) - USD to AUD @ 1.4247",29.12,,0',
      '2026-09-23,"SELL AMD.NAS 6 USD 615.94 189411314 NT2678442-004 0.7147",5150.74,,0',
    ];
    const res = detectBrokerFormat(lines);
    expect(res.detected).toBe('nabtrade');
    expect(res.headerIndex).toBe(0);
  });

  it('returns null when headers are unrecognized or generic', () => {
    const lines = [
      'Date,Description,Debit,Credit,Balance',
      '01/01/2024,Coffee,4.50,,1000.00',
    ];
    const res = detectBrokerFormat(lines);
    expect(res.detected).toBeNull();
    expect(res.headerIndex).toBe(-1);
  });
});

import { describe, it, expect } from 'vitest';
import { escapeCSVField, formatTransactionsCSV } from './export';
import type { TransactionItem } from '../../components/TransactionLedger';

describe('CSV Exporter (export.ts)', () => {
  describe('escapeCSVField', () => {
    it('returns empty string for null or undefined', () => {
      expect(escapeCSVField(null)).toBe('');
      expect(escapeCSVField(undefined)).toBe('');
    });

    it('returns plain alphanumeric strings without quotes', () => {
      expect(escapeCSVField('AAPL')).toBe('AAPL');
      expect(escapeCSVField('100.50')).toBe('100.50');
    });

    it('wraps fields with commas in double quotes', () => {
      expect(escapeCSVField('BHP Group, Limited')).toBe('"BHP Group, Limited"');
    });

    it('doubles internal quotation marks following RFC 4180', () => {
      expect(escapeCSVField('Saying "Hello"')).toBe('"Saying ""Hello"""');
    });

    it('wraps fields with newlines in double quotes', () => {
      expect(escapeCSVField("Line 1\nLine 2")).toBe('"Line 1\nLine 2"');
    });

    it('neutralizes spreadsheet formula injection (CWE-1236) while preserving pure numbers', () => {
      expect(escapeCSVField('=SUM(A1:B10)')).toBe("'=SUM(A1:B10)");
      expect(escapeCSVField('@malicious')).toBe("'@malicious");
      expect(escapeCSVField('\t=cmd')).toBe("'\t=cmd");
      expect(escapeCSVField('-cmd|calc')).toBe("'-cmd|calc");
      expect(escapeCSVField('+cmd|calc')).toBe("'+cmd|calc");
      // Pure numbers should be preserved
      expect(escapeCSVField('-14500.00')).toBe('-14500.00');
      expect(escapeCSVField('+100')).toBe('+100');
    });
  });

  describe('formatTransactionsCSV', () => {
    const mockTxs: TransactionItem[] = [
      {
        id: 'tx-newest',
        type: 'SELL',
        symbol: 'AAPL',
        instrumentName: 'Apple Inc.',
        tradeDate: '2025-06-15',
        quantity: '50',
        price: { amount: '220.00', currencyCode: 'USD' },
        amount: { amount: '11000.00', currencyCode: 'USD' },
        fee: { amount: '19.95', currencyCode: 'USD' },
        notes: 'Taking profit',
        createdAt: '2025-06-15T12:00:00Z',
      },
      {
        id: 'tx-oldest',
        type: 'DEPOSIT',
        symbol: null,
        instrumentName: 'Cash',
        tradeDate: '2023-01-10',
        quantity: null,
        price: null,
        amount: { amount: '15000.00', currencyCode: 'AUD' },
        fee: { amount: '0.00', currencyCode: 'AUD' },
        notes: 'Initial account funding',
        createdAt: '2023-01-10T09:00:00Z',
      },
      {
        id: 'tx-mid-buy',
        type: 'BUY',
        symbol: 'AAPL',
        instrumentName: 'Apple Inc.',
        tradeDate: '2023-01-15',
        quantity: '100',
        price: { amount: '145.00', currencyCode: 'USD' },
        amount: { amount: '14500.00', currencyCode: 'USD' },
        fee: { amount: '9.95', currencyCode: 'USD' },
        notes: 'Initial buy',
        createdAt: '2023-01-15T10:00:00Z',
        withholdingTax: { amount: '0.00', currencyCode: 'USD' },
        fxRateToBase: '1.485000',
        externalRef: 'BROKER-REF-001',
      },
    ];

    it('prepends UTF-8 BOM (\\uFEFF) for Excel compatibility', () => {
      const csv = formatTransactionsCSV(mockTxs);
      expect(csv.startsWith('\uFEFF')).toBe(true);
    });

    it('strictly sorts transactions in ascending order (oldest first)', () => {
      const csv = formatTransactionsCSV(mockTxs);
      const lines = csv.replace('\uFEFF', '').split('\r\n');

      // Header row
      expect(lines[0]).toContain('Date,Type,Symbol,Instrument Name');
      expect(lines[0]).toContain('Franking Credit,Franking %');

      // Row 1 should be the oldest: 2023-01-10 (DEPOSIT)
      expect(lines[1]).toContain('2023-01-10,DEPOSIT,,Cash');

      // Row 2 should be the middle: 2023-01-15 (BUY)
      expect(lines[2]).toContain('2023-01-15,BUY,AAPL,Apple Inc.');

      // Row 3 should be the newest: 2025-06-15 (SELL)
      expect(lines[3]).toContain('2025-06-15,SELL,AAPL,Apple Inc.');
    });

    it('prioritizes DEPOSIT before BUY on the same trade date', () => {
      const sameDayTxs: TransactionItem[] = [
        {
          id: 'tx-buy',
          type: 'BUY',
          symbol: 'BHP.AX',
          instrumentName: 'BHP Group',
          tradeDate: '2024-04-01',
          quantity: '10',
          price: { amount: '44.00', currencyCode: 'AUD' },
          amount: { amount: '440.00', currencyCode: 'AUD' },
          fee: { amount: '5.00', currencyCode: 'AUD' },
          createdAt: '2024-04-01T10:00:00Z',
        },
        {
          id: 'tx-dep',
          type: 'DEPOSIT',
          tradeDate: '2024-04-01',
          amount: { amount: '1000.00', currencyCode: 'AUD' },
          fee: { amount: '0.00', currencyCode: 'AUD' },
          createdAt: '2024-04-01T11:00:00Z',
        },
      ];

      const csv = formatTransactionsCSV(sameDayTxs);
      const lines = csv.replace('\uFEFF', '').split('\r\n');

      expect(lines[1]).toContain('2024-04-01,DEPOSIT');
      expect(lines[2]).toContain('2024-04-01,BUY');
    });

    it('formats BUY cash flow as negative net amount and SELL as positive', () => {
      const csv = formatTransactionsCSV(mockTxs);
      const lines = csv.replace('\uFEFF', '').split('\r\n');

      // BUY net amount should be negative: -14500.00
      expect(lines[2]).toContain('14500.00,-14500.00,9.95');

      // SELL net amount should be positive: 11000.00
      expect(lines[3]).toContain('11000.00,11000.00,19.95');
    });

    it('formats SPLIT transactions with ratio in quantity and zero amounts', () => {
      const splitTx: TransactionItem = {
        id: 'tx-split',
        type: 'SPLIT',
        symbol: 'NVDA',
        instrumentName: 'NVIDIA Corporation',
        tradeDate: '2024-06-10',
        quantity: '10',
        amount: { amount: '0.00', currencyCode: 'USD' },
        fee: { amount: '0.00', currencyCode: 'USD' },
        notes: '10-for-1 forward split',
        createdAt: '2024-06-10T08:00:00Z',
      };

      const csv = formatTransactionsCSV([splitTx]);
      const lines = csv.replace('\uFEFF', '').split('\r\n');

      expect(lines[1]).toContain('2024-06-10,SPLIT,NVDA,NVIDIA Corporation,10,,USD,0.00,0.00');
    });
  });
});

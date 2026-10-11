import { describe, it, expect } from 'vitest';
import {
  validateAddAssetPriceInput,
  detectPriceConflict,
  type ExistingPriceCheck,
} from './AddAssetPriceModal';

describe('AddAssetPriceModal logic & validation', () => {
  const TODAY = '2026-10-11';

  describe('validateAddAssetPriceInput', () => {
    it('accepts valid input with positive decimal price and current or past date', () => {
      const res = validateAddAssetPriceInput('aapl', '2026-10-10', '230.50', TODAY);
      expect(res.isValid).toBe(true);
      expect(res.cleanSymbol).toBe('AAPL');
      expect(res.cleanDate).toBe('2026-10-10');
      expect(res.cleanPrice).toBe('230.50');
      expect(res.error).toBeUndefined();
    });

    it('accepts prices on the current date', () => {
      const res = validateAddAssetPriceInput('BHP.AX', '2026-10-11', '42.15', TODAY);
      expect(res.isValid).toBe(true);
      expect(res.cleanSymbol).toBe('BHP.AX');
      expect(res.cleanDate).toBe('2026-10-11');
      expect(res.cleanPrice).toBe('42.15');
    });

    it('rejects future price dates', () => {
      const res = validateAddAssetPriceInput('AAPL', '2026-10-12', '230.50', TODAY);
      expect(res.isValid).toBe(false);
      expect(res.error).toBe('Price date cannot be in the future.');
    });

    it('rejects missing or empty symbol', () => {
      const res = validateAddAssetPriceInput('   ', '2026-10-10', '230.50', TODAY);
      expect(res.isValid).toBe(false);
      expect(res.error).toBe('Symbol, date, and closing price are required.');
    });

    it('rejects missing or empty price date', () => {
      const res = validateAddAssetPriceInput('AAPL', '', '230.50', TODAY);
      expect(res.isValid).toBe(false);
      expect(res.error).toBe('Symbol, date, and closing price are required.');
    });

    it('rejects missing or empty price', () => {
      const res = validateAddAssetPriceInput('AAPL', '2026-10-10', '', TODAY);
      expect(res.isValid).toBe(false);
      expect(res.error).toBe('Symbol, date, and closing price are required.');
    });

    it('rejects zero closing price', () => {
      const res = validateAddAssetPriceInput('AAPL', '2026-10-10', '0.00', TODAY);
      expect(res.isValid).toBe(false);
      expect(res.error).toBe('Closing price must be a valid positive number.');
    });

    it('rejects negative closing price', () => {
      const res = validateAddAssetPriceInput('AAPL', '2026-10-10', '-15.50', TODAY);
      expect(res.isValid).toBe(false);
      expect(res.error).toBe('Closing price must be a valid positive number.');
    });

    it('rejects non-numeric closing price strings', () => {
      const res = validateAddAssetPriceInput('AAPL', '2026-10-10', 'invalid_price', TODAY);
      expect(res.isValid).toBe(false);
      expect(res.error).toBe('Closing price must be a valid positive number.');
    });
  });

  describe('detectPriceConflict', () => {
    const existingRecords: ExistingPriceCheck[] = [
      {
        symbol: 'AAPL',
        priceDate: '2026-10-10',
        priceAmount: '228.40',
        currencyCode: 'USD',
      },
      {
        symbol: 'MSFT',
        priceDate: '2026-10-10',
        priceAmount: '415.20',
        currencyCode: 'USD',
      },
      {
        symbol: 'BHP.AX',
        priceDate: '2026-10-09',
        priceAmount: '41.80',
        currencyCode: 'AUD',
      },
    ];

    it('returns null when no existing prices are provided', () => {
      expect(detectPriceConflict('AAPL', '2026-10-10', [])).toBeNull();
    });

    it('returns null when date has no recorded price for the symbol', () => {
      const conflict = detectPriceConflict('AAPL', '2026-10-08', existingRecords);
      expect(conflict).toBeNull();
    });

    it('returns existing record when symbol and date match', () => {
      const conflict = detectPriceConflict('AAPL', '2026-10-10', existingRecords);
      expect(conflict).not.toBeNull();
      expect(conflict?.priceAmount).toBe('228.40');
      expect(conflict?.currencyCode).toBe('USD');
    });

    it('handles case-insensitive symbol comparison', () => {
      const conflict = detectPriceConflict('aapl', '2026-10-10', existingRecords);
      expect(conflict).not.toBeNull();
      expect(conflict?.priceAmount).toBe('228.40');
    });

    it('does not conflict across different symbols on the same date', () => {
      const conflict = detectPriceConflict('GOOG', '2026-10-10', existingRecords);
      expect(conflict).toBeNull();
    });

    it('matches records where symbol is omitted (scoped to single asset)', () => {
      const singleAssetRecords: ExistingPriceCheck[] = [
        {
          priceDate: '2026-10-10',
          priceAmount: '150.00',
          currencyCode: 'USD',
        },
      ];
      const conflict = detectPriceConflict('ANY_SYM', '2026-10-10', singleAssetRecords);
      expect(conflict).not.toBeNull();
      expect(conflict?.priceAmount).toBe('150.00');
    });
  });
});

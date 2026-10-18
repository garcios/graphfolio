import { describe, it, expect } from 'vitest';
import {
  generateDividendWithdrawalNotes,
  buildDividendWithdrawalInput,
} from './WithdrawDividendModal';

describe('Withdraw Against Dividend (US-TX-105)', () => {
  describe('generateDividendWithdrawalNotes', () => {
    it('generates standardized reference notes for a stock dividend', () => {
      const notes = generateDividendWithdrawalNotes('BHP.AX', '2024-04-15');
      expect(notes).toBe('Withdrawal for dividend: BHP.AX on 2024-04-15');
    });

    it('trims leading/trailing whitespace from symbol and tradeDate', () => {
      const notes = generateDividendWithdrawalNotes('   AAPL   ', '  2025-05-20  ');
      expect(notes).toBe('Withdrawal for dividend: AAPL on 2025-05-20');
    });

    it('falls back to "Cash" when symbol is null or empty', () => {
      expect(generateDividendWithdrawalNotes(null, '2024-04-15')).toBe(
        'Withdrawal for dividend: Cash on 2024-04-15'
      );
      expect(generateDividendWithdrawalNotes('', '2024-04-15')).toBe(
        'Withdrawal for dividend: Cash on 2024-04-15'
      );
      expect(generateDividendWithdrawalNotes('   ', '2024-04-15')).toBe(
        'Withdrawal for dividend: Cash on 2024-04-15'
      );
    });

    it('handles missing or empty tradeDate gracefully', () => {
      expect(generateDividendWithdrawalNotes('BHP.AX', null)).toBe(
        'Withdrawal for dividend: BHP.AX on unspecified date'
      );
      expect(generateDividendWithdrawalNotes('BHP.AX', '')).toBe(
        'Withdrawal for dividend: BHP.AX on unspecified date'
      );
    });
  });

  describe('buildDividendWithdrawalInput', () => {
    it('mirrors tradeDate, amount, and currencyCode exactly from source dividend', () => {
      const dividendTx = {
        tradeDate: '2024-04-15',
        amount: {
          amount: '450.00',
          currencyCode: 'AUD',
        },
        symbol: 'BHP.AX',
      };

      const input = buildDividendWithdrawalInput(dividendTx);

      expect(input).toEqual({
        type: 'WITHDRAWAL',
        tradeDate: '2024-04-15',
        amount: '450.00',
        currencyCode: 'AUD',
        notes: 'Withdrawal for dividend: BHP.AX on 2024-04-15',
        fee: '0.00',
        feeCurrencyCode: 'AUD',
      });
    });

    it('preserves multi-currency codes (e.g. USD, EUR, GBP)', () => {
      const dividendTx = {
        tradeDate: '2024-06-10',
        amount: {
          amount: '125.50',
          currencyCode: 'USD',
        },
        symbol: 'AAPL',
      };

      const input = buildDividendWithdrawalInput(dividendTx);
      expect(input.currencyCode).toBe('USD');
      expect(input.feeCurrencyCode).toBe('USD');
      expect(input.amount).toBe('125.50');
      expect(input.tradeDate).toBe('2024-06-10');
    });

    it('respects user-provided custom notes when specified', () => {
      const dividendTx = {
        tradeDate: '2024-04-15',
        amount: {
          amount: '450.00',
          currencyCode: 'AUD',
        },
        symbol: 'BHP.AX',
      };

      const customNotes = 'Swept to Commonwealth Bank living expenses account';
      const input = buildDividendWithdrawalInput(dividendTx, customNotes);

      expect(input.notes).toBe(customNotes);
    });

    it('omits notes if user clears notes to empty whitespace', () => {
      const dividendTx = {
        tradeDate: '2024-04-15',
        amount: {
          amount: '450.00',
          currencyCode: 'AUD',
        },
        symbol: 'BHP.AX',
      };

      const input = buildDividendWithdrawalInput(dividendTx, '   ');
      expect(input.notes).toBeUndefined();
    });

    it('preserves high decimal cents strings without floating-point distortion', () => {
      const dividendTx = {
        tradeDate: '2024-09-30',
        amount: {
          amount: '9876543.21',
          currencyCode: 'AUD',
        },
        symbol: 'VAS.AX',
      };

      const input = buildDividendWithdrawalInput(dividendTx);
      expect(input.amount).toBe('9876543.21');
    });
  });
});

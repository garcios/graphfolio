import { describe, it, expect } from 'vitest';
import {
  validateUpdateTransaction,
  buildUpdateTransactionPayload,
} from './EditTransactionModal';

describe('EditTransactionModal', () => {
  describe('validateUpdateTransaction', () => {
    it('validates required fields for BUY and SELL', () => {
      expect(() =>
        validateUpdateTransaction({
          type: 'BUY',
          symbol: '',
          quantity: '10',
          price: '150.00',
        })
      ).toThrowError(/Please select or specify an asset\/symbol/i);

      expect(() =>
        validateUpdateTransaction({
          type: 'BUY',
          symbol: 'AAPL',
          quantity: '0',
          price: '150.00',
        })
      ).toThrowError(/Quantity must be greater than zero/i);

      expect(() =>
        validateUpdateTransaction({
          type: 'SELL',
          symbol: 'AAPL',
          quantity: '10',
          price: '-5.00',
        })
      ).toThrowError(/Price cannot be negative/i);

      expect(() =>
        validateUpdateTransaction({
          type: 'BUY',
          symbol: 'AAPL',
          quantity: '10',
          price: '150.00',
        })
      ).not.toThrow();
    });

    it('validates required fields for DIVIDEND', () => {
      expect(() =>
        validateUpdateTransaction({
          type: 'DIVIDEND',
          symbol: '',
          amount: '100.00',
        })
      ).toThrowError(/Please select or specify a symbol for dividend/i);

      expect(() =>
        validateUpdateTransaction({
          type: 'DIVIDEND',
          symbol: 'BHP.AX',
          amount: '0',
        })
      ).toThrowError(/Amount must be greater than zero/i);

      expect(() =>
        validateUpdateTransaction({
          type: 'DIVIDEND',
          symbol: 'BHP.AX',
          amount: '500.00',
        })
      ).not.toThrow();
    });

    it('validates required fields for DEPOSIT and WITHDRAWAL', () => {
      expect(() =>
        validateUpdateTransaction({
          type: 'DEPOSIT',
          amount: '0',
        })
      ).toThrowError(/Amount must be greater than zero/i);

      expect(() =>
        validateUpdateTransaction({
          type: 'WITHDRAWAL',
          amount: '250.00',
        })
      ).not.toThrow();
    });

    it('validates required fields for SPLIT', () => {
      expect(() =>
        validateUpdateTransaction({
          type: 'SPLIT',
          symbol: '',
          splitRatioTo: '2',
          splitRatioFrom: '1',
        })
      ).toThrowError(/Please select or specify an asset for the stock split/i);

      expect(() =>
        validateUpdateTransaction({
          type: 'SPLIT',
          symbol: 'NVDA',
          splitRatioTo: '0',
          splitRatioFrom: '1',
        })
      ).toThrowError(/New shares ratio must be greater than zero/i);

      expect(() =>
        validateUpdateTransaction({
          type: 'SPLIT',
          symbol: 'NVDA',
          splitRatioTo: '10',
          splitRatioFrom: '1',
        })
      ).not.toThrow();
    });
  });

  describe('buildUpdateTransactionPayload', () => {
    it('builds BUY payload with correct quantity, price, and amount', () => {
      const payload = buildUpdateTransactionPayload({
        type: 'BUY',
        symbol: 'AAPL',
        tradeDate: '2025-01-15',
        currencyCode: 'USD',
        quantity: '10',
        price: '150.00',
        amount: '1500.00',
        fee: '9.95',
        feeCurrencyCode: 'USD',
        notes: 'Updated purchase note',
      });

      expect(payload).toEqual({
        type: 'BUY',
        symbol: 'AAPL',
        tradeDate: '2025-01-15',
        currencyCode: 'USD',
        quantity: '10',
        price: '150.00',
        amount: '1500.00',
        fee: '9.95',
        feeCurrencyCode: 'USD',
        notes: 'Updated purchase note',
      });
    });

    it('builds DIVIDEND payload with franking credit and percentage', () => {
      const payload = buildUpdateTransactionPayload({
        type: 'DIVIDEND',
        symbol: 'BHP.AX',
        tradeDate: '2025-03-20',
        currencyCode: 'AUD',
        amount: '700.00',
        fee: '0.00',
        frankingCredit: '300.00',
        frankingPct: '100',
        notes: 'Interim fully franked dividend',
      });

      expect(payload).toEqual({
        type: 'DIVIDEND',
        symbol: 'BHP.AX',
        tradeDate: '2025-03-20',
        currencyCode: 'AUD',
        amount: '700.00',
        fee: '0.00',
        feeCurrencyCode: undefined,
        frankingCredit: '300.00',
        frankingPct: '100',
        notes: 'Interim fully franked dividend',
      });
    });

    it('builds SPLIT payload with computed ratio multiplier', () => {
      const payload = buildUpdateTransactionPayload({
        type: 'SPLIT',
        symbol: 'NVDA',
        tradeDate: '2024-06-10',
        currencyCode: 'USD',
        splitRatioTo: '10',
        splitRatioFrom: '1',
      });

      expect(payload).toEqual({
        type: 'SPLIT',
        symbol: 'NVDA',
        tradeDate: '2024-06-10',
        currencyCode: 'USD',
        quantity: '10',
        price: '0.00',
        amount: '0.00',
        fee: undefined,
        feeCurrencyCode: undefined,
        notes: undefined,
      });
    });

    it('builds DEPOSIT payload with amount and currency', () => {
      const payload = buildUpdateTransactionPayload({
        type: 'DEPOSIT',
        symbol: '',
        tradeDate: '2025-01-01',
        currencyCode: 'AUD',
        amount: '5000.00',
      });

      expect(payload).toEqual({
        type: 'DEPOSIT',
        tradeDate: '2025-01-01',
        currencyCode: 'AUD',
        amount: '5000.00',
        fee: undefined,
        feeCurrencyCode: undefined,
        notes: undefined,
      });
    });
  });
});

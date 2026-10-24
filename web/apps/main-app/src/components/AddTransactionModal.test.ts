import { describe, it, expect } from 'vitest';
import {
  transactionTypes,
  validateTransactionInput,
  buildAddTransactionPayload,
} from './AddTransactionModal';

describe('Add Transaction Modal - Interest Support (US-TX-107)', () => {
  describe('transactionTypes list', () => {
    it('contains INTEREST in the supported types', () => {
      expect(transactionTypes).toContain('INTEREST');
    });

    it('orders INTEREST logically adjacent to DIVIDEND and before SPLIT', () => {
      const divIndex = transactionTypes.indexOf('DIVIDEND');
      const interestIndex = transactionTypes.indexOf('INTEREST');
      const splitIndex = transactionTypes.indexOf('SPLIT');

      expect(divIndex).toBeGreaterThanOrEqual(0);
      expect(interestIndex).toBe(divIndex + 1);
      expect(splitIndex).toBe(interestIndex + 1);
    });
  });

  describe('validateTransactionInput for INTEREST', () => {
    it('passes when positive amount is provided without symbol or quantity', () => {
      expect(() =>
        validateTransactionInput({
          type: 'INTEREST',
          amount: '125.50',
        })
      ).not.toThrow();
    });

    it('throws error when amount is missing or empty', () => {
      expect(() =>
        validateTransactionInput({
          type: 'INTEREST',
          amount: '',
        })
      ).toThrow('Amount must be greater than zero.');
    });

    it('throws error when amount is zero', () => {
      expect(() =>
        validateTransactionInput({
          type: 'INTEREST',
          amount: '0.00',
        })
      ).toThrow('Amount must be greater than zero.');
    });

    it('throws error when amount is negative', () => {
      expect(() =>
        validateTransactionInput({
          type: 'INTEREST',
          amount: '-50.00',
        })
      ).toThrow('Amount must be greater than zero.');
    });
  });

  describe('buildAddTransactionPayload for INTEREST', () => {
    it('builds a clean symbol-free payload for cash interest', () => {
      const payload = buildAddTransactionPayload({
        type: 'INTEREST',
        tradeDate: '2026-10-18',
        currencyCode: 'AUD',
        amount: '250.00',
        notes: 'Monthly CMA Yield',
      });

      expect(payload).toEqual({
        type: 'INTEREST',
        tradeDate: '2026-10-18',
        currencyCode: 'AUD',
        amount: '250.00',
        notes: 'Monthly CMA Yield',
        fee: undefined,
        feeCurrencyCode: undefined,
      });

      // Symbol and quantity should not be present
      expect(payload.symbol).toBeUndefined();
      expect(payload.quantity).toBeUndefined();
      expect(payload.price).toBeUndefined();
    });

    it('handles interest without optional notes', () => {
      const payload = buildAddTransactionPayload({
        type: 'INTEREST',
        tradeDate: '2026-10-18',
        currencyCode: 'USD',
        amount: '75.25',
      });

      expect(payload).toEqual({
        type: 'INTEREST',
        tradeDate: '2026-10-18',
        currencyCode: 'USD',
        amount: '75.25',
        notes: undefined,
        fee: undefined,
        feeCurrencyCode: undefined,
      });
    });
  });

  describe('zero regression for existing transaction types', () => {
    it('BUY requires symbol, quantity, and price', () => {
      expect(() =>
        validateTransactionInput({
          type: 'BUY',
          quantity: '10',
          price: '150.00',
        })
      ).toThrow('Please select or specify a symbol.');

      expect(() =>
        validateTransactionInput({
          type: 'BUY',
          symbol: 'AAPL',
          quantity: '0',
          price: '150.00',
        })
      ).toThrow('Quantity must be greater than zero.');
    });

    it('DIVIDEND requires symbol and amount', () => {
      expect(() =>
        validateTransactionInput({
          type: 'DIVIDEND',
          amount: '100.00',
        })
      ).toThrow('Please select or specify a symbol for dividend.');

      expect(() =>
        validateTransactionInput({
          type: 'DIVIDEND',
          symbol: 'BHP.AX',
          amount: '0.00',
        })
      ).toThrow('Amount must be greater than zero.');
    });

    it('DEPOSIT requires positive amount and allows symbol-free entry', () => {
      expect(() =>
        validateTransactionInput({
          type: 'DEPOSIT',
          amount: '1000.00',
        })
      ).not.toThrow();
    });

    it('FX_CONVERSION validates positive source amount, distinct target currency, and target amount/rate', () => {
      // Valid conversion: 1500 AUD to 1000 USD @ 0.6667
      expect(() =>
        validateTransactionInput({
          type: 'FX_CONVERSION',
          currencyCode: 'AUD',
          targetCurrency: 'USD',
          amount: '1500.00',
          quantity: '1000.00',
          price: '0.6667',
        })
      ).not.toThrow();

      // Zero or missing source amount
      expect(() =>
        validateTransactionInput({
          type: 'FX_CONVERSION',
          currencyCode: 'AUD',
          targetCurrency: 'USD',
          amount: '0.00',
          quantity: '1000.00',
        })
      ).toThrow('Source amount must be greater than zero.');

      // Identical currencies
      expect(() =>
        validateTransactionInput({
          type: 'FX_CONVERSION',
          currencyCode: 'USD',
          targetCurrency: 'USD',
          amount: '1000.00',
          quantity: '1000.00',
        })
      ).toThrow('Source and target currency cannot be identical.');

      // Missing target amount and price
      expect(() =>
        validateTransactionInput({
          type: 'FX_CONVERSION',
          currencyCode: 'AUD',
          targetCurrency: 'USD',
          amount: '1500.00',
          quantity: '0.00',
          price: '0.00',
        })
      ).toThrow('Target amount received or conversion rate must be greater than zero.');
    });
  });

  describe('buildAddTransactionPayload for FX_CONVERSION', () => {
    it('builds a complete FX conversion payload with target currency as symbol', () => {
      const payload = buildAddTransactionPayload({
        type: 'FX_CONVERSION',
        tradeDate: '2026-10-23',
        currencyCode: 'AUD',
        targetCurrency: 'USD',
        amount: '1500.00',
        quantity: '1000.00',
        price: '0.6667',
        fee: '5.00',
        feeCurrencyCode: 'AUD',
      });

      expect(payload).toEqual({
        type: 'FX_CONVERSION',
        tradeDate: '2026-10-23',
        currencyCode: 'AUD',
        symbol: 'USD',
        amount: '1500.00',
        quantity: '1000.00',
        price: '0.6667',
        fee: '5.00',
        feeCurrencyCode: 'AUD',
        notes: 'Converted 1500.00 AUD to 1000.00 USD @ 0.6667',
      });
    });
  });

  describe('buildAddTransactionPayload for DIVIDEND with Franking Credits', () => {
    it('builds a dividend payload with explicit franking credit and percentage', () => {
      const payload = buildAddTransactionPayload({
        type: 'DIVIDEND',
        tradeDate: '2026-10-24',
        currencyCode: 'AUD',
        symbol: 'BHP.AX',
        amount: '700.00',
        frankingCredit: '300.00',
        frankingPct: '100',
      });

      expect(payload).toEqual({
        type: 'DIVIDEND',
        tradeDate: '2026-10-24',
        currencyCode: 'AUD',
        symbol: 'BHP.AX',
        amount: '700.00',
        frankingCredit: '300.00',
        frankingPct: '100',
      });
    });

    it('omits franking fields when not provided or zero', () => {
      const payload = buildAddTransactionPayload({
        type: 'DIVIDEND',
        tradeDate: '2026-10-24',
        currencyCode: 'USD',
        symbol: 'AAPL',
        amount: '150.00',
      });

      expect(payload).toEqual({
        type: 'DIVIDEND',
        tradeDate: '2026-10-24',
        currencyCode: 'USD',
        symbol: 'AAPL',
        amount: '150.00',
      });
      expect(payload.frankingCredit).toBeUndefined();
      expect(payload.frankingPct).toBeUndefined();
    });
  });
});

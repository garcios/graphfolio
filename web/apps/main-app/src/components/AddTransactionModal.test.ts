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
  });
});

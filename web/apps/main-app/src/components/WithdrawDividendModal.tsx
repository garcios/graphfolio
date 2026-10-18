import React, { useState, useEffect } from 'react';
import { client } from '@graphfolio/api-client';
import { formatMoney } from '@graphfolio/ui';
import type { TransactionItem } from './TransactionLedger';
import './AddTransactionModal.css';
import './WithdrawDividendModal.css';

/**
 * Generates standardized descriptive notes for a withdrawal created against a dividend.
 */
export function generateDividendWithdrawalNotes(
  symbol?: string | null,
  tradeDate?: string | null
): string {
  const assetLabel = symbol?.trim() ? symbol.trim() : 'Cash';
  const dateLabel = tradeDate?.trim() ? tradeDate.trim() : 'unspecified date';
  return `Withdrawal for dividend: ${assetLabel} on ${dateLabel}`;
}

/**
 * Constructs the GraphQL TransactionInput for a matching withdrawal transaction.
 */
export function buildDividendWithdrawalInput(
  dividendTx: {
    tradeDate: string;
    amount: { amount: string; currencyCode: string };
    symbol?: string | null;
  },
  customNotes?: string
): {
  type: 'WITHDRAWAL';
  tradeDate: string;
  amount: string;
  currencyCode: string;
  notes?: string;
  fee: string;
  feeCurrencyCode: string;
} {
  const notes = customNotes !== undefined
    ? customNotes.trim()
    : generateDividendWithdrawalNotes(dividendTx.symbol, dividendTx.tradeDate);

  return {
    type: 'WITHDRAWAL',
    tradeDate: dividendTx.tradeDate,
    amount: dividendTx.amount.amount,
    currencyCode: dividendTx.amount.currencyCode,
    notes: notes || undefined,
    fee: '0.00',
    feeCurrencyCode: dividendTx.amount.currencyCode,
  };
}

export interface WithdrawDividendModalProps {
  isOpen: boolean;
  dividendTx: TransactionItem | null;
  onClose: () => void;
  onSuccess: (updatedPortfolio: any) => void;
}

export const WithdrawDividendModal: React.FC<WithdrawDividendModalProps> = ({
  isOpen,
  dividendTx,
  onClose,
  onSuccess,
}) => {
  const [tradeDate, setTradeDate] = useState('');
  const [amount, setAmount] = useState('');
  const [currencyCode, setCurrencyCode] = useState('USD');
  const [notes, setNotes] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (dividendTx) {
      setTradeDate(dividendTx.tradeDate || '');
      setAmount(dividendTx.amount?.amount || '');
      setCurrencyCode(dividendTx.amount?.currencyCode || 'USD');
      setNotes(generateDividendWithdrawalNotes(dividendTx.symbol, dividendTx.tradeDate));
      setError(null);
    }
  }, [dividendTx]);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isOpen && !loading) {
        onClose();
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, loading, onClose]);

  if (!isOpen || !dividendTx) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    const parsedAmount = parseFloat(amount);
    if (isNaN(parsedAmount) || parsedAmount <= 0) {
      setError('Amount must be a positive number greater than zero.');
      return;
    }

    if (!tradeDate) {
      setError('Withdrawal date is required.');
      return;
    }

    setLoading(true);

    try {
      const input = buildDividendWithdrawalInput(
        {
          tradeDate,
          amount: { amount, currencyCode },
          symbol: dividendTx.symbol,
        },
        notes
      );

      const res = await client.mutation({
        addTransaction: {
          __args: {
            input,
          },
          transactionId: true,
          portfolio: {
            totalValue: { amount: true, currencyCode: true },
            todayReturnAmount: { amount: true, currencyCode: true },
            todayReturnPercent: true,
            annualizedReturnPercent: true,
            cashBalance: { amount: true, currencyCode: true },
            investments: {
              id: true,
              ticker: true,
              name: true,
              price: { amount: true, currencyCode: true },
              quantity: true,
              totalValue: { amount: true, currencyCode: true },
              todayReturnAmount: { amount: true, currencyCode: true },
              todayReturnPercent: true,
              totalReturnAmount: { amount: true, currencyCode: true },
              totalReturnPercent: true,
            },
          },
        },
      });

      if (res.addTransaction?.portfolio) {
        onSuccess(res.addTransaction.portfolio);
        onClose();
      } else {
        setError('Failed to record withdrawal transaction.');
      }
    } catch (err: any) {
      console.error('Error creating withdrawal transaction against dividend:', err);
      setError(err.message || 'An unexpected error occurred while saving the withdrawal.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="modal-backdrop" onClick={onClose} role="dialog" aria-modal="true" aria-labelledby="withdraw-modal-title">
      <div className="modal-container withdraw-modal-container" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <div className="withdraw-modal-header-group">
            <span className="withdraw-badge">CASH SWEEP</span>
            <h3 id="withdraw-modal-title" className="modal-title">Withdraw Dividend Cash</h3>
          </div>
          <button
            type="button"
            className="modal-close-btn"
            onClick={onClose}
            aria-label="Close modal"
            disabled={loading}
          >
            &times;
          </button>
        </div>

        <form onSubmit={handleSubmit} className="modal-form">
          {error && <div className="modal-error">{error}</div>}

          {/* Source Dividend Info Box */}
          <div className="withdraw-source-card">
            <div className="withdraw-source-title">Source Dividend Distribution</div>
            <div className="withdraw-source-grid">
              <div className="withdraw-source-item">
                <span className="label">Asset</span>
                <span className="value asset-highlight">
                  {dividendTx.symbol || 'Cash'}
                  {dividendTx.instrumentName && ` - ${dividendTx.instrumentName}`}
                </span>
              </div>
              <div className="withdraw-source-item">
                <span className="label">Dividend Date</span>
                <span className="value">{dividendTx.tradeDate}</span>
              </div>
              <div className="withdraw-source-item">
                <span className="label">Net Dividend</span>
                <span className="value amount-highlight">{formatMoney(dividendTx.amount)}</span>
              </div>
            </div>
          </div>

          <div className="withdraw-info-callout">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <circle cx="12" cy="12" r="10" />
              <line x1="12" y1="16" x2="12" y2="12" />
              <line x1="12" y1="8" x2="12.01" y2="8" />
            </svg>
            <div>
              This action creates an exact matching cash withdrawal on the distribution date. It offsets your cash balance without reducing your historical dividend income return.
            </div>
          </div>

          {/* Form inputs matching the dividend */}
          <div className="form-row">
            <div className="form-group">
              <label htmlFor="withdraw-tradeDate">Withdrawal Date</label>
              <input
                id="withdraw-tradeDate"
                type="date"
                value={tradeDate}
                onChange={(e) => setTradeDate(e.target.value)}
                required
                disabled={loading}
              />
            </div>

            <div className="form-group">
              <label htmlFor="withdraw-amount">Amount ({currencyCode})</label>
              <div className="input-with-adornment">
                <input
                  id="withdraw-amount"
                  type="number"
                  step="any"
                  min="0.01"
                  value={amount}
                  onChange={(e) => setAmount(e.target.value)}
                  placeholder="0.00"
                  required
                  disabled={loading}
                />
                <span className="currency-adornment">{currencyCode}</span>
              </div>
            </div>
          </div>

          <div className="form-group">
            <label htmlFor="withdraw-notes">Notes / Reference</label>
            <input
              id="withdraw-notes"
              type="text"
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              placeholder="e.g. Swept to personal bank account"
              disabled={loading}
            />
          </div>

          <div className="modal-actions">
            <button
              type="button"
              className="btn-secondary"
              onClick={onClose}
              disabled={loading}
            >
              Cancel
            </button>
            <button
              type="submit"
              className="btn-primary btn-withdraw-confirm"
              disabled={loading || !amount || !tradeDate}
            >
              {loading ? 'Recording...' : 'Confirm Withdrawal'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

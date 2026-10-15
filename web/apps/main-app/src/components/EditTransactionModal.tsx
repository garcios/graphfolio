import React, { useState, useEffect } from 'react';
import { client } from '@graphfolio/api-client';
import type { TransactionItem } from './TransactionLedger';
import './AddTransactionModal.css';

type TransactionType = 'BUY' | 'SELL' | 'DIVIDEND' | 'SPLIT' | 'DEPOSIT' | 'WITHDRAWAL';

interface InstrumentOption {
  id: string;
  symbol: string;
  name: string;
  currencyCode: string;
}

export interface CurrencyOption {
  code: string;
  name?: string;
  symbol?: string;
}

interface EditTransactionModalProps {
  isOpen: boolean;
  transaction: TransactionItem;
  onClose: () => void;
  onSuccess: (updatedPortfolio: any, updatedTx: any) => void;
  preferredCurrency?: string;
  supportedCurrencies?: CurrencyOption[];
}

export const EditTransactionModal: React.FC<EditTransactionModalProps> = ({
  isOpen,
  transaction,
  onClose,
  onSuccess,
  preferredCurrency,
  supportedCurrencies,
}) => {
  const [type, setType] = useState<TransactionType>(() => {
    const validTypes: TransactionType[] = ['BUY', 'SELL', 'DIVIDEND', 'SPLIT', 'DEPOSIT', 'WITHDRAWAL'];
    return validTypes.includes(transaction.type as TransactionType)
      ? (transaction.type as TransactionType)
      : 'BUY';
  });

  const [symbol, setSymbol] = useState(transaction.symbol || '');
  const [tradeDate, setTradeDate] = useState(transaction.tradeDate || '');
  const [quantity, setQuantity] = useState(transaction.quantity || '');
  const [price, setPrice] = useState(transaction.price?.amount || '');
  const [amount, setAmount] = useState(transaction.amount?.amount || '');
  const [fee, setFee] = useState(transaction.fee?.amount || '0.00');
  const [feeCurrencyCode, setFeeCurrencyCode] = useState(transaction.fee?.currencyCode || preferredCurrency || 'USD');
  const [currencyCode, setCurrencyCode] = useState(transaction.amount?.currencyCode || 'USD');
  const [notes, setNotes] = useState(transaction.notes || '');

  const [splitRatioTo, setSplitRatioTo] = useState(transaction.quantity || '2');
  const [splitRatioFrom, setSplitRatioFrom] = useState('1');

  const [instruments, setInstruments] = useState<InstrumentOption[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const availableCurrencies = Array.from(
    new Set([
      preferredCurrency,
      ...(supportedCurrencies || []).map((c) => c.code),
      'USD',
      'AUD',
      'EUR',
      'GBP',
      'CAD',
      'JPY',
      'CHF',
    ].filter(Boolean) as string[])
  );

  // Fetch instruments on mount
  useEffect(() => {
    client
      .query({
        instruments: {
          id: true,
          symbol: true,
          name: true,
          currencyCode: true,
        },
      })
      .then((res) => {
        if (res.instruments) {
          setInstruments(res.instruments);
          if (!symbol && res.instruments.length > 0) {
            setSymbol(res.instruments[0].symbol);
            setCurrencyCode(res.instruments[0].currencyCode);
          }
        }
      })
      .catch((err) => {
        console.error('Failed to fetch instruments:', err);
      });
  }, []);

  // Keyboard shortcut: Escape to close
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isOpen && !loading) {
        onClose();
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, loading, onClose]);

  const handleQuantityChange = (val: string) => {
    setQuantity(val);
    if (type === 'BUY' || type === 'SELL') {
      const q = parseFloat(val);
      const p = parseFloat(price);
      if (!isNaN(q) && !isNaN(p) && q > 0 && p >= 0) {
        setAmount((q * p).toFixed(2));
      }
    }
  };

  const handlePriceChange = (val: string) => {
    setPrice(val);
    if (type === 'BUY' || type === 'SELL') {
      const q = parseFloat(quantity);
      const p = parseFloat(val);
      if (!isNaN(q) && !isNaN(p) && q > 0 && p >= 0) {
        setAmount((q * p).toFixed(2));
      }
    }
  };

  const handleInstrumentChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const sym = e.target.value;
    setSymbol(sym);
    const found = instruments.find((i) => i.symbol === sym);
    if (found) {
      setCurrencyCode(found.currencyCode);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setLoading(true);

    try {
      // Validate inputs
      if (type === 'BUY' || type === 'SELL') {
        if (!symbol) throw new Error('Please select or specify an asset/symbol.');
        if (!quantity || parseFloat(quantity) <= 0) throw new Error('Quantity must be greater than zero.');
        if (!price || parseFloat(price) < 0) throw new Error('Price cannot be negative.');
      } else if (type === 'DEPOSIT' || type === 'WITHDRAWAL') {
        if (!amount || parseFloat(amount) <= 0) throw new Error('Amount must be greater than zero.');
      } else if (type === 'DIVIDEND') {
        if (!symbol) throw new Error('Please select or specify a symbol for dividend.');
        if (!amount || parseFloat(amount) <= 0) throw new Error('Amount must be greater than zero.');
      } else if (type === 'SPLIT') {
        if (!symbol) throw new Error('Please select or specify an asset for the stock split.');
        const to = parseFloat(splitRatioTo);
        const from = parseFloat(splitRatioFrom);
        if (isNaN(to) || to <= 0) throw new Error('New shares ratio must be greater than zero.');
        if (isNaN(from) || from <= 0) throw new Error('Old shares ratio must be greater than zero.');
      }

      const input: any = {
        type,
        tradeDate,
        currencyCode,
        notes: notes || undefined,
        fee: fee ? fee : undefined,
        feeCurrencyCode: (type === 'BUY' || type === 'SELL') ? feeCurrencyCode : (feeCurrencyCode || undefined),
      };

      if (type === 'BUY' || type === 'SELL') {
        input.symbol = symbol;
        input.quantity = quantity;
        input.price = price;
        input.amount = amount || (parseFloat(quantity) * parseFloat(price)).toFixed(2);
      } else if (type === 'DEPOSIT' || type === 'WITHDRAWAL') {
        input.amount = amount;
      } else if (type === 'DIVIDEND') {
        input.symbol = symbol;
        input.amount = amount;
      } else if (type === 'SPLIT') {
        const to = parseFloat(splitRatioTo);
        const from = parseFloat(splitRatioFrom);
        const ratioMultiplier = (to / from).toFixed(8).replace(/\.?0+$/, '');
        input.symbol = symbol;
        input.quantity = ratioMultiplier;
        input.price = '0.00';
        input.amount = '0.00';
      }

      const res = await client.mutation({
        updateTransaction: {
          __args: {
            id: transaction.id,
            input,
          },
          success: true,
          transaction: {
            id: true,
            type: true,
            symbol: true,
            instrumentName: true,
            tradeDate: true,
            quantity: true,
            price: { amount: true, currencyCode: true },
            amount: { amount: true, currencyCode: true },
            fee: { amount: true, currencyCode: true },
            notes: true,
            createdAt: true,
          },
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
              capitalGainAmount: { amount: true, currencyCode: true },
              incomeAmount: { amount: true, currencyCode: true },
              currencyGainAmount: { amount: true, currencyCode: true },
              isInternational: true,
              isClosed: true,
            },
          },
        },
      });

      if (res.updateTransaction && res.updateTransaction.success) {
        onSuccess(res.updateTransaction.portfolio, res.updateTransaction.transaction);
        onClose();
      } else {
        setError('Failed to update transaction. Please try again.');
      }
    } catch (err: any) {
      console.error('Error updating transaction:', err);
      setError(err?.message || 'Failed to update transaction. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div className="modal-backdrop" onClick={() => !loading && onClose()}>
      <div className="modal-container" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h2 className="modal-title">✏️ Edit Transaction</h2>
          <button className="modal-close-btn" onClick={onClose} type="button" disabled={loading}>
            &times;
          </button>
        </div>

        {/* Transaction Type Tabs */}
        <div className="type-tabs">
          {(['BUY', 'SELL', 'DIVIDEND', 'SPLIT', 'DEPOSIT', 'WITHDRAWAL'] as TransactionType[]).map((t) => (
            <button
              key={t}
              type="button"
              className={`type-tab ${type === t ? 'active' : ''} ${type === 'SELL' ? 'sell' : ''} ${type === 'SPLIT' ? 'split' : ''}`}
              onClick={() => {
                setType(t);
                setError(null);
              }}
            >
              {t}
            </button>
          ))}
        </div>

        <form className="modal-form" onSubmit={handleSubmit}>
          {error && <div className="error-banner">{error}</div>}

          {/* Asset Selection for trade and corporate action types */}
          {(type === 'BUY' || type === 'SELL' || type === 'DIVIDEND' || type === 'SPLIT') && (
            <div className="form-group">
              <label>Asset / Ticker</label>
              <select
                className="form-select"
                value={symbol}
                onChange={handleInstrumentChange}
              >
                {instruments.map((inst) => (
                  <option key={inst.id} value={inst.symbol}>
                    {inst.symbol} - {inst.name} ({inst.currencyCode})
                  </option>
                ))}
              </select>
            </div>
          )}

          {/* Trade Date & Execution Currency */}
          <div className="form-row">
            <div className="form-group">
              <label>Trade Date</label>
              <input
                type="date"
                className="form-input"
                value={tradeDate}
                onChange={(e) => setTradeDate(e.target.value)}
                required
              />
            </div>

            <div className="form-group">
              <label>Currency</label>
              <select
                className="form-select"
                value={currencyCode}
                onChange={(e) => setCurrencyCode(e.target.value)}
                disabled={type === 'BUY' || type === 'SELL' || type === 'SPLIT'}
              >
                {availableCurrencies.map((c) => (
                  <option key={c} value={c}>
                    {c}
                  </option>
                ))}
              </select>
            </div>
          </div>

          {/* BUY & SELL: Quantity, Price, Total Amount, Fee */}
          {(type === 'BUY' || type === 'SELL') && (
            <>
              <div className="form-row">
                <div className="form-group">
                  <label>Quantity (Shares)</label>
                  <input
                    type="number"
                    step="any"
                    className="form-input"
                    placeholder="e.g. 10.5"
                    value={quantity}
                    onChange={(e) => handleQuantityChange(e.target.value)}
                    required
                  />
                </div>
                <div className="form-group">
                  <label>Price ({currencyCode})</label>
                  <input
                    type="number"
                    step="any"
                    className="form-input"
                    placeholder="e.g. 150.25"
                    value={price}
                    onChange={(e) => handlePriceChange(e.target.value)}
                    required
                  />
                </div>
              </div>

              <div className="form-row">
                <div className="form-group">
                  <label>Total Value ({currencyCode})</label>
                  <input
                    type="number"
                    step="any"
                    className="form-input"
                    placeholder="Auto-calculated (Qty × Price)"
                    value={amount}
                    onChange={(e) => setAmount(e.target.value)}
                    required
                  />
                </div>
                <div className="form-group">
                  <label>Brokerage Fee</label>
                  <div className="fee-input-group">
                    <input
                      type="number"
                      step="any"
                      className="form-input"
                      placeholder="0.00"
                      value={fee}
                      onChange={(e) => setFee(e.target.value)}
                    />
                    <select
                      className="form-select fee-currency-select"
                      value={feeCurrencyCode}
                      onChange={(e) => setFeeCurrencyCode(e.target.value)}
                    >
                      {availableCurrencies.map((c) => (
                        <option key={c} value={c}>
                          {c}
                        </option>
                      ))}
                    </select>
                  </div>
                </div>
              </div>
            </>
          )}

          {/* SPLIT: Ratio Inputs */}
          {type === 'SPLIT' && (
            <div className="form-row">
              <div className="form-group">
                <label>New Shares (For)</label>
                <input
                  type="number"
                  step="any"
                  className="form-input"
                  placeholder="e.g. 2"
                  value={splitRatioTo}
                  onChange={(e) => setSplitRatioTo(e.target.value)}
                  required
                />
              </div>
              <div className="form-group">
                <label>Old Shares (From)</label>
                <input
                  type="number"
                  step="any"
                  className="form-input"
                  placeholder="e.g. 1"
                  value={splitRatioFrom}
                  onChange={(e) => setSplitRatioFrom(e.target.value)}
                  required
                />
              </div>
            </div>
          )}

          {/* DIVIDEND: Net Amount & Fee */}
          {type === 'DIVIDEND' && (
            <div className="form-row">
              <div className="form-group">
                <label>Dividend Amount ({currencyCode})</label>
                <input
                  type="number"
                  step="any"
                  className="form-input"
                  placeholder="e.g. 250.00"
                  value={amount}
                  onChange={(e) => setAmount(e.target.value)}
                  required
                />
              </div>
              <div className="form-group">
                <label>Withholding Tax / Fee ({currencyCode})</label>
                <input
                  type="number"
                  step="any"
                  className="form-input"
                  placeholder="0.00"
                  value={fee}
                  onChange={(e) => setFee(e.target.value)}
                />
              </div>
            </div>
          )}

          {/* CASH: Deposit & Withdrawal Amount & Fee */}
          {(type === 'DEPOSIT' || type === 'WITHDRAWAL') && (
            <div className="form-row">
              <div className="form-group">
                <label>Amount ({currencyCode})</label>
                <input
                  type="number"
                  step="any"
                  className="form-input"
                  placeholder="e.g. 1000.00"
                  value={amount}
                  onChange={(e) => setAmount(e.target.value)}
                  required
                />
              </div>
              <div className="form-group">
                <label>Transfer Fee ({currencyCode})</label>
                <input
                  type="number"
                  step="any"
                  className="form-input"
                  placeholder="0.00"
                  value={fee}
                  onChange={(e) => setFee(e.target.value)}
                />
              </div>
            </div>
          )}

          {/* Notes */}
          <div className="form-group">
            <label>Notes (Optional)</label>
            <input
              type="text"
              className="form-input"
              placeholder="e.g. Rebalanced position, updated brokerage fee"
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
            />
          </div>

          <div className="modal-actions">
            <button
              type="button"
              className="btn-cancel"
              onClick={onClose}
              disabled={loading}
            >
              Cancel
            </button>
            <button
              type="submit"
              className={`btn-submit ${type === 'SELL' ? 'btn-sell' : ''}`}
              disabled={loading}
            >
              {loading ? 'Saving Changes...' : 'Save Changes'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

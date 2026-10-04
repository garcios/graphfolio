import React, { useState, useEffect } from 'react';
import { createClient } from '../generated';
import './AddTransactionModal.css';

const client = createClient({
  url: 'http://localhost:8080/query',
});

type TransactionType = 'BUY' | 'SELL' | 'DEPOSIT' | 'WITHDRAWAL' | 'DIVIDEND';

interface InstrumentOption {
  id: string;
  symbol: string;
  name: string;
  currencyCode: string;
}

interface AddTransactionModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: (updatedPortfolio: any) => void;
}

export const AddTransactionModal: React.FC<AddTransactionModalProps> = ({
  isOpen,
  onClose,
  onSuccess,
}) => {
  const [type, setType] = useState<TransactionType>('BUY');
  const [symbol, setSymbol] = useState('AAPL');
  const [tradeDate, setTradeDate] = useState(() => new Date().toISOString().split('T')[0]);
  const [quantity, setQuantity] = useState('');
  const [price, setPrice] = useState('');
  const [amount, setAmount] = useState('');
  const [fee, setFee] = useState('0.00');
  const [currencyCode, setCurrencyCode] = useState('USD');
  const [notes, setNotes] = useState('');

  const [instruments, setInstruments] = useState<InstrumentOption[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Fetch available instruments on mount
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
        if (res.instruments && res.instruments.length > 0) {
          setInstruments(res.instruments);
          setSymbol(res.instruments[0].symbol);
          setCurrencyCode(res.instruments[0].currencyCode);
        }
      })
      .catch((err) => {
        console.warn('Failed to fetch instruments:', err);
      });
  }, []);

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


  // Handle instrument selection
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
        if (!symbol) throw new Error('Please select or specify a symbol.');
        if (!quantity || parseFloat(quantity) <= 0) throw new Error('Quantity must be greater than zero.');
        if (!price || parseFloat(price) < 0) throw new Error('Price cannot be negative.');
      } else if (type === 'DEPOSIT' || type === 'WITHDRAWAL') {
        if (!amount || parseFloat(amount) <= 0) throw new Error('Amount must be greater than zero.');
      } else if (type === 'DIVIDEND') {
        if (!symbol) throw new Error('Please select or specify a symbol for dividend.');
        if (!amount || parseFloat(amount) <= 0) throw new Error('Amount must be greater than zero.');
      }

      const input: any = {
        type,
        tradeDate,
        currencyCode,
        notes: notes || undefined,
        fee: fee ? fee : undefined,
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
      }

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
        // Reset form
        setQuantity('');
        setPrice('');
        setAmount('');
        setNotes('');
      }
    } catch (err: any) {
      console.error('Error adding transaction:', err);
      setError(err?.message || 'Failed to add transaction. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div className="modal-backdrop" onClick={onClose}>
      <div className="modal-container" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h2 className="modal-title">Record Transaction</h2>
          <button className="modal-close-btn" onClick={onClose} type="button">
            &times;
          </button>
        </div>

        {/* Transaction Type Segmented Control */}
        <div className="type-tabs">
          {(['BUY', 'SELL', 'DEPOSIT', 'WITHDRAWAL', 'DIVIDEND'] as TransactionType[]).map((t) => (
            <button
              key={t}
              type="button"
              className={`type-tab ${type === t ? 'active' : ''} ${type === 'SELL' ? 'sell' : ''}`}
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

          {/* Instrument Selection for BUY/SELL/DIVIDEND */}
          {(type === 'BUY' || type === 'SELL' || type === 'DIVIDEND') && (
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

          {/* Trade Date & Currency */}
          <div className="form-row">
            <div className="form-group">
              <label>Date</label>
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
              >
                <option value="USD">USD</option>
                <option value="EUR">EUR</option>
                <option value="GBP">GBP</option>
                <option value="AUD">AUD</option>
              </select>
            </div>
          </div>

          {/* Trade specific: Quantity & Price */}
          {(type === 'BUY' || type === 'SELL') && (
            <div className="form-row">
              <div className="form-group">
                <label>Quantity</label>
                <input
                  type="number"
                  step="any"
                  placeholder="0.00"
                  className="form-input"
                  value={quantity}
                  onChange={(e) => handleQuantityChange(e.target.value)}
                  required
                />
              </div>

              <div className="form-group">
                <label>Price per Share</label>
                <input
                  type="number"
                  step="any"
                  placeholder="0.00"
                  className="form-input"
                  value={price}
                  onChange={(e) => handlePriceChange(e.target.value)}
                  required
                />
              </div>
            </div>
          )}

          {/* Total Amount & Brokerage Fee */}
          <div className="form-row">
            <div className="form-group">
              <label>Total Amount</label>
              <input
                type="number"
                step="any"
                placeholder="0.00"
                className="form-input"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
                required
              />
              {(type === 'BUY' || type === 'SELL') && (
                <span className="form-calculated-note">
                  Auto-calculated from quantity &times; price
                </span>
              )}
            </div>

            {(type === 'BUY' || type === 'SELL') && (
              <div className="form-group">
                <label>Brokerage Fee</label>
                <input
                  type="number"
                  step="any"
                  placeholder="0.00"
                  className="form-input"
                  value={fee}
                  onChange={(e) => setFee(e.target.value)}
                />
              </div>
            )}
          </div>

          {/* Notes */}
          <div className="form-group">
            <label>Notes (Optional)</label>
            <input
              type="text"
              placeholder="e.g. Dollar-cost averaging, dividend reinvestment..."
              className="form-input"
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
            />
          </div>

          {/* Modal Actions */}
          <div className="modal-actions">
            <button
              type="button"
              className="btn-secondary"
              onClick={onClose}
              disabled={loading}
            >
              Cancel
            </button>
            <button type="submit" className="btn-primary" disabled={loading}>
              {loading ? 'Recording...' : `Record ${type}`}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

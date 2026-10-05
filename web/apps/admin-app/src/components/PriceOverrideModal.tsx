import React, { useState } from 'react';
import { Modal, Input, Button } from '@graphfolio/ui';
import './PriceOverrideModal.css';

export interface PriceOverrideData {
  symbol: string;
  priceDate: string;
  price: string;
  reason: string;
  recomputeValuations: boolean;
}

interface PriceOverrideModalProps {
  isOpen: boolean;
  onClose: () => void;
  defaultSymbol?: string;
  onSubmit: (data: PriceOverrideData) => Promise<void>;
}

export const PriceOverrideModal: React.FC<PriceOverrideModalProps> = ({
  isOpen,
  onClose,
  defaultSymbol = 'AAPL',
  onSubmit,
}) => {
  const [symbol, setSymbol] = useState(defaultSymbol);
  const [prevDefaultSymbol, setPrevDefaultSymbol] = useState(defaultSymbol);
  const [priceDate, setPriceDate] = useState(() => new Date().toISOString().split('T')[0]);
  const [price, setPrice] = useState('');
  const [reason, setReason] = useState('');
  const [recomputeValuations, setRecomputeValuations] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  if (defaultSymbol !== prevDefaultSymbol) {
    setPrevDefaultSymbol(defaultSymbol);
    setSymbol(defaultSymbol);
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const cleanSymbol = symbol.trim().toUpperCase();
    const cleanPrice = price.trim();
    const cleanDate = priceDate.trim();
    const cleanReason = reason.trim();

    if (!cleanSymbol || !cleanDate || !cleanPrice) {
      setError('Symbol, date, and closing price are required.');
      return;
    }

    const numPrice = Number(cleanPrice);
    if (isNaN(numPrice) || numPrice <= 0) {
      setError('Closing price must be a valid positive number.');
      return;
    }

    try {
      setIsSubmitting(true);
      setError(null);
      await onSubmit({
        symbol: cleanSymbol,
        priceDate: cleanDate,
        price: cleanPrice,
        reason: cleanReason,
        recomputeValuations,
      });
      setPrice('');
      setReason('');
      onClose();
    } catch (err: any) {
      setError(err?.message || 'Failed to submit price override.');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Manual Price Override"
      size="md"
      footer={
        <>
          <Button variant="ghost" onClick={onClose} disabled={isSubmitting}>
            Cancel
          </Button>
          <Button variant="danger" onClick={handleSubmit} isLoading={isSubmitting}>
            Apply Price Override
          </Button>
        </>
      }
    >
      <form onSubmit={handleSubmit} className="price-override-form">
        <div className="price-override-form__warning">
          <strong>⚠️ Caution:</strong> Applying a manual closing price override modifies the
          authoritative pricing ledger and will trigger a retroactive recalculation of all portfolio
          valuations and time-weighted returns (TWR) for this asset.
        </div>

        {error && <div className="instrument-form__error">{error}</div>}

        <div className="price-override-form__row">
          <Input
            label="Ticker Symbol *"
            value={symbol}
            onChange={(e) => setSymbol(e.target.value)}
            disabled={isSubmitting}
            required
          />
          <Input
            label="Effective Date *"
            type="date"
            value={priceDate}
            onChange={(e) => setPriceDate(e.target.value)}
            disabled={isSubmitting}
            required
          />
        </div>

        <Input
          label="Corrected Closing Price *"
          type="number"
          step="0.000001"
          placeholder="e.g. 182.50"
          value={price}
          onChange={(e) => setPrice(e.target.value)}
          disabled={isSubmitting}
          required
        />

        <Input
          label="Audit Justification / Source *"
          placeholder="e.g. Corporate split adjustment / Bad vendor feed tick"
          value={reason}
          onChange={(e) => setReason(e.target.value)}
          disabled={isSubmitting}
          required
        />

        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginTop: '0.25rem' }}>
          <input
            id="recompute-checkbox"
            type="checkbox"
            checked={recomputeValuations}
            onChange={(e) => setRecomputeValuations(e.target.checked)}
            disabled={isSubmitting}
            style={{ width: '1rem', height: '1rem', cursor: 'pointer', accentColor: 'var(--accent-primary, #6366f1)' }}
          />
          <label
            htmlFor="recompute-checkbox"
            style={{ fontSize: '0.85rem', color: 'var(--text-secondary, #a1a1aa)', cursor: 'pointer' }}
          >
            Trigger retroactive portfolio valuation recalculations
          </label>
        </div>
      </form>
    </Modal>
  );
};

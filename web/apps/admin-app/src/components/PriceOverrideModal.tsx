import React, { useState } from 'react';
import { Modal, Input, Button } from '@graphfolio/ui';
import './PriceOverrideModal.css';

export interface PriceOverrideData {
  symbol: string;
  priceDate: string;
  price: string;
  currencyCode: string;
  reason: string;
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
  const [priceDate, setPriceDate] = useState(() => new Date().toISOString().split('T')[0]);
  const [price, setPrice] = useState('');
  const [currencyCode, setCurrencyCode] = useState('USD');
  const [reason, setReason] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!symbol.trim() || !priceDate.trim() || !price.trim()) {
      setError('Symbol, date, and closing price are required.');
      return;
    }

    try {
      setIsSubmitting(true);
      setError(null);
      await onSubmit({
        symbol: symbol.trim().toUpperCase(),
        priceDate: priceDate.trim(),
        price: price.trim(),
        currencyCode: currencyCode.trim().toUpperCase(),
        reason: reason.trim(),
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

        <div className="price-override-form__row">
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
            label="Currency *"
            value={currencyCode}
            onChange={(e) => setCurrencyCode(e.target.value)}
            disabled={isSubmitting}
            required
          />
        </div>

        <Input
          label="Audit Justification / Source *"
          placeholder="e.g. Corporate split adjustment / Bad vendor feed tick"
          value={reason}
          onChange={(e) => setReason(e.target.value)}
          disabled={isSubmitting}
          required
        />
      </form>
    </Modal>
  );
};

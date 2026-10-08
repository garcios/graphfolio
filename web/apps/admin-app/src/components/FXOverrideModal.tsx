import React, { useState, useEffect } from 'react';
import { Modal, Input, Button } from '@graphfolio/ui';
import './FXOverrideModal.css';

export interface FXOverrideData {
  baseCurrency: string;
  quoteCurrency: string;
  rateDate: string;
  rate: string;
  reason: string;
  recomputeValuations: boolean;
}

interface FXOverrideModalProps {
  isOpen: boolean;
  onClose: () => void;
  defaultBaseCurrency?: string;
  defaultQuoteCurrency?: string;
  defaultRateDate?: string;
  defaultRate?: string;
  onSubmit: (data: FXOverrideData) => Promise<void>;
}

export const FXOverrideModal: React.FC<FXOverrideModalProps> = ({
  isOpen,
  onClose,
  defaultBaseCurrency = 'EUR',
  defaultQuoteCurrency = 'USD',
  defaultRateDate,
  defaultRate = '',
  onSubmit,
}) => {
  const [baseCurrency, setBaseCurrency] = useState(defaultBaseCurrency);
  const [quoteCurrency, setQuoteCurrency] = useState(defaultQuoteCurrency);
  const [rateDate, setRateDate] = useState(() => defaultRateDate || new Date().toISOString().split('T')[0]);
  const [rate, setRate] = useState(defaultRate);
  const [reason, setReason] = useState('');
  const [recomputeValuations, setRecomputeValuations] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (isOpen) {
      setBaseCurrency(defaultBaseCurrency);
      setQuoteCurrency(defaultQuoteCurrency);
      setRateDate(defaultRateDate || new Date().toISOString().split('T')[0]);
      setRate(defaultRate);
      setReason('');
      setError(null);
    }
  }, [isOpen, defaultBaseCurrency, defaultQuoteCurrency, defaultRateDate, defaultRate]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const cleanBase = baseCurrency.trim().toUpperCase();
    const cleanQuote = quoteCurrency.trim().toUpperCase();
    const cleanRate = rate.trim();
    const cleanDate = rateDate.trim();
    const cleanReason = reason.trim();

    if (!cleanBase || cleanBase.length !== 3) {
      setError('Base currency must be a 3-letter ISO code (e.g. EUR).');
      return;
    }
    if (!cleanQuote || cleanQuote.length !== 3) {
      setError('Quote currency must be a 3-letter ISO code (e.g. USD).');
      return;
    }
    if (cleanBase === cleanQuote) {
      setError('Base and quote currencies cannot be identical.');
      return;
    }
    if (!cleanDate) {
      setError('Rate date is required.');
      return;
    }
    if (!cleanRate) {
      setError('Exchange rate is required.');
      return;
    }

    const numRate = Number(cleanRate);
    if (isNaN(numRate) || numRate <= 0) {
      setError('Exchange rate must be a valid positive number.');
      return;
    }

    try {
      setIsSubmitting(true);
      setError(null);
      await onSubmit({
        baseCurrency: cleanBase,
        quoteCurrency: cleanQuote,
        rateDate: cleanDate,
        rate: cleanRate,
        reason: cleanReason,
        recomputeValuations,
      });
      setRate('');
      setReason('');
      onClose();
    } catch (err: any) {
      setError(err?.message || 'Failed to record FX rate override.');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Record Foreign Exchange (FX) Rate Override"
      size="md"
      footer={
        <>
          <Button variant="ghost" onClick={onClose} disabled={isSubmitting}>
            Cancel
          </Button>
          <Button variant="danger" onClick={handleSubmit} isLoading={isSubmitting}>
            Apply FX Override
          </Button>
        </>
      }
    >
      <form onSubmit={handleSubmit} className="fx-override-form">
        <div className="fx-override-form__warning">
          <strong>⚠️ Caution:</strong> Applying a manual foreign exchange rate override alters historical
          FX conversion in the authoritative ledger and optionally triggers retroactive recalculation of portfolio
          valuations and time-weighted returns (TWR).
        </div>

        {error && <div className="fx-override-form__error">{error}</div>}

        <div className="fx-override-form__row">
          <Input
            label="Base Currency"
            value={baseCurrency}
            onChange={(e) => setBaseCurrency(e.target.value.toUpperCase())}
            placeholder="EUR"
            maxLength={3}
            required
            disabled={isSubmitting}
          />
          <Input
            label="Quote Currency"
            value={quoteCurrency}
            onChange={(e) => setQuoteCurrency(e.target.value.toUpperCase())}
            placeholder="USD"
            maxLength={3}
            required
            disabled={isSubmitting}
          />
        </div>

        <div className="fx-override-form__row">
          <Input
            label="Rate Date"
            type="date"
            value={rateDate}
            onChange={(e) => setRateDate(e.target.value)}
            required
            disabled={isSubmitting}
          />
          <Input
            label="Exchange Rate (1 Base = X Quote)"
            type="number"
            step="any"
            min="0.000001"
            value={rate}
            onChange={(e) => setRate(e.target.value)}
            placeholder="1.085000"
            required
            disabled={isSubmitting}
          />
        </div>

        <Input
          label="Audit Justification / Note"
          value={reason}
          onChange={(e) => setReason(e.target.value)}
          placeholder="e.g. ECB fixing correction, holiday fallback adjustment"
          disabled={isSubmitting}
        />

        <label className="fx-override-form__checkbox">
          <input
            type="checkbox"
            checked={recomputeValuations}
            onChange={(e) => setRecomputeValuations(e.target.checked)}
            disabled={isSubmitting}
          />
          <span>Retroactively recompute portfolio valuations affected by this pair</span>
        </label>
      </form>
    </Modal>
  );
};

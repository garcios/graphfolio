import React, { useState, useEffect, useMemo } from 'react';
import { Modal, Input, Select, Button } from '@graphfolio/ui';
import './AddAssetPriceModal.css';

export interface InstrumentOption {
  symbol: string;
  name: string;
  currencyCode: string;
}

export interface AddAssetPriceSubmitData {
  symbol: string;
  priceDate: string;
  price: string;
  reason?: string;
  recomputeValuations: boolean;
}

export interface ExistingPriceCheck {
  symbol?: string;
  priceDate: string;
  priceAmount: string;
  currencyCode?: string;
}

export interface AddAssetPriceModalProps {
  isOpen: boolean;
  onClose: () => void;
  instruments: InstrumentOption[];
  defaultSymbol?: string;
  defaultDate?: string;
  defaultPrice?: string;
  existingPrices?: ExistingPriceCheck[];
  onSubmit: (data: AddAssetPriceSubmitData) => Promise<void>;
}

export function validateAddAssetPriceInput(
  symbol: string,
  priceDate: string,
  price: string,
  todayISO: string = new Date().toISOString().split('T')[0]
): { isValid: boolean; error?: string; cleanSymbol?: string; cleanDate?: string; cleanPrice?: string } {
  const cleanSymbol = symbol.trim().toUpperCase();
  const cleanDate = priceDate.trim();
  const cleanPrice = price.trim();

  if (!cleanSymbol || !cleanDate || !cleanPrice) {
    return { isValid: false, error: 'Symbol, date, and closing price are required.' };
  }

  if (cleanDate > todayISO) {
    return { isValid: false, error: 'Price date cannot be in the future.' };
  }

  const numPrice = Number(cleanPrice);
  if (isNaN(numPrice) || numPrice <= 0) {
    return { isValid: false, error: 'Closing price must be a valid positive number.' };
  }

  return { isValid: true, cleanSymbol, cleanDate, cleanPrice };
}

export function detectPriceConflict(
  symbol: string,
  priceDate: string,
  existingPrices: ExistingPriceCheck[] = []
): ExistingPriceCheck | null {
  if (!existingPrices || existingPrices.length === 0) return null;
  const cleanSymbol = symbol.trim().toUpperCase();
  return (
    existingPrices.find(
      (p) => (!p.symbol || p.symbol.toUpperCase() === cleanSymbol) && p.priceDate === priceDate
    ) || null
  );
}

export const AddAssetPriceModal: React.FC<AddAssetPriceModalProps> = ({
  isOpen,
  onClose,
  instruments,
  defaultSymbol = '',
  defaultDate,
  defaultPrice = '',
  existingPrices = [],
  onSubmit,
}) => {
  const getTodayISO = () => new Date().toISOString().split('T')[0];

  const [symbol, setSymbol] = useState(defaultSymbol);
  const [priceDate, setPriceDate] = useState(() => defaultDate || getTodayISO());
  const [price, setPrice] = useState(defaultPrice);
  const [reason, setReason] = useState('');
  const [recomputeValuations, setRecomputeValuations] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Synchronize state whenever modal opens or defaults change
  useEffect(() => {
    if (isOpen) {
      const initialSymbol = defaultSymbol || (instruments.length > 0 ? instruments[0].symbol : '');
      setSymbol(initialSymbol);
      setPriceDate(defaultDate || getTodayISO());
      setPrice(defaultPrice || '');
      setReason('');
      setRecomputeValuations(true);
      setError(null);
    }
  }, [isOpen, defaultSymbol, defaultDate, defaultPrice, instruments]);

  // Resolve selected instrument details and quote currency
  const selectedInstrument = useMemo(() => {
    return instruments.find((i) => i.symbol.toUpperCase() === symbol.toUpperCase());
  }, [instruments, symbol]);

  const currencyCode = selectedInstrument?.currencyCode || 'USD';

  // Conflict detection: Check if an existing mark exists for this symbol and date
  const existingRecord = useMemo(() => {
    return detectPriceConflict(symbol, priceDate, existingPrices);
  }, [existingPrices, symbol, priceDate]);

  const isUpdatingExisting = Boolean(existingRecord);

  const instrumentOptions = useMemo(() => {
    return instruments.map((inst) => ({
      label: `${inst.symbol} — ${inst.name} [${inst.currencyCode}]`,
      value: inst.symbol,
    }));
  }, [instruments]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const validation = validateAddAssetPriceInput(symbol, priceDate, price, getTodayISO());
    if (!validation.isValid || !validation.cleanSymbol || !validation.cleanDate || !validation.cleanPrice) {
      setError(validation.error || 'Invalid form input.');
      return;
    }

    const cleanReason = reason.trim();

    try {
      setIsSubmitting(true);
      setError(null);
      await onSubmit({
        symbol: validation.cleanSymbol,
        priceDate: validation.cleanDate,
        price: validation.cleanPrice,
        reason: cleanReason || undefined,
        recomputeValuations,
      });
      onClose();
    } catch (err: any) {
      setError(err?.message || 'Failed to record asset price.');
    } finally {
      setIsSubmitting(false);
    }
  };

  const modalTitle = isUpdatingExisting
    ? `Update Closing Price: ${symbol || 'Asset'}`
    : `Add Asset Price${symbol ? `: ${symbol}` : ''}`;

  return (
    <Modal
      isOpen={isOpen}
      onClose={isSubmitting ? () => {} : onClose}
      title={modalTitle}
      size="md"
      footer={
        <>
          <Button variant="ghost" onClick={onClose} disabled={isSubmitting}>
            Cancel
          </Button>
          <Button
            variant={isUpdatingExisting ? 'secondary' : 'primary'}
            onClick={handleSubmit}
            isLoading={isSubmitting}
          >
            {isUpdatingExisting ? 'Update Closing Price' : 'Record Asset Price'}
          </Button>
        </>
      }
    >
      <form onSubmit={handleSubmit} className="add-price-form">
        {existingRecord && (
          <div className="add-price-form__notice">
            <span>ℹ️</span>
            <div>
              <strong>Existing Mark Found:</strong> A closing price of{' '}
              <strong>
                {existingRecord.priceAmount} {existingRecord.currencyCode || currencyCode}
              </strong>{' '}
              is currently recorded for <strong>{priceDate}</strong>. Submitting will update this record in the authoritative ledger.
            </div>
          </div>
        )}

        {error && <div className="add-price-form__error">{error}</div>}

        <div className="add-price-form__row">
          {instrumentOptions.length > 0 ? (
            <Select
              label="Tradable Asset *"
              value={symbol}
              onChange={(e) => setSymbol(e.target.value)}
              options={instrumentOptions}
              disabled={isSubmitting}
              required
            />
          ) : (
            <Input
              label="Ticker Symbol *"
              value={symbol}
              onChange={(e) => setSymbol(e.target.value.toUpperCase())}
              placeholder="e.g. AAPL"
              disabled={isSubmitting}
              required
            />
          )}

          <Input
            label="Pricing Date *"
            type="date"
            value={priceDate}
            max={getTodayISO()}
            onChange={(e) => setPriceDate(e.target.value)}
            disabled={isSubmitting}
            required
          />
        </div>

        <div className="add-price-form__price-group">
          <label className="gf-input-label">Closing Price ({currencyCode}) *</label>
          <div className="add-price-form__price-input-wrapper">
            <Input
              type="number"
              step="any"
              placeholder="e.g. 182.50"
              value={price}
              onChange={(e) => setPrice(e.target.value)}
              disabled={isSubmitting}
              required
              style={{ paddingRight: '4.5rem' }}
            />
            <span className="add-price-form__currency-badge">{currencyCode}</span>
          </div>
        </div>

        <Input
          label="Audit Source / Notes"
          placeholder="e.g. Initial valuation mark, AFR EOD sheet, OTC quote (default: manual)"
          value={reason}
          onChange={(e) => setReason(e.target.value)}
          disabled={isSubmitting}
          helperText="Optional note explaining the provenance or reason for manual entry."
        />

        <div className="add-price-form__checkbox-row">
          <input
            id="add-price-recompute"
            type="checkbox"
            checked={recomputeValuations}
            onChange={(e) => setRecomputeValuations(e.target.checked)}
            disabled={isSubmitting}
            style={{ marginTop: '0.2rem', cursor: 'pointer', accentColor: 'var(--accent-primary, #6366f1)' }}
          />
          <label htmlFor="add-price-recompute" className="add-price-form__checkbox-label">
            <span className="add-price-form__checkbox-title">
              Recompute affected portfolio valuations & returns
            </span>
            <span className="add-price-form__checkbox-desc">
              Retroactively recalibrates daily portfolio valuations and time-weighted returns (TWR) for portfolios holding this asset.
            </span>
          </label>
        </div>
      </form>
    </Modal>
  );
};

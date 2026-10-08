import React, { useState, useMemo } from 'react';
import { Modal, Input, Select, Button } from '@graphfolio/ui';
import './AddInstrumentModal.css';

export interface NewInstrumentData {
  symbol: string;
  exchangeCode: string;
  name: string;
  assetClass: string;
  currencyCode: string;
  isin?: string;
}

export interface ExchangeOption {
  code: string;
  name: string;
}

interface AddInstrumentModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: NewInstrumentData) => Promise<void>;
  exchanges: ExchangeOption[];
  exchangesLoading?: boolean;
  exchangesError?: string | null;
}

const ISIN_REGEX = /^[A-Z]{2}[A-Z0-9]{9}[0-9]$/;

export const AddInstrumentModal: React.FC<AddInstrumentModalProps> = ({
  isOpen,
  onClose,
  onSubmit,
  exchanges,
  exchangesLoading = false,
  exchangesError = null,
}) => {
  const [symbol, setSymbol] = useState('');
  const [exchangeCode, setExchangeCode] = useState('');
  const [name, setName] = useState('');
  const [assetClass, setAssetClass] = useState('EQUITY');
  const [currencyCode, setCurrencyCode] = useState('USD');
  const [isin, setIsin] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const exchangeOptions = useMemo(
    () => exchanges.map((x) => ({ label: `${x.code} — ${x.name}`, value: x.code })),
    [exchanges]
  );

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const cleanSymbol = symbol.trim().toUpperCase();
    const cleanExchange = exchangeCode.trim().toUpperCase();
    const cleanName = name.trim();
    const cleanCurrency = currencyCode.trim().toUpperCase();
    const cleanIsin = isin.trim().toUpperCase();

    if (!cleanSymbol || !cleanExchange || !cleanName) {
      setError('Symbol, exchange code, and instrument name are required.');
      return;
    }

    if (cleanCurrency.length !== 3) {
      setError('Currency code must be exactly 3 characters (e.g. USD, EUR, GBP).');
      return;
    }

    if (cleanIsin && !ISIN_REGEX.test(cleanIsin)) {
      setError('ISIN must be 12 alphanumeric characters conforming to ISO 6166 (e.g. US0378331005).');
      return;
    }

    try {
      setIsSubmitting(true);
      setError(null);
      await onSubmit({
        symbol: cleanSymbol,
        exchangeCode: cleanExchange,
        name: cleanName,
        assetClass,
        currencyCode: cleanCurrency,
        isin: cleanIsin || undefined,
      });
      // Reset form
      setSymbol('');
      setExchangeCode('');
      setName('');
      setAssetClass('EQUITY');
      setCurrencyCode('USD');
      setIsin('');
      onClose();
    } catch (err: any) {
      setError(err?.message || 'Failed to register instrument.');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Register Tradable Asset"
      size="md"
      footer={
        <>
          <Button variant="ghost" onClick={onClose} disabled={isSubmitting}>
            Cancel
          </Button>
          <Button variant="primary" onClick={handleSubmit} isLoading={isSubmitting}>
            Register Asset
          </Button>
        </>
      }
    >
      <form onSubmit={handleSubmit} className="instrument-form">
        {error && <div className="instrument-form__error">{error}</div>}

        <div className="instrument-form__row">
          <Input
            label="Ticker / Symbol *"
            placeholder="e.g. NVDA"
            value={symbol}
            onChange={(e) => setSymbol(e.target.value)}
            disabled={isSubmitting}
            required
          />
          <Select
            id="instrument-exchange-code"
            label="Exchange Code *"
            placeholder={exchangesLoading ? 'Loading exchanges…' : 'Select Exchange'}
            value={exchangeCode}
            onChange={(e) => setExchangeCode(e.target.value)}
            options={exchangeOptions}
            disabled={isSubmitting || exchangesLoading || !!exchangesError}
            error={exchangesError ?? undefined}
            required
          />
        </div>

        <Input
          label="Full Legal / Commercial Name *"
          placeholder="e.g. NVIDIA Corporation"
          value={name}
          onChange={(e) => setName(e.target.value)}
          disabled={isSubmitting}
          required
        />

        <div className="instrument-form__row">
          <Select
            label="Asset Class"
            value={assetClass}
            onChange={(e) => setAssetClass(e.target.value)}
            disabled={isSubmitting}
            options={[
              { label: 'Equity / Stock', value: 'EQUITY' },
              { label: 'Exchange-Traded Fund (ETF)', value: 'ETF' },
              { label: 'Mutual / Index Fund', value: 'FUND' },
              { label: 'Fixed Income / Bond', value: 'BOND' },
              { label: 'Cryptocurrency', value: 'CRYPTO' },
              { label: 'Cash Equivalent', value: 'CASH_EQUIVALENT' },
            ]}
          />
          <Input
            label="Trading Currency *"
            placeholder="USD"
            value={currencyCode}
            onChange={(e) => setCurrencyCode(e.target.value)}
            disabled={isSubmitting}
            required
          />
        </div>

        <Input
          label="ISIN (Optional)"
          placeholder="e.g. US67066G1040"
          value={isin}
          onChange={(e) => setIsin(e.target.value)}
          disabled={isSubmitting}
        />
      </form>
    </Modal>
  );
};

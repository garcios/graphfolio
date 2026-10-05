import React, { useState } from 'react';
import { Modal, Input, Select, Button } from '@graphfolio/ui';
import './AddInstrumentModal.css';

export interface NewInstrumentData {
  symbol: string;
  name: string;
  assetClass: string;
  currencyCode: string;
}

interface AddInstrumentModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: NewInstrumentData) => Promise<void>;
}

export const AddInstrumentModal: React.FC<AddInstrumentModalProps> = ({
  isOpen,
  onClose,
  onSubmit,
}) => {
  const [symbol, setSymbol] = useState('');
  const [name, setName] = useState('');
  const [assetClass, setAssetClass] = useState('EQUITY');
  const [currencyCode, setCurrencyCode] = useState('USD');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!symbol.trim() || !name.trim()) {
      setError('Symbol and instrument name are required.');
      return;
    }

    try {
      setIsSubmitting(true);
      setError(null);
      await onSubmit({
        symbol: symbol.trim().toUpperCase(),
        name: name.trim(),
        assetClass,
        currencyCode: currencyCode.trim().toUpperCase(),
      });
      // Reset form
      setSymbol('');
      setName('');
      setAssetClass('EQUITY');
      setCurrencyCode('USD');
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
          label="Full Legal / Commercial Name *"
          placeholder="e.g. NVIDIA Corporation"
          value={name}
          onChange={(e) => setName(e.target.value)}
          disabled={isSubmitting}
          required
        />

        <Select
          label="Asset Class"
          value={assetClass}
          onChange={(e) => setAssetClass(e.target.value)}
          disabled={isSubmitting}
          options={[
            { label: 'Equity / Stock', value: 'EQUITY' },
            { label: 'Exchange-Traded Fund (ETF)', value: 'ETF' },
            { label: 'Cryptocurrency', value: 'CRYPTO' },
            { label: 'Commodity', value: 'COMMODITY' },
            { label: 'Fixed Income / Bond', value: 'BOND' },
          ]}
        />
      </form>
    </Modal>
  );
};

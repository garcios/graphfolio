import React, { useState } from 'react';
import { Modal, Input, Button } from '@graphfolio/ui';
import './BackfillModal.css';

export interface BackfillSubmitData {
  fromDate: string;
  toDate: string;
  symbols?: string[];
  currencyPairs?: string[];
  backfillAssets: boolean;
  backfillFx: boolean;
  recomputeValuations: boolean;
}

export interface BackfillModalResult {
  success: boolean;
  pricesSynced: number;
  fxRatesSynced: number;
  message: string;
  warnings: string[];
}

export interface BackfillModalProps {
  isOpen: boolean;
  onClose: () => void;
  initialFromDate?: string;
  initialToDate?: string;
  initialSymbols?: string[];
  initialCurrencyPairs?: string[];
  initialBackfillAssets?: boolean;
  initialBackfillFx?: boolean;
  initialRecomputeValuations?: boolean;
  rateLimitRemaining?: number;
  rateLimitBudget?: number;
  onSubmit: (data: BackfillSubmitData) => Promise<BackfillModalResult>;
  onSuccess?: (result: BackfillModalResult) => void;
}

type PresetKey = '30D' | '90D' | 'YTD' | '1Y' | 'MAX';

function formatISODate(d: Date): string {
  return d.toISOString().split('T')[0];
}

function calculatePresetDate(preset: PresetKey): string {
  const now = new Date();
  switch (preset) {
    case '30D': {
      const d = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000);
      return formatISODate(d);
    }
    case '90D': {
      const d = new Date(now.getTime() - 90 * 24 * 60 * 60 * 1000);
      return formatISODate(d);
    }
    case 'YTD': {
      const d = new Date(now.getFullYear(), 0, 1);
      return formatISODate(d);
    }
    case '1Y': {
      const d = new Date(now.getTime() - 365 * 24 * 60 * 60 * 1000);
      return formatISODate(d);
    }
    case 'MAX': {
      const d = new Date(now.getFullYear() - 5, now.getMonth(), now.getDate());
      return formatISODate(d);
    }
  }
}

export const BackfillModal: React.FC<BackfillModalProps> = ({
  isOpen,
  onClose,
  initialFromDate,
  initialToDate,
  initialSymbols = [],
  initialCurrencyPairs = [],
  initialBackfillAssets = true,
  initialBackfillFx = true,
  initialRecomputeValuations = true,
  rateLimitRemaining = 500,
  rateLimitBudget = 500,
  onSubmit,
  onSuccess,
}) => {
  const [today] = useState(() => formatISODate(new Date()));

  const [fromDate, setFromDate] = useState(() => initialFromDate || calculatePresetDate('30D'));
  const [toDate, setToDate] = useState(() => initialToDate || formatISODate(new Date()));
  const [activePreset, setActivePreset] = useState<PresetKey | 'CUSTOM'>('30D');

  const [backfillAssets, setBackfillAssets] = useState(initialBackfillAssets);
  const [assetScope, setAssetScope] = useState<'all' | 'specific'>(() =>
    initialSymbols.length > 0 ? 'specific' : 'all'
  );
  const [symbols, setSymbols] = useState<string[]>(initialSymbols);
  const [newSymbolInput, setNewSymbolInput] = useState('');

  const [backfillFx, setBackfillFx] = useState(initialBackfillFx);
  const [fxScope, setFxScope] = useState<'all' | 'specific'>(() =>
    initialCurrencyPairs.length > 0 ? 'specific' : 'all'
  );
  const [currencyPairs, setCurrencyPairs] = useState<string[]>(initialCurrencyPairs);
  const [newPairInput, setNewPairInput] = useState('');

  const [recomputeValuations, setRecomputeValuations] = useState(initialRecomputeValuations);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<BackfillModalResult | null>(null);

  // Sync state when modal opens or initial props change
  const [prevIsOpen, setPrevIsOpen] = useState(isOpen);
  if (isOpen && !prevIsOpen) {
    setPrevIsOpen(true);
    setFromDate(initialFromDate || calculatePresetDate('30D'));
    setToDate(initialToDate || today);
    setActivePreset('30D');
    setBackfillAssets(initialBackfillAssets);
    setAssetScope(initialSymbols.length > 0 ? 'specific' : 'all');
    setSymbols(initialSymbols);
    setBackfillFx(initialBackfillFx);
    setFxScope(initialCurrencyPairs.length > 0 ? 'specific' : 'all');
    setCurrencyPairs(initialCurrencyPairs);
    setRecomputeValuations(initialRecomputeValuations);
    setError(null);
    setResult(null);
  } else if (!isOpen && prevIsOpen) {
    setPrevIsOpen(false);
  }

  const handlePresetClick = (preset: PresetKey) => {
    setActivePreset(preset);
    setFromDate(calculatePresetDate(preset));
    setToDate(today);
  };

  const handleFromDateChange = (val: string) => {
    setFromDate(val);
    setActivePreset('CUSTOM');
  };

  const handleToDateChange = (val: string) => {
    setToDate(val);
    setActivePreset('CUSTOM');
  };

  const handleAddSymbol = (e?: React.FormEvent) => {
    if (e) e.preventDefault();
    const clean = newSymbolInput.trim().toUpperCase();
    if (!clean) return;
    if (!symbols.includes(clean)) {
      setSymbols([...symbols, clean]);
    }
    setNewSymbolInput('');
  };

  const handleRemoveSymbol = (sym: string) => {
    setSymbols(symbols.filter((s) => s !== sym));
  };

  const handleAddPair = (e?: React.FormEvent) => {
    if (e) e.preventDefault();
    let clean = newPairInput.trim().toUpperCase().replace(/-/g, '/');
    if (!clean) return;
    if (clean.length === 6 && !clean.includes('/')) {
      clean = `${clean.slice(0, 3)}/${clean.slice(3)}`;
    }
    if (!currencyPairs.includes(clean)) {
      setCurrencyPairs([...currencyPairs, clean]);
    }
    setNewPairInput('');
  };

  const handleRemovePair = (pair: string) => {
    setCurrencyPairs(currencyPairs.filter((p) => p !== pair));
  };

  // Telemetry API call estimation
  const startMs = new Date(fromDate).getTime();
  const endMs = new Date(toDate).getTime();
  const daysDiff = Math.max(1, Math.round((endMs - startMs) / (1000 * 3600 * 24)));

  let estimatedCalls = 0;
  if (backfillAssets) {
    const assetCount = assetScope === 'specific' ? Math.max(1, symbols.length) : 10;
    estimatedCalls += assetCount * Math.max(1, Math.ceil(daysDiff / 365));
  }
  if (backfillFx) {
    const fxCount = fxScope === 'specific' ? Math.max(1, currencyPairs.length) : 7;
    estimatedCalls += fxCount * Math.max(1, Math.ceil(daysDiff / 90));
  }

  const isHighTokenConsumption = estimatedCalls > rateLimitRemaining * 0.5;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!fromDate || !toDate) {
      setError('Start date and end date are required.');
      return;
    }

    if (fromDate > toDate) {
      setError('Start date cannot be after end date.');
      return;
    }

    if (fromDate > today || toDate > today) {
      setError('Date range cannot include future dates.');
      return;
    }

    if (!backfillAssets && !backfillFx) {
      setError('Please select at least one data target: Asset Prices or Foreign Exchange Rates.');
      return;
    }

    if (backfillAssets && assetScope === 'specific' && symbols.length === 0) {
      setError('Please specify at least one asset symbol or choose "All Active Instruments".');
      return;
    }

    if (backfillFx && fxScope === 'specific' && currencyPairs.length === 0) {
      setError('Please specify at least one currency pair or choose "All Active Currencies".');
      return;
    }

    try {
      setIsSubmitting(true);
      const res = await onSubmit({
        fromDate,
        toDate,
        symbols: backfillAssets && assetScope === 'specific' ? symbols : undefined,
        currencyPairs: backfillFx && fxScope === 'specific' ? currencyPairs : undefined,
        backfillAssets,
        backfillFx,
        recomputeValuations,
      });

      setResult(res);
      if (onSuccess) {
        onSuccess(res);
      }
    } catch (err: any) {
      setError(err?.message || 'Failed to execute historical backfill.');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="⚡ Historical Market Data Backfill"
      size="lg"
      footer={
        <>
          <Button variant="ghost" onClick={onClose} disabled={isSubmitting}>
            {result ? 'Close' : 'Cancel'}
          </Button>
          {!result && (
            <Button
              variant="primary"
              onClick={handleSubmit}
              isLoading={isSubmitting}
            >
              ⚡ Execute Backfill
            </Button>
          )}
        </>
      }
    >
      <form onSubmit={handleSubmit} className="backfill-modal-form">
        <p className="backfill-modal__description">
          Pull authoritative historical closing prices and foreign exchange rates across an inclusive
          date range from external market providers into the PostgreSQL ledger.
        </p>

        {error && <div className="backfill-modal__error">{error}</div>}

        {result && (
          <div className="backfill-modal__success">
            <strong>✅ {result.message}</strong>
            <span>
              Synced: <strong>{result.pricesSynced}</strong> asset prices,{' '}
              <strong>{result.fxRatesSynced}</strong> FX fixing rates.
            </span>
            {result.warnings && result.warnings.length > 0 && (
              <ul className="backfill-modal__warnings-list">
                {result.warnings.map((w, idx) => (
                  <li key={idx}>⚠️ {w}</li>
                ))}
              </ul>
            )}
          </div>
        )}

        {!result && (
          <>
            {/* Date Range Section */}
            <div className="backfill-modal__section">
              <h4 className="backfill-modal__section-title">Date Horizon</h4>
              <div className="backfill-modal__date-grid">
                <div className="backfill-modal__date-field">
                  <label htmlFor="backfill-from-date">Start Date (From)</label>
                  <Input
                    id="backfill-from-date"
                    type="date"
                    value={fromDate}
                    max={today}
                    onChange={(e) => handleFromDateChange(e.target.value)}
                  />
                </div>
                <div className="backfill-modal__date-field">
                  <label htmlFor="backfill-to-date">End Date (To)</label>
                  <Input
                    id="backfill-to-date"
                    type="date"
                    value={toDate}
                    max={today}
                    onChange={(e) => handleToDateChange(e.target.value)}
                  />
                </div>
              </div>

              {/* Quick Presets */}
              <div className="backfill-modal__presets">
                <span className="backfill-modal__presets-label">Presets:</span>
                {(['30D', '90D', 'YTD', '1Y', 'MAX'] as PresetKey[]).map((preset) => (
                  <button
                    key={preset}
                    type="button"
                    className={`backfill-modal__preset-chip ${
                      activePreset === preset ? 'backfill-modal__preset-chip--active' : ''
                    }`}
                    onClick={() => handlePresetClick(preset)}
                  >
                    {preset === 'MAX' ? '5 Years (Max)' : preset}
                  </button>
                ))}
              </div>
            </div>

            {/* Scope & Data Feeds Section */}
            <div className="backfill-modal__section">
              <h4 className="backfill-modal__section-title">Target Feeds & Scope</h4>

              {/* Asset Prices Feed */}
              <div
                className={`backfill-modal__feed-card ${
                  backfillAssets ? 'backfill-modal__feed-card--enabled' : ''
                }`}
              >
                <label className="backfill-modal__checkbox-label">
                  <input
                    type="checkbox"
                    checked={backfillAssets}
                    onChange={(e) => setBackfillAssets(e.target.checked)}
                  />
                  <span>Backfill Asset Prices (Equities, ETFs & Funds)</span>
                </label>

                {backfillAssets && (
                  <div className="backfill-modal__feed-suboptions">
                    <div className="backfill-modal__radio-group">
                      <label className="backfill-modal__radio-label">
                        <input
                          type="radio"
                          name="assetScope"
                          value="all"
                          checked={assetScope === 'all'}
                          onChange={() => setAssetScope('all')}
                        />
                        <span>All Active Instruments</span>
                      </label>
                      <label className="backfill-modal__radio-label">
                        <input
                          type="radio"
                          name="assetScope"
                          value="specific"
                          checked={assetScope === 'specific'}
                          onChange={() => setAssetScope('specific')}
                        />
                        <span>Specific Symbols</span>
                      </label>
                    </div>

                    {assetScope === 'specific' && (
                      <div className="backfill-modal__tag-container">
                        <div className="backfill-modal__tags">
                          {symbols.map((sym) => (
                            <span key={sym} className="backfill-modal__tag">
                              {sym}
                              <button
                                type="button"
                                className="backfill-modal__tag-remove"
                                onClick={() => handleRemoveSymbol(sym)}
                                aria-label={`Remove ${sym}`}
                              >
                                ✕
                              </button>
                            </span>
                          ))}
                        </div>
                        <div className="backfill-modal__tag-input-row">
                          <Input
                            placeholder="Symbol (e.g. AAPL, NVDA)"
                            value={newSymbolInput}
                            onChange={(e) => setNewSymbolInput(e.target.value)}
                            onKeyDown={(e) => {
                              if (e.key === 'Enter') {
                                e.preventDefault();
                                handleAddSymbol();
                              }
                            }}
                          />
                          <Button
                            type="button"
                            variant="secondary"
                            size="sm"
                            onClick={handleAddSymbol}
                          >
                            + Add
                          </Button>
                        </div>
                      </div>
                    )}
                  </div>
                )}
              </div>

              {/* FX Rates Feed */}
              <div
                className={`backfill-modal__feed-card ${
                  backfillFx ? 'backfill-modal__feed-card--enabled' : ''
                }`}
              >
                <label className="backfill-modal__checkbox-label">
                  <input
                    type="checkbox"
                    checked={backfillFx}
                    onChange={(e) => setBackfillFx(e.target.checked)}
                  />
                  <span>Backfill Foreign Exchange Rates (ECB Daily Fixings)</span>
                </label>

                {backfillFx && (
                  <div className="backfill-modal__feed-suboptions">
                    <div className="backfill-modal__radio-group">
                      <label className="backfill-modal__radio-label">
                        <input
                          type="radio"
                          name="fxScope"
                          value="all"
                          checked={fxScope === 'all'}
                          onChange={() => setFxScope('all')}
                        />
                        <span>All Active Currencies</span>
                      </label>
                      <label className="backfill-modal__radio-label">
                        <input
                          type="radio"
                          name="fxScope"
                          value="specific"
                          checked={fxScope === 'specific'}
                          onChange={() => setFxScope('specific')}
                        />
                        <span>Specific Pairs</span>
                      </label>
                    </div>

                    {fxScope === 'specific' && (
                      <div className="backfill-modal__tag-container">
                        <div className="backfill-modal__tags">
                          {currencyPairs.map((pair) => (
                            <span key={pair} className="backfill-modal__tag">
                              {pair}
                              <button
                                type="button"
                                className="backfill-modal__tag-remove"
                                onClick={() => handleRemovePair(pair)}
                                aria-label={`Remove ${pair}`}
                              >
                                ✕
                              </button>
                            </span>
                          ))}
                        </div>
                        <div className="backfill-modal__tag-input-row">
                          <Input
                            placeholder="Pair (e.g. EUR/USD, AUD/USD)"
                            value={newPairInput}
                            onChange={(e) => setNewPairInput(e.target.value)}
                            onKeyDown={(e) => {
                              if (e.key === 'Enter') {
                                e.preventDefault();
                                handleAddPair();
                              }
                            }}
                          />
                          <Button
                            type="button"
                            variant="secondary"
                            size="sm"
                            onClick={handleAddPair}
                          >
                            + Add
                          </Button>
                        </div>
                      </div>
                    )}
                  </div>
                )}
              </div>
            </div>

            {/* Execution Options */}
            <div className="backfill-modal__section">
              <h4 className="backfill-modal__section-title">Valuation Engine Options</h4>
              <label className="backfill-modal__checkbox-label">
                <input
                  type="checkbox"
                  checked={recomputeValuations}
                  onChange={(e) => setRecomputeValuations(e.target.checked)}
                />
                <span style={{ fontSize: '0.8125rem', fontWeight: 500 }}>
                  Recompute historical daily portfolio valuations after backfilling marks
                </span>
              </label>
            </div>

            {/* Telemetry Indicator */}
            <div
              className={`backfill-modal__telemetry ${
                isHighTokenConsumption ? 'backfill-modal__telemetry--warning' : ''
              }`}
            >
              <span className="backfill-modal__telemetry-icon">
                {isHighTokenConsumption ? '⚠️' : 'ℹ️'}
              </span>
              <div className="backfill-modal__telemetry-content">
                <span>
                  Estimated External API Requests: <strong>~{estimatedCalls}</strong> requests (
                  <strong>{rateLimitRemaining}</strong> remaining of {rateLimitBudget} quota).
                </span>
                {isHighTokenConsumption && (
                  <span style={{ fontSize: '0.75rem', opacity: 0.9 }}>
                    Caution: This backfill span may consume more than 50% of the active provider token
                    budget.
                  </span>
                )}
              </div>
            </div>
          </>
        )}
      </form>
    </Modal>
  );
};

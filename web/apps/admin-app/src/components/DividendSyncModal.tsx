import React, { useState } from 'react';
import { Modal, Input, Button } from '@graphfolio/ui';
import './DividendSyncModal.css';

export type DividendSyncScope = 'ALL' | 'SYMBOLS' | 'EXCHANGE';
export type DividendProviderChoice = 'DEFAULT' | 'YAHOO' | 'TWELVE_DATA';

export interface DividendSyncSubmitData {
  symbols?: string[];
  exchangeCode?: string;
  syncHistorical: boolean;
  forceProvider?: string;
}

export interface DividendSyncModalResult {
  success: boolean;
  schedulesSynced: number;
  primarySourceCount: number;
  fallbackSourceCount: number;
  placeholdersUpgraded: number;
  durationMs: number;
  message: string;
  failedSymbols: string[];
}

export interface DividendSyncModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: DividendSyncSubmitData) => Promise<DividendSyncModalResult>;
  onSuccess?: (result: DividendSyncModalResult) => void;
}

export const DividendSyncModal: React.FC<DividendSyncModalProps> = ({
  isOpen,
  onClose,
  onSubmit,
  onSuccess,
}) => {
  const [scope, setScope] = useState<DividendSyncScope>('ALL');
  const [symbolsInput, setSymbolsInput] = useState('');
  const [exchangeInput, setExchangeInput] = useState('');
  const [syncHistorical, setSyncHistorical] = useState(true);
  const [providerChoice, setProviderChoice] = useState<DividendProviderChoice>('DEFAULT');
  const [isSyncing, setIsSyncing] = useState(false);
  const [result, setResult] = useState<DividendSyncModalResult | null>(null);
  const [error, setError] = useState<string | null>(null);

  if (!isOpen) return null;

  const handleClose = () => {
    if (isSyncing) return;
    setResult(null);
    setError(null);
    onClose();
  };

  const parseSymbols = (raw: string): string[] => {
    return raw
      .split(/[,;\s]+/)
      .map((s) => s.trim().toUpperCase())
      .filter((s) => s.length > 0);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setResult(null);

    const parsedSymbols = scope === 'SYMBOLS' ? parseSymbols(symbolsInput) : undefined;
    if (scope === 'SYMBOLS' && (!parsedSymbols || parsedSymbols.length === 0)) {
      setError('Please provide at least one instrument ticker symbol.');
      return;
    }

    const exCode = scope === 'EXCHANGE' ? exchangeInput.trim().toUpperCase() : undefined;
    if (scope === 'EXCHANGE' && !exCode) {
      setError('Please provide an exchange code (e.g. ASX, NASDAQ, NYSE).');
      return;
    }

    const submitData: DividendSyncSubmitData = {
      symbols: parsedSymbols,
      exchangeCode: exCode,
      syncHistorical,
      forceProvider: providerChoice === 'DEFAULT' ? undefined : providerChoice,
    };

    setIsSyncing(true);
    try {
      const res = await onSubmit(submitData);
      setResult(res);
      if (onSuccess) {
        onSuccess(res);
      }
    } catch (err: any) {
      setError(err?.message || 'Dividend sync failed');
    } finally {
      setIsSyncing(false);
    }
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={handleClose}
      title="Sync Corporate Dividends & Forward Cadence"
      className="dividend-sync-modal"
    >
      <form onSubmit={handleSubmit} className="dividend-sync-form">
        <p className="dividend-sync-intro">
          Automate corporate action ingestion, official board declaration scheduling, and forward
          12-month cadence projection across portfolio holdings.
        </p>

        {/* Scope Selector */}
        <div className="form-group">
          <label className="form-label">Ingestion Scope</label>
          <div className="scope-options">
            <label className={`scope-option ${scope === 'ALL' ? 'active' : ''}`}>
              <input
                type="radio"
                name="scope"
                value="ALL"
                checked={scope === 'ALL'}
                onChange={() => setScope('ALL')}
                disabled={isSyncing}
              />
              <span>All Active Holdings</span>
            </label>

            <label className={`scope-option ${scope === 'SYMBOLS' ? 'active' : ''}`}>
              <input
                type="radio"
                name="scope"
                value="SYMBOLS"
                checked={scope === 'SYMBOLS'}
                onChange={() => setScope('SYMBOLS')}
                disabled={isSyncing}
              />
              <span>Specific Tickers</span>
            </label>

            <label className={`scope-option ${scope === 'EXCHANGE' ? 'active' : ''}`}>
              <input
                type="radio"
                name="scope"
                value="EXCHANGE"
                checked={scope === 'EXCHANGE'}
                onChange={() => setScope('EXCHANGE')}
                disabled={isSyncing}
              />
              <span>By Exchange</span>
            </label>
          </div>
        </div>

        {scope === 'SYMBOLS' && (
          <div className="form-group">
            <label className="form-label" htmlFor="symbolsInput">
              Ticker Symbols (comma or space separated)
            </label>
            <Input
              id="symbolsInput"
              placeholder="e.g. AAPL, BHP.AX, MSFT, CBA.AX"
              value={symbolsInput}
              onChange={(e) => setSymbolsInput(e.target.value)}
              disabled={isSyncing}
            />
          </div>
        )}

        {scope === 'EXCHANGE' && (
          <div className="form-group">
            <label className="form-label" htmlFor="exchangeInput">
              Exchange MIC / Code
            </label>
            <Input
              id="exchangeInput"
              placeholder="e.g. ASX, NASDAQ, NYSE, LSE"
              value={exchangeInput}
              onChange={(e) => setExchangeInput(e.target.value)}
              disabled={isSyncing}
            />
          </div>
        )}

        {/* Provider Cascade Mode */}
        <div className="form-group">
          <label className="form-label">Provider Resolution Mode</label>
          <div className="provider-options">
            <label className={`provider-option ${providerChoice === 'DEFAULT' ? 'active' : ''}`}>
              <input
                type="radio"
                name="provider"
                value="DEFAULT"
                checked={providerChoice === 'DEFAULT'}
                onChange={() => setProviderChoice('DEFAULT')}
                disabled={isSyncing}
              />
              <div>
                <strong>Default Resilient Cascade</strong>
                <p>Yahoo Finance unmetered primary ➔ Twelve Data fallback on error/empty</p>
              </div>
            </label>

            <label className={`provider-option ${providerChoice === 'YAHOO' ? 'active' : ''}`}>
              <input
                type="radio"
                name="provider"
                value="YAHOO"
                checked={providerChoice === 'YAHOO'}
                onChange={() => setProviderChoice('YAHOO')}
                disabled={isSyncing}
              />
              <div>
                <strong>Force Yahoo Finance</strong>
                <p>Fast calendarEvents & chart distributions without consuming API credits</p>
              </div>
            </label>

            <label className={`provider-option ${providerChoice === 'TWELVE_DATA' ? 'active' : ''}`}>
              <input
                type="radio"
                name="provider"
                value="TWELVE_DATA"
                checked={providerChoice === 'TWELVE_DATA'}
                onChange={() => setProviderChoice('TWELVE_DATA')}
                disabled={isSyncing}
              />
              <div>
                <strong>Force Twelve Data</strong>
                <p>Strict fallback provider (governed by 8 credits/min rate limiter)</p>
              </div>
            </label>
          </div>
        </div>

        {/* Historical 5Y Series Toggle */}
        <div className="form-group checkbox-group">
          <label className="checkbox-label">
            <input
              type="checkbox"
              checked={syncHistorical}
              onChange={(e) => setSyncHistorical(e.target.checked)}
              disabled={isSyncing}
            />
            <span>Ingest 5Y historical dividend distributions for CAGR & cadence recognition</span>
          </label>
        </div>

        {/* Telemetry info */}
        <div className="telemetry-info-card">
          <span className="telemetry-icon">🛡️</span>
          <div>
            <strong>Override Protection Active:</strong>
            <p>
              Manual adjustments with <code>is_override = true</code> cannot be overwritten by
              automated sync runs. Forward unannounced cycles will be synthesized with{' '}
              <code>is_estimated = true</code>.
            </p>
          </div>
        </div>

        {/* Error notification */}
        {error && <div className="dividend-sync-error">❌ {error}</div>}

        {/* Results view */}
        {result && (
          <div className="dividend-sync-result">
            <h4>✓ Dividend Sync Completed</h4>
            <div className="result-stats">
              <div className="result-stat">
                <span className="stat-label">Schedules Synced</span>
                <span className="stat-value">{result.schedulesSynced}</span>
              </div>
              <div className="result-stat">
                <span className="stat-label">Yahoo (Primary)</span>
                <span className="stat-value">{result.primarySourceCount}</span>
              </div>
              <div className="result-stat">
                <span className="stat-label">Twelve Data (Fallback)</span>
                <span className="stat-value">{result.fallbackSourceCount}</span>
              </div>
              <div className="result-stat">
                <span className="stat-label">Estimates Upgraded</span>
                <span className="stat-value">{result.placeholdersUpgraded}</span>
              </div>
              <div className="result-stat">
                <span className="stat-label">Duration</span>
                <span className="stat-value">{result.durationMs}ms</span>
              </div>
            </div>
            {result.failedSymbols && result.failedSymbols.length > 0 && (
              <p className="failed-symbols-warning">
                ⚠️ Failed symbols: {result.failedSymbols.join(', ')}
              </p>
            )}
          </div>
        )}

        <div className="dividend-sync-actions">
          <Button variant="ghost" type="button" onClick={handleClose} disabled={isSyncing}>
            {result ? 'Done' : 'Cancel'}
          </Button>
          <Button variant="primary" type="submit" isLoading={isSyncing}>
            ⚡ Launch Dividend Sync
          </Button>
        </div>
      </form>
    </Modal>
  );
};

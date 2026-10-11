import React, { useState, useRef } from 'react';
import { client } from '@graphfolio/api-client';
import { parseBrokerCSV, type SupportedBroker, type ParseResult, type NormalizedTransactionRow } from '../services/csv';
import './ImportTransactionsModal.css';

interface ImportTransactionsModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: (updatedPortfolio: any, message: string) => void;
}

type PreviewTab = 'all' | 'ready' | 'duplicates' | 'errors';

export const ImportTransactionsModal: React.FC<ImportTransactionsModalProps> = ({
  isOpen,
  onClose,
  onSuccess,
}) => {
  const [broker, setBroker] = useState<SupportedBroker>('auto');
  const [fileContent, setFileContent] = useState<string | null>(null);
  const [fileName, setFileName] = useState<string | null>(null);
  const [parseResult, setParseResult] = useState<ParseResult | null>(null);
  const [duplicateRefs, setDuplicateRefs] = useState<Set<string>>(new Set());
  const [isParsing, setIsParsing] = useState(false);
  const [isCheckingDuplicates, setIsCheckingDuplicates] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [skipDuplicates, setSkipDuplicates] = useState(true);
  const [activeTab, setActiveTab] = useState<PreviewTab>('all');
  const [errorBanner, setErrorBanner] = useState<string | null>(null);
  const [isDragging, setIsDragging] = useState(false);

  const fileInputRef = useRef<HTMLInputElement>(null);

  if (!isOpen) return null;

  const handleProcessFile = async (name: string, content: string, selectedBroker: SupportedBroker) => {
    setIsParsing(true);
    setErrorBanner(null);
    setParseResult(null);
    setDuplicateRefs(new Set());

    try {
      const result = await parseBrokerCSV(content, selectedBroker);
      setParseResult(result);
      setFileName(name);
      setFileContent(content);

      if (result.validRows.length > 0) {
        setIsCheckingDuplicates(true);
        const refs = result.validRows.map((r) => r.externalRef);
        try {
          const dupRes = await client.query({
            checkTransactionDuplicates: {
              __args: { externalRefs: refs },
            },
          });
          const existing = new Set<string>(dupRes.checkTransactionDuplicates || []);
          setDuplicateRefs(existing);
        } catch (dupErr: any) {
          console.warn('Pre-flight duplicate check failed:', dupErr);
        } finally {
          setIsCheckingDuplicates(false);
        }
      }
    } catch (err: any) {
      setErrorBanner(err?.message || 'Failed to parse CSV file.');
    } finally {
      setIsParsing(false);
    }
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = (event) => {
      const text = event.target?.result as string;
      if (text) {
        handleProcessFile(file.name, text, broker);
      }
    };
    reader.readAsText(file);
  };

  const handleBrokerChange = (newBroker: SupportedBroker) => {
    setBroker(newBroker);
    if (fileName && fileContent) {
      handleProcessFile(fileName, fileContent, newBroker);
    }
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(true);
  };

  const handleDragLeave = () => {
    setIsDragging(false);
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
    const file = e.dataTransfer.files?.[0];
    if (!file) return;

    if (!file.name.toLowerCase().endsWith('.csv')) {
      setErrorBanner('Please upload a valid .csv file.');
      return;
    }

    const reader = new FileReader();
    reader.onload = (event) => {
      const text = event.target?.result as string;
      if (text) {
        handleProcessFile(file.name, text, broker);
      }
    };
    reader.readAsText(file);
  };

  const handleReset = () => {
    setFileName(null);
    setFileContent(null);
    setParseResult(null);
    setDuplicateRefs(new Set());
    setErrorBanner(null);
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

  const handleImportSubmit = async () => {
    if (!parseResult || parseResult.validRows.length === 0) return;

    setIsSubmitting(true);
    setErrorBanner(null);

    const rowsToImport = parseResult.validRows.filter(
      (r) => !skipDuplicates || !duplicateRefs.has(r.externalRef)
    );

    if (rowsToImport.length === 0) {
      setErrorBanner('All valid transactions are marked as duplicates. Uncheck "Skip duplicates" if you wish to re-import.');
      setIsSubmitting(false);
      return;
    }

    try {
      const res = await client.mutation({
        importTransactions: {
          __args: {
            input: {
              transactions: rowsToImport.map((r) => ({
                externalRef: r.externalRef,
                symbol: r.symbol,
                type: r.type,
                tradeDate: r.tradeDate,
                settleDate: r.settleDate,
                quantity: r.quantity,
                price: r.price,
                amount: r.amount,
                fee: r.fee,
                currencyCode: r.currencyCode,
                notes: r.notes,
              })),
              skipDuplicates,
            },
          },
          success: true,
          importedCount: true,
          skippedCount: true,
          message: true,
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

      if (res.importTransactions && res.importTransactions.success) {
        onSuccess(
          res.importTransactions.portfolio,
          res.importTransactions.message || `Successfully imported ${res.importTransactions.importedCount} transactions!`
        );
        onClose();
      } else {
        setErrorBanner(res.importTransactions?.message || 'Failed to import transactions.');
      }
    } catch (err: any) {
      console.error('Import error:', err);
      setErrorBanner(err?.message || 'An error occurred during import.');
    } finally {
      setIsSubmitting(false);
    }
  };

  const validRows = parseResult?.validRows || [];
  const errors = parseResult?.errors || [];
  const readyRows = validRows.filter((r) => !duplicateRefs.has(r.externalRef));
  const dupRows = validRows.filter((r) => duplicateRefs.has(r.externalRef));

  const countToImport = skipDuplicates ? readyRows.length : validRows.length;

  const displayedRows: NormalizedTransactionRow[] =
    activeTab === 'all'
      ? validRows
      : activeTab === 'ready'
      ? readyRows
      : activeTab === 'duplicates'
      ? dupRows
      : [];

  return (
    <div className="modal-backdrop" onClick={onClose}>
      <div className="import-modal-container" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h2 className="modal-title">Import Broker CSV (CommSec / nabtrade)</h2>
          <button className="modal-close-btn" onClick={onClose} type="button">
            &times;
          </button>
        </div>

        <div className="import-modal-body">
          {errorBanner && <div className="error-banner">{errorBanner}</div>}

          {/* Config row: Broker selector */}
          <div className="upload-config-row">
            <div className="broker-select-group">
              <span className="broker-select-label">Broker Format:</span>
              <select
                className="broker-select-input"
                value={broker}
                onChange={(e) => handleBrokerChange(e.target.value as SupportedBroker)}
              >
                <option value="auto">Auto-detect (Recommended)</option>
                <option value="commsec">CommSec (Commonwealth Securities)</option>
                <option value="nabtrade">nabtrade (National Australia Bank)</option>
              </select>
            </div>

            {fileName && (
              <button type="button" className="btn-secondary" onClick={handleReset}>
                Change File
              </button>
            )}
          </div>

          {/* Dropzone */}
          {!fileName ? (
            <div
              className={`csv-dropzone ${isDragging ? 'dragging' : ''}`}
              onDragOver={handleDragOver}
              onDragLeave={handleDragLeave}
              onDrop={handleDrop}
              onClick={() => fileInputRef.current?.click()}
            >
              <input
                ref={fileInputRef}
                type="file"
                accept=".csv"
                style={{ display: 'none' }}
                onChange={handleFileChange}
              />
              <span className="dropzone-icon">📥</span>
              <span className="dropzone-title">
                {isParsing ? 'Parsing CSV...' : 'Click or drag & drop broker CSV file here'}
              </span>
              <span className="dropzone-subtitle">
                Supports CommSec transaction ledger exports and nabtrade trade confirmation CSVs
              </span>
            </div>
          ) : (
            <div className="dropzone-file-info">
              <span className="dropzone-icon" style={{ fontSize: '1.25rem' }}>📄</span>
              <span className="file-info-name">{fileName}</span>
              {parseResult && (
                <span className="file-info-broker-badge">{parseResult.broker}</span>
              )}
              {isCheckingDuplicates && (
                <span style={{ fontSize: '0.75rem', color: '#f59e0b' }}>
                  Checking duplicates...
                </span>
              )}
            </div>
          )}

          {/* Metric Summary Cards */}
          {parseResult && (
            <div className="metrics-summary-grid">
              <div className="metric-card-summary">
                <span className="metric-summary-label">Total Rows</span>
                <span className="metric-summary-value">{parseResult.totalRowsRead}</span>
              </div>
              <div className="metric-card-summary ready">
                <span className="metric-summary-label">Ready to Import</span>
                <span className="metric-summary-value">{readyRows.length}</span>
              </div>
              <div className="metric-card-summary duplicates">
                <span className="metric-summary-label">Duplicates</span>
                <span className="metric-summary-value">{dupRows.length}</span>
              </div>
              <div className="metric-card-summary errors">
                <span className="metric-summary-label">Malformed</span>
                <span className="metric-summary-value">{errors.length}</span>
              </div>
            </div>
          )}

          {/* Preview Tabs & Table */}
          {parseResult && (
            <>
              <div className="preview-tabs-row">
                <div className="preview-tabs">
                  <button
                    type="button"
                    className={`preview-tab-btn ${activeTab === 'all' ? 'active' : ''}`}
                    onClick={() => setActiveTab('all')}
                  >
                    All Valid ({validRows.length})
                  </button>
                  <button
                    type="button"
                    className={`preview-tab-btn ${activeTab === 'ready' ? 'active' : ''}`}
                    onClick={() => setActiveTab('ready')}
                  >
                    Ready ({readyRows.length})
                  </button>
                  <button
                    type="button"
                    className={`preview-tab-btn ${activeTab === 'duplicates' ? 'active' : ''}`}
                    onClick={() => setActiveTab('duplicates')}
                  >
                    Duplicates ({dupRows.length})
                  </button>
                  <button
                    type="button"
                    className={`preview-tab-btn ${activeTab === 'errors' ? 'active' : ''}`}
                    onClick={() => setActiveTab('errors')}
                  >
                    Errors ({errors.length})
                  </button>
                </div>
              </div>

              {activeTab !== 'errors' ? (
                <div className="preview-table-wrapper">
                  <table className="preview-table">
                    <thead>
                      <tr>
                        <th>Status</th>
                        <th>Row</th>
                        <th>Date</th>
                        <th>Type</th>
                        <th>Symbol</th>
                        <th>Units</th>
                        <th>Price</th>
                        <th>Fee</th>
                        <th>Amount</th>
                      </tr>
                    </thead>
                    <tbody>
                      {displayedRows.length === 0 ? (
                        <tr>
                          <td colSpan={9} style={{ textAlign: 'center', padding: '1.5rem', color: 'var(--text-secondary)' }}>
                            No rows in this filter.
                          </td>
                        </tr>
                      ) : (
                        displayedRows.map((row) => {
                          const isDup = duplicateRefs.has(row.externalRef);
                          return (
                            <tr key={row.externalRef + row.rowNumber}>
                              <td>
                                <span className={`status-pill ${isDup ? 'duplicate' : 'ready'}`}>
                                  {isDup ? 'Duplicate' : 'Ready'}
                                </span>
                              </td>
                              <td>#{row.rowNumber}</td>
                              <td>{row.tradeDate}</td>
                              <td>
                                <span className={`status-pill ${row.type.toLowerCase()}`}>
                                  {row.type}
                                </span>
                              </td>
                              <td style={{ fontWeight: 600 }}>{row.symbol || '—'}</td>
                              <td>{row.symbol ? row.quantity : '—'}</td>
                              <td>{row.symbol ? `$${row.price}` : '—'}</td>
                              <td>${row.fee}</td>
                              <td style={{ fontWeight: 600 }}>${row.amount}</td>
                            </tr>
                          );
                        })
                      )}
                    </tbody>
                  </table>
                </div>
              ) : (
                <div className="error-list-container">
                  {errors.length === 0 ? (
                    <div style={{ padding: '1rem', color: '#10b981', textAlign: 'center' }}>
                      No malformed or invalid rows detected!
                    </div>
                  ) : (
                    errors.map((err, idx) => (
                      <div key={idx} className="error-row-card">
                        <div className="error-row-header">
                          <span className="error-row-badge">Row #{err.rowNumber}</span>
                          <span className="error-row-msg">{err.message}</span>
                        </div>
                        <div className="error-row-raw">{err.rawValue}</div>
                      </div>
                    ))
                  )}
                </div>
              )}
            </>
          )}
        </div>

        {/* Modal Footer */}
        <div className="import-modal-footer">
          <label className="skip-duplicates-checkbox-group">
            <input
              type="checkbox"
              checked={skipDuplicates}
              onChange={(e) => setSkipDuplicates(e.target.checked)}
            />
            <span className="skip-duplicates-label">Skip duplicates (Recommended)</span>
          </label>

          <div className="footer-actions-group">
            <button type="button" className="btn-secondary" onClick={onClose}>
              Cancel
            </button>
            <button
              type="button"
              className="btn-import-submit"
              disabled={isSubmitting || !parseResult || countToImport === 0}
              onClick={handleImportSubmit}
            >
              {isSubmitting ? (
                <>
                  <span className="spinner-icon"></span>
                  Importing...
                </>
              ) : (
                `Import ${countToImport} Transaction${countToImport === 1 ? '' : 's'}`
              )}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

import React, { useState, useEffect } from 'react';
import { client } from '@graphfolio/api-client';
import { formatMoney, formatQuantity } from '@graphfolio/ui';
import { EditTransactionModal } from './EditTransactionModal';
import { formatTransactionsCSV, downloadCSV } from '../services/csv/export';
import './TransactionLedger.css';

interface Money {
  amount: string;
  currencyCode: string;
}

export interface TransactionItem {
  id: string;
  type: 'BUY' | 'SELL' | 'DIVIDEND' | 'DEPOSIT' | 'WITHDRAWAL' | 'INTEREST' | 'FEE' | 'TAX' | 'TRANSFER_IN' | 'TRANSFER_OUT' | 'FX_CONVERSION' | 'SPLIT';
  symbol?: string | null;
  instrumentName?: string | null;
  tradeDate: string;
  quantity?: string | null;
  price?: Money | null;
  amount: Money;
  fee: Money;
  notes?: string | null;
  createdAt: string;
  withholdingTax?: Money | null;
  fxRateToBase?: string | null;
  externalRef?: string | null;
}

interface TransactionLedgerProps {
  onTransactionDeleted?: (updatedPortfolio: any) => void;
  onTransactionUpdated?: (updatedPortfolio: any) => void;
  onOpenImportModal?: () => void;
  refreshTrigger?: number;
  preferredCurrency?: string;
  supportedCurrencies?: any[];
}

const filterTypes: Array<{ label: string; value: string | null }> = [
  { label: 'All', value: null },
  { label: 'Buy', value: 'BUY' },
  { label: 'Sell', value: 'SELL' },
  { label: 'Dividend', value: 'DIVIDEND' },
  { label: 'Split', value: 'SPLIT' },
  { label: 'Deposit', value: 'DEPOSIT' },
  { label: 'Withdrawal', value: 'WITHDRAWAL' },
];

export const TransactionLedger: React.FC<TransactionLedgerProps> = ({
  onTransactionDeleted,
  onTransactionUpdated,
  onOpenImportModal,
  refreshTrigger = 0,
  preferredCurrency,
  supportedCurrencies,
}) => {
  const [transactions, setTransactions] = useState<TransactionItem[]>([]);
  const [totalCount, setTotalCount] = useState(0);
  const [page, setPage] = useState(1);
  const pageSize = 20;

  const [selectedType, setSelectedType] = useState<string | null>(null);
  const [symbolQuery, setSymbolQuery] = useState('');
  const [reloadCount, setReloadCount] = useState(0);

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Export state
  const [isExporting, setIsExporting] = useState(false);
  const [exportNotice, setExportNotice] = useState<string | null>(null);

  // Edit modal state
  const [txToEdit, setTxToEdit] = useState<TransactionItem | null>(null);

  // Delete modal state
  const [txToDelete, setTxToDelete] = useState<TransactionItem | null>(null);
  const [deleting, setDeleting] = useState(false);
  const [deleteError, setDeleteError] = useState<string | null>(null);

  useEffect(() => {
    let isCancelled = false;
    const timer = setTimeout(() => {
      const args: any = {
        page,
        pageSize,
      };
      if (selectedType) {
        args.type = selectedType;
      }
      const trimmed = symbolQuery.trim();
      if (trimmed) {
        args.symbol = trimmed.toUpperCase();
      }

      client
        .query({
          transactions: {
            __args: args,
            totalCount: true,
            page: true,
            pageSize: true,
            items: {
              id: true,
              type: true,
              symbol: true,
              instrumentName: true,
              tradeDate: true,
              quantity: true,
              price: { amount: true, currencyCode: true },
              amount: { amount: true, currencyCode: true },
              fee: { amount: true, currencyCode: true },
              withholdingTax: { amount: true, currencyCode: true },
              fxRateToBase: true,
              externalRef: true,
              notes: true,
              createdAt: true,
            },
          },
        })
        .then((res: any) => {
          if (!isCancelled) {
            if (res.transactions) {
              setTransactions(res.transactions.items || []);
              setTotalCount(res.transactions.totalCount || 0);
            }
            setLoading(false);
          }
        })
        .catch((err: any) => {
          if (!isCancelled) {
            console.error('Failed to fetch transactions:', err);
            setError('Failed to load transaction history.');
            setLoading(false);
          }
        });
    }, 250);

    return () => {
      isCancelled = true;
      clearTimeout(timer);
    };
  }, [page, pageSize, selectedType, symbolQuery, refreshTrigger, reloadCount]);

  const handleTypeSelect = (typeVal: string | null) => {
    setSelectedType(typeVal);
    setLoading(true);
    setPage(1);
  };

  const handleSearchChange = (val: string) => {
    setSymbolQuery(val);
    setLoading(true);
    setPage(1);
  };

  const handleConfirmDelete = async () => {
    if (!txToDelete) return;
    setDeleting(true);
    setDeleteError(null);

    try {
      const res = await client.mutation({
        deleteTransaction: {
          __args: { id: txToDelete.id },
          success: true,
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

      if (res.deleteTransaction && res.deleteTransaction.success) {
        if (onTransactionDeleted) {
          onTransactionDeleted(res.deleteTransaction.portfolio);
        }
        setTxToDelete(null);
        // If current page is now empty and not on page 1, shift back 1 page
        if (transactions.length === 1 && page > 1) {
          setPage((p) => p - 1);
        } else {
          setReloadCount((c) => c + 1);
        }
      } else {
        setDeleteError('Failed to delete transaction. Please try again.');
      }
    } catch (err: any) {
      console.error('Error deleting transaction:', err);
      setDeleteError(err.message || 'An unexpected error occurred while deleting.');
    } finally {
      setDeleting(false);
    }
  };

  const handleExportCSV = async () => {
    if (isExporting || totalCount === 0) return;
    setIsExporting(true);
    setExportNotice(null);

    try {
      const args: any = {};
      if (selectedType) {
        args.type = selectedType;
      }
      const trimmed = symbolQuery.trim();
      if (trimmed) {
        args.symbol = trimmed.toUpperCase();
      }

      const res: any = await client.query({
        exportTransactions: {
          __args: args,
          id: true,
          type: true,
          symbol: true,
          instrumentName: true,
          tradeDate: true,
          quantity: true,
          price: { amount: true, currencyCode: true },
          amount: { amount: true, currencyCode: true },
          fee: { amount: true, currencyCode: true },
          withholdingTax: { amount: true, currencyCode: true },
          fxRateToBase: true,
          externalRef: true,
          notes: true,
          createdAt: true,
        },
      });

      const items: TransactionItem[] = res.exportTransactions || [];
      if (items.length === 0) {
        setExportNotice('No transactions found to export.');
        setTimeout(() => setExportNotice(null), 3000);
        return;
      }

      const csvContent = formatTransactionsCSV(items);
      const dateStr = new Date().toISOString().split('T')[0];
      let filename = `graphfolio-transactions-${dateStr}.csv`;
      if (trimmed) {
        filename = `graphfolio-transactions-${trimmed.toLowerCase()}-${dateStr}.csv`;
      } else if (selectedType) {
        filename = `graphfolio-transactions-${selectedType.toLowerCase()}-${dateStr}.csv`;
      }

      downloadCSV(csvContent, filename);
      setExportNotice(`Exported ${items.length} transaction${items.length === 1 ? '' : 's'} (oldest first).`);
      setTimeout(() => setExportNotice(null), 4000);
    } catch (err: any) {
      console.error('Failed to export transactions:', err);
      setError('Failed to export transactions. Please try again.');
    } finally {
      setIsExporting(false);
    }
  };

  const totalPages = Math.max(1, Math.ceil(totalCount / pageSize));
  const startItem = totalCount === 0 ? 0 : (page - 1) * pageSize + 1;
  const endItem = Math.min(page * pageSize, totalCount);

  const getBadgeClass = (type: string) => {
    switch (type) {
      case 'BUY':
        return 'badge-type buy';
      case 'SELL':
        return 'badge-type sell';
      case 'DIVIDEND':
        return 'badge-type dividend';
      case 'SPLIT':
        return 'badge-type split';
      case 'DEPOSIT':
        return 'badge-type deposit';
      case 'WITHDRAWAL':
        return 'badge-type withdrawal';
      default:
        return 'badge-type other';
    }
  };

  return (
    <div className="transaction-ledger-container">
      <div className="ledger-card">
        <div className="ledger-header">
          <div className="ledger-title-group">
            <h3>Transaction Ledger</h3>
            <p>Authoritative record of all cash flows and trades with automatic projection replay</p>
          </div>
          <div className="ledger-actions-group">
            <button
              type="button"
              className="btn-export-csv"
              onClick={handleExportCSV}
              disabled={isExporting || totalCount === 0}
              title="Download transactions sorted chronologically (oldest first)"
            >
              {isExporting ? (
                <>
                  <span className="spinner-sm" /> Exporting...
                </>
              ) : (
                <>
                  <span className="btn-icon">📤</span> Export CSV
                </>
              )}
            </button>
            {onOpenImportModal && (
              <button
                type="button"
                className="btn-import-csv"
                onClick={onOpenImportModal}
                title="Import trades from CommSec or nabtrade CSV"
              >
                📥 Import CSV
              </button>
            )}
          </div>
        </div>

        {exportNotice && (
          <div className="ledger-toast-notice">
            <span>✓ {exportNotice}</span>
          </div>
        )}

        {/* Toolbar: Filters and Search */}
        <div className="ledger-toolbar">
          <div className="type-filter-pills">
            {filterTypes.map((ft) => (
              <button
                key={ft.label}
                type="button"
                className={`filter-pill ${selectedType === ft.value ? 'active' : ''}`}
                onClick={() => handleTypeSelect(ft.value)}
              >
                {ft.label}
              </button>
            ))}
          </div>

          <div className="filter-search-box">
            <span className="search-icon">🔍</span>
            <input
              type="text"
              className="filter-search-input"
              placeholder="Search ticker (e.g. AAPL)..."
              value={symbolQuery}
              onChange={(e) => handleSearchChange(e.target.value)}
            />
            {symbolQuery && (
              <button
                type="button"
                className="filter-search-clear"
                onClick={() => handleSearchChange('')}
                title="Clear filter"
              >
                ✕
              </button>
            )}
          </div>
        </div>

        {/* Content Table */}
        {loading ? (
          <div className="ledger-state-msg">Loading transaction ledger...</div>
        ) : error ? (
          <div className="ledger-state-msg error">{error}</div>
        ) : transactions.length === 0 ? (
          <div className="ledger-state-msg">No transactions found matching your criteria.</div>
        ) : (
          <div className="ledger-table-wrapper">
            <table className="ledger-table">
              <thead>
                <tr>
                  <th>Date</th>
                  <th>Type</th>
                  <th>Asset</th>
                  <th>Quantity</th>
                  <th>Price</th>
                  <th>Amount</th>
                  <th>Fee</th>
                  <th>Notes</th>
                  <th style={{ width: '80px', textAlign: 'center' }}></th>
                </tr>
              </thead>
              <tbody>
                {transactions.map((tx) => {
                  const hasFee = tx.fee && Number(tx.fee.amount) > 0;
                  return (
                    <tr key={tx.id}>
                      <td>{tx.tradeDate}</td>
                      <td>
                        <span className={getBadgeClass(tx.type)}>{tx.type}</span>
                      </td>
                      <td>
                        <div className="cell-asset">
                          {tx.symbol ? (
                            <>
                              <span className="asset-symbol">{tx.symbol}</span>
                              {tx.instrumentName && (
                                <span className="asset-name">{tx.instrumentName}</span>
                              )}
                            </>
                          ) : (
                            <span className="asset-cash">Cash</span>
                          )}
                        </div>
                      </td>
                      <td>{tx.type === 'SPLIT' && tx.quantity ? `${formatQuantity(tx.quantity)}x` : formatQuantity(tx.quantity)}</td>
                      <td>{tx.price && tx.type !== 'SPLIT' ? formatMoney(tx.price) : '-'}</td>
                      <td className="cell-amount">{tx.type === 'SPLIT' ? '-' : formatMoney(tx.amount)}</td>
                      <td>{hasFee ? formatMoney(tx.fee) : '-'}</td>
                      <td>
                        <span className="cell-notes" title={tx.notes || ''}>
                          {tx.notes || '-'}
                        </span>
                      </td>
                      <td style={{ textAlign: 'center' }}>
                        <div className="row-actions">
                          <button
                            type="button"
                            className="btn-edit-row"
                            onClick={() => setTxToEdit(tx)}
                            title="Edit transaction"
                          >
                            <svg
                              width="14"
                              height="14"
                              viewBox="0 0 24 24"
                              fill="none"
                              stroke="currentColor"
                              strokeWidth="2"
                              strokeLinecap="round"
                              strokeLinejoin="round"
                            >
                              <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" />
                              <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
                            </svg>
                          </button>
                          <button
                            type="button"
                            className="btn-delete-row"
                            onClick={() => setTxToDelete(tx)}
                            title="Delete transaction"
                          >
                            <svg
                              width="14"
                              height="14"
                              viewBox="0 0 24 24"
                              fill="none"
                              stroke="currentColor"
                              strokeWidth="2"
                              strokeLinecap="round"
                              strokeLinejoin="round"
                            >
                              <path d="M3 6h18" />
                              <path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6" />
                              <path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2" />
                              <line x1="10" y1="11" x2="10" y2="17" />
                              <line x1="14" y1="11" x2="14" y2="17" />
                            </svg>
                          </button>
                        </div>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}

        {/* Pagination Bar */}
        {totalCount > 0 && (
          <div className="ledger-pagination">
            <span className="pagination-info">
              Showing {startItem}–{endItem} of {totalCount} transactions
            </span>
            <div className="pagination-controls">
              <button
                type="button"
                className="btn-pagination"
                disabled={page <= 1 || loading}
                onClick={() => setPage((p) => Math.max(1, p - 1))}
              >
                ← Prev
              </button>
              <span className="pagination-page-indicator">
                {page} / {totalPages}
              </span>
              <button
                type="button"
                className="btn-pagination"
                disabled={page >= totalPages || loading}
                onClick={() => setPage((p) => p + 1)}
              >
                Next →
              </button>
            </div>
          </div>
        )}
      </div>

      {/* Delete Confirmation Modal */}
      {txToDelete && (
        <div className="modal-overlay" onClick={() => !deleting && setTxToDelete(null)}>
          <div className="delete-modal-content" onClick={(e) => e.stopPropagation()}>
            <div className="delete-modal-header">
              <div className="delete-modal-icon">⚠️</div>
              <h4>Delete Transaction</h4>
            </div>

            <div className="delete-warning-banner">
              <strong>Warning:</strong> Deleting this transaction will permanently remove it from your
              ledger and automatically re-calculate your portfolio holdings, tax lots, and cost basis.
            </div>

            <div className="delete-tx-summary">
              <div className="tx-summary-row">
                <span>Type:</span>
                <span className={getBadgeClass(txToDelete.type)}>{txToDelete.type}</span>
              </div>
              <div className="tx-summary-row">
                <span>Date:</span>
                <span>{txToDelete.tradeDate}</span>
              </div>
              {txToDelete.symbol && (
                <div className="tx-summary-row">
                  <span>Asset:</span>
                  <span>{txToDelete.symbol}</span>
                </div>
              )}
              {txToDelete.quantity && (
                <div className="tx-summary-row">
                  <span>Quantity:</span>
                  <span>{txToDelete.quantity}</span>
                </div>
              )}
              <div className="tx-summary-row">
                <span>Amount:</span>
                <span>{formatMoney(txToDelete.amount)}</span>
              </div>
            </div>

            {deleteError && <div className="delete-error-msg">{deleteError}</div>}

            <div className="delete-modal-actions">
              <button
                type="button"
                className="btn-cancel-delete"
                disabled={deleting}
                onClick={() => setTxToDelete(null)}
              >
                Cancel
              </button>
              <button
                type="button"
                className="btn-confirm-delete"
                disabled={deleting}
                onClick={handleConfirmDelete}
              >
                {deleting ? 'Deleting...' : 'Delete Transaction'}
              </button>
            </div>
          </div>
        </div>
      )}
      {/* Edit Transaction Modal */}
      {txToEdit && (
        <EditTransactionModal
          isOpen={true}
          transaction={txToEdit}
          onClose={() => setTxToEdit(null)}
          preferredCurrency={preferredCurrency}
          supportedCurrencies={supportedCurrencies}
          onSuccess={(updatedPortfolio) => {
            setTxToEdit(null);
            setReloadCount((c) => c + 1);
            if (onTransactionUpdated) {
              onTransactionUpdated(updatedPortfolio);
            }
          }}
        />
      )}
    </div>
  );
};

import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { client, type TransactionType } from '@graphfolio/api-client';
import {
  Button,
  Badge,
  Card,
  Table,
  TableHead,
  TableBody,
  TableRow,
  TableHeaderCell,
  TableCell,
  Input,
  Select,
  formatMoney,
  formatDate,
} from '@graphfolio/ui';
import {
  AddAssetPriceModal,
  type AddAssetPriceSubmitData,
  type InstrumentOption,
} from './AddAssetPriceModal';
import { BackfillModal, type BackfillSubmitData, type BackfillModalResult } from './BackfillModal';
import './MissingPricesReport.css';

export interface MissingPriceItem {
  transactionId: string;
  portfolioId: string;
  instrumentId: string;
  symbol: string;
  assetName: string;
  exchangeMic: string;
  tradeDate: string;
  type: TransactionType;
  quantity: string;
  executionPrice: { amount: string; currencyCode: string };
  currencyCode: string;
}

export interface MissingPricesReportProps {
  onNotify: (msg: string) => void;
  onUpdateCount?: (count: number) => void;
}

type PresetKey = '30D' | '90D' | 'YTD' | '1Y' | 'ALL';

const PAGE_SIZE = 50;

function formatISODate(d: Date): string {
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, '0');
  const day = String(d.getDate()).padStart(2, '0');
  return `${y}-${m}-${day}`;
}

export function calculatePresetDates(preset: PresetKey): { from: string; to: string } {
  const now = new Date();
  const todayISO = formatISODate(now);
  switch (preset) {
    case '30D': {
      const d = new Date(now);
      d.setDate(d.getDate() - 30);
      return { from: formatISODate(d), to: todayISO };
    }
    case '90D': {
      const d = new Date(now);
      d.setDate(d.getDate() - 90);
      return { from: formatISODate(d), to: todayISO };
    }
    case 'YTD': {
      const d = new Date(now.getFullYear(), 0, 1);
      return { from: formatISODate(d), to: todayISO };
    }
    case '1Y': {
      const d = new Date(now);
      d.setDate(d.getDate() - 365);
      return { from: formatISODate(d), to: todayISO };
    }
    case 'ALL': {
      return { from: '', to: '' };
    }
  }
}

export function calculateDateSpanDays(fromStr?: string | null, toStr?: string | null): string {
  if (!fromStr || !toStr) return '—';
  const d1 = new Date(fromStr);
  const d2 = new Date(toStr);
  if (isNaN(d1.getTime()) || isNaN(d2.getTime())) return '—';
  const diffTime = Math.abs(d2.getTime() - d1.getTime());
  const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24)) + 1;
  return `${diffDays} Day${diffDays === 1 ? '' : 's'}`;
}

export function getTypeBadgeVariant(type: TransactionType): 'success' | 'info' | 'warning' | 'purple' | 'neutral' {
  switch (type) {
    case 'BUY':
      return 'success';
    case 'SELL':
      return 'info';
    case 'DIVIDEND':
      return 'warning';
    case 'SPLIT':
      return 'purple';
    default:
      return 'neutral';
  }
}

export const MissingPricesReport: React.FC<MissingPricesReportProps> = ({ onNotify, onUpdateCount }) => {
  const [items, setItems] = useState<MissingPriceItem[]>([]);
  const [totalCount, setTotalCount] = useState(0);
  const [uniqueInstrumentsCount, setUniqueInstrumentsCount] = useState(0);
  const [earliestTradeDate, setEarliestTradeDate] = useState<string | null>(null);
  const [latestTradeDate, setLatestTradeDate] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  // Filters
  const [symbolFilter, setSymbolFilter] = useState('ALL');
  const [typeFilter, setTypeFilter] = useState('ALL');
  const [fromDate, setFromDate] = useState('');
  const [toDate, setToDate] = useState('');
  const [activePreset, setActivePreset] = useState<string | null>(null);
  const [page, setPage] = useState(1);

  // Available metadata
  const [availableSymbols, setAvailableSymbols] = useState<string[]>([]);
  const [availableInstruments, setAvailableInstruments] = useState<InstrumentOption[]>([]);

  // Modal states
  const [isBackfillModalOpen, setIsBackfillModalOpen] = useState(false);
  const [backfillModalProps, setBackfillModalProps] = useState<{
    symbols?: string[];
    fromDate?: string;
    toDate?: string;
  }>({});

  const [isAddPriceModalOpen, setIsAddPriceModalOpen] = useState(false);
  const [targetPriceModalData, setTargetPriceModalData] = useState<{
    symbol: string;
    priceDate?: string;
    price?: string;
  }>({ symbol: 'AAPL' });

  const totalPages = Math.max(1, Math.ceil(totalCount / PAGE_SIZE));
  const isDateRangeInvalid = Boolean(fromDate && toDate && fromDate > toDate);

  // Fetch instruments for dropdowns
  const fetchInstruments = useCallback(() => {
    client
      .query({
        allInstruments: {
          symbol: true,
          name: true,
          currencyCode: true,
          isActive: true,
        },
      })
      .then((res) => {
        if (res.allInstruments && res.allInstruments.length > 0) {
          const instOptions: InstrumentOption[] = res.allInstruments.map((i) => ({
            symbol: i.symbol,
            name: i.name,
            currencyCode: i.currencyCode,
          }));
          setAvailableInstruments(instOptions);

          const syms = Array.from(new Set(res.allInstruments.map((i) => i.symbol))).sort();
          setAvailableSymbols(syms);
        }
      })
      .catch((err) => console.warn('Could not fetch instrument symbols:', err));
  }, []);

  // Fetch missing price transactions from GraphQL BFF
  const fetchReport = useCallback(() => {
    if (fromDate && toDate && fromDate > toDate) {
      return;
    }

    setLoading(true);
    client
      .query({
        missingPriceTransactions: {
          __args: {
            symbol: symbolFilter === 'ALL' ? undefined : symbolFilter,
            fromDate: fromDate || undefined,
            toDate: toDate || undefined,
            limit: PAGE_SIZE,
            offset: (page - 1) * PAGE_SIZE,
          },
          items: {
            transactionId: true,
            portfolioId: true,
            instrumentId: true,
            symbol: true,
            assetName: true,
            exchangeMic: true,
            tradeDate: true,
            type: true,
            quantity: true,
            executionPrice: {
              amount: true,
              currencyCode: true,
            },
            currencyCode: true,
          },
          totalCount: true,
          uniqueInstrumentsCount: true,
          earliestTradeDate: true,
          latestTradeDate: true,
        },
      })
      .then((res) => {
        const report = res.missingPriceTransactions;
        if (report) {
          setItems(report.items as MissingPriceItem[]);
          setTotalCount(report.totalCount);
          setUniqueInstrumentsCount(report.uniqueInstrumentsCount);
          setEarliestTradeDate(report.earliestTradeDate ?? null);
          setLatestTradeDate(report.latestTradeDate ?? null);
          onUpdateCount?.(report.totalCount);
        }
      })
      .catch((err) => {
        console.error('Failed to fetch missing prices report:', err);
        onNotify(`Fetch report failed: ${err?.message || 'Server error'}`);
      })
      .finally(() => {
        setLoading(false);
      });
  }, [symbolFilter, fromDate, toDate, page, onNotify, onUpdateCount]);

  useEffect(() => {
    fetchInstruments();
  }, [fetchInstruments]);

  useEffect(() => {
    fetchReport();
  }, [fetchReport]);

  const handlePresetClick = (preset: PresetKey) => {
    setActivePreset(preset);
    const { from, to } = calculatePresetDates(preset);
    setFromDate(from);
    setToDate(to);
    setPage(1);
  };

  const handleCustomDateChange = (type: 'from' | 'to', value: string) => {
    setActivePreset(null);
    if (type === 'from') {
      setFromDate(value);
    } else {
      setToDate(value);
    }
    setPage(1);
  };

  const handleSymbolFilterChange = (val: string) => {
    setSymbolFilter(val);
    setPage(1);
  };

  // Launch Backfill Modal for a single transaction row
  const handleBackfillRow = (item: MissingPriceItem) => {
    setBackfillModalProps({
      symbols: [item.symbol],
      fromDate: item.tradeDate,
      toDate: item.tradeDate,
    });
    setIsBackfillModalOpen(true);
  };

  // Launch Backfill Modal for all detected missing prices
  const handleBackfillAll = () => {
    const todayISO = formatISODate(new Date());
    setBackfillModalProps({
      symbols: symbolFilter === 'ALL' ? undefined : [symbolFilter],
      fromDate: earliestTradeDate || fromDate || todayISO,
      toDate: latestTradeDate || toDate || todayISO,
    });
    setIsBackfillModalOpen(true);
  };

  // Launch Manual Add Price modal for a single transaction row
  const handleAddManualPriceRow = (item: MissingPriceItem) => {
    setTargetPriceModalData({
      symbol: item.symbol,
      priceDate: item.tradeDate,
      price: item.executionPrice.amount,
    });
    setIsAddPriceModalOpen(true);
  };

  // Backfill execute submission handler
  const handleExecuteBackfill = async (submitData: BackfillSubmitData): Promise<BackfillModalResult> => {
    try {
      const res = await client.mutation({
        triggerBackfill: {
          __args: {
            input: {
              fromDate: submitData.fromDate,
              toDate: submitData.toDate,
              symbols: submitData.symbols,
              currencyPairs: submitData.currencyPairs,
              backfillAssets: submitData.backfillAssets,
              backfillFx: submitData.backfillFx,
              recomputeValuations: submitData.recomputeValuations,
            },
          },
          success: true,
          pricesSynced: true,
          fxRatesSynced: true,
          message: true,
          warnings: true,
        },
      });

      const payload = res.triggerBackfill;
      if (payload.success) {
        onNotify(payload.message || 'Historical prices backfilled successfully.');
        fetchReport();
      } else {
        onNotify(`Backfill warning: ${payload.message}`);
      }

      return {
        success: payload.success,
        pricesSynced: payload.pricesSynced,
        fxRatesSynced: payload.fxRatesSynced,
        message: payload.message,
        warnings: payload.warnings || [],
      };
    } catch (err: any) {
      console.error('Trigger backfill error:', err);
      onNotify(`Backfill failed: ${err?.message || 'Server error'}`);
      throw err;
    }
  };

  // Manual price override submission handler
  const handleSavePriceOverride = async (submitData: AddAssetPriceSubmitData): Promise<void> => {
    try {
      const res = await client.mutation({
        recordPriceOverride: {
          __args: {
            input: {
              symbol: submitData.symbol,
              priceDate: submitData.priceDate,
              price: submitData.price,
              reason: submitData.reason,
              recomputeValuations: submitData.recomputeValuations,
            },
          },
          price: {
            price: {
              amount: true,
              currencyCode: true,
            },
            priceDate: true,
            symbol: true,
          },
          valuationsRecomputed: true,
        },
      });

      onNotify(`Saved price for ${res.recordPriceOverride.price.symbol} on ${res.recordPriceOverride.price.priceDate}.`);
      fetchReport();
    } catch (err: any) {
      console.error('Record asset price error:', err);
      onNotify(`Record asset price failed: ${err?.message || 'Server error'}`);
      throw err;
    }
  };

  // Client-side transaction type filter if specified
  const filteredItems = useMemo(() => {
    if (typeFilter === 'ALL') return items;
    return items.filter((item) => item.type === typeFilter);
  }, [items, typeFilter]);

  const dateSpanText = useMemo(() => {
    return calculateDateSpanDays(earliestTradeDate, latestTradeDate);
  }, [earliestTradeDate, latestTradeDate]);

  return (
    <div className="missing-prices-report">
      {/* 1. Summary KPI Header */}
      <div className="missing-prices-report__stats">
        <Card className="missing-price-stat-card">
          <div className="stat-label">Unpriced Trades</div>
          <div className={`stat-value ${totalCount > 0 ? 'missing-price-stat-card .stat-value--alert' : ''}`}>
            {loading ? '…' : totalCount}
          </div>
        </Card>

        <Card className="missing-price-stat-card">
          <div className="stat-label">Affected Assets</div>
          <div className="stat-value">{loading ? '…' : uniqueInstrumentsCount}</div>
        </Card>

        <Card className="missing-price-stat-card">
          <div className="stat-label">Earliest Missing Date</div>
          <div className="stat-value">
            {loading ? '…' : earliestTradeDate ? formatDate(earliestTradeDate) : '—'}
          </div>
        </Card>

        <Card className="missing-price-stat-card">
          <div className="stat-label">Date Span Coverage</div>
          <div className="stat-value">{loading ? '…' : dateSpanText}</div>
        </Card>
      </div>

      {/* 2. Filter & Action Toolbar */}
      <Card>
        <div className="missing-prices-report__toolbar">
          <div className="missing-prices-report__filters">
            {/* Symbol Filter */}
            <div className="missing-prices-report__symbol-filter">
              <Select
                value={symbolFilter}
                onChange={(e) => handleSymbolFilterChange(e.target.value)}
                options={[
                  { value: 'ALL', label: 'All Symbols' },
                  ...availableSymbols.map((s) => ({ value: s, label: s })),
                ]}
              />
            </div>

            {/* Transaction Type Filter */}
            <div className="missing-prices-report__type-filter">
              <Select
                value={typeFilter}
                onChange={(e) => setTypeFilter(e.target.value)}
                options={[
                  { value: 'ALL', label: 'All Types' },
                  { value: 'BUY', label: 'BUY' },
                  { value: 'SELL', label: 'SELL' },
                  { value: 'SPLIT', label: 'SPLIT' },
                  { value: 'TRANSFER_IN', label: 'TRANSFER_IN' },
                  { value: 'TRANSFER_OUT', label: 'TRANSFER_OUT' },
                ]}
              />
            </div>

            {/* Date Inputs */}
            <div className="missing-prices-report__date-input-group">
              <Input
                type="date"
                value={fromDate}
                onChange={(e) => handleCustomDateChange('from', e.target.value)}
                className="missing-prices-report__date-input"
                title="Filter from trade date"
              />
            </div>

            <span style={{ color: 'var(--text-muted, #71717a)' }}>→</span>

            <div className="missing-prices-report__date-input-group">
              <Input
                type="date"
                value={toDate}
                onChange={(e) => handleCustomDateChange('to', e.target.value)}
                className="missing-prices-report__date-input"
                title="Filter to trade date"
              />
            </div>

            {/* Presets */}
            <div className="missing-prices-report__presets">
              {(['30D', '90D', 'YTD', '1Y', 'ALL'] as PresetKey[]).map((preset) => (
                <Button
                  key={preset}
                  variant={activePreset === preset ? 'primary' : 'secondary'}
                  size="sm"
                  onClick={() => handlePresetClick(preset)}
                >
                  {preset}
                </Button>
              ))}
            </div>

            {isDateRangeInvalid && (
              <span className="missing-prices-report__date-error">
                ⚠️ From Date must be earlier than or equal to To Date
              </span>
            )}
          </div>

          {/* Action Buttons */}
          <div className="missing-prices-report__actions">
            <Button
              variant="secondary"
              size="sm"
              onClick={fetchReport}
              disabled={loading || isDateRangeInvalid}
            >
              🔄 Refresh
            </Button>

            <Button
              variant="primary"
              size="sm"
              onClick={handleBackfillAll}
              disabled={totalCount === 0 || loading}
            >
              ⚡ Backfill All Missing
            </Button>
          </div>
        </div>
      </Card>

      {/* 3. Diagnostic Ledger Table */}
      <Card>
        {loading && items.length === 0 ? (
          <div style={{ padding: '3rem', textAlign: 'center', color: 'var(--text-secondary, #a1a1aa)' }}>
            Loading diagnostic report…
          </div>
        ) : totalCount === 0 ? (
          <div className="missing-prices-report__empty">
            <div className="missing-prices-report__empty-icon">🛡️</div>
            <div className="missing-prices-report__empty-title">All Transactions Fully Priced</div>
            <div className="missing-prices-report__empty-text">
              Every recorded trade event has verified daily closing marks on its trade date. Historical valuation
              replays and cost basis engines are fully reconciled.
            </div>
            <Button variant="secondary" size="sm" onClick={fetchReport}>
              Re-verify Missing Prices
            </Button>
          </div>
        ) : (
          <>
            <Table>
              <TableHead>
                <TableRow>
                  <TableHeaderCell>Trade Date</TableHeaderCell>
                  <TableHeaderCell>Symbol & Asset</TableHeaderCell>
                  <TableHeaderCell>Type</TableHeaderCell>
                  <TableHeaderCell style={{ textAlign: 'right' }}>Quantity</TableHeaderCell>
                  <TableHeaderCell style={{ textAlign: 'right' }}>Execution Price</TableHeaderCell>
                  <TableHeaderCell>Status</TableHeaderCell>
                  <TableHeaderCell style={{ textAlign: 'right' }}>Actions</TableHeaderCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {filteredItems.map((item) => (
                  <TableRow key={item.transactionId}>
                    <TableCell style={{ fontVariantNumeric: 'tabular-nums', fontWeight: 500 }}>
                      {formatDate(item.tradeDate)}
                    </TableCell>

                    <TableCell>
                      <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                        <span style={{ fontWeight: 600, color: '#ffffff' }}>{item.symbol}</span>
                        {item.exchangeMic && <Badge variant="neutral">{item.exchangeMic}</Badge>}
                      </div>
                      <div style={{ fontSize: '0.8rem', color: 'var(--text-secondary, #a1a1aa)' }}>
                        {item.assetName}
                      </div>
                    </TableCell>

                    <TableCell>
                      <Badge variant={getTypeBadgeVariant(item.type)}>{item.type}</Badge>
                    </TableCell>

                    <TableCell style={{ textAlign: 'right', fontVariantNumeric: 'tabular-nums' }}>
                      {item.quantity}
                    </TableCell>

                    <TableCell style={{ textAlign: 'right', fontVariantNumeric: 'tabular-nums' }}>
                      {formatMoney(item.executionPrice.amount, item.executionPrice.currencyCode)}
                    </TableCell>

                    <TableCell>
                      <Badge variant="warning">Missing Mark</Badge>
                    </TableCell>

                    <TableCell style={{ textAlign: 'right' }}>
                      <div className="missing-prices-report__actions-cell" style={{ justifyContent: 'flex-end' }}>
                        <Button
                          variant="secondary"
                          size="sm"
                          onClick={() => handleBackfillRow(item)}
                          title={`Backfill ${item.symbol} closing price on ${item.tradeDate}`}
                        >
                          ⚡ Backfill Range
                        </Button>

                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleAddManualPriceRow(item)}
                          title={`Manually record price mark for ${item.symbol}`}
                        >
                          ✏️ Add Price
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>

            {/* 4. Pagination Bar */}
            <div className="missing-prices-report__pagination">
              <div className="missing-prices-report__pagination-info">
                Showing {Math.min((page - 1) * PAGE_SIZE + 1, totalCount)} to{' '}
                {Math.min(page * PAGE_SIZE, totalCount)} of {totalCount} unpriced transactions
              </div>

              <div className="missing-prices-report__pagination-controls">
                <Button
                  variant="secondary"
                  size="sm"
                  onClick={() => setPage((p) => Math.max(1, p - 1))}
                  disabled={page <= 1 || loading}
                >
                  Previous
                </Button>

                <span className="missing-prices-report__page-indicator">
                  Page {page} of {totalPages}
                </span>

                <Button
                  variant="secondary"
                  size="sm"
                  onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                  disabled={page >= totalPages || loading}
                >
                  Next
                </Button>
              </div>
            </div>
          </>
        )}
      </Card>

      {/* 5. Historical Backfill Modal */}
      <BackfillModal
        isOpen={isBackfillModalOpen}
        onClose={() => setIsBackfillModalOpen(false)}
        initialSymbols={backfillModalProps.symbols}
        initialFromDate={backfillModalProps.fromDate}
        initialToDate={backfillModalProps.toDate}
        initialBackfillAssets={true}
        initialBackfillFx={false}
        initialRecomputeValuations={true}
        onSubmit={handleExecuteBackfill}
      />

      {/* 6. Manual Add Asset Price Modal */}
      <AddAssetPriceModal
        isOpen={isAddPriceModalOpen}
        onClose={() => setIsAddPriceModalOpen(false)}
        instruments={availableInstruments}
        defaultSymbol={targetPriceModalData.symbol}
        defaultDate={targetPriceModalData.priceDate}
        defaultPrice={targetPriceModalData.price}
        onSubmit={handleSavePriceOverride}
      />
    </div>
  );
};

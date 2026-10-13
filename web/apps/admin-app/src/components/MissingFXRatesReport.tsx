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
import { BackfillModal, type BackfillSubmitData, type BackfillModalResult } from './BackfillModal';
import { FXOverrideModal, type FXOverrideData } from './FXOverrideModal';
import './MissingFXRatesReport.css';

export interface MissingFXRateItem {
  transactionId: string;
  portfolioId: string;
  portfolioName: string;
  portfolioBaseCurrency: string;
  tradeDate: string;
  type: TransactionType;
  instrumentId?: string | null;
  symbol?: string | null;
  assetName?: string | null;
  amount: { amount: string; currencyCode: string };
  transactionCurrency: string;
  requiredPair: string;
  recordedFxRate?: string | null;
}

export interface MissingFXRatesReportProps {
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

export function calculateBufferDate(dateStr: string, bufferDays = 3): string {
  const d = new Date(dateStr);
  if (isNaN(d.getTime())) return dateStr;
  d.setDate(d.getDate() - bufferDays);
  return formatISODate(d);
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

export const MissingFXRatesReport: React.FC<MissingFXRatesReportProps> = ({ onNotify, onUpdateCount }) => {
  const [items, setItems] = useState<MissingFXRateItem[]>([]);
  const [totalCount, setTotalCount] = useState(0);
  const [uniquePairsCount, setUniquePairsCount] = useState(0);
  const [earliestTradeDate, setEarliestTradeDate] = useState<string | null>(null);
  const [latestTradeDate, setLatestTradeDate] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  // Filters
  const [pairFilter, setPairFilter] = useState('ALL');
  const [typeFilter, setTypeFilter] = useState('ALL');
  const [fromDate, setFromDate] = useState('');
  const [toDate, setToDate] = useState('');
  const [activePreset, setActivePreset] = useState<string | null>(null);
  const [page, setPage] = useState(1);

  // Available pairs detected
  const [availablePairs, setAvailablePairs] = useState<string[]>([]);

  // Modal states
  const [isBackfillModalOpen, setIsBackfillModalOpen] = useState(false);
  const [backfillModalProps, setBackfillModalProps] = useState<{
    currencyPairs?: string[];
    fromDate?: string;
    toDate?: string;
  }>({});

  const [isOverrideModalOpen, setIsOverrideModalOpen] = useState(false);
  const [targetOverrideData, setTargetOverrideData] = useState<{
    base: string;
    quote: string;
    date: string;
    rate?: string;
  }>({
    base: 'USD',
    quote: 'AUD',
    date: formatISODate(new Date()),
  });

  const totalPages = Math.max(1, Math.ceil(totalCount / PAGE_SIZE));
  const isDateRangeInvalid = Boolean(fromDate && toDate && fromDate > toDate);

  // Fetch missing FX rate transactions from GraphQL BFF
  const fetchReport = useCallback(() => {
    if (fromDate && toDate && fromDate > toDate) {
      return;
    }

    setLoading(true);
    client
      .query({
        missingFXRateTransactions: {
          __args: {
            currencyPair: pairFilter === 'ALL' ? undefined : pairFilter,
            fromDate: fromDate || undefined,
            toDate: toDate || undefined,
            type: typeFilter === 'ALL' ? undefined : (typeFilter as TransactionType),
            limit: PAGE_SIZE,
            offset: (page - 1) * PAGE_SIZE,
          },
          items: {
            transactionId: true,
            portfolioId: true,
            portfolioName: true,
            portfolioBaseCurrency: true,
            tradeDate: true,
            type: true,
            instrumentId: true,
            symbol: true,
            assetName: true,
            amount: {
              amount: true,
              currencyCode: true,
            },
            transactionCurrency: true,
            requiredPair: true,
            recordedFxRate: true,
          },
          totalCount: true,
          uniquePairsCount: true,
          earliestTradeDate: true,
          latestTradeDate: true,
        },
      })
      .then((res) => {
        const report = res.missingFXRateTransactions;
        if (report) {
          const mappedItems: MissingFXRateItem[] = (report.items || []).map((it) => ({
            transactionId: it.transactionId,
            portfolioId: it.portfolioId,
            portfolioName: it.portfolioName,
            portfolioBaseCurrency: it.portfolioBaseCurrency,
            tradeDate: it.tradeDate,
            type: it.type as TransactionType,
            instrumentId: it.instrumentId,
            symbol: it.symbol,
            assetName: it.assetName,
            amount: {
              amount: it.amount?.amount ?? '0',
              currencyCode: it.amount?.currencyCode ?? it.transactionCurrency,
            },
            transactionCurrency: it.transactionCurrency,
            requiredPair: it.requiredPair,
            recordedFxRate: it.recordedFxRate,
          }));

          setItems(mappedItems);
          setTotalCount(report.totalCount || 0);
          setUniquePairsCount(report.uniquePairsCount || 0);
          setEarliestTradeDate(report.earliestTradeDate ?? null);
          setLatestTradeDate(report.latestTradeDate ?? null);

          // Track unique currency pairs for filter selector
          const discoveredPairs = Array.from(new Set(mappedItems.map((i) => i.requiredPair))).filter(Boolean).sort();
          if (discoveredPairs.length > 0) {
            setAvailablePairs((prev) => Array.from(new Set([...prev, ...discoveredPairs])).sort());
          }

          if (onUpdateCount) {
            onUpdateCount(report.totalCount || 0);
          }
        }
      })
      .catch((err) => {
        console.error('Failed to fetch missing FX rates report:', err);
        onNotify('Error loading Missing FX Rates Report');
      })
      .finally(() => {
        setLoading(false);
      });
  }, [pairFilter, typeFilter, fromDate, toDate, page, onNotify, onUpdateCount]);

  useEffect(() => {
    fetchReport();
  }, [fetchReport]);

  // Handlers for Preset clicks
  const handlePresetSelect = (preset: PresetKey) => {
    setActivePreset(preset);
    const { from, to } = calculatePresetDates(preset);
    setFromDate(from);
    setToDate(to);
    setPage(1);
  };

  // Trigger single row BackfillModal
  const handleRowBackfill = (item: MissingFXRateItem) => {
    const bufferedFromDate = calculateBufferDate(item.tradeDate, 3);
    setBackfillModalProps({
      currencyPairs: [item.requiredPair],
      fromDate: bufferedFromDate,
      toDate: item.tradeDate,
    });
    setIsBackfillModalOpen(true);
  };

  // Trigger global BackfillModal for all missing pairs
  const handleGlobalBackfill = () => {
    const targetPairs = pairFilter !== 'ALL'
      ? [pairFilter]
      : availablePairs.length > 0
      ? availablePairs
      : items.map((i) => i.requiredPair).filter(Boolean);

    const from = earliestTradeDate ? calculateBufferDate(earliestTradeDate, 3) : fromDate || formatISODate(new Date());
    const to = latestTradeDate || toDate || formatISODate(new Date());

    setBackfillModalProps({
      currencyPairs: Array.from(new Set(targetPairs)),
      fromDate: from,
      toDate: to,
    });
    setIsBackfillModalOpen(true);
  };

  // Trigger manual FX override modal
  const handleRowManualOverride = (item: MissingFXRateItem) => {
    setTargetOverrideData({
      base: item.transactionCurrency,
      quote: item.portfolioBaseCurrency,
      date: item.tradeDate,
      rate: item.recordedFxRate || undefined,
    });
    setIsOverrideModalOpen(true);
  };

  // Backfill execution
  const executeBackfill = async (data: BackfillSubmitData): Promise<BackfillModalResult> => {
    try {
      const res = await client.mutation({
        triggerBackfill: {
          __args: {
            input: {
              fromDate: data.fromDate,
              toDate: data.toDate,
              symbols: data.symbols,
              currencyPairs: data.currencyPairs,
              backfillAssets: data.backfillAssets,
              backfillFx: data.backfillFx,
              recomputeValuations: data.recomputeValuations,
            },
          },
          success: true,
          pricesSynced: true,
          fxRatesSynced: true,
          message: true,
          warnings: true,
        },
      });

      return {
        success: res.triggerBackfill.success,
        pricesSynced: res.triggerBackfill.pricesSynced,
        fxRatesSynced: res.triggerBackfill.fxRatesSynced,
        message: res.triggerBackfill.message,
        warnings: res.triggerBackfill.warnings || [],
      };
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Backfill failed';
      return {
        success: false,
        pricesSynced: 0,
        fxRatesSynced: 0,
        message,
        warnings: [message],
      };
    }
  };

  // Manual FX rate override submit
  const executeOverride = async (data: FXOverrideData) => {
    try {
      await client.mutation({
        recordFXRateOverride: {
          __args: {
            input: {
              baseCurrency: data.baseCurrency,
              quoteCurrency: data.quoteCurrency,
              rateDate: data.rateDate,
              rate: data.rate,
              reason: data.reason || 'Missing FX Rates Report manual override',
              recomputeValuations: data.recomputeValuations,
            },
          },
          rate: {
            rateDate: true,
            rate: true,
          },
          valuationsRecomputed: true,
        },
      });

      setIsOverrideModalOpen(false);
      onNotify(`FX rate for ${data.baseCurrency}/${data.quoteCurrency} on ${data.rateDate} recorded.`);
      fetchReport();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to record FX override';
      onNotify(msg);
      throw err;
    }
  };

  // Filtered items (client-side safety)
  const displayItems = useMemo(() => {
    return items.filter((it) => {
      if (typeFilter !== 'ALL' && it.type !== typeFilter) {
        return false;
      }
      return true;
    });
  }, [items, typeFilter]);

  return (
    <div className="missing-fx-report">
      {/* 1. Summary KPI Cards */}
      <div className="missing-fx-report__stats">
        <Card className="missing-fx-stat-card">
          <div className="stat-label">Unconverted Trades</div>
          <div className={`stat-value ${totalCount > 0 ? 'missing-fx-stat-card .stat-value--alert' : ''}`}>
            {loading ? '...' : totalCount}
          </div>
        </Card>

        <Card className="missing-fx-stat-card">
          <div className="stat-label">Impacted Pairs</div>
          <div className="stat-value">{loading ? '...' : uniquePairsCount}</div>
        </Card>

        <Card className="missing-fx-stat-card">
          <div className="stat-label">Earliest Missing Date</div>
          <div className="stat-value">
            {loading ? '...' : earliestTradeDate ? formatDate(earliestTradeDate) : '—'}
          </div>
        </Card>

        <Card className="missing-fx-stat-card">
          <div className="stat-label">Date Span Coverage</div>
          <div className="stat-value">
            {loading ? '...' : calculateDateSpanDays(earliestTradeDate, latestTradeDate)}
          </div>
        </Card>
      </div>

      {/* 2. Filters & Toolbar */}
      <Card>
        <div className="missing-fx-report__toolbar">
          <div className="missing-fx-report__filters">
            {/* Pair Selector */}
            <div className="missing-fx-report__pair-filter">
              <Select
                id="missing-fx-pair-filter"
                value={pairFilter}
                onChange={(e) => {
                  setPairFilter(e.target.value);
                  setPage(1);
                }}
              >
                <option value="ALL">All Pairs ({availablePairs.length || 'All'})</option>
                {availablePairs.map((pair) => (
                  <option key={pair} value={pair}>
                    {pair}
                  </option>
                ))}
              </Select>
            </div>

            {/* Type Filter */}
            <div className="missing-fx-report__type-filter">
              <Select
                id="missing-fx-type-filter"
                value={typeFilter}
                onChange={(e) => {
                  setTypeFilter(e.target.value);
                  setPage(1);
                }}
              >
                <option value="ALL">All Types</option>
                <option value="BUY">BUY</option>
                <option value="SELL">SELL</option>
                <option value="DIVIDEND">DIVIDEND</option>
                <option value="DEPOSIT">DEPOSIT</option>
                <option value="WITHDRAWAL">WITHDRAWAL</option>
                <option value="TRANSFER_IN">TRANSFER IN</option>
                <option value="TRANSFER_OUT">TRANSFER OUT</option>
                <option value="INTEREST">INTEREST</option>
                <option value="FEE">FEE</option>
                <option value="TAX">TAX</option>
              </Select>
            </div>

            {/* From Date */}
            <div className="missing-fx-report__date-input-group">
              <Input
                id="missing-fx-from-date"
                type="date"
                className="missing-fx-report__date-input"
                value={fromDate}
                onChange={(e) => {
                  setFromDate(e.target.value);
                  setActivePreset(null);
                  setPage(1);
                }}
              />
            </div>

            <span>to</span>

            {/* To Date */}
            <div className="missing-fx-report__date-input-group">
              <Input
                id="missing-fx-to-date"
                type="date"
                className="missing-fx-report__date-input"
                value={toDate}
                onChange={(e) => {
                  setToDate(e.target.value);
                  setActivePreset(null);
                  setPage(1);
                }}
              />
            </div>

            {/* Quick Presets */}
            <div className="missing-fx-report__presets">
              {(['30D', '90D', 'YTD', '1Y', 'ALL'] as PresetKey[]).map((p) => (
                <Button
                  key={p}
                  variant={activePreset === p ? 'primary' : 'secondary'}
                  size="sm"
                  onClick={() => handlePresetSelect(p)}
                >
                  {p}
                </Button>
              ))}
            </div>

            {isDateRangeInvalid && (
              <span className="missing-fx-report__date-error">
                Start date cannot be after end date
              </span>
            )}
          </div>

          <div className="missing-fx-report__actions">
            <Button
              variant="primary"
              size="sm"
              disabled={totalCount === 0 || loading}
              onClick={handleGlobalBackfill}
            >
              Backfill All Missing FX
            </Button>
          </div>
        </div>
      </Card>

      {/* 3. Diagnostic Report Table */}
      <Card>
        {loading && items.length === 0 ? (
          <div style={{ padding: '2.5rem', textAlign: 'center', color: '#a1a1aa' }}>
            Checking trade ledger and foreign exchange fixings...
          </div>
        ) : displayItems.length === 0 ? (
          <div className="missing-fx-report__empty">
            <div className="missing-fx-report__empty-icon">🛡️</div>
            <div className="missing-fx-report__empty-title">All Cross-Currency Transactions Covered</div>
            <p className="missing-fx-report__empty-desc">
              Every recorded foreign currency transaction has a matching daily closing exchange rate on its trade date.
              No currency conversion gaps detected.
            </p>
            <Button variant="secondary" size="sm" onClick={fetchReport}>
              Re-run FX Verification
            </Button>
          </div>
        ) : (
          <>
            <Table>
              <TableHead>
                <TableRow>
                  <TableHeaderCell>Trade Date</TableHeaderCell>
                  <TableHeaderCell>Required Pair</TableHeaderCell>
                  <TableHeaderCell>Type</TableHeaderCell>
                  <TableHeaderCell>Asset / Symbol</TableHeaderCell>
                  <TableHeaderCell>Transacted Amount</TableHeaderCell>
                  <TableHeaderCell>Portfolio (Base)</TableHeaderCell>
                  <TableHeaderCell>Recorded FX</TableHeaderCell>
                  <TableHeaderCell>Status</TableHeaderCell>
                  <TableHeaderCell>Actions</TableHeaderCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {displayItems.map((item) => (
                  <TableRow key={item.transactionId}>
                    {/* Trade Date */}
                    <TableCell>{formatDate(item.tradeDate)}</TableCell>

                    {/* Required Pair */}
                    <TableCell>
                      <strong style={{ fontFamily: 'monospace', letterSpacing: '0.05em' }}>
                        {item.requiredPair}
                      </strong>
                    </TableCell>

                    {/* Type Badge */}
                    <TableCell>
                      <Badge variant={getTypeBadgeVariant(item.type)}>{item.type}</Badge>
                    </TableCell>

                    {/* Asset / Symbol */}
                    <TableCell>
                      {item.symbol ? (
                        <div className="missing-fx-report__asset-meta">
                          <span style={{ fontWeight: 600 }}>{item.symbol}</span>
                          {item.assetName && (
                            <span className="missing-fx-report__asset-name">{item.assetName}</span>
                          )}
                        </div>
                      ) : (
                        <span style={{ color: '#a1a1aa', fontStyle: 'italic' }}>Cash Movement</span>
                      )}
                    </TableCell>

                    {/* Amount */}
                    <TableCell>
                      <span style={{ fontWeight: 600 }}>
                        {formatMoney(item.amount.amount, item.amount.currencyCode)}
                      </span>
                    </TableCell>

                    {/* Portfolio */}
                    <TableCell>
                      <div className="missing-fx-report__portfolio-pill">
                        <span>{item.portfolioName}</span>
                        <Badge variant="neutral">[{item.portfolioBaseCurrency}]</Badge>
                      </div>
                    </TableCell>

                    {/* Recorded FX Rate */}
                    <TableCell>
                      {item.recordedFxRate ? (
                        <div className="missing-fx-report__recorded-fx">
                          <span>{item.recordedFxRate}</span>
                          <span className="missing-fx-report__recorded-note">(Broker snapshot)</span>
                        </div>
                      ) : (
                        <span style={{ color: '#71717a' }}>—</span>
                      )}
                    </TableCell>

                    {/* Price Status */}
                    <TableCell>
                      <Badge variant="warning">Missing Fixing</Badge>
                    </TableCell>

                    {/* Action Buttons */}
                    <TableCell>
                      <div className="missing-fx-report__actions-cell">
                        <Button
                          variant="secondary"
                          size="sm"
                          onClick={() => handleRowBackfill(item)}
                        >
                          Backfill Pair
                        </Button>
                        <Button
                          variant="secondary"
                          size="sm"
                          onClick={() => handleRowManualOverride(item)}
                        >
                          Manual FX
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>

            {/* Pagination Controls */}
            {totalPages > 1 && (
              <div className="missing-fx-report__pagination">
                <div>
                  Showing {(page - 1) * PAGE_SIZE + 1}–
                  {Math.min(page * PAGE_SIZE, totalCount)} of {totalCount} records
                </div>
                <div className="missing-fx-report__pagination-controls">
                  <Button
                    variant="secondary"
                    size="sm"
                    disabled={page <= 1}
                    onClick={() => setPage((p) => Math.max(1, p - 1))}
                  >
                    &lt; Prev
                  </Button>
                  <span>
                    Page {page} of {totalPages}
                  </span>
                  <Button
                    variant="secondary"
                    size="sm"
                    disabled={page >= totalPages}
                    onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                  >
                    Next &gt;
                  </Button>
                </div>
              </div>
            )}
          </>
        )}
      </Card>

      {/* 4. Connected Modals */}
      <BackfillModal
        isOpen={isBackfillModalOpen}
        onClose={() => setIsBackfillModalOpen(false)}
        initialCurrencyPairs={backfillModalProps.currencyPairs}
        initialFromDate={backfillModalProps.fromDate}
        initialToDate={backfillModalProps.toDate}
        initialBackfillAssets={false}
        initialBackfillFx={true}
        initialRecomputeValuations={true}
        onSubmit={executeBackfill}
        onSuccess={(result) => {
          onNotify(result.message || 'FX backfill job completed successfully.');
          fetchReport();
        }}
      />

      <FXOverrideModal
        isOpen={isOverrideModalOpen}
        onClose={() => setIsOverrideModalOpen(false)}
        defaultBaseCurrency={targetOverrideData.base}
        defaultQuoteCurrency={targetOverrideData.quote}
        defaultRateDate={targetOverrideData.date}
        defaultRate={targetOverrideData.rate}
        onSubmit={executeOverride}
      />
    </div>
  );
};

import React, { useState, useEffect, useCallback } from 'react';
import { client } from '@graphfolio/api-client';
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
import { PriceOverrideModal, type PriceOverrideData } from './PriceOverrideModal';
import { BackfillModal, type BackfillSubmitData, type BackfillModalResult } from './BackfillModal';
import './PriceManagement.css';

interface PriceRecord {
  id: string;
  symbol: string;
  priceDate: string;
  price: { amount: string; currencyCode: string };
  source: string;
  updatedAt: string;
}

interface PriceManagementProps {
  onNotify: (msg: string) => void;
}

type PresetKey = '7D' | '30D' | '90D' | 'YTD' | '1Y';

const PAGE_SIZE = 50;

function formatISODate(d: Date): string {
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, '0');
  const day = String(d.getDate()).padStart(2, '0');
  return `${y}-${m}-${day}`;
}

function calculatePresetDates(preset: PresetKey): { from: string; to: string } {
  const now = new Date();
  const todayISO = formatISODate(now);
  switch (preset) {
    case '7D': {
      const d = new Date(now);
      d.setDate(d.getDate() - 7);
      return { from: formatISODate(d), to: todayISO };
    }
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
  }
}

export const PriceManagement: React.FC<PriceManagementProps> = ({ onNotify }) => {
  const [prices, setPrices] = useState<PriceRecord[]>([]);
  const [totalCount, setTotalCount] = useState(0);
  const [loading, setLoading] = useState(true);
  const [symbolFilter, setSymbolFilter] = useState('ALL');
  const [fromDate, setFromDate] = useState('');
  const [toDate, setToDate] = useState('');
  const [activePreset, setActivePreset] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [isSyncing, setIsSyncing] = useState(false);
  const [isOverrideModalOpen, setIsOverrideModalOpen] = useState(false);
  const [isBackfillModalOpen, setIsBackfillModalOpen] = useState(false);
  const [selectedSymbolForOverride, setSelectedSymbolForOverride] = useState('AAPL');
  const [availableSymbols, setAvailableSymbols] = useState<string[]>([]);
  const [latestPriceDate, setLatestPriceDate] = useState<string>('—');

  const totalPages = Math.max(1, Math.ceil(totalCount / PAGE_SIZE));
  const isDateRangeInvalid = Boolean(fromDate && toDate && fromDate > toDate);

  // Load available instruments from BFF for the dropdown filter
  const fetchInstruments = useCallback(() => {
    client
      .query({
        allInstruments: {
          symbol: true,
          isActive: true,
        },
      })
      .then((res) => {
        if (res.allInstruments && res.allInstruments.length > 0) {
          const syms = Array.from(new Set(res.allInstruments.map((i) => i.symbol))).sort();
          setAvailableSymbols(syms);
          if (syms.length > 0 && !syms.includes(selectedSymbolForOverride)) {
            setSelectedSymbolForOverride(syms[0]);
          }
        }
      })
      .catch((err) => console.warn('Could not fetch instrument symbols for filter:', err));
  }, [selectedSymbolForOverride]);

  const fetchPrices = useCallback(() => {
    if (fromDate && toDate && fromDate > toDate) {
      return; // Don't query invalid ranges
    }

    setLoading(true);
    client
      .query({
        instrumentPrices: {
          __args: {
            symbol: symbolFilter === 'ALL' ? undefined : symbolFilter,
            fromDate: fromDate || undefined,
            toDate: toDate || undefined,
            limit: PAGE_SIZE,
            offset: (page - 1) * PAGE_SIZE,
          },
          items: {
            id: true,
            symbol: true,
            priceDate: true,
            price: {
              amount: true,
              currencyCode: true,
            },
            source: true,
            updatedAt: true,
          },
          totalCount: true,
        },
      })
      .then((res) => {
        if (res.instrumentPrices) {
          setPrices(res.instrumentPrices.items);
          setTotalCount(res.instrumentPrices.totalCount);
          if (res.instrumentPrices.items.length > 0 && page === 1) {
            setLatestPriceDate(res.instrumentPrices.items[0].priceDate);
          }
        }
        setLoading(false);
      })
      .catch((err) => {
        console.error('Failed to query instrument prices:', err);
        onNotify(`Error loading prices: ${err?.message || 'Server error'}`);
        setLoading(false);
      });
  }, [symbolFilter, fromDate, toDate, page, onNotify]);

  useEffect(() => {
    fetchInstruments();
  }, [fetchInstruments]);

  useEffect(() => {
    fetchPrices();
  }, [fetchPrices]);

  const handleApplyPreset = (preset: PresetKey) => {
    const dates = calculatePresetDates(preset);
    setFromDate(dates.from);
    setToDate(dates.to);
    setActivePreset(preset);
    setPage(1);
  };

  const handleClearFilters = () => {
    setSymbolFilter('ALL');
    setFromDate('');
    setToDate('');
    setActivePreset(null);
    setPage(1);
  };

  const handleTriggerSync = async () => {
    setIsSyncing(true);
    try {
      const res = await client.mutation({
        triggerMarketSync: {
          __args: {
            symbols: symbolFilter !== 'ALL' ? [symbolFilter] : undefined,
            syncFx: true,
          },
          success: true,
          pricesSynced: true,
          fxRatesSynced: true,
          message: true,
        },
      });

      if (res.triggerMarketSync.success) {
        onNotify(res.triggerMarketSync.message || 'Market data ingestion completed successfully.');
        fetchPrices();
      } else {
        onNotify(`Market sync issue: ${res.triggerMarketSync.message}`);
      }
    } catch (err: any) {
      console.error('Market sync error:', err);
      onNotify(`Market sync failed: ${err?.message || 'Server error'}`);
    } finally {
      setIsSyncing(false);
    }
  };

  const handleApplyOverride = async (data: PriceOverrideData) => {
    try {
      const res = await client.mutation({
        recordPriceOverride: {
          __args: {
            input: {
              symbol: data.symbol,
              priceDate: data.priceDate,
              price: data.price,
              reason: data.reason || undefined,
              recomputeValuations: data.recomputeValuations,
            },
          },
          price: {
            id: true,
            symbol: true,
            priceDate: true,
            price: {
              amount: true,
              currencyCode: true,
            },
            source: true,
            updatedAt: true,
          },
          valuationsRecomputed: true,
        },
      });

      const overridePrice = res.recordPriceOverride.price;
      onNotify(
        `Closing price override applied for ${data.symbol} on ${data.priceDate}: ${formatMoney(overridePrice.price)}${
          res.recordPriceOverride.valuationsRecomputed ? ' (Valuations recalibrated)' : ''
        }`
      );
      fetchPrices();
    } catch (err: any) {
      console.error('Price override error:', err);
      onNotify(`Price override failed: ${err?.message || 'Server error'}`);
      throw err;
    }
  };

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
        onNotify(payload.message || 'Asset prices backfilled successfully.');
        fetchPrices();
      } else {
        onNotify(`Backfill warning: ${payload.message}`);
      }

      return {
        success: payload.success,
        pricesSynced: payload.pricesSynced,
        fxRatesSynced: payload.fxRatesSynced,
        message: payload.message,
        warnings: payload.warnings,
      };
    } catch (err: any) {
      console.error('Backfill asset prices failed:', err);
      onNotify(`Backfill failed: ${err?.message || 'Server error'}`);
      throw err;
    }
  };

  return (
    <div className="price-mgmt">
      <div className="price-mgmt__stats">
        <Card className="price-stat-card">
          <div className="stat-label">Active Tracked Assets</div>
          <div className="stat-value">{availableSymbols.length}</div>
        </Card>
        <Card className="price-stat-card">
          <div className="stat-label">Latest Pricing Date</div>
          <div className="stat-value">{latestPriceDate !== '—' ? formatDate(latestPriceDate) : '—'}</div>
        </Card>
        <Card className="price-stat-card">
          <div className="stat-label">Authoritative Closing Records</div>
          <div className="stat-value">{totalCount}</div>
        </Card>
      </div>

      <div className="price-mgmt__toolbar">
        <div className="price-mgmt__filters">
          <Select
            className="price-mgmt__symbol-filter"
            value={symbolFilter}
            onChange={(e) => {
              setSymbolFilter(e.target.value);
              setPage(1);
            }}
            options={[
              { label: 'All Symbols', value: 'ALL' },
              ...availableSymbols.map((s) => ({ label: s, value: s })),
            ]}
          />

          <div className="price-mgmt__date-input-group">
            <Input
              className="price-mgmt__date-input"
              type="date"
              placeholder="From Date"
              value={fromDate}
              onChange={(e) => {
                setFromDate(e.target.value);
                setActivePreset(null);
                setPage(1);
              }}
              error={isDateRangeInvalid ? 'Invalid range' : undefined}
            />
          </div>

          <div className="price-mgmt__date-input-group">
            <Input
              className="price-mgmt__date-input"
              type="date"
              placeholder="To Date"
              value={toDate}
              onChange={(e) => {
                setToDate(e.target.value);
                setActivePreset(null);
                setPage(1);
              }}
              error={isDateRangeInvalid ? 'Must be ≥ From' : undefined}
            />
          </div>

          <div className="price-mgmt__presets">
            {(['7D', '30D', '90D', 'YTD', '1Y'] as const).map((preset) => (
              <Button
                key={preset}
                size="sm"
                variant={activePreset === preset ? 'primary' : 'ghost'}
                onClick={() => handleApplyPreset(preset)}
              >
                {preset}
              </Button>
            ))}
          </div>

          {(symbolFilter !== 'ALL' || fromDate || toDate) && (
            <Button
              size="sm"
              variant="ghost"
              onClick={handleClearFilters}
            >
              Clear Filters
            </Button>
          )}

          {isDateRangeInvalid && (
            <span className="price-mgmt__date-error">
              Start date cannot be after end date
            </span>
          )}
        </div>

        <div className="price-mgmt__actions">
          <Button
            variant="secondary"
            isLoading={isSyncing}
            onClick={handleTriggerSync}
          >
            ⚡ Trigger EOD Ingestion
          </Button>

          <Button
            variant="secondary"
            onClick={() => setIsBackfillModalOpen(true)}
          >
            ⚡ Backfill Asset Prices
          </Button>

          <Button
            variant="primary"
            onClick={() => {
              setSelectedSymbolForOverride(symbolFilter === 'ALL' ? (availableSymbols[0] || 'AAPL') : symbolFilter);
              setIsOverrideModalOpen(true);
            }}
          >
            + Manual Price Override
          </Button>
        </div>
      </div>

      {loading ? (
        <div style={{ padding: '2rem', textAlign: 'center', color: 'var(--text-secondary)' }}>
          Loading authoritative closing prices...
        </div>
      ) : (
        <>
          <Table hoverable striped>
            <TableHead>
              <TableRow>
                <TableHeaderCell>Date</TableHeaderCell>
                <TableHeaderCell>Symbol</TableHeaderCell>
                <TableHeaderCell>Closing Price</TableHeaderCell>
                <TableHeaderCell>Currency</TableHeaderCell>
                <TableHeaderCell>Data Source</TableHeaderCell>
                <TableHeaderCell>Last Updated</TableHeaderCell>
                <TableHeaderCell align="right">Actions</TableHeaderCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {prices.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={7} align="center">
                    No closing price records found for this filter.
                  </TableCell>
                </TableRow>
              ) : (
                prices.map((rec) => (
                  <TableRow key={rec.id}>
                    <TableCell>{formatDate(rec.priceDate)}</TableCell>
                    <TableCell>
                      <strong style={{ fontFamily: 'var(--font-mono, monospace)' }}>
                        {rec.symbol}
                      </strong>
                    </TableCell>
                    <TableCell>
                      <strong>{formatMoney(rec.price)}</strong>
                    </TableCell>
                    <TableCell>{rec.price.currencyCode}</TableCell>
                    <TableCell>
                      <Badge
                        variant={
                          rec.source === 'manual' || rec.source === 'MANUAL_OVERRIDE'
                            ? 'warning'
                            : rec.source === 'twelve_data' || rec.source === 'TWELVE_DATA'
                            ? 'info'
                            : 'neutral'
                        }
                      >
                        {rec.source}
                      </Badge>
                    </TableCell>
                    <TableCell style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>
                      {new Date(rec.updatedAt).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                    </TableCell>
                    <TableCell align="right">
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => {
                          setSelectedSymbolForOverride(rec.symbol);
                          setIsOverrideModalOpen(true);
                        }}
                      >
                        Override
                      </Button>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>

          <div className="price-mgmt__pagination">
            <div className="price-mgmt__pagination-info">
              Showing {prices.length > 0 ? (page - 1) * PAGE_SIZE + 1 : 0} to{' '}
              {Math.min(page * PAGE_SIZE, totalCount)} of {totalCount} records
            </div>
            <div className="price-mgmt__pagination-controls">
              <Button
                size="sm"
                variant="ghost"
                disabled={page <= 1 || loading}
                onClick={() => setPage((p) => Math.max(1, p - 1))}
              >
                Previous
              </Button>
              <span className="price-mgmt__page-indicator">
                Page {page} of {totalPages}
              </span>
              <Button
                size="sm"
                variant="ghost"
                disabled={page >= totalPages || loading}
                onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
              >
                Next
              </Button>
            </div>
          </div>
        </>
      )}

      <PriceOverrideModal
        isOpen={isOverrideModalOpen}
        onClose={() => setIsOverrideModalOpen(false)}
        defaultSymbol={selectedSymbolForOverride}
        onSubmit={handleApplyOverride}
      />

      <BackfillModal
        isOpen={isBackfillModalOpen}
        onClose={() => setIsBackfillModalOpen(false)}
        initialSymbols={symbolFilter !== 'ALL' ? [symbolFilter] : []}
        initialFromDate={fromDate || undefined}
        initialToDate={toDate || undefined}
        initialBackfillAssets={true}
        initialBackfillFx={false}
        initialRecomputeValuations={true}
        onSubmit={handleExecuteBackfill}
        onSuccess={() => fetchPrices()}
      />
    </div>
  );
};


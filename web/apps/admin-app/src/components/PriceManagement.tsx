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

export const PriceManagement: React.FC<PriceManagementProps> = ({ onNotify }) => {
  const [prices, setPrices] = useState<PriceRecord[]>([]);
  const [totalCount, setTotalCount] = useState(0);
  const [loading, setLoading] = useState(true);
  const [symbolFilter, setSymbolFilter] = useState('ALL');
  const [dateFilter, setDateFilter] = useState('');
  const [isSyncing, setIsSyncing] = useState(false);
  const [isOverrideModalOpen, setIsOverrideModalOpen] = useState(false);
  const [selectedSymbolForOverride, setSelectedSymbolForOverride] = useState('AAPL');
  const [availableSymbols, setAvailableSymbols] = useState<string[]>([]);
  const [latestPriceDate, setLatestPriceDate] = useState<string>('—');

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
    setLoading(true);
    client
      .query({
        instrumentPrices: {
          __args: {
            symbol: symbolFilter === 'ALL' ? undefined : symbolFilter,
            fromDate: dateFilter || undefined,
            toDate: dateFilter || undefined,
            limit: 50,
            offset: 0,
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
          if (res.instrumentPrices.items.length > 0) {
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
  }, [symbolFilter, dateFilter, onNotify]);

  useEffect(() => {
    fetchInstruments();
  }, [fetchInstruments]);

  useEffect(() => {
    fetchPrices();
  }, [fetchPrices]);

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
            onChange={(e) => setSymbolFilter(e.target.value)}
            options={[
              { label: 'All Symbols', value: 'ALL' },
              ...availableSymbols.map((s) => ({ label: s, value: s })),
            ]}
          />
          <Input
            className="price-mgmt__date-filter"
            type="date"
            value={dateFilter}
            onChange={(e) => setDateFilter(e.target.value)}
          />
          {(symbolFilter !== 'ALL' || dateFilter) && (
            <Button
              size="sm"
              variant="ghost"
              onClick={() => {
                setSymbolFilter('ALL');
                setDateFilter('');
              }}
            >
              Clear Filters
            </Button>
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
      )}

      <PriceOverrideModal
        isOpen={isOverrideModalOpen}
        onClose={() => setIsOverrideModalOpen(false)}
        defaultSymbol={selectedSymbolForOverride}
        onSubmit={handleApplyOverride}
      />
    </div>
  );
};

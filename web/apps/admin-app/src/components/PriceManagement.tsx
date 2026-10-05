import React, { useState, useEffect } from 'react';
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
  price: string;
  currencyCode: string;
  source: 'TWELVE_DATA' | 'YAHOO_FINANCE' | 'MANUAL_OVERRIDE' | 'ECB';
  reason?: string;
  updatedAt: string;
}

interface PriceManagementProps {
  onNotify: (msg: string) => void;
}

const INITIAL_PRICES: PriceRecord[] = [
  {
    id: 'p-1',
    symbol: 'AAPL',
    priceDate: '2026-10-02',
    price: '228.45',
    currencyCode: 'USD',
    source: 'TWELVE_DATA',
    updatedAt: '2026-10-02T21:05:00Z',
  },
  {
    id: 'p-2',
    symbol: 'MSFT',
    priceDate: '2026-10-02',
    price: '430.15',
    currencyCode: 'USD',
    source: 'TWELVE_DATA',
    updatedAt: '2026-10-02T21:05:00Z',
  },
  {
    id: 'p-3',
    symbol: 'GOOGL',
    priceDate: '2026-10-02',
    price: '185.70',
    currencyCode: 'USD',
    source: 'TWELVE_DATA',
    updatedAt: '2026-10-02T21:05:00Z',
  },
  {
    id: 'p-4',
    symbol: 'VOO',
    priceDate: '2026-10-02',
    price: '524.80',
    currencyCode: 'USD',
    source: 'YAHOO_FINANCE',
    updatedAt: '2026-10-02T21:10:00Z',
  },
  {
    id: 'p-5',
    symbol: 'BTC',
    priceDate: '2026-10-02',
    price: '64500.00',
    currencyCode: 'USD',
    source: 'TWELVE_DATA',
    updatedAt: '2026-10-02T21:00:00Z',
  },
  {
    id: 'p-6',
    symbol: 'AAPL',
    priceDate: '2026-10-01',
    price: '226.75',
    currencyCode: 'USD',
    source: 'TWELVE_DATA',
    updatedAt: '2026-10-01T21:05:00Z',
  },
  {
    id: 'p-7',
    symbol: 'MSFT',
    priceDate: '2026-10-01',
    price: '428.90',
    currencyCode: 'USD',
    source: 'TWELVE_DATA',
    updatedAt: '2026-10-01T21:05:00Z',
  },
];

export const PriceManagement: React.FC<PriceManagementProps> = ({ onNotify }) => {
  const [prices, setPrices] = useState<PriceRecord[]>(INITIAL_PRICES);
  const [symbolFilter, setSymbolFilter] = useState('ALL');
  const [dateFilter, setDateFilter] = useState('');
  const [isSyncing, setIsSyncing] = useState(false);
  const [isOverrideModalOpen, setIsOverrideModalOpen] = useState(false);
  const [selectedSymbolForOverride, setSelectedSymbolForOverride] = useState('AAPL');

  // Load available instruments from BFF for the dropdown filter
  const [availableSymbols, setAvailableSymbols] = useState<string[]>(['AAPL', 'MSFT', 'GOOGL', 'VOO', 'BTC']);

  useEffect(() => {
    client
      .query({
        instruments: {
          symbol: true,
        },
      })
      .then((res) => {
        if (res.instruments && res.instruments.length > 0) {
          const syms = Array.from(new Set(res.instruments.map((i) => i.symbol)));
          setAvailableSymbols(syms);
        }
      })
      .catch((err) => console.warn('Could not fetch instrument symbols for filter:', err));
  }, []);

  const handleTriggerSync = async () => {
    setIsSyncing(true);
    // Simulate pipeline ingestion step
    setTimeout(() => {
      setIsSyncing(false);
      onNotify('Market data ingestion pipeline triggered! 5 EOD prices refreshed.');
    }, 1200);
  };

  const handleApplyOverride = async (data: PriceOverrideData) => {
    const newRecord: PriceRecord = {
      id: `override-${Date.now()}`,
      symbol: data.symbol,
      priceDate: data.priceDate,
      price: data.price,
      currencyCode: data.currencyCode,
      source: 'MANUAL_OVERRIDE',
      reason: data.reason,
      updatedAt: new Date().toISOString(),
    };

    setPrices((prev) => [newRecord, ...prev]);
    onNotify(
      `Closing price override applied for ${data.symbol} on ${data.priceDate}: $${data.price}`
    );
  };

  const filteredPrices = prices.filter((rec) => {
    const matchesSymbol = symbolFilter === 'ALL' || rec.symbol === symbolFilter;
    const matchesDate = !dateFilter || rec.priceDate === dateFilter;
    return matchesSymbol && matchesDate;
  });

  return (
    <div className="price-mgmt">
      <div className="price-mgmt__stats">
        <Card className="price-stat-card">
          <div className="stat-label">Active Tracked Assets</div>
          <div className="stat-value">{availableSymbols.length}</div>
        </Card>
        <Card className="price-stat-card">
          <div className="stat-label">Latest Pricing Date</div>
          <div className="stat-value">2026-10-02</div>
        </Card>
        <Card className="price-stat-card">
          <div className="stat-label">Pipeline Feed Status</div>
          <div className="stat-value" style={{ color: 'var(--accent-green)' }}>
            Healthy (EOD Fixings OK)
          </div>
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
              setSelectedSymbolForOverride(symbolFilter === 'ALL' ? 'AAPL' : symbolFilter);
              setIsOverrideModalOpen(true);
            }}
          >
            + Manual Price Override
          </Button>
        </div>
      </div>

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
          {filteredPrices.length === 0 ? (
            <TableRow>
              <TableCell colSpan={7} align="center">
                No closing price records found for this filter.
              </TableCell>
            </TableRow>
          ) : (
            filteredPrices.map((rec) => (
              <TableRow key={rec.id}>
                <TableCell>{formatDate(rec.priceDate)}</TableCell>
                <TableCell>
                  <strong style={{ fontFamily: 'var(--font-mono, monospace)' }}>
                    {rec.symbol}
                  </strong>
                </TableCell>
                <TableCell>
                  <strong>{formatMoney({ amount: rec.price, currencyCode: rec.currencyCode })}</strong>
                </TableCell>
                <TableCell>{rec.currencyCode}</TableCell>
                <TableCell>
                  <Badge
                    variant={
                      rec.source === 'MANUAL_OVERRIDE'
                        ? 'warning'
                        : rec.source === 'TWELVE_DATA'
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

      <PriceOverrideModal
        isOpen={isOverrideModalOpen}
        onClose={() => setIsOverrideModalOpen(false)}
        defaultSymbol={selectedSymbolForOverride}
        onSubmit={handleApplyOverride}
      />
    </div>
  );
};

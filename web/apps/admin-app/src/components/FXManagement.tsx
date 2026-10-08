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
  formatDate,
  formatDecimal,
} from '@graphfolio/ui';
import { FXTrendChart, type FXTimeframeKey } from './FXTrendChart';
import { FXOverrideModal, type FXOverrideData } from './FXOverrideModal';
import './FXManagement.css';

interface CurrencyPairSummary {
  baseCurrency: string;
  quoteCurrency: string;
  pair: string;
  latestRate: string;
  latestDate: string;
  latestSource: string;
  previousRate?: string | null;
  change1dAmount?: string | null;
  change1dPct?: string | null;
  totalRecords: number;
  firstDate: string;
  lastDate: string;
}

interface FXRateItem {
  baseCurrency: string;
  quoteCurrency: string;
  pair: string;
  rateDate: string;
  rate: string;
  invertedRate: string;
  source: string;
}

interface FXManagementProps {
  onNotify: (msg: string) => void;
}

const DEFAULT_POPULAR_PAIRS = ['EUR/USD', 'USD/GBP', 'USD/AUD', 'USD/JPY', 'EUR/GBP', 'USD/CAD'];

export const FXManagement: React.FC<FXManagementProps> = ({ onNotify }) => {
  const [pairs, setPairs] = useState<CurrencyPairSummary[]>([]);
  const [selectedBase, setSelectedBase] = useState('EUR');
  const [selectedQuote, setSelectedQuote] = useState('USD');
  const [isInverted, setIsInverted] = useState(false);
  const [timeframe, setTimeframe] = useState<FXTimeframeKey>('1Y');

  // Rates ledger state
  const [rates, setRates] = useState<FXRateItem[]>([]);
  const [totalCount, setTotalCount] = useState(0);
  const [page, setPage] = useState(1);
  const pageSize = 20;
  const [fromDate, setFromDate] = useState('');
  const [toDate, setToDate] = useState('');
  const [loadingRates, setLoadingRates] = useState(true);

  // Sync & Modal state
  const [isSyncing, setIsSyncing] = useState(false);
  const [isOverrideModalOpen, setIsOverrideModalOpen] = useState(false);
  const [modalPrefill, setModalPrefill] = useState<{
    base: string;
    quote: string;
    date?: string;
    rate?: string;
  }>({
    base: 'EUR',
    quote: 'USD',
  });

  // 1. Fetch available currency pairs
  const fetchPairs = useCallback(() => {
    client
      .query({
        currencyPairs: {
          baseCurrency: true,
          quoteCurrency: true,
          pair: true,
          latestRate: true,
          latestDate: true,
          latestSource: true,
          previousRate: true,
          change1dAmount: true,
          change1dPct: true,
          totalRecords: true,
          firstDate: true,
          lastDate: true,
        },
      })
      .then((res) => {
        if (res.currencyPairs && res.currencyPairs.length > 0) {
          const fetched = res.currencyPairs as CurrencyPairSummary[];
          setPairs(fetched);

          // If current selection is not in the list, set to first
          const pairExists = fetched.some(
            (p) => p.baseCurrency === selectedBase && p.quoteCurrency === selectedQuote
          );
          if (!pairExists && fetched.length > 0) {
            setSelectedBase(fetched[0].baseCurrency);
            setSelectedQuote(fetched[0].quoteCurrency);
          }
        }
      })
      .catch((err) => console.warn('Could not fetch currency pairs:', err));
  }, [selectedBase, selectedQuote]);

  useEffect(() => {
    fetchPairs();
  }, [fetchPairs]);

  // 2. Fetch ledger rates with filters & pagination
  const fetchRates = useCallback(() => {
    setLoadingRates(true);
    client
      .query({
        fxRates: {
          __args: {
            baseCurrency: selectedBase,
            quoteCurrency: selectedQuote,
            fromDate: fromDate || undefined,
            toDate: toDate || undefined,
            limit: pageSize,
            offset: (page - 1) * pageSize,
          },
          items: {
            baseCurrency: true,
            quoteCurrency: true,
            pair: true,
            rateDate: true,
            rate: true,
            invertedRate: true,
            source: true,
          },
          totalCount: true,
        },
      })
      .then((res) => {
        if (res.fxRates) {
          setRates(res.fxRates.items as FXRateItem[]);
          setTotalCount(res.fxRates.totalCount);
        }
        setLoadingRates(false);
      })
      .catch((err) => {
        console.warn('Could not fetch FX rates:', err);
        setLoadingRates(false);
      });
  }, [selectedBase, selectedQuote, fromDate, toDate, page, pageSize]);

  useEffect(() => {
    fetchRates();
  }, [fetchRates]);

  // Handle active pair info
  const activePair = pairs.find(
    (p) => p.baseCurrency === selectedBase && p.quoteCurrency === selectedQuote
  );

  const activePairLabel = isInverted
    ? `${selectedQuote}/${selectedBase}`
    : `${selectedBase}/${selectedQuote}`;

  const latestRateNum = activePair
    ? isInverted && Number(activePair.latestRate) > 0
      ? 1 / Number(activePair.latestRate)
      : Number(activePair.latestRate)
    : 0;

  const change1dAmountNum = activePair && activePair.change1dAmount
    ? isInverted && Number(activePair.latestRate) > 0 && activePair.previousRate
      ? 1 / Number(activePair.latestRate) - 1 / Number(activePair.previousRate)
      : Number(activePair.change1dAmount)
    : null;

  const change1dPctNum = activePair && activePair.change1dPct
    ? isInverted && activePair.previousRate && Number(activePair.previousRate) > 0
      ? ((1 / Number(activePair.latestRate) - 1 / Number(activePair.previousRate)) / (1 / Number(activePair.previousRate))) * 100
      : Number(activePair.change1dPct)
    : null;

  const isChangePositive = change1dAmountNum !== null && change1dAmountNum >= 0;

  // Toggle inversion
  const handleInvert = () => {
    setIsInverted(!isInverted);
  };

  // Select quick pair
  const handleSelectPair = (base: string, quote: string) => {
    setSelectedBase(base);
    setSelectedQuote(quote);
    setIsInverted(false);
    setPage(1);
  };

  // Trigger FX Sync
  const handleSyncFx = async () => {
    try {
      setIsSyncing(true);
      const res = await client.mutation({
        triggerMarketSync: {
          __args: {
            syncFx: true,
          },
          success: true,
          fxRatesSynced: true,
          message: true,
        },
      });
      if (res.triggerMarketSync) {
        onNotify(res.triggerMarketSync.message || `Synchronized ${res.triggerMarketSync.fxRatesSynced} FX rates.`);
        fetchPairs();
        fetchRates();
      }
    } catch (err: any) {
      onNotify(err?.message || 'Failed to trigger FX market synchronization.');
    } finally {
      setIsSyncing(false);
    }
  };

  // Open override modal
  const handleOpenOverride = (base = selectedBase, quote = selectedQuote, date?: string, rate?: string) => {
    setModalPrefill({ base, quote, date, rate });
    setIsOverrideModalOpen(true);
  };

  // Submit override
  const handleOverrideSubmit = async (data: FXOverrideData) => {
    const res = await client.mutation({
      recordFXRateOverride: {
        __args: {
          input: {
            baseCurrency: data.baseCurrency,
            quoteCurrency: data.quoteCurrency,
            rateDate: data.rateDate,
            rate: data.rate,
            reason: data.reason || undefined,
            recomputeValuations: data.recomputeValuations,
          },
        },
        valuationsRecomputed: true,
        rate: {
          pair: true,
          rate: true,
          rateDate: true,
        },
      },
    });

    if (res.recordFXRateOverride) {
      onNotify(
        `Applied FX override for ${data.baseCurrency}/${data.quoteCurrency} on ${data.rateDate} (${data.rate}).`
      );
      fetchPairs();
      fetchRates();
    }
  };

  const totalPages = Math.max(1, Math.ceil(totalCount / pageSize));

  return (
    <div className="fx-mgmt">
      {/* Top Toolbar: Quick Pair Pills & Action Buttons */}
      <div className="fx-mgmt__toolbar">
        <div className="fx-mgmt__pair-pills">
          {pairs.length > 0
            ? pairs.slice(0, 7).map((p) => {
                const isActive = p.baseCurrency === selectedBase && p.quoteCurrency === selectedQuote;
                return (
                  <button
                    key={p.pair}
                    type="button"
                    className={`fx-pill ${isActive ? 'fx-pill--active' : ''}`}
                    onClick={() => handleSelectPair(p.baseCurrency, p.quoteCurrency)}
                  >
                    {p.pair}
                  </button>
                );
              })
            : DEFAULT_POPULAR_PAIRS.map((pairStr) => {
                const [b, q] = pairStr.split('/');
                const isActive = b === selectedBase && q === selectedQuote;
                return (
                  <button
                    key={pairStr}
                    type="button"
                    className={`fx-pill ${isActive ? 'fx-pill--active' : ''}`}
                    onClick={() => handleSelectPair(b, q)}
                  >
                    {pairStr}
                  </button>
                );
              })}
        </div>

        <div className="fx-mgmt__actions">
          <Button variant="ghost" onClick={handleInvert}>
            <span>⇄ Invert ({isInverted ? 'Quote ⇄ Base' : 'Base ⇄ Quote'})</span>
          </Button>

          <Button variant="secondary" onClick={handleSyncFx} isLoading={isSyncing}>
            <span>⚡ Sync FX Rates</span>
          </Button>

          <Button variant="primary" onClick={() => handleOpenOverride()}>
            <span>+ Override Rate</span>
          </Button>
        </div>
      </div>

      {/* KPI Summary Cards */}
      <div className="fx-mgmt__stats">
        <Card className="fx-stat-card">
          <div className="stat-label">Current Spot Rate ({activePairLabel})</div>
          <div className="stat-value">
            {latestRateNum > 0 ? latestRateNum.toFixed(6) : '—'}
          </div>
          <div className="stat-sub">
            {activePair ? `As of ${formatDate(activePair.latestDate)}` : 'No observations'}
          </div>
        </Card>

        <Card className="fx-stat-card">
          <div className="stat-label">1-Day Change</div>
          <div className="stat-value">
            {change1dAmountNum !== null ? (
              <Badge variant={isChangePositive ? 'success' : 'danger'}>
                {isChangePositive ? '+' : ''}
                {change1dAmountNum.toFixed(6)} ({isChangePositive ? '+' : ''}
                {change1dPctNum?.toFixed(2)}%)
              </Badge>
            ) : (
              '—'
            )}
          </div>
          <div className="stat-sub">Prior business day delta</div>
        </Card>

        <Card className="fx-stat-card">
          <div className="stat-label">Primary Feed Source</div>
          <div className="stat-value">
            <Badge variant="info">{activePair?.latestSource || 'ECB Fixing'}</Badge>
          </div>
          <div className="stat-sub">Official Reference Benchmark</div>
        </Card>

        <Card className="fx-stat-card">
          <div className="stat-label">Total Observations</div>
          <div className="stat-value">
            {activePair ? activePair.totalRecords : totalCount}
          </div>
          <div className="stat-sub">
            {activePair?.firstDate ? `${formatDate(activePair.firstDate)} → ${formatDate(activePair.lastDate)}` : 'Daily history points'}
          </div>
        </Card>
      </div>

      {/* Interactive Trend Chart */}
      <FXTrendChart
        baseCurrency={selectedBase}
        quoteCurrency={selectedQuote}
        isInverted={isInverted}
        timeframe={timeframe}
        onTimeframeChange={setTimeframe}
      />

      {/* Authoritative Rates Ledger (Table) */}
      <Card className="fx-ledger-card">
        <div className="fx-ledger-card__header">
          <div className="fx-ledger-card__title">
            Authoritative Exchange Rates Ledger ({selectedBase}/{selectedQuote})
          </div>

          <div className="fx-ledger-card__filters">
            <div className="fx-ledger-card__date-input">
              <Input
                type="date"
                placeholder="From Date"
                value={fromDate}
                onChange={(e) => {
                  setFromDate(e.target.value);
                  setPage(1);
                }}
              />
            </div>
            <div className="fx-ledger-card__date-input">
              <Input
                type="date"
                placeholder="To Date"
                value={toDate}
                onChange={(e) => {
                  setToDate(e.target.value);
                  setPage(1);
                }}
              />
            </div>
            {(fromDate || toDate) && (
              <Button
                variant="ghost"
                onClick={() => {
                  setFromDate('');
                  setToDate('');
                  setPage(1);
                }}
              >
                Clear
              </Button>
            )}
          </div>
        </div>

        <Table>
          <TableHead>
            <TableRow>
              <TableHeaderCell>Rate Date</TableHeaderCell>
              <TableHeaderCell>Currency Pair</TableHeaderCell>
              <TableHeaderCell>Exchange Rate</TableHeaderCell>
              <TableHeaderCell>Reciprocal Rate (1/X)</TableHeaderCell>
              <TableHeaderCell>Benchmark Source</TableHeaderCell>
              <TableHeaderCell style={{ textAlign: 'right' }}>Actions</TableHeaderCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {loadingRates ? (
              <TableRow>
                <TableCell colSpan={6} style={{ textAlign: 'center', padding: '2rem' }}>
                  Loading exchange rates...
                </TableCell>
              </TableRow>
            ) : rates.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6} style={{ textAlign: 'center', padding: '2rem' }}>
                  No exchange rate records found for this currency pair and filter criteria.
                </TableCell>
              </TableRow>
            ) : (
              rates.map((r, idx) => (
                <TableRow key={`${r.pair}-${r.rateDate}-${idx}`}>
                  <TableCell>{formatDate(r.rateDate)}</TableCell>
                  <TableCell>
                    <strong>{r.pair || `${r.baseCurrency}/${r.quoteCurrency}`}</strong>
                  </TableCell>
                  <TableCell style={{ fontVariantNumeric: 'tabular-nums' }}>
                    {formatDecimal(r.rate, 6)}
                  </TableCell>
                  <TableCell style={{ fontVariantNumeric: 'tabular-nums', color: 'var(--text-secondary, #a1a1aa)' }}>
                    {formatDecimal(r.invertedRate, 6)}
                  </TableCell>
                  <TableCell>
                    <Badge variant={r.source.startsWith('manual') ? 'warning' : 'info'}>
                      {r.source}
                    </Badge>
                  </TableCell>
                  <TableCell style={{ textAlign: 'right' }}>
                    <Button
                      variant="ghost"
                      onClick={() =>
                        handleOpenOverride(r.baseCurrency, r.quoteCurrency, r.rateDate, r.rate)
                      }
                    >
                      Override
                    </Button>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>

        {/* Pagination Bar */}
        <div className="fx-ledger-card__pagination">
          <div className="fx-pagination__info">
            Showing {rates.length > 0 ? (page - 1) * pageSize + 1 : 0} to{' '}
            {Math.min(page * pageSize, totalCount)} of {totalCount} observations
          </div>

          <div className="fx-pagination__controls">
            <Button
              variant="ghost"
              disabled={page <= 1 || loadingRates}
              onClick={() => setPage((p) => Math.max(1, p - 1))}
            >
              Previous
            </Button>
            <span style={{ fontSize: '0.85rem', color: 'var(--text-secondary, #a1a1aa)' }}>
              Page {page} of {totalPages}
            </span>
            <Button
              variant="ghost"
              disabled={page >= totalPages || loadingRates}
              onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
            >
              Next
            </Button>
          </div>
        </div>
      </Card>

      {/* Manual Rate Override Modal */}
      <FXOverrideModal
        isOpen={isOverrideModalOpen}
        onClose={() => setIsOverrideModalOpen(false)}
        defaultBaseCurrency={modalPrefill.base}
        defaultQuoteCurrency={modalPrefill.quote}
        defaultRateDate={modalPrefill.date}
        defaultRate={modalPrefill.rate}
        onSubmit={handleOverrideSubmit}
      />
    </div>
  );
};

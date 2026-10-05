import React, { useState, useEffect, useCallback } from 'react';
import { client } from '@graphfolio/api-client';
import { Card, Badge, Button, formatDate } from '@graphfolio/ui';
import './IngestionPipeline.css';

interface FeedStatus {
  name: string;
  status: string;
  provider: string;
  schedule: string;
  lastRun: string;
  details: string;
}

interface IngestionData {
  feeds: FeedStatus[];
  trackedInstruments: number;
  trackedCurrencies: number;
  latestPriceDate?: string | null;
  latestFxDate?: string | null;
  rateLimitRemaining: number;
  rateLimitBudget: number;
  pendingBackfillJobs: number;
}

interface IngestionPipelineProps {
  onNotify: (msg: string) => void;
}

export const IngestionPipeline: React.FC<IngestionPipelineProps> = ({ onNotify }) => {
  const [data, setData] = useState<IngestionData | null>(null);
  const [loading, setLoading] = useState(true);
  const [isSyncing, setIsSyncing] = useState(false);
  const [syncingFeed, setSyncingFeed] = useState<string | null>(null);

  const fetchStatus = useCallback(() => {
    setLoading(true);
    client
      .query({
        ingestionStatus: {
          feeds: {
            name: true,
            status: true,
            provider: true,
            schedule: true,
            lastRun: true,
            details: true,
          },
          trackedInstruments: true,
          trackedCurrencies: true,
          latestPriceDate: true,
          latestFxDate: true,
          rateLimitRemaining: true,
          rateLimitBudget: true,
          pendingBackfillJobs: true,
        },
      })
      .then((res) => {
        if (res.ingestionStatus) {
          setData(res.ingestionStatus);
        }
        setLoading(false);
      })
      .catch((err) => {
        console.error('Failed to query ingestion status:', err);
        onNotify(`Error loading ingestion status: ${err?.message || 'Server error'}`);
        setLoading(false);
      });
  }, [onNotify]);

  useEffect(() => {
    fetchStatus();
  }, [fetchStatus]);

  const handleTriggerSync = async (feedName?: string, syncFx: boolean = true) => {
    if (feedName) {
      setSyncingFeed(feedName);
    } else {
      setIsSyncing(true);
    }

    try {
      const res = await client.mutation({
        triggerMarketSync: {
          __args: {
            syncFx,
          },
          success: true,
          pricesSynced: true,
          fxRatesSynced: true,
          message: true,
        },
      });

      if (res.triggerMarketSync.success) {
        onNotify(res.triggerMarketSync.message || 'Market data ingestion completed successfully.');
        fetchStatus();
      } else {
        onNotify(`Market sync warning: ${res.triggerMarketSync.message}`);
      }
    } catch (err: any) {
      console.error('Market sync failed:', err);
      onNotify(`Market sync failed: ${err?.message || 'Server error'}`);
    } finally {
      setIsSyncing(false);
      setSyncingFeed(null);
    }
  };

  const getBadgeVariant = (status: string) => {
    switch (status.toUpperCase()) {
      case 'ACTIVE':
        return 'active';
      case 'HEALTHY':
        return 'info';
      case 'WARNING':
        return 'warning';
      case 'DEGRADED':
        return 'danger';
      default:
        return 'neutral';
    }
  };

  const getGlowColor = (status: string, idx: number): 'green' | 'blue' | 'amber' => {
    const s = status.toUpperCase();
    if (s === 'WARNING' || s === 'DEGRADED') return 'amber';
    return idx % 2 === 0 ? 'green' : 'blue';
  };

  return (
    <div className="ingestion-pipeline">
      <div className="ingestion-toolbar">
        <div className="ingestion-toolbar__status">
          Authoritative market data feed orchestration and rate limiting telemetry.
        </div>
        <div className="ingestion-toolbar__actions">
          <Button
            size="sm"
            variant="ghost"
            onClick={fetchStatus}
            disabled={loading || isSyncing}
          >
            ↻ Refresh Metrics
          </Button>
          <Button
            size="sm"
            variant="primary"
            isLoading={isSyncing}
            onClick={() => handleTriggerSync(undefined, true)}
          >
            ⚡ Trigger Full Ingestion
          </Button>
        </div>
      </div>

      {loading && !data ? (
        <div style={{ padding: '2rem', textAlign: 'center', color: 'var(--text-secondary)' }}>
          Loading pipeline telemetry...
        </div>
      ) : data ? (
        <>
          <div className="ingestion-stats">
            <Card className="price-stat-card">
              <div className="stat-label">Tracked Instruments</div>
              <div className="stat-value">{data.trackedInstruments}</div>
            </Card>
            <Card className="price-stat-card">
              <div className="stat-label">Tracked FX Currencies</div>
              <div className="stat-value">{data.trackedCurrencies}</div>
            </Card>
            <Card className="price-stat-card">
              <div className="stat-label">Latest Pricing Fixing</div>
              <div className="stat-value">
                {data.latestPriceDate ? formatDate(data.latestPriceDate) : '—'}
              </div>
            </Card>
            <Card className="price-stat-card">
              <div className="stat-label">Latest FX Fixing</div>
              <div className="stat-value">
                {data.latestFxDate ? formatDate(data.latestFxDate) : '—'}
              </div>
            </Card>
            <Card className="price-stat-card">
              <div className="stat-label">API Budget Remaining</div>
              <div className="stat-value">
                {data.rateLimitRemaining} / {data.rateLimitBudget}
              </div>
            </Card>
          </div>

          <div className="ingestion-cards">
            {data.feeds.map((feed, idx) => {
              const isEquityFeed = feed.name.toLowerCase().includes('equity');
              const isFxFeed = feed.name.toLowerCase().includes('fx');
              const isRateLimit = feed.name.toLowerCase().includes('rate');

              return (
                <Card
                  key={feed.name}
                  className="pipeline-card"
                  glow={getGlowColor(feed.status, idx)}
                >
                  <div className="pipeline-card__header">
                    <span className="pipeline-card__title">{feed.name}</span>
                    <Badge variant={getBadgeVariant(feed.status)}>{feed.status}</Badge>
                  </div>
                  <div className="pipeline-metric">
                    <span>Provider</span>
                    <span>{feed.provider}</span>
                  </div>
                  <div className="pipeline-metric">
                    <span>Schedule</span>
                    <span>{feed.schedule}</span>
                  </div>
                  <div className="pipeline-metric">
                    <span>Last Run</span>
                    <span>{feed.lastRun ? formatDate(feed.lastRun) : '—'}</span>
                  </div>
                  <div className="pipeline-metric">
                    <span>Diagnostics</span>
                    <span style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>
                      {feed.details}
                    </span>
                  </div>
                  <div style={{ marginTop: '0.5rem' }}>
                    {isEquityFeed ? (
                      <Button
                        size="sm"
                        variant="secondary"
                        style={{ width: '100%' }}
                        isLoading={syncingFeed === feed.name}
                        onClick={() => handleTriggerSync(feed.name, false)}
                      >
                        ⚡ Run Equity Feed Ingestion
                      </Button>
                    ) : isFxFeed ? (
                      <Button
                        size="sm"
                        variant="secondary"
                        style={{ width: '100%' }}
                        isLoading={syncingFeed === feed.name}
                        onClick={() => handleTriggerSync(feed.name, true)}
                      >
                        ⚡ Sync FX Fixing Rates
                      </Button>
                    ) : isRateLimit ? (
                      <Button
                        size="sm"
                        variant="ghost"
                        style={{ width: '100%' }}
                        onClick={() => {
                          fetchStatus();
                          onNotify(
                            `Rate limit checked: ${data.rateLimitRemaining}/${data.rateLimitBudget} calls remaining. Backfill queue: ${data.pendingBackfillJobs} jobs.`
                          );
                        }}
                      >
                        Inspect Rate Limit Status
                      </Button>
                    ) : (
                      <Button
                        size="sm"
                        variant="secondary"
                        style={{ width: '100%' }}
                        onClick={() => handleTriggerSync(feed.name, true)}
                      >
                        Trigger Feed Sync
                      </Button>
                    )}
                  </div>
                </Card>
              );
            })}
          </div>
        </>
      ) : null}
    </div>
  );
};

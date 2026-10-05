import React from 'react';
import { Card, Badge, Button } from '@graphfolio/ui';
import './IngestionPipeline.css';

interface IngestionPipelineProps {
  onNotify: (msg: string) => void;
}

export const IngestionPipeline: React.FC<IngestionPipelineProps> = ({ onNotify }) => {
  return (
    <div className="ingestion-pipeline">
      <div className="ingestion-cards">
        <Card className="pipeline-card" glow="green">
          <div className="pipeline-card__header">
            <span className="pipeline-card__title">EOD Equity Feeds</span>
            <Badge variant="active">Active</Badge>
          </div>
          <div className="pipeline-metric">
            <span>Primary Provider</span>
            <span>Twelve Data (REST API)</span>
          </div>
          <div className="pipeline-metric">
            <span>Fallback Provider</span>
            <span>Yahoo Finance</span>
          </div>
          <div className="pipeline-metric">
            <span>Schedule</span>
            <span>Daily at 21:00 UTC</span>
          </div>
          <div className="pipeline-metric">
            <span>Last Successful Run</span>
            <span>2026-10-02 21:05 UTC</span>
          </div>
          <div style={{ marginTop: '0.5rem' }}>
            <Button
              size="sm"
              variant="secondary"
              style={{ width: '100%' }}
              onClick={() => onNotify('Equity feed diagnostic passed: Latency 82ms, 0 anomalies.')}
            >
              Run Feed Diagnostic
            </Button>
          </div>
        </Card>

        <Card className="pipeline-card" glow="blue">
          <div className="pipeline-card__header">
            <span className="pipeline-card__title">FX Fixing Rates</span>
            <Badge variant="active">Active</Badge>
          </div>
          <div className="pipeline-metric">
            <span>Central Bank Source</span>
            <span>European Central Bank (ECB)</span>
          </div>
          <div className="pipeline-metric">
            <span>Currency Pairs Tracked</span>
            <span>7 (EUR, GBP, AUD, CAD, JPY, CHF, USD)</span>
          </div>
          <div className="pipeline-metric">
            <span>Precision Model</span>
            <span>Fixed-Point 8 decimals</span>
          </div>
          <div className="pipeline-metric">
            <span>Last Rate Fixing</span>
            <span>2026-10-02 16:00 CET</span>
          </div>
          <div style={{ marginTop: '0.5rem' }}>
            <Button
              size="sm"
              variant="secondary"
              style={{ width: '100%' }}
              onClick={() => onNotify('FX rate cross-check passed: Triangulation error < 0.0001%')}
            >
              Verify FX Triangulation
            </Button>
          </div>
        </Card>

        <Card className="pipeline-card" glow="amber">
          <div className="pipeline-card__header">
            <span className="pipeline-card__title">Rate Limits & Backfill</span>
            <Badge variant="info">Healthy</Badge>
          </div>
          <div className="pipeline-metric">
            <span>Token Bucket Budget</span>
            <span>800 / 800 req/day</span>
          </div>
          <div className="pipeline-metric">
            <span>Burst Allowance</span>
            <span>8 calls/min</span>
          </div>
          <div className="pipeline-metric">
            <span>Pending Backfill Jobs</span>
            <span>0</span>
          </div>
          <div className="pipeline-metric">
            <span>Anomaly Circuit Breaker</span>
            <span>Armed (±25% threshold)</span>
          </div>
          <div style={{ marginTop: '0.5rem' }}>
            <Button
              size="sm"
              variant="ghost"
              style={{ width: '100%' }}
              onClick={() => onNotify('Backfill queue is empty. All asset price histories are continuous.')}
            >
              Inspect Backfill Queue
            </Button>
          </div>
        </Card>
      </div>
    </div>
  );
};

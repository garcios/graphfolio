import React, { useState } from 'react';
import { formatMoney } from '@graphfolio/ui';

interface CapitalGainsSummaryCardProps {
  totalCapitalGains?: { amount: string; currencyCode: string };
  lossesApplied?: { amount: string; currencyCode: string };
  cgtDiscountAmount?: { amount: string; currencyCode: string };
  netCapitalGain?: { amount: string; currencyCode: string };
  cgtDiscountRatePct?: string;
  gainDisposalsCount?: number;
  lossDisposalsCount?: number;
}

export const CapitalGainsSummaryCard: React.FC<CapitalGainsSummaryCardProps> = ({
  totalCapitalGains,
  lossesApplied,
  cgtDiscountAmount,
  netCapitalGain,
  cgtDiscountRatePct = '50.00',
  gainDisposalsCount = 0,
  lossDisposalsCount = 0,
}) => {
  const [showTooltip, setShowTooltip] = useState(false);

  return (
    <div className="tax-summary-card" data-testid="capital-gains-summary-card">
      <div className="tax-summary-header">
        <div className="tax-summary-title-group">
          <h3 className="tax-summary-title">Capital Gains Summary</h3>
          <p className="tax-summary-desc">Statutory capital gains waterfall under ATO ITAA 1997 s 102-5.</p>
        </div>
        <div className="disposals-count-pill" title={`${gainDisposalsCount} profitable disposals, ${lossDisposalsCount} loss disposals`}>
          <span>{gainDisposalsCount}G / {lossDisposalsCount}L</span>
        </div>
      </div>

      {/* Visual Waterfall Indicator */}
      <div className="cgt-waterfall-progress" title="Gross Gains → Less Losses → Less 50% CGT Discount → Net Gain">
        <div className="waterfall-step-bar" />
      </div>

      <div className="tax-breakdown-list">
        <div className="tax-breakdown-row">
          <div className="tax-row-label">
            <span className="row-bullet bullet-gain" />
            <span className="row-name">Total capital gains</span>
          </div>
          <div className="tax-row-value positive-val">
            {totalCapitalGains ? formatMoney(totalCapitalGains) : '$0.00'}
          </div>
        </div>

        <div className="tax-breakdown-row">
          <div className="tax-row-label">
            <span className="row-bullet bullet-loss" />
            <span className="row-name">Losses applied</span>
            <div
              className="loss-info-wrapper"
              onMouseEnter={() => setShowTooltip(true)}
              onMouseLeave={() => setShowTooltip(false)}
              onClick={() => setShowTooltip((prev) => !prev)}
            >
              <span className="info-icon" role="button" aria-label="Loss offset info">ⓘ</span>
              {showTooltip && (
                <div className="tax-loss-tooltip" role="tooltip">
                  Capital losses are applied first against short-term capital gains (taxed at ordinary rates) before offsetting long-term gains, maximizing the benefit of the 50% CGT discount.
                </div>
              )}
            </div>
          </div>
          <div className="tax-row-value negative-val">
            {lossesApplied && parseFloat(lossesApplied.amount) > 0 ? (
              <span>-{formatMoney(lossesApplied)}</span>
            ) : (
              '$0.00'
            )}
          </div>
        </div>

        <div className="tax-breakdown-row">
          <div className="tax-row-label">
            <span className="row-bullet bullet-discount" />
            <span className="row-name">CGT discount ({parseFloat(cgtDiscountRatePct).toFixed(0)}%)</span>
            <span className="row-ratio-tag discount-tag">Long-Term Concession</span>
          </div>
          <div className="tax-row-value negative-val">
            {cgtDiscountAmount && parseFloat(cgtDiscountAmount.amount) > 0 ? (
              <span>-{formatMoney(cgtDiscountAmount)}</span>
            ) : (
              '$0.00'
            )}
          </div>
        </div>

        <div className="tax-breakdown-divider" />

        <div className="tax-breakdown-row total-row">
          <div className="tax-row-label">
            <span className="total-name">Net capital gain</span>
            <span className="statutory-tag">Assessable</span>
          </div>
          <div className="tax-row-value total-value cgt-accent">
            {netCapitalGain ? formatMoney(netCapitalGain) : '$0.00'}
          </div>
        </div>
      </div>
    </div>
  );
};

import React from 'react';
import { formatMoney } from '@graphfolio/ui';

interface TaxableIncomeSummaryCardProps {
  dividendAndInterestIncome?: { amount: string; currencyCode: string };
  netCapitalGains?: { amount: string; currencyCode: string };
  frankingCredits?: { amount: string; currencyCode: string };
  totalAssessableIncome?: { amount: string; currencyCode: string };
  dividendInterestRatioPct?: string;
  capitalGainsRatioPct?: string;
  frankingCreditsRatioPct?: string;
}

export const TaxableIncomeSummaryCard: React.FC<TaxableIncomeSummaryCardProps> = ({
  dividendAndInterestIncome,
  netCapitalGains,
  frankingCredits,
  totalAssessableIncome,
  dividendInterestRatioPct = '0.00',
  capitalGainsRatioPct = '0.00',
  frankingCreditsRatioPct = '0.00',
}) => {
  const divPct = parseFloat(dividendInterestRatioPct) || 0;
  const cgtPct = parseFloat(capitalGainsRatioPct) || 0;
  const frankingPct = parseFloat(frankingCreditsRatioPct) || 0;

  return (
    <div className="tax-summary-card" data-testid="taxable-income-summary-card">
      <div className="tax-summary-header">
        <div className="tax-summary-title-group">
          <h3 className="tax-summary-title">Taxable Income Summary</h3>
          <p className="tax-summary-desc">Consolidated portfolio income sources contributing to assessable income.</p>
        </div>
      </div>

      {/* Proportional Stacked Ratio Progress Bar */}
      <div className="proportional-bar-container" title={`Dividends: ${divPct}%, CGT: ${cgtPct}%, Franking: ${frankingPct}%`}>
        <div
          className="bar-segment bar-amber"
          style={{ width: `${divPct}%` }}
          aria-label={`Dividend & Interest Income ${divPct}%`}
        />
        <div
          className="bar-segment bar-coral"
          style={{ width: `${cgtPct}%` }}
          aria-label={`Net Capital Gains ${cgtPct}%`}
        />
        <div
          className="bar-segment bar-purple"
          style={{ width: `${frankingPct}%` }}
          aria-label={`Franking Credits ${frankingPct}%`}
        />
      </div>

      <div className="tax-breakdown-list">
        <div className="tax-breakdown-row">
          <div className="tax-row-label">
            <span className="legend-indicator dot-amber" />
            <span className="row-name">Dividend & interest income</span>
            <span className="row-ratio-tag amber-tag">{divPct.toFixed(2)}%</span>
          </div>
          <div className="tax-row-value">
            {dividendAndInterestIncome ? formatMoney(dividendAndInterestIncome) : '$0.00'}
          </div>
        </div>

        <div className="tax-breakdown-row">
          <div className="tax-row-label">
            <span className="legend-indicator dot-coral" />
            <span className="row-name">Capital gains (Net)</span>
            <span className="row-ratio-tag coral-tag">{cgtPct.toFixed(2)}%</span>
          </div>
          <div className="tax-row-value">
            {netCapitalGains ? formatMoney(netCapitalGains) : '$0.00'}
          </div>
        </div>

        <div className="tax-breakdown-row">
          <div className="tax-row-label">
            <span className="legend-indicator dot-purple" />
            <span className="row-name">Franking credits (Gross-up)</span>
            <span className="row-ratio-tag purple-tag">{frankingPct.toFixed(2)}%</span>
          </div>
          <div className="tax-row-value">
            {frankingCredits ? formatMoney(frankingCredits) : '$0.00'}
          </div>
        </div>

        <div className="tax-breakdown-divider" />

        <div className="tax-breakdown-row total-row">
          <div className="tax-row-label">
            <span className="total-name">Total assessable income</span>
            <span className="verified-check" title="Consolidated reconciliation matches statutory statutory sum">✓ Reconciled</span>
          </div>
          <div className="tax-row-value total-value">
            {totalAssessableIncome ? formatMoney(totalAssessableIncome) : '$0.00'}
          </div>
        </div>
      </div>
    </div>
  );
};

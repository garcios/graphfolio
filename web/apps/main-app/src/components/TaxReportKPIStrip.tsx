import React from 'react';
import { formatMoney } from '@graphfolio/ui';

interface TaxReportKPIStripProps {
  totalAssessableIncome?: { amount: string; currencyCode: string };
  netCapitalGain?: { amount: string; currencyCode: string };
  frankingCredits?: { amount: string; currencyCode: string };
  foreignTaxOffsets?: { amount: string; currencyCode: string };
  fiscalYearLabel?: string;
}

export const TaxReportKPIStrip: React.FC<TaxReportKPIStripProps> = ({
  totalAssessableIncome,
  netCapitalGain,
  frankingCredits,
  foreignTaxOffsets,
  fiscalYearLabel,
}) => {
  return (
    <div className="tax-kpi-strip" data-testid="tax-kpi-strip">
      <div className="tax-kpi-card highlight-card">
        <div className="tax-kpi-header">
          <span className="tax-kpi-title">Total Assessable Income</span>
          {fiscalYearLabel && <span className="tax-fy-pill">{fiscalYearLabel}</span>}
        </div>
        <div className="tax-kpi-value primary-value">
          {totalAssessableIncome ? formatMoney(totalAssessableIncome) : '—'}
        </div>
        <div className="tax-kpi-subtext">
          Statutory declared gross investment income
        </div>
      </div>

      <div className="tax-kpi-card">
        <div className="tax-kpi-header">
          <span className="tax-kpi-title">Net Capital Gain</span>
          <span className="tax-badge-pill cgt-pill">CGT</span>
        </div>
        <div className="tax-kpi-value cgt-value">
          {netCapitalGain ? formatMoney(netCapitalGain) : '—'}
        </div>
        <div className="tax-kpi-subtext">
          After losses & 50% discount concession
        </div>
      </div>

      <div className="tax-kpi-card">
        <div className="tax-kpi-header">
          <span className="tax-kpi-title">Franking Credits</span>
          <span className="tax-badge-pill franking-pill">Imputation</span>
        </div>
        <div className="tax-kpi-value franking-value">
          {frankingCredits ? formatMoney(frankingCredits) : '—'}
        </div>
        <div className="tax-kpi-subtext">
          Refundable corporate tax offset
        </div>
      </div>

      <div className="tax-kpi-card">
        <div className="tax-kpi-header">
          <span className="tax-kpi-title">Foreign Tax Offsets (FITO)</span>
          <span className="tax-badge-pill fito-pill">Withholding</span>
        </div>
        <div className="tax-kpi-value fito-value">
          {foreignTaxOffsets ? formatMoney(foreignTaxOffsets) : '—'}
        </div>
        <div className="tax-kpi-subtext">
          Foreign withholding tax credit offset
        </div>
      </div>
    </div>
  );
};

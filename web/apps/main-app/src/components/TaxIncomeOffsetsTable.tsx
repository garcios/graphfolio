import React from 'react';
import { formatMoney } from '@graphfolio/ui';

interface TaxIncomeOffsetsTableProps {
  taxableIncome?: {
    dividendAndInterestIncome?: { amount: string; currencyCode: string };
    netCapitalGains?: { amount: string; currencyCode: string };
    frankingCredits?: { amount: string; currencyCode: string };
    totalAssessableIncome?: { amount: string; currencyCode: string };
  };
  taxOffsets?: {
    frankingCreditsOffset?: { amount: string; currencyCode: string };
    foreignTaxOffsets?: { amount: string; currencyCode: string };
    totalAvailableOffsets?: { amount: string; currencyCode: string };
  };
  fiscalYearLabel?: string;
}

export const TaxIncomeOffsetsTable: React.FC<TaxIncomeOffsetsTableProps> = ({
  taxableIncome,
  taxOffsets,
  fiscalYearLabel = 'FY',
}) => {
  return (
    <div className="tax-subview-panel" data-testid="tax-income-offsets-table">
      <div className="tax-panel-header">
        <div className="tax-panel-title-group">
          <h3 className="tax-panel-title">Income & Tax Offsets Reconciliation</h3>
          <p className="tax-panel-desc">
            Consolidated investment income streams and attached statutory tax offsets for {fiscalYearLabel}.
          </p>
        </div>
      </div>

      <div className="tax-offsets-grid">
        {/* Income Stream Cards */}
        <div className="offset-card income-summary-box">
          <h4 className="offset-card-title">Assessable Investment Income</h4>
          <p className="offset-card-desc">Income items that must be declared in your tax return as assessable income.</p>

          <div className="offset-items-list">
            <div className="offset-item-row">
              <span className="offset-item-label">Cash Dividends & Interest</span>
              <span className="offset-item-value">
                {taxableIncome?.dividendAndInterestIncome
                  ? formatMoney(taxableIncome.dividendAndInterestIncome)
                  : '$0.00'}
              </span>
            </div>
            <div className="offset-item-row">
              <span className="offset-item-label">Franking Credits (Gross-up)</span>
              <span className="offset-item-value">
                {taxableIncome?.frankingCredits
                  ? formatMoney(taxableIncome.frankingCredits)
                  : '$0.00'}
              </span>
            </div>
            <div className="offset-item-row">
              <span className="offset-item-label">Net Realized Capital Gains</span>
              <span className="offset-item-value">
                {taxableIncome?.netCapitalGains
                  ? formatMoney(taxableIncome.netCapitalGains)
                  : '$0.00'}
              </span>
            </div>
            <div className="tax-breakdown-divider" />
            <div className="offset-item-row total-row">
              <span className="offset-item-label font-bold">Total Assessable Income</span>
              <span className="offset-item-value total-value">
                {taxableIncome?.totalAssessableIncome
                  ? formatMoney(taxableIncome.totalAssessableIncome)
                  : '$0.00'}
              </span>
            </div>
          </div>
        </div>

        {/* Tax Offsets Cards */}
        <div className="offset-card credits-summary-box">
          <h4 className="offset-card-title">Available Tax Offsets & Credits</h4>
          <p className="offset-card-desc">Dollar-for-dollar tax offsets that reduce net tax payable to the revenue authority.</p>

          <div className="offset-items-list">
            <div className="offset-item-row">
              <div>
                <span className="offset-item-label font-bold">Franking Tax Offset</span>
                <div className="offset-subtext">Refundable tax offset from Australian franked company dividends</div>
              </div>
              <span className="offset-item-value franking-text">
                {taxOffsets?.frankingCreditsOffset
                  ? formatMoney(taxOffsets.frankingCreditsOffset)
                  : '$0.00'}
              </span>
            </div>

            <div className="offset-item-row">
              <div>
                <span className="offset-item-label font-bold">Foreign Tax Offset (FITO)</span>
                <div className="offset-subtext">Non-refundable foreign withholding taxes paid at source</div>
              </div>
              <span className="offset-item-value fito-text">
                {taxOffsets?.foreignTaxOffsets
                  ? formatMoney(taxOffsets.foreignTaxOffsets)
                  : '$0.00'}
              </span>
            </div>

            <div className="tax-breakdown-divider" />

            <div className="offset-item-row total-row">
              <span className="offset-item-label font-bold">Total Available Tax Offsets</span>
              <span className="offset-item-value total-value">
                {taxOffsets?.totalAvailableOffsets
                  ? formatMoney(taxOffsets.totalAvailableOffsets)
                  : '$0.00'}
              </span>
            </div>
          </div>
        </div>
      </div>

      <div className="tax-guidance-box">
        <h5 className="guidance-title">Statutory Tax Filing Note</h5>
        <p className="guidance-body">
          Under statutory imputation systems (e.g. Australian ITAA 1997 Division 207), franking credits must be added to your assessable income (imputation gross-up) and simultaneously claimed as a tax offset. If your franking credit offset exceeds tax payable, the excess is refunded directly to individual taxpayers. Foreign withholding taxes (FITO) are non-refundable offsets and cannot create a cash refund.
        </p>
      </div>
    </div>
  );
};

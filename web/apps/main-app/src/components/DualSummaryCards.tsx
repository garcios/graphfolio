import React from 'react';
import { TaxableIncomeSummaryCard } from './TaxableIncomeSummaryCard';
import { CapitalGainsSummaryCard } from './CapitalGainsSummaryCard';

interface DualSummaryCardsProps {
  summary?: {
    taxableIncome?: {
      dividendAndInterestIncome?: { amount: string; currencyCode: string };
      netCapitalGains?: { amount: string; currencyCode: string };
      frankingCredits?: { amount: string; currencyCode: string };
      totalAssessableIncome?: { amount: string; currencyCode: string };
      dividendInterestRatioPct?: string;
      capitalGainsRatioPct?: string;
      frankingCreditsRatioPct?: string;
    };
    capitalGains?: {
      totalCapitalGains?: { amount: string; currencyCode: string };
      lossesApplied?: { amount: string; currencyCode: string };
      cgtDiscountAmount?: { amount: string; currencyCode: string };
      netCapitalGain?: { amount: string; currencyCode: string };
      cgtDiscountRatePct?: string;
      gainDisposalsCount?: number;
      lossDisposalsCount?: number;
    };
  };
}

export const DualSummaryCards: React.FC<DualSummaryCardsProps> = ({ summary }) => {
  return (
    <div className="dual-summary-cards-container" data-testid="dual-summary-cards">
      <TaxableIncomeSummaryCard
        dividendAndInterestIncome={summary?.taxableIncome?.dividendAndInterestIncome}
        netCapitalGains={summary?.taxableIncome?.netCapitalGains}
        frankingCredits={summary?.taxableIncome?.frankingCredits}
        totalAssessableIncome={summary?.taxableIncome?.totalAssessableIncome}
        dividendInterestRatioPct={summary?.taxableIncome?.dividendInterestRatioPct}
        capitalGainsRatioPct={summary?.taxableIncome?.capitalGainsRatioPct}
        frankingCreditsRatioPct={summary?.taxableIncome?.frankingCreditsRatioPct}
      />
      <CapitalGainsSummaryCard
        totalCapitalGains={summary?.capitalGains?.totalCapitalGains}
        lossesApplied={summary?.capitalGains?.lossesApplied}
        cgtDiscountAmount={summary?.capitalGains?.cgtDiscountAmount}
        netCapitalGain={summary?.capitalGains?.netCapitalGain}
        cgtDiscountRatePct={summary?.capitalGains?.cgtDiscountRatePct}
        gainDisposalsCount={summary?.capitalGains?.gainDisposalsCount}
        lossDisposalsCount={summary?.capitalGains?.lossDisposalsCount}
      />
    </div>
  );
};

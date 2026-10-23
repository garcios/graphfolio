import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { client } from '@graphfolio/api-client';
import { calculateFiscalYearBounds } from './UserPreferencesModal';
import { TaxReportKPIStrip } from './TaxReportKPIStrip';
import { DualSummaryCards } from './DualSummaryCards';
import { RealizedDisposalsTable, type RealizedDisposalItem } from './RealizedDisposalsTable';
import { OpenTaxLotsTable, type TaxLotItem } from './OpenTaxLotsTable';
import { TaxIncomeOffsetsTable } from './TaxIncomeOffsetsTable';
import './TaxLotsReport.css';

export type TaxSubView = 'summary' | 'disposals' | 'open-lots' | 'income';

export interface TaxLotsReportProps {
  preferredCurrency?: string;
  financialYearEndMonth?: number;
  financialYearEndDay?: number;
}

export const TaxLotsReport: React.FC<TaxLotsReportProps> = ({
  financialYearEndMonth,
  financialYearEndDay,
}) => {
  const [activeTab, setActiveTab] = useState<TaxSubView>('summary');
  const [fyEndMonth, setFyEndMonth] = useState<number>(financialYearEndMonth ?? 6);
  const [fyEndDay, setFyEndDay] = useState<number>(financialYearEndDay ?? 30);

  // Fetch preferences if not passed as props
  useEffect(() => {
    if (financialYearEndMonth !== undefined && financialYearEndDay !== undefined) {
      setFyEndMonth(financialYearEndMonth);
      setFyEndDay(financialYearEndDay);
    } else {
      client
        .query({
          userPreferences: {
            financialYearEndMonth: true,
            financialYearEndDay: true,
          },
        })
        .then((res: any) => {
          if (res?.userPreferences) {
            if (res.userPreferences.financialYearEndMonth != null) {
              setFyEndMonth(res.userPreferences.financialYearEndMonth);
            }
            if (res.userPreferences.financialYearEndDay != null) {
              setFyEndDay(res.userPreferences.financialYearEndDay);
            }
          }
        })
        .catch(() => {});
    }
  }, [financialYearEndMonth, financialYearEndDay]);

  // Determine current FY bounds and year options
  const currentBounds = useMemo(() => {
    return calculateFiscalYearBounds(fyEndMonth, fyEndDay);
  }, [fyEndMonth, fyEndDay]);

  const currentFYYear = currentBounds.endDate.getFullYear();
  const [selectedYear, setSelectedYear] = useState<number>(currentFYYear);

  // Keep selectedYear synced when currentBounds initializes
  useEffect(() => {
    if (currentFYYear) {
      setSelectedYear(currentFYYear);
    }
  }, [currentFYYear]);

  // Generate selectable financial years (current down to 5 years back)
  const fyOptions = useMemo(() => {
    const list: { year: number; label: string }[] = [];
    const base = currentFYYear || new Date().getFullYear();
    for (let i = 0; i < 5; i++) {
      const y = base - i;
      const label = fyEndMonth === 12 && fyEndDay === 31 ? `FY${y}` : `FY${String(y).slice(-2)}`;
      list.push({ year: y, label });
    }
    return list;
  }, [currentFYYear, fyEndMonth, fyEndDay]);

  // Report state
  const [reportData, setReportData] = useState<any>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [isExporting, setIsExporting] = useState<boolean>(false);

  const fetchReport = useCallback(() => {
    setLoading(true);
    setError(null);

    client
      .query({
        taxReport: {
          __args: {
            financialYear: selectedYear,
          },
          summary: {
            fiscalYearLabel: true,
            baseCurrency: true,
            taxableIncome: {
              dividendAndInterestIncome: { amount: true, currencyCode: true },
              netCapitalGains: { amount: true, currencyCode: true },
              frankingCredits: { amount: true, currencyCode: true },
              totalAssessableIncome: { amount: true, currencyCode: true },
              dividendInterestRatioPct: true,
              capitalGainsRatioPct: true,
              frankingCreditsRatioPct: true,
            },
            capitalGains: {
              totalCapitalGains: { amount: true, currencyCode: true },
              lossesApplied: { amount: true, currencyCode: true },
              cgtDiscountAmount: { amount: true, currencyCode: true },
              netCapitalGain: { amount: true, currencyCode: true },
              cgtDiscountRatePct: true,
              gainDisposalsCount: true,
              lossDisposalsCount: true,
            },
            taxOffsets: {
              frankingCreditsOffset: { amount: true, currencyCode: true },
              foreignTaxOffsets: { amount: true, currencyCode: true },
              totalAvailableOffsets: { amount: true, currencyCode: true },
            },
          },
          disposals: {
            disposalId: true,
            taxLotId: true,
            sellTransactionId: true,
            instrumentId: true,
            ticker: true,
            name: true,
            acquiredDate: true,
            disposedDate: true,
            quantityDisposed: true,
            costBasisRelievedBase: { amount: true, currencyCode: true },
            proceedsBase: { amount: true, currencyCode: true },
            realizedPnlBase: { amount: true, currencyCode: true },
            realizedPnlPercent: true,
            holdingPeriodDays: true,
            termClassification: true,
            isCgtDiscountEligible: true,
            discountedTaxablePnlBase: { amount: true, currencyCode: true },
          },
          openLots: {
            id: true,
            instrumentId: true,
            ticker: true,
            name: true,
            openTransactionId: true,
            acquiredDate: true,
            closedDate: true,
            originalQuantity: true,
            remainingQuantity: true,
            costBasis: { amount: true, currencyCode: true },
            costBasisBase: { amount: true, currencyCode: true },
            unitCostBase: { amount: true, currencyCode: true },
            currentPrice: { amount: true, currencyCode: true },
            marketValueBase: { amount: true, currencyCode: true },
            unrealizedPnlBase: { amount: true, currencyCode: true },
            unrealizedPnlPercent: true,
            holdingPeriodDays: true,
            termClassification: true,
            isCgtDiscountEligible: true,
            isLossHarvestingCandidate: true,
            hasWashSaleRisk: true,
          },
          fromDate: true,
          toDate: true,
          fiscalYearLabel: true,
          baseCurrency: true,
        },
      })
      .then((res: any) => {
        if (res?.taxReport) {
          setReportData(res.taxReport);
        } else {
          setReportData(null);
        }
      })
      .catch((err: any) => {
        setError(err.message || 'Failed to load tax report data.');
      })
      .finally(() => {
        setLoading(false);
      });
  }, [selectedYear]);

  useEffect(() => {
    fetchReport();
  }, [fetchReport]);

  // CSV Exporter
  const handleExportCSV = async () => {
    try {
      setIsExporting(true);
      const res: any = await client.query({
        exportTaxScheduleCSV: {
          __args: {
            financialYear: selectedYear,
          },
        },
      });

      if (res?.exportTaxScheduleCSV) {
        const blob = new Blob([res.exportTaxScheduleCSV], { type: 'text/csv;charset=utf-8;' });
        const url = URL.createObjectURL(blob);
        const link = document.createElement('a');
        link.href = url;
        const fyLabel = reportData?.fiscalYearLabel || `FY${selectedYear}`;
        link.setAttribute('download', `tax_schedule_${fyLabel}.csv`);
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        URL.revokeObjectURL(url);
      }
    } catch (err: any) {
      alert(`Export failed: ${err.message || 'Unknown error'}`);
    } finally {
      setIsExporting(false);
    }
  };

  const summary = reportData?.summary;
  const disposals: RealizedDisposalItem[] = reportData?.disposals || [];
  const openLots: TaxLotItem[] = reportData?.openLots || [];

  const tlhCandidateCount = useMemo(() => {
    return openLots.filter((l) => l.isLossHarvestingCandidate).length;
  }, [openLots]);

  return (
    <div className="tax-lots-report" data-testid="tax-lots-report">
      {/* Header and Controls */}
      <div className="tax-report-header">
        <div className="tax-title-group">
          <h2>Tax Lots & Capital Gains Report</h2>
          <p>
            Statutory tax calculations, realized capital gains waterfall, loss harvesting scanner, and Australian CGT schedule.
          </p>
        </div>

        <div className="tax-controls-bar">
          <select
            className="tax-select"
            value={selectedYear}
            onChange={(e) => setSelectedYear(Number(e.target.value))}
            aria-label="Financial Year"
          >
            {fyOptions.map((opt) => (
              <option key={opt.year} value={opt.year}>
                {opt.label} ({opt.year})
              </option>
            ))}
          </select>

          <button
            type="button"
            className="tax-btn-secondary"
            onClick={fetchReport}
            disabled={loading}
            title="Reload report data"
          >
            ↻ Refresh
          </button>
        </div>
      </div>

      {/* Error Banner */}
      {error && (
        <div className="tax-error-banner">
          <span>⚠️ {error}</span>
          <button type="button" onClick={fetchReport} className="tax-btn-retry">
            Retry
          </button>
        </div>
      )}

      {/* Loading Skeleton / Spinner */}
      {loading && (
        <div className="tax-loading-state">
          <div className="tax-spinner" />
          <p>Computing tax lots, applying loss allocation rules, and compiling CGT schedule...</p>
        </div>
      )}

      {/* Main Content Area */}
      {!loading && (
        <>
          {/* Top 4 KPI Metric Cards */}
          <TaxReportKPIStrip
            totalAssessableIncome={summary?.taxableIncome?.totalAssessableIncome}
            netCapitalGain={summary?.capitalGains?.netCapitalGain}
            frankingCredits={summary?.taxableIncome?.frankingCredits}
            foreignTaxOffsets={summary?.taxOffsets?.foreignTaxOffsets}
            fiscalYearLabel={summary?.fiscalYearLabel || reportData?.fiscalYearLabel}
          />

          {/* Sub-Navigation Pills */}
          <div className="tax-subnav-pills" role="tablist">
            <button
              type="button"
              role="tab"
              aria-selected={activeTab === 'summary'}
              className={`tax-subnav-btn ${activeTab === 'summary' ? 'active' : ''}`}
              onClick={() => setActiveTab('summary')}
            >
              <span>📊 Tax Return Summary</span>
            </button>

            <button
              type="button"
              role="tab"
              aria-selected={activeTab === 'disposals'}
              className={`tax-subnav-btn ${activeTab === 'disposals' ? 'active' : ''}`}
              onClick={() => setActiveTab('disposals')}
            >
              <span>📑 Realized Disposals</span>
              {disposals.length > 0 && <span className="tab-badge">{disposals.length}</span>}
            </button>

            <button
              type="button"
              role="tab"
              aria-selected={activeTab === 'open-lots'}
              className={`tax-subnav-btn ${activeTab === 'open-lots' ? 'active' : ''}`}
              onClick={() => setActiveTab('open-lots')}
            >
              <span>📦 Open Tax Lots & TLH</span>
              {tlhCandidateCount > 0 ? (
                <span className="tab-badge pulse-badge" title={`${tlhCandidateCount} TLH candidates`}>
                  {tlhCandidateCount} TLH
                </span>
              ) : (
                openLots.length > 0 && <span className="tab-badge">{openLots.length}</span>
              )}
            </button>

            <button
              type="button"
              role="tab"
              aria-selected={activeTab === 'income'}
              className={`tax-subnav-btn ${activeTab === 'income' ? 'active' : ''}`}
              onClick={() => setActiveTab('income')}
            >
              <span>💰 Income & Offsets</span>
            </button>
          </div>

          {/* Tab Sub-View Renderers */}
          {activeTab === 'summary' && (
            <div className="tax-summary-view-content">
              <DualSummaryCards summary={summary} />
              <div className="tax-summary-quick-preview">
                <RealizedDisposalsTable
                  disposals={disposals.slice(0, 5)}
                  onExportCSV={handleExportCSV}
                  isExporting={isExporting}
                />
              </div>
            </div>
          )}

          {activeTab === 'disposals' && (
            <RealizedDisposalsTable
              disposals={disposals}
              onExportCSV={handleExportCSV}
              isExporting={isExporting}
            />
          )}

          {activeTab === 'open-lots' && <OpenTaxLotsTable lots={openLots} />}

          {activeTab === 'income' && (
            <TaxIncomeOffsetsTable
              taxableIncome={summary?.taxableIncome}
              taxOffsets={summary?.taxOffsets}
              fiscalYearLabel={summary?.fiscalYearLabel || reportData?.fiscalYearLabel}
            />
          )}
        </>
      )}
    </div>
  );
};

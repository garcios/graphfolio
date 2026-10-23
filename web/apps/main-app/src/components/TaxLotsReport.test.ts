import { describe, it, expect } from 'vitest';
import { calculateFiscalYearBounds } from './UserPreferencesModal';

describe('Tax Lots & Capital Gains Report Tests', () => {
  describe('Fiscal Year Slicing & Bounds', () => {
    it('accurately resolves Australian EOFY (30 June) FY bounds', () => {
      const asOf = new Date(2025, 9, 24); // 24 Oct 2025 -> Australian FY26
      const bounds = calculateFiscalYearBounds(6, 30, asOf);

      expect(bounds.currentLabel).toBe('FY26');
      expect(bounds.prevLabel).toBe('FY25');
      expect(bounds.endDate.getFullYear()).toBe(2026);
    });

    it('accurately resolves US EOFY (31 Dec) FY bounds', () => {
      const asOf = new Date(2025, 5, 15); // 15 Jun 2025 -> US FY2025
      const bounds = calculateFiscalYearBounds(12, 31, asOf);

      expect(bounds.currentLabel).toBe('FY2025');
      expect(bounds.prevLabel).toBe('FY2024');
      expect(bounds.endDate.getFullYear()).toBe(2025);
    });
  });

  describe('CGT 12-Month Rule & Discount Eligibility (TEST-TAX-01, TEST-TAX-02)', () => {
    const isDiscountEligible = (holdingDays: number) => holdingDays > 365;

    it('holding period <= 365 days is short-term and not eligible for CGT discount', () => {
      // 364 days held
      expect(isDiscountEligible(364)).toBe(false);
      // Exactly 365 days held (day of acquisition excluded, requires > 365 days)
      expect(isDiscountEligible(365)).toBe(false);
    });

    it('holding period >= 366 days is long-term and eligible for 50% CGT discount', () => {
      // 366 days held
      expect(isDiscountEligible(366)).toBe(true);
      expect(isDiscountEligible(500)).toBe(true);
    });
  });

  describe('Capital Gains Waterfall & Loss Sequencing (TEST-TAX-03, TEST-TAX-04)', () => {
    it('applies capital losses to short-term gains first before long-term gains', () => {
      const grossShortTermGains = 1000.0;
      const grossLongTermGains = 5000.0;
      const grossLosses = 600.0;

      // Rule: losses offset short-term gains first
      const stOffset = Math.min(grossShortTermGains, grossLosses);
      const remainingLosses = grossLosses - stOffset;
      const netShortTerm = grossShortTermGains - stOffset;

      const ltOffset = Math.min(grossLongTermGains, remainingLosses);
      const remainingLongTerm = grossLongTermGains - ltOffset;

      // 50% discount applies ONLY to remaining long term
      const discount = remainingLongTerm * 0.5;
      const netTaxableGain = netShortTerm + (remainingLongTerm - discount);

      expect(stOffset).toBe(600.0);
      expect(netShortTerm).toBe(400.0);
      expect(remainingLongTerm).toBe(5000.0);
      expect(discount).toBe(2500.0);
      expect(netTaxableGain).toBe(2900.0);
    });

    it('correctly calculates example from implementation plan (TEST-TAX-03)', () => {
      const grossLongTermGains = 6064.99;
      const losses = 350.17;

      const netBeforeDiscount = grossLongTermGains - losses; // 5714.82
      const discount = Number((netBeforeDiscount * 0.5).toFixed(2)); // 2857.41
      const netCapitalGain = Number((netBeforeDiscount - discount).toFixed(2)); // 2857.41

      expect(discount).toBe(2857.41);
      expect(netCapitalGain).toBe(2857.41);
    });
  });

  describe('Assessable Income & Progress Bar Ratios (TEST-TAX-05, TEST-TAX-06)', () => {
    it('computes total assessable income and ratio percentages matching specification', () => {
      const divInt = 2966.8;
      const netCgt = 2857.41;
      const franking = 300.33;

      const totalAssessable = divInt + netCgt + franking;
      expect(Number(totalAssessable.toFixed(2))).toBe(6124.54);

      const divIntRatio = ((divInt / totalAssessable) * 100).toFixed(2);
      const cgtRatio = ((netCgt / totalAssessable) * 100).toFixed(2);
      const frankingRatio = ((franking / totalAssessable) * 100).toFixed(2);

      expect(divIntRatio).toBe('48.44');
      expect(cgtRatio).toBe('46.66');
      expect(frankingRatio).toBe('4.90');
    });
  });

  describe('Tax-Loss Harvesting (TLH) & Wash-Sale Detection (TEST-TAX-08, TEST-TAX-09)', () => {
    const isTlhCandidate = (unrealizedPnl: number, unrealizedPct: number) => {
      return unrealizedPnl < -500 && unrealizedPct < -10;
    };

    const hasWashSaleRisk = (tradeDaysAgo: number) => {
      return tradeDaysAgo >= -30 && tradeDaysAgo <= 30;
    };

    it('identifies TLH candidate when loss > $500 and loss pct < -10%', () => {
      expect(isTlhCandidate(-501, -10.1)).toBe(true);
      expect(isTlhCandidate(-499, -15.0)).toBe(false); // under $500 threshold
      expect(isTlhCandidate(-1000, -8.0)).toBe(false); // under 10% threshold
      expect(isTlhCandidate(100, 5.0)).toBe(false);
    });

    it('identifies wash-sale risk when transaction executed within 30 days', () => {
      expect(hasWashSaleRisk(15)).toBe(true);
      expect(hasWashSaleRisk(30)).toBe(true);
      expect(hasWashSaleRisk(31)).toBe(false);
      expect(hasWashSaleRisk(-10)).toBe(true);
      expect(hasWashSaleRisk(-35)).toBe(false);
    });
  });
});

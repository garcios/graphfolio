import { describe, it, expect } from 'vitest';
import { calculateFiscalYearBounds } from './UserPreferencesModal';

const formatDateToISO = (d: Date): string => {
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, '0');
  const day = String(d.getDate()).padStart(2, '0');
  return `${y}-${m}-${day}`;
};

describe('Cash Flow Report - Last Financial Year Preset', () => {
  describe('calculateFiscalYearBounds for Last Financial Year', () => {
    it('calculates exact ISO date bounds for Australian EOFY (30 June)', () => {
      // Evaluated on October 22, 2024
      const asOf = new Date(2024, 9, 22);
      const bounds = calculateFiscalYearBounds(6, 30, asOf);

      const fromDate = formatDateToISO(bounds.prevStartDate);
      const toDate = formatDateToISO(bounds.prevEndDate);

      expect(fromDate).toBe('2023-07-01');
      expect(toDate).toBe('2024-06-30');
      expect(bounds.prevLabel).toBe('FY24');
      expect(bounds.prevPeriod).toBe('1 Jul 2023 – 30 Jun 2024 (FY24)');
    });

    it('calculates exact ISO date bounds for Australian EOFY in early calendar year (March 2025)', () => {
      // Evaluated on March 15, 2025 (still in Australian FY25)
      const asOf = new Date(2025, 2, 15);
      const bounds = calculateFiscalYearBounds(6, 30, asOf);

      const fromDate = formatDateToISO(bounds.prevStartDate);
      const toDate = formatDateToISO(bounds.prevEndDate);

      expect(fromDate).toBe('2023-07-01');
      expect(toDate).toBe('2024-06-30');
      expect(bounds.prevLabel).toBe('FY24');
    });

    it('rolls over to next financial year on July 1 (July 1, 2025)', () => {
      // Evaluated on July 1, 2025 (now in Australian FY26, last FY is FY25)
      const asOf = new Date(2025, 6, 1);
      const bounds = calculateFiscalYearBounds(6, 30, asOf);

      const fromDate = formatDateToISO(bounds.prevStartDate);
      const toDate = formatDateToISO(bounds.prevEndDate);

      expect(fromDate).toBe('2024-07-01');
      expect(toDate).toBe('2025-06-30');
      expect(bounds.prevLabel).toBe('FY25');
      expect(bounds.prevPeriod).toBe('1 Jul 2024 – 30 Jun 2025 (FY25)');
    });

    it('calculates exact ISO date bounds for United Kingdom EOFY (5 April)', () => {
      // Evaluated on October 22, 2024
      const asOf = new Date(2024, 9, 22);
      const bounds = calculateFiscalYearBounds(4, 5, asOf);

      const fromDate = formatDateToISO(bounds.prevStartDate);
      const toDate = formatDateToISO(bounds.prevEndDate);

      expect(fromDate).toBe('2023-04-06');
      expect(toDate).toBe('2024-04-05');
      expect(bounds.prevLabel).toBe('FY24');
    });

    it('calculates exact ISO date bounds for Calendar Year EOFY (31 December)', () => {
      // Evaluated on October 22, 2024
      const asOf = new Date(2024, 9, 22);
      const bounds = calculateFiscalYearBounds(12, 31, asOf);

      const fromDate = formatDateToISO(bounds.prevStartDate);
      const toDate = formatDateToISO(bounds.prevEndDate);

      expect(fromDate).toBe('2023-01-01');
      expect(toDate).toBe('2023-12-31');
      expect(bounds.prevLabel).toBe('FY2023');
      expect(bounds.prevPeriod).toBe('1 Jan 2023 – 31 Dec 2023 (FY2023)');
    });
  });

  describe('CSV Export naming for Last Financial Year', () => {
    it('constructs reporting period header and CSV filename with FY label', () => {
      const asOf = new Date(2024, 9, 22);
      const bounds = calculateFiscalYearBounds(6, 30, asOf);
      const reportData = { toDate: '2024-06-30' };

      const periodStr = `Last Financial Year (${bounds.prevLabel}: ${bounds.prevPeriod})`;
      const filename = `GraphFolio_CashFlow_${bounds.prevLabel}_${reportData.toDate}.csv`;

      expect(periodStr).toBe('Last Financial Year (FY24: 1 Jul 2023 – 30 Jun 2024 (FY24))');
      expect(filename).toBe('GraphFolio_CashFlow_FY24_2024-06-30.csv');
    });
  });
});

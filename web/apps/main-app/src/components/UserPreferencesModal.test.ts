import { describe, it, expect } from 'vitest';
import {
  calculateFiscalYearBounds,
  detectPresetId,
  getMaxDaysInMonth,
  EOFY_PRESETS,
} from './UserPreferencesModal';

describe('User Preferences EOFY Configuration (US-PREF-102)', () => {
  describe('calculateFiscalYearBounds', () => {
    it('calculates Australian financial year bounds (30 June)', () => {
      // Evaluated mid-year: October 22, 2024
      const asOf = new Date(2024, 9, 22); // Month index 9 = October
      const bounds = calculateFiscalYearBounds(6, 30, asOf);

      expect(bounds.currentPeriod).toBe('1 Jul 2024 – 30 Jun 2025 (FY25)');
      expect(bounds.prevPeriod).toBe('1 Jul 2023 – 30 Jun 2024 (FY24)');
      expect(bounds.currentLabel).toBe('FY25');
      expect(bounds.prevLabel).toBe('FY24');
    });

    it('calculates Australian financial year bounds on exact EOFY day (30 June 2025)', () => {
      const asOf = new Date(2025, 5, 30, 12, 0, 0); // Month index 5 = June
      const bounds = calculateFiscalYearBounds(6, 30, asOf);

      expect(bounds.currentPeriod).toBe('1 Jul 2024 – 30 Jun 2025 (FY25)');
      expect(bounds.prevPeriod).toBe('1 Jul 2023 – 30 Jun 2024 (FY24)');
    });

    it('rolls over to next financial year on July 1', () => {
      const asOf = new Date(2025, 6, 1); // Month index 6 = July
      const bounds = calculateFiscalYearBounds(6, 30, asOf);

      expect(bounds.currentPeriod).toBe('1 Jul 2025 – 30 Jun 2026 (FY26)');
      expect(bounds.prevPeriod).toBe('1 Jul 2024 – 30 Jun 2025 (FY25)');
      expect(bounds.currentLabel).toBe('FY26');
    });

    it('calculates United Kingdom financial year bounds (5 April)', () => {
      // Evaluated on October 22, 2024
      const asOf = new Date(2024, 9, 22);
      const bounds = calculateFiscalYearBounds(4, 5, asOf);

      expect(bounds.currentPeriod).toBe('6 Apr 2024 – 5 Apr 2025 (FY25)');
      expect(bounds.prevPeriod).toBe('6 Apr 2023 – 5 Apr 2024 (FY24)');
      expect(bounds.currentLabel).toBe('FY25');
    });

    it('calculates Calendar Year financial bounds (31 December)', () => {
      // Evaluated on October 22, 2024
      const asOf = new Date(2024, 9, 22);
      const bounds = calculateFiscalYearBounds(12, 31, asOf);

      expect(bounds.currentPeriod).toBe('1 Jan 2024 – 31 Dec 2024 (FY2024)');
      expect(bounds.prevPeriod).toBe('1 Jan 2023 – 31 Dec 2023 (FY2023)');
      expect(bounds.currentLabel).toBe('FY2024');
      expect(bounds.prevLabel).toBe('FY2023');
    });

    it('calculates New Zealand / India financial year bounds (31 March)', () => {
      const asOf = new Date(2024, 9, 22);
      const bounds = calculateFiscalYearBounds(3, 31, asOf);

      expect(bounds.currentPeriod).toBe('1 Apr 2024 – 31 Mar 2025 (FY25)');
      expect(bounds.prevPeriod).toBe('1 Apr 2023 – 31 Mar 2024 (FY24)');
    });

    it('calculates South Africa financial year bounds (28 February)', () => {
      const asOf = new Date(2024, 9, 22);
      const bounds = calculateFiscalYearBounds(2, 28, asOf);

      expect(bounds.currentPeriod).toBe('1 Mar 2024 – 28 Feb 2025 (FY25)');
      expect(bounds.prevPeriod).toBe('1 Mar 2023 – 29 Feb 2024 (FY24)');
    });

    it('handles leap year leap-day resolution for February 29', () => {
      // 2024 is a leap year; 2025 is not
      const asOf = new Date(2024, 0, 15); // Jan 15, 2024
      const bounds = calculateFiscalYearBounds(2, 29, asOf);

      expect(bounds.currentPeriod).toBe('1 Mar 2023 – 29 Feb 2024 (FY24)');
    });
  });

  describe('detectPresetId', () => {
    it('detects standard country presets correctly', () => {
      expect(detectPresetId(6, 30)).toBe('AU');
      expect(detectPresetId(4, 5)).toBe('UK');
      expect(detectPresetId(12, 31)).toBe('US');
      expect(detectPresetId(3, 31)).toBe('NZ');
      expect(detectPresetId(2, 28)).toBe('ZA');
    });

    it('returns CUSTOM for non-standard dates', () => {
      expect(detectPresetId(7, 15)).toBe('CUSTOM');
      expect(detectPresetId(6, 15)).toBe('CUSTOM');
      expect(detectPresetId(1, 1)).toBe('CUSTOM');
    });
  });

  describe('getMaxDaysInMonth', () => {
    it('returns 29 for February to permit leap-year setups', () => {
      expect(getMaxDaysInMonth(2)).toBe(29);
    });

    it('returns 30 for 30-day months (April, June, September, November)', () => {
      expect(getMaxDaysInMonth(4)).toBe(30);
      expect(getMaxDaysInMonth(6)).toBe(30);
      expect(getMaxDaysInMonth(9)).toBe(30);
      expect(getMaxDaysInMonth(11)).toBe(30);
    });

    it('returns 31 for 31-day months', () => {
      expect(getMaxDaysInMonth(1)).toBe(31);
      expect(getMaxDaysInMonth(3)).toBe(31);
      expect(getMaxDaysInMonth(5)).toBe(31);
      expect(getMaxDaysInMonth(7)).toBe(31);
      expect(getMaxDaysInMonth(8)).toBe(31);
      expect(getMaxDaysInMonth(10)).toBe(31);
      expect(getMaxDaysInMonth(12)).toBe(31);
    });
  });

  describe('EOFY_PRESETS', () => {
    it('contains all 5 institutional regional presets with correct month and day', () => {
      expect(EOFY_PRESETS).toHaveLength(5);
      const au = EOFY_PRESETS.find((p) => p.id === 'AU');
      expect(au).toEqual({ id: 'AU', label: 'Australia (30 Jun)', flag: '🇦🇺', month: 6, day: 30 });

      const uk = EOFY_PRESETS.find((p) => p.id === 'UK');
      expect(uk).toEqual({ id: 'UK', label: 'UK (5 Apr)', flag: '🇬🇧', month: 4, day: 5 });

      const us = EOFY_PRESETS.find((p) => p.id === 'US');
      expect(us).toEqual({ id: 'US', label: 'USA / Canada (31 Dec)', flag: '🇺🇸', month: 12, day: 31 });

      const nz = EOFY_PRESETS.find((p) => p.id === 'NZ');
      expect(nz).toEqual({ id: 'NZ', label: 'NZ / India (31 Mar)', flag: '🇳🇿', month: 3, day: 31 });

      const za = EOFY_PRESETS.find((p) => p.id === 'ZA');
      expect(za).toEqual({ id: 'ZA', label: 'South Africa (28 Feb)', flag: '🇿🇦', month: 2, day: 28 });
    });
  });
});

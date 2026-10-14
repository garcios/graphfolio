import { describe, it, expect } from 'vitest';
import {
  formatISODate,
  calculatePresetDates,
  calculatePresetDate,
  PRESET_OPTIONS,
} from './BackfillModal';

describe('BackfillModal preset and date helpers', () => {
  describe('formatISODate', () => {
    it('formats dates as YYYY-MM-DD using local date components', () => {
      const d = new Date(2026, 9, 14); // 2026-10-14
      expect(formatISODate(d)).toBe('2026-10-14');
    });

    it('pads single-digit month and day with leading zeroes', () => {
      const d = new Date(2026, 0, 5); // 2026-01-05
      expect(formatISODate(d)).toBe('2026-01-05');
    });
  });

  describe('PRESET_OPTIONS', () => {
    it('includes Yesterday as a selectable preset option', () => {
      expect(PRESET_OPTIONS).toContain('Yesterday');
      expect(PRESET_OPTIONS).toEqual(['Yesterday', '30D', '90D', 'YTD', '1Y', 'MAX']);
    });
  });

  describe('calculatePresetDates', () => {
    describe('Yesterday preset', () => {
      it('populates both Start Date and End Date with yesterday date', () => {
        const ref = new Date(2026, 9, 14); // 2026-10-14
        const result = calculatePresetDates('Yesterday', ref);
        expect(result.from).toBe('2026-10-13');
        expect(result.to).toBe('2026-10-13');
        expect(result.from).toBe(result.to);
      });

      it('correctly handles month rollback for Yesterday preset', () => {
        const ref = new Date(2026, 2, 1); // 2026-03-01
        const result = calculatePresetDates('Yesterday', ref);
        expect(result.from).toBe('2026-02-28');
        expect(result.to).toBe('2026-02-28');
      });

      it('correctly handles leap-year February rollback for Yesterday preset', () => {
        const ref = new Date(2024, 2, 1); // 2024-03-01 (leap year)
        const result = calculatePresetDates('Yesterday', ref);
        expect(result.from).toBe('2024-02-29');
        expect(result.to).toBe('2024-02-29');
      });

      it('correctly handles year rollback for Yesterday preset on New Year Day', () => {
        const ref = new Date(2026, 0, 1); // 2026-01-01
        const result = calculatePresetDates('Yesterday', ref);
        expect(result.from).toBe('2025-12-31');
        expect(result.to).toBe('2025-12-31');
      });
    });

    describe('Standard range presets', () => {
      const ref = new Date(2026, 9, 14); // 2026-10-14

      it('calculates 30D bounds ending on refDate', () => {
        const result = calculatePresetDates('30D', ref);
        expect(result.to).toBe('2026-10-14');
        expect(result.from).toBe('2026-09-14');
      });

      it('calculates 90D bounds ending on refDate', () => {
        const result = calculatePresetDates('90D', ref);
        expect(result.to).toBe('2026-10-14');
        expect(result.from).toBe('2026-07-16');
      });

      it('calculates YTD bounds starting at Jan 1st and ending on refDate', () => {
        const result = calculatePresetDates('YTD', ref);
        expect(result.to).toBe('2026-10-14');
        expect(result.from).toBe('2026-01-01');
      });

      it('calculates 1Y bounds ending on refDate', () => {
        const result = calculatePresetDates('1Y', ref);
        expect(result.to).toBe('2026-10-14');
        expect(result.from).toBe('2025-10-14');
      });

      it('calculates MAX (5 Years) bounds ending on refDate', () => {
        const result = calculatePresetDates('MAX', ref);
        expect(result.to).toBe('2026-10-14');
        expect(result.from).toBe('2021-10-14');
      });
    });
  });

  describe('calculatePresetDate helper', () => {
    it('returns the from date of the preset', () => {
      const ref = new Date(2026, 9, 14);
      expect(calculatePresetDate('Yesterday', ref)).toBe('2026-10-13');
      expect(calculatePresetDate('YTD', ref)).toBe('2026-01-01');
      expect(calculatePresetDate('30D', ref)).toBe('2026-09-14');
    });
  });
});

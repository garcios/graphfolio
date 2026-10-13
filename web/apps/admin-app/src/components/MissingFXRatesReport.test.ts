import { describe, it, expect } from 'vitest';
import {
  calculatePresetDates,
  calculateDateSpanDays,
  calculateBufferDate,
  getTypeBadgeVariant,
} from './MissingFXRatesReport';

describe('MissingFXRatesReport helper utilities', () => {
  describe('calculateBufferDate', () => {
    it('applies a default 3-day buffer backwards to cover weekend FX gaps', () => {
      expect(calculateBufferDate('2024-03-15', 3)).toBe('2024-03-12');
      expect(calculateBufferDate('2024-03-01', 3)).toBe('2024-02-27');
    });

    it('returns the original input if the string is invalid', () => {
      expect(calculateBufferDate('not-a-date', 3)).toBe('not-a-date');
    });

    it('defaults bufferDays to 3 if omitted', () => {
      expect(calculateBufferDate('2024-05-10')).toBe('2024-05-07');
    });
  });

  describe('calculateDateSpanDays', () => {
    it('returns dash when either date is null, undefined, or empty', () => {
      expect(calculateDateSpanDays(null, null)).toBe('—');
      expect(calculateDateSpanDays('2024-01-01', null)).toBe('—');
      expect(calculateDateSpanDays(undefined, '2024-01-01')).toBe('—');
      expect(calculateDateSpanDays('', '')).toBe('—');
    });

    it('returns 1 Day when fromDate and toDate are the same', () => {
      expect(calculateDateSpanDays('2024-02-14', '2024-02-14')).toBe('1 Day');
    });

    it('accurately calculates multi-day spans inclusive of boundary days', () => {
      expect(calculateDateSpanDays('2024-01-01', '2024-01-10')).toBe('10 Days');
      expect(calculateDateSpanDays('2024-01-01', '2024-01-31')).toBe('31 Days');
    });

    it('handles inverted date bounds symmetrically without negative days', () => {
      expect(calculateDateSpanDays('2024-01-10', '2024-01-01')).toBe('10 Days');
    });

    it('returns dash for invalid date strings', () => {
      expect(calculateDateSpanDays('not-a-date', '2024-01-10')).toBe('—');
    });
  });

  describe('calculatePresetDates', () => {
    it('returns empty bounds for ALL preset', () => {
      const res = calculatePresetDates('ALL');
      expect(res.from).toBe('');
      expect(res.to).toBe('');
    });

    it('returns valid ISO date strings for 30D where from <= to', () => {
      const res = calculatePresetDates('30D');
      expect(res.from).toMatch(/^\d{4}-\d{2}-\d{2}$/);
      expect(res.to).toMatch(/^\d{4}-\d{2}-\d{2}$/);
      expect(new Date(res.from).getTime()).toBeLessThanOrEqual(new Date(res.to).getTime());
    });

    it('returns valid ISO date strings for 90D', () => {
      const res = calculatePresetDates('90D');
      expect(res.from).toMatch(/^\d{4}-\d{2}-\d{2}$/);
      expect(res.to).toMatch(/^\d{4}-\d{2}-\d{2}$/);
    });

    it('returns Jan 1 of current year for YTD', () => {
      const res = calculatePresetDates('YTD');
      const currentYear = new Date().getFullYear();
      expect(res.from).toBe(`${currentYear}-01-01`);
    });

    it('returns valid ISO date strings for 1Y', () => {
      const res = calculatePresetDates('1Y');
      expect(res.from).toMatch(/^\d{4}-\d{2}-\d{2}$/);
      expect(res.to).toMatch(/^\d{4}-\d{2}-\d{2}$/);
    });
  });

  describe('getTypeBadgeVariant', () => {
    it('maps BUY to success', () => {
      expect(getTypeBadgeVariant('BUY')).toBe('success');
    });

    it('maps SELL to info', () => {
      expect(getTypeBadgeVariant('SELL')).toBe('info');
    });

    it('maps DIVIDEND to warning', () => {
      expect(getTypeBadgeVariant('DIVIDEND')).toBe('warning');
    });

    it('maps SPLIT to purple', () => {
      expect(getTypeBadgeVariant('SPLIT')).toBe('purple');
    });

    it('maps cash flow and transfer types to neutral', () => {
      expect(getTypeBadgeVariant('DEPOSIT')).toBe('neutral');
      expect(getTypeBadgeVariant('WITHDRAWAL')).toBe('neutral');
      expect(getTypeBadgeVariant('TRANSFER_IN')).toBe('neutral');
      expect(getTypeBadgeVariant('TRANSFER_OUT')).toBe('neutral');
      expect(getTypeBadgeVariant('FEE')).toBe('neutral');
    });
  });
});

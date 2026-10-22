import { describe, it, expect } from 'vitest';

/**
 * Pure calculation logic tests for Dividend Calendar & Yield Analytics (US-DIV-101)
 */

export function isTradeDateEntitled(tradeDate: string, exDate: string): boolean {
  return tradeDate < exDate;
}

export function calculateYield(annualIncome: number, baseValue: number): number | null {
  if (baseValue <= 0) return null;
  return (annualIncome / baseValue) * 100;
}

export function calculateYieldExpansion(yieldOnCost: number | null, currentYield: number): number | null {
  if (yieldOnCost === null) return null;
  return yieldOnCost - currentYield;
}

export function simulateDRIPCompounding(
  startingVal: number,
  startingIncome: number,
  years: number,
  divGrowthPct: number,
  capAppPct: number
): { terminalDrip: number; terminalCash: number; benefit: number } {
  let dripShares = 1000;
  let unitPrice = startingVal / dripShares;
  let dps = startingIncome / dripShares;

  // DRIP simulation
  for (let yr = 1; yr <= years; yr++) {
    const annualDiv = dripShares * dps;
    const additionalShares = unitPrice > 0 ? annualDiv / unitPrice : 0;
    dripShares += additionalShares;
    unitPrice *= (1 + capAppPct / 100);
    dps *= (1 + divGrowthPct / 100);
  }
  const terminalDrip = dripShares * unitPrice;

  // Cash simulation (shares constant at 1000)
  let cashUnitPrice = startingVal / 1000;
  for (let yr = 1; yr <= years; yr++) {
    cashUnitPrice *= (1 + capAppPct / 100);
  }
  const terminalCash = 1000 * cashUnitPrice;

  return {
    terminalDrip,
    terminalCash,
    benefit: terminalDrip - terminalCash,
  };
}

describe('Dividend Calendar & Entitlement Tracking (US-DIV-101)', () => {
  describe('isTradeDateEntitled', () => {
    it('returns true when trade settlement precedes ex-dividend date', () => {
      expect(isTradeDateEntitled('2026-05-10', '2026-05-15')).toBe(true);
    });

    it('returns false when trade date equals ex-dividend date (bought on ex-date)', () => {
      expect(isTradeDateEntitled('2026-05-15', '2026-05-15')).toBe(false);
    });

    it('returns false when trade date is after ex-dividend date', () => {
      expect(isTradeDateEntitled('2026-05-16', '2026-05-15')).toBe(false);
    });
  });

  describe('Current Yield and Yield on Cost (YOC)', () => {
    it('calculates current yield accurately', () => {
      const yld = calculateYield(450, 10000);
      expect(yld).toBeCloseTo(4.5, 4);
    });

    it('calculates yield on cost accurately when cost basis is lower than market value', () => {
      const yoc = calculateYield(450, 6000);
      expect(yoc).toBeCloseTo(7.5, 4);
    });

    it('safely handles zero or negative cost basis without throwing division by zero', () => {
      expect(calculateYield(450, 0)).toBeNull();
      expect(calculateYield(450, -100)).toBeNull();
    });

    it('calculates compounding expansion when YOC exceeds current yield', () => {
      const curYield = calculateYield(450, 10000)!; // 4.5%
      const yoc = calculateYield(450, 6000)!;      // 7.5%
      const expansion = calculateYieldExpansion(yoc, curYield);
      expect(expansion).toBeCloseTo(3.0, 4);
    });

    it('returns null expansion if YOC is null', () => {
      expect(calculateYieldExpansion(null, 4.5)).toBeNull();
    });
  });

  describe('simulateDRIPCompounding', () => {
    it('demonstrates positive compounding advantage over cash withdrawals', () => {
      const res = simulateDRIPCompounding(100000, 4000, 10, 5, 7);
      expect(res.terminalDrip).toBeGreaterThan(res.terminalCash);
      expect(res.benefit).toBeGreaterThan(0);
      expect(res.terminalDrip).toBeGreaterThan(250000);
    });

    it('matches terminal values for 0 horizon years', () => {
      const res = simulateDRIPCompounding(100000, 4000, 0, 5, 7);
      expect(res.terminalDrip).toBe(100000);
      expect(res.terminalCash).toBe(100000);
      expect(res.benefit).toBe(0);
    });
  });
});

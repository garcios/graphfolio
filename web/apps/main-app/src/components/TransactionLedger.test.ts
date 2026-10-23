import { describe, it, expect } from 'vitest';
import {
  toggleSortDirection,
  buildTransactionQueryArgs,
  getSortHeaderAria,
  getSortHeaderTitle,
  filterTypes,
  getBadgeClass,
  getEmptyStateMessage,
} from './TransactionLedger';

describe('Transaction Ledger Date Sorting (US-TX-104)', () => {
  describe('toggleSortDirection', () => {
    it('toggles from DESC (newest first) to ASC (oldest first)', () => {
      expect(toggleSortDirection('DESC')).toBe('ASC');
    });

    it('toggles from ASC (oldest first) to DESC (newest first)', () => {
      expect(toggleSortDirection('ASC')).toBe('DESC');
    });
  });

  describe('buildTransactionQueryArgs', () => {
    it('builds default query arguments with DESC sort direction', () => {
      const args = buildTransactionQueryArgs(1, 20, 'DESC');
      expect(args).toEqual({
        page: 1,
        pageSize: 20,
        sortDirection: 'DESC',
      });
    });

    it('builds query arguments with ASC sort direction', () => {
      const args = buildTransactionQueryArgs(1, 20, 'ASC');
      expect(args).toEqual({
        page: 1,
        pageSize: 20,
        sortDirection: 'ASC',
      });
    });

    it('preserves active type filter across sort direction changes', () => {
      const descArgs = buildTransactionQueryArgs(1, 20, 'DESC', 'DIVIDEND');
      expect(descArgs).toEqual({
        page: 1,
        pageSize: 20,
        sortDirection: 'DESC',
        type: 'DIVIDEND',
      });

      const ascArgs = buildTransactionQueryArgs(1, 20, 'ASC', 'DIVIDEND');
      expect(ascArgs).toEqual({
        page: 1,
        pageSize: 20,
        sortDirection: 'ASC',
        type: 'DIVIDEND',
      });
    });

    it('preserves and trims/uppercases symbol search query', () => {
      const args = buildTransactionQueryArgs(1, 20, 'ASC', null, '  bhp.ax  ');
      expect(args).toEqual({
        page: 1,
        pageSize: 20,
        sortDirection: 'ASC',
        symbol: 'BHP.AX',
      });
    });

    it('omits empty or whitespace-only symbol search', () => {
      const args = buildTransactionQueryArgs(1, 20, 'DESC', null, '   ');
      expect(args.symbol).toBeUndefined();
      expect(args).toEqual({
        page: 1,
        pageSize: 20,
        sortDirection: 'DESC',
      });
    });

    it('combines type filter, symbol query, and sort direction simultaneously', () => {
      const args = buildTransactionQueryArgs(2, 20, 'ASC', 'BUY', 'aapl');
      expect(args).toEqual({
        page: 2,
        pageSize: 20,
        sortDirection: 'ASC',
        type: 'BUY',
        symbol: 'AAPL',
      });
    });
  });

  describe('Accessibility & Header Helpers', () => {
    describe('getSortHeaderAria', () => {
      it('returns "descending" when sortDirection is DESC', () => {
        expect(getSortHeaderAria('DESC')).toBe('descending');
      });

      it('returns "ascending" when sortDirection is ASC', () => {
        expect(getSortHeaderAria('ASC')).toBe('ascending');
      });
    });

    describe('getSortHeaderTitle', () => {
      it('returns descriptive title indicating click will switch to oldest first when DESC', () => {
        const title = getSortHeaderTitle('DESC');
        expect(title).toContain('Newest first');
        expect(title).toContain('click to sort oldest first');
      });

      it('returns descriptive title indicating click will switch to newest first when ASC', () => {
        const title = getSortHeaderTitle('ASC');
        expect(title).toContain('Oldest first');
        expect(title).toContain('click to sort newest first');
      });
    });
  });

  describe('Deterministic Ordering Integration', () => {
    it('verifies sortDirection parameter maps to GraphQL model values', () => {
      const directions: Array<'DESC' | 'ASC'> = ['DESC', 'ASC'];
      for (const dir of directions) {
        const args = buildTransactionQueryArgs(1, 20, dir);
        expect(['DESC', 'ASC']).toContain(args.sortDirection);
      }
    });

    it('resets page to 1 on sort toggle simulation', () => {
      let currentPage = 4;
      let currentSort: 'DESC' | 'ASC' = 'DESC';

      // Simulate handleSortDate
      const handleSortDate = () => {
        currentSort = toggleSortDirection(currentSort);
        currentPage = 1;
      };

      handleSortDate();
      expect(currentSort).toBe('ASC');
      expect(currentPage).toBe(1);

      handleSortDate();
      expect(currentSort).toBe('DESC');
      expect(currentPage).toBe(1);
    });
  });
});

describe('Transaction Ledger Interest Filter (US-TX-106)', () => {
  describe('filterTypes constant', () => {
    it('contains the Interest filter pill with value INTEREST', () => {
      const interestPill = filterTypes.find((f) => f.value === 'INTEREST');
      expect(interestPill).toBeDefined();
      expect(interestPill?.label).toBe('Interest');
    });

    it('positions Interest immediately following Dividend in the filter order', () => {
      const dividendIdx = filterTypes.findIndex((f) => f.value === 'DIVIDEND');
      const interestIdx = filterTypes.findIndex((f) => f.value === 'INTEREST');
      expect(dividendIdx).toBeGreaterThanOrEqual(0);
      expect(interestIdx).toBe(dividendIdx + 1);
    });

    it('contains all expected transaction type filters with All as first entry', () => {
      expect(filterTypes[0]).toEqual({ label: 'All', value: null });
      const values = filterTypes.map((f) => f.value);
      expect(values).toEqual([
        null,
        'BUY',
        'SELL',
        'DIVIDEND',
        'INTEREST',
        'SPLIT',
        'DEPOSIT',
        'WITHDRAWAL',
        'FX_CONVERSION',
      ]);
    });
  });

  describe('buildTransactionQueryArgs with INTEREST', () => {
    it('passes type = "INTEREST" when selectedType is INTEREST', () => {
      const args = buildTransactionQueryArgs(1, 20, 'DESC', 'INTEREST');
      expect(args).toEqual({
        page: 1,
        pageSize: 20,
        sortDirection: 'DESC',
        type: 'INTEREST',
      });
    });

    it('combines INTEREST type filter with ASC sorting and symbol search', () => {
      const args = buildTransactionQueryArgs(2, 20, 'ASC', 'INTEREST', 'cma');
      expect(args).toEqual({
        page: 2,
        pageSize: 20,
        sortDirection: 'ASC',
        type: 'INTEREST',
        symbol: 'CMA',
      });
    });

    it('preserves INTEREST filter across sort direction changes', () => {
      const descArgs = buildTransactionQueryArgs(1, 20, 'DESC', 'INTEREST');
      const ascArgs = buildTransactionQueryArgs(1, 20, 'ASC', 'INTEREST');
      expect(descArgs.type).toBe('INTEREST');
      expect(ascArgs.type).toBe('INTEREST');
    });
  });

  describe('getBadgeClass', () => {
    it('returns "badge-type interest" for INTEREST type', () => {
      expect(getBadgeClass('INTEREST')).toBe('badge-type interest');
    });

    it('returns correct badge classes for standard types', () => {
      expect(getBadgeClass('BUY')).toBe('badge-type buy');
      expect(getBadgeClass('SELL')).toBe('badge-type sell');
      expect(getBadgeClass('DIVIDEND')).toBe('badge-type dividend');
      expect(getBadgeClass('SPLIT')).toBe('badge-type split');
      expect(getBadgeClass('DEPOSIT')).toBe('badge-type deposit');
      expect(getBadgeClass('WITHDRAWAL')).toBe('badge-type withdrawal');
      expect(getBadgeClass('FX_CONVERSION')).toBe('badge-type fx');
    });

    it('returns "badge-type other" for unknown or unexpected types', () => {
      expect(getBadgeClass('UNKNOWN')).toBe('badge-type other');
      expect(getBadgeClass('')).toBe('badge-type other');
    });
  });

  describe('getEmptyStateMessage', () => {
    it('returns interest-specific empty message when selectedType is INTEREST', () => {
      expect(getEmptyStateMessage('INTEREST')).toBe(
        'No interest transactions found for this portfolio.'
      );
    });

    it('returns interest and symbol empty message when both are present', () => {
      expect(getEmptyStateMessage('INTEREST', 'cma')).toBe(
        'No interest transactions found matching "CMA".'
      );
    });

    it('handles uppercase trimming for symbol search with interest filter', () => {
      expect(getEmptyStateMessage('INTEREST', '  cash  ')).toBe(
        'No interest transactions found matching "CASH".'
      );
    });

    it('returns default empty message when no filters are active', () => {
      expect(getEmptyStateMessage(null, '')).toBe(
        'No transactions found matching your criteria.'
      );
    });

    it('returns symbol-specific message when only symbol query is active', () => {
      expect(getEmptyStateMessage(null, 'bhp.ax')).toBe(
        'No transactions found matching "BHP.AX".'
      );
    });

    it('returns general type-specific message for other transaction types', () => {
      expect(getEmptyStateMessage('DIVIDEND')).toBe(
        'No dividend transactions found for this portfolio.'
      );
    });
  });
});

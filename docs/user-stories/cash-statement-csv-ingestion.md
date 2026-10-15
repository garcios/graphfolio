# User Story: Cash Account Statement CSV Ingestion

> **Story ID**: US-CSV-102  
> **Epic**: Transaction Ledger & Statement Import  
> **Target**: Primary Investor Application (`web/apps/main-app`), CSV Ingestion Engine (`web/apps/main-app/src/services/csv/`), BFF GraphQL Orchestrator (`bff/`), Portfolio Microservice (`services/portfolio-api`)  
> **Status**: **DONE**  
> **Implementation Plan**: [cash-statement-csv-ingestion-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/cash-statement-csv-ingestion-implementation-plan.md)  
> **Related Stories**: [multi-broker-csv-ingestion.md](file:///Users/oscargarcia/workspace/graphfolio/docs/user-stories/multi-broker-csv-ingestion.md)  
> **Standards Reference**: [financial-calculations skill](file:///Users/oscargarcia/workspace/graphfolio/.agents/skills/financial-calculations/SKILL.md), [AGENTS.md](file:///Users/oscargarcia/workspace/graphfolio/AGENTS.md)  

---

## 1. User Story Statement

> **As an** investor importing historical brokerage records,  
> **I want to** upload bank and brokerage cash account CSV statements (such as nabtrade Cash Account exports) containing deposits, withdrawals, interest credits, and dividend transfers,  
> **So that** my portfolio cash balance accurately reflects my funded capital, eliminating artificial negative cash balances caused by debiting stock purchases against a zero cash baseline.

---

## 2. Business Context & Financial Integrity

1. **Negative Cash Balance Remediation**:
   - When investors import equity trade confirmations without corresponding cash account records, hundreds of stock purchases are debited against an initial balance of `$0.00`, resulting in severe negative cash distortions (e.g. `-A$250,187.80`).
2. **Cash Statement Row Archetypes**:
   - Authoritative broker cash statements (`Date,Type,Description,Debit,Credit,Balance`) combine equity trades with funding and cash events:
     - Cash Deposits: `Credit > 0` with descriptions like `FUNDS TRANSFER ... deposit [Name]` $\rightarrow$ mapped to `DEPOSIT`.
     - Cash Withdrawals: `Debit > 0` with descriptions like `FUNDS TRANSFER ... transfer [Name]` $\rightarrow$ mapped to `WITHDRAWAL`.
     - Monthly Interest Income: `INTEREST` credits $\rightarrow$ mapped to `INTEREST`.
     - Domestic Dividends: `FUNDS TRANSFER DIVIDEND - [TICKER]` credits $\rightarrow$ mapped to `DIVIDEND` linked to the ticker.
     - International Dividends: `DIVIDEND on [TICKER].US (WHT of USD [WHT])` credits $\rightarrow$ mapped to `DIVIDEND` with withholding tax metadata.
     - Informational Policy Rows: `InterestChange` notifications $\rightarrow$ ignored cleanly.
3. **Empty Symbol Bypass**:
   - Unlike stock transactions, pure cash events (`DEPOSIT`, `WITHDRAWAL`, `INTEREST`) carry an empty symbol (`""`) and bypass equity instrument catalog validation while updating cash balances in the ledger.

---

## 3. Acceptance Criteria

### 3.1 CSV Dialect & Row Pattern Recognition
- [x] **AC-1 (Cash Account Header Detection)**: Client-side detector recognizes `Date,Type,Description,Debit,Credit,Balance` headers as nabtrade cash account format.
- [x] **AC-2 (Deposit & Withdrawal Parsing)**: Cash deposit rows map to `NormalizedTxType: DEPOSIT` and withdrawals to `WITHDRAWAL` in portfolio base currency (e.g. `AUD`).
- [x] **AC-3 (Monthly Interest Parsing)**: Bank interest credits map to `NormalizedTxType: INTEREST` with gross income amounts.
- [x] **AC-4 (Domestic & International Dividends)**: Domestic dividend rows extract ticker symbols from `FUNDS TRANSFER DIVIDEND - {TICKER}`. US equity dividends parse gross dividends, net proceeds, and withholding tax amounts.
- [x] **AC-5 (Informational Row Filtering)**: Non-transactional rows (such as interest rate changes) are skipped cleanly without throwing parser errors.

### 3.2 Deduplication & External Reference Generation
- [x] **AC-6 (Cash External References)**: Deterministic unique references are generated for cash events (e.g. `nabtrade:cash:{ref}` or hash-based keys) to prevent duplicate deposit imports.
- [x] **AC-7 (Trade Deduplication Synergy)**: Equity trades present in the cash statement match order IDs from trade confirmation exports, preventing double-counting trades.

### 3.3 Backend Persistence & Ledger Replay
- [x] **AC-8 (Zero-Symbol Ingestion Support)**: Backend `BatchImportTransactions` accepts transactions with empty instrument symbols for cash movements without foreign key errors.
- [x] **AC-9 (Cash Balance Recalculation)**: Ledger projection engine (`RebuildProjections`) processes all cash movements chronologically, restoring true non-negative cash balances.

---

## 4. Technical & Implementation Notes

- **CSV Parser**: `web/apps/main-app/src/services/csv/parsers/nabtrade.ts` implements regex parsing for cash account transaction descriptions.
- **Detector**: `web/apps/main-app/src/services/csv/detector.ts` includes `Date,Type,Description,Debit,Credit,Balance` signatures.
- **Unit Tests**: `web/apps/main-app/src/services/csv/parsers/nabtrade.test.ts` tests all 7 row archetypes including deposits, interest, and dividends.
- **Backend Service**: `services/portfolio-api/internal/service/transaction.go` supports symbol-less cash transactions in batch imports.

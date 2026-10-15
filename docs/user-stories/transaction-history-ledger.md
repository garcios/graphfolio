# User Story: Transaction History Ledger & Safe Deletion

> **Story ID**: US-TX-102  
> **Epic**: Transaction Ledger & Trade Management  
> **Target**: Primary Investor Application (`web/apps/main-app`), BFF GraphQL Orchestrator (`bff/`), Portfolio Microservice (`services/portfolio-api`), Protobuf (`proto/portfolio/v1/portfolio.proto`), Database (`portfolio.transactions`)  
> **Status**: **DONE**  
> **Implementation Plan**: [transaction-history-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/transaction-history-implementation-plan.md)  
> **Roadmap Reference**: [FEATURES.md (Feature #8)](file:///Users/oscargarcia/workspace/graphfolio/FEATURES.md#L18)  
> **Standards Reference**: [financial-calculations skill](file:///Users/oscargarcia/workspace/graphfolio/.agents/skills/financial-calculations/SKILL.md), [AGENTS.md](file:///Users/oscargarcia/workspace/graphfolio/AGENTS.md)  

---

## 1. User Story Statements

### 1.1 Ledger History Inspection & Filtering
> **As an** investor tracking my past market orders,  
> **I want to** view a complete, paginated transaction history table with filtering by transaction type (`BUY`, `SELL`, `DIVIDEND`, `DEPOSIT`, `WITHDRAWAL`, `SPLIT`) and ticker symbol,  
> **So that** I can audit past execution prices, brokerage fees, trade dates, and cash movements.

### 1.2 Safe Transaction Deletion & Re-Projection
> **As an** investor who entered an incorrect transaction,  
> **I want to** delete an errant transaction record and have my entire portfolio holdings, tax lots, cash balances, and valuations automatically recalculated,  
> **So that** my portfolio state remains 100% accurate without manual ledger adjustments or orphaned records.

---

## 2. Business Context & Event-Sourced Replay

1. **Authoritative Ledger Architecture**:
   - `portfolio.transactions` is the authoritative transaction log.
   - Any transaction deletion requires removing the transaction row from the database and immediately executing `ProcessLedger` over all remaining transactions for the portfolio.
2. **Negative Inventory Prevention**:
   - Deleting a `BUY` transaction that provided shares for subsequent `SELL` orders is guarded; if removing a transaction causes share quantities to drop below zero at any historical point, the operation is blocked to preserve physical inventory integrity.
3. **Valuation Reconciliation Hook**:
   - Deleting a past transaction invalidates daily valuation snapshots forward from the deleted trade's date, automatically triggering valuation backfill to restore accurate performance curves.

---

## 3. Acceptance Criteria

### 3.1 Paginated Ledger Presentation & Filtering
- [x] **AC-1 (Transaction Ledger View)**: `TransactionLedger.tsx` displays a responsive table listing Trade Date, Type badge, Symbol, Units, Unit Price, Brokerage Fee, Net Amount, and Currency.
- [x] **AC-2 (Multi-Criteria Filtering)**: Dropdowns allow filtering by transaction type (`ALL`, `BUY`, `SELL`, `DIVIDEND`, `DEPOSIT`, `WITHDRAWAL`, `SPLIT`) and searching by asset symbol.
- [x] **AC-3 (Server-Side Pagination)**: Query supports `page` and `pageSize`, displaying current records and total count.

### 3.2 Safe Transaction Deletion
- [x] **AC-4 (Delete Action & Confirmation)**: Each transaction row provides a "Delete" action button opening a confirmation modal detailing the transaction to be removed.
- [x] **AC-5 (Deterministic Projection Rebuild)**: Deleting a transaction removes the row and atomically invokes `RebuildProjections`, recalculating all tax lots, holdings, and cash balances from remaining records.
- [x] **AC-6 (Valuation Replay Trigger)**: Deleting a historical transaction automatically triggers daily valuation backfills from the deleted trade's date forward.
- [x] **AC-7 (Instant UI Synchronization)**: Upon successful deletion, the ledger table refreshes, dashboard metrics update immediately, and a confirmation toast is displayed.

---

## 4. Technical & Implementation Notes

- **Protobuf**: `proto/portfolio/v1/portfolio.proto` defines `ListTransactions` (with `TransactionFilter`, `TransactionPage`) and `DeleteTransaction`.
- **Backend Service**: `services/portfolio-api/internal/service/transaction.go` coordinates `DeleteTransaction`, `RebuildProjections`, and `BackfillPortfolioValuations`.
- **Web UI**: `web/apps/main-app/src/components/TransactionLedger.tsx` and `.css`.
- **BFF GraphQL**: Exposes `transactions` query and `deleteTransaction` mutation.

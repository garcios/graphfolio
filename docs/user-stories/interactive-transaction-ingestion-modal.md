# User Story: Interactive Transaction Ingestion Modal

> **Story ID**: US-TX-101  
> **Epic**: Transaction Ledger & Trade Ingestion  
> **Target**: Primary Investor Application (`web/apps/main-app`), BFF GraphQL Orchestrator (`bff/`), Portfolio Microservice (`services/portfolio-api`), Protobuf (`proto/portfolio/v1/portfolio.proto`)  
> **Status**: **DONE**  
> **Implementation Plan**: [add-transactions-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/add-transactions-implementation-plan.md)  
> **Roadmap Reference**: [FEATURES.md (Feature #6)](file:///Users/oscargarcia/workspace/graphfolio/FEATURES.md#L16)  
> **Standards Reference**: [financial-calculations skill](file:///Users/oscargarcia/workspace/graphfolio/.agents/skills/financial-calculations/SKILL.md), [AGENTS.md](file:///Users/oscargarcia/workspace/graphfolio/AGENTS.md)  

---

## 1. User Story Statement

> **As an** active investor,  
> **I want to** enter new investment trades and cash events through an interactive, multi-type transaction modal with real-time asset lookup, multi-currency fee handling, stock split ratio presets, and date validation,  
> **So that** my portfolio holdings, cash balances, tax lots, and total annualized return metrics are immediately and accurately updated without requiring a full page refresh.

---

## 2. Business Context & Financial Rules

1. **Append-Only Event Sourcing**: Transactions recorded via the ingestion modal are appended directly to `portfolio.transactions`. They trigger an automatic replay of the deterministic ledger projection engine (`ProcessLedger`), deriving updated tax lots (`portfolio.tax_lots`), lot disposals (`portfolio.lot_disposals`), open holdings (`portfolio.holdings`), and cash balances (`portfolio.cash_balances`).
2. **Supported Event Types**:
   - `BUY`: Adds to equity position, creates tax lots, and debits cash.
   - `SELL`: Disposes tax lots (Average Cost or FIFO), records realized capital gain/loss, and credits cash net of brokerage fees.
   - `DIVIDEND`: Credits income cash, records foreign withholding tax deductions, and attributes to investment yield.
   - `DEPOSIT`: Increases cash account balance without altering equity holdings.
   - `WITHDRAWAL`: Decreases cash account balance.
   - `SPLIT`: Adjusts lot share counts and per-share cost basis according to split ratios while preserving total historical cost basis.
3. **Strict Zero Floating-Point Precision**: Every unit quantity, execution price, brokerage fee, and withholding tax is captured and transmitted with exact arbitrary precision (`common.v1.Decimal` / `common.v1.Money`).

---

## 3. Acceptance Criteria

### 3.1 Modal Trigger & Presentation
- [x] **AC-1 (Modal Access)**: An "+ Add Transaction" action button is prominent in the dashboard header and ledger view, opening a dark glassmorphic modal (`AddTransactionModal.tsx`).
- [x] **AC-2 (Type Selector)**: Segmented control or selector allows choosing between `BUY`, `SELL`, `DIVIDEND`, `SPLIT`, `DEPOSIT`, and `WITHDRAWAL`.
- [x] **AC-3 (Dynamic Form Fields)**:
  - For `BUY` / `SELL`: Instrument search, trade date, quantity, unit price, brokerage fee, and fee currency selector.
  - For `DIVIDEND`: Instrument selector, record date, gross dividend amount, withholding tax, and payment currency.
  - For `SPLIT`: Instrument selector, split execution date, split ratio presets (2:1, 3:1, 4:1, 10:1, and reverse split options 1:2, 1:5, 1:10) or custom ratio inputs.
  - For `DEPOSIT` / `WITHDRAWAL`: Amount, currency, execution date, and optional reference notes.

### 3.2 Form Validation & Multi-Currency Handling
- [x] **AC-4 (Field Validation)**: Positive non-zero numbers enforced for quantity and price; future dates beyond today are rejected; required fields prevent submission when blank.
- [x] **AC-5 (Fee Currency Decoupling)**: Brokerage fee currency defaults to the investor's preferred display currency, permitting multi-currency trading (e.g. trading USD equities while paying AUD brokerage fees).
- [x] **AC-6 (Instrument Discovery)**: Dynamic typeahead search queries `ListInstruments` filtering by symbol and company name.

### 3.3 Mutation Execution & Ledger Synchronization
- [x] **AC-7 (BFF GraphQL & gRPC Mutation)**: Submitting triggers GraphQL `addTransaction` calling gRPC `AddTransaction`.
- [x] **AC-8 (Deterministic Projection Rebuild)**: Ingestion atomically invokes `ProcessLedger`, updating holding quantities, volume-weighted average buy price, and cash balances.
- [x] **AC-9 (Valuation Replay Hook)**: Adding a transaction with a trade date prior to the latest valuation triggers daily valuation backfills from the trade date forward.
- [x] **AC-10 (Optimistic UI & Feedback)**: Upon successful submission, modal closes, displays a success toast notification, and immediately updates dashboard cards, holding tables, and performance curves.

---

## 4. Technical & Implementation Notes

- **Protobuf**: `proto/portfolio/v1/portfolio.proto` defines `AddTransactionRequest` and `AddTransactionResponse`.
- **Backend Service**: `services/portfolio-api/internal/service/transaction.go` manages insertion, negative inventory validation, projection rebuilds (`RebuildProjections`), and valuation reconciliation (`BackfillPortfolioValuations`).
- **Web UI**: `web/apps/main-app/src/components/AddTransactionModal.tsx` styled with `@graphfolio/ui` tokens.

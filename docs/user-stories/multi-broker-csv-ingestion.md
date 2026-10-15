# User Story: Multi-Broker CSV Ingestion & Reconciliation (CommSec & nabtrade)

> **Story ID**: US-CSV-101  
> **Epic**: Transaction Ledger & Statement Import  
> **Target**: Primary Investor Application (`web/apps/main-app`), Client CSV Engine (`web/apps/main-app/src/services/csv/`), BFF GraphQL Orchestrator (`bff/`), Portfolio Microservice (`services/portfolio-api`), PostgreSQL (`portfolio` schema)  
> **Status**: **DONE**  
> **Implementation Plan**: [multi-broker-csv-ingestion-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/multi-broker-csv-ingestion-implementation-plan.md)  
> **Roadmap Reference**: [FEATURES.md (Feature #21)](file:///Users/oscargarcia/workspace/graphfolio/FEATURES.md#L31)  
> **Standards Reference**: [financial-calculations skill](file:///Users/oscargarcia/workspace/graphfolio/.agents/skills/financial-calculations/SKILL.md), [AGENTS.md](file:///Users/oscargarcia/workspace/graphfolio/AGENTS.md)  

---

## 1. User Story Statement

> **As an** investor trading across Australian brokerage platforms (CommSec and nabtrade),  
> **I want to** drag and drop historical CSV trade statements and confirmations into an import modal, automatically detect the broker dialect, preview parsed records with duplicate detection, and batch import them,  
> **So that** my entire multi-year transaction history is ingested in seconds with zero manual data entry and complete ledger accuracy.

---

## 2. Business Context & Ingestion Requirements

1. **Multi-Broker Australian Dialects**:
   - CommSec and nabtrade are Australia's leading retail brokers, but their CSV exports differ drastically in header names, trade action labels, date formats (`DD/MM/YYYY`), currency formatting (currency symbols, thousand commas, negative numbers in accounting parentheses), and metadata header/footer rows.
2. **Client-Side Validation & Security**:
   - File parsing and dialect detection execute entirely on the client using **Zod** schemas, providing instant feedback without sending unparsed files across the wire.
3. **Deterministic Idempotency & Deduplication**:
   - Each transaction generates a unique deterministic external reference (`external_ref`), constructed from broker trade confirmation numbers or cryptographic hashes (`SHA-256`) of normalized trade tuples.
   - Pre-flight duplicate checking via the BFF queries existing ledger references, allowing investors to re-import cumulative statements while skipping already-imported trades.
4. **Atomic Batch Ingestion & Replay**:
   - Backend `BatchImportTransactions` executes single-pass database insertion (`ON CONFLICT (portfolio_id, external_ref) DO NOTHING`), immediately triggering projection rebuilds (`RebuildProjections`) and historical valuation backfills.

---

## 3. Acceptance Criteria

### 3.1 Broker Detection & Parsing Engine
- [x] **AC-1 (Automatic Dialect Sniffing)**: System identifies whether an uploaded CSV is CommSec or nabtrade based on header signatures, while offering a manual override dropdown.
- [x] **AC-2 (CommSec Format Parsing)**: Correctly parses CommSec trade confirmations (`Details`, `Date`, `Reference`, `Units`, `Unit Price ($)`, `Brokerage ($)`, `Total ($)`) and historical ledger exports.
- [x] **AC-3 (nabtrade Format Parsing)**: Correctly parses nabtrade confirmation exports, strips leading metadata rows and account headers, and normalizes order IDs.
- [x] **AC-4 (Australian Locale Sanitization)**: Parses `DD/MM/YYYY` dates into ISO `YYYY-MM-DD`, strips currency symbols (`$`, `A$`), handles thousand separators, and converts accounting parentheses `(123.45)` to negative numbers.

### 3.2 Pre-Flight Preview & Duplicate Detection
- [x] **AC-5 (Interactive Import Modal)**: `ImportTransactionsModal.tsx` provides a drag-and-drop file zone, summary metric cards (Total Rows, Valid Trades, Duplicates, Errors), and a tabbed record preview.
- [x] **AC-6 (Pre-Flight Duplicate Check)**: Client queries `checkTransactionDuplicates(portfolioId, externalRefs)`; duplicates are visually flagged and excluded from the import batch by default.
- [x] **AC-7 (Line-Level Error Inspection)**: Malformed rows display actionable error messages (e.g. invalid date or unparseable numbers) without blocking valid rows.

### 3.3 Atomic Batch Persistence & Replay
- [x] **AC-8 (Batch Persistence RPC)**: `BatchImportTransactions` persists valid transactions atomically via `pgx.Batch` with `ON CONFLICT DO NOTHING`.
- [x] **AC-9 (Ledger Projection Replay)**: Ingestion triggers a single-pass `RebuildProjections`, recomputing tax lots, average cost bases, cash balances, and holdings.
- [x] **AC-10 (Valuation Backfill Hook)**: Historical daily valuations are automatically backfilled from the earliest imported trade date forward.

---

## 4. Technical & Implementation Notes

- **CSV Parsers**: `web/apps/main-app/src/services/csv/` (`detector.ts`, `types.ts`, `parsers/commsec.ts`, `parsers/nabtrade.ts`).
- **Modal Component**: `web/apps/main-app/src/components/ImportTransactionsModal.tsx` and `.css`.
- **Protobuf**: `proto/portfolio/v1/portfolio.proto` defines `CheckTransactionDuplicates` and `BatchImportTransactions`.
- **Backend Service**: `services/portfolio-api/internal/service/transaction.go` implements duplicate checks, batch inserts, and post-import valuation replay.

# User Story: PostgreSQL Multi-Schema Database Architecture

> **Story ID**: US-DB-101  
> **Epic**: Database Infrastructure & Relational Persistence  
> **Target**: Database Engine (`PostgreSQL 17+`), Migration Runner (`pkg/database/migrate.go`), Microservices (`services/portfolio-api`, `services/user-api`), Shared Infrastructure (`pkg/database/`)  
> **Status**: **DONE**  
> **Implementation Plan**: [db-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/db-implementation-plan.md)  
> **Roadmap Reference**: [FEATURES.md (Feature #1)](file:///Users/oscargarcia/workspace/graphfolio/FEATURES.md#L11)  
> **Standards Reference**: [AGENTS.md](file:///Users/oscargarcia/workspace/graphfolio/AGENTS.md)  

---

## 1. User Story Statement

> **As a** platform software engineer,  
> **I want to** establish a multi-schema PostgreSQL persistence layer with schema isolation (`portfolio` and `users`), dedicated service credentials, UUIDv7 time-ordered primary keys, and exact arbitrary-precision `NUMERIC` types,  
> **So that** domain microservices have an append-only transaction ledger, deterministic projection tables, and audit logs that maintain strict zero floating-point arithmetic and prevent schema coupling.

---

## 2. Architectural & Technical Requirements

1. **Schema Isolation & Dedicated Roles**:
   - Single PostgreSQL database (`graphfolio`) containing isolated schemas: `portfolio` (owned by `portfolio-api` via role `portfolio_svc`) and `users` (owned by `user-api` via role `user_svc`).
   - Strict ban on cross-schema foreign keys, joins, or table grants to preserve independent deployability.
2. **Exact Fixed-Point Numeric Types**:
   - Total prohibition of IEEE 754 floating-point types (`REAL`, `DOUBLE PRECISION`).
   - Quantities use `NUMERIC(28,10)` for micro-fractional shares; prices use `NUMERIC(20,8)`; money amounts use `NUMERIC(20,6)`; exchange rates use `NUMERIC(20,10)`.
3. **Time-Ordered Primary Keys**:
   - Standardize on `UUIDv7` (`DEFAULT uuidv7()`), providing globally unique, sequential identifiers optimized for B-tree index performance and safe API exposure.
4. **Append-Only Ledger & Derived Projections**:
   - `portfolio.transactions` serves as the immutable event ledger.
   - `portfolio.tax_lots`, `portfolio.lot_disposals`, `portfolio.holdings`, and `portfolio.cash_balances` are derived projections, rebuildable via ledger replay.
5. **Database Migration Pipeline**:
   - Schema versioning managed via embedded plain SQL `up`/`down` migrations (`golang-migrate`), running automatically on microservice boot.

---

## 3. Acceptance Criteria

### 3.1 Bootstrap, Schemas & Security Roles
- [x] **AC-1 (Idempotent Bootstrap Script)**: `scripts/db/bootstrap.sql` initializes database `graphfolio`, provisions roles `portfolio_svc` and `user_svc`, and creates schemas `portfolio` and `users` with least-privilege permissions.
- [x] **AC-2 (Zero Cross-Schema Coupling)**: No foreign keys or SQL queries cross schema boundaries between `portfolio` and `users`.

### 3.2 Schema Migrations & Relational Tables
- [x] **AC-3 (Versioned Migration Runner)**: Embedded `golang-migrate` runs migration scripts from `migrations/` on application startup.
- [x] **AC-4 (Reference Data Tables)**: `portfolio.currencies` (ISO 4217), `portfolio.exchanges` (ISO 10383 MIC), and `portfolio.instruments` (with ISIN validation) created and populated.
- [x] **AC-5 (Transaction Ledger Table)**: `portfolio.transactions` supports transaction types enum (`BUY`, `SELL`, `DIVIDEND`, `SPLIT`, `DEPOSIT`, `WITHDRAWAL`, `INTEREST`, `FEE`, `TAX`), transaction dates, exact quantities, unit prices, execution fees, withholding taxes, and external refs.
- [x] **AC-6 (Projection Tables)**: Relational schema includes `portfolio.tax_lots`, `portfolio.lot_disposals`, `portfolio.holdings`, `portfolio.cash_balances`, and `portfolio.portfolio_valuations`.
- [x] **AC-7 (Market Pricing & FX Tables)**: Relational schema includes `portfolio.instrument_prices` and `portfolio.fx_rates` with unique constraints on `(instrument_id, price_date)` and `(base_currency, quote_currency, rate_date)`.

### 3.3 Connection Pooling & Type Registration
- [x] **AC-8 (pgx Connection Pooling)**: `pkg/database/postgres.go` configures `pgxpool.Pool` with connection lifecycle management and automated type registration for `shopspring/decimal.Decimal` on every connection.
- [x] **AC-9 (Development Seeds)**: `seeds/dev_seed.sql` populates development fixtures including demo investors, multi-year transaction histories, and pricing matrices.

---

## 4. Technical & Implementation Notes

- **Scripts**: `scripts/db/bootstrap.sql` and `scripts/db/teardown.sql`.
- **Shared Package**: `pkg/database/` provides DSN sanitization (`SanitizeDSN`), migration runner (`migrate.go`), and pool factory (`postgres.go`).
- **Migrations**: `services/portfolio-api/migrations/` (migrations 000001 through 000010) and `services/user-api/migrations/`.
- **Type Integration**: `github.com/jackc/pgx-shopspring-decimal` registered on `pgxpool.Config.AfterConnect`.

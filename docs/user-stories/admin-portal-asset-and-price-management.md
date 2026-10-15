# User Story: Admin Portal Asset, Price & Ingestion Management Backend

> **Story ID**: US-ADM-101  
> **Epic**: Internal Administration & Market Data Operations  
> **Target**: Admin Portal Application (`web/apps/admin-app`), BFF GraphQL Orchestrator (`bff/`), Portfolio Microservice (`services/portfolio-api`), Protobuf (`proto/portfolio/v1/portfolio.proto`), PostgreSQL (`portfolio` schema)  
> **Status**: **DONE**  
> **Implementation Plan**: [admin-portal-backend-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/admin-portal-backend-implementation-plan.md)  
> **Roadmap Reference**: [FEATURES.md (Feature #10)](file:///Users/oscargarcia/workspace/graphfolio/FEATURES.md#L20)  
> **Standards Reference**: [AGENTS.md](file:///Users/oscargarcia/workspace/graphfolio/AGENTS.md)  

---

## 1. User Story Statements

### 1.1 Master Instrument Directory Management
> **As a** platform operations administrator,  
> **I want to** register, inspect, activate, deactivate, and update tradable instruments in the master asset directory,  
> **So that** investors can search and trade verified assets listing on supported global exchanges with valid MIC and ISO 4217 currencies.

### 1.2 Closing Price Ledger & Manual Overrides
> **As a** financial operations specialist,  
> **I want to** inspect daily asset closing prices and submit audit-justified manual price overrides,  
> **So that** erroneous market data feeds or missing end-of-day quotes can be corrected and affected portfolio valuations recalibrated accurately.

### 1.3 Ingestion Pipeline Diagnostics & On-Demand Synchronization
> **As an** operations engineer,  
> **I want to** monitor the health of external market data ingestion feeds (ECB FX fixings, Twelve Data, Yahoo Finance) and trigger manual syncs,  
> **So that** I can guarantee data freshness, monitor rate limit budgets, and resolve stale prices proactively.

---

## 2. Business Context & Operational Rules

1. **Strict Service Boundary**: All administrative operations run through the internal Admin Portal (`admin-app` on `:5174`) via dedicated gRPC and GraphQL admin APIs, completely segregated from the public investor application.
2. **Authoritative Master Directory**: Instruments must have valid uppercase symbols, recognized exchange MIC codes (`portfolio.exchanges`), ISO 4217 currency codes (`portfolio.currencies`), and valid ISO 6166 ISINs when supplied.
3. **Audit Trail for Price Adjustments**: When an administrator records a price override, the source is flagged as `manual`, and a mandatory reason note is stored for compliance and historical auditability.
4. **Retroactive Valuation Recalibration**: Overriding a price triggers automated recalculation of historical portfolio valuations holding that asset on the affected trade dates.

---

## 3. Acceptance Criteria

### 3.1 Tradable Instrument Management
- [x] **AC-1 (List Instruments)**: Exposes `allInstruments` GraphQL query / `ListAllInstruments` gRPC RPC returning complete instrument metadata (symbol, name, exchange code, currency, ISIN, active status).
- [x] **AC-2 (Register Asset)**: Modal form in `AssetManagement.tsx` submits `createInstrument` validating symbol uniqueness, exchange FK constraint, and currency code.
- [x] **AC-3 (Toggle Status)**: Instant active/inactive toggles allow administrators to freeze assets from new user trading while preserving existing historical ledgers.

### 3.2 Closing Price Ledger & Overrides
- [x] **AC-4 (Price Ledger Inspection)**: `PriceManagement.tsx` queries `instrumentPrices` with symbol filtering and pagination.
- [x] **AC-5 (Manual Price Override)**: Submitting `recordPriceOverride` updates `portfolio.instrument_prices` with `source = 'manual'`, recording `override_reason`.
- [x] **AC-6 (Valuation Recalculation Hook)**: Price overrides trigger portfolio revaluations for any portfolios holding the modified asset on the specified date.

### 3.3 Pipeline Ingestion Diagnostics
- [x] **AC-7 (Feed Health Status)**: `IngestionPipeline.tsx` renders health status cards for ECB FX and Equity feeds displaying last sync timestamp, tracked instrument count, and status indicators (`HEALTHY`, `DEGRADED`, `STALE`).
- [x] **AC-8 (Trigger Market Sync)**: Clicking "Trigger Sync" invokes `triggerMarketSync` mutation, scheduling immediate external feed pulls.

---

## 4. Technical & Implementation Notes

- **Protobuf**: `proto/portfolio/v1/portfolio.proto` defines `ListAllInstruments`, `CreateInstrument`, `UpdateInstrument`, `ListInstrumentPrices`, `RecordPriceOverride`, `GetIngestionStatus`, and `TriggerMarketSync`.
- **Backend Service**: `services/portfolio-api/internal/service/admin.go` coordinates instrument validation, price upserts, and pipeline telemetry.
- **BFF Orchestrator**: `bff/graph/schema.resolvers.go` maps admin GraphQL queries and mutations to gRPC microservices with proper error mapping.
- **Admin UI**: `web/apps/admin-app/src/components/` hosts `AssetManagement.tsx`, `PriceManagement.tsx`, and `IngestionPipeline.tsx`.

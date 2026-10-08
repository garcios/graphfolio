# GraphFolio Features Matrix & Roadmap

This document catalogs all implemented features, in-progress components, and planned future enhancements across GraphFolio. Each entry references its architectural implementation plan, current status, and target layer breakdown.

---

## 1. Feature Status Summary

| # | Feature | Status | Primary Layers | Implementation Plan |
|---|---|---|---|---|
| **1** | **PostgreSQL Multi-Schema Database Architecture** | **DONE** | Database, Migrations, Seed | [db-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/db-implementation-plan.md) |
| **2** | **Portfolio Service & Ledger Projection Engine** | **DONE** | Microservice, Projections | [portfolio-service-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/portfolio-service-implementation-plan.md) |
| **3** | **Exact Fixed-Point Decimal Arithmetic** | **DONE** | Proto, Domain, BFF, Web | [portfolio-service-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/portfolio-service-implementation-plan.md) |
| **4** | **Backend-for-Frontend (BFF) GraphQL Layer** | **DONE** | BFF (`gqlgen`), gRPC Client | [portfolio-service-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/portfolio-service-implementation-plan.md) |
| **5** | **Investor Dashboard Web Application** | **DONE** | Web (React, Vite, GenQL) | [portfolio-service-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/portfolio-service-implementation-plan.md) |
| **6** | **Interactive Transaction Ingestion Modal** | **DONE** | Proto, Svc, BFF, Web | [add-transactions-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/add-transactions-implementation-plan.md) |
| **7** | **Interactive SVG Performance Chart & Time Range Filtering** | **DONE** | Proto, Svc, BFF, Web | [performance-chart-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/performance-chart-implementation-plan.md) |
| **8** | **Transaction History & Ledger Management** | **DONE** | Proto, Svc, BFF, Web | [transaction-history-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/transaction-history-implementation-plan.md) |
| **9** | **Frontend Workspace Architecture (Main App & Admin Portal)** | **DONE** | Web (Monorepo, React, Vite), BFF | [web-workspace-refactoring-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/web-workspace-refactoring-plan.md) |
| **10** | **Admin Portal Backend & BFF (Asset, Price & Ingestion Management)** | **DONE** | Proto, Svc, BFF, Web | [admin-portal-backend-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/admin-portal-backend-implementation-plan.md) |
| **11** | **Dynamic Cost Basis Method Switching (Average Cost ↔ FIFO)** | **PLANNED** | Proto, Svc, BFF, Web | [cost-basis-switching-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/cost-basis-switching-implementation-plan.md) |
| **12** | **User Preferences & Display Currency** | **DONE** | Proto, User API, BFF, Web | [user-preferences-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/user-preferences-implementation-plan.md) |
| **13** | **Portfolio Valuation Engine & Historical Backfill** | **DONE** | Database, Proto, Svc, Worker | [portfolio-valuation-engine-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/portfolio-valuation-engine-implementation-plan.md) |
| **14** | **Market Data Ingestion (Asset Prices & FX Rates)** | **DONE** | Database, Svc, Worker | [market-data-ingestion-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/market-data-ingestion-implementation-plan.md) |
| **15** | **Look-Through Fundamental & Cash Flow Quality Engine** | **PLANNED** | Database, Proto, Svc, BFF, Web | [fundamental-cash-flow-engine-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/fundamental-cash-flow-engine-implementation-plan.md) |
| **16** | **Currency Pair Historical Prices (Admin Portal)** | **DONE** | Proto, Svc, BFF, Web | [currency-pair-historical-prices-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/currency-pair-historical-prices-implementation-plan.md) |
| **17** | **Admin UI Historical Market Data Backfill (Asset Prices & FX Rates)** | **DONE** | Proto, Svc, BFF, Web | [admin-historical-backfill-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/admin-historical-backfill-implementation-plan.md) |
| **18** | **Delete Asset & Price Cascade (Admin Portal)** | **DONE** | Database, Proto, Svc, BFF, Web | [delete-asset-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/delete-asset-implementation-plan.md) |
| **19** | **Ingestion Job History View (Admin Portal)** | **PLANNED** | Database, Proto, Svc, BFF, Web | [ingestion-job-history-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/ingestion-job-history-implementation-plan.md) |
| **20** | **Tax Lot Inspector & Capital Gains Reports** | **ROADMAP** | Svc, BFF, Web | *(Future Plan)* |
| **21** | **Dividend Calendar & Yield Analytics** | **ROADMAP** | Svc, BFF, Web | *(Future Plan)* |
| **22** | **Real-Time Market Data & WebSocket Price Ticker** | **ROADMAP** | Market Data, Svc, Web | *(Future Plan)* |

---

## 2. Completed Features (`DONE`)

### 2.1 PostgreSQL Multi-Schema Database Architecture
- **Status**: **DONE**
- **Plan Reference**: [db-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/db-implementation-plan.md)
- **Description**: Robust relational foundation on PostgreSQL 17+ featuring schema isolation (`portfolio` and `users` schemas) with dedicated service roles (`portfolio_svc`, `user_svc`). Includes 6 versioned schema migrations (`golang-migrate`), UUIDv7 native primary keys, foreign currency constraints, transaction types enum, tax lot tables, and projection storage.
- **Key Files**:
  - [bootstrap.sql](file:///Users/oscargarcia/workspace/graphfolio/scripts/db/bootstrap.sql)
  - [migrations/000001_reference_data.up.sql](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/migrations/000001_reference_data.up.sql)
  - [migrations/000006_projections.up.sql](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/migrations/000006_projections.up.sql)
  - [dev_seed.sql](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/seeds/dev_seed.sql)

### 2.2 Portfolio Service & Deterministic Ledger Projections
- **Status**: **DONE**
- **Plan Reference**: [portfolio-service-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/portfolio-service-implementation-plan.md)
- **Description**: Standalone Go gRPC microservice (`services/portfolio-api` on `:50051`) implementing an event-sourced ledger projection engine. Holding quantities, cost bases, cash balances, and valuations are derived deterministically by replaying historical transactions through `service.ProcessLedger` with atomic row-level locking (`SELECT ... FOR UPDATE`).
- **Key Files**:
  - [server.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/server.go)
  - [projection.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/service/projection.go)
  - [calculator.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/domain/calculator.go)
  - [postgres.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/repository/postgres.go)

### 2.3 Exact Fixed-Point Decimal Arithmetic
- **Status**: **DONE**
- **Plan Reference**: [portfolio-service-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/portfolio-service-implementation-plan.md)
- **Description**: Strict zero floating-point arithmetic policy. Every financial amount, price, quantity, and percentage is represented with exact arbitrary precision using `shopspring/decimal.Decimal` in Go, custom `common.v1.Decimal` / `common.v1.Money` protobuf types, and a custom string-serialized `Decimal` scalar in GraphQL.
- **Key Files**:
  - [decimal.proto](file:///Users/oscargarcia/workspace/graphfolio/proto/common/v1/decimal.proto)
  - [decimalpb.go](file:///Users/oscargarcia/workspace/graphfolio/pkg/decimalpb/decimalpb.go)
  - [decimal.go](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/model/decimal.go)

### 2.4 Backend-for-Frontend (BFF) GraphQL Layer
- **Status**: **DONE**
- **Plan Reference**: [portfolio-service-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/portfolio-service-implementation-plan.md)
- **Description**: High-performance GraphQL orchestrator built with Go and `gqlgen` running on `:8080`. Connects to gRPC microservices, stitches domain models together, exposes strongly-typed queries and mutations, and handles CORS for frontend clients.
- **Key Files**:
  - [schema.graphqls](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.graphqls)
  - [schema.resolvers.go](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.resolvers.go)
  - [helpers.go](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/helpers.go)
  - [main.go](file:///Users/oscargarcia/workspace/graphfolio/bff/cmd/server/main.go)

### 2.5 Investor Dashboard Web Application
- **Status**: **DONE**
- **Plan Reference**: [portfolio-service-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/portfolio-service-implementation-plan.md) & [genql-usage.md](file:///Users/oscargarcia/workspace/graphfolio/docs/genql-usage.md)
- **Description**: Glassmorphic, dark-mode single-page application built with React 19, TypeScript, and Vite. Queries the BFF using an auto-generated type-safe client (`genql`), rendering total portfolio valuation, day's gain/loss, annualized returns, cash balance, and a comprehensive holdings table with currency formatting.
- **Key Files**:
  - [Dashboard.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/main-app/src/components/Dashboard.tsx)
  - [Dashboard.css](file:///Users/oscargarcia/workspace/graphfolio/web/apps/main-app/src/components/Dashboard.css)
  - [main.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/main-app/src/main.tsx)

### 2.6 Interactive Transaction Ingestion Modal
- **Status**: **DONE**
- **Plan Reference**: [add-transactions-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/add-transactions-implementation-plan.md)
- **Description**: End-to-end transaction entry flow allowing investors to record `BUY`, `SELL`, `DIVIDEND`, `DEPOSIT`, and `WITHDRAWAL` transactions. Features dynamic instrument lookup, trade date validation, cash balance checks, automatic projection rebuilds, and instantaneous UI refresh without full page reload.
- **Key Files**:
  - [portfolio.proto](file:///Users/oscargarcia/workspace/graphfolio/proto/portfolio/v1/portfolio.proto)
  - [AddTransactionModal.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/main-app/src/components/AddTransactionModal.tsx)
  - [AddTransactionModal.css](file:///Users/oscargarcia/workspace/graphfolio/web/apps/main-app/src/components/AddTransactionModal.css)
  - [transaction.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/service/transaction.go)

### 2.7 Interactive SVG Performance Chart & Time Range Filtering
- **Status**: **DONE**
- **Plan Reference**: [performance-chart-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/performance-chart-implementation-plan.md)
- **Description**: Replaced static placeholder chart with an interactive, data-driven SVG performance curve. Driven by time-series snapshots in `portfolio.portfolio_valuations` (with a 365-day seed series), queried across `1D`, `1W`, `1M`, `1Y`, and `ALL` timeframes. Features smooth Catmull-Rom Bezier curves, dynamic gradient glows, hover crosshairs with micro-animations, and floating glassmorphism tooltips showing total value, market/cash splits, and cumulative return.
- **Key Files**:
  - [portfolio.proto](file:///Users/oscargarcia/workspace/graphfolio/proto/portfolio/v1/portfolio.proto)
  - [history.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/service/history.go)
  - [schema.graphqls](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.graphqls)
  - [schema.resolvers.go](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.resolvers.go)
  - [PerformanceChart.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/main-app/src/components/PerformanceChart.tsx)
  - [PerformanceChart.css](file:///Users/oscargarcia/workspace/graphfolio/web/apps/main-app/src/components/PerformanceChart.css)

### 2.8 Transaction History & Ledger Management
- **Status**: **DONE**
- **Plan Reference**: [transaction-history-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/transaction-history-implementation-plan.md)
- **Description**: Full transaction ledger history with server-side pagination, multi-criteria filtering by transaction type and symbol, and safe transaction deletion with automatic deterministic projection replay (`RebuildProjections`). Deleting an entry atomically clears derived lots, disposals, holdings, and cash, re-running the ledger to update metrics and UI immediately.
- **Key Files**:
  - [portfolio.proto](file:///Users/oscargarcia/workspace/graphfolio/proto/portfolio/v1/portfolio.proto)
  - [transaction.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/service/transaction.go)
  - [schema.graphqls](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.graphqls)
  - [schema.resolvers.go](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.resolvers.go)
  - [TransactionLedger.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/main-app/src/components/TransactionLedger.tsx)
  - [TransactionLedger.css](file:///Users/oscargarcia/workspace/graphfolio/web/apps/main-app/src/components/TransactionLedger.css)

### 2.9 Frontend Workspace Architecture (`main-app` & `admin-app`)
- **Status**: **DONE**
- **Plan Reference**: [web-workspace-refactoring-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/web-workspace-refactoring-plan.md)
- **Description**: Transitioned `web/` to an npm workspace monorepo supporting multiple frontend applications. Cleanly decouples the primary user-facing application (`main-app` on port 5173) and the internal administrative portal (`admin-app` on port 5174) while eliminating configuration duplication through shared `tsconfig.base.json`, Oxlint configs, and Vite bundler settings.
- **Key Packages & Files**:
  - **Shared Design System**: [`@graphfolio/ui`](file:///Users/oscargarcia/workspace/graphfolio/web/packages/ui) with tokens, reset, formatters, and atomic components (`Button`, `Modal`, `Card`, `Badge`, `Table`, `Input`, `Select`).
  - **Shared API Client**: [`@graphfolio/api-client`](file:///Users/oscargarcia/workspace/graphfolio/web/packages/api-client) with GenQL generated client and singleton factory.
  - **Primary Application**: [`@graphfolio/main-app`](file:///Users/oscargarcia/workspace/graphfolio/web/apps/main-app)
  - **Internal Admin Portal**: [`@graphfolio/admin-app`](file:///Users/oscargarcia/workspace/graphfolio/web/apps/admin-app) (Asset Management, Price Management, Ingestion Pipeline)
  - [Makefile](file:///Users/oscargarcia/workspace/graphfolio/Makefile) targets: `run-web`, `run-admin`, `run-all-web`, `build-web`, `install-web`

### 2.10 Admin Portal Backend & BFF (Asset, Price & Ingestion Management)
- **Status**: **DONE**
- **Plan Reference**: [admin-portal-backend-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/admin-portal-backend-implementation-plan.md)
- **Description**: Full administrative operations and market orchestration backend. Extends `portfolio-api` and GraphQL BFF with master instrument directory management (asset listing, creation with exchange code dropdown populated from `portfolio.exchanges` reference data, ISO 6166 ISIN validation, active/inactive toggles), closing price ledger queries (`portfolio.instrument_prices`) and manual price overrides with audit justifications and retroactive valuation recalibrations, and ingestion pipeline diagnostics reporting feed health, ECB FX fixings, rate limit budgets, and on-demand market data synchronization. Connected directly to `@graphfolio/admin-app`.
- **Key Files**:
  - [portfolio.proto](file:///Users/oscargarcia/workspace/graphfolio/proto/portfolio/v1/portfolio.proto)
  - [admin.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/service/admin.go)
  - [postgres.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/repository/postgres.go)
  - [schema.graphqls](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.graphqls)
  - [schema.resolvers.go](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.resolvers.go)
  - [AddInstrumentModal.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/admin-app/src/components/AddInstrumentModal.tsx)
  - [AssetManagement.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/admin-app/src/components/AssetManagement.tsx)
  - [PriceManagement.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/admin-app/src/components/PriceManagement.tsx)
  - [IngestionPipeline.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/admin-app/src/components/IngestionPipeline.tsx)

### 2.11 Market Data Ingestion Pipeline (Daily EOD Pricing, ECB FX & Backfill Synchronizer)
- **Status**: **DONE**
- **Plan Reference**: [market-data-ingestion-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/market-data-ingestion-implementation-plan.md)
- **Description**: Automated EOD price and foreign exchange synchronization pipeline. Implements European Central Bank (ECB) daily and 90-day XML reference fixing feed parser with exact 10-decimal triangulation, resilient Twelve Data primary and Yahoo Finance fallback equity price adapters, token-bucket rate limiting (`golang.org/x/time/rate`) with exponential backoff and randomized jitter on HTTP 429/5xx, automated transaction ingestion backfill hooks that detect unpriced asset ranges upon trade entry, and a standalone scheduled CLI worker (`cmd/market-ingest/main.go`) runnable via `make ingest-market-data`.
- **Key Files**:
  - [provider.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/marketdata/provider.go)
  - [validation.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/marketdata/validation.go)
  - [triangulation.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/marketdata/triangulation.go)
  - [limiter.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/marketdata/limiter.go)
  - [ecb_provider.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/marketdata/ecb_provider.go)
  - [equity_provider.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/marketdata/equity_provider.go)
  - [ingestion.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/service/ingestion.go)
  - [transaction.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/service/transaction.go)
  - [main.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/cmd/market-ingest/main.go)
  - [Makefile](file:///Users/oscargarcia/workspace/graphfolio/Makefile) target: `make ingest-market-data`

### 2.12 User Preferences & Display Currency
- **Status**: **DONE**
- **Plan Reference**: [user-preferences-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/user-preferences-implementation-plan.md)
- **Description**: User profile identity and personal preference management system. Features the dedicated `services/user-api` gRPC microservice on `:50052` backed by the isolated `users` PostgreSQL schema (`users.users`), managing display name, theme (`DARK`, `LIGHT`, `SYSTEM`), and base display currency. Integrated via BFF GraphQL queries (`userPreferences`, `supportedCurrencies`) and mutation (`updateUserPreferences`), with dynamic portfolio revaluation synchronization. Investor frontend features the interactive `UserPreferencesModal` accessible from the dashboard profile badge, multi-currency rich dropdown (`USD`, `EUR`, `GBP`, `AUD`, `CAD`, `JPY`, `CHF`), segmented theme picker, live character counter, and real-time toast feedback.
- **Key Files**:
  - [user.proto](file:///Users/oscargarcia/workspace/graphfolio/proto/user/v1/user.proto)
  - [000002_add_theme.up.sql](file:///Users/oscargarcia/workspace/graphfolio/services/user-api/migrations/000002_add_theme.up.sql)
  - [dev_seed.sql](file:///Users/oscargarcia/workspace/graphfolio/services/user-api/seeds/dev_seed.sql)
  - [user.go](file:///Users/oscargarcia/workspace/graphfolio/services/user-api/internal/domain/user.go)
  - [repository.go](file:///Users/oscargarcia/workspace/graphfolio/services/user-api/internal/repository/repository.go)
  - [postgres.go](file:///Users/oscargarcia/workspace/graphfolio/services/user-api/internal/repository/postgres.go)
  - [service.go](file:///Users/oscargarcia/workspace/graphfolio/services/user-api/internal/service/service.go)
  - [server.go](file:///Users/oscargarcia/workspace/graphfolio/services/user-api/internal/server.go)
  - [schema.graphqls](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.graphqls)
  - [schema.resolvers.go](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.resolvers.go)
  - [UserPreferencesModal.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/main-app/src/components/UserPreferencesModal.tsx)
  - [UserPreferencesModal.css](file:///Users/oscargarcia/workspace/graphfolio/web/apps/main-app/src/components/UserPreferencesModal.css)
  - [Dashboard.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/main-app/src/components/Dashboard.tsx)

### 2.13 Portfolio Valuation Engine & Historical Backfill
- **Status**: **DONE**
- **Plan Reference**: [portfolio-valuation-engine-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/portfolio-valuation-engine-implementation-plan.md)
- **Description**: Comprehensive portfolio valuation engine, mathematical analytics, and historical time-series backfill system. Evaluates exact daily valuation snapshots ($V_t = \text{Market Value} + \text{Cash Value}$, $F_t = \text{Deposits} - \text{Withdrawals}$, $R_t = \frac{V_t - (V_{t-1} + F_t)}{V_{t-1} + F_t}$, and cumulative Time-Weighted Return $\text{TWR}_t$ geometric linking index) with zero floating-point drift using `shopspring/decimal`. Features GIPS-compliant Annualized Return (CAGR) calculations, repository batch upsert routines (`UpsertValuationsBatch`), and matrix queries with Last Observation Carried Forward (LOCF) for missing asset closing prices and triangulated FX rates. Includes automatic transaction ingestion hooks triggering historical replay on past-dated transactions and current-day snapshots, valuation backfill upon transaction deletion, gRPC `RebuildValuations` endpoint, and a dedicated background CLI / daemon worker (`cmd/worker/main.go`, `make run-valuation-job`) scheduled for daily market close (22:00 UTC).
- **Key Files**:
  - [portfolio.proto](file:///Users/oscargarcia/workspace/graphfolio/proto/portfolio/v1/portfolio.proto)
  - [valuation.go (Domain)](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/domain/valuation.go)
  - [valuation_test.go (Domain)](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/domain/valuation_test.go)
  - [queries.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/repository/queries.go)
  - [postgres.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/repository/postgres.go)
  - [valuation.go (Service)](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/service/valuation.go)
  - [valuation_test.go (Service)](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/service/valuation_test.go)
  - [transaction.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/service/transaction.go)
  - [server.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/server.go)
  - [server_test.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/server_test.go)
  - [main.go (Worker)](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/cmd/worker/main.go)
  - [Makefile](file:///Users/oscargarcia/workspace/graphfolio/Makefile)

### 2.12 Currency Pair Historical Prices (Admin Portal)
- **Status**: **DONE**
- **Plan Reference**: [currency-pair-historical-prices-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/currency-pair-historical-prices-implementation-plan.md)
- **Description**: Foreign Exchange (FX) currency pair inspection, reciprocal rate triangulation, historical time-series analytics, and manual pricing override system. Features a dedicated "Currency Pairs (FX)" portal view in the internal Admin Portal (`:5174`), interactive SVG historical trend curves with glowing area fills across `1W`, `1M`, `1Y`, and `ALL` timeframes, instant direct/reciprocal inversion toggle (`Base ⇄ Quote`), real-time operational telemetry (Spot Rate, 1D delta/percentage, and benchmark provider badges), paginated authoritative ledger (`portfolio.fx_rates`), and modal dialogs for audit-justified manual rate overrides. End-to-end gRPC RPCs, GraphQL BFF resolvers, and typed GenQL client integration strictly adhering to the zero floating-point arithmetic policy.
- **Key Files**:
  - [portfolio.proto](file:///Users/oscargarcia/workspace/graphfolio/proto/portfolio/v1/portfolio.proto)
  - [fx.go (Domain)](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/domain/fx.go)
  - [queries.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/repository/queries.go)
  - [postgres.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/repository/postgres.go)
  - [fx.go (Service)](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/service/fx.go)
  - [fx_test.go (Service)](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/service/fx_test.go)
  - [server.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/server.go)
  - [server_test.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/server_test.go)
  - [schema.graphqls](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.graphqls)
  - [schema.resolvers.go](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.resolvers.go)
  - [helpers.go](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/helpers.go)
  - [FXManagement.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/admin-app/src/components/FXManagement.tsx)
  - [FXTrendChart.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/admin-app/src/components/FXTrendChart.tsx)
  - [FXOverrideModal.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/admin-app/src/components/FXOverrideModal.tsx)
  - [AdminLayout.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/admin-app/src/components/AdminLayout.tsx)

### 2.17 Admin UI Historical Market Data Backfill (Asset Prices & FX Rates)
- **Status**: **DONE**
- **Plan Reference**: [admin-historical-backfill-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/admin-historical-backfill-implementation-plan.md)
- **Scope**: `proto/portfolio/v1/`, `services/portfolio-api`, `bff/`, `web/apps/admin-app`
- **Description**: Interactive multi-day historical market data backfill engine orchestrating batch ingestion of daily closing prices and foreign exchange reference fixings over custom date spans. Features an interactive, glassmorphic modal (`BackfillModal.tsx`) with dynamic date preset chips (`30D`, `90D`, `YTD`, `1Y`, `ALL`), granular scope selection (All Active vs Specific Symbols / Pairs), real-time API token budget impact projections, and optional automated retroactive valuation reconciliation. Screen integrations across `IngestionPipeline` (Backfill Console), `PriceManagement` ("Backfill Asset Prices" toolbar action), and `FXManagement` ("Backfill FX Rates" toolbar action). End-to-end gRPC `TriggerBackfill` RPC, GraphQL `triggerBackfill` mutation, and typed GenQL client integration strictly adhering to the zero floating-point arithmetic policy.
- **Key Files**:
  - [portfolio.proto](file:///Users/oscargarcia/workspace/graphfolio/proto/portfolio/v1/portfolio.proto)
  - [backfill.go (Domain)](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/domain/backfill.go)
  - [admin.go (Service)](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/service/admin.go)
  - [admin_test.go (Service)](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/service/admin_test.go)
  - [server.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/server.go)
  - [server_test.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/server_test.go)
  - [schema.graphqls](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.graphqls)
  - [schema.resolvers.go](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.resolvers.go)
  - [schema.resolvers_test.go](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.resolvers_test.go)
  - [BackfillModal.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/admin-app/src/components/BackfillModal.tsx)
  - [BackfillModal.css](file:///Users/oscargarcia/workspace/graphfolio/web/apps/admin-app/src/components/BackfillModal.css)
  - [IngestionPipeline.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/admin-app/src/components/IngestionPipeline.tsx)
  - [PriceManagement.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/admin-app/src/components/PriceManagement.tsx)
  - [FXManagement.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/admin-app/src/components/FXManagement.tsx)

### 2.15 Delete Asset & Price Cascade (Admin Portal)
- **Status**: **DONE**
- **Plan Reference**: [delete-asset-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/delete-asset-implementation-plan.md)
- **Description**: Enables operations administrators to permanently remove tradable assets from the GraphFolio master directory. Automatically cascades deletion to all historical closing marks in `portfolio.instrument_prices` within an atomic database transaction. Protects referential integrity by rejecting deletion if an asset is held or actively referenced by user transactions (`ErrInstrumentInUse`, `codes.FailedPrecondition`). Includes safety confirmation modal in the admin web UI.
- **Key Files**:
  - [000007_cascade_instrument_prices.up.sql](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/migrations/000007_cascade_instrument_prices.up.sql)
  - [portfolio.proto](file:///Users/oscargarcia/workspace/graphfolio/proto/portfolio/v1/portfolio.proto)
  - [repository.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/repository/repository.go)
  - [queries.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/repository/queries.go)
  - [postgres.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/repository/postgres.go)
  - [admin.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/service/admin.go)
  - [server.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/server.go)
  - [schema.graphqls](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.graphqls)
  - [schema.resolvers.go](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.resolvers.go)
  - [DeleteAssetModal.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/admin-app/src/components/DeleteAssetModal.tsx)
  - [DeleteAssetModal.css](file:///Users/oscargarcia/workspace/graphfolio/web/apps/admin-app/src/components/DeleteAssetModal.css)
  - [AssetManagement.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/admin-app/src/components/AssetManagement.tsx)

---

## 3. Planned Features (`PLANNED`)

### 3.1 Dynamic Cost Basis Method Switching (Average Cost ↔ FIFO)
- **Status**: **PLANNED**
- **Plan Reference**: [cost-basis-switching-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/cost-basis-switching-implementation-plan.md)
- **Scope**: `proto/`, `services/portfolio-api`, `bff/`, `web/`
- **Highlights**:
  - Segmented toggle control: `[ Average Cost | FIFO ]` in dashboard header.
  - Dynamically switches `portfolio.portfolios.cost_basis_method`.
  - Replays historical transaction ledger to instantly update realized capital gains, holding cost basis, and total return percentages.
  - Info tooltip explaining tax optimization differences (pooling vs selling oldest shares).

### 3.2 Look-Through Fundamental & Cash Flow Quality Engine
- **Status**: **PLANNED**
- **Plan Reference**: [fundamental-cash-flow-engine-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/fundamental-cash-flow-engine-implementation-plan.md)
- **Scope**: `proto/`, `services/portfolio-api`, `bff/`, `web/`
- **Highlights**:
  - Table `portfolio.instrument_fundamentals` tracking GAAP/IFRS audited line items (NOPAT, Invested Capital, OCF, Maintenance CapEx, Diluted Shares, NIBCL).
  - Proportional look-through calculations: investor's exact share of Operating Cash Flow, Owner Earnings, FCF, and Revenue.
  - Portfolio Capital Allocation Scorecard: Weighted Average ROIC, Reinvestment Rate, Intrinsic Compounding Rate ($g = \text{ROIC} \times \text{RR}$), and Economic Value Added spread ($\text{ROIC} - \text{WACC}$).
  - Earnings quality detection via the Sloan Accrual Ratio to highlight divergence between reported Net Income and cash realization.
  - "Business Owner" dashboard view contrasting Market Value growth against Business Intrinsic Value growth.

### 3.3 Ingestion Job History View (Admin Portal)
- **Status**: **PLANNED**
- **Plan Reference**: [ingestion-job-history-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/ingestion-job-history-implementation-plan.md)
- **Scope**: Database (`portfolio.ingestion_jobs`), `proto/portfolio/v1/`, `services/portfolio-api`, `bff/`, `web/apps/admin-app`
- **Highlights**:
  - PostgreSQL schema table `portfolio.ingestion_jobs` tracking execution lifecycle, timestamps, record counts, and errors.
  - Dedicated historical ledger view (`IngestionJobHistory.tsx`) and tab integration in `IngestionPipeline` with auto-refresh.
  - Exact UTC timestamps for execution start, finish, and computed elapsed duration (e.g. `4.2s`, `1m 24s`, or `"Running (15s)..."`).
  - Multi-state visual badges: `SUCCESS` (green), `FAILED` (red), `PARTIAL_SUCCESS` (amber), and `IN_PROGRESS` (pulsing blue).
  - Granular record split metrics displaying successful vs failed rows (`420 ok / 12 fail`).
  - Server-side sorting by start date (newest first by default) with single-pass `COUNT(*) OVER()` pagination.
  - Interactive drawer/modal to inspect error diagnostics, stack traces, and execution parameters for failed imports.

---

## 4. Future Roadmap Ideas (`ROADMAP`)

### 4.1 Tax Lot Inspector & Capital Gains Reports
- **Status**: **ROADMAP**
- **Description**: Detailed inspection of individual tax lots (`portfolio.tax_lots`) and closed lot disposals (`portfolio.lot_disposals`). Enables investors to see which specific lot was sold, purchase date, holding period, and short-term vs long-term capital gains classification for annual tax filing.

### 4.2 Dividend Calendar & Yield Analytics
- **Status**: **ROADMAP**
- **Description**: Monthly dividend income projections, ex-dividend dates calendar, yield on cost (YOC), and automated dividend reinvestment plan (DRIP) modeling.

### 4.3 Real-Time Market Data & WebSocket Price Ticker
- **Status**: **ROADMAP**
- **Description**: Live market data feed via WebSockets for real-time portfolio valuation updates, intraday price fluctuations, and automated FX rate synchronization.

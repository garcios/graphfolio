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
| **20** | **Asset Prices Date Range Filter & Pagination (Admin Portal)** | **DONE** | Web, BFF, Svc | [asset-prices-date-range-filter-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/asset-prices-date-range-filter-implementation-plan.md) |
| **21** | **Multi-Broker CSV Ingestion & Cash Account Statements (CommSec & nabtrade)** | **DONE** | Proto, Svc, BFF, Web | [multi-broker-csv-ingestion-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/multi-broker-csv-ingestion-implementation-plan.md), [cash-statement-csv-ingestion-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/cash-statement-csv-ingestion-implementation-plan.md) |
| **22** | **Tax Lot Inspector & Capital Gains Reports** | **ROADMAP** | Svc, BFF, Web | *(Future Plan)* |
| **23** | **Dividend Calendar & Yield Analytics** | **ROADMAP** | Svc, BFF, Web | *(Future Plan)* |
| **24** | **Real-Time Market Data & WebSocket Price Ticker** | **ROADMAP** | Market Data, Svc, Web | *(Future Plan)* |
| **25** | **Investment Holding Return Attribution (Capital Gain, Income & Currency Gain/Loss)** | **DONE** | Proto, Svc, BFF, Web | [holding-return-attribution-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/holding-return-attribution-implementation-plan.md) |
| **26** | **Average Buy Price Holdings Column** | **DONE** | Proto, Svc, BFF, Web | [average-buy-price-column-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/average-buy-price-column-implementation-plan.md) |
| **27** | **Portfolio Cash Flow Report Generation** | **DONE** | Proto, Svc, BFF, Web | [portfolio-cash-flow-report-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/portfolio-cash-flow-report-implementation-plan.md) |
| **28** | **GIPS CAGR Annualized Return & Valuation Cost-Basis Resiliency** | **DONE** | Database, Svc, Domain, Worker | [annualized-return-twr-fix-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/annualized-return-twr-fix-implementation-plan.md) |
| **29** | **Missing Prices Report & Date Range Backfill (Admin Portal)** | **DONE** | Proto, Svc, BFF, Web | [admin-missing-prices-report-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/admin-missing-prices-report-implementation-plan.md) |
| **30** | **Include Closed Positions in Dashboard Investments Table** | **DONE** | Proto, Svc, BFF, Web | [include-closed-positions-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/include-closed-positions-implementation-plan.md) |
| **31** | **Modified Dietz Annualized Return & Cash-Flow Weighting** | **DONE** | Proto, Svc, BFF, Web | [annualized-return-modified-dietz-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/annualized-return-modified-dietz-implementation-plan.md) |
| **32** | **Missing FX Rates Report & Date Range Backfill (Admin Portal)** | **DONE** | Proto, Svc, BFF, Web | [admin-missing-fx-rates-report-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/admin-missing-fx-rates-report-implementation-plan.md) |
| **33** | **Portfolio Benchmark Comparison & Relative Performance Analytics (S&P 500, NASDAQ-100, ASX 200)** | **DONE** | Database, Proto, Svc, BFF, Web | [benchmark-performance-comparison-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/benchmark-performance-comparison-implementation-plan.md) |
| **34** | **Download Transactions from Transaction Ledger (Ascending Order)** | **DONE** | Proto, Svc, BFF, Web | [download-transactions-ledger-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/download-transactions-ledger-implementation-plan.md) |
| **35** | **User Registration, Authentication & Session Management (Main App)** | **DONE** | Proto, User API, BFF, Web | [user-registration-and-login-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/user-registration-and-login-implementation-plan.md) |
| **36** | **User Password Change & Session Invalidation (Main App)** | **DONE** | Proto, User API, BFF, Web | [user-change-password-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/user-change-password-implementation-plan.md) |
| **37** | **In-Place Transaction Editing in Transaction Ledger** | **PLANNED** | Proto, Svc, BFF, Web | [edit-transaction-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/edit-transaction-implementation-plan.md) |
| **38** | **Admin Portal Authentication, Auto-Provisioning & Mandatory First-Login Password Rotation** | **DONE** | Database, Proto, User API, BFF, Web | [admin-portal-login-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/admin-portal-login-implementation-plan.md) |
| **39** | **Bidirectional Date Sorting in Transaction Ledger (`DESC` ↔ `ASC`)** | **DONE** | Proto, Svc, BFF, Web | [sort-transactions-by-date-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/sort-transactions-by-date-implementation-plan.md) |
| **40** | **Create Cash Withdrawal Against Dividend Transaction** | **DONE** | Web, BFF, Svc | [withdraw-against-dividend-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/withdraw-against-dividend-implementation-plan.md) |
| **41** | **Filter Transactions by Interest Type in Transaction Ledger** | **DONE** | Web, BFF, Svc, Proto, DB | [filter-transactions-by-interest-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/filter-transactions-by-interest-implementation-plan.md) |
| **42** | **Record Cash Interest Transactions (Yield Ingestion & Cash Account Reconciliation)** | **DONE** | Proto, Svc, BFF, Web | [record-interest-transaction-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/record-interest-transaction-implementation-plan.md) |
| **43** | **User Preferences — End of Financial Year (EOFY) Configuration** | **DONE** | Database, Proto, User API, BFF, Web | [user-preferences-financial-year-end-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/user-preferences-financial-year-end-implementation-plan.md) |



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
- **Description**: End-to-end transaction entry flow allowing investors to record `BUY`, `SELL`, `DIVIDEND`, `SPLIT`, `DEPOSIT`, and `WITHDRAWAL` transactions. Features dynamic instrument lookup, custom brokerage fee currency selection (defaulting to the investor's preferred display currency with exact multi-currency cash balance & tax lot cost basis adjustments), stock split ratio input with popular presets (2:1, 3:1, 4:1, 10:1, reverse splits) preserving total cost basis, trade date validation, automatic projection rebuilds, and instantaneous UI refresh without full page reload.
- **Key Files**:
  - [portfolio.proto](file:///Users/oscargarcia/workspace/graphfolio/proto/portfolio/v1/portfolio.proto)
  - [AddTransactionModal.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/main-app/src/components/AddTransactionModal.tsx)
  - [AddTransactionModal.css](file:///Users/oscargarcia/workspace/graphfolio/web/apps/main-app/src/components/AddTransactionModal.css)
  - [transaction.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/service/transaction.go)
  - [projection.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/service/projection.go)

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

### 2.16 Asset Prices Date Range Filter & Pagination (Admin Portal)
- **Status**: **DONE**
- **Plan Reference**: [asset-prices-date-range-filter-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/asset-prices-date-range-filter-implementation-plan.md)
- **Scope**: `web/apps/admin-app`, `bff/`, `services/portfolio-api`
- **Description**: Upgrades the Closing Prices Ledger (`PriceManagement`) in the Admin Portal with dual date range filtering (`fromDate` and `toDate`), quick range preset pill buttons (`7D`, `30D`, `90D`, `YTD`, `1Y`), client-side date range validation guard (`fromDate > toDate`), and full multi-page pagination controls (`page`, `pageSize = 50`, `totalPages`, Prev/Next navigation, slice record counter). Includes integration with `BackfillModal` passing active date range bounds, and unit test suites across `portfolio-api` (service clamping/bounds, gRPC parsing) and BFF GraphQL resolvers verifying bounded, half-open, and paginated price queries.
- **Key Files**:
  - [PriceManagement.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/admin-app/src/components/PriceManagement.tsx)
  - [PriceManagement.css](file:///Users/oscargarcia/workspace/graphfolio/web/apps/admin-app/src/components/PriceManagement.css)
  - [admin.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/service/admin.go)
  - [admin_test.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/service/admin_test.go)
  - [server.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/server.go)
  - [server_test.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/server_test.go)
  - [schema.resolvers_test.go](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.resolvers_test.go)

### 2.21 Multi-Broker CSV Ingestion & Cash Account Statements (CommSec & nabtrade)
- **Status**: **DONE**
- **Plan Reference**: [multi-broker-csv-ingestion-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/multi-broker-csv-ingestion-implementation-plan.md), [cash-statement-csv-ingestion-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/cash-statement-csv-ingestion-implementation-plan.md)
- **Scope**: `proto/portfolio/v1/`, `services/portfolio-api`, `bff/`, `web/apps/main-app`
- **Description**: Robust end-to-end multi-broker CSV transaction import supporting leading Australian retail brokers (CommSec and nabtrade). Features automated client-side header signature detection with manual override, broker dialect parsers handling Australian date formatting (`DD/MM/YYYY`), currency cleansing (`$`, thousand commas, accounting negative parentheses), and nabtrade metadata header/footer stripping. Incorporates full cash account statement ingestion (`Date,Type,Description,Debit,Credit,Balance`), automatically capturing cash deposits (`DEPOSIT`), cash withdrawals (`WITHDRAWAL`), monthly cash interest (`INTEREST`), and domestic ETF/stock dividends (`FUNDS TRANSFER DIVIDEND`) with ticker alias mapping, while cleanly ignoring policy notifications (`InterestChange`) to eliminate negative cash balances. Supports empty-symbol cash imports bypassing instrument catalog lookups, idempotent hash generation (`SHA-256`) and native trade confirmation number mapping into `portfolio.transactions.external_ref`, pre-flight duplicate checking against existing portfolio records via GraphQL BFF, and an interactive glassmorphic modal (`ImportTransactionsModal.tsx`) with summary metric cards, tabbed data preview, and line-level error inspection. Microservice backend implements atomic batch persistence via `pgx.Batch` (`ON CONFLICT (portfolio_id, external_ref) DO NOTHING`), single-pass deterministic projection rebuild (`RebuildProjections`), and automated retroactive portfolio valuation backfills.
- **Key Files**:
  - [portfolio.proto](file:///Users/oscargarcia/workspace/graphfolio/proto/portfolio/v1/portfolio.proto)
  - [transaction.go (Domain)](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/domain/transaction.go)
  - [queries.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/repository/queries.go)
  - [postgres.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/repository/postgres.go)
  - [transaction.go (Service)](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/service/transaction.go)
  - [transaction_test.go (Service)](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/service/transaction_test.go)
  - [server.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/server.go)
  - [server_test.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/server_test.go)
  - [schema.graphqls](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.graphqls)
  - [schema.resolvers.go](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.resolvers.go)
  - [schema.resolvers_test.go](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.resolvers_test.go)
  - [csv services](file:///Users/oscargarcia/workspace/graphfolio/web/apps/main-app/src/services/csv/) (`types.ts`, `utils.ts`, `detector.ts`, `parsers/commsec.ts`, `parsers/nabtrade.ts`)
  - [ImportTransactionsModal.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/main-app/src/components/ImportTransactionsModal.tsx)
  - [ImportTransactionsModal.css](file:///Users/oscargarcia/workspace/graphfolio/web/apps/main-app/src/components/ImportTransactionsModal.css)
  - [TransactionLedger.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/main-app/src/components/TransactionLedger.tsx)
  - [Dashboard.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/main-app/src/components/Dashboard.tsx)

### 2.25 Investment Holding Return Attribution (Capital Gain, Income & Currency Gain/Loss)
- **Status**: **DONE**
- **Plan Reference**: [holding-return-attribution-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/holding-return-attribution-implementation-plan.md)
- **User Story**: [investment-holding-capital-gain-income-currency.md](file:///Users/oscargarcia/workspace/graphfolio/docs/user-stories/investment-holding-capital-gain-income-currency.md)
- **Scope**: `proto/portfolio/v1/`, `services/portfolio-api`, `bff/`, `web/apps/main-app`
- **Description**: Robust 3-pillar multi-currency return attribution engine decomposing holding performance into Capital Gain/Loss (pure asset price movement), Income (cumulative dividends and cash interest net of withholding tax), and Currency Gain/Loss (FX fluctuations on foreign denominated assets relative to portfolio base currency). Features exact historical acquisition exchange rate weighting ($\overline{\text{FX}}_0 = \frac{C_{\text{base}}}{C_{\text{local}}}$) cleanly separating equity appreciation from FX currency tailwinds/headwinds, while preserving the mathematical attribution identity $\text{CapGain}_{\text{base}} + \text{FXGain}_{\text{base}} \equiv V_{\text{base}} - C_{\text{base}}$. Domestic equities automatically display a neutral em-dash (`—`) with zero currency risk. The primary investor table in `Dashboard.tsx` expands to 9 responsive columns with an `INTL` asset badge for foreign holdings, Yield on Cost (YOC%), and an itemized subtotal footer row reconciling the sum of Capital Gain, Income, and Currency Gain against Total Portfolio Return.
- **Key Files**:
  - [portfolio.proto](file:///Users/oscargarcia/workspace/graphfolio/proto/portfolio/v1/portfolio.proto)
  - [holding.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/domain/holding.go)
  - [calculator.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/domain/calculator.go)
  - [calculator_test.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/domain/calculator_test.go)
  - [server.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/server.go)
  - [schema.graphqls](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.graphqls)
  - [helpers.go](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/helpers.go)
  - [Dashboard.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/main-app/src/components/Dashboard.tsx)
  - [Dashboard.css](file:///Users/oscargarcia/workspace/graphfolio/web/apps/main-app/src/components/Dashboard.css)

### 2.26 Average Buy Price Holdings Column
- **Status**: **DONE**
- **Plan Reference**: [average-buy-price-column-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/average-buy-price-column-implementation-plan.md)
- **User Story**: [user_story_average_buy_price_column.md](file:///Users/oscargarcia/workspace/graphfolio/docs/user-stories/user_story_average_buy_price_column.md)
- **Scope**: `proto/portfolio/v1/`, `services/portfolio-api`, `bff/`, `web/apps/main-app`
- **Description**: Displays the volume-weighted average purchase price across all open lots directly alongside the market Price column in the investments table. Denominated in the instrument trading currency to allow direct 1-to-1 comparison against current execution prices. Incorporates stock splits seamlessly through deterministic ledger replay, supports interactive column sorting with informational tooltip, renders empty placeholders (`—`) for unpriced holdings, and adjusts table footer spans (`colSpan={4}`).
- **Key Files**:
  - [portfolio.proto](file:///Users/oscargarcia/workspace/graphfolio/proto/portfolio/v1/portfolio.proto)
  - [holding.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/domain/holding.go)
  - [calculator.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/domain/calculator.go)
  - [calculator_test.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/domain/calculator_test.go)
  - [server.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/server.go)
  - [schema.graphqls](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.graphqls)
  - [helpers.go](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/helpers.go)
  - [schema.resolvers_test.go](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.resolvers_test.go)
  - [Dashboard.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/main-app/src/components/Dashboard.tsx)
  - [Dashboard.css](file:///Users/oscargarcia/workspace/graphfolio/web/apps/main-app/src/components/Dashboard.css)

### 2.27 Portfolio Cash Flow Report Generation
- **Status**: **DONE**
- **Plan Reference**: [portfolio-cash-flow-report-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/portfolio-cash-flow-report-implementation-plan.md)
- **Scope**: `proto/portfolio/v1/`, `services/portfolio-api`, `bff/`, `web/apps/main-app`
- **Highlights**:
  - GIPS, US GAAP ASC 946, and IFRS 9 trade settlement accounting for complete portfolio cash flows.
  - Multi-timeframe filters: MTD, YTD, Trailing 12M, 6M, 3M, Inception, and Custom date ranges.
  - High-precision cash reconciliation: $\text{Starting Cash Balance} + \text{Total Inflows} - \text{Total Outflows} \equiv \text{Ending Cash Balance}$.
  - KPI summary cards and category subtotals for capital deposits, dividends, interest, sales proceeds, withdrawals, purchases, fees, and taxes.
  - Itemized transaction ledger with sequential running cash balances and multi-column sorting.
  - Unified **Reports** tab hub (`ReportsView.tsx`) in the investor dashboard housing the Cash Flow Report (`CashFlowReport.tsx`) alongside upcoming reporting modules.
  - One-click formatted CSV report export for tax accounting and reconciliation audits.
- **Key Files**:
  - [portfolio.proto](file:///Users/oscargarcia/workspace/graphfolio/proto/portfolio/v1/portfolio.proto)
  - [cash_flow.go (domain)](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/domain/cash_flow.go)
  - [cash_flow.go (service)](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/service/cash_flow.go)
  - [cash_flow_test.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/service/cash_flow_test.go)
  - [server.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/server.go)
  - [schema.graphqls](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.graphqls)
  - [helpers.go](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/helpers.go)
  - [schema.resolvers.go](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.resolvers.go)
  - [schema.resolvers_test.go](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.resolvers_test.go)
  - [Dashboard.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/main-app/src/components/Dashboard.tsx)
  - [ReportsView.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/main-app/src/components/ReportsView.tsx)
  - [CashFlowReport.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/main-app/src/components/CashFlowReport.tsx)
  - [CashFlowReport.css](file:///Users/oscargarcia/workspace/graphfolio/web/apps/main-app/src/components/CashFlowReport.css)

### 2.28 GIPS CAGR Annualized Return & Valuation Cost-Basis Resiliency
- **Status**: **DONE**
- **Plan Reference**: [annualized-return-twr-fix-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/annualized-return-twr-fix-implementation-plan.md) & [annualized-twr-return-investigation-report.md](file:///Users/oscargarcia/workspace/graphfolio/docs/annualized-twr-return-investigation-report.md)
- **Description**: Robust valuation engine defense and true GIPS-compliant performance measurement. Features US GAAP ASC 820 Level 2/3 cost basis fallback (`h.CostBasisBase` / `h.CostBasis / h.Quantity`) in `valuation.go` (`SnapshotValuation` & `RebuildValuations`) whenever closing market prices are missing from the database, eliminating artificial asset write-downs and false $-573\%$ daily return drops on inflow days. Integrates inception tracking via `GetEarliestValuation` and GIPS Standard 2.A.20 CAGR annualization in `calculator.go` $((\text{TWRIndex})^{\frac{365.25}{D}} - 1) \times 100$ for periods $\ge 365$ days, preserving cumulative period returns for sub-annual holding periods.
- **Key Files**:
  - [valuation.go (service)](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/service/valuation.go)
  - [valuation_test.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/service/valuation_test.go)
  - [calculator.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/domain/calculator.go)
  - [calculator_test.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/domain/calculator_test.go)
  - [service.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/service/service.go)
  - [repository.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/repository/repository.go)
### 2.29 Missing Prices Report & Date Range Backfill (Admin Portal)
- **Status**: **DONE**
- **Plan Reference**: [admin-missing-prices-report-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/admin-missing-prices-report-implementation-plan.md)
- **User Story**: [admin-missing-prices-report.md](file:///Users/oscargarcia/workspace/graphfolio/docs/user-stories/admin-missing-prices-report.md)
- **Description**: Comprehensive diagnostic report in the internal Admin Portal detecting historical pricing anomalies where transactions (`BUY`, `SELL`, `DIVIDEND`, `SPLIT`) lack matching EOD closing marks in `portfolio.instrument_prices` on their execution `trade_date`. Features top-level KPI cards (Unpriced Trades, Affected Assets, Earliest Missing Date, Date Span Coverage), dual-input date range filtering with fast presets (`30D`, `90D`, `YTD`, `1Y`, `ALL`), symbol and transaction type filters, and direct one-click remediation actions. Operators can launch `BackfillModal` pre-populated with affected symbols and date bounds to trigger targeted provider syncs and valuation recomputations, or use `AddAssetPriceModal` for manual price marks.
- **Key Files**:
  - [portfolio.proto](file:///Users/oscargarcia/workspace/graphfolio/proto/portfolio/v1/portfolio.proto)
  - [missing_price.go (domain)](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/domain/missing_price.go)
  - [queries.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/repository/queries.go)
  - [postgres.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/repository/postgres.go)
  - [admin.go (service)](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/service/admin.go)
  - [server.go (gRPC)](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/server.go)
  - [schema.graphqls](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.graphqls)
  - [schema.resolvers.go](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.resolvers.go)
  - [MissingPricesReport.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/admin-app/src/components/MissingPricesReport.tsx)
  - [MissingPricesReport.css](file:///Users/oscargarcia/workspace/graphfolio/web/apps/admin-app/src/components/MissingPricesReport.css)
  - [AdminLayout.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/admin-app/src/components/AdminLayout.tsx)
  - [App.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/admin-app/src/App.tsx)

### 2.30 Include Closed Positions in Dashboard Investments Table
- **Status**: **DONE**
- **User Story**: [include-closed-positions-dashboard.md](file:///Users/oscargarcia/workspace/graphfolio/docs/user-stories/include-closed-positions-dashboard.md)
- **Plan Reference**: [include-closed-positions-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/include-closed-positions-implementation-plan.md)
- **Scope**: `proto/portfolio/v1/`, `services/portfolio-api`, `bff/`, `web/apps/main-app`
- **Description**: Allows investors to view fully exited/liquidated positions ($Q = 0$) directly within the "Your Investments" table via an interactive header checkbox toggle. Reconstructs lifetime performance (realized capital gain, accumulated dividends/income, currency impact, and net total return) from historical tax lots without zero-division or valuation distortion, preserving $0.00 current market value and displaying a distinct `CLOSED` badge.
- **Key Files**:
  - [portfolio.proto](file:///Users/oscargarcia/workspace/graphfolio/proto/portfolio/v1/portfolio.proto)
  - [holding.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/domain/holding.go)
  - [calculator.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/domain/calculator.go)
  - [queries.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/repository/queries.go)
  - [postgres.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/repository/postgres.go)
  - [service.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/service/service.go)
  - [server.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/server.go)
  - [schema.graphqls](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.graphqls)
  - [schema.resolvers.go](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.resolvers.go)
  - [Dashboard.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/main-app/src/components/Dashboard.tsx)
  - [Dashboard.css](file:///Users/oscargarcia/workspace/graphfolio/web/apps/main-app/src/components/Dashboard.css)

### 2.31 Modified Dietz Annualized Return & Cash-Flow Weighting
- **Status**: **DONE**
- **User Story**: [annualized-return-modified-dietz.md](file:///Users/oscargarcia/workspace/graphfolio/docs/user-stories/annualized-return-modified-dietz.md)
- **Plan Reference**: [annualized-return-modified-dietz-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/annualized-return-modified-dietz-implementation-plan.md)
- **Scope**: `proto/portfolio/v1/`, `services/portfolio-api`, `bff/`, `web/apps/main-app`
- **Description**: Accurate, industry-standard money-weighted rate of return calculation using the Modified Dietz Method adhering strictly to GIPS Standard 2.A.20. Weights external cash flows ($DEPOSIT$, $WITHDRAWAL$) on an Actual/Actual day count basis against portfolio start value. Features compound geometric annualization $((1 + R_{\text{Dietz}})^{\frac{365.25}{CD}} - 1) \times 100\%$ for multi-year periods ($\ge 365$ days) with sub-annual period annualization guard, optional multi-year monthly sub-period linking to prevent cash flow dilution, and an interactive glassmorphic modal with complete capital decomposition and itemized cash-flow weight inspection table.
- **Key Files**:
  - [portfolio.proto](file:///Users/oscargarcia/workspace/graphfolio/proto/portfolio/v1/portfolio.proto)
  - [dietz.go (domain)](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/domain/dietz.go)
  - [dietz_test.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/domain/dietz_test.go)
  - [queries.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/repository/queries.go)
  - [postgres.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/repository/postgres.go)
  - [dietz.go (service)](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/service/dietz.go)
  - [dietz_test.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/service/dietz_test.go)
  - [server.go (gRPC)](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/server.go)
  - [server_test.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/server_test.go)
  - [schema.graphqls](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.graphqls)
  - [helpers.go](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/helpers.go)
  - [schema.resolvers.go](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.resolvers.go)
  - [schema.resolvers_test.go](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.resolvers_test.go)
  - [ModifiedDietzModal.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/main-app/src/components/ModifiedDietzModal.tsx)
  - [ModifiedDietzModal.css](file:///Users/oscargarcia/workspace/graphfolio/web/apps/main-app/src/components/ModifiedDietzModal.css)
  - [Dashboard.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/main-app/src/components/Dashboard.tsx)
  - [Dashboard.css](file:///Users/oscargarcia/workspace/graphfolio/web/apps/main-app/src/components/Dashboard.css)

### 2.32 Missing FX Rates Report & Date Range Backfill (Admin Portal)
- **Status**: **DONE**
- **User Story**: [admin-missing-fx-rates-report.md](file:///Users/oscargarcia/workspace/graphfolio/docs/user-stories/admin-missing-fx-rates-report.md)
- **Plan Reference**: [admin-missing-fx-rates-report-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/admin-missing-fx-rates-report-implementation-plan.md)
- **Scope**: `proto/portfolio/v1/`, `services/portfolio-api`, `bff/`, `web/apps/admin-app`
- **Description**: Diagnostic and audit report identifying all recorded cross-currency transactions ($t.\text{currency\_code} \neq p.\text{base\_currency}$) lacking an authoritative daily closing foreign exchange rate in `portfolio.fx_rates` on their trade date. Supports bidirectional direct and inverted quote lookups ($1/\text{rate}$), high-precision decimal exchange rate storage, and interactive single-click integration to pre-populate [BackfillModal.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/admin-app/src/components/BackfillModal.tsx) (with a 3-day buffer for weekend/holiday trade dates) or [FXOverrideModal.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/admin-app/src/components/FXOverrideModal.tsx) to record manual rates with automatic valuation recomputation.
- **Key Files**:
  - [portfolio.proto](file:///Users/oscargarcia/workspace/graphfolio/proto/portfolio/v1/portfolio.proto)
  - [missing_fx.go (domain)](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/domain/missing_fx.go)
  - [queries.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/repository/queries.go)
  - [postgres.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/repository/postgres.go)
  - [admin.go (service)](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/service/admin.go)
  - [admin_test.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/service/admin_test.go)
  - [server.go (gRPC)](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/server.go)
  - [server_test.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/server_test.go)
  - [schema.graphqls](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.graphqls)
  - [helpers.go](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/helpers.go)
  - [schema.resolvers.go](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.resolvers.go)
  - [schema.resolvers_test.go](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.resolvers_test.go)
  - [MissingFXRatesReport.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/admin-app/src/components/MissingFXRatesReport.tsx)
  - [MissingFXRatesReport.css](file:///Users/oscargarcia/workspace/graphfolio/web/apps/admin-app/src/components/MissingFXRatesReport.css)
  - [MissingFXRatesReport.test.ts](file:///Users/oscargarcia/workspace/graphfolio/web/apps/admin-app/src/components/MissingFXRatesReport.test.ts)
  - [AdminLayout.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/admin-app/src/components/AdminLayout.tsx)

### 2.33 Portfolio Benchmark Comparison & Relative Performance Analytics
- **Status**: **DONE**
- **User Story**: [benchmark-performance-comparison.md](file:///Users/oscargarcia/workspace/graphfolio/docs/user-stories/benchmark-performance-comparison.md)
- **Plan Reference**: [benchmark-performance-comparison-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/benchmark-performance-comparison-implementation-plan.md)
- **Scope**: Database (`portfolio.benchmarks`), `proto/portfolio/v1/`, `services/portfolio-api`, `bff/`, `web/apps/main-app`
- **Description**: Institutional-grade comparative performance analysis engine allowing investors to overlay market benchmarks (S&P 500 Total Return `SP500TR`, NASDAQ-100 `QQQ`, ASX 200 `VAS`) directly against their portfolio performance. Features GIPS Standard 2.A compliant origin re-basing ($0.00\%$ at $T_0$), sub-1-year annualization prohibition ($CD < 365$), dual currency modes (Portfolio Base Currency with daily spot FX translation vs Native Benchmark Currency unhedged), active risk attribution (Active Alpha, Beta $\beta$, Tracking Error $\text{TE}$, Information Ratio $\text{IR}$), and interactive dual-curve SVG visualization with synchronized hover tooltips.
- **Key Files**:
  - [000010_benchmarks.up.sql](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/migrations/000010_benchmarks.up.sql)
  - [portfolio.proto](file:///Users/oscargarcia/workspace/graphfolio/proto/portfolio/v1/portfolio.proto)
  - [benchmark.go (domain)](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/domain/benchmark.go)
  - [benchmark_test.go (domain)](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/domain/benchmark_test.go)
  - [repository.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/repository/repository.go)
  - [queries.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/repository/queries.go)
  - [postgres.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/repository/postgres.go)
  - [benchmark.go (service)](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/service/benchmark.go)
  - [benchmark_test.go (service)](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/service/benchmark_test.go)
  - [server.go (gRPC)](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/server.go)
  - [server_test.go (gRPC)](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/server_test.go)
  - [schema.graphqls](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.graphqls)
  - [helpers.go](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/helpers.go)
  - [schema.resolvers.go](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.resolvers.go)
  - [schema.resolvers_test.go](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.resolvers_test.go)
  - [PerformanceChart.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/main-app/src/components/PerformanceChart.tsx)
  - [PerformanceChart.css](file:///Users/oscargarcia/workspace/graphfolio/web/apps/main-app/src/components/PerformanceChart.css)
  - [BenchmarkMetricsStrip.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/main-app/src/components/BenchmarkMetricsStrip.tsx)
  - [BenchmarkMetricsStrip.css](file:///Users/oscargarcia/workspace/graphfolio/web/apps/main-app/src/components/BenchmarkMetricsStrip.css)

### 2.34 User Registration, Authentication & Session Management (Main App)
- **Status**: **DONE**
- **User Story**: [user-registration-and-login.md](file:///Users/oscargarcia/workspace/graphfolio/docs/user-stories/user-registration-and-login.md) (`US-AUTH-101`)
- **Plan Reference**: [user-registration-and-login-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/user-registration-and-login-implementation-plan.md)
- **Scope**: Proto (`proto/user/v1/user.proto`, `proto/portfolio/v1/portfolio.proto`), Database (`users.users`, `users.sessions`), User Service (`services/user-api`), BFF (`bff/`), Main App (`web/apps/main-app`)
- **Highlights**:
  - Secure investor registration with Argon2id cryptographic password hashing, real-time complexity validation, and automatic initial default portfolio provisioning in the chosen base currency.
  - Dual-token session security architecture: short-lived (15-minute) JWT access tokens in memory and long-lived (30-day) `httpOnly`, `Secure`, `SameSite=Strict` rotating refresh cookies with token family reuse detection.
  - Decoupling BFF resolvers from the hardcoded `defaultUserID` using contextual authorization middleware (`ctx.UserID`).
  - Dark glassmorphic login and registration views in `@graphfolio/ui` aesthetics, accessible forms, silent background session refresh, and explicit sign-out action in the navigation profile menu.

### 2.35 User Password Change & Session Invalidation (Main App)
- **Status**: **DONE**
- **User Story**: [user-change-password.md](file:///Users/oscargarcia/workspace/graphfolio/docs/user-stories/user-change-password.md) (`US-AUTH-102`)
- **Plan Reference**: [user-change-password-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/user-change-password-implementation-plan.md)
- **Scope**: Proto (`proto/user/v1/user.proto`), Database (`users.users`, `users.sessions`), User Service (`services/user-api`), BFF (`bff/`), Main App (`web/apps/main-app`)
- **Highlights**:
  - Secure credential rotation conforming to OWASP ASVS v4.0 Level 2 and NIST SP 800-63B memorized secret verifier specifications.
  - Mandatory current password re-authentication before permitting changes, protecting against unauthorized workstation takeover.
  - Real-time password complexity enforcement ($\ge 10$ characters, uppercase, lowercase, digit, special symbol) via live `PasswordStrengthMeter` checklist and match verification.
  - Prevention of identical password reuse and constant-time dummy hash verification against user enumeration timing attacks.
  - Option to revoke all other active sessions and background refresh tokens across devices while preserving uninterrupted authentication on the current active session.
  - Dark glassmorphic `ChangePasswordModal` launched from `UserProfileMenu` with password visibility toggles, loading state, error banners, and auto-dismissing success notifications.

### 2.36 Download Transactions from Transaction Ledger (Ascending Order)
- **Status**: **DONE**
- **User Story**: [download-transactions-ledger.md](file:///Users/oscargarcia/workspace/graphfolio/docs/user-stories/download-transactions-ledger.md) (`US-TX-103`)
- **Plan Reference**: [download-transactions-ledger-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/download-transactions-ledger-implementation-plan.md)
- **Scope**: Proto (`proto/portfolio/v1/portfolio.proto`), Database (`portfolio.transactions`), Portfolio Service (`services/portfolio-api`), BFF (`bff/`), Main App (`web/apps/main-app`)
- **Highlights**:
  - Direct CSV download action ("Export CSV") integrated into the `TransactionLedger.tsx` header toolbar alongside "Import CSV".
  - Strict ascending chronological ordering (`trade_date ASC, created_at ASC, id ASC`) so financial events flow forward from inception, ensuring seamless ingestion into tax tools and external accounting engines without negative inventory errors.
  - Complete dataset export across all pages, bypassing UI pagination limits (20 items/page).
  - Filter-aware export respecting active transaction type filters (e.g. Dividend only) and symbol searches while maintaining chronological integrity.
  - Comprehensive CSV schema including trade date, transaction type, ticker, company name, quantity, price, gross/net amounts, fee, fee currency, withholding tax, FX rate to base, external broker reference, and user notes.
  - Robust security and formatting guards: RFC 4180 escaping, UTF-8 BOM encoding for Excel compatibility, and formula injection sanitization (CWE-1236) protecting against leading `=, +, -, @` characters.

### 2.37 Admin Portal Authentication, Auto-Provisioning & Mandatory First-Login Password Rotation
- **Status**: **DONE**
- **User Story**: [admin-portal-login-and-provisioning.md](file:///Users/oscargarcia/workspace/graphfolio/docs/user-stories/admin-portal-login-and-provisioning.md) (`US-ADM-103`)
- **Plan Reference**: [admin-portal-login-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/admin-portal-login-implementation-plan.md)
- **Scope**: Database (`users.users`), Proto (`proto/user/v1/user.proto`), User Microservice (`services/user-api`), BFF (`bff/`), Admin Portal (`web/apps/admin-app`)
- **Highlights**:
  - Automated idempotent database migration (`000004_add_role_and_password_policy.up.sql`) and seed (`dev_seed.sql`) auto-provisioning `admin@graphfolio.app` with default password `Password123!` (Argon2id RFC 9106 hash: `$argon2id$v=19$m=65536,t=3,p=4$z48VmX+qFO9X2FkAfV/Z6Q$WW1NOG+O6TwEwlSDkDIfae3Xi0xOePDpNg+rd7lkWsw`), `role = 'ADMIN'`, and `must_change_password = true`.
  - Strict Role-Based Access Control (RBAC): `role text NOT NULL DEFAULT 'USER' CHECK (role IN ('USER', 'ADMIN'))` in PostgreSQL, `UserRole` enum in Protobuf and GraphQL, and embedded `Role` in JWT access token claims. Standard investors (`role = 'USER'`) attempting access to the admin portal (:5174) are denied with an explicit 403 `AccessDeniedView` offering an immediate sign-out action.
  - Mandatory first-login password rotation: `must_change_password` flag intercepts admin login and strictly enforces completion of `AdminChangePasswordView`. Operational console tabs (Assets, Prices, Missing Prices, Missing FX, FX Management, Ingestion Pipeline) remain unmounted and blocked until rotated.
  - NIST SP 800-63B password complexity enforcement ($\ge 10$ characters, uppercase, lowercase, digit, symbol) with real-time `PasswordStrengthMeter` checklist, match confirmation, and optional multi-session revocation. Calling `UpdatePassword` in `services/user-api` automatically resets `must_change_password = false`, instantly unlocking the portal.
  - Admin portal session lifecycle via `AdminAuthContext` and `AdminAuthProvider`: silent `httpOnly` cookie refresh on mount, background token renewal, in-memory JWT bearer storage, dynamic `@graphfolio/api-client` GenQL client injection with `credentials: 'include'`.
  - Top-right administrative profile header in `AdminLayout` displaying user email, `ADMIN` badge, "Change Password" modal launcher, and "Sign Out" dropdown action.
  - Client-side security state machine in `web/apps/admin-app/src/App.tsx`:
    $$\text{Loading} \longrightarrow \text{Unauthenticated (AdminLoginView)} \longrightarrow \text{Non-Admin (AccessDeniedView)} \longrightarrow \text{MustChangePassword (AdminChangePasswordView)} \longrightarrow \text{AdminLayout}$$
  - Full automated test coverage in `services/user-api` (`server_test.go`), BFF (`resolvers_test.go`), and Admin App (`AdminAuth.test.ts`).

### 2.38 Bidirectional Date Sorting in Transaction Ledger (`DESC` ↔ `ASC`)
- **Status**: **DONE**
- **User Story**: [sort-transactions-by-date-ledger.md](file:///Users/oscargarcia/workspace/graphfolio/docs/user-stories/sort-transactions-by-date-ledger.md) (`US-TX-104`)
- **Plan Reference**: [sort-transactions-by-date-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/sort-transactions-by-date-implementation-plan.md)
- **Scope**: Proto (`SortDirection`), Database (`portfolio.transactions`), Portfolio Service (`services/portfolio-api`), BFF (`bff/`), Main App (`web/apps/main-app`)
- **Highlights**:
  - Interactive bidirectional sort toggle on the "Date" column header in `TransactionLedger.tsx` supporting reverse chronological descending (`DESC`, newest first) and chronological ascending (`ASC`, oldest first).
  - Dynamic sort indicator chevrons (`▲/▼`), accessible ARIA markup (`role="columnheader"`, `aria-sort="ascending" | "descending"`), and full keyboard interaction (`Enter` / `Space`).
  - Seamless state preservation: toggling sort direction preserves active transaction type filters (`selectedType`) and symbol queries (`symbolQuery`), automatically resetting to page 1.
  - End-to-end backend contract support: `SortDirection` enum defined in `proto/portfolio/v1/portfolio.proto` and exposed via GraphQL `schema.graphqls`.
  - Deterministic PostgreSQL repository queries (`listTransactionsAscSQL` and `listTransactionsDescSQL`) incorporating event lifecycle tie-breaking (`DEPOSIT` before `BUY` before `SELL` for ASC; reverse for DESC).

### 2.39 Create Cash Withdrawal Against Dividend Transaction
- **Status**: **DONE**
- **User Story**: [withdraw-against-dividend-transaction.md](file:///Users/oscargarcia/workspace/graphfolio/docs/user-stories/withdraw-against-dividend-transaction.md) (`US-TX-105`)
- **Plan Reference**: [withdraw-against-dividend-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/withdraw-against-dividend-implementation-plan.md)
- **Scope**: Web (`web/apps/main-app`), BFF (`bff/`), Portfolio Service (`services/portfolio-api`)
- **Highlights**:
  - Contextual row-level action button rendered exclusively on `DIVIDEND` items in `TransactionLedger.tsx`.
  - Dark glassmorphic modal (`WithdrawDividendModal.tsx`) pre-populating trade date, currency, and exact net dividend amount.
  - Standardized automated audit reference memo: `"Withdrawal for dividend: SYMBOL on YYYY-MM-DD"`.
  - Eliminates "ghost cash" accumulation and preserves 100% bank statement cash reconciliation ($\Delta \text{Cash} = +D - D = \$0.00$) for investors who sweep dividend income into personal checking/savings accounts.
  - Guaranteed intra-day ordering safety during ledger replay (`DIVIDEND` with priority 5 executes before `WITHDRAWAL` with priority 11).

### 2.40 Filter Transactions by Interest Type in Transaction Ledger
- **Status**: **DONE**
- **User Story**: [filter-transactions-by-interest-ledger.md](file:///Users/oscargarcia/workspace/graphfolio/docs/user-stories/filter-transactions-by-interest-ledger.md) (`US-TX-106`)
- **Plan Reference**: [filter-transactions-by-interest-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/filter-transactions-by-interest-implementation-plan.md)
- **Scope**: Web (`web/apps/main-app`), BFF (`bff/`), Portfolio Service (`services/portfolio-api`), Database (`portfolio.transactions`)
- **Highlights**:
  - Dedicated "Interest" filter pill added to `filterTypes` toolbar in `TransactionLedger.tsx`, positioned alongside income cash flows.
  - Distinctive cyan/teal glassmorphic badge styling (`.badge-type.interest` with `#22d3ee`) in `TransactionLedger.css`.
  - Contextual zero-state empty message (`getEmptyStateMessage`) informing the investor when no interest records exist for the active portfolio or search criteria.
  - Filtered CSV data export generating interest-only audit files (`graphfolio-transactions-interest-YYYY-MM-DD.csv`).
  - Comprehensive unit test coverage verifying filter pill constants, query argument building, badge class resolution, and empty state messaging.

### 2.41 Record Cash Interest Transactions (Yield Ingestion & Cash Account Reconciliation)
- **Status**: **DONE**
- **User Story**: [record-interest-transaction.md](file:///Users/oscargarcia/workspace/graphfolio/docs/user-stories/record-interest-transaction.md) (`US-TX-107`)
- **Plan Reference**: [record-interest-transaction-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/record-interest-transaction-implementation-plan.md)
- **Scope**: Proto (`proto/portfolio/v1/portfolio.proto`), Database (`portfolio.transactions`, `portfolio.cash_balances`), Portfolio Service (`services/portfolio-api`), BFF (`bff/`), Main App (`web/apps/main-app`)
- **Highlights**:
  - Direct manual recording of cash yield/interest payments via dedicated "Interest" type tab in `AddTransactionModal.tsx`.
  - Dynamic form adaptation hiding equity-specific fields (Symbol, Quantity, Price, Split) and presenting Trade Date, Amount, Currency, and Notes.
  - Strict preservation of GIPS® TWR and Modified Dietz investment return metrics by classifying cash yield as internal income rather than external capital contributions ($F_t = 0$).
  - Exact fixed-point arithmetic (`shopspring/decimal.Decimal`) updating portfolio cash balances and reflecting in the Transaction Ledger and Cash Flow Report without full page reloads.

### 2.42 User Preferences — End of Financial Year (EOFY) Configuration
- **Status**: **DONE**
- **User Story**: [user-preferences-financial-year-end.md](file:///Users/oscargarcia/workspace/graphfolio/docs/user-stories/user-preferences-financial-year-end.md) (`US-PREF-102`)
- **Plan Reference**: [user-preferences-financial-year-end-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/user-preferences-financial-year-end-implementation-plan.md)
- **Scope**: Database (`users.users`), Proto (`proto/user/v1/user.proto`), User Service (`services/user-api`), BFF (`bff/`), Main App (`web/apps/main-app`, `web/packages/api-client`)
- **Highlights**:
  - Schema extension with `financial_year_end_month` and `financial_year_end_day` on `users.users` table with standard-based validation constraints (1–12 months, 1–31 days, leap-year-safe February clamping).
  - Protobuf RPC enhancement in `proto/user/v1/user.proto` (`UserPreferences` and `UpdateUserPreferencesRequest`) and gRPC microservice implementation.
  - GraphQL API exposure in BFF (`schema.graphqls`, `schema.resolvers.go`) mapping user fiscal calendar parameters.
  - Interactive preferences modal (`UserPreferencesModal.tsx`) with one-click major tax jurisdiction presets (Australia 30 Jun, UK 5 Apr, US/Canada 31 Dec, NZ/India 31 Mar, South Africa 28 Feb), custom month/day dropdowns with dynamic day clamping, and live fiscal calendar preview card displaying current & previous fiscal year date ranges.
  - Complete multi-layer test suite: domain validator unit tests in Go (`user_test.go`), gRPC service and server unit tests, BFF GraphQL resolver tests (`schema.resolvers_test.go`), and frontend Vitest suite (`UserPreferencesModal.test.ts`).

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
  - Single-pass `COUNT(*) OVER()` pagination with server-side sorting by start date.
  - Interactive drawer/modal to inspect error diagnostics, stack traces, and execution parameters for failed imports.

### 3.4 Benchmark Instrument Market Price Ingestion & Historical Backfill
- **Status**: **PLANNED**
- **User Story**: [benchmark-instrument-price-ingestion.md](file:///Users/oscargarcia/workspace/graphfolio/docs/user-stories/benchmark-instrument-price-ingestion.md) (`US-MKT-016`)
- **Plan Reference**: [benchmark-instrument-price-ingestion-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/benchmark-instrument-price-ingestion-plan.md)
- **Scope**: Market Data (`services/portfolio-api/internal/marketdata`), Database (`portfolio.benchmarks`, `portfolio.instrument_prices`), CLI Worker (`cmd/market-ingest/`), Admin Portal (`web/apps/admin-app`)
- **Highlights**:
  - Dedicated benchmark discovery query integrating all `portfolio.benchmarks` rows into the daily EOD sync workflow, regardless of whether any user holds them.
  - Multi-provider ticker and symbol resolution (e.g. mapping internal `SP500TR` to Yahoo `^SP500TR` or ETF proxy `SPY`, and `VAS` with `XASX` to `VAS.AX`).
  - Total Return adjusted close (`adjclose`) ingestion for ETF proxies to ensure multi-year reinvested dividend comparability and eliminate dividend bias.
  - Multi-year historical backfill capability via CLI (`cmd/market-ingest -symbol SP500TR -from 2020-01-01`) and Admin backfill triggers.
  - Integration with Admin Portal `MissingPricesReport` to identify and backfill missing benchmark quotes with a single click.

### 3.5 In-Place Transaction Editing in Transaction Ledger
- **Status**: **PLANNED**
- **User Story**: [edit-transaction-ledger.md](file:///Users/oscargarcia/workspace/graphfolio/docs/user-stories/edit-transaction-ledger.md) (`US-TX-102`)
- **Plan Reference**: [edit-transaction-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/edit-transaction-implementation-plan.md)
- **Scope**: Proto (`proto/portfolio/v1/portfolio.proto`), Database (`portfolio.transactions`, `portfolio.tax_lots`, `portfolio.projections_holdings`), Portfolio Service (`services/portfolio-api`), BFF (`bff/`), Main App (`web/apps/main-app`)
- **Highlights**:
  - In-place transaction editing directly from the `TransactionLedger.tsx` table via a dedicated pencil icon action button.
  - Dark glassmorphic `EditTransactionModal` pre-populated with trade parameters (date, ticker, quantity, execution price, net amount, brokerage fees, and notes).
  - Short-sale prevention and Negative Inventory Guard verifying that editing historical trades never causes interim inventory to dip below zero.
  - Atomic ledger projection replay (`RebuildProjections`) deterministically re-evaluating FIFO tax-lot relief, open position cost basis, realized gains, and cash balance.
  - Automatic multi-day daily portfolio valuation backfill (`BackfillPortfolioValuations`) re-anchoring daily valuation snapshots from $\min(T_{\text{old}}, T_{\text{new}})$ through today.

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

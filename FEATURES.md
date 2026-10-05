# GraphFolio Features Matrix & Roadmap

This document catalogs all implemented features, in-progress components, and planned future enhancements across GraphFolio. Each entry references its architectural implementation plan, current status, and target layer breakdown.

---

## 1. Feature Status Summary

| # | Feature | Status | Primary Layers | Implementation Plan |
|---|---|---|---|---|
| **1** | **PostgreSQL Multi-Schema Database Architecture** | **DONE** | Database, Migrations, Seed | [db-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/db-implementation-plan.md) |
| **2** | **Portfolio Service & Ledger Projection Engine** | **DONE** | Microservice, Projections | [portfolio-service-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/portfolio-service-implementation-plan.md) |
| **3** | **Exact Fixed-Point Decimal Arithmetic** | **DONE** | Proto, Domain, BFF, Web | [portfolio-service-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/portfolio-service-implementation-plan.md) |
| **4** | **Backend-for-Frontend (BFF) GraphQL Layer** | **DONE** | BFF (`gqlgen`), gRPC Client | [portfolio-service-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/portfolio-service-implementation-plan.md) |
| **5** | **Investor Dashboard Web Application** | **DONE** | Web (React, Vite, GenQL) | [portfolio-service-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/portfolio-service-implementation-plan.md) |
| **6** | **Interactive Transaction Ingestion Modal** | **DONE** | Proto, Svc, BFF, Web | [add-transactions-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/add-transactions-implementation-plan.md) |
| **7** | **Interactive SVG Performance Chart & Time Range Filtering** | **DONE** | Proto, Svc, BFF, Web | [performance-chart-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/performance-chart-implementation-plan.md) |
| **8** | **Transaction History & Ledger Management** | **DONE** | Proto, Svc, BFF, Web | [transaction-history-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/transaction-history-implementation-plan.md) |
| **9** | **Frontend Workspace Architecture (Main App & Admin Portal)** | **DONE** | Web (Monorepo, React, Vite), BFF | [web-workspace-refactoring-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/web-workspace-refactoring-plan.md) |
| **10** | **Admin Portal Backend & BFF (Asset, Price & Ingestion Management)** | **DONE** | Proto, Svc, BFF, Web | [admin-portal-backend-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/admin-portal-backend-implementation-plan.md) |
| **11** | **Dynamic Cost Basis Method Switching (Average Cost ↔ FIFO)** | **PLANNED** | Proto, Svc, BFF, Web | [cost-basis-switching-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/cost-basis-switching-implementation-plan.md) |
| **12** | **User Preferences & Display Currency** | **PLANNED** | Proto, User API, BFF, Web | [user-preferences-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/user-preferences-implementation-plan.md) |
| **13** | **Portfolio Valuation Engine & Historical Backfill** | **PLANNED** | Database, Svc, Worker | [portfolio-valuation-engine-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/portfolio-valuation-engine-implementation-plan.md) |
| **14** | **Market Data Ingestion (Asset Prices & FX Rates)** | **PLANNED** | Database, Svc, Worker | [market-data-ingestion-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/market-data-ingestion-implementation-plan.md) |
| **15** | **Look-Through Fundamental & Cash Flow Quality Engine** | **PLANNED** | Database, Proto, Svc, BFF, Web | [fundamental-cash-flow-engine-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/fundamental-cash-flow-engine-implementation-plan.md) |
| **16** | **Tax Lot Inspector & Capital Gains Reports** | **ROADMAP** | Svc, BFF, Web | *(Future Plan)* |
| **17** | **Dividend Calendar & Yield Analytics** | **ROADMAP** | Svc, BFF, Web | *(Future Plan)* |
| **18** | **Real-Time Market Data & WebSocket Price Ticker** | **ROADMAP** | Market Data, Svc, Web | *(Future Plan)* |

---

## 2. Completed Features (`DONE`)

### 2.1 PostgreSQL Multi-Schema Database Architecture
- **Status**: **DONE**
- **Plan Reference**: [db-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/db-implementation-plan.md)
- **Description**: Robust relational foundation on PostgreSQL 17+ featuring schema isolation (`portfolio` and `users` schemas) with dedicated service roles (`portfolio_svc`, `user_svc`). Includes 6 versioned schema migrations (`golang-migrate`), UUIDv7 native primary keys, foreign currency constraints, transaction types enum, tax lot tables, and projection storage.
- **Key Files**:
  - [bootstrap.sql](file:///Users/oscargarcia/workspace/graphfolio/scripts/db/bootstrap.sql)
  - [migrations/000001_reference_data.up.sql](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/migrations/000001_reference_data.up.sql)
  - [migrations/000006_projections.up.sql](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/migrations/000006_projections.up.sql)
  - [dev_seed.sql](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/seeds/dev_seed.sql)

### 2.2 Portfolio Service & Deterministic Ledger Projections
- **Status**: **DONE**
- **Plan Reference**: [portfolio-service-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/portfolio-service-implementation-plan.md)
- **Description**: Standalone Go gRPC microservice (`services/portfolio-api` on `:50051`) implementing an event-sourced ledger projection engine. Holding quantities, cost bases, cash balances, and valuations are derived deterministically by replaying historical transactions through `service.ProcessLedger` with atomic row-level locking (`SELECT ... FOR UPDATE`).
- **Key Files**:
  - [server.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/server.go)
  - [projection.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/service/projection.go)
  - [calculator.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/domain/calculator.go)
  - [postgres.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/repository/postgres.go)

### 2.3 Exact Fixed-Point Decimal Arithmetic
- **Status**: **DONE**
- **Plan Reference**: [portfolio-service-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/portfolio-service-implementation-plan.md)
- **Description**: Strict zero floating-point arithmetic policy. Every financial amount, price, quantity, and percentage is represented with exact arbitrary precision using `shopspring/decimal.Decimal` in Go, custom `common.v1.Decimal` / `common.v1.Money` protobuf types, and a custom string-serialized `Decimal` scalar in GraphQL.
- **Key Files**:
  - [decimal.proto](file:///Users/oscargarcia/workspace/graphfolio/proto/common/v1/decimal.proto)
  - [decimalpb.go](file:///Users/oscargarcia/workspace/graphfolio/pkg/decimalpb/decimalpb.go)
  - [decimal.go](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/model/decimal.go)

### 2.4 Backend-for-Frontend (BFF) GraphQL Layer
- **Status**: **DONE**
- **Plan Reference**: [portfolio-service-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/portfolio-service-implementation-plan.md)
- **Description**: High-performance GraphQL orchestrator built with Go and `gqlgen` running on `:8080`. Connects to gRPC microservices, stitches domain models together, exposes strongly-typed queries and mutations, and handles CORS for frontend clients.
- **Key Files**:
  - [schema.graphqls](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.graphqls)
  - [schema.resolvers.go](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.resolvers.go)
  - [helpers.go](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/helpers.go)
  - [main.go](file:///Users/oscargarcia/workspace/graphfolio/bff/cmd/server/main.go)

### 2.5 Investor Dashboard Web Application
- **Status**: **DONE**
- **Plan Reference**: [portfolio-service-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/portfolio-service-implementation-plan.md) & [genql-usage.md](file:///Users/oscargarcia/workspace/graphfolio/docs/genql-usage.md)
- **Description**: Glassmorphic, dark-mode single-page application built with React 19, TypeScript, and Vite. Queries the BFF using an auto-generated type-safe client (`genql`), rendering total portfolio valuation, day's gain/loss, annualized returns, cash balance, and a comprehensive holdings table with currency formatting.
- **Key Files**:
  - [Dashboard.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/main-app/src/components/Dashboard.tsx)
  - [Dashboard.css](file:///Users/oscargarcia/workspace/graphfolio/web/apps/main-app/src/components/Dashboard.css)
  - [main.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/main-app/src/main.tsx)

### 2.6 Interactive Transaction Ingestion Modal
- **Status**: **DONE**
- **Plan Reference**: [add-transactions-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/add-transactions-implementation-plan.md)
- **Description**: End-to-end transaction entry flow allowing investors to record `BUY`, `SELL`, `DIVIDEND`, `DEPOSIT`, and `WITHDRAWAL` transactions. Features dynamic instrument lookup, trade date validation, cash balance checks, automatic projection rebuilds, and instantaneous UI refresh without full page reload.
- **Key Files**:
  - [portfolio.proto](file:///Users/oscargarcia/workspace/graphfolio/proto/portfolio/v1/portfolio.proto)
  - [AddTransactionModal.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/main-app/src/components/AddTransactionModal.tsx)
  - [AddTransactionModal.css](file:///Users/oscargarcia/workspace/graphfolio/web/apps/main-app/src/components/AddTransactionModal.css)
  - [transaction.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/service/transaction.go)

### 2.7 Interactive SVG Performance Chart & Time Range Filtering
- **Status**: **DONE**
- **Plan Reference**: [performance-chart-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/performance-chart-implementation-plan.md)
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
- **Plan Reference**: [transaction-history-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/transaction-history-implementation-plan.md)
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
- **Plan Reference**: [web-workspace-refactoring-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/web-workspace-refactoring-plan.md)
- **Description**: Transitioned `web/` to an npm workspace monorepo supporting multiple frontend applications. Cleanly decouples the primary user-facing application (`main-app` on port 5173) and the internal administrative portal (`admin-app` on port 5174) while eliminating configuration duplication through shared `tsconfig.base.json`, Oxlint configs, and Vite bundler settings.
- **Key Packages & Files**:
  - **Shared Design System**: [`@graphfolio/ui`](file:///Users/oscargarcia/workspace/graphfolio/web/packages/ui) with tokens, reset, formatters, and atomic components (`Button`, `Modal`, `Card`, `Badge`, `Table`, `Input`, `Select`).
  - **Shared API Client**: [`@graphfolio/api-client`](file:///Users/oscargarcia/workspace/graphfolio/web/packages/api-client) with GenQL generated client and singleton factory.
  - **Primary Application**: [`@graphfolio/main-app`](file:///Users/oscargarcia/workspace/graphfolio/web/apps/main-app)
  - **Internal Admin Portal**: [`@graphfolio/admin-app`](file:///Users/oscargarcia/workspace/graphfolio/web/apps/admin-app) (Asset Management, Price Management, Ingestion Pipeline)
  - [Makefile](file:///Users/oscargarcia/workspace/graphfolio/Makefile) targets: `run-web`, `run-admin`, `run-all-web`, `build-web`, `install-web`

### 2.10 Admin Portal Backend & BFF (Asset, Price & Ingestion Management)
- **Status**: **DONE**
- **Plan Reference**: [admin-portal-backend-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/admin-portal-backend-implementation-plan.md)
- **Description**: Full administrative operations and market orchestration backend. Extends `portfolio-api` and GraphQL BFF with master instrument directory management (asset listing, creation with exchange code and ISO 6166 ISIN validation, active/inactive toggles), closing price ledger queries (`portfolio.instrument_prices`) and manual price overrides with audit justifications and retroactive valuation recalibrations, and ingestion pipeline diagnostics reporting feed health, ECB FX fixings, rate limit budgets, and on-demand market data synchronization. Connected directly to `@graphfolio/admin-app`.
- **Key Files**:
  - [portfolio.proto](file:///Users/oscargarcia/workspace/graphfolio/proto/portfolio/v1/portfolio.proto)
  - [admin.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/service/admin.go)
  - [postgres.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/repository/postgres.go)
  - [schema.graphqls](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.graphqls)
  - [schema.resolvers.go](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.resolvers.go)
  - [AssetManagement.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/admin-app/src/components/AssetManagement.tsx)
  - [PriceManagement.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/admin-app/src/components/PriceManagement.tsx)
  - [IngestionPipeline.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/admin-app/src/components/IngestionPipeline.tsx)

### 2.11 Market Data Ingestion Pipeline (Daily EOD Pricing, ECB FX & Backfill Synchronizer)
- **Status**: **DONE**
- **Plan Reference**: [market-data-ingestion-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/market-data-ingestion-implementation-plan.md)
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

---

## 3. Planned Features (`PLANNED`)

### 3.1 Dynamic Cost Basis Method Switching (Average Cost ↔ FIFO)
- **Status**: **PLANNED**
- **Plan Reference**: [cost-basis-switching-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/cost-basis-switching-implementation-plan.md)
- **Scope**: `proto/`, `services/portfolio-api`, `bff/`, `web/`
- **Highlights**:
  - Segmented toggle control: `[ Average Cost | FIFO ]` in dashboard header.
  - Dynamically switches `portfolio.portfolios.cost_basis_method`.
  - Replays historical transaction ledger to instantly update realized capital gains, holding cost basis, and total return percentages.
  - Info tooltip explaining tax optimization differences (pooling vs selling oldest shares).

### 3.2 User Preferences & Display Currency
- **Status**: **PLANNED**
- **Plan Reference**: [user-preferences-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/user-preferences-implementation-plan.md)
- **Scope**: `proto/user/v1/`, `services/user-api`, `bff/`, `web/`
- **Highlights**:
  - Implementation of `user-api` gRPC microservice on `:50052` backed by `users` PostgreSQL schema.
  - Viewing and editing user preferences: Display Name, Theme (`DARK`, `LIGHT`, `SYSTEM`), and Display Currency.
  - Multi-currency selector (`USD`, `EUR`, `GBP`, `AUD`, `CAD`, `JPY`, `CHF`).
  - Dynamic currency conversion across portfolio valuation and holding metrics via live FX rates.
  - `UserPreferencesModal` component triggered from the user profile badge in the navigation bar.

### 3.3 Portfolio Valuation Engine & Historical Backfill
- **Status**: **PLANNED**
- **Plan Reference**: [portfolio-valuation-engine-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/portfolio-valuation-engine-implementation-plan.md)
- **Scope**: `services/portfolio-api`
- **Highlights**:
  - Automated daily End-of-Day (EOD) portfolio valuation scheduler and snapshot engine.
  - Historical ledger replay worker computing daily valuation points from historical transactions + historical market close & FX rates.
  - Daily sub-period return and cumulative Time-Weighted Return (TWR) index calculations with zero-drift decimal arithmetic.
  - Automatic historical replay hook triggered upon past-dated transaction additions.

### 3.4 Market Data Ingestion (Asset Prices & FX Rates)
- **Status**: **PLANNED**
- **Plan Reference**: [market-data-ingestion-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/market-data-ingestion-implementation-plan.md)
- **Scope**: `services/portfolio-api` (`internal/marketdata/`, `cmd/market-ingest/`)
- **Highlights**:
  - Automated End-of-Day (EOD) closing price and official FX fixing rate ingestion pipeline.
  - Multi-provider adapter architecture (Twelve Data, Yahoo Finance, European Central Bank SDMX/XML, Open Exchange Rates).
  - High-throughput batch upsert queries with idempotent conflict handling for `portfolio.instrument_prices` and `portfolio.fx_rates`.
  - Token-bucket rate limiting (`golang.org/x/time/rate`), exponential backoff retry policies, and price spike anomaly detection.
  - Automatic historical price and FX backfill triggered upon transaction ingestion for unpriced assets.

### 3.5 Look-Through Fundamental & Cash Flow Quality Engine
- **Status**: **PLANNED**
- **Plan Reference**: [fundamental-cash-flow-engine-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/fundamental-cash-flow-engine-implementation-plan.md)
- **Scope**: `proto/`, `services/portfolio-api`, `bff/`, `web/`
- **Highlights**:
  - Table `portfolio.instrument_fundamentals` tracking GAAP/IFRS audited line items (NOPAT, Invested Capital, OCF, Maintenance CapEx, Diluted Shares, NIBCL).
  - Proportional look-through calculations: investor's exact share of Operating Cash Flow, Owner Earnings, FCF, and Revenue.
  - Portfolio Capital Allocation Scorecard: Weighted Average ROIC, Reinvestment Rate, Intrinsic Compounding Rate ($g = \text{ROIC} \times \text{RR}$), and Economic Value Added spread ($\text{ROIC} - \text{WACC}$).
  - Earnings quality detection via the Sloan Accrual Ratio to highlight divergence between reported Net Income and cash realization.
  - "Business Owner" dashboard view contrasting Market Value growth against Business Intrinsic Value growth.

### 3.6 Currency Pair Historical Prices (Admin Portal)
- **Status**: **PLANNED**
- **Plan Reference**: [currency-pair-historical-prices-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/currency-pair-historical-prices-implementation-plan.md)
- **Scope**: `proto/portfolio/v1/`, `services/portfolio-api`, `bff/`, `web/apps/admin-app`
- **Highlights**:
  - Dedicated "Currency Pairs (FX)" management view in the internal Admin Portal (`:5174`).
  - Interactive SVG historical exchange rate trend chart with timeframe filtering (`1W`, `1M`, `3M`, `1Y`, `ALL`) and hover crosshairs.
  - Dynamic currency pair selector with reciprocal rate inversion toggle (`Base ⇄ Quote`, e.g. `EUR/USD` ↔ `USD/EUR`) computed with exact 10-decimal fixed-point precision.
  - Real-time operational KPI telemetry: Spot Rate, 24h/1D change (% and delta), Period High/Low, Total Historical Fixings, and Data Source benchmark (`ECB`, `Twelve Data`).
  - Filterable, paginated authoritative rates ledger (`portfolio.fx_rates`) with date pickers, direct & inverted rates, and manual rate override modal dialog.
  - End-to-end gRPC, GraphQL BFF resolvers, and typed GenQL client integration strictly adhering to the zero floating-point arithmetic policy.

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

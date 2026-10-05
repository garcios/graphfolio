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
| **9** | **Dynamic Cost Basis Method Switching (Average Cost ↔ FIFO)** | **PLANNED** | Proto, Svc, BFF, Web | [cost-basis-switching-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/cost-basis-switching-implementation-plan.md) |
| **10** | **User Preferences & Display Currency** | **PLANNED** | Proto, User API, BFF, Web | [user-preferences-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/user-preferences-implementation-plan.md) |
| **11** | **Portfolio Valuation Engine & Historical Backfill** | **PLANNED** | Database, Svc, Worker | [portfolio-valuation-engine-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/portfolio-valuation-engine-implementation-plan.md) |
| **12** | **Market Data Ingestion (Asset Prices & FX Rates)** | **PLANNED** | Database, Svc, Worker | [market-data-ingestion-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/market-data-ingestion-implementation-plan.md) |
| **13** | **Tax Lot Inspector & Capital Gains Reports** | **ROADMAP** | Svc, BFF, Web | *(Future Plan)* |
| **14** | **Dividend Calendar & Yield Analytics** | **ROADMAP** | Svc, BFF, Web | *(Future Plan)* |
| **15** | **Real-Time Market Data & WebSocket Price Ticker** | **ROADMAP** | Market Data, Svc, Web | *(Future Plan)* |

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
  - [Dashboard.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/src/components/Dashboard.tsx)
  - [Dashboard.css](file:///Users/oscargarcia/workspace/graphfolio/web/src/components/Dashboard.css)
  - [main.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/src/main.tsx)

### 2.6 Interactive Transaction Ingestion Modal
- **Status**: **DONE**
- **Plan Reference**: [add-transactions-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/add-transactions-implementation-plan.md)
- **Description**: End-to-end transaction entry flow allowing investors to record `BUY`, `SELL`, `DIVIDEND`, `DEPOSIT`, and `WITHDRAWAL` transactions. Features dynamic instrument lookup, trade date validation, cash balance checks, automatic projection rebuilds, and instantaneous UI refresh without full page reload.
- **Key Files**:
  - [portfolio.proto](file:///Users/oscargarcia/workspace/graphfolio/proto/portfolio/v1/portfolio.proto)
  - [AddTransactionModal.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/src/components/AddTransactionModal.tsx)
  - [AddTransactionModal.css](file:///Users/oscargarcia/workspace/graphfolio/web/src/components/AddTransactionModal.css)
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
  - [PerformanceChart.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/src/components/PerformanceChart.tsx)
  - [PerformanceChart.css](file:///Users/oscargarcia/workspace/graphfolio/web/src/components/PerformanceChart.css)

### 2.8 Transaction History & Ledger Management
- **Status**: **DONE**
- **Plan Reference**: [transaction-history-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/transaction-history-implementation-plan.md)
- **Description**: Full transaction ledger history with server-side pagination, multi-criteria filtering by transaction type and symbol, and safe transaction deletion with automatic deterministic projection replay (`RebuildProjections`). Deleting an entry atomically clears derived lots, disposals, holdings, and cash, re-running the ledger to update metrics and UI immediately.
- **Key Files**:
  - [portfolio.proto](file:///Users/oscargarcia/workspace/graphfolio/proto/portfolio/v1/portfolio.proto)
  - [transaction.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/service/transaction.go)
  - [schema.graphqls](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.graphqls)
  - [schema.resolvers.go](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.resolvers.go)
  - [TransactionLedger.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/src/components/TransactionLedger.tsx)
  - [TransactionLedger.css](file:///Users/oscargarcia/workspace/graphfolio/web/src/components/TransactionLedger.css)

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

### 3.3 User Preferences & Display Currency
- **Status**: **PLANNED**
- **Plan Reference**: [user-preferences-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/user-preferences-implementation-plan.md)
- **Scope**: `proto/user/v1/`, `services/user-api`, `bff/`, `web/`
- **Highlights**:
  - Implementation of `user-api` gRPC microservice on `:50052` backed by `users` PostgreSQL schema.
  - Viewing and editing user preferences: Display Name, Theme (`DARK`, `LIGHT`, `SYSTEM`), and Display Currency.
  - Multi-currency selector (`USD`, `EUR`, `GBP`, `AUD`, `CAD`, `JPY`, `CHF`).
  - Dynamic currency conversion across portfolio valuation and holding metrics via live FX rates.
  - `UserPreferencesModal` component triggered from the user profile badge in the navigation bar.

### 3.4 Portfolio Valuation Engine & Historical Backfill
- **Status**: **PLANNED**
- **Plan Reference**: [portfolio-valuation-engine-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/portfolio-valuation-engine-implementation-plan.md)
- **Scope**: `services/portfolio-api`
- **Highlights**:
  - Automated daily End-of-Day (EOD) portfolio valuation scheduler and snapshot engine.
  - Historical ledger replay worker computing daily valuation points from historical transactions + historical market close & FX rates.
  - Daily sub-period return and cumulative Time-Weighted Return (TWR) index calculations with zero-drift decimal arithmetic.
  - Automatic historical replay hook triggered upon past-dated transaction additions.

### 3.5 Market Data Ingestion (Asset Prices & FX Rates)
- **Status**: **PLANNED**
- **Plan Reference**: [market-data-ingestion-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/market-data-ingestion-implementation-plan.md)
- **Scope**: `services/portfolio-api` (`internal/marketdata/`, `cmd/market-ingest/`)
- **Highlights**:
  - Automated End-of-Day (EOD) closing price and official FX fixing rate ingestion pipeline.
  - Multi-provider adapter architecture (Twelve Data, Yahoo Finance, European Central Bank SDMX/XML, Open Exchange Rates).
  - High-throughput batch upsert queries with idempotent conflict handling for `portfolio.instrument_prices` and `portfolio.fx_rates`.
  - Token-bucket rate limiting (`golang.org/x/time/rate`), exponential backoff retry policies, and price spike anomaly detection.
  - Automatic historical price and FX backfill triggered upon transaction ingestion for unpriced assets.

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

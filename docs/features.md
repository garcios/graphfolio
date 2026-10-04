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
| **7** | **Interactive SVG Performance Chart & Time Range Filtering** | **PLANNED** | Proto, Svc, BFF, Web | [performance-chart-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/performance-chart-implementation-plan.md) |
| **8** | **Transaction History & Ledger Management** | **PLANNED** | Proto, Svc, BFF, Web | [transaction-history-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/transaction-history-implementation-plan.md) |
| **9** | **Dynamic Cost Basis Method Switching (Average Cost ↔ FIFO)** | **PLANNED** | Proto, Svc, BFF, Web | [cost-basis-switching-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/cost-basis-switching-implementation-plan.md) |
| **10** | **User Preferences & Display Currency** | **PLANNED** | Proto, User API, BFF, Web | [user-preferences-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/user-preferences-implementation-plan.md) |
| **11** | **Tax Lot Inspector & Capital Gains Reports** | **ROADMAP** | Svc, BFF, Web | *(Future Plan)* |
| **12** | **Dividend Calendar & Yield Analytics** | **ROADMAP** | Svc, BFF, Web | *(Future Plan)* |

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

---

## 3. Planned Features (`PLANNED`)

### 3.1 Interactive SVG Performance Chart & Time Range Filtering
- **Status**: **PLANNED**
- **Plan Reference**: [performance-chart-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/performance-chart-implementation-plan.md)
- **Scope**: `proto/`, `services/portfolio-api`, `bff/`, `web/`
- **Highlights**:
  - Replaces static placeholder asset with an interactive SVG performance chart.
  - Time range selector buttons: `1W`, `1M`, `3M`, `YTD`, `1Y`, `ALL`.
  - Queries daily Time-Weighted Return (TWR) and valuation curves from `portfolio.portfolio_valuations`.
  - Interactive crosshairs, tooltips showing date and valuation, gradient fills, and performance delta pills.

### 3.2 Transaction History & Ledger Management
- **Status**: **PLANNED**
- **Plan Reference**: [transaction-history-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/transaction-history-implementation-plan.md)
- **Scope**: `proto/`, `services/portfolio-api`, `bff/`, `web/`
- **Highlights**:
  - Paginated transaction history table (`1–20 of N`) with window function `COUNT(*) OVER()`.
  - Multi-criteria filtering by transaction type (`BUY`, `SELL`, `DIVIDEND`, `DEPOSIT`, `WITHDRAWAL`) and ticker symbol.
  - Safe transaction deletion with warning modal.
  - Deleting an entry triggers automatic projection replay (`RebuildProjections`), recalculating open tax lots, disposals, holdings, and cash atomically.

### 3.3 Dynamic Cost Basis Method Switching (Average Cost ↔ FIFO)
- **Status**: **PLANNED**
- **Plan Reference**: [cost-basis-switching-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/cost-basis-switching-implementation-plan.md)
- **Scope**: `proto/`, `services/portfolio-api`, `bff/`, `web/`
- **Highlights**:
  - Segmented toggle control: `[ Average Cost | FIFO ]` in dashboard header.
  - Dynamically switches `portfolio.portfolios.cost_basis_method`.
  - Replays historical transaction ledger to instantly update realized capital gains, holding cost basis, and total return percentages.
  - Info tooltip explaining tax optimization differences (pooling vs selling oldest shares).

### 3.4 User Preferences & Display Currency
- **Status**: **PLANNED**
- **Plan Reference**: [user-preferences-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/user-preferences-implementation-plan.md)
- **Scope**: `proto/user/v1/`, `services/user-api`, `bff/`, `web/`
- **Highlights**:
  - Implementation of `user-api` gRPC microservice on `:50052` backed by `users` PostgreSQL schema.
  - Viewing and editing user preferences: Display Name, Theme (`DARK`, `LIGHT`, `SYSTEM`), and Display Currency.
  - Multi-currency selector (`USD`, `EUR`, `GBP`, `AUD`, `CAD`, `JPY`, `CHF`).
  - Dynamic currency conversion across portfolio valuation and holding metrics via live FX rates.
  - `UserPreferencesModal` component triggered from the user profile badge in the navigation bar.

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

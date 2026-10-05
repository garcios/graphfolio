# Agent Guidelines & Repository Architecture (`AGENTS.md`)

This file contains the architectural rules, coding standards, and verification procedures that all AI agents operating in the **GraphFolio** repository must adhere to.

---

## 1. Project Overview

A modern portfolio tracker built for serious investors. Move beyond simple price charts to accurately measure your total annualized returns—including dividends, currency fluctuations, and corporate actions—all in one clear, consolidated dashboard.

---

## 2. Layered Architecture & Separation of Concerns

### Layer Breakdown
- **`proto/` (The Contract)**:
  Protocol Buffers are the single source of truth for microservice RPCs and shared types. Stored at the root to eliminate schema drift between gRPC servers and the BFF client wrappers.
  - `proto/common/v1/decimal.proto`: High-precision fixed-point `Decimal` (value, scale) and `Money` (amount, currency_code) types.
  - `proto/portfolio/v1/portfolio.proto`: Portfolio service (`GetPortfolio`, `AddTransaction`, `ListInstruments`, `GetPortfolioHistory`, `ListTransactions`, `DeleteTransaction`, `ListAllInstruments`, `CreateInstrument`, `UpdateInstrument`, `ListInstrumentPrices`, `RecordPriceOverride`, `GetIngestionStatus`, `TriggerMarketSync`), investment summary, time-series valuation points, performance metrics, and administrative operations.

- **`pkg/` (Shared Infrastructure)**:
  Strictly non-domain-specific code shared across backend modules:
  - `pkg/database/`: PostgreSQL connection pooling (`pgx/v5`), auto-discovery of `.env` files, automated `shopspring/decimal` type registration on every new connection, DSN query parameter sanitization (`SanitizeDSN`), and embedded migration runner (`golang-migrate`).
  - `pkg/decimalpb/`: Exact conversions between protobuf `Decimal`/`Money` and `shopspring/decimal.Decimal`.
  - `pkg/logger/`: Standardized structured logging.
  - `pkg/middleware/`: gRPC interceptors (authentication, metrics, tracing).

- **`services/` (The Domain Microservices)**:
  Each microservice is an isolated Go module (`services/portfolio-api`, `services/user-api`):
  - Domain models and calculators live in `internal/domain/` (`portfolio`, `holding`, `transaction`, `instrument`, `price`, `ingestion`, `tax_lot`, `history`).
  - Persistence logic lives in `internal/repository/` with `pgxpool.Pool` queries and transactions (`InsertTransaction`, `ListTransactions`, `DeleteTransaction`, `FindInstrumentBySymbol`, `SaveProjectionsTx`, `GetPortfolioValuations`, `ListAllInstruments`, `CreateInstrument`, `UpdateInstrument`, `ListInstrumentPrices`, `UpsertInstrumentPrice`, `GetIngestionMetrics`).
  - Business logic, calculation services, transaction ingestion (`AddTransaction`), transaction deletion (`DeleteTransaction`), ledger replay projections (`RebuildProjections`), paginated ledger queries (`ListTransactions`), administrative operations (`CreateInstrument`, `UpdateInstrument`, `RecordPriceOverride`, `GetIngestionStatus`, `TriggerMarketSync`), and historical valuation time-series retrieval (`GetPortfolioHistory`) live in `internal/service/`.
  - Transport adapters (gRPC servers) live in `internal/` (`server.go`) and `cmd/server/main.go`.
  - Microservices only interact with their dedicated database schemas and other gRPC APIs. They have zero awareness of GraphQL.

- **`bff/` (The Orchestrator)**:
  The Backend-for-Frontend is a Go service using `gqlgen` (GraphQL). It translates GraphQL queries (`portfolio`, `instruments`, `portfolioHistory`, `transactions`, `allInstruments`, `instrumentPrices`, `ingestionStatus`) and mutations (`addTransaction`, `deleteTransaction`, `createInstrument`, `updateInstrument`, `recordPriceOverride`, `triggerMarketSync`) into gRPC calls across domain microservices, maps exact decimal scalars, stitches data together, and shields the frontend from microservice topology.

- **`web/` (The Consumer)**:
  Workspace monorepo containing multiple frontend applications and shared libraries:
  - `apps/main-app`: Primary investor-facing application (Port 5173). Interactive portfolio dashboard, performance chart, transaction ledger, and trade ingestion modal.
  - `apps/admin-app`: Internal administrative portal (Port 5174). Master instrument/asset directory management, closing price ledger, manual price overrides, and market ingestion monitoring.
  - `packages/ui`: Shared design system (`@graphfolio/ui`) with dark glassmorphic design tokens, atomic components (`Button`, `Modal`, `Card`, `Badge`, `Table`, `Input`, `Select`), and precision financial formatters.
  - `packages/api-client`: Shared auto-generated typed GraphQL client (`@graphfolio/api-client`) communicating with the BFF.

---

## 3. Project Structure

```text
├── proto/                        # 1. Single Source of Truth for APIs
│   ├── common/v1/
│   │   └── decimal.proto         # Decimal and Money contracts
│   ├── portfolio/v1/
│   │   └── portfolio.proto       # Portfolio gRPC service (GetPortfolio, AddTransaction, ListInstruments, GetPortfolioHistory, ListTransactions, DeleteTransaction)
│   └── user/v1/
│       └── user.proto            # User gRPC service definition
│
├── pkg/                          # 2. Shared Libraries (Backend)
│   ├── database/                 # pgxpool connection pooling, config, migration runner
│   │   ├── config.go             # Environment parsing with automatic .env loading
│   │   ├── postgres.go           # Pool factory with decimal type registration
│   │   └── migrate.go            # Embedded golang-migrate runner
│   ├── decimalpb/                # Proto Decimal/Money <-> shopspring conversion
│   └── go.mod
│
├── scripts/                      # 3. Database bootstrap & teardown
│   └── db/
│       ├── bootstrap.sql         # Idempotent DB, role (portfolio_svc, user_svc), and schema creation
│       └── teardown.sql          # DB drop script
│
├── services/                     # 4. Core gRPC Microservices
│   ├── portfolio-api/
│   │   ├── cmd/server/           # Service entrypoint (gRPC on :50051)
│   │   ├── internal/
│   │   │   ├── domain/           # Domain entities (Portfolio, Holding, Transaction, Instrument, Price, Ingestion, TaxLot, Money, History)
│   │   │   │   ├── transaction.go# Transaction, TransactionFilter, TransactionPage
│   │   │   │   ├── price.go      # InstrumentPrice, PriceFilter, PriceOverrideInput
│   │   │   │   ├── ingestion.go  # FeedHealth, IngestionStatus
│   │   │   │   ├── history.go    # ValuationPoint, PortfolioHistory, CalculatePortfolioHistory
│   │   │   │   └── ...
│   │   │   ├── repository/       # Repository interface & PostgreSQL (pgx) queries
│   │   │   │   ├── postgres.go   # Queries: ListTransactions, ListAllInstruments, UpsertInstrumentPrice, etc.
│   │   │   │   └── mocks/        # Uber-go mock repository (MockRepository)
│   │   │   ├── service/          # PortfolioService, ledger projection engine, transaction & admin manager
│   │   │   │   ├── transaction.go# AddTransaction, ListTransactions, DeleteTransaction (with RebuildProjections)
│   │   │   │   ├── admin.go      # CreateInstrument, UpdateInstrument, RecordPriceOverride, GetIngestionStatus, TriggerMarketSync
│   │   │   │   ├── admin_test.go # Table-driven unit tests for admin operations
│   │   │   │   ├── transaction_test.go # Unit tests for transaction ingestion, filtering, and deletion
│   │   │   │   ├── history.go    # GetPortfolioHistory and timeframe boundary logic
│   │   │   │   ├── history_test.go# Unit tests for history filtering and period return math
│   │   │   │   └── mocks/        # Uber-go mock service (MockPortfolioService)
│   │   │   ├── server.go         # gRPC PortfolioServiceServer implementation (ListTransactions, RecordPriceOverride, etc.)
│   │   │   └── server_test.go    # gRPC server unit tests with MockPortfolioService
│   │   ├── migrations/           # Schema migrations (000001 to 000006)
│   │   ├── seeds/                # Development seed data (dev_seed.sql with 365-day history)
│   │   └── go.mod
│   └── user-api/
│       ├── cmd/server/           # Service entrypoint (gRPC on :50052)
│       ├── migrations/           # User schema migrations (000001)
│       └── go.mod
│
├── bff/                          # 5. GraphQL Backend-for-Frontend
│   ├── cmd/server/               # BFF entrypoint (GraphQL server on :8080)
│   ├── graph/
│   │   ├── schema.graphqls       # GraphQL schema (Queries: portfolio, transactions; Mutations: addTransaction, deleteTransaction)
│   │   ├── schema.resolvers.go   # Resolver implementations calling gRPC (portfolio, portfolioHistory, transactions, deleteTransaction)
│   │   ├── schema.resolvers_test.go # Unit tests for resolvers
│   │   ├── helpers.go            # Domain-to-GraphQL conversion helpers (history, transactions, timeframe)
│   │   └── model/
│   │       ├── decimal.go        # Custom Decimal scalar unmarshaler/marshaler
│   │       └── models_gen.go     # Generated GraphQL models
│   └── go.mod
│
├── web/                          # 6. Frontend Workspace Monorepo
│   ├── package.json              # Workspace root ("workspaces": ["apps/*", "packages/*"])
│   ├── tsconfig.base.json        # Shared TypeScript compiler options
│   ├── apps/
│   │   ├── main-app/             # Primary Investor Application (:5173)
│   │   │   ├── src/
│   │   │   │   ├── components/   # Dashboard, PerformanceChart, TransactionLedger, AddTransactionModal
│   │   │   │   ├── App.tsx
│   │   │   │   └── main.tsx
│   │   │   ├── vite.config.ts    # Configured with resolve.alias for live package HMR
│   │   │   └── package.json      # @graphfolio/main-app
│   │   └── admin-app/            # Internal Operations Portal (:5174)
│   │       ├── src/
│   │       │   ├── components/   # AdminLayout, AssetManagement, PriceManagement, IngestionPipeline
│   │       │   ├── App.tsx
│   │       │   └── main.tsx
│   │       ├── vite.config.ts    # Port 5174 with package aliases
│   │       └── package.json      # @graphfolio/admin-app
│   └── packages/
│       ├── ui/                   # Shared Design System (@graphfolio/ui)
│       │   ├── src/
│       │   │   ├── styles/       # tokens.css, reset.css
│       │   │   ├── components/   # Button, Modal, Card, Badge, Table, Input, Select
│       │   │   └── utils/        # formatMoney, formatPercent, formatDate, cn
│       │   └── package.json
│       └── api-client/           # Shared GraphQL Client (@graphfolio/api-client)
│           ├── src/
│           │   ├── generated/    # Target for `make generate`
│           │   └── client.ts     # Configurable GraphQL client singleton
│           └── package.json
│
├── docs/                         # Architecture designs & implementation plans
│   ├── db-implementation-plan.md
│   ├── portfolio-service-implementation-plan.md
│   ├── add-transactions-implementation-plan.md
│   ├── performance-chart-implementation-plan.md
│   ├── portfolio-valuation-engine-implementation-plan.md
│   ├── market-data-ingestion-implementation-plan.md
│   ├── transaction-history-implementation-plan.md
│   ├── cost-basis-switching-implementation-plan.md
│   ├── user-preferences-implementation-plan.md
│   ├── competitive-analysis.md
│   ├── fundamental-cash-flow-engine-implementation-plan.md
│   └── genql-usage.md
│
├── FEATURES.md                   # Master features matrix & roadmap (DONE, PLANNED, ROADMAP)
├── Makefile                      # Standardized commands (make proto, make generate, make run, make test)
├── go.work                       # Go workspace mapping (bff, pkg, proto, portfolio-api, user-api)
└── .env.example                  # Environment variable template
```

---

## 4. Coding & Architecture Standards

### 4.1 Exact Precision & Currency Arithmetic
- **Zero Floating-Point Policy**: Never use `float32` or `float64` for money, asset quantities, prices, or percentages.
- Always use `github.com/shopspring/decimal` for Go domain types and arithmetic.
- Use `pkg/decimalpb` to convert between Go decimals and protobuf `common.v1.Decimal` / `common.v1.Money`.
- In GraphQL, money values use the custom `Decimal` scalar; never coerce to IEEE 754 float in the BFF.

### 4.2 Database Isolation & Schemas
- Local PostgreSQL 17+ runs on `localhost:5432`.
- Each microservice owns its own isolated schema within the `graphfolio` database:
  - `portfolio-api` owns the `portfolio` schema and connects as `portfolio_svc`.
  - `user-api` owns the `users` schema and connects as `user_svc`.
- Migration tool: `golang-migrate` tracks versions in separate schema tables (`schema_migrations` within each schema).
- All DSN strings must be sanitized through `database.SanitizeDSN` before passing to `pgxpool.ParseConfig`.
- All `pgxpool` connections must have the `pgx-shopspring-decimal` extension registered via `poolCfg.AfterConnect`.

### 4.3 Ledger Replay & Cost Basis Methods
- The transaction ledger (`portfolio.transactions`) is the source of truth for portfolio history.
- Adding transactions (`BUY`, `SELL`, `DEPOSIT`, `WITHDRAWAL`, `DIVIDEND`) or deleting existing transactions (`DeleteTransaction`) alters `portfolio.transactions` and immediately invokes `service.RebuildProjections`.
- Holding balances, tax lots, lot disposals, and cash balances are deterministic projections built by `service.ProcessLedger`.
- Projection rebuilds acquire a row lock (`SELECT ... FOR UPDATE` on `portfolio.portfolios`) to serialize concurrent updates and atomically save recomputed state.
- Both `AVERAGE_COST` (default) and `FIFO` cost-basis relieve algorithms are supported in `internal/service/projection.go`.

### 4.4 Mocking & Unit Testing Standards
- **Mock Generation**: Generated with `go.uber.org/mock` (`mockgen`).
  - Repository mock: `services/portfolio-api/internal/repository/mocks/mock_repository.go`
  - Service mock: `services/portfolio-api/internal/service/mocks/mock_service.go`
  - Generated via `make generate` or `go generate ./...`.
- **Zero 3rd-Party Assertions**: Use standard library `testing` package exclusively (`t.Run`, `t.Errorf`, `t.Fatalf`). Do **NOT** use `testify` (`assert`/`require`).
- **Decimal Assertions**: Compare decimals using `.Equal()`, `.IsZero()`, or `.IsPositive()`; never compare structs with `==`.
- **Race Condition Safety**: All tests must pass race detection: `go test -v -race ./...`.

### 4.5 Historical Valuations & Performance Metrics
- `portfolio.portfolio_valuations` stores daily valuation snapshots (total value, cash balance, cost basis) per portfolio.
- Historical queries (`GetPortfolioHistory` in gRPC, `portfolioHistory` in GraphQL) accept a timeframe parameter (`1D`, `1W`, `1M`, `YTD`, `1Y`, `ALL`).
- Service boundaries determine cutoff dates in UTC (`time.Now().UTC()`); `YTD` starts at January 1st 00:00:00 UTC of the current year.
- Period returns compute net value change (`end_value - start_value`) and percentage return (`(end_value - start_value) / start_value * 100`) using exact `shopspring/decimal.Decimal` arithmetic with zero-start fallback handling.
- The web frontend renders historical points using a custom interactive SVG cubic Bezier curve (`PerformanceChart`) with hovering crosshairs and dynamic positive/negative gradient styling.

### 4.6 Transaction Ledger, Pagination & Deletion Integrity
- Paginated transaction queries use the `TransactionFilter` domain model (`PortfolioID`, `InstrumentID`, `Type`, `Limit`, `Offset`).
- Persistence layer uses PostgreSQL single-pass `COUNT(*) OVER() AS total_count` window function with dynamic SQL filters to retrieve transactions and exact total counts in a single round-trip.
- Transaction deletion (`DeleteTransaction`) validates ownership, deletes the ledger record in `portfolio.transactions`, and immediately executes `RebuildProjections` under an exclusive portfolio lock to atomically clean up orphaned tax lots, holding projections, and recompute cash balances.
- The web frontend renders historical records in `TransactionLedger` with type-specific color badges, date/amount formatting, pagination controls, and confirmation modals before triggering deletion.

### 4.7 Master Instrument Directory, Pricing Overrides & Ingestion Diagnostics
- **Asset Directory Management**: `portfolio.instruments` stores authoritative reference data. Adding an instrument requires validating non-empty uppercase symbol and exchange code, valid ISO 4217 3-letter currency code, allowed asset class (`EQUITY`, `ETF`, `FUND`, `BOND`, `CRYPTO`, `CASH_EQUIVALENT`), and optional 12-character ISO 6166 ISIN format (`^[A-Z]{2}[A-Z0-9]{9}[0-9]$`).
- **Authoritative Price Ledger & Overrides**: `portfolio.instrument_prices` tracks daily closing marks and vendor sources. Manual price overrides (`RecordPriceOverride`) enforce positive decimal prices (`price.IsPositive()`), record mandatory audit justification notes with source `'manual'`, and optionally trigger retroactive recalibrations of portfolio valuations (`recompute_valuations = true`).
- **Ingestion Telemetry & On-Demand Synchronization**: gRPC `GetIngestionStatus` aggregates active instrument counts, tracked currency pairs, latest pricing dates, and token-bucket budget metrics alongside feed status metadata (`Twelve Data`, `Yahoo Finance`, `ECB`). On-demand market synchronization (`TriggerMarketSync`) enables immediate batch pricing updates and FX triangulation checks.
- **Admin Portal UI**: The internal operations portal (`apps/admin-app` on `:5174`) directly integrates with the BFF via live GenQL queries and mutations across `AssetManagement`, `PriceManagement`, and `IngestionPipeline` views.

---

## 5. Verification Commands

Before committing changes, execute the following commands to ensure consistency and test health:

```bash
# 1. Regenerate code contracts, GraphQL models, genql client, and Go mocks
make generate

# 2. Check formatting
test -z "$(gofmt -s -l services/ pkg/ bff/)" || (echo "Unformatted files found:" && gofmt -s -l services/ pkg/ bff/ && exit 1)

# 3. Static analysis
go vet ./services/portfolio-api/... ./pkg/... ./bff/...

# 4. Run test suites with race detection
make test
```
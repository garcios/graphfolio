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
  - `proto/portfolio/v1/portfolio.proto`: Portfolio service (`GetPortfolio`, `UpdatePortfolioBaseCurrency`, `AddTransaction`, `ListInstruments`, `GetPortfolioHistory`, `ListTransactions`, `DeleteTransaction`, `ListAllInstruments`, `CreateInstrument`, `UpdateInstrument`, `ListInstrumentPrices`, `RecordPriceOverride`, `GetIngestionStatus`, `TriggerMarketSync`, `RebuildValuations`), investment summary, time-series valuation points, performance metrics, and administrative operations.
  - `proto/user/v1/user.proto`: User service (`GetUserPreferences`, `UpdateUserPreferences`, `ListSupportedCurrencies`), user profile preferences (`theme`, `display_currency`, `display_name`), and currency reference metadata.

- **`pkg/` (Shared Infrastructure)**:
  Strictly non-domain-specific code shared across backend modules:
  - `pkg/database/`: PostgreSQL connection pooling (`pgx/v5`), auto-discovery of `.env` files, automated `shopspring/decimal` type registration on every new connection, DSN query parameter sanitization (`SanitizeDSN`), and embedded migration runner (`golang-migrate`).
  - `pkg/decimalpb/`: Exact conversions between protobuf `Decimal`/`Money` and `shopspring/decimal.Decimal`.
  - `pkg/logger/`: Standardized structured logging.
  - `pkg/middleware/`: gRPC interceptors (authentication, metrics, tracing).

- **`services/` (The Domain Microservices)**:
  Each microservice is an isolated Go module (`services/portfolio-api`, `services/user-api`):
  - **`services/portfolio-api` (:50051)**:
    - Domain models and calculators live in `internal/domain/` (`portfolio`, `holding`, `transaction`, `instrument`, `price`, `ingestion`, `tax_lot`, `history`, `valuation`).
    - Persistence logic lives in `internal/repository/` with `pgxpool.Pool` queries and transactions (`InsertTransaction`, `ListTransactions`, `DeleteTransaction`, `UpdatePortfolioBaseCurrency`, `FindInstrumentBySymbol`, `SaveProjectionsTx`, `GetPortfolioValuations`, `ListAllInstruments`, `CreateInstrument`, `UpdateInstrument`, `ListInstrumentPrices`, `UpsertInstrumentPrice`, `GetIngestionMetrics`, `UpsertValuationsBatch`, `GetLatestValuationBefore`, `GetHistoricalPriceMatrix`, `GetHistoricalFXMatrix`, `DeleteValuationsFromDate`, `ListActivePortfolios`).
    - Business logic, calculation services, transaction ingestion (`AddTransaction`), transaction deletion (`DeleteTransaction`), ledger replay projections (`RebuildProjections`), dynamic base currency re-anchoring (`UpdatePortfolioBaseCurrency`), paginated ledger queries (`ListTransactions`), administrative operations (`CreateInstrument`, `UpdateInstrument`, `RecordPriceOverride`, `GetIngestionStatus`, `TriggerMarketSync`), historical valuation time-series retrieval (`GetPortfolioHistory`), and daily valuation snapshots & multi-day historical backfill replay engine (`ValuationService`, `RebuildValuations`) live in `internal/service/`.
    - Transport adapters (gRPC servers) live in `internal/` (`server.go`), `cmd/server/main.go`, and scheduled CLI / background valuation worker in `cmd/worker/main.go`.
  - **`services/user-api` (:50052)**:
    - Domain models live in `internal/domain/` (`User`, `CurrencyInfo`, `SupportedCurrencies`).
    - Persistence logic lives in `internal/repository/` with `pgxpool.Pool` queries (`GetUserByID`, `UpdateUserPreferences`).
    - Business logic, input validation (display name bounds, theme enum, ISO 4217 currencies), and demo user alias resolution (`1` ↔ `018f0000-0000-7000-8000-000000000001`) live in `internal/service/`.
    - Transport adapters (gRPC servers) live in `internal/` (`server.go`) and `cmd/server/main.go`.
  - Microservices only interact with their dedicated database schemas and other gRPC APIs. They have zero awareness of GraphQL.

- **`bff/` (The Orchestrator)**:
  The Backend-for-Frontend is a Go service using `gqlgen` (GraphQL). It translates GraphQL queries (`portfolio`, `instruments`, `portfolioHistory`, `transactions`, `allInstruments`, `instrumentPrices`, `ingestionStatus`, `userPreferences`, `supportedCurrencies`) and mutations (`addTransaction`, `deleteTransaction`, `createInstrument`, `updateInstrument`, `recordPriceOverride`, `triggerMarketSync`, `updateUserPreferences`) into gRPC calls across domain microservices, maps exact decimal scalars, orchestrates dynamic currency re-anchoring between `user-api` and `portfolio-api`, and shields the frontend from microservice topology.

- **`web/` (The Consumer)**:
  Workspace monorepo containing multiple frontend applications and shared libraries:
  - `apps/main-app`: Primary investor-facing application (Port 5173). Interactive portfolio dashboard, performance chart, transaction ledger, trade ingestion modal, and investor profile preferences modal (`UserPreferencesModal`).
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
│   │   └── portfolio.proto       # Portfolio gRPC service (GetPortfolio, UpdatePortfolioBaseCurrency, AddTransaction, ListInstruments, GetPortfolioHistory, ListTransactions, DeleteTransaction)
│   └── user/v1/
│       └── user.proto            # User gRPC service definition (GetUserPreferences, UpdateUserPreferences, ListSupportedCurrencies)
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
│   │   ├── cmd/worker/           # Scheduled EOD & on-demand valuation worker (runnable via make run-valuation-job)
│   │   ├── internal/
│   │   │   ├── domain/           # Domain entities (Portfolio, Holding, Transaction, Instrument, Price, Ingestion, TaxLot, Money, History, Valuation)
│   │   │   │   ├── transaction.go# Transaction, TransactionFilter, TransactionPage
│   │   │   │   ├── valuation.go  # PortfolioValuationSnapshot, sub-period return, TWR linking, GIPS CAGR
│   │   │   │   ├── price.go      # InstrumentPrice, PriceFilter, PriceOverrideInput
│   │   │   │   ├── ingestion.go  # FeedHealth, IngestionStatus
│   │   │   │   ├── history.go    # ValuationPoint, PortfolioHistory, CalculatePortfolioHistory
│   │   │   │   └── ...
│   │   │   ├── repository/       # Repository interface & PostgreSQL (pgx) queries
│   │   │   │   ├── postgres.go   # Queries: ListTransactions, ListAllInstruments, UpsertValuationsBatch, GetHistoricalPriceMatrix, etc.
│   │   │   │   └── mocks/        # Uber-go mock repository (MockRepository)
│   │   │   ├── service/          # PortfolioService, ValuationService, ledger projection engine, transaction & admin manager
│   │   │   │   ├── transaction.go# AddTransaction, ListTransactions, DeleteTransaction (with RebuildProjections & valuation hooks)
│   │   │   │   ├── valuation.go  # ValuationService (SnapshotValuation, BackfillPortfolioValuations, RunDailyValuationJob)
│   │   │   │   ├── valuation_test.go # Multi-day historical backfill replay and daily job unit tests
│   │   │   │   ├── admin.go      # CreateInstrument, UpdateInstrument, RecordPriceOverride, GetIngestionStatus, TriggerMarketSync
│   │   │   │   ├── admin_test.go # Table-driven unit tests for admin operations
│   │   │   │   ├── transaction_test.go # Unit tests for transaction ingestion, filtering, deletion, and valuation hooks
│   │   │   │   ├── history.go    # GetPortfolioHistory and timeframe boundary logic
│   │   │   │   ├── history_test.go# Unit tests for history filtering and period return math
│   │   │   │   └── mocks/        # Uber-go mock service (MockPortfolioService, MockValuationService)
│   │   │   ├── server.go         # gRPC PortfolioServiceServer implementation (ListTransactions, RecordPriceOverride, RebuildValuations, etc.)
│   │   │   └── server_test.go    # gRPC server unit tests with MockPortfolioService
│   │   ├── migrations/           # Schema migrations (000001 to 000006)
│   │   ├── seeds/                # Development seed data (dev_seed.sql with 365-day history)
│   │   └── go.mod
│   └── user-api/
│       ├── cmd/server/           # Service entrypoint (gRPC on :50052)
│       ├── internal/             # Domain entities, repository, service, and gRPC server
│       │   ├── domain/           # Domain entities (User, Theme, Currency)
│       │   ├── repository/       # Repository interface & PostgreSQL (pgx) queries
│       │   ├── service/          # UserService (preferences, validation, supported currencies)
│       │   ├── server.go         # gRPC UserServiceServer implementation
│       │   └── server_test.go    # gRPC server unit tests with MockUserService
│       ├── migrations/           # User schema migrations (000001_initial_schema, 000002_add_theme)
│       ├── seeds/                # User dev fixtures (dev_seed.sql with demo investor)
│       └── go.mod
│
├── bff/                          # 5. GraphQL Backend-for-Frontend
│   ├── cmd/server/               # BFF entrypoint (GraphQL server on :8080)
│   ├── graph/
│   │   ├── schema.graphqls       # GraphQL schema (Queries: portfolio, transactions, userPreferences; Mutations: addTransaction, deleteTransaction, updateUserPreferences)
│   │   ├── schema.resolvers.go   # Resolver implementations calling gRPC (portfolio, portfolioHistory, transactions, deleteTransaction, userPreferences, updateUserPreferences)
│   │   ├── schema.resolvers_test.go # Unit tests for resolvers
│   │   ├── helpers.go            # Domain-to-GraphQL conversion helpers (history, transactions, timeframe, preferences)
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
│   │   │   │   ├── components/   # Dashboard, PerformanceChart, TransactionLedger, AddTransactionModal, UserPreferencesModal
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
├── docs/                         # Architecture designs, plans & documentation
│   ├── plans/                    # System implementation plans
│   │   ├── db-implementation-plan.md
│   │   ├── portfolio-service-implementation-plan.md
│   │   ├── add-transactions-implementation-plan.md
│   │   ├── performance-chart-implementation-plan.md
│   │   ├── portfolio-valuation-engine-implementation-plan.md
│   │   ├── market-data-ingestion-implementation-plan.md
│   │   ├── transaction-history-implementation-plan.md
│   │   ├── cost-basis-switching-implementation-plan.md
│   │   ├── user-preferences-implementation-plan.md
│   │   ├── fundamental-cash-flow-engine-implementation-plan.md
│   │   ├── currency-pair-historical-prices-implementation-plan.md
│   │   ├── admin-historical-backfill-implementation-plan.md
│   │   ├── ingestion-job-history-implementation-plan.md
│   │   └── web-workspace-refactoring-plan.md
│   ├── user-stories/             # Agile user stories & acceptance criteria
│   ├── competitive-analysis.md
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

### 4.8 Market Data Ingestion Pipeline & FX Triangulation Standards
- **Vendor Interfaces & Zero Float Policy**: All external feeds (`TwelveDataProvider`, `YahooFinanceProvider`, `ECBProvider`) implement `marketdata.PriceProvider` or `marketdata.FXRateProvider`. Closing marks and exchange rates are parsed strictly into `shopspring/decimal.Decimal` (using `json.Number` for JSON or string parsing for XML) with zero IEEE 754 float drift.
- **ECB XML & 10-Decimal Triangulation**: European Central Bank reference fixings parse official XML feeds. `TriangulationEngine` resolves direct, inverse, and cross-currency rates using an anchor currency (default `EUR`) with exact 10-decimal precision. Last Observation Carried Forward (LOCF) carries the last available trading day mark over weekends and market holidays.
- **Resilience & Rate Limiting**: Outbound HTTP traffic is governed by `RateLimiter` (`golang.org/x/time/rate`) and exponential backoff retry with randomized jitter (`ExecuteWithRetry`) on transient HTTP 429 and 5xx status codes.
- **Transaction Ingestion Backfill**: Recording a trade in `AddTransaction` for an unpriced instrument automatically verifies range coverage (`HasPricesForRange`) and triggers historical backfill (`BackfillInstrumentPrices`) before `RebuildProjections` updates holding valuations and performance curves.
- **Scheduled & On-Demand CLI Ingestion**: `cmd/market-ingest/main.go` (runnable via `make ingest-market-data`) orchestrates daily batch synchronization and historical date-range backfills.

### 4.9 User Identity, Personal Preferences & Dynamic Currency Re-anchoring
- **User Domain Microservice Isolation**: `services/user-api` (Port `:50052`) owns the `users` schema (`users.users`), completely decoupled from the portfolio ledger.
- **Preference Invariants & Validation**: User profile models validate non-empty display names (1-100 characters), supported ISO 4217 currency codes (`USD`, `EUR`, `GBP`, `AUD`, `CAD`, `JPY`, `CHF`), and supported theme options (`DARK`, `LIGHT`, `SYSTEM`).
- **Dynamic Portfolio Base Currency Re-anchoring**: Calling `UpdatePortfolioBaseCurrency` on `portfolio-api` atomically updates `portfolio.portfolios.base_currency` and scales both historical valuation time-series (`portfolio.portfolio_valuations`) and holding cost bases (`portfolio.holdings`) using official exchange rates via direct, inverse, or triangulated market marks.
- **BFF Orchestration & Currency Synchronization**: The BFF ensures single-source-of-truth consistency across microservices:
  - When fetching the portfolio (`Query.portfolio`), the BFF verifies that the portfolio's base currency matches the user's preferred currency, automatically triggering `UpdatePortfolioBaseCurrency` if a mismatch is detected.
  - When mutating preferences (`Mutation.updateUserPreferences`), the BFF updates `user-api` and immediately synchronizes `portfolio-api` within the same workflow, returning the newly re-scaled portfolio state.
- **Reactive UI Synchronization**: The investor frontend (`apps/main-app`) re-renders cards, tables, and metric badges upon preference changes, and explicitly remounts the SVG `PerformanceChart` with the new currency key to prevent canvas and tooltip currency ghosting.

### 4.10 Portfolio Valuation Engine, Historical Backfill & Scheduled Worker
- **Mathematical Modeling & Zero Drift**: Daily valuation points compute exact market value ($V_t = \sum \text{Holding Value} + \sum \text{Cash Value}$), net cash flows ($F_t = \text{Deposits} - \text{Withdrawals}$), sub-period return ($R_t = \frac{V_t - (V_{t-1} + F_t)}{V_{t-1} + F_t}$), and cumulative Time-Weighted Return ($\text{TWR}_t = \text{TWR}_{t-1} \times (1 + R_t)$) with exact `shopspring/decimal.Decimal` fixed-point precision. First-funding deposits and zero-start edge cases isolate market movement with $0\%$ return baseline.
- **GIPS Annualized Return (CAGR)**: Cumulative compounding returns over periods exceeding 365.25 days are annualized via $( \text{TWR}_t )^{365.25 / \text{days}} - 1$. Periods under 365.25 days report non-annualized cumulative returns in compliance with GIPS 2.A.20 standards.
- **LOCF Matrix Retrieval & Triangulation**: Historical replay fetches asset closing prices and FX rates across multi-day windows using optimized matrix queries with Last Observation Carried Forward (LOCF) to prevent weekend/holiday missing marks.
- **Transaction Ingestion & Deletion Hooks**: `AddTransaction` automatically triggers `BackfillPortfolioValuations` when `tradeDate.Before(todayUTC)` and records `SnapshotValuation` on current-day trades. `DeleteTransaction` triggers retrospective backfill starting from `portfolio.CreatedAt`.
- **Scheduled EOD Worker & On-Demand CLI**: `services/portfolio-api/cmd/worker/main.go` (runnable via `make run-valuation-job`) supports daily valuation snapshots across all active portfolios (`RunDailyValuationJob`), historical date backfills (`-from`), specific portfolio UUID targets (`-portfolio`), and a continuous background daemon mode (`-daemon`, `-cron-hour=22` UTC market close).

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
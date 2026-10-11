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
  - `proto/portfolio/v1/portfolio.proto`: Portfolio service (`GetPortfolio`, `UpdatePortfolioBaseCurrency`, `AddTransaction`, `ListInstruments`, `GetPortfolioHistory`, `ListTransactions`, `DeleteTransaction`, `ListAllInstruments`, `CreateInstrument`, `UpdateInstrument`, `DeleteInstrument`, `ListExchanges`, `ListInstrumentPrices`, `RecordPriceOverride`, `GetIngestionStatus`, `TriggerMarketSync`, `TriggerBackfill`, `RebuildValuations`, `ListCurrencyPairs`, `ListFXRates`, `GetCurrencyPairHistory`, `RecordFXRateOverride`, `CheckTransactionDuplicates`, `BatchImportTransactions`), investment summary with 3-pillar return attribution (`capital_gain_amount`, `capital_gain_percent`, `income_amount`, `income_yield_percent`, `currency_gain_amount`, `currency_gain_percent`, `is_international`), time-series valuation points, performance metrics, corporate actions (`SPLIT`), and administrative foreign exchange / pricing operations.
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
    - Domain models and calculators live in `internal/domain/` (`portfolio`, `holding`, `transaction`, `instrument`, `exchange`, `price`, `ingestion`, `tax_lot`, `history`, `valuation`, `fx`, `backfill`, `calculator` with exact 3-pillar multi-currency return attribution).
    - Persistence logic lives in `internal/repository/` with `pgxpool.Pool` queries and transactions (`InsertTransaction`, `ListTransactions`, `DeleteTransaction`, `UpdatePortfolioBaseCurrency`, `FindInstrumentBySymbol`, `SaveProjectionsTx`, `GetPortfolioValuations`, `ListAllInstruments`, `CreateInstrument`, `UpdateInstrument`, `DeleteInstrument`, `ListExchanges`, `ListInstrumentPrices`, `UpsertInstrumentPrice`, `GetIngestionMetrics`, `UpsertValuationsBatch`, `GetLatestValuationBefore`, `GetHistoricalPriceMatrix`, `GetHistoricalFXMatrix`, `DeleteValuationsFromDate`, `ListActivePortfolios`, `ListActiveInstruments`, `ListActiveCurrencies`, `FindPortfoliosHoldingInstrument`, `ListCurrencyPairs`, `GetCurrencyPairHistory`, `ListFXRates`, `UpsertFXRate`, `FindExistingExternalRefs`, `BatchInsertTransactions`).
    - Business logic, calculation services, transaction ingestion (`AddTransaction`), transaction deletion (`DeleteTransaction`), batch statement import (`BatchImportTransactions`), external reference deduplication (`CheckTransactionDuplicates`), ledger replay projections (`RebuildProjections`), dynamic base currency re-anchoring (`UpdatePortfolioBaseCurrency`), paginated ledger queries (`ListTransactions`), administrative operations (`CreateInstrument`, `UpdateInstrument`, `DeleteInstrument`, `ListExchanges`, `RecordPriceOverride`, `GetIngestionStatus`, `TriggerMarketSync`, `TriggerBackfill`), foreign exchange management & rate overrides (`ListCurrencyPairs`, `GetCurrencyPairHistory`, `ListFXRates`, `RecordFXRateOverride`), historical valuation time-series retrieval (`GetPortfolioHistory`), and daily valuation snapshots & multi-day historical backfill replay engine (`ValuationService`, `RebuildValuations`) live in `internal/service/`.
    - Transport adapters (gRPC servers) live in `internal/` (`server.go`), `cmd/server/main.go`, and scheduled CLI / background valuation worker in `cmd/worker/main.go`.
  - **`services/user-api` (:50052)**:
    - Domain models live in `internal/domain/` (`User`, `CurrencyInfo`, `SupportedCurrencies`).
    - Persistence logic lives in `internal/repository/` with `pgxpool.Pool` queries (`GetUserByID`, `UpdateUserPreferences`).
    - Business logic, input validation (display name bounds, theme enum, ISO 4217 currencies), and demo user alias resolution (`1` ↔ `018f0000-0000-7000-8000-000000000001`) live in `internal/service/`.
    - Transport adapters (gRPC servers) live in `internal/` (`server.go`) and `cmd/server/main.go`.
  - Microservices only interact with their dedicated database schemas and other gRPC APIs. They have zero awareness of GraphQL.

- **`bff/` (The Orchestrator)**:
  The Backend-for-Frontend is a Go service using `gqlgen` (GraphQL). It translates GraphQL queries (`portfolio`, `instruments`, `portfolioHistory`, `transactions`, `allInstruments`, `exchanges`, `instrumentPrices`, `ingestionStatus`, `userPreferences`, `supportedCurrencies`, `currencyPairs`, `currencyPairHistory`, `fxRates`, `checkTransactionDuplicates`) and mutations (`addTransaction`, `deleteTransaction`, `importTransactions`, `createInstrument`, `updateInstrument`, `deleteInstrument`, `recordPriceOverride`, `recordFXRateOverride`, `triggerMarketSync`, `triggerBackfill`, `updateUserPreferences`) into gRPC calls across domain microservices, maps exact decimal scalars, exposes investment return attribution fields (`capitalGainAmount`, `incomeAmount`, `currencyGainAmount`, `isInternational`), orchestrates dynamic currency re-anchoring between `user-api` and `portfolio-api`, and shields the frontend from microservice topology.

- **`web/` (The Consumer)**:
  Workspace monorepo containing multiple frontend applications and shared libraries:
  - `apps/main-app`: Primary investor-facing application (Port 5173). Interactive portfolio dashboard with widened desktop layout (`max-width: 1680px`) and 10-column holdings return attribution breakdown (`Asset`, `Price`, `Avg Buy Price`, `Quantity`, `Total Value`, `Capital Gain`, `Income`, `Currency Gain`, `Total Return`, `Today's Return`), performance chart, transaction ledger, trade ingestion modal (`AddTransactionModal` with Buy, Sell, Deposit, Withdrawal, Dividend, and Split support), multi-broker CSV statement import modal (`ImportTransactionsModal`), client-side CSV parsers (`src/services/csv/`), and investor profile preferences modal (`UserPreferencesModal`).
  - `apps/admin-app`: Internal administrative portal (Port 5174). Master instrument directory, closing price ledger with dual date range filtering, presets, and pagination (`PriceManagement`), foreign exchange currency pairs management (`FXManagement`, `FXTrendChart`, `FXOverrideModal`), market ingestion monitoring, and historical market data range backfills (`BackfillModal`).
  - `packages/ui`: Shared design system (`@graphfolio/ui`) with dark glassmorphic design tokens, atomic components (`Button`, `Modal`, `Card`, `Badge`, `Table`, `Input`, `Select`), and precision financial formatters.
  - `packages/api-client`: Shared auto-generated typed GraphQL client (`@graphfolio/api-client`) communicating with the BFF.

---

## 3. Project Structure

```text
├── proto/                        # 1. Single Source of Truth for APIs
│   ├── common/v1/
│   │   └── decimal.proto         # Decimal and Money contracts
│   ├── portfolio/v1/
│   │   └── portfolio.proto       # Portfolio gRPC service (GetPortfolio, UpdatePortfolioBaseCurrency, AddTransaction, ListInstruments, GetPortfolioHistory, ListTransactions, DeleteTransaction, admin, FX & TriggerBackfill RPCs)
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
│   │   │   ├── domain/           # Domain entities (Portfolio, Holding, Transaction, Instrument, Price, Ingestion, TaxLot, Money, History, Valuation, FX, Backfill)
│   │   │   │   ├── transaction.go# Transaction, TransactionFilter, TransactionPage
│   │   │   │   ├── backfill.go   # BackfillInput, BackfillResult (historical market data backfill jobs)
│   │   │   │   ├── valuation.go  # PortfolioValuationSnapshot, sub-period return, TWR linking, GIPS CAGR
│   │   │   │   ├── price.go      # InstrumentPrice, PriceFilter, PriceOverrideInput
│   │   │   │   ├── fx.go         # CurrencyPair, FXRate, FXHistoryPoint, CurrencyPairHistory
│   │   │   │   ├── ingestion.go  # FeedHealth, IngestionStatus
│   │   │   │   ├── history.go    # ValuationPoint, PortfolioHistory, CalculatePortfolioHistory
│   │   │   │   └── ...
│   │   │   ├── repository/       # Repository interface & PostgreSQL (pgx) queries
│   │   │   │   ├── postgres.go   # Queries: ListTransactions, ListAllInstruments, UpsertValuationsBatch, GetHistoricalPriceMatrix, ListCurrencyPairs, etc.
│   │   │   │   └── mocks/        # Uber-go mock repository (MockRepository)
│   │   │   ├── service/          # PortfolioService, ValuationService, ledger projection engine, transaction & admin manager
│   │   │   │   ├── transaction.go# AddTransaction, ListTransactions, DeleteTransaction (with RebuildProjections & valuation hooks)
│   │   │   │   ├── valuation.go  # ValuationService (SnapshotValuation, BackfillPortfolioValuations, RunDailyValuationJob)
│   │   │   │   ├── valuation_test.go # Multi-day historical backfill replay and daily job unit tests
│   │   │   │   ├── fx.go         # Foreign exchange service (ListCurrencyPairs, GetCurrencyPairHistory, ListFXRates, RecordFXRateOverride)
│   │   │   │   ├── fx_test.go    # FX unit tests for period bounds, 10-decimal reciprocal math, and overrides
│   │   │   │   ├── admin.go      # CreateInstrument, UpdateInstrument, RecordPriceOverride, GetIngestionStatus, TriggerMarketSync, TriggerBackfill
│   │   │   │   ├── admin_test.go # Table-driven unit tests for admin operations (incl. backfill validation, scope resolution, partial failures)
│   │   │   │   ├── ingestion.go  # IngestionService (IngestDailyMarketData, BackfillInstrumentPrices, BackfillCurrencyPair)
│   │   │   │   ├── transaction_test.go # Unit tests for transaction ingestion, filtering, deletion, and valuation hooks
│   │   │   │   ├── history.go    # GetPortfolioHistory and timeframe boundary logic
│   │   │   │   ├── history_test.go# Unit tests for history filtering and period return math
│   │   │   │   └── mocks/        # Uber-go mocks (MockPortfolioService, MockValuationService, MockIngestionService)
│   │   │   ├── server.go         # gRPC PortfolioServiceServer implementation (ListTransactions, RecordPriceOverride, RebuildValuations, ListCurrencyPairs, TriggerBackfill, etc.)
│   │   │   └── server_test.go    # gRPC server unit tests with MockPortfolioService
│   │   ├── migrations/           # Schema migrations (000001 to 000009)
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
│   │   ├── schema.graphqls       # GraphQL schema (Queries: portfolio, transactions, userPreferences; Mutations: addTransaction, deleteTransaction, triggerBackfill, updateUserPreferences)
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
│   │       │   ├── components/   # AdminLayout, AssetManagement, PriceManagement, FXManagement, FXTrendChart, FXOverrideModal, BackfillModal, IngestionPipeline
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
│   │   ├── asset-prices-date-range-filter-implementation-plan.md
│   │   ├── ingestion-job-history-implementation-plan.md
│   │   ├── multi-broker-csv-ingestion-implementation-plan.md
│   │   ├── cash-statement-csv-ingestion-implementation-plan.md
│   │   ├── holding-return-attribution-implementation-plan.md
│   │   └── web-workspace-refactoring-plan.md
│   ├── user-stories/             # Agile user stories & acceptance criteria
│   │   └── investment-holding-capital-gain-income-currency.md
│   ├── negative-cash-balance-investigation-report.md # Cash ledger overdraft root cause & remediation report
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
- Adding transactions (`BUY`, `SELL`, `DEPOSIT`, `WITHDRAWAL`, `DIVIDEND`, `SPLIT`) or deleting existing transactions (`DeleteTransaction`) alters `portfolio.transactions` and immediately invokes `service.RebuildProjections`.
- Holding balances, tax lots, lot disposals, and cash balances are deterministic projections built by `service.ProcessLedger`.
- Projection rebuilds acquire a row lock (`SELECT ... FOR UPDATE` on `portfolio.portfolios`) to serialize concurrent updates and atomically save recomputed state.
- Both `AVERAGE_COST` (default) and `FIFO` cost-basis relieve algorithms are supported in `internal/service/projection.go`.
- **Corporate Action Stock Splits (`SPLIT`)**:
  - Encoded with a ratio where `quantity` represents the new share factor and `price` represents the old share factor (e.g., a 2-for-1 split records `quantity = 2`, `price = 1`, yielding split multiplier $M = \frac{2}{1} = 2.0$; a 1-for-2 reverse split records `quantity = 1`, `price = 2`, yielding multiplier $M = 0.5$).
  - Evaluated in `service.ProcessLedger` as a zero-cash transaction that scales the share quantities of all open tax lots acquired prior to the split date: $Q_{\text{lot, new}} = Q_{\text{lot, old}} \times M$. Both `OriginalQuantity` and `RemainingQuantity` are scaled.
  - Total invested cost basis of each tax lot is invariant, which proportionally adjusts the effective unit cost basis ($C_{\text{unit, new}} = \frac{C_{\text{total}}}{Q_{\text{lot, new}}}$) and preserves cash balances with zero cash flow impact ($0.00$).
  - Subsequent disposals (`SELL`) relieve split-adjusted tax lots at their newly calibrated unit cost basis under both `AVERAGE_COST` and `FIFO`.
  - Deleting a split transaction restores pre-split lot quantities and unadjusted unit costs during the deterministic ledger replay.

### 4.4 Mocking & Unit Testing Standards
- **Mock Generation**: Generated with `go.uber.org/mock` (`mockgen`).
  - Repository mock: `services/portfolio-api/internal/repository/mocks/mock_repository.go`
  - Service mock: `services/portfolio-api/internal/service/mocks/mock_service.go`
  - Ingestion service mock: `services/portfolio-api/internal/service/mocks/mock_ingestion_service.go` (`MockIngestionService`)
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
- The web frontend renders historical records in `TransactionLedger` with type-specific color badges (including glowing purple badge and split multiplier for `SPLIT` transactions), date/amount formatting, pagination controls, and confirmation modals before triggering deletion.

### 4.7 Master Instrument Directory, Pricing Overrides & Ingestion Diagnostics
- **Asset Directory Management**: `portfolio.instruments` stores authoritative reference data. Adding an instrument requires validating non-empty uppercase symbol and exchange code, valid ISO 4217 3-letter currency code, allowed asset class (`EQUITY`, `ETF`, `FUND`, `BOND`, `CRYPTO`, `CASH_EQUIVALENT`), and optional 12-character ISO 6166 ISIN format (`^[A-Z]{2}[A-Z0-9]{9}[0-9]$`).
- **Authoritative Price Ledger & Overrides**: `portfolio.instrument_prices` tracks daily closing marks and vendor sources. Manual price overrides (`RecordPriceOverride`) enforce positive decimal prices (`price.IsPositive()`), record mandatory audit justification notes with source `'manual'`, and optionally trigger retroactive recalibrations of portfolio valuations (`recompute_valuations = true`).
- **Ingestion Telemetry & On-Demand Synchronization**: gRPC `GetIngestionStatus` aggregates active instrument counts, tracked currency pairs, latest pricing dates, and token-bucket budget metrics alongside feed status metadata (`Twelve Data`, `Yahoo Finance`, `ECB`). On-demand market synchronization (`TriggerMarketSync`) enables immediate batch pricing updates and FX triangulation checks.
- **Date Range Filtering & Ledger Pagination**: `portfolio.instrument_prices` queries in `ListInstrumentPrices` support bounded and half-open date ranges (`FromDate`, `ToDate`) and clamped pagination (`Limit` [1, 200] defaulting to 50, `Offset` >= 0). `PriceManagement` in `apps/admin-app` enforces client-side date ordering guards (`fromDate <= toDate`), provides quick range preset buttons (`7D`, `30D`, `90D`, `YTD`, `1Y`), passes active bounds into `BackfillModal`, and renders glassmorphic pagination controls with tabular page indicators and Prev/Next navigation.
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

### 4.11 Foreign Exchange (FX) Currency Pair Analytics, History & Rate Ledger Standards
- **Authoritative FX Ledger**: The PostgreSQL table `portfolio.fx_rates` stores daily closing reference marks and vendor sources (`ECB`, `TWELVE_DATA`, `manual`). Direct and inverted currency rates are queried via `ListCurrencyPairs`, `GetCurrencyPairHistory`, and `ListFXRates`.
- **Zero Float & 10-Decimal Reciprocal Math**: Never use IEEE 754 floating-point arithmetic for exchange rates. All rates in Go use `shopspring/decimal.Decimal`, protobuf uses `common.v1.Decimal`, and GraphQL uses the custom `Decimal` scalar. Reciprocal rates are calculated with exact 10-decimal precision: `decimal.NewFromInt(1).DivRound(rate, 10)`.
- **Timeframe Boundary Resolution & High/Low**: `GetCurrencyPairHistory` calculates date boundaries in UTC (`time.Now().UTC()`) for `1W` (7 days), `1M` (30 days), `1Y` (365 days), and `ALL`. Computes start/end rates, absolute period change, percentage change, and scans period high and low extremes with exact decimal comparison (`.GreaterThan()`, `.LessThan()`).
- **Manual Rate Overrides & Audit Trail**: Manual rate overrides via gRPC `RecordFXRateOverride` and GraphQL `recordFXRateOverride` enforce positive rates (`rate.IsPositive()`), reject future dates (`rateDate > today`), record mandatory audit justification notes with source prefix `'manual: '`, and optionally trigger retroactive recalibrations of all affected portfolio valuations (`recompute_valuations = true`).
- **Admin Portal UI Integration**: The internal operations portal (`apps/admin-app` on `:5174`) provides a dedicated `FXManagement` view featuring interactive SVG trend curves (`FXTrendChart`) with direct/reciprocal inversion toggles (`Base ⇄ Quote`), real-time spot rate and 1D delta KPI cards, date range filtering, and modal dialogs (`FXOverrideModal`) for audit-trail rate overrides.

### 4.12 Admin Historical Market Data Backfill (Asset Prices & FX Rates)
- **End-to-End Contract**: gRPC `TriggerBackfill(TriggerBackfillRequest)` → GraphQL `triggerBackfill(input: TriggerBackfillInput!): BackfillPayload!`. Dates travel as ISO `YYYY-MM-DD` strings and are parsed into `domain.BackfillInput` (`internal/domain/backfill.go`); results map from `domain.BackfillResult` (`success`, `pricesSynced`, `fxRatesSynced`, `message`, `warnings`).
- **Validation Invariants** (`PortfolioService.TriggerBackfill` in `internal/service/admin.go`): both dates required (`ErrDateRequired`), normalized to UTC midnight; `toDate` must not precede `fromDate` (`ErrInvalidDateRange`); no future dates (`ErrFutureDate`); maximum window of 1826 days / ~5 years (`ErrBackfillRangeTooLarge`); at least one target enabled (`ErrNoBackfillTarget`). The gRPC layer defaults an empty `to_date` to today (UTC), maps these to `codes.InvalidArgument`, and all other failures to `codes.Internal`.
- **Scope Resolution**: Assets — explicit `symbols` (trimmed, uppercased, resolved via `FindInstrumentBySymbol`) or all active instruments (`ListActiveInstruments`). FX — explicit `currencyPairs` (`"EUR/USD"` parsed by `parseCurrencyPair`) or pairs generated from `ListActiveCurrencies` via `generateCurrencyPairs`.
- **Partial-Failure Semantics**: Unknown symbols, malformed pairs, and per-instrument / per-pair provider errors are collected as non-fatal `warnings`; the batch continues. Only repository failures listing active instruments abort the job.
- **Valuation Reconciliation**: When `recomputeValuations = true`, `BackfillPortfolioValuations(fromDate)` runs only for portfolios holding the backfilled instruments (`FindPortfoliosHoldingInstrument`); FX-only backfills that synced rates recompute all active portfolios (`ListActivePortfolios`).
- **BFF Defaults**: `backfillAssets` and `backfillFx` default to `true`, `recomputeValuations` defaults to `false`; `warnings` is always a non-null list.
- **Admin Portal UI**: `BackfillModal` (`apps/admin-app`) provides date presets (`30D`, `90D`, `YTD`, `1Y`, `ALL`), All-Active vs Specific scope selection with symbol/pair chips, API token budget impact projection from `ingestionStatus`, and an optional recompute toggle. Integrated in `IngestionPipeline` (Backfill Console), `PriceManagement` ("Backfill Asset Prices", pre-filled with the active symbol), and `FXManagement` ("Backfill FX Rates", pre-filled with the active pair); each view refreshes its data on completion.

### 4.13 Multi-Broker CSV Ingestion & Flexible Symbol Resolution Standards
- **Multi-Broker Dialect Parsing**: Client-side parsers (`CommSec` and `nabtrade` in `apps/main-app/src/services/csv/`) support leading Australian retail brokers. Automatically detects format via header signature matching (`detectBrokerFormat`), normalizes Australian date formats (`DD/MM/YYYY`), strips non-numeric currency characters and accounting negative parentheses, matches nabtrade international and domestic cash account trade debits/credits with native AUD fee derivation, and isolates legal metadata headers/footers.
- **Cash Account Statement Parsing**: Ingests full nabtrade cash account statements (`Date,Type,Description,Debit,Credit,Balance`). Automatically maps cash deposits (`DEPOSIT` from `FUNDS TRANSFER ... deposit {name}` with confirmation ref `nabtrade:cash:{ref}`), cash withdrawals (`WITHDRAWAL` from `... transfer {name}` on debit), monthly interest credits (`INTEREST` from `INTEREST` on credit), and domestic ETF/stock dividends (`DIVIDEND` from `FUNDS TRANSFER DIVIDEND - {TICKER}` with ticker extraction and company-name alias matching). Informational policy rows (`InterestChange`, `Please note from...`) are silently skipped without raising parse errors.
- **Idempotency & Deduplication**: Trade confirmation numbers and SHA-256 hashed transaction attributes are stored in `portfolio.transactions.external_ref`. The UI performs pre-flight batch deduplication queries (`CheckTransactionDuplicates` gRPC / `checkTransactionDuplicates` GraphQL) to visually flag existing items. The persistence layer enforces idempotent insertion via `pgx.Batch` with `ON CONFLICT (portfolio_id, external_ref) DO NOTHING`.
- **Flexible Instrument Suffix Resolution**: Database query `findInstrumentBySymbolSQL` and transaction filtering `listTransactionsSQL` resolve tickers whether queried by base root symbol, market provider `.AX` suffix, or Australian broker `.ASX` notation (`WHERE UPPER(symbol) = UPPER($1) OR UPPER(symbol) = UPPER($1) || '.AX' OR (UPPER($1) LIKE '%.AX' AND UPPER(symbol) = UPPER(REPLACE($1, '.AX', ''))) ...`), giving precedence to exact symbol matches.
- **Auto-Provisioning Unmapped Instruments & Empty Symbol Cash Ingestion**: In `BatchImportTransactions`, when an imported transaction specifies an unknown ticker (`ErrInstrumentNotFound`), the service automatically provisions the instrument (e.g. `SYMBOL.AX`, exchange `XASX`, currency `AUD`, asset class `EQUITY`). When an imported row specifies an empty symbol (`item.Symbol == ""` for `DEPOSIT`, `WITHDRAWAL`, or `INTEREST`), the service skips instrument catalog lookups, assigns `instID = nil`, preserves custom currency codes (`AUD`), and updates cash balances upon ledger replay.
- **Single-Pass Projection Rebuild & Valuation Hooks**: Multi-transaction imports execute ledger replay (`RebuildProjections`) once under an exclusive portfolio lock (`SELECT ... FOR UPDATE`), followed by retroactive historical valuation backfills (`BackfillPortfolioValuations`) starting from the earliest imported trade date.
- **Interactive Ingestion UI**: `ImportTransactionsModal` (`apps/main-app`) provides drag-and-drop file upload, broker format selector, KPI metrics cards (total rows, net buy/sell amount, total fees), tabbed preview of valid and duplicate rows with inline error diagnostics, custom status pills (`deposit`, `withdrawal`, `interest`), placeholder dashes for cash movements, and direct ledger integration.

### 4.14 Investment Holding Return Attribution & Multi-Currency Decomposition Standards
- **Three-Pillar Return Decomposition**: `services/portfolio-api/internal/domain/calculator.go` computes an exact 3-pillar return attribution for each holding in `CalculateInvestment`:
  1. **Capital Gain/Loss**: Local price appreciation/depreciation converted to base currency at the effective acquisition exchange rate:
     $$\text{CapGain}_{\text{base}} = (V_{\text{local}} - C_{\text{local}}) \times \overline{\text{FX}}_0$$
     $$\text{CapGain}\% = \frac{V_{\text{local}} - C_{\text{local}}}{C_{\text{local}}} \times 100$$
  2. **Income (Dividends & Distributions)**: Cumulative gross cash distributions received in base currency:
     $$\text{Income}_{\text{base}} = \text{Dividends}_{\text{base}}$$
     $$\text{Yield on Cost (YOC)}\% = \frac{\text{Dividends}_{\text{base}}}{C_{\text{base}}} \times 100$$
  3. **Currency Gain/Loss (FX Fluctuation)**: Currency movement between acquisition and current valuation mark:
     $$\text{FXGain}_{\text{base}} = V_{\text{local}} \times (\text{FX}_t - \overline{\text{FX}}_0)$$
     $$\text{FXGain}\% = \frac{\text{FX}_t - \overline{\text{FX}}_0}{\overline{\text{FX}}_0} \times 100$$
- **Weighted-Average Acquisition FX Rate**: Defined as $\overline{\text{FX}}_0 = \frac{C_{\text{base}}}{C_{\text{local}}}$, weighting the exchange rates of past purchases by their invested cost basis. If local cost is zero ($C_{\text{local}} = 0$), $\overline{\text{FX}}_0$ defaults to current market $\text{FX}_t$.
- **Mathematical Identity Invariant**: For all multi-currency assets, capital gain plus currency gain exactly equals the unrealized gain in base currency:
  $$\text{CapGain}_{\text{base}} + \text{FXGain}_{\text{base}} \equiv V_{\text{base}} - C_{\text{base}}$$
  This identity is mathematically verified in domain unit tests with zero tolerance for floating-point error.
- **Domestic Asset Neutrality**: For domestic holdings (where instrument currency equals portfolio base currency or $\text{FX}_t = 1.0$), `IsInternational = false`, `CurrencyGainAmount = 0.00`, and `CurrencyGainPercent = 0.00%`. The frontend renders a clean neutral dash (`—`) in place of currency gain metrics.
- **Unpriced & Zero Cost Handling**: Unpriced assets fallback to local valuation equal to local cost ($V_{\text{local}} = C_{\text{local}}$), yielding $0.00$ capital gain and $0.00\%$ return while safely passing through accumulated dividend income.
- **Responsive Presentation & Expanded Container**: `web/apps/main-app` renders the 10-column return attribution table (`Asset`, `Price`, `Avg Buy Price`, `Quantity`, `Total Value`, `Capital Gain`, `Income`, `Currency Gain`, `Total Return`, `Today's Return`) within an expanded `.app-container` (`max-width: 1680px` on desktop) to eliminate horizontal scrolling. International assets feature a subtle cyan `INTL` badge. The table footer includes an itemized subtotal row summing portfolio-wide Capital Gain, Cumulative Income, and Currency Gain.

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
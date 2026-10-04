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
  - `proto/portfolio/v1/portfolio.proto`: Portfolio service (`GetPortfolio`, `AddTransaction`, `ListInstruments`), investment summary, and performance metrics.

- **`pkg/` (Shared Infrastructure)**:
  Strictly non-domain-specific code shared across backend modules:
  - `pkg/database/`: PostgreSQL connection pooling (`pgx/v5`), auto-discovery of `.env` files, automated `shopspring/decimal` type registration on every new connection, DSN query parameter sanitization (`SanitizeDSN`), and embedded migration runner (`golang-migrate`).
  - `pkg/decimalpb/`: Exact conversions between protobuf `Decimal`/`Money` and `shopspring/decimal.Decimal`.
  - `pkg/logger/`: Standardized structured logging.
  - `pkg/middleware/`: gRPC interceptors (authentication, metrics, tracing).

- **`services/` (The Domain Microservices)**:
  Each microservice is an isolated Go module (`services/portfolio-api`, `services/user-api`):
  - Domain models and calculators live in `internal/domain/` (`portfolio`, `holding`, `transaction`, `instrument`, `tax_lot`).
  - Persistence logic lives in `internal/repository/` with `pgxpool.Pool` queries and transactions (`InsertTransaction`, `FindInstrumentBySymbol`, `SaveProjectionsTx`).
  - Business logic, calculation services, transaction ingestion (`AddTransaction`), and ledger replay projections (`RebuildProjections`) live in `internal/service/`.
  - Transport adapters (gRPC servers) live in `internal/` (`server.go`) and `cmd/server/main.go`.
  - Microservices only interact with their dedicated database schemas and other gRPC APIs. They have zero awareness of GraphQL.

- **`bff/` (The Orchestrator)**:
  The Backend-for-Frontend is a Go service using `gqlgen` (GraphQL). It translates GraphQL queries (`portfolio`, `instruments`) and mutations (`addTransaction`) into gRPC calls across domain microservices, maps exact decimal scalars, stitches data together, and shields the frontend from microservice topology.

- **`web/` (The Consumer)**:
  Standalone Vite + React + TypeScript frontend. Interacts exclusively with the BFF GraphQL endpoint using a strongly-typed auto-generated client (`genql`). Includes interactive dashboard and transaction ingestion modal (`AddTransactionModal`).

---

## 3. Project Structure

```text
├── proto/                        # 1. Single Source of Truth for APIs
│   ├── common/v1/
│   │   └── decimal.proto         # Decimal and Money contracts
│   ├── portfolio/v1/
│   │   └── portfolio.proto       # Portfolio gRPC service (GetPortfolio, AddTransaction, ListInstruments)
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
│   │   │   ├── domain/           # Domain entities (Portfolio, Holding, Transaction, Instrument, TaxLot, Money)
│   │   │   ├── repository/       # Repository interface & PostgreSQL (pgx) queries
│   │   │   │   └── mocks/        # Uber-go mock repository (MockRepository)
│   │   │   ├── service/          # PortfolioService, transaction ingestion & ledger projection engine
│   │   │   │   └── mocks/        # Uber-go mock service (MockPortfolioService)
│   │   │   ├── server.go         # gRPC PortfolioServiceServer implementation
│   │   │   └── server_test.go    # gRPC server unit tests with MockPortfolioService
│   │   ├── migrations/           # Schema migrations (000001 to 000006)
│   │   ├── seeds/                # Development seed data (dev_seed.sql)
│   │   └── go.mod
│   └── user-api/
│       ├── cmd/server/           # Service entrypoint (gRPC on :50052)
│       ├── migrations/           # User schema migrations (000001)
│       └── go.mod
│
├── bff/                          # 5. GraphQL Backend-for-Frontend
│   ├── cmd/server/               # BFF entrypoint (GraphQL server on :8080)
│   ├── graph/
│   │   ├── schema.graphqls       # GraphQL schema (Queries, Mutations, custom Decimal scalar)
│   │   ├── schema.resolvers.go   # Resolver implementations calling gRPC
│   │   ├── helpers.go            # Domain-to-GraphQL conversion helpers
│   │   └── model/
│   │       ├── decimal.go        # Custom Decimal scalar unmarshaler/marshaler
│   │       └── models_gen.go     # Generated GraphQL models
│   └── go.mod
│
├── web/                          # 6. Frontend Application
│   ├── src/
│   │   ├── components/           # Reusable UI components (Dashboard, AddTransactionModal, etc.)
│   │   ├── generated/            # GenQL auto-generated typed client
│   │   └── main.tsx              # Application entrypoint
│   ├── package.json
│   └── tsconfig.json
│
├── docs/                         # Architecture designs & implementation plans
│   ├── db-implementation-plan.md
│   ├── portfolio-service-implementation-plan.md
│   ├── add-transactions-implementation-plan.md
│   └── genql-usage.md
│
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
- Adding transactions (`BUY`, `SELL`, `DEPOSIT`, `WITHDRAWAL`, `DIVIDEND`) appends records to `portfolio.transactions` and immediately invokes `service.RebuildProjections`.
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
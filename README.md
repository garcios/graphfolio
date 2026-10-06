# GraphFolio

A modern portfolio tracker built for serious investors. GraphFolio accurately measures your total annualized returns—including dividends, currency fluctuations, and corporate actions—all in one clear, consolidated dashboard.

---

## Key Features

- **Consolidated Dashboard**: View all your investments across multiple asset classes in a single, unified view.
- **Exact Decimal Arithmetic**: Zero floating-point drift. All money, quantities, prices, and rates use fixed-point decimal arithmetic from database to browser.
- **Interactive Performance Chart**: Timeframe-selectable SVG performance chart (1D, 1W, 1M, YTD, 1Y, ALL) rendering cubic Bezier curves, dynamic profit/loss color gradients, crosshair tracking, inspection tooltips, and exact period returns.
- **Interactive Transaction Logging**: Record Buys, Sells, Cash Deposits, Withdrawals, and Dividends directly in the UI. Holding quantities, cost bases, cash balances, and returns re-project deterministically in real time.
- **Transaction History & Ledger Management**: Paginated transaction history with type filtering, exact execution prices, fee tracking, and safe deletion with immediate atomic ledger replay.
- **Transaction-Ledger Architecture**: Immutable transaction ledger acts as the authoritative source of truth, deterministically projecting holdings, cash balances, and valuations.
- **Flexible Cost Basis Accounting**: Native support for both **Average Cost** (`AVERAGE_COST`, default) and **FIFO** (`FIFO`) tax lot relief strategies.
- **User Preferences & Dynamic Multi-Currency Re-anchoring**: Manage investor profile display name, UI theme (`DARK`, `LIGHT`, `SYSTEM`), and base display currency (`USD`, `EUR`, `GBP`, `AUD`, `CAD`, `JPY`, `CHF`). Changing preferred currency automatically triggers atomic base currency re-anchoring on `portfolio-api`, re-scaling valuations and holdings cost bases via live FX triangulation without data drift.
- **Clean Microservice Monorepo**: Contract-first gRPC services with a Go GraphQL Backend-for-Frontend (BFF) and strongly-typed frontend queries.
- **Responsive UI**: Glassmorphic, dark-mode dashboard built with React 19, TypeScript, Vite, modal transaction entry, investor preferences dialog, and instant reactive state refresh.

> 📋 *For a comprehensive list of completed milestones, in-progress components, and planned roadmap items, see [`FEATURES.md`](file:///Users/oscargarcia/workspace/graphfolio/FEATURES.md).*

---

## Tech Stack

| Layer | Technologies |
| --- | --- |
| **Frontend** | React 19, TypeScript, Vite, Vanilla CSS, GenQL |
| **Backend-for-Frontend (BFF)** | Go 1.26+, GraphQL (`gqlgen`), gRPC Client |
| **Core Microservices** | Go 1.26+, gRPC, Protocol Buffers (`protoc-gen-go`, `protoc-gen-go-grpc`) |
| **Database & Persistence** | PostgreSQL 17+ (18+ supported), `pgx/v5`, `golang-migrate` |
| **Mocking & Testing** | `go.uber.org/mock` (`mockgen`), standard library Go test runner with `-race` |
| **Orchestration** | GNU Make, Go Workspaces (`go.work`) |

---

## Prerequisites

Ensure the following tools are installed on your machine:

- **Go 1.26+**
- **Node.js 20+** and **npm**
- **PostgreSQL 17+** running locally at `localhost:5432` (`psql`, `pg_isready`)
- **`golang-migrate` CLI**: `brew install golang-migrate`
- **Protobuf Compiler (`protoc`)**: `brew install protobuf` *(only required if editing `.proto` files)*

---

## Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/garcios/graphfolio.git
cd graphfolio
```

### 2. Configure Environment

Copy the example environment file:

```bash
cp .env.example .env
```

The default `.env` is pre-configured for a local PostgreSQL instance:

```env
PG_ADMIN_URL=postgres://$(USER)@localhost:5432/postgres
PORTFOLIO_DB_URL=postgres://portfolio_svc:portfolio@localhost:5432/graphfolio?sslmode=disable
USER_DB_URL=postgres://user_svc:user@localhost:5432/graphfolio?sslmode=disable
MIGRATE_ON_START=false
```

### 3. Install Dependencies

Install all frontend npm packages and backend Go module dependencies:

```bash
make install
```

### 4. Bootstrap and Migrate Database

Set up database roles, schemas, migrations, and development seed data:

```bash
# Verify PostgreSQL is running
make db-check

# Bootstrap roles (portfolio_svc, user_svc), database (graphfolio), and schemas
make db-bootstrap

# Run all schema migrations for portfolio-api and user-api
make migrate-up

# Seed development market data, instruments, demo portfolios, and transactions
make db-seed

# (Alternatively, run all four steps with: make db-reset)
```

### 5. Generate Code Contracts, Models, and Mocks

Compile protobufs, generate GraphQL models, regenerate the typed GenQL frontend client, and generate Go unit test mocks:

```bash
make generate
```

### 6. Start Development Servers

Start the Portfolio gRPC API, the User gRPC API, the GraphQL BFF, and the Vite frontend concurrently:

```bash
make run
```

- **Primary Investor Web App**: [http://localhost:5173](http://localhost:5173)
- **Internal Admin Portal**: [http://localhost:5174](http://localhost:5174) *(run separately via `make run-admin`)*
- **GraphQL Playground (BFF)**: [http://localhost:8080/](http://localhost:8080/)
- **Portfolio gRPC API**: `localhost:50051`
- **User gRPC API**: `localhost:50052`

---

## Architecture & Data Flow

GraphFolio enforces a strict separation of concerns across layered boundaries:

```text
React Components (Dashboard, PerformanceChart, TransactionLedger, UserPreferencesModal)
      │
      ▼  (Typed queries & mutations via GenQL)
GraphQL Backend-for-Frontend (BFF)  [localhost:8080]
      ├── gRPC (proto/portfolio/v1)
      │   ▼
      │   Portfolio API Service  [localhost:50051]
      │   └── pgxpool -> PostgreSQL [localhost:5432/graphfolio] -> portfolio schema
      │
      └── gRPC (proto/user/v1)
          ▼
          User API Service  [localhost:50052]
          └── pgxpool -> PostgreSQL [localhost:5432/graphfolio] -> users schema
```

### Exact Decimal Precision

Financial applications cannot tolerate IEEE 754 binary floating-point rounding errors:
- Contract layer defines fixed-point types in [`proto/common/v1/decimal.proto`](file:///Users/oscargarcia/workspace/graphfolio/proto/common/v1/decimal.proto).
- Shared package [`pkg/decimalpb`](file:///Users/oscargarcia/workspace/graphfolio/pkg/decimalpb/decimalpb.go) converts losslessly between protobuf messages and `shopspring/decimal.Decimal`.
- PostgreSQL drivers automatically register the `pgx-shopspring-decimal` extension on every pool connection.
- GraphQL exposes a dedicated `Decimal` scalar in [`bff/graph/model/decimal.go`](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/model/decimal.go).

### Ledger Replay & Cost Basis Methods

1. **Transaction Ledger**: All investment events (buys, sells, dividends, transfers, deposits, withdrawals) are appended to `portfolio.transactions`.
2. **Deterministic Projections**: When transactions are added, deleted, or cost-basis methods switch, the replay engine in [`services/portfolio-api/internal/service/projection.go`](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/service/projection.go) re-evaluates all transactions to produce:
   - `tax_lots` and `lot_disposals`
   - `holding_projections` (quantity, cost basis, realized PnL, dividend income)
   - `cash_balances`
3. **Atomic Replacement**: Rebuild runs under a row lock (`SELECT ... FOR UPDATE` on `portfolio.portfolios`) to serialize concurrent updates and atomically save recomputed projections in `SaveProjectionsTx`.
4. **Accounting Methods**:
   - `AVERAGE_COST`: Disposals proportionally relieve cost basis across all open tax lots.
   - `FIFO`: Disposals relieve the earliest acquired tax lots first.

### User Identity, Preferences & Dynamic Currency Re-anchoring

1. **User Domain Microservice**: `services/user-api` (Port `:50052`) owns the `users.users` table, isolating user profile identity, display name, UI theme (`DARK`, `LIGHT`, `SYSTEM`), and preferred display currency (`USD`, `EUR`, `GBP`, `AUD`, `CAD`, `JPY`, `CHF`).
2. **Atomic Base Currency Synchronization**: When an investor updates their display currency in `UserPreferencesModal`:
   - BFF dispatches `UpdateUserPreferences` to `user-api` to persist user profile settings.
   - BFF dispatches `UpdatePortfolioBaseCurrency` to `portfolio-api` within an atomic database transaction.
   - `portfolio.portfolios.base_currency` is updated, and historical valuation snapshots (`portfolio.portfolio_valuations`) and holding cost bases (`portfolio.holdings`) are re-scaled proportionally via live FX rates with triangulation fallback.
   - The GraphQL mutation returns the re-calculated portfolio, instantly updating all cards, tables, and metrics without floating-point error.
3. **Query-Level Consistency**: The GraphQL `portfolio` query automatically verifies the investor's preferred display currency with `user-api` and aligns `portfolio-api`'s base currency if necessary.

---

## Directory Structure

```text
├── proto/                        # Single Source of Truth for APIs (Protobuf definitions)
│   ├── common/v1/decimal.proto   # Decimal and Money contracts
│   ├── portfolio/v1/             # Portfolio gRPC service (GetPortfolio, UpdatePortfolioBaseCurrency, AddTransaction, etc.)
│   └── user/v1/user.proto        # User gRPC service (GetUserPreferences, UpdateUserPreferences, ListSupportedCurrencies)
├── pkg/                          # Shared Go infrastructure
│   ├── database/                 # pgx connection pooling, auto .env loading, migration runner
│   └── decimalpb/                # Decimal/Money conversions between proto and shopspring
├── scripts/db/                   # Database bootstrap and teardown scripts
├── services/                     # Domain Microservices
│   ├── portfolio-api/            # Portfolio business logic, ledger replay, valuations, migrations, seeds
│   │   ├── cmd/server/           # Application entrypoint (:50051)
│   │   ├── internal/             # Domain entities, repository, service, and gRPC server
│   │   ├── migrations/           # Versioned schema migrations (000001 - 000006)
│   │   └── seeds/                # Seed fixtures (dev_seed.sql with 365-day history)
│   └── user-api/                 # User domain microservice (:50052)
│       ├── cmd/server/           # Application entrypoint (:50052)
│       ├── internal/             # User domain, repository (pgx), service, and gRPC server
│       ├── migrations/           # Versioned schema migrations (000001, 000002_add_theme)
│       └── seeds/                # User dev fixtures (dev_seed.sql with demo investor)
├── bff/                          # GraphQL Backend-for-Frontend (gqlgen on :8080)
│   ├── graph/                    # Schema, resolvers, helpers, model
│   └── cmd/server/               # BFF entrypoint
├── web/                          # Frontend Workspace Monorepo
│   ├── apps/
│   │   ├── main-app/             # Primary Investor React App (:5173)
│   │   │   └── src/components/   # Dashboard, PerformanceChart, TransactionLedger, UserPreferencesModal
│   │   └── admin-app/            # Internal Operations Portal (:5174)
│   └── packages/
│       ├── ui/                   # Shared Design System (@graphfolio/ui)
│       └── api-client/           # Shared GraphQL Client (@graphfolio/api-client)
├── docs/                         # Implementation plans and guides
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
│   ├── user-stories/             # Product specifications & acceptance criteria
│   ├── competitive-analysis.md
│   └── genql-usage.md
├── FEATURES.md                   # Master features matrix & roadmap (completed & planned)
├── Makefile                      # Standardized development workflows
└── go.work                       # Go workspace mapping modules
```

---

## Testing & Quality Assurance

### Running Tests

Execute unit tests across all Go workspace modules with race condition detection:

```bash
make test
```

### Unit Tests with Mocks (`go.uber.org/mock`)

Service and transport layer unit tests run with zero external dependencies using mocks generated by `mockgen`:
- **Repository Mock**: [`MockRepository`](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/repository/mocks/mock_repository.go) allows verifying `PortfolioService` business calculations, error handling, and projection rebuilds without a database.
- **Service Mock**: [`MockPortfolioService`](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/service/mocks/mock_service.go) allows verifying gRPC server request routing, error code mapping (`codes.NotFound`, `codes.Internal`), and fallback behavior.

To regenerate mocks after modifying Go interfaces:

```bash
make generate
```

---

## Available Make Commands

| Command | Description |
| --- | --- |
| `make install` | Installs npm packages and downloads all Go module dependencies |
| `make proto` | Compiles `.proto` contracts into Go packages |
| `make generate` | Runs proto compilation, gqlgen generation, genql client build, and mockgen |
| `make db-check` | Tests if local PostgreSQL is accepting connections on localhost:5432 |
| `make db-bootstrap` | Idempotently creates database roles (`portfolio_svc`, `user_svc`), DB, and schemas |
| `make db-drop` | Drops local development database |
| `make migrate-up` | Applies all pending migrations across all microservices |
| `make migrate-down` | Rolls back the latest migration step across services |
| `make db-seed` | Seeds development instruments, market data, and demo portfolios |
| `make db-reset` | Recreates database from scratch: drop, bootstrap, migrate, and seed |
| `make test` | Runs unit tests across all modules in `go.work` with race detection |
| `make run-portfolio` | Starts the Portfolio gRPC microservice on port 50051 |
| `make run-user` | Starts the User gRPC microservice on port 50052 |
| `make run-bff` | Starts the GraphQL BFF on port 8080 |
| `make run-web` | Starts the Primary Investor frontend (`main-app`) on port 5173 |
| `make run-admin` | Starts the Internal Operations Portal (`admin-app`) on port 5174 |
| `make run-all-web` | Concurrently starts both `main-app` and `admin-app` |
| `make build-web` | Builds production bundles for all web workspaces |
| `make run` | Concurrently starts Portfolio API, User API, BFF, and React web servers |

---

## Documentation Links

- **Repository Rules & Guidelines**: [`AGENTS.md`](./AGENTS.md)
- **Frontend Workspace Architecture Plan**: [`docs/plans/web-workspace-refactoring-plan.md`](./docs/plans/web-workspace-refactoring-plan.md)
- **Database Architecture & Schema Design**: [`docs/plans/db-implementation-plan.md`](./docs/plans/db-implementation-plan.md)
- **Portfolio Service Implementation Plan**: [`docs/plans/portfolio-service-implementation-plan.md`](./docs/plans/portfolio-service-implementation-plan.md)
- **Add Transactions Implementation Plan**: [`docs/plans/add-transactions-implementation-plan.md`](./docs/plans/add-transactions-implementation-plan.md)
- **Performance Chart Implementation Plan**: [`docs/plans/performance-chart-implementation-plan.md`](./docs/plans/performance-chart-implementation-plan.md)
- **Transaction History Implementation Plan**: [`docs/plans/transaction-history-implementation-plan.md`](./docs/plans/transaction-history-implementation-plan.md)
- **Cost Basis Switching Plan**: [`docs/plans/cost-basis-switching-implementation-plan.md`](./docs/plans/cost-basis-switching-implementation-plan.md)
- **Portfolio Valuation Engine Plan**: [`docs/plans/portfolio-valuation-engine-implementation-plan.md`](./docs/plans/portfolio-valuation-engine-implementation-plan.md)
- **Market Data Ingestion Plan**: [`docs/plans/market-data-ingestion-implementation-plan.md`](./docs/plans/market-data-ingestion-implementation-plan.md)
- **User Preferences Implementation Plan**: [`docs/plans/user-preferences-implementation-plan.md`](./docs/plans/user-preferences-implementation-plan.md)
- **Fundamental & Cash Flow Quality Engine Plan**: [`docs/plans/fundamental-cash-flow-engine-implementation-plan.md`](./docs/plans/fundamental-cash-flow-engine-implementation-plan.md)
- **Currency Pair Historical Prices Plan**: [`docs/plans/currency-pair-historical-prices-implementation-plan.md`](./docs/plans/currency-pair-historical-prices-implementation-plan.md)
- **Admin Historical Backfill Plan**: [`docs/plans/admin-historical-backfill-implementation-plan.md`](./docs/plans/admin-historical-backfill-implementation-plan.md)
- **Ingestion Job History Plan**: [`docs/plans/ingestion-job-history-implementation-plan.md`](./docs/plans/ingestion-job-history-implementation-plan.md)
- **Competitive Strategy Analysis**: [`docs/competitive-analysis.md`](./docs/competitive-analysis.md)
- **Frontend GraphQL Setup**: [`docs/genql-usage.md`](./docs/genql-usage.md)



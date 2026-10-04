# GraphFolio

A modern portfolio tracker built for serious investors. GraphFolio accurately measures your total annualized returns—including dividends, currency fluctuations, and corporate actions—all in one clear, consolidated dashboard.

---

## Key Features

- **Consolidated Dashboard**: View all your investments across multiple asset classes in a single, unified view.
- **Exact Decimal Arithmetic**: Zero floating-point drift. All money, quantities, prices, and rates use fixed-point decimal arithmetic from database to browser.
- **Transaction-Ledger Architecture**: Immutable transaction ledger acts as the source of truth, deterministically projecting holdings, cash balances, and valuations.
- **Flexible Cost Basis Accounting**: Native support for both **Average Cost** (`AVERAGE_COST`, default) and **FIFO** (`FIFO`) tax lot relief strategies.
- **Clean Microservice Monorepo**: Contract-first gRPC services with a Go GraphQL Backend-for-Frontend (BFF) and strongly-typed frontend queries.
- **Responsive UI**: Glassmorphic, dark-mode dashboard built with React 19, TypeScript, and Vite.

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

Start the Portfolio gRPC API, the GraphQL BFF, and the Vite frontend concurrently:

```bash
make run
```

- **Web Frontend**: [http://localhost:5173](http://localhost:5173)
- **GraphQL Playground (BFF)**: [http://localhost:8080/](http://localhost:8080/)
- **Portfolio gRPC API**: `localhost:50051`

---

## Architecture & Data Flow

GraphFolio enforces a strict separation of concerns across layered boundaries:

```text
React Component (Dashboard)
      │
      ▼  (Typed queries via GenQL)
GraphQL Backend-for-Frontend (BFF)  [localhost:8080]
      │
      ▼  (gRPC via Protocol Buffers)
Portfolio API Service  [localhost:50051]
      │
      ▼  (pgxpool connection pool with shopspring/decimal)
PostgreSQL Database  [localhost:5432/graphfolio]
      ├── portfolio schema (owned by portfolio_svc)
      └── users schema     (owned by user_svc)
```

### Exact Decimal Precision

Financial applications cannot tolerate IEEE 754 binary floating-point rounding errors:
- Contract layer defines fixed-point types in [`proto/common/v1/decimal.proto`](file:///Users/oscargarcia/workspace/graphfolio/proto/common/v1/decimal.proto).
- Shared package [`pkg/decimalpb`](file:///Users/oscargarcia/workspace/graphfolio/pkg/decimalpb/decimalpb.go) converts losslessly between protobuf messages and `shopspring/decimal.Decimal`.
- PostgreSQL drivers automatically register the `pgx-shopspring-decimal` extension on every pool connection.
- GraphQL exposes a dedicated `Decimal` scalar in [`bff/graph/model/decimal.go`](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/model/decimal.go).

### Ledger Replay & Cost Basis Methods

1. **Transaction Ledger**: All investment events (buys, sells, dividends, transfers, deposits, withdrawals) are appended to `portfolio.transactions`.
2. **Deterministic Projections**: The replay engine in [`services/portfolio-api/internal/service/projection.go`](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/service/projection.go) re-evaluates all transactions to produce:
   - `tax_lots` and `lot_disposals`
   - `holding_projections` (quantity, cost basis, realized PnL, dividend income)
   - `cash_balances`
3. **Accounting Methods**:
   - `AVERAGE_COST`: Disposals proportionally relieve cost basis across all open tax lots.
   - `FIFO`: Disposals relieve the earliest acquired tax lots first.

---

## Directory Structure

```text
├── proto/                        # Single Source of Truth for APIs (Protobuf definitions)
│   ├── common/v1/decimal.proto   # Decimal and Money contracts
│   └── portfolio/v1/             # Portfolio gRPC service contract
├── pkg/                          # Shared Go infrastructure
│   ├── database/                 # pgx connection pooling, auto .env loading, migration runner
│   └── decimalpb/                # Decimal/Money conversions between proto and shopspring
├── scripts/db/                   # Database bootstrap and teardown scripts
├── services/                     # Domain Microservices
│   ├── portfolio-api/            # Portfolio business logic, ledger replay, migrations, seeds
│   │   ├── cmd/server/           # Application entrypoint
│   │   ├── internal/             # Domain calculator, repository, service, and gRPC handler
│   │   ├── migrations/           # Versioned schema migrations (000001 - 000006)
│   │   └── seeds/                # Seed fixtures (dev_seed.sql)
│   └── user-api/                 # User domain microservice and migrations
├── bff/                          # GraphQL Backend-for-Frontend (gqlgen)
│   ├── graph/                    # Schema, resolvers, helpers, custom Decimal scalar
│   └── cmd/server/               # BFF entrypoint
├── web/                          # Frontend React application
│   ├── src/components/           # Reusable UI elements (Dashboard, etc.)
│   └── src/generated/            # Auto-generated typed GenQL client
├── docs/                         # Implementation plans and guides
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
| `make run-web` | Starts the Vite React frontend on port 5173 |
| `make run` | Concurrently starts Portfolio API, BFF, and React web servers |

---

## Documentation Links

- **Repository Rules & Guidelines**: [`AGENTS.md`](./AGENTS.md)
- **Database Architecture & Schema Design**: [`docs/db-implementation-plan.md`](./docs/db-implementation-plan.md)
- **Portfolio Service Implementation Plan**: [`docs/portfolio-service-implementation-plan.md`](./docs/portfolio-service-implementation-plan.md)
- **Frontend GraphQL Setup**: [`docs/genql-usage.md`](./docs/genql-usage.md)
